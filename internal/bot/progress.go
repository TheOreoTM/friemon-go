package bot

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"

	"github.com/TheOreoTM/friemon-go/internal/data"
	"github.com/TheOreoTM/friemon-go/internal/game"
	"github.com/TheOreoTM/friemon-go/internal/store"
)

// giveXP adds message XP to the author's party and announces any level-ups in one message.
func (b *Bot) giveXP(ctx context.Context, e *events.GuildMessageCreate) {
	ups, err := b.store.AddPartyXP(ctx, e.Message.Author.ID, game.RollMessageXP(newRand()))
	if err != nil {
		log.Printf("add XP for %s: %v", e.Message.Author.ID, err)
		return
	}
	if len(ups) == 0 {
		return
	}

	var lines []string
	for _, up := range ups {
		c := characterData(up.CharacterID)
		line := fmt.Sprintf("Your **%s** is now level %d!", c.Name, up.NewLevel)
		if newMoves := game.NewMovesBetween(c, up.OldLevel, up.NewLevel); len(newMoves) > 0 {
			line += fmt.Sprintf(" It can now learn %s.", moveNames(newMoves))
		}
		lines = append(lines, line)
	}
	msg := fmt.Sprintf("Congratulations %s!\n%s", e.Message.Author.Mention(), strings.Join(lines, "\n"))
	if _, err := e.Client().Rest.CreateMessage(e.ChannelID, discord.MessageCreate{Content: msg}); err != nil {
		log.Printf("announce level up in %s: %v", e.ChannelID, err)
	}
}

// handleSelect makes a character the party leader.
func (b *Bot) handleSelect(e *handler.CommandEvent) error {
	user := e.User().ID
	idx := e.SlashCommandInteractionData().Int("number")
	oc, ok, err := b.store.GetCharacter(e.Ctx, user, idx)
	if err != nil {
		return err
	}
	if !ok {
		return e.CreateMessage(ephemeral(fmt.Sprintf("You do not have a character #%d.", idx)))
	}
	if b.battles.inBattle(user) {
		return e.CreateMessage(ephemeral("You cannot change your party during a battle."))
	}

	party, err := b.store.PartyIdxs(e.Ctx, user)
	if err != nil {
		return err
	}
	party, dropped := game.PartyWithLeader(party, idx)
	if err := b.store.SetParty(e.Ctx, user, party); err != nil {
		return err
	}

	c := characterData(oc.CharacterID)
	msg := fmt.Sprintf("Your level %d **%s** (#%d) now leads your party.", oc.Level, c.Name, idx)
	if dropped != 0 {
		msg += fmt.Sprintf(" #%d left the party to make room.", dropped)
	}
	return e.CreateMessage(discord.MessageCreate{Content: msg})
}

func (b *Bot) handleInfo(e *handler.CommandEvent) error {
	oc, ok, err := b.characterOption(e.Ctx, e.User().ID, e.SlashCommandInteractionData())
	if err != nil {
		return err
	}
	if !ok {
		return e.CreateMessage(ephemeral("You do not have that character. Check `/collection` for your numbers."))
	}

	c := characterData(oc.CharacterID)
	nature := natureOf(oc.Nature)
	stats := game.CalcStats(c.BaseStats, oc.IVs, oc.EVs, oc.Level, nature)

	var d strings.Builder
	if oc.Level < game.MaxLevel {
		fmt.Fprintf(&d, "**XP:** %d/%d\n", oc.XP, game.XPToNext(oc.Level))
	} else {
		d.WriteString("**XP:** Max level\n")
	}
	fmt.Fprintf(&d, "**Types:** %s\n", typeNames(c.Types))
	fmt.Fprintf(&d, "**Nature:** %s\n\n", natureText(nature))

	for _, s := range data.MainStats {
		label := shortStatName(s)
		// Mark the stats the nature changes, like the main games do.
		if !nature.Neutral() && s == nature.Up {
			label += " ▲"
		} else if !nature.Neutral() && s == nature.Down {
			label += " ▼"
		}
		fmt.Fprintf(&d, "**%s:** %d – IV: %d/%d – EV: %d\n", label, stats.Get(s), oc.IVs.Get(s), game.MaxIV, oc.EVs.Get(s))
	}
	fmt.Fprintf(&d, "**Total IV:** %.2f%%\n**Total EV:** %d/%d", game.IVPercent(oc.IVs), oc.EVs.Total(), game.MaxEVTotal)

	moves := "None. Use `/learn` to teach one."
	if len(oc.Moves) > 0 {
		moves = moveNames(oc.Moves)
	}

	embed := discord.NewEmbed().
		WithTitlef("Level %d %s %s", oc.Level, c.DisplayEmoji(), c.Name).
		WithDescription(d.String()).
		AddField("Moves", moves, false).
		WithFooterTextf("Displaying character #%d.", oc.Idx).
		WithColor(embedColor)
	return e.CreateMessage(discord.MessageCreate{Embeds: []discord.Embed{embed}})
}

// natureOf looks up a nature by ID, falling back to a neutral one.
func natureOf(id string) data.Nature {
	if n, ok := data.NatureByID(id); ok {
		return n
	}
	n, _ := data.NatureByID("hardy")
	return n
}

// natureText describes a nature, like "Adamant (+Attack, -Magic Attack)".
func natureText(n data.Nature) string {
	if n.Neutral() {
		return n.Name + " (no change)"
	}
	return fmt.Sprintf("%s (+%s, -%s)", n.Name, n.Up, n.Down)
}

// shortStatName is the stat name used in /info.
func shortStatName(s data.Stat) string {
	switch s {
	case data.StatMagicAttack:
		return "Magic Atk"
	case data.StatMagicDefense:
		return "Magic Def"
	}
	return s.String()
}

// characterOption loads the character given by the optional "number" option,
// or the user's selected character if it is left out.
func (b *Bot) characterOption(ctx context.Context, userID snowflake.ID, d discord.SlashCommandInteractionData) (store.OwnedCharacter, bool, error) {
	if idx, ok := d.OptInt("number"); ok {
		return b.store.GetCharacter(ctx, userID, idx)
	}
	return b.store.SelectedCharacter(ctx, userID)
}

// characterData looks up a character's static data, with a stand-in if it was removed from the game.
func characterData(id string) data.Character {
	if c, ok := data.CharacterByID(id); ok {
		return c
	}
	return data.Character{ID: id, Name: id}
}

// moveNames joins move IDs into readable names, like "Zoltraak, Slash".
func moveNames(ids []string) string {
	names := make([]string, len(ids))
	for i, id := range ids {
		names[i] = id
		if m, ok := data.MoveByID(id); ok {
			names[i] = m.Name
		}
	}
	return strings.Join(names, ", ")
}

func typeNames(types []data.Type) string {
	names := make([]string, len(types))
	for i, t := range types {
		names[i] = t.String()
	}
	return strings.Join(names, " / ")
}
