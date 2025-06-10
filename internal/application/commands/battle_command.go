package commands

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/theoreotm/friemon/constants"
	"github.com/theoreotm/friemon/internal/application/bot"
	"github.com/theoreotm/friemon/internal/core/game"
)

func init() {
	Commands["battle"] = cmdBattle
}

var cmdBattle = &Command{
	Cmd: discord.SlashCommandCreate{
		Name:        "battle",
		Description: "Challenge another user to a battle.",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionUser{
				Name:        "user",
				Description: "The user you want to challenge.",
				Required:    true,
			},
		},
	},
	Handler:  handleBattle,
	Category: "Battle",
}

func handleBattle(b *bot.Bot) handler.CommandHandler {
	return func(e *handler.CommandEvent) error {
		challenger := e.User()
		challenged := e.SlashCommandInteractionData().User("user")

		if challenger.ID == challenged.ID {
			e.CreateMessage(ErrorMessage("You cannot challenge yourself!"))
			return nil
		}
		if challenged.Bot {
			e.CreateMessage(ErrorMessage("You cannot challenge a bot!"))
			return nil
		}

		// Check if either player is already in a battle
		if _, inBattle := b.BattleManager.GetPlayerBattle(challenger.ID); inBattle {
			e.CreateMessage(ErrorMessage("You are already in a battle!"))
			return nil
		}
		if _, inBattle := b.BattleManager.GetPlayerBattle(challenged.ID); inBattle {
			e.CreateMessage(ErrorMessage(fmt.Sprintf("%s is already in a battle!", challenged.Username)))
			return nil
		}

		// Create the challenge
		challenge, err := b.BattleManager.CreateChallenge(challenger.ID, challenged.ID, e.ChannelID(), game.DefaultGameSettings())
		if err != nil {
			e.CreateMessage(ErrorMessage(err.Error()))
			return nil
		}

		// Send challenge embed
		embed := discord.NewEmbedBuilder().
			SetTitle("⚔️ Battle Challenge!").
			SetDescription(fmt.Sprintf("%s has challenged %s to a battle!", challenger.Mention(), challenged.Mention())).
			SetColor(constants.ColorInfo).
			AddField("Rules", "3v3, Level 100", true).
			AddField("Expires", discord.TimestampStyleRelative.Format(int64(challenge.ExpiresAt.Second())), true).
			Build()

		components := discord.NewActionRow(
			discord.NewSuccessButton("Accept", fmt.Sprintf("battle_challenge_accept:%s", challenge.ID)),
			discord.NewDangerButton("Decline", fmt.Sprintf("battle_challenge_decline:%s", challenge.ID)),
		)

		e.CreateMessage(discord.MessageCreate{
			Embeds:     []discord.Embed{embed},
			Components: []discord.ContainerComponent{components},
		})

		return nil
	}

}
