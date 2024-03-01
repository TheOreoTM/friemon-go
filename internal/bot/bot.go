package bot

import (
	"fmt"
	"os"

	"github.com/andersfylling/disgord"
	"github.com/sirupsen/logrus"
	"github.com/theoreotm/gommand"
	"github.com/theoreotm/gordinal/internal/commands"
	"github.com/theoreotm/gordinal/internal/embeds"
	"github.com/theoreotm/gordinal/internal/events"
)

func Start() {
	router := gommand.NewRouter(&gommand.RouterConfig{
		PrefixCheck: gommand.MultiplePrefixCheckers(gommand.StaticPrefix(">"), gommand.MentionPrefix),
	})

	client := disgord.New(disgord.Config{
		BotToken:    os.Getenv("DISCORD_TOKEN"),
		Intents:     disgord.AllIntents(),
		Logger:      logrus.New(),
		ProjectName: "Gordinal",
		Presence: &disgord.UpdateStatusPayload{
			Game: &disgord.Activity{
				Name: "with your feelings",
			},
		},
	})

	router.AddErrorHandler(func(ctx *gommand.Context, err error) bool {
		switch err.(type) {
		case *gommand.CommandNotFound, *gommand.CommandBlank:
			ctx.Reply(embeds.Error("Command Not Found", err, false))
			return true
		case *gommand.InvalidTransformation:
			ctx.Reply(embeds.Error("Invalid Type", err, false))
			return true
		case *gommand.IncorrectPermissions:
			ctx.Reply(embeds.Error("Missing Permissions", err, false))
			return true
		case *gommand.InvalidArgCount:
			ctx.Reply(embeds.Error("Missing Arguments", err, false))
			return true
		case *gommand.PanicError:
			ctx.Session.Logger().Error(err)
			ctx.Reply(embeds.Error("Panic", err, true))
			return false
		default:
			ctx.Session.Logger().Error(err)
			ctx.Reply(embeds.Error("Handled Error:", err, true))
			return false
		}
	})

	commands.Register(router)

	events.Register(client)

	router.Hook(client)

	defer client.Gateway().StayConnectedUntilInterrupted()

	client.Gateway().BotReady(func() {
		fmt.Println("Bot is ready!")
	})
}
