package bot

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"

	"github.com/TheOreoTM/friemon-go/internal/data"
	"github.com/TheOreoTM/friemon-go/internal/game"
)

// wildSpawns tracks the uncaught character in each channel. It lives in memory,
// so a restart clears it; a new spawn replaces the old one, like Pokétwo.
type wildSpawns struct {
	mu        sync.Mutex
	byChannel map[snowflake.ID]data.Character
}

func newWildSpawns() *wildSpawns {
	return &wildSpawns{byChannel: map[snowflake.ID]data.Character{}}
}

func (w *wildSpawns) set(channelID snowflake.ID, c data.Character) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.byChannel[channelID] = c
}

func (w *wildSpawns) get(channelID snowflake.ID) (data.Character, bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	c, ok := w.byChannel[channelID]
	return c, ok
}

// catchResult says what happened when someone guessed.
type catchResult int

const (
	noSpawn catchResult = iota
	wrongName
	caught
)

// take removes the channel's spawn if guess names it. Checking and removing
// under one lock means two players guessing at once cannot both catch it.
func (w *wildSpawns) take(channelID snowflake.ID, guess string) (data.Character, catchResult) {
	w.mu.Lock()
	defer w.mu.Unlock()
	c, ok := w.byChannel[channelID]
	if !ok {
		return data.Character{}, noSpawn
	}
	if !game.Matches(c, guess) {
		return data.Character{}, wrongName
	}
	delete(w.byChannel, channelID)
	return c, caught
}

// putBack restores a spawn after a failed save, unless a new one has appeared since.
func (w *wildSpawns) putBack(channelID snowflake.ID, c data.Character) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if _, ok := w.byChannel[channelID]; !ok {
		w.byChannel[channelID] = c
	}
}

// onMessage gives XP to the author's selected character and counts the message towards a spawn.
func (b *Bot) onMessage(e *events.GuildMessageCreate) {
	author := e.Message.Author
	if author.Bot || e.Message.WebhookID != nil {
		return
	}
	now := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if b.xpCooldown.Allow(author.ID.String(), now) {
		b.giveXP(ctx, e)
	}
	if b.counter.Record(e.GuildID.String(), author.ID.String(), now) {
		b.spawn(ctx, e)
	}
}

// spawn posts a random character in the server's spawn channel, if one is set.
func (b *Bot) spawn(ctx context.Context, e *events.GuildMessageCreate) {
	channelID, ok, err := b.store.SpawnChannel(ctx, e.GuildID)
	if err != nil {
		log.Printf("spawn channel for guild %s: %v", e.GuildID, err)
		return
	}
	if !ok {
		return // no spawn channel set with /setspawn yet
	}

	c := game.PickSpawn(newRand(), data.Characters)
	embed := discord.NewEmbed().
		WithTitle("A wild character has appeared!").
		WithDescriptionf("*%s*\n\nGuess who it is with `/catch`.", c.Description).
		WithColor(embedColor)
	if _, err := e.Client().Rest.CreateMessage(channelID, discord.MessageCreate{Embeds: []discord.Embed{embed}}); err != nil {
		log.Printf("post spawn in channel %s: %v", channelID, err) // often a missing permission
		return
	}
	b.wild.set(channelID, c)
}

func (b *Bot) handleCatch(e *handler.CommandEvent) error {
	channelID := e.Channel().ID()
	c, result := b.wild.take(channelID, e.SlashCommandInteractionData().String("name"))
	switch result {
	case noSpawn:
		return e.CreateMessage(ephemeral("There is no wild character here right now."))
	case wrongName:
		return e.CreateMessage(discord.MessageCreate{Content: "That is the wrong name!"})
	}

	rolled := game.RollCaught(newRand(), c)
	idx, err := b.store.AddCaught(e.Ctx, e.User().ID, rolled)
	if err != nil {
		b.wild.putBack(channelID, c) // let someone try again
		return err
	}
	return e.CreateMessage(discord.MessageCreate{Content: fmt.Sprintf(
		"Congratulations %s! You caught a level %d **%s**! It was added to your collection as #%d.",
		e.User().Mention(), rolled.Level, c.Name, idx,
	)})
}

func (b *Bot) handleHint(e *handler.CommandEvent) error {
	c, ok := b.wild.get(e.Channel().ID())
	if !ok {
		return e.CreateMessage(ephemeral("There is no wild character here right now."))
	}
	// "_" means italics in Discord, so escape it to show the blanks literally.
	hint := strings.ReplaceAll(game.Hint(newRand(), c.Name), "_", `\_`)
	return e.CreateMessage(discord.MessageCreate{Content: "The character is " + hint + "."})
}
