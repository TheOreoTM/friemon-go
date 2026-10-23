package bot

import (
	"errors"
	"fmt"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	"github.com/TheOreoTM/friemon-go/internal/game"
)

func (b *Bot) handlePartyView(e *handler.CommandEvent) error {
	party, err := b.store.Party(e.Ctx, e.User().ID)
	if err != nil {
		return err
	}
	if len(party) == 0 {
		return e.CreateMessage(ephemeral("You have no characters yet. Catch one first!"))
	}

	var d strings.Builder
	for slot := range game.MaxParty {
		if slot >= len(party) {
			fmt.Fprintf(&d, "`%d`　*empty*\n", slot+1)
			continue
		}
		oc := party[slot]
		c := characterData(oc.CharacterID)
		fmt.Fprintf(&d, "`%d`　%s **%s**%sLvl. %d%s#%d", slot+1, c.DisplayEmoji(), c.Name, sep, oc.Level, sep, oc.Idx)
		if slot == 0 {
			d.WriteString(sep + "Leader")
		}
		d.WriteString("\n")
	}

	embed := discord.NewEmbed().
		WithTitle("Your party").
		WithDescription(d.String()).
		WithFooterTextf("Chatting gives the leader full XP and the others %d%%. Use /select to change the leader.", game.PartyXPShare).
		WithColor(embedColor)
	return e.CreateMessage(discord.MessageCreate{Embeds: []discord.Embed{embed}})
}

func (b *Bot) handlePartyAdd(e *handler.CommandEvent) error {
	return b.changeParty(e, game.PartyAdd, "joined your party")
}

func (b *Bot) handlePartyRemove(e *handler.CommandEvent) error {
	return b.changeParty(e, game.PartyRemove, "left your party")
}

// changeParty applies a party change (add or remove) for the "number" option.
// Passing the function in as an argument lets add and remove share all the checks.
func (b *Bot) changeParty(e *handler.CommandEvent, change func([]int, int) ([]int, error), done string) error {
	user := e.User().ID
	idx := e.SlashCommandInteractionData().Int("number")
	if b.battles.inBattle(user) {
		return e.CreateMessage(ephemeral("You cannot change your party during a battle."))
	}
	oc, ok, err := b.store.GetCharacter(e.Ctx, user, idx)
	if err != nil {
		return err
	}
	if !ok {
		return e.CreateMessage(ephemeral(fmt.Sprintf("You do not have a character #%d.", idx)))
	}

	party, err := b.store.PartyIdxs(e.Ctx, user)
	if err != nil {
		return err
	}
	party, err = change(party, idx)
	switch {
	case errors.Is(err, game.ErrPartyFull):
		return e.CreateMessage(ephemeral(fmt.Sprintf("Your party already has %d characters. Remove one first with `/party remove`.", game.MaxParty)))
	case err != nil:
		// The other game errors are already written for players.
		return e.CreateMessage(ephemeral(capitalize(err.Error()) + "."))
	}
	if err := b.store.SetParty(e.Ctx, user, party); err != nil {
		return err
	}
	c := characterData(oc.CharacterID)
	return e.CreateMessage(discord.MessageCreate{Content: fmt.Sprintf("Your level %d **%s** (#%d) %s.", oc.Level, c.Name, idx, done)})
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
