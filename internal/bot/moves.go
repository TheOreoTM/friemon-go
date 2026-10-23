package bot

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	"github.com/TheOreoTM/friemon-go/internal/data"
	"github.com/TheOreoTM/friemon-go/internal/game"
)

func (b *Bot) handleMoves(e *handler.CommandEvent) error {
	oc, ok, err := b.store.SelectedCharacter(e.Ctx, e.User().ID)
	if err != nil {
		return err
	}
	if !ok {
		return e.CreateMessage(ephemeral("You have no characters yet. Catch one first!"))
	}
	c := characterData(oc.CharacterID)

	var current strings.Builder
	for _, id := range oc.Moves {
		current.WriteString(moveLine(id) + "\n")
	}
	if len(oc.Moves) == 0 {
		current.WriteString("None")
	}

	// List the whole learnset, marking moves that are known or still locked.
	var learnset strings.Builder
	for _, entry := range c.Learnset {
		mark := "✅" // can learn now
		switch {
		case slices.Contains(oc.Moves, entry.MoveID):
			mark = "☑️" // already known
		case entry.Level > oc.Level:
			mark = "🔒"
		}
		fmt.Fprintf(&learnset, "%s Lvl. %d • %s\n", mark, entry.Level, moveLine(entry.MoveID))
	}

	embed := discord.NewEmbed().
		WithTitlef("Level %d %s moves", oc.Level, c.Name).
		AddField(fmt.Sprintf("Current moves (%d/%d)", len(oc.Moves), game.MaxMoves), current.String(), false).
		AddField("Learnset", learnset.String(), false).
		WithFooterText("✅ can learn  ☑️ known  🔒 needs a higher level • Use /learn to teach a move.").
		WithColor(embedColor)
	return e.CreateMessage(discord.MessageCreate{Embeds: []discord.Embed{embed}})
}

// moveLine describes a move in one line, like "**Zoltraak** (Destruction, 60 power)".
func moveLine(id string) string {
	m, ok := data.MoveByID(id)
	if !ok {
		return id
	}
	if m.Category == data.Status {
		return fmt.Sprintf("**%s** (%s, status)", m.Name, m.Type)
	}
	return fmt.Sprintf("**%s** (%s, %d power)", m.Name, m.Type, m.Power)
}

// findMove accepts a move ID (what autocomplete sends) or a typed name.
func findMove(input string) (data.Move, bool) {
	if m, ok := data.MoveByID(input); ok {
		return m, true
	}
	want := game.Normalize(input)
	for _, m := range data.Moves {
		if game.Normalize(m.Name) == want {
			return m, true
		}
	}
	return data.Move{}, false
}

func (b *Bot) handleLearn(e *handler.CommandEvent) error {
	oc, ok, err := b.store.SelectedCharacter(e.Ctx, e.User().ID)
	if err != nil {
		return err
	}
	if !ok {
		return e.CreateMessage(ephemeral("You have no characters yet. Catch one first!"))
	}
	c := characterData(oc.CharacterID)

	move, found := findMove(e.SlashCommandInteractionData().String("move"))
	if !found {
		return e.CreateMessage(ephemeral("That move does not exist. Check `/moves` for what your character can learn."))
	}
	if err := game.CheckLearn(c, oc.Level, oc.Moves, move.ID); err != nil {
		return e.CreateMessage(ephemeral(learnErrorMessage(err, c, move)))
	}

	if len(oc.Moves) < game.MaxMoves {
		if err := b.store.SetMoves(e.Ctx, e.User().ID, oc.Idx, append(oc.Moves, move.ID)); err != nil {
			return err
		}
		return e.CreateMessage(discord.MessageCreate{Content: fmt.Sprintf("Your **%s** learned **%s**!", c.Name, move.Name)})
	}

	// Already knows 4 moves: ask which one to forget. The custom ID carries everything the
	// select-menu handler needs, since it runs later as a separate interaction.
	options := make([]discord.StringSelectMenuOption, len(oc.Moves))
	for i, id := range oc.Moves {
		name := id
		if m, ok := data.MoveByID(id); ok {
			name = m.Name
		}
		options[i] = discord.NewStringSelectMenuOption(name, id)
	}
	customID := fmt.Sprintf("/learn-replace/%s/%d/%s", e.User().ID, oc.Idx, move.ID)
	return e.CreateMessage(discord.MessageCreate{
		Content: fmt.Sprintf("Your **%s** already knows %d moves. Which one should it forget to learn **%s**?",
			c.Name, game.MaxMoves, move.Name),
		Components: []discord.LayoutComponent{
			discord.NewActionRow(discord.NewStringSelectMenu(customID, "Choose a move to forget", options...)),
		},
		Flags: discord.MessageFlagEphemeral,
	})
}

// handleLearnReplace runs when the player picks a move to forget from the /learn menu.
func (b *Bot) handleLearnReplace(e *handler.ComponentEvent) error {
	// Only the player who ran /learn may answer. The menu is ephemeral, but check anyway.
	if e.Vars["user"] != e.User().ID.String() {
		return e.CreateMessage(ephemeral("This is not your menu."))
	}
	idx, err := strconv.Atoi(e.Vars["idx"])
	if err != nil {
		return err
	}
	newMove, ok := data.MoveByID(e.Vars["move"])
	if !ok {
		return e.UpdateMessage(contentOnly("That move no longer exists."))
	}

	// Load the character again: it may have changed since the menu was sent.
	oc, ok, err := b.store.GetCharacter(e.Ctx, e.User().ID, idx)
	if err != nil {
		return err
	}
	if !ok {
		return e.UpdateMessage(contentOnly("You no longer have that character."))
	}
	c := characterData(oc.CharacterID)
	if err := game.CheckLearn(c, oc.Level, oc.Moves, newMove.ID); err != nil {
		return e.UpdateMessage(contentOnly(learnErrorMessage(err, c, newMove)))
	}

	old := e.StringSelectMenuInteractionData().Values[0]
	i := slices.Index(oc.Moves, old)
	if i < 0 {
		return e.UpdateMessage(contentOnly("Your character no longer knows that move. Run `/learn` again."))
	}
	moves := slices.Clone(oc.Moves)
	moves[i] = newMove.ID
	if err := b.store.SetMoves(e.Ctx, e.User().ID, idx, moves); err != nil {
		return err
	}
	return e.UpdateMessage(contentOnly(fmt.Sprintf("Your **%s** forgot **%s** and learned **%s**!",
		c.Name, moveNames([]string{old}), newMove.Name)))
}

// autocompleteLearn suggests moves the selected character can learn right now.
func (b *Bot) autocompleteLearn(e *handler.AutocompleteEvent) error {
	choices := []discord.AutocompleteChoice{}
	oc, ok, err := b.store.SelectedCharacter(e.Ctx, e.User().ID)
	if err != nil || !ok {
		return e.AutocompleteResult(choices) // an empty list; errors cannot be shown here
	}
	c := characterData(oc.CharacterID)
	typed := game.Normalize(e.Data.String("move"))

	for _, id := range game.LearnableMoves(c, oc.Level) {
		m, ok := data.MoveByID(id)
		if !ok || slices.Contains(oc.Moves, id) || !strings.Contains(game.Normalize(m.Name), typed) {
			continue
		}
		choices = append(choices, discord.AutocompleteChoiceString{
			Name:  fmt.Sprintf("%s (%s)", m.Name, m.Type),
			Value: m.ID,
		})
		if len(choices) == 25 { // Discord's limit
			break
		}
	}
	return e.AutocompleteResult(choices)
}

func learnErrorMessage(err error, c data.Character, m data.Move) string {
	switch {
	case errors.Is(err, game.ErrAlreadyKnown):
		return fmt.Sprintf("Your **%s** already knows **%s**.", c.Name, m.Name)
	case errors.Is(err, game.ErrLevelTooLow):
		for _, e := range c.Learnset {
			if e.MoveID == m.ID {
				return fmt.Sprintf("Your **%s** needs to be level %d to learn **%s**.", c.Name, e.Level, m.Name)
			}
		}
	case errors.Is(err, game.ErrNotInSet):
		return fmt.Sprintf("**%s** cannot learn **%s**.", c.Name, m.Name)
	}
	return err.Error()
}

// contentOnly replaces a message's text and removes its components (like a used select menu).
func contentOnly(content string) discord.MessageUpdate {
	return discord.NewMessageUpdate().WithContent(content).ClearComponents()
}
