package main

import (
	"fmt"
	"os"

	"github.com/Karitham/corde"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/theoreotm/gordinal/discord"
	"github.com/urfave/cli/v2"
)

func main() {
	log.Logger = log.Level(zerolog.DebugLevel)

	disc := &discordCmd{}

	app := &cli.App{
		Name:        "idk",
		Usage:       "Run the bot, and use utils",
		Version:     "0.0.1",
		Description: "A bot for discord",
		Commands: []*cli.Command{
			{
				Name:    "register",
				Aliases: []string{"r"},
				Usage:   "Register the bot commands",
				Action:  disc.register,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "BOT_TOKEN",
						EnvVars:     []string{"DISCORD_TOKEN", "BOT_TOKEN"},
						Destination: &disc.botToken,
						Required:    true,
					},
					&cliSnowflake{
						EnvVars: []string{"DISCORD_GUILD_ID", "GUILD_ID"},
						Dest:    disc.guildID,
					},
					&cli.StringFlag{
						EnvVars:     []string{"DISCORD_APP_ID", "APP_ID"},
						Destination: &disc.appID,
						Required:    true,
					},
				},
			},
		},
	}
}

type discordCmd struct {
	botToken string
	appID    string
	guildID  *corde.Snowflake
}

func (dc *discordCmd) register(c *cli.Context) error {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	bot := &discord.Bot{
		AppID:    corde.SnowflakeFromString(dc.appID),
		BotToken: dc.botToken,
		GuildID:  dc.guildID,
	}

	if err := bot.RegisterCommands(); err != nil {
		return fmt.Errorf("error registering commands %v", err)
	}
	return nil
}

func (dc *discordCmd) run(c *cli.Context) error {
	disc := discord.New(&discord.Bot{
		AppID:    corde.SnowflakeFromString(dc.appID),
		GuildID:  dc.guildID,
		BotToken: dc.botToken,
	})

	return disc.ListenAndServe(":" + dc.port)
}
