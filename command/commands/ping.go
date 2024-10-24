package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/theoreotm/gordinal/command"
)

type PingCommand struct {
	*command.BaseCommand
}

func NewPingCommand() *PingCommand {
	return &PingCommand{
		BaseCommand: command.NewBaseCommand("ping", "Ping the bot", true),
	}
}

func (c *PingCommand) Execute(s *discordgo.Session, m *command.Message) error {
	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Pong! %v", m.Args.PickStringRest()))
	return err
}

func (c *PingCommand) Name() string {
	return c.BaseCommand.Name
}

func (c *PingCommand) Description() string {
	return c.BaseCommand.Description
}

func (c *PingCommand) Help() string {
	return c.BaseCommand.Help
}

func (c *PingCommand) Active() bool {
	return c.BaseCommand.Active
}
