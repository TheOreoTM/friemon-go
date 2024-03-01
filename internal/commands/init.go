package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type Command struct {
	Meta    *discordgo.ApplicationCommand
	Handler func(s *discordgo.Session, i *discordgo.InteractionCreate)
}

var Commands = map[string]*Command{}

func ExecuteCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if h, ok := Commands[i.ApplicationCommandData().Name]; ok {
		h.Handler(s, i)
	}
}

func Register(s *discordgo.Session, guildID *string) {
	for _, command := range Commands {
		if guildID != nil {
			_, err := s.ApplicationCommandCreate(s.State.User.ID, *guildID, command.Meta)
			if err != nil {
				fmt.Printf("Cannot create '%v' command: %v\n", command.Meta.Name, err)
			}
		} else {
			_, err := s.ApplicationCommandCreate(s.State.User.ID, "", command.Meta)
			if err != nil {
				fmt.Printf("Cannot create '%v' command: %v\n", command.Meta.Name, err)
			}
		}
	}
}
