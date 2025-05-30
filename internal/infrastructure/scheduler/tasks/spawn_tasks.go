package tasks

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
)

// TaskData represents flexible task data (copied from scheduler to avoid import cycle)
type TaskData map[string]interface{}

// Utility methods for TaskData
func (td TaskData) MustString(key string) string {
	val, ok := td[key]
	if !ok {
		panic(fmt.Sprintf("key %s not found or not string", key))
	}
	str, ok := val.(string)
	if !ok {
		panic(fmt.Sprintf("key %s is not a string", key))
	}
	return str
}

// Dependencies interface to avoid import cycles
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

// SpawnTaskHandlers contains handlers for spawn-related scheduled tasks
type SpawnTaskHandlers struct {
	deps BotDependencies
}

func NewSpawnTaskHandlers(deps BotDependencies) *SpawnTaskHandlers {
	return &SpawnTaskHandlers{deps: deps}
}

func (h *SpawnTaskHandlers) DisableSpawnButton(ctx context.Context, data TaskData) error {
	channelID := snowflake.MustParse(data.MustString("channel_id"))
	messageID := snowflake.MustParse(data.MustString("message_id"))

	slog.Info("Disabling spawn button",
		slog.String("channel_id", channelID.String()),
		slog.String("message_id", messageID.String()))

	// Get the message
	message, err := h.deps.GetRestClient().GetMessage(channelID, messageID)
	if err != nil {
		slog.Error("Failed to get message for spawn button disable", slog.Any("err", err))
		return err
	}

	// Update button to disabled state
	button, exists := message.ButtonByID("/claim")
	if !exists {
		slog.Error("Failed to find button to disable")
		return nil
	}

	_, err = h.deps.GetRestClient().UpdateMessage(
		channelID,
		messageID,
		discord.NewMessageUpdateBuilder().
			AddActionRow(button.AsDisabled().WithLabel("Character has wandered away")).
			Build())

	if err != nil {
		slog.Error("Failed to disable spawn button", slog.Any("err", err))
		return err
	}

	// Clean up cached character
	h.deps.GetCache().DeleteChannelCharacter(channelID)

	slog.Info("Successfully disabled spawn button")
	return nil
}

func (h *SpawnTaskHandlers) CleanupChannel(ctx context.Context, data TaskData) error {
	channelID := snowflake.MustParse(data.MustString("channel_id"))

	slog.Info("Starting channel cleanup", slog.String("channel_id", channelID.String()))

	// Reset interaction count and clean up character data
	h.deps.GetCache().ResetInteractionCount(channelID)
	h.deps.GetCache().DeleteChannelCharacter(channelID)

	slog.Info("Channel cleanup completed")
	return nil
}
