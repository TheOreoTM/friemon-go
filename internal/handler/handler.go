package handler

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

type Command struct {
	Meta         *discordgo.ApplicationCommand
	ChatInputRun func(s *discordgo.Session, i *discordgo.InteractionCreate)
	MessageRun   func(s *discordgo.Session, m *discordgo.MessageCreate, args *Args)
}

type Handler struct {
	options  *SetupOptions
	Session  *discordgo.Session
	Commands map[string]*Command
}

type SetupOptions struct {
	Prefix        string
	CaseSensitive bool
}

var (
	Prefix        string
	CaseSensitive bool
)

func Setup(s *discordgo.Session, cmds map[string]*Command, options *SetupOptions) *Handler {
	if options == nil {
		panic("Setup options cannot be nil")
	} else {
		if options.Prefix == "" {
			options.Prefix = "!"
		}
	}

	handler := &Handler{
		Session:  s,
		options:  options,
		Commands: cmds,
	}

	Prefix = options.Prefix
	CaseSensitive = options.CaseSensitive

	return handler
}

func (h *Handler) LoadCommands() {
	if h.Session == nil {
		panic("Session cannot be nil")
	}

	fmt.Printf("Loading %v commands\n", len(h.Commands))

	h.Session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		handleMessageCommands(s, m, h.Commands)
	})

	h.Session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		handleChatInputCommands(s, i, h.Commands)
	})

}

func handleMessageCommands(s *discordgo.Session, m *discordgo.MessageCreate, cmds map[string]*Command) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	hasPrefix := strings.HasPrefix(m.Content, Prefix)
	if !hasPrefix {
		return
	}

	content := m.Content

	command := strings.TrimSpace(strings.TrimPrefix(content, Prefix))
	commandName := strings.Split(command, " ")[0]

	startTime := time.Now()
	args := ParseArgs(s, m.Message)
	if args == nil {
		args = &Args{}
	}
	endTime := time.Now()

	fmt.Printf("Parsed args in %v\n", endTime.Sub(startTime))

	if h, ok := cmds[commandName]; ok {
		go h.MessageRun(s, m, args)
	}
}

func handleChatInputCommands(s *discordgo.Session, i *discordgo.InteractionCreate, cmds map[string]*Command) {
	if h, ok := cmds[i.ApplicationCommandData().Name]; ok {
		go h.ChatInputRun(s, i)
	}
}
