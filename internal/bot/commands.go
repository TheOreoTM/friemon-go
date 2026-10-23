package bot

import (
	"fmt"
	"log"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/omit"
)

// guildOnly limits a command to servers (no DMs).
var guildOnly = []discord.InteractionContextType{discord.InteractionContextTypeGuild}

// commands is the list registered with Discord on startup.
var commands = []discord.ApplicationCommandCreate{
	discord.SlashCommandCreate{Name: "ping", Description: "Check that the bot is alive"},
	discord.SlashCommandCreate{
		Name:        "catch",
		Description: "Catch the wild character by guessing its name",
		Contexts:    guildOnly,
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{Name: "name", Description: "Who is it?", Required: true},
		},
	},
	discord.SlashCommandCreate{
		Name:        "hint",
		Description: "Reveal some letters of the wild character's name",
		Contexts:    guildOnly,
	},
	discord.SlashCommandCreate{
		Name:        "collection",
		Description: "List the characters you have caught",
	},
	discord.SlashCommandCreate{
		Name:        "select",
		Description: "Choose the character that gains XP when you chat",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionInt{
				Name: "number", Description: "Its number in /collection", Required: true, MinValue: omit.Ptr(1),
			},
		},
	},
	discord.SlashCommandCreate{
		Name:        "info",
		Description: "Show a character's stats and moves",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionInt{
				Name: "number", Description: "Defaults to your selected character", MinValue: omit.Ptr(1),
			},
		},
	},
	discord.SlashCommandCreate{
		Name:        "moves",
		Description: "Show your selected character's moves and what it can learn",
	},
	discord.SlashCommandCreate{
		Name:        "learn",
		Description: "Teach your selected character a move",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{
				Name: "move", Description: "The move to learn", Required: true, Autocomplete: true,
			},
		},
	},
	discord.SlashCommandCreate{
		Name:        "setspawn",
		Description: "Set the channel where characters spawn",
		Contexts:    guildOnly,
		// Only members with Manage Server see this command by default.
		DefaultMemberPermissions: omit.NewPtr(discord.PermissionManageGuild),
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionChannel{
				Name:         "channel",
				Description:  "Defaults to this channel",
				ChannelTypes: []discord.ChannelType{discord.ChannelTypeGuildText},
			},
		},
	},
}

// newRouter maps each command to the method that answers it.
// Paths look like "/name", or "/name/sub" for subcommands.
func (b *Bot) newRouter() *handler.Mux {
	r := handler.New()
	r.Command("/ping", handlePing)
	r.Command("/catch", b.handleCatch)
	r.Command("/hint", b.handleHint)
	r.Command("/collection", b.handleCollection)
	r.Command("/setspawn", b.handleSetSpawn)
	r.Command("/select", b.handleSelect)
	r.Command("/info", b.handleInfo)
	r.Command("/moves", b.handleMoves)
	r.Command("/learn", b.handleLearn)
	r.Autocomplete("/learn", b.autocompleteLearn)
	// {user}, {idx} and {move} are variables, read in the handler from e.Vars.
	r.Component("/learn-replace/{user}/{idx}/{move}", b.handleLearnReplace)

	// When a handler returns an error, log it and tell the player something went wrong.
	r.Error(func(e *handler.InteractionEvent, err error) {
		log.Printf("command %s failed: %v", e.Interaction.ID(), err)
		_ = e.CreateMessage(ephemeral("Something went wrong. Please try again.")) // fails if we already replied; that is fine
	})
	return r
}

func handlePing(e *handler.CommandEvent) error {
	latency := e.Client().Gateway.Latency()
	return e.CreateMessage(discord.MessageCreate{
		Content: fmt.Sprintf("Pong! Gateway latency: %dms", latency.Milliseconds()),
	})
}

func (b *Bot) handleSetSpawn(e *handler.CommandEvent) error {
	guildID := e.GuildID()
	if guildID == nil {
		return e.CreateMessage(ephemeral("This only works in a server."))
	}
	channelID := e.Channel().ID()
	if ch, ok := e.SlashCommandInteractionData().OptChannel("channel"); ok {
		channelID = ch.ID
	}
	if err := b.store.SetSpawnChannel(e.Ctx, *guildID, channelID); err != nil {
		return err
	}
	return e.CreateMessage(discord.MessageCreate{Content: fmt.Sprintf("Characters will now spawn in <#%s>.", channelID)})
}

// ephemeral builds a reply only the command's user can see.
func ephemeral(content string) discord.MessageCreate {
	return discord.MessageCreate{Content: content, Flags: discord.MessageFlagEphemeral}
}
