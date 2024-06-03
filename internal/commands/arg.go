package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/theoreotm/gordinal/internal/handler"
)

func init() {
	cmds["arg"] = &handler.Command{
		Meta: &discordgo.ApplicationCommand{
			Name:        "arg",
			Description: "Test out the args!",
		},
		MessageRun:   messageRunArg,
		ChatInputRun: chatInputRunArg,
	}
}

func messageRunArg(s *discordgo.Session, m *discordgo.MessageCreate, args *handler.Args) {
	fmt.Printf("Args: %v\n", args)

	user := args.GetMember()
	if user == nil {
		s.ChannelMessageSend(m.ChannelID, "User not found")
		return
	}

	s.ChannelMessageSend(m.ChannelID, user.User.Username)
}

func chatInputRunArg(s *discordgo.Session, i *discordgo.InteractionCreate) {
	messageToEcho := i.ApplicationCommandData().Options[0].StringValue()

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: messageToEcho,
		},
	})
}
