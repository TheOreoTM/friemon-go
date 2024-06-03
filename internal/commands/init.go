package commands

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/theoreotm/gordinal/internal/handler"
)

var cmds = make(map[string]*handler.Command)
var registeredCommands = []*discordgo.ApplicationCommand{}

func Register(s *discordgo.Session, guildID *string) (map[string]*handler.Command, []*discordgo.ApplicationCommand) {
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
