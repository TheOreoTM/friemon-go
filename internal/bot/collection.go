package bot

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/paginator"

	"github.com/TheOreoTM/friemon-go/internal/data"
	"github.com/TheOreoTM/friemon-go/internal/game"
)

const perPage = 20

// sep is a wide gap around the dots; a plain space would collapse in Discord.
const sep = "　•　" // 　 is the "ideographic space", a full-width space

func (b *Bot) handleCollection(e *handler.CommandEvent) error {
	user := e.User()
	total, err := b.store.CountCharacters(e.Ctx, user.ID)
	if err != nil {
		return err
	}
	if total == 0 {
		return e.CreateMessage(ephemeral("You have not caught any characters yet. Watch for spawns and use `/catch`!"))
	}

	return b.pages.Create(e.Respond, paginator.Pages{
		ID:         e.ID().String(),
		Pages:      (total + perPage - 1) / perPage, // integer division that rounds up
		Creator:    user.ID,                         // only the person who ran the command can turn pages
		ExpireMode: paginator.ExpireModeAfterLastUsage,
		// PageFunc runs each time a page is shown, so only that page is loaded.
		PageFunc: func(page int, embed discord.Embed) discord.Embed {
			embed = embed.WithTitle("Your characters")
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			list, err := b.store.ListCharacters(ctx, user.ID, perPage, page*perPage)
			if err != nil {
				log.Printf("list characters for %s: %v", user.ID, err)
				return embed.WithDescription("Could not load this page. Please try again.")
			}
			if len(list) == 0 {
				return embed.WithDescription("Nothing on this page.")
			}

			// Right-align the numbers so the names line up, padding to the widest number on the page.
			width := len(strconv.Itoa(list[len(list)-1].Idx))
			var lines strings.Builder
			for _, oc := range list {
				name, emoji := oc.CharacterID, "❓" // fall back if a character was removed from data
				if c, ok := data.CharacterByID(oc.CharacterID); ok {
					name, emoji = c.Name, c.DisplayEmoji()
				}
				if oc.Favorite {
					name += " ❤️"
				}
				// %*d pads the number to `width` characters, taking the width as an argument.
				fmt.Fprintf(&lines, "`%*d`　%s **%s**%sLvl. %d%s%.2f%%\n",
					width, oc.Idx, emoji, name, sep, oc.Level, sep, game.IVPercent(oc.IVs))
			}

			first := page*perPage + 1
			return embed.
				WithDescription(lines.String()).
				WithFooterTextf("Showing %d–%d out of %d.", first, first+len(list)-1, total)
		},
	}, false)
}
