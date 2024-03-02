package commands

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

type Command struct {
	Meta         *discordgo.ApplicationCommand
	ChatInputRun func(s *discordgo.Session, i *discordgo.InteractionCreate)
	MessageRun   func(s *discordgo.Session, m *discordgo.MessageCreate, args ...string)
}

var cmds = make(map[string]*Command)
var registeredCommands = []*discordgo.ApplicationCommand{}

func Register(s *discordgo.Session, guildID *string) (commands map[string]*Command, registeredCommands []*discordgo.ApplicationCommand) {
	fmt.Printf("[Commands] Registering %v commands\n", len(cmds))
	for _, command := range cmds {
		registeredCommand, err := s.ApplicationCommandCreate(s.State.User.ID, *guildID, command.Meta)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v\n", command.Meta.Name, err)
		}
		fmt.Printf("[Command] Registered '%v' command\n", registeredCommand.Name)
		registeredCommands = append(registeredCommands, registeredCommand)
	}

	return cmds, registeredCommands
}

func Unregister(s *discordgo.Session, guildID *string) {
	for _, command := range registeredCommands {
		err := s.ApplicationCommandDelete(s.State.User.ID, *guildID, command.ID)
		if err != nil {
			log.Panicf("Cannot delete '%v' command: %v", command.Name, err)
		}
	}
}
