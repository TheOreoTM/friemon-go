// Package bot wires Discord events to game logic.
package bot

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/paginator"
	"github.com/disgoorg/snowflake/v2"

	"github.com/TheOreoTM/friemon-go/internal/game"
	"github.com/TheOreoTM/friemon-go/internal/store"
)

// embedColor is used on all the bot's embeds.
const embedColor = 0x7fb7a4

// Bot holds the Discord client and everything the handlers need.
type Bot struct {
	client     *bot.Client
	store      *store.Store
	guildIDs   []snowflake.ID // if set, commands register to these guilds only (instant updates while developing)
	counter    *game.SpawnCounter
	xpCooldown *game.Cooldown
	wild       *wildSpawns
	battles    *battles
	pages      *paginator.Manager
}

// New creates a bot. It does not connect yet; call Start for that.
// guildID may be empty to register commands globally.
func New(token, guildID string, st *store.Store) (*Bot, error) {
	b := &Bot{
		store:      st,
		counter:    game.NewSpawnCounter(game.DefaultSpawnConfig),
		xpCooldown: game.NewCooldown(game.XPCooldown),
		wild:       newWildSpawns(),
		battles:    newBattles(),
		pages:      paginator.New(paginator.WithEmbedColor(embedColor)),
	}
	if guildID != "" {
		id, err := snowflake.Parse(guildID)
		if err != nil {
			return nil, fmt.Errorf("parse guild ID: %w", err)
		}
		b.guildIDs = []snowflake.ID{id}
	}

	client, err := disgo.New(token,
		// GuildMessages lets us count messages. We never read their text,
		// so we do not need the privileged Message Content intent.
		bot.WithGatewayConfigOpts(gateway.WithIntents(gateway.IntentGuilds, gateway.IntentGuildMessages)),
		// Handle each event in its own goroutine, so a slow database call
		// for one player never holds up everyone else.
		bot.WithEventManagerConfigOpts(bot.WithAsyncEventsEnabled()),
		bot.WithEventListeners(b.newRouter(), b.pages),
		bot.WithEventListenerFunc(b.onMessage),
		bot.WithEventListenerFunc(func(e *events.Ready) {
			log.Printf("logged in as %s", e.User.Username)
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}
	b.client = client
	return b, nil
}

// Start registers slash commands and opens the gateway connection.
func (b *Bot) Start(ctx context.Context) error {
	// SyncCommands replaces the full command list, so removed commands disappear too.
	if err := handler.SyncCommands(b.client, commands, b.guildIDs); err != nil {
		return fmt.Errorf("register commands: %w", err)
	}
	if err := b.client.OpenGateway(ctx); err != nil {
		return fmt.Errorf("open gateway: %w", err)
	}
	// "go" runs a function in the background on its own goroutine.
	go b.expireBattles(ctx)
	log.Printf("%d command(s) registered", len(commands))
	return nil
}

// Close disconnects from Discord.
func (b *Bot) Close(ctx context.Context) {
	b.client.Close(ctx)
}

// newRand returns a fresh random generator. A *rand.Rand is not safe to share
// between goroutines, so each event makes its own; it is cheap.
func newRand() *rand.Rand {
	return rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
}
