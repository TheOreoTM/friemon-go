package common

import "fmt"

var (
	ErrMissingDiscordBotPrefix = fmt.Errorf("missing discord bot prefix")
	ErrMissingDiscordBotToken  = fmt.Errorf("missing discord bot token")
	ErrCommandNotFound         = fmt.Errorf("command not found")
)
