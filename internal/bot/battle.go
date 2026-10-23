package bot

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"

	"github.com/TheOreoTM/friemon-go/internal/data"
	"github.com/TheOreoTM/friemon-go/internal/game"
)

const (
	challengeTimeout  = 2 * time.Minute // how long a challenge can wait to be accepted
	battleIdleTimeout = 5 * time.Minute // a battle with no actions for this long ends
)

type player struct {
	id   snowflake.ID
	name string
}

type challenge struct {
	challenger, opponent player
	created              time.Time
}

// activeBattle is a battle in progress plus the Discord details needed to show it.
type activeBattle struct {
	id         string
	players    [2]player
	game       *game.Battle
	choices    [2]*game.Action // what each player picked this step; nil until they pick
	step       int             // goes up whenever the battle moves on, so old buttons stop working
	log        []game.Event    // what happened since the current message was posted
	channelID  snowflake.ID
	messageID  snowflake.ID // the latest battle message; earlier turns stay in the chat
	lastAction time.Time
}

func (ab *activeBattle) playerIndex(id snowflake.ID) int {
	for i, p := range ab.players {
		if p.id == id {
			return i
		}
	}
	return -1
}

// replacing reports whether the battle is waiting for a fainted character to be replaced.
func (ab *activeBattle) replacing() bool {
	return ab.game.NeedsReplacement(0) || ab.game.NeedsReplacement(1)
}

// needsInput reports whether player s has to choose something right now.
func (ab *activeBattle) needsInput(s int) bool {
	if ab.game.Over() {
		return false
	}
	if ab.replacing() {
		return ab.game.NeedsReplacement(s)
	}
	return true
}

// waitingFor lists the players who still have to choose.
func (ab *activeBattle) waitingFor() []string {
	var names []string
	for s, p := range ab.players {
		if ab.needsInput(s) && ab.choices[s] == nil {
			names = append(names, p.name)
		}
	}
	return names
}

// battles holds all challenges and battles in memory; a restart ends them.
// One mutex guards everything, which is simple and plenty fast at this scale.
type battles struct {
	mu         sync.Mutex
	challenges map[string]*challenge
	active     map[string]*activeBattle
	byUser     map[snowflake.ID]*activeBattle
}

func newBattles() *battles {
	return &battles{
		challenges: map[string]*challenge{},
		active:     map[string]*activeBattle{},
		byUser:     map[snowflake.ID]*activeBattle{},
	}
}

// end removes a finished battle and returns the EVs to save. The caller must hold the lock.
func (bs *battles) end(ab *activeBattle) []evSave {
	delete(bs.active, ab.id)
	for _, p := range ab.players {
		delete(bs.byUser, p.id)
	}
	return evSaves(ab)
}

func (bs *battles) inBattle(id snowflake.ID) bool {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	_, ok := bs.byUser[id]
	return ok
}

// evSave is one character's new EV total after a battle.
type evSave struct {
	userID snowflake.ID
	idx    int
	evs    data.Stats
}

// evSaves adds up the EVs each character earned from its knockouts.
func evSaves(ab *activeBattle) []evSave {
	type key struct {
		side int
		f    *game.Fighter
	}
	gains := map[key]data.Stats{}
	var order []key
	for _, ko := range ab.game.KOs {
		k := key{ko.Side, ko.Scorer}
		if _, ok := gains[k]; !ok {
			order = append(order, k)
		}
		g := gains[k]
		for _, s := range data.MainStats {
			g.Set(s, g.Get(s)+ko.Fainted.EVYield.Get(s))
		}
		gains[k] = g
	}
	var saves []evSave
	for _, k := range order {
		saves = append(saves, evSave{ab.players[k.side].id, k.f.Idx, game.AddEVs(k.f.EVs, gains[k])})
	}
	return saves
}

func (b *Bot) saveEVs(saves []evSave) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	for _, s := range saves {
		if err := b.store.SetEVs(ctx, s.userID, s.idx, s.evs); err != nil {
			log.Printf("save EVs for %s #%d: %v", s.userID, s.idx, err)
		}
	}
}

// handleBattle sends a challenge to another player.
func (b *Bot) handleBattle(e *handler.CommandEvent) error {
	me := e.User()
	opp := e.SlashCommandInteractionData().User("opponent")
	switch {
	case opp.ID == me.ID:
		return e.CreateMessage(ephemeral("You cannot battle yourself."))
	case opp.Bot:
		return e.CreateMessage(ephemeral("You cannot battle a bot."))
	case b.battles.inBattle(me.ID):
		return e.CreateMessage(ephemeral("You are already in a battle."))
	case b.battles.inBattle(opp.ID):
		return e.CreateMessage(ephemeral(opp.EffectiveName() + " is already in a battle."))
	}

	// Check the challenger's party now; the opponent's is checked when they accept.
	if _, problem, err := b.loadSide(e.Ctx, player{me.ID, me.EffectiveName()}); err != nil {
		return err
	} else if problem != "" {
		return e.CreateMessage(ephemeral(problem))
	}

	id := e.ID().String()
	b.battles.mu.Lock()
	b.battles.challenges[id] = &challenge{
		challenger: player{me.ID, me.EffectiveName()},
		opponent:   player{opp.ID, opp.EffectiveName()},
		created:    time.Now(),
	}
	b.battles.mu.Unlock()

	return e.CreateMessage(discord.MessageCreate{
		Content: fmt.Sprintf("%s, %s has challenged you to a party battle! You have %d minutes to accept.",
			opp.Mention(), me.Mention(), int(challengeTimeout.Minutes())),
		Components: []discord.LayoutComponent{discord.NewActionRow(
			discord.NewSuccessButton("Accept", "/battle-accept/"+id),
			discord.NewDangerButton("Decline", "/battle-decline/"+id),
		)},
	})
}

func (b *Bot) handleBattleAccept(e *handler.ComponentEvent) error {
	id := e.Vars["id"]
	b.battles.mu.Lock()
	ch, ok := b.battles.challenges[id]
	if ok && ch.opponent.id == e.User().ID {
		delete(b.battles.challenges, id) // taken now, so a double click cannot start two battles
	}
	b.battles.mu.Unlock()

	switch {
	case !ok || time.Since(ch.created) > challengeTimeout:
		return e.UpdateMessage(contentOnly("This challenge has expired."))
	case ch.opponent.id != e.User().ID:
		return e.CreateMessage(ephemeral("Only " + ch.opponent.name + " can accept this challenge."))
	}

	// Load both parties now, since either player may have changed theirs.
	var sides [2]*game.Side
	for i, p := range []player{ch.challenger, ch.opponent} {
		side, problem, err := b.loadSide(e.Ctx, p)
		if err != nil {
			return err
		}
		if problem != "" {
			return e.UpdateMessage(contentOnly("The battle could not start. " + problem))
		}
		sides[i] = side
	}

	g, start := game.NewBattle(sides[0], sides[1])
	ab := &activeBattle{
		id:         id,
		players:    [2]player{ch.challenger, ch.opponent},
		game:       g,
		log:        start,
		channelID:  e.Channel().ID(),
		messageID:  e.Message.ID,
		lastAction: time.Now(),
	}
	b.battles.mu.Lock()
	_, busy0 := b.battles.byUser[ab.players[0].id]
	_, busy1 := b.battles.byUser[ab.players[1].id]
	if busy0 || busy1 {
		b.battles.mu.Unlock()
		return e.UpdateMessage(contentOnly("The battle could not start: one of you is already in another battle."))
	}
	b.battles.active[id] = ab
	for _, p := range ab.players {
		b.battles.byUser[p.id] = ab
	}
	update := battleUpdate(ab, true)
	b.battles.mu.Unlock()

	return e.UpdateMessage(update)
}

func (b *Bot) handleBattleDecline(e *handler.ComponentEvent) error {
	id := e.Vars["id"]
	b.battles.mu.Lock()
	ch, ok := b.battles.challenges[id]
	allowed := ok && (e.User().ID == ch.opponent.id || e.User().ID == ch.challenger.id)
	if allowed {
		delete(b.battles.challenges, id)
	}
	b.battles.mu.Unlock()

	switch {
	case !ok:
		return e.UpdateMessage(contentOnly("This challenge has expired."))
	case !allowed:
		return e.CreateMessage(ephemeral("This is not your challenge."))
	case e.User().ID == ch.challenger.id:
		return e.UpdateMessage(contentOnly(ch.challenger.name + " cancelled the challenge."))
	}
	return e.UpdateMessage(contentOnly(ch.opponent.name + " declined the challenge."))
}

// handleBattleChoose shows a player their options in a message only they can see:
// their moves with PP, and who they can switch to.
func (b *Bot) handleBattleChoose(e *handler.ComponentEvent) error {
	b.battles.mu.Lock()
	ab, ok := b.battles.active[e.Vars["id"]]
	var reply discord.MessageCreate
	switch {
	case !ok:
		reply = ephemeral("This battle has ended.")
	case ab.playerIndex(e.User().ID) < 0:
		reply = ephemeral("You are not in this battle.")
	default:
		s := ab.playerIndex(e.User().ID)
		switch {
		case !ab.needsInput(s):
			reply = ephemeral("Waiting for your opponent to send out a character.")
		case ab.choices[s] != nil:
			reply = ephemeral("You already chose this turn. Waiting for your opponent.")
		default:
			reply = actionMenu(ab, s)
		}
	}
	b.battles.mu.Unlock()
	return e.CreateMessage(reply)
}

// actionMenu builds the private move and switch buttons. The caller must hold the lock.
func actionMenu(ab *activeBattle, s int) discord.MessageCreate {
	g := ab.game
	f := g.Active(s)
	prefix := fmt.Sprintf("/battle-act/%s/%d/", ab.id, ab.step)

	var rows []discord.LayoutComponent
	content := fmt.Sprintf("**%s** • %d/%d HP\nChoose a move or switch:", f.Name, f.HP, f.Stats.HP)

	if ab.replacing() {
		content = fmt.Sprintf("**%s** fainted! Choose who to send out:", f.Name)
	} else {
		var moves []discord.InteractiveComponent
		if f.HasPP() {
			for i, ms := range f.Moves {
				label := fmt.Sprintf("%s (%d/%d)", ms.Move.Name, ms.PP, ms.Move.PP)
				btn := discord.NewPrimaryButton(label, prefix+"m"+strconv.Itoa(i))
				if ms.PP == 0 {
					btn = btn.AsDisabled()
				}
				moves = append(moves, btn)
			}
		} else {
			moves = append(moves, discord.NewDangerButton("Struggle", prefix+"m0"))
		}
		rows = append(rows, discord.NewActionRow(moves...))
	}

	var switches []discord.InteractiveComponent
	for i, member := range g.Sides[s].Party {
		if g.CanSwitchTo(s, i) {
			label := fmt.Sprintf("%s (%d/%d HP)", member.Name, member.HP, member.Stats.HP)
			switches = append(switches, discord.NewSecondaryButton(label, prefix+"s"+strconv.Itoa(i)))
		}
	}
	if len(switches) > 0 {
		rows = append(rows, discord.NewActionRow(switches...))
	}

	return discord.MessageCreate{Content: content, Components: rows, Flags: discord.MessageFlagEphemeral}
}

// handleBattleAct records a move or switch. When everyone needed has chosen, the battle moves on:
// a full turn posts a new message, and sending out a replacement updates the current one.
func (b *Bot) handleBattleAct(e *handler.ComponentEvent) error {
	step, err := strconv.Atoi(e.Vars["step"])
	act := e.Vars["action"]
	if err != nil || len(act) < 2 {
		return fmt.Errorf("bad battle button ID %q", e.Data.CustomID())
	}
	idx, err := strconv.Atoi(act[1:])
	if err != nil {
		return fmt.Errorf("bad battle button ID %q", e.Data.CustomID())
	}
	action := game.Action{Switch: act[0] == 's', Index: idx}

	b.battles.mu.Lock()
	ab, ok := b.battles.active[e.Vars["id"]]
	var (
		reply     string
		messageID snowflake.ID           // the battle message as it was when we held the lock
		edit      *discord.MessageUpdate // change to that message, if any
		next      *discord.MessageCreate // a new message after a full turn
		saves     []evSave
	)
	s := -1
	if ok {
		s = ab.playerIndex(e.User().ID)
	}
	switch {
	case !ok:
		reply = "This battle has ended."
	case s < 0:
		reply = "You are not in this battle."
	case step != ab.step:
		reply = "Those buttons are out of date. Press **Choose action** again."
	case !ab.needsInput(s) || ab.choices[s] != nil:
		reply = "You have nothing to choose right now."
	case !validAction(ab, s, action):
		reply = "You cannot do that right now."
	default:
		ab.choices[s] = &action
		ab.lastAction = time.Now()
		reply = "You chose " + describeAction(ab.game, s, action) + "."
		messageID = ab.messageID

		if len(ab.waitingFor()) > 0 {
			reply += " Waiting for your opponent."
			u := battleUpdate(ab, true)
			edit = &u
			break
		}

		if ab.replacing() {
			// Replacements are added to the current message instead of starting a new one.
			for side := range ab.players {
				if ab.game.NeedsReplacement(side) {
					ev, err := ab.game.Replace(side, ab.choices[side].Index)
					if err != nil {
						log.Printf("battle %s replace: %v", ab.id, err)
					}
					ab.log = append(ab.log, ev...)
				}
			}
			ab.choices = [2]*game.Action{}
			ab.step++
			u := battleUpdate(ab, true)
			edit = &u
			break
		}

		// Close off this turn's message: final footer, no buttons. Then play the turn.
		u := battleUpdate(ab, false)
		edit = &u
		ev, err := ab.game.PlayTurn(newRand(), [2]game.Action{*ab.choices[0], *ab.choices[1]})
		ab.choices = [2]*game.Action{}
		ab.step++
		if err != nil {
			log.Printf("battle %s turn: %v", ab.id, err)
		}
		ab.log = ev
		if ab.game.Over() {
			saves = b.battles.end(ab)
		}
		c := battleCreate(ab)
		next = &c
	}
	b.battles.mu.Unlock()

	// Answer the click first (Discord wants a reply within 3 seconds), then update the battle.
	if err := e.UpdateMessage(contentOnly(reply)); err != nil {
		return err
	}
	rest := e.Client().Rest
	if edit != nil {
		if _, err := rest.UpdateMessage(ab.channelID, messageID, *edit); err != nil {
			log.Printf("update battle message %s: %v", ab.id, err)
		}
	}
	if len(saves) > 0 {
		b.saveEVs(saves)
	}
	if next == nil {
		return nil
	}

	msg, err := rest.CreateMessage(ab.channelID, *next)
	if err != nil {
		return fmt.Errorf("send next battle turn: %w", err)
	}

	// Point the battle at the new message. A player may have already chosen
	// while it was being sent; if so, refresh its footer.
	b.battles.mu.Lock()
	ab.messageID = msg.ID
	var refresh *discord.MessageUpdate
	if !ab.game.Over() && (ab.choices[0] != nil || ab.choices[1] != nil) {
		u := battleUpdate(ab, true)
		refresh = &u
	}
	b.battles.mu.Unlock()
	if refresh != nil {
		if _, err := rest.UpdateMessage(ab.channelID, msg.ID, *refresh); err != nil {
			log.Printf("refresh battle message %s: %v", ab.id, err)
		}
	}
	return nil
}

// validAction checks a choice before it is stored. The caller must hold the lock.
func validAction(ab *activeBattle, s int, a game.Action) bool {
	g := ab.game
	if a.Switch {
		return g.CanSwitchTo(s, a.Index)
	}
	if ab.replacing() {
		return false // only switching is allowed when replacing a fainted character
	}
	f := g.Active(s)
	if !f.HasPP() {
		return true // struggle
	}
	return a.Index >= 0 && a.Index < len(f.Moves) && f.Moves[a.Index].PP > 0
}

func describeAction(g *game.Battle, s int, a game.Action) string {
	if a.Switch {
		return "to send out **" + g.Sides[s].Party[a.Index].Name + "**"
	}
	f := g.Active(s)
	if !f.HasPP() {
		return "**Struggle**"
	}
	return "**" + f.Moves[a.Index].Move.Name + "**"
}

func (b *Bot) handleBattleFlee(e *handler.ComponentEvent) error {
	b.battles.mu.Lock()
	ab, ok := b.battles.active[e.Vars["id"]]
	s := -1
	var update discord.MessageUpdate
	var saves []evSave
	if ok {
		s = ab.playerIndex(e.User().ID)
		if s >= 0 {
			ab.game.Forfeit(s)
			saves = b.battles.end(ab)
			fled := fmt.Sprintf("%s fled! **%s wins!**", ab.players[s].name, ab.players[1-s].name)
			desc := strings.TrimSpace(battleLog(ab, ab.log) + "\n\n" + fled)
			update = battleUpdate(ab, false).WithEmbeds(battleEmbed(ab, false).WithDescription(desc))
		}
	}
	b.battles.mu.Unlock()

	switch {
	case !ok:
		return e.UpdateMessage(discord.NewMessageUpdate().ClearComponents())
	case s < 0:
		return e.CreateMessage(ephemeral("You are not in this battle."))
	}
	if len(saves) > 0 {
		b.saveEVs(saves)
	}
	return e.UpdateMessage(update)
}

// expireBattles runs in the background, ending challenges and battles nobody is using.
func (b *Bot) expireBattles(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		// select waits for whichever channel is ready first: a tick, or the context ending.
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		type edit struct {
			channelID, messageID snowflake.ID
			update               discord.MessageUpdate
		}
		var edits []edit
		var saves []evSave

		b.battles.mu.Lock()
		for id, ch := range b.battles.challenges {
			if time.Since(ch.created) > challengeTimeout {
				delete(b.battles.challenges, id) // its buttons now answer "expired"
			}
		}
		for _, ab := range b.battles.active {
			if time.Since(ab.lastAction) > battleIdleTimeout {
				saves = append(saves, b.battles.end(ab)...) // deleting from a map while ranging over it is allowed in Go
				u := battleUpdate(ab, false).WithEmbeds(battleEmbed(ab, false).
					WithDescription("The battle ended because nobody moved for 5 minutes."))
				edits = append(edits, edit{ab.channelID, ab.messageID, u})
			}
		}
		b.battles.mu.Unlock()

		for _, ed := range edits {
			if _, err := b.client.Rest.UpdateMessage(ed.channelID, ed.messageID, ed.update); err != nil {
				log.Printf("expire battle message: %v", err)
			}
		}
		b.saveEVs(saves)
	}
}

// loadSide builds a battle side from a player's party. If they cannot battle,
// problem explains why in words meant for players.
func (b *Bot) loadSide(ctx context.Context, p player) (side *game.Side, problem string, err error) {
	party, err := b.store.Party(ctx, p.id)
	if err != nil {
		return nil, "", err
	}
	if len(party) == 0 {
		return nil, p.name + " has no characters yet.", nil
	}
	side = &game.Side{Player: p.name}
	for _, oc := range party {
		c := characterData(oc.CharacterID)
		f := game.NewFighter(game.FighterSpec{
			Character: c, Idx: oc.Idx, Level: oc.Level, Nature: natureOf(oc.Nature),
			IVs: oc.IVs, EVs: oc.EVs, MoveIDs: oc.Moves,
		})
		if len(f.Moves) == 0 {
			return nil, fmt.Sprintf("%s's %s (#%d) knows no moves. Teach it one with `/learn`, or take it out with `/party remove`.",
				p.name, c.Name, oc.Idx), nil
		}
		side.Party = append(side.Party, f)
	}
	return side, "", nil
}

// winColor marks a finished battle.
const winColor = 0xf1c40f

// battleEmbed draws the battle as it stands: what happened since this message was posted,
// and both sides' active character and party. The caller must hold the battles lock.
func battleEmbed(ab *activeBattle, buttons bool) discord.Embed {
	g := ab.game
	embed := discord.NewEmbed().WithColor(embedColor)
	switch {
	case g.Over():
		embed = embed.WithTitlef("🏆 %s wins!", ab.players[g.Winner].name).WithColor(winColor)
	case g.Turn == 0:
		embed = embed.WithTitlef("⚔️ %s vs %s", ab.players[0].name, ab.players[1].name)
	default:
		embed = embed.WithTitlef("⚔️ Turn %d", g.Turn)
	}

	desc := battleLog(ab, ab.log)
	if g.Turn == 0 && !g.Over() {
		desc += "\n\nPress **Choose action** to pick a move or switch."
	}
	embed = embed.WithDescription(desc)

	// Inline fields sit side by side, one per player.
	for s, p := range ab.players {
		embed = embed.AddField(p.name, sideSummary(g, s), true)
	}

	waiting := ab.waitingFor()
	switch {
	case g.Over():
	case buttons && ab.replacing():
		embed = embed.WithFooterTextf("Waiting for %s to send out a character", strings.Join(waiting, " and "))
	case buttons:
		embed = embed.WithFooterTextf("Turn %d • Waiting for %s", g.Turn+1, strings.Join(waiting, " and "))
	case len(waiting) == 0:
		embed = embed.WithFooterText("Both players have chosen.")
	}
	return embed
}

// sideSummary shows a side's active character, its condition and stat changes, and the party.
func sideSummary(g *game.Battle, s int) string {
	f := g.Active(s)
	var b strings.Builder
	fmt.Fprintf(&b, "**%s** • Lvl. %d\n%s", f.Name, f.Level, typeNames(f.Types))
	if f.Condition != data.NoCondition {
		fmt.Fprintf(&b, " • **%s**", conditionWord(f.Condition))
	}
	fmt.Fprintf(&b, "\n`%s`\n**%d**/%d HP", hpBar(f.HP, f.Stats.HP), f.HP, f.Stats.HP)

	var stages []string
	for _, st := range []data.Stat{data.StatAttack, data.StatDefense, data.StatMagicAttack,
		data.StatMagicDefense, data.StatSpeed, data.StatAccuracy, data.StatEvasion} {
		if n := f.Stage(st); n != 0 {
			stages = append(stages, fmt.Sprintf("%s %+d", shortStat(st), n))
		}
	}
	if len(stages) > 0 {
		b.WriteString("\n" + strings.Join(stages, ", "))
	}

	b.WriteString("\n")
	for _, member := range g.Sides[s].Party {
		if member.Fainted() {
			b.WriteString("🔴")
		} else {
			b.WriteString("🟢")
		}
	}
	return b.String()
}

func shortStat(s data.Stat) string {
	switch s {
	case data.StatAttack:
		return "Atk"
	case data.StatDefense:
		return "Def"
	case data.StatMagicAttack:
		return "M.Atk"
	case data.StatMagicDefense:
		return "M.Def"
	case data.StatSpeed:
		return "Spe"
	case data.StatAccuracy:
		return "Acc"
	case data.StatEvasion:
		return "Eva"
	}
	return s.String()
}

func battleButtons(ab *activeBattle) discord.LayoutComponent {
	return discord.NewActionRow(
		discord.NewPrimaryButton("Choose action", "/battle-choose/"+ab.id),
		discord.NewDangerButton("Flee", "/battle-flee/"+ab.id),
	)
}

// battleUpdate edits the current battle message, with or without its buttons.
func battleUpdate(ab *activeBattle, buttons bool) discord.MessageUpdate {
	buttons = buttons && !ab.game.Over()
	update := discord.NewMessageUpdate().WithContent("").WithEmbeds(battleEmbed(ab, buttons))
	if !buttons {
		return update.ClearComponents()
	}
	return update.WithComponents(battleButtons(ab))
}

// battleCreate is the new message sent after each turn.
func battleCreate(ab *activeBattle) discord.MessageCreate {
	buttons := !ab.game.Over()
	msg := discord.MessageCreate{Embeds: []discord.Embed{battleEmbed(ab, buttons)}}
	if buttons {
		msg.Components = []discord.LayoutComponent{battleButtons(ab)}
	}
	return msg
}

// hpBar draws HP as 10 blocks, like ███████░░░.
func hpBar(hp, maxHP int) string {
	filled := 10 * hp / max(maxHP, 1)
	if hp > 0 && filled == 0 {
		filled = 1 // show a sliver while still standing
	}
	return strings.Repeat("█", filled) + strings.Repeat("░", 10-filled)
}
