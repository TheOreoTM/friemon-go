package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
	"github.com/theoreotm/friemon/friemon"
	"github.com/theoreotm/friemon/friemon/db"
)

const (
	registrationDuration = 1 * time.Minute
	questionDuration     = 30 * time.Second
	endStateDuration     = 1 * time.Minute
)

type TriviaHandler struct {
	bot   *friemon.Bot
	redis *db.RedisClient
	ctx   context.Context
}

func NewTriviaHandler(bot *friemon.Bot, redis *db.RedisClient) *TriviaHandler {
	return &TriviaHandler{
		bot:   bot,
		redis: redis,
		ctx:   context.Background(),
	}
}

func (h *TriviaHandler) StartTrivia(e *events.ApplicationCommandInteractionCreate) error {
	// Check if trivia is already running
	state, err := h.redis.GetTriviaState(h.ctx)
	if err == nil && state != "" {
		return e.CreateMessage(discord.MessageCreate{
			Content: "A trivia game is already in progress!",
		})
	}

	// Initialize trivia state
	err = h.redis.SetTriviaState(h.ctx, db.TriviaStateStarting)
	if err != nil {
		return fmt.Errorf("failed to set trivia state: %v", err)
	}

	// Clean up any previous game data
	err = h.redis.CleanupTrivia(h.ctx)
	if err != nil {
		return fmt.Errorf("failed to cleanup previous trivia: %v", err)
	}

	// Set up questions (you'll need to implement question generation/loading)
	questions := generateQuestions() // Implement this function
	err = h.redis.SetTriviaQuestions(h.ctx, questions)
	if err != nil {
		return fmt.Errorf("failed to set trivia questions: %v", err)
	}

	// Move to waiting state and send registration message
	err = h.redis.SetTriviaState(h.ctx, db.TriviaStateWaiting)
	if err != nil {
		return fmt.Errorf("failed to set trivia state: %v", err)
	}

	// Create registration embed
	embed := discord.NewEmbedBuilder().
		SetTitle("Trivia Game Starting!").
		SetDescription("React to this message to join the game!").
		SetColor(0x00ff00).
		SetFooter("Registration closes in 1 minute", "").
		Build()

	// Send registration message
	err = e.CreateMessage(discord.MessageCreate{
		Embeds: []discord.Embed{embed},
	})
	if err != nil {
		return fmt.Errorf("failed to send registration message: %v", err)
	}

	// Start registration timer
	time.AfterFunc(registrationDuration, func() {
		h.startGame(e.ChannelID())
	})

	return nil
}

func (h *TriviaHandler) startGame(channelID snowflake.ID) {
	// Check if we have enough players
	users, err := h.redis.GetRegisteredTriviaUsers(h.ctx)
	if err != nil || len(users) == 0 {
		h.redis.CleanupTrivia(h.ctx)
		h.bot.Client.Rest().CreateMessage(channelID, discord.MessageCreate{
			Content: "Not enough players registered for trivia. Game cancelled.",
		})
		return
	}

	// Set game state to in progress
	err = h.redis.SetTriviaState(h.ctx, db.TriviaStateInProgress)
	if err != nil {
		h.bot.Client.Rest().CreateMessage(channelID, discord.MessageCreate{
			Content: "Error starting trivia game.",
		})
		return
	}

	// Start asking questions
	h.askNextQuestion(channelID)
}

func (h *TriviaHandler) askNextQuestion(channelID snowflake.ID) {
	// Get next question
	question, err := h.redis.GetNextTriviaQuestion(h.ctx)
	if err != nil {
		// No more questions, end the game
		h.endGame(channelID)
		return
	}

	// Create question embed
	embed := discord.NewEmbedBuilder().
		SetTitle("Trivia Question").
		SetDescription(question).
		SetColor(0x0000ff).
		SetFooter("You have 30 seconds to answer!", "").
		Build()

	// Send question
	h.bot.Client.Rest().CreateMessage(channelID, discord.MessageCreate{
		Embeds: []discord.Embed{embed},
	})

	// Wait for the question duration
	time.AfterFunc(questionDuration, func() {
		h.askNextQuestion(channelID)
	})
}

func (h *TriviaHandler) endGame(channelID snowflake.ID) {
	// Set game state to ended
	err := h.redis.SetTriviaState(h.ctx, db.TriviaStateEnded)
	if err != nil {
		h.bot.Client.Rest().CreateMessage(channelID, discord.MessageCreate{
			Content: "Error ending trivia game.",
		})
		return
	}

	// Get final scores
	scores, err := h.redis.GetTriviaScores(h.ctx)
	if err != nil {
		h.bot.Client.Rest().CreateMessage(channelID, discord.MessageCreate{
			Content: "Error retrieving final scores.",
		})
		return
	}

	// Create results embed
	embed := discord.NewEmbedBuilder().
		SetTitle("Trivia Game Results").
		SetColor(0xff0000)

	// Add scores to embed
	description := "Final Scores:\n"
	for _, score := range scores {
		description += fmt.Sprintf("<@%s>: %.0f points\n", score.Member, score.Score)
	}
	embed.SetDescription(description)

	// Send results
	h.bot.Client.Rest().CreateMessage(channelID, discord.MessageCreate{
		Embeds: []discord.Embed{embed.Build()},
	})

	// Clean up after delay
	time.AfterFunc(endStateDuration, func() {
		h.redis.CleanupTrivia(h.ctx)
	})
}

// Helper function to generate questions (implement this based on your needs)
func generateQuestions() []string {
	// This is a placeholder - implement your own question generation logic
	return []string{
		"What is the capital of France?",
		"Who wrote Romeo and Juliet?",
		"What is the chemical symbol for gold?",
		// Add more questions...
	}
}
