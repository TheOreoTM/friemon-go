package tasks

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
	"github.com/theoreotm/friemon/internal/types"
)

type BotDependencies interface {
	GetRestClient() RestClient
	GetCache() CacheClient
}

type RestClient interface {
	GetMessage(channelID, messageID snowflake.ID, opts ...rest.RequestOpt) (*discord.Message, error)
	UpdateMessage(channelID, messageID snowflake.ID, messageUpdate discord.MessageUpdate, opts ...rest.RequestOpt) (*discord.Message, error)
}

type CacheClient interface {
	DeleteChannelCharacter(channelID snowflake.ID) error
	ResetInteractionCount(channelID snowflake.ID) error
}

type SpawnTaskHandlers struct {
	deps BotDependencies
}

func NewSpawnTaskHandlers(deps BotDependencies) *SpawnTaskHandlers {
	return &SpawnTaskHandlers{
		deps: deps,
	}
}

func (h *SpawnTaskHandlers) DisableSpawnButton(ctx context.Context, data types.TaskData) error {
	channelIDStr := data.MustString("channel_id")
	messageIDStr := data.MustString("message_id")

	channelID, err := snowflake.Parse(channelIDStr)
	if err != nil {
		return fmt.Errorf("invalid channel_id: %w", err)
	}

	messageID, err := snowflake.Parse(messageIDStr)
	if err != nil {
		return fmt.Errorf("invalid message_id: %w", err)
	}

	slog.Info("Disabling spawn button",
		"channel_id", channelID,
		"message_id", messageID,
	)

	// Get the current message
	restClient := h.deps.GetRestClient()
	message, err := restClient.GetMessage(channelID, messageID)
	if err != nil {
		return fmt.Errorf("failed to get message: %w", err)
	}

	button, exists := message.ButtonByID("/claim")
	if !exists {
		return errors.New("failed to find button")
	}

	if button.Disabled {
		return nil
	}

	_, err = restClient.UpdateMessage(
		channelID,
		messageID,
		discord.NewMessageUpdateBuilder().
			AddActionRow(button.AsDisabled().WithLabel("Character has wandered away")).
			Build())

	if err != nil {
		return fmt.Errorf("failed to update message: %w", err)
	}

	slog.Info("Successfully disabled spawn button",
		"channel_id", channelID,
		"message_id", messageID,
	)

	return nil
}

func (h *SpawnTaskHandlers) CleanupChannel(ctx context.Context, data types.TaskData) error {
	channelIDStr := data.MustString("channel_id")

	channelID, err := snowflake.Parse(channelIDStr)
	if err != nil {
		return fmt.Errorf("invalid channel_id: %w", err)
	}

	slog.Debug("Cleaning up channel cache", "channel_id", channelID)

	cache := h.deps.GetCache()

	// Clean up cached character
	if err := cache.DeleteChannelCharacter(channelID); err != nil {
		slog.Warn("Failed to delete cached character",
			"channel_id", channelID,
			"error", err,
		)
	}

	// Reset interaction count
	if err := cache.ResetInteractionCount(channelID); err != nil {
		slog.Warn("Failed to reset interaction count",
			"channel_id", channelID,
			"error", err,
		)
	}

	slog.Debug("Channel cleanup completed", "channel_id", channelID)
	return nil
}
