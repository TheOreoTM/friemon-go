package handler

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type Argument struct {
	Message *discordgo.Message
	Value   any
	Type    string
}

type Args struct {
	Arguments []*Argument
}

func ParseArgs(session *discordgo.Session, message *discordgo.Message) *Args {
	args := Args{}
	args.Arguments = make([]*Argument, 0)

	parts := strings.Split(message.Content, " ")
	for i := 1; i < len(parts); i++ {
		fmt.Printf("Part: %v\n", parts[i])

		user := resolveUser(session, parts[i])
		if user != nil {
			args.Arguments = append(args.Arguments, &Argument{
				Message: message,
				Value:   user,
				Type:    "User",
			})
			fmt.Printf("User: %v\n", user)
			continue
		}

		guild, err := session.Guild(message.GuildID)
		if err != nil {
			return nil
		}

		member, err := resolveMember(session, guild, parts[i])
		if err != nil {
			return nil
		}

		if member != nil {
			args.Arguments = append(args.Arguments, &Argument{
				Message: message,
				Value:   member,
				Type:    "Member",
			})
			continue
		}

		args.Arguments = append(args.Arguments, &Argument{
			Message: message,
			Value:   parts[i],
			Type:    "String",
		})

	}

	return &args
}

func (a *Args) GetUser() *discordgo.User {
	for i, arg := range a.Arguments {
		if arg.Type == "User" {
			a.Arguments = append(a.Arguments[:i], a.Arguments[i+1:]...)
			return arg.Value.(*discordgo.User)
		}
	}

	return nil
}
