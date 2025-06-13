package components

import (
	"fmt"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/google/uuid"
	"github.com/theoreotm/friemon/constants"
	"github.com/theoreotm/friemon/internal/application/bot"
	"github.com/theoreotm/friemon/internal/core/game"
)

func init() {
	Components["/battle_challenge_accept"] = HandleChallengeAccept
	Components["/battle_challenge_decline"] = HandleChallengeDecline
	// We will add more handlers here as we create them
}

func HandleChallengeAccept(b *bot.Bot) handler.ComponentHandler {
	return func(e *handler.ComponentEvent) error {
		data := e.Vars["challenge_id"]
		fmt.Printf("Challenge ID: %s\n", data)

		challengeIDStr := strings.SplitN(e.Data.CustomID(), "/", 2)[1]
		challengeID, err := uuid.Parse(challengeIDStr)
		fmt.Printf("Challenge ID: %s\n", challengeID)
		if err != nil {
			e.CreateMessage(discord.MessageCreate{Content: "Invalid challenge ID.", Flags: discord.MessageFlagEphemeral})
			return err
		}

		challenge, exists := b.BattleManager.GetChallenge(e.User().ID)
		if !exists || challenge.ID != challengeID {
			e.CreateMessage(discord.MessageCreate{Content: "This challenge is not for you or has expired.", Flags: discord.MessageFlagEphemeral})
			return err
		}

		if e.User().ID != challenge.Challenged {
			e.CreateMessage(discord.MessageCreate{Content: "You are not the one being challenged.", Flags: discord.MessageFlagEphemeral})
			return err
		}

		// Accept the challenge and create the battle
		battle, err := b.BattleManager.AcceptChallenge(e.User().ID)
		if err != nil {
			e.UpdateMessage(discord.NewMessageUpdateBuilder().SetContentf("Failed to start battle: %s", err.Error()).Build())
			return err
		}

		player1, err := b.Client.Rest().GetUser(battle.Player1.ID)
		if err != nil {
			e.UpdateMessage(discord.NewMessageUpdateBuilder().SetContentf("Failed to get player 1 user: %s", err.Error()).Build())
			return err
		}

		player2, err := b.Client.Rest().GetUser(battle.Player2.ID)
		if err != nil {
			e.UpdateMessage(discord.NewMessageUpdateBuilder().SetContentf("Failed to get player 2 user: %s", err.Error()).Build())
			return err
		}
		// Create the main battle thread
		thread, err := e.Client().Rest().CreatePostInThreadChannel(
			e.Channel().ID(),
			discord.ThreadChannelPostCreate{
				Name:                fmt.Sprintf("Battle: %s vs %s", player1.Username, player2.Username),
				AutoArchiveDuration: discord.AutoArchiveDuration1h,
			},
		)
		if err != nil {
			e.UpdateMessage(discord.NewMessageUpdateBuilder().SetContentf("Failed to create battle thread: %s", err.Error()).Build())
			return err
		}
		battle.ThreadID = thread.ID()

		// Update original message
		e.UpdateMessage(discord.NewMessageUpdateBuilder().SetContentf("Battle accepted! Go to %s to watch!", thread.Mention()).Build())

		// Send initial messages to the threads
		sendTeamSelection(b, battle, battle.Player1)
		sendTeamSelection(b, battle, battle.Player2)

		b.Client.Rest().CreateMessage(thread.ID(), discord.MessageCreate{
			Content: "The battle is about to begin! Players are selecting their teams.",
		})

		return nil
	}
}

func HandleChallengeDecline(b *bot.Bot) handler.ComponentHandler {
	return func(e *handler.ComponentEvent) error {
		// Similar logic to accept, but decline the challenge
		challengeIDStr := strings.Split(e.Data.CustomID(), ":")[1]
		challengeID, err := uuid.Parse(challengeIDStr)
		if err != nil {
			e.CreateMessage(discord.MessageCreate{Content: "Invalid challenge ID.", Flags: discord.MessageFlagEphemeral})
			return err
		}

		challenge, exists := b.BattleManager.GetChallenge(e.User().ID)
		if !exists || challenge.ID != challengeID {
			e.CreateMessage(discord.MessageCreate{Content: "This challenge is not for you or has expired.", Flags: discord.MessageFlagEphemeral})
			return err
		}

		if e.User().ID != challenge.Challenged {
			e.CreateMessage(discord.MessageCreate{Content: "You are not the one being challenged.", Flags: discord.MessageFlagEphemeral})
			return err
		}

		b.BattleManager.DeclineChallenge(e.User().ID)

		e.UpdateMessage(discord.NewMessageUpdateBuilder().SetContentf("%s declined the challenge.", e.User().Mention()).Build())

		return nil
	}
}

func sendTeamSelection(b *bot.Bot, battle *game.Battle, player *game.BattlePlayer) {
	// This would create a private thread for each player
	// For simplicity, we'll just send an ephemeral message for now.
	// In a real implementation, you'd create a private thread.

	userChars, err := b.DB.GetCharactersForUser(b.Context, player.ID)
	if err != nil {
		b.Client.Rest().CreateMessage(battle.ThreadID, discord.MessageCreate{Content: fmt.Sprintf("Failed to get characters for %s", player.ID)})
		return
	}

	if len(userChars) < battle.Settings.TeamSize {
		b.Client.Rest().CreateMessage(battle.ThreadID, discord.MessageCreate{Content: fmt.Sprintf("%s does not have enough characters to battle!", player.ID)})
		// Here you would cancel the battle
		return
	}

	options := make([]discord.StringSelectMenuOption, 0, len(userChars))
	for _, char := range userChars {
		options = append(options, discord.NewStringSelectMenuOption(
			fmt.Sprintf("Lvl %d %s", char.Level, char.CharacterName()),
			char.ID.String(),
		).WithEmoji(discord.ComponentEmoji{Name: "⚔️"}))
	}

	selectMenu := discord.NewStringSelectMenu(
		fmt.Sprintf("battle_team_select:%s", battle.ID),
		"Select your team...",
		options...,
	).WithMaxValues(battle.Settings.TeamSize).WithMinValues(battle.Settings.TeamSize)

	embed := discord.NewEmbedBuilder().
		SetTitle("Team Selection").
		SetDescription(fmt.Sprintf("Choose your team of %d characters for the battle.", battle.Settings.TeamSize)).
		SetColor(constants.ColorInfo).
		Build()

	// This message should be ephemeral and sent to the player
	b.Client.Rest().CreateMessage(battle.ChannelID, discord.MessageCreate{
		Content:    fmt.Sprintf("<@%s>, it's your turn to select a team.", player.ID),
		Embeds:     []discord.Embed{embed},
		Components: []discord.ContainerComponent{discord.NewActionRow(selectMenu)},
		Flags:      discord.MessageFlagEphemeral,
	})
}

// NOTE: The full implementation would require many more component handlers for:
// - battle_team_select
// - battle_team_confirm
// - battle_action_attack
// - battle_action_switch
// - battle_exec_move
// - battle_exec_switch
//
// Each of these would interact with the BattleManager and send updated embeds.
// This file would become very large. For brevity, I've shown the initial challenge flow.
// The subsequent flows would follow a similar pattern of:
// 1. Parsing the custom ID.
// 2. Finding the battle in the BattleManager.
// 3. Performing the action (e.g., adding a character to a team, submitting a move).
// 4. Checking if the battle state can advance (e.g., both players confirmed teams, both players chose a move).
// 5. If so, advance the state (e.g., start the battle, process the turn).
// 6. Update the Discord messages with new embeds and components.
