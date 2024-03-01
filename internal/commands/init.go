package commands

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

type command struct {
	Meta    *discordgo.ApplicationCommand
	Handler func(s *discordgo.Session, i *discordgo.InteractionCreate)
}

var cmds = map[string]*command{}
var registeredCommands = []*discordgo.ApplicationCommand{}

func ExecuteCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if h, ok := cmds[i.ApplicationCommandData().Name]; ok {
		h.Handler(s, i)
	}
}

func Register(s *discordgo.Session, guildID *string) {
	for _, command := range cmds {
		registeredCommand, err := s.ApplicationCommandCreate(s.State.User.ID, *guildID, command.Meta)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v\n", command.Meta.Name, err)
		}
		fmt.Printf("[Command] Registered '%v' command\n", registeredCommand.Name)
		registeredCommands = append(registeredCommands, registeredCommand)
	}
}

func Unregister(s *discordgo.Session, guildID *string) {
	for _, command := range registeredCommands {
		err := s.ApplicationCommandDelete(s.State.User.ID, *guildID, command.ID)
		if err != nil {
			log.Panicf("Cannot delete '%v' command: %v", command.Name, err)
		}
	}
}
