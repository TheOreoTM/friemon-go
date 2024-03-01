package commands

import (
	"github.com/theoreotm/gommand"
	"github.com/theoreotm/gordinal/internal/embeds"
)

func init() {
	cmds = append(cmds, &gommand.Command{
		Name:        "ping",
		Aliases:     []string{"latency", "pong"},
		Description: "Gets the bot's heartbeat (websocket) latency.",
		Category:    infoCategory,
		Function:    ping,
	})
}

func ping(ctx *gommand.Context) error {
	latency, _ := ctx.Session.AvgHeartbeatLatency()

	_, err := ctx.Reply(embeds.Info(
		"🏓 "+latency.String(),
		"",
		"",
	))
	return err
}
