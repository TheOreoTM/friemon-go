package handler

import (
	"fmt"
	"strings"
	"sync"

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

	var wg sync.WaitGroup
	var mu sync.Mutex

	for i := 1; i < len(parts); i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			user := resolveUser(session, parts[i])
			if user != nil {
				mu.Lock()
				args.Arguments = append(args.Arguments, &Argument{
					Message: message,
					Value:   user,
					Type:    "User",
				})
				mu.Unlock()
				return
			}

			guild, err := session.State.Guild(message.GuildID)
			if err != nil {
				return
			}

			member, err := resolveMember(session, guild, parts[i])
			if err != nil {
				return
			}

			if member != nil {
				mu.Lock()
				args.Arguments = append(args.Arguments, &Argument{
					Message: message,
					Value:   member,
					Type:    "Member",
				})
				mu.Unlock()
				return
			}

			mu.Lock()
			args.Arguments = append(args.Arguments, &Argument{
				Message: message,
				Value:   parts[i],
				Type:    "String",
			})
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	return &args
}

func (a *Args) GetUser() *discordgo.User {
	for i, arg := range a.Arguments {
		if arg.Type == "User" {
			removeArgument(a.Arguments, i)
			return arg.Value.(*discordgo.User)
		}
	}

	return nil
}

func (a *Args) GetMember() *discordgo.Member {
	for i, arg := range a.Arguments {
		if arg.Type == "Member" {
			removeArgument(a.Arguments, i)
			fmt.Printf("Member: %v\n", arg.Value.(*discordgo.Member).User.Username)
			return arg.Value.(*discordgo.Member)
		}
	}

	return nil
}

func removeArgument(args []*Argument, index int) []*Argument {
	return append(args[:index], args[index+1:]...)
}
