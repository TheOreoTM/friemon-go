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

// numberOption is the optional "number" option shared by several commands.
func numberOption(description string) discord.ApplicationCommandOptionInt {
	return discord.ApplicationCommandOptionInt{Name: "number", Description: description, MinValue: omit.Ptr(1)}
}

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
		Description: "Make a character your party leader",
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
			numberOption("Defaults to your party leader"),
		},
	},
	discord.SlashCommandCreate{
		Name:        "moves",
		Description: "Show a character's moves and what it can learn",
		Options: []discord.ApplicationCommandOption{
			numberOption("Defaults to your party leader"),
		},
	},
	discord.SlashCommandCreate{
		Name:        "learn",
		Description: "Teach a character a move",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{
				Name: "move", Description: "The move to learn", Required: true, Autocomplete: true,
			},
			numberOption("Defaults to your party leader"),
		},
	},
	discord.SlashCommandCreate{
		Name:        "party",
		Description: "Manage your party of up to 4: it battles together and shares chat XP",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionSubCommand{Name: "view", Description: "Show your party"},
			discord.ApplicationCommandOptionSubCommand{
				Name: "add", Description: "Add a character to your party",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionInt{Name: "number", Description: "Its number in /collection", Required: true, MinValue: omit.Ptr(1)},
				},
			},
			discord.ApplicationCommandOptionSubCommand{
				Name: "remove", Description: "Take a character out of your party",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionInt{Name: "number", Description: "Its number in /collection", Required: true, MinValue: omit.Ptr(1)},
				},
			},
		},
	},
	discord.SlashCommandCreate{
		Name:        "battle",
		Description: "Challenge another player to a party battle",
		Contexts:    guildOnly,
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionUser{Name: "opponent", Description: "Who to battle", Required: true},
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
	r.Route("/party", func(r handler.Router) {
		r.Command("/view", b.handlePartyView)
		r.Command("/add", b.handlePartyAdd)
		r.Command("/remove", b.handlePartyRemove)
	})
	r.Command("/info", b.handleInfo)
	r.Command("/moves", b.handleMoves)
	r.Command("/learn", b.handleLearn)
	r.Autocomplete("/learn", b.autocompleteLearn)
	// {user}, {idx} and {move} are variables, read in the handler from e.Vars.
	r.Component("/learn-replace/{user}/{idx}/{move}", b.handleLearnReplace)
	r.Command("/battle", b.handleBattle)
	r.Component("/battle-accept/{id}", b.handleBattleAccept)
	r.Component("/battle-decline/{id}", b.handleBattleDecline)
	r.Component("/battle-choose/{id}", b.handleBattleChoose)
	r.Component("/battle-act/{id}/{step}/{action}", b.handleBattleAct)
	r.Component("/battle-flee/{id}", b.handleBattleFlee)

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
