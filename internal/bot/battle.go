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

	"github.com/TheOreoTM/friemon-go/internal/game"
)

const (
	challengeTimeout  = 2 * time.Minute // how long a challenge can wait to be accepted
	battleIdleTimeout = 5 * time.Minute // a battle with no moves for this long ends
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
	choices    [2]int   // each player's chosen move index this turn, or -1 if not chosen yet
	log        []string // what happened last turn
	channelID  snowflake.ID
	messageID  snowflake.ID
	lastAction time.Time
}

// playerIndex returns 0 or 1 for a player in this battle, or -1 for anyone else.
func (ab *activeBattle) playerIndex(id snowflake.ID) int {
	for i, p := range ab.players {
		if p.id == id {
			return i
		}
	}
	return -1
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

// end removes a finished battle. The caller must hold the lock.
func (bs *battles) end(ab *activeBattle) {
	delete(bs.active, ab.id)
	for _, p := range ab.players {
		delete(bs.byUser, p.id)
	}
}

func (bs *battles) inBattle(id snowflake.ID) bool {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	_, ok := bs.byUser[id]
	return ok
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

	// Check the challenger can fight now; the opponent is checked when they accept.
	if _, problem, err := b.loadFighter(e.Ctx, player{me.ID, me.EffectiveName()}); err != nil {
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
		Content: fmt.Sprintf("%s, %s has challenged you to a battle! You have %d minutes to accept.",
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

	// Load both characters now, since either player may have changed their selection.
	var fighters [2]*game.Fighter
	for i, p := range []player{ch.challenger, ch.opponent} {
		f, problem, err := b.loadFighter(e.Ctx, p)
		if err != nil {
			return err
		}
		if problem != "" {
			return e.UpdateMessage(contentOnly("The battle could not start. " + problem))
		}
		fighters[i] = f
	}

	ab := &activeBattle{
		id:         id,
		players:    [2]player{ch.challenger, ch.opponent},
		game:       game.NewBattle(fighters[0], fighters[1]),
		choices:    [2]int{-1, -1},
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

// handleBattleChoose shows a player their moves in a message only they can see.
func (b *Bot) handleBattleChoose(e *handler.ComponentEvent) error {
	b.battles.mu.Lock()
	ab, ok := b.battles.active[e.Vars["id"]]
	i := -1
	var moves []string
	var turn int
	alreadyChose := false
	if ok {
		i = ab.playerIndex(e.User().ID)
		if i >= 0 {
			alreadyChose = ab.choices[i] >= 0
			turn = ab.game.Turn
			for _, m := range ab.game.Fighters[i].Moves {
				moves = append(moves, m.Name)
			}
		}
	}
	b.battles.mu.Unlock()

	switch {
	case !ok:
		return e.CreateMessage(ephemeral("This battle has ended."))
	case i < 0:
		return e.CreateMessage(ephemeral("You are not in this battle."))
	case alreadyChose:
		return e.CreateMessage(ephemeral("You already chose a move this turn. Waiting for your opponent."))
	}

	// The turn number is in each button's ID, so buttons from an old turn stop working.
	buttons := make([]discord.InteractiveComponent, len(moves))
	for idx, name := range moves {
		buttons[idx] = discord.NewPrimaryButton(name, fmt.Sprintf("/battle-move/%s/%d/%d", ab.id, turn, idx))
	}
	return e.CreateMessage(discord.MessageCreate{
		Content:    "Choose a move:",
		Components: []discord.LayoutComponent{discord.NewActionRow(buttons...)},
		Flags:      discord.MessageFlagEphemeral,
	})
}

func (b *Bot) handleBattleMove(e *handler.ComponentEvent) error {
	turn, err1 := strconv.Atoi(e.Vars["turn"])
	idx, err2 := strconv.Atoi(e.Vars["idx"])
	if err1 != nil || err2 != nil {
		return fmt.Errorf("bad battle button ID %q", e.Data.CustomID())
	}

	b.battles.mu.Lock()
	ab, ok := b.battles.active[e.Vars["id"]]
	var (
		reply      string
		messageID  snowflake.ID           // the battle message as it was when we held the lock
		edit       *discord.MessageUpdate // change to that message, if any
		nextTurn   *discord.MessageCreate // a new message for the next turn, if the turn finished
		turnPlayed bool
	)
	switch {
	case !ok:
		reply = "This battle has ended."
	case ab.playerIndex(e.User().ID) < 0:
		reply = "You are not in this battle."
	case turn != ab.game.Turn:
		reply = "Those buttons are from an earlier turn."
	default:
		i := ab.playerIndex(e.User().ID)
		fighter := ab.game.Fighters[i]
		switch {
		case ab.choices[i] >= 0:
			reply = "You already chose a move this turn."
		case idx < 0 || idx >= len(fighter.Moves):
			reply = "That move is not available."
		default:
			ab.choices[i] = idx
			ab.lastAction = time.Now()
			reply = fmt.Sprintf("You chose **%s**.", fighter.Moves[idx].Name)
			messageID = ab.messageID

			if ab.choices[0] >= 0 && ab.choices[1] >= 0 {
				// Close off this turn's message: final footer, no buttons.
				u := battleUpdate(ab, false)
				edit = &u

				ab.log = ab.game.PlayTurn(newRand(), ab.choices)
				ab.choices = [2]int{-1, -1}
				if ab.game.Over() {
					b.battles.end(ab)
				}
				c := battleCreate(ab)
				nextTurn = &c
				turnPlayed = true
			} else {
				reply += " Waiting for your opponent."
				u := battleUpdate(ab, true) // same message, new "waiting for" footer
				edit = &u
			}
		}
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
	if !turnPlayed {
		return nil
	}

	msg, err := rest.CreateMessage(ab.channelID, *nextTurn)
	if err != nil {
		return fmt.Errorf("send next battle turn: %w", err)
	}

	// Point the battle at the new message. A player may have already picked a move
	// for the next turn while it was being sent; if so, refresh its footer.
	b.battles.mu.Lock()
	ab.messageID = msg.ID
	var refresh *discord.MessageUpdate
	if !ab.game.Over() && ab.choices != [2]int{-1, -1} {
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

func (b *Bot) handleBattleFlee(e *handler.ComponentEvent) error {
	b.battles.mu.Lock()
	ab, ok := b.battles.active[e.Vars["id"]]
	i := -1
	var update discord.MessageUpdate
	if ok {
		i = ab.playerIndex(e.User().ID)
		if i >= 0 {
			ab.game.Forfeit(i)
			ab.log = []string{fmt.Sprintf("%s fled! **%s wins!**", ab.players[i].name, ab.players[1-i].name)}
			b.battles.end(ab)
			update = battleUpdate(ab, false)
		}
	}
	b.battles.mu.Unlock()

	switch {
	case !ok:
		return e.UpdateMessage(discord.NewMessageUpdate().ClearComponents())
	case i < 0:
		return e.CreateMessage(ephemeral("You are not in this battle."))
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

		b.battles.mu.Lock()
		for id, ch := range b.battles.challenges {
			if time.Since(ch.created) > challengeTimeout {
				delete(b.battles.challenges, id) // its buttons now answer "expired"
			}
		}
		for _, ab := range b.battles.active {
			if time.Since(ab.lastAction) > battleIdleTimeout {
				ab.log = []string{"The battle ended because nobody moved for 5 minutes."}
				b.battles.end(ab) // deleting from a map while ranging over it is allowed in Go
				edits = append(edits, edit{ab.channelID, ab.messageID, battleUpdate(ab, false)})
			}
		}
		b.battles.mu.Unlock()

		for _, ed := range edits {
			if _, err := b.client.Rest.UpdateMessage(ed.channelID, ed.messageID, ed.update); err != nil {
				log.Printf("expire battle message: %v", err)
			}
		}
	}
}

// loadFighter builds a fighter from a player's selected character. If they cannot
// battle, problem explains why in words meant for players.
func (b *Bot) loadFighter(ctx context.Context, p player) (f *game.Fighter, problem string, err error) {
	oc, ok, err := b.store.SelectedCharacter(ctx, p.id)
	if err != nil {
		return nil, "", err
	}
	if !ok {
		return nil, p.name + " has no characters yet.", nil
	}
	c := characterData(oc.CharacterID)
	f = game.NewFighter(fmt.Sprintf("%s's %s", p.name, c.Name), c, oc.Level, oc.IVs, oc.Moves)
	if len(f.Moves) == 0 {
		return nil, fmt.Sprintf("%s's %s knows no moves. Use `/learn` first.", p.name, c.Name), nil
	}
	return f, "", nil
}

// winColor marks a finished battle.
const winColor = 0xf1c40f

// battleEmbed draws the battle as it stands. Each turn gets its own message, so
// the embed shows what happened last turn and both characters' HP after it.
// The caller must hold the battles lock.
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

	if len(ab.log) > 0 {
		embed = embed.WithDescription(strings.Join(ab.log, "\n"))
	} else {
		embed = embed.WithDescription("The battle begins! Press **Choose move** to pick your move.")
	}

	// Inline fields sit side by side, one per fighter.
	for _, f := range g.Fighters {
		embed = embed.AddField(f.Name, fmt.Sprintf("Lvl. %d • %s\n`%s`\n**%d**/%d HP",
			f.Level, typeNames(f.Types), hpBar(f.HP, f.Stats.HP), f.HP, f.Stats.HP), true)
	}

	var waiting []string
	for i, p := range ab.players {
		if ab.choices[i] < 0 {
			waiting = append(waiting, p.name)
		}
	}
	switch {
	case g.Over():
	case buttons:
		embed = embed.WithFooterTextf("Turn %d • Waiting for %s", g.Turn+1, strings.Join(waiting, " and "))
	case len(waiting) == 0:
		embed = embed.WithFooterText("Both players have chosen.")
	}
	return embed
}

func battleButtons(ab *activeBattle) discord.LayoutComponent {
	return discord.NewActionRow(
		discord.NewPrimaryButton("Choose move", "/battle-choose/"+ab.id),
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
