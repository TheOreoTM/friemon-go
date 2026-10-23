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

// giveXP adds message XP to the author's selected character and announces level-ups.
func (b *Bot) giveXP(ctx context.Context, e *events.GuildMessageCreate) {
	up, err := b.store.AddXPToSelected(ctx, e.Message.Author.ID, game.RollMessageXP(newRand()))
	if err != nil {
		log.Printf("add XP for %s: %v", e.Message.Author.ID, err)
		return
	}
	if up == nil {
		return
	}

	c := characterData(up.CharacterID)
	msg := fmt.Sprintf("Congratulations %s! Your **%s** is now level %d!", e.Message.Author.Mention(), c.Name, up.NewLevel)
	if newMoves := game.NewMovesBetween(c, up.OldLevel, up.NewLevel); len(newMoves) > 0 {
		msg += fmt.Sprintf("\nIt can now learn %s. Use `/learn` to teach it.", moveNames(newMoves))
	}
	if _, err := e.Client().Rest.CreateMessage(e.ChannelID, discord.MessageCreate{Content: msg}); err != nil {
		log.Printf("announce level up in %s: %v", e.ChannelID, err)
	}
}

func (b *Bot) handleSelect(e *handler.CommandEvent) error {
	idx := e.SlashCommandInteractionData().Int("number")
	ok, err := b.store.SelectCharacter(e.Ctx, e.User().ID, idx)
	if err != nil {
		return err
	}
	if !ok {
		return e.CreateMessage(ephemeral(fmt.Sprintf("You do not have a character #%d.", idx)))
	}
	oc, _, err := b.store.GetCharacter(e.Ctx, e.User().ID, idx)
	if err != nil {
		return err
	}
	c := characterData(oc.CharacterID)
	return e.CreateMessage(discord.MessageCreate{Content: fmt.Sprintf(
		"You selected your level %d **%s** (#%d).", oc.Level, c.Name, idx)})
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
	stats := game.CalcStats(c.BaseStats, oc.IVs, oc.Level)

	var d strings.Builder
	if oc.Level < game.MaxLevel {
		fmt.Fprintf(&d, "**XP:** %d/%d\n", oc.XP, game.XPToNext(oc.Level))
	} else {
		d.WriteString("**XP:** Max level\n")
	}
	fmt.Fprintf(&d, "**Types:** %s\n\n", typeNames(c.Types))

	statLine := func(label string, value, iv int) {
		fmt.Fprintf(&d, "**%s:** %d – IV: %d/%d\n", label, value, iv, game.MaxIV)
	}
	statLine("HP", stats.HP, oc.IVs.HP)
	statLine("Attack", stats.Attack, oc.IVs.Attack)
	statLine("Defense", stats.Defense, oc.IVs.Defense)
	statLine("Magic Atk", stats.MagicAttack, oc.IVs.MagicAttack)
	statLine("Magic Def", stats.MagicDefense, oc.IVs.MagicDefense)
	statLine("Speed", stats.Speed, oc.IVs.Speed)
	fmt.Fprintf(&d, "**Total IV:** %.2f%%", game.IVPercent(oc.IVs))

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
