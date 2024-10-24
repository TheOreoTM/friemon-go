package command

import (
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
	"github.com/theoreotm/gordinal/common"
)

const (
	Version       = "v0.1.0-go-rewrite"
	Prefix        = "!"
	ephemeralFlag = 64
)

var (
	Commands = make(map[string]ICommand)
)

type Message struct {
	*discordgo.MessageCreate
	Args *Arguments
}

type ICommand interface {
	Active() bool
	Help() string
	Description() string
	Name() string
	Execute(*discordgo.Session, *Message) error
}

type BaseCommand struct {
	Active      bool
	Name        string
	Description string
	Help        string
}

type Command struct {
	Prefix   string
	Session  *discordgo.Session
	Commands map[string]ICommand
}

func New(session *discordgo.Session) (*Command, error) {
	if len(Prefix) == 0 {
		return nil, common.ErrMissingDiscordBotPrefix
	}

	return &Command{
		Prefix:   Prefix,
		Session:  session,
		Commands: make(map[string]ICommand),
	}, nil
}

func NewBaseCommand(name, description string, active bool) *BaseCommand {
	return &BaseCommand{
		Name:        strings.ToLower(name),
		Description: description,
		Help:        "",
		Active:      active,
	}
}

func (c *Command) Register(commands ...ICommand) {
	for _, command := range commands {
		log.Printf("Registering command %s", command.Name())
		Commands[command.Name()] = command
	}
}

func Process(s *discordgo.Session, m *discordgo.MessageCreate) error {
	content := strings.Split(strings.TrimPrefix(m.Content, Prefix), " ")
	maybeCommandName, maybeArguments := content[0], content[1:]

	command, exists := Commands[strings.ToLower(maybeCommandName)]
	if !exists {
		return common.ErrCommandNotFound
	}

	if err := command.Execute(s, &Message{m, NewArguments(maybeArguments)}); err != nil {
		return err
	}

	return nil
}

func OnMessageCreateHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot || !strings.HasPrefix(m.Content, Prefix) || len(m.Content) == 0 {
		return
	}

	if err := Process(s, m); err != nil {
		lit.Error("command runtime exception: %v", err)
	}
}
