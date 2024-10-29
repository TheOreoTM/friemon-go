package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
	"github.com/peterbourgon/ff/v3"
	"github.com/theoreotm/gordinal/command"
)

func main() {
	fmt.Printf(`
	______         _   _
	|  ___|       | | (_)
	| |_ __ _  ___| |_ _  ___  _ __
	|  _/ _' |/ __| __| |/ _ \| '_ \
	| || (_| | (__| |_| | (_) | | | |
	\_| \__,_|\___|\__|_|\___/|_| |_| %s
`, "v0.0.1")

	fs := flag.NewFlagSet("faction", flag.ExitOnError)
	token := fs.String("token", "", "Discord Authentication Token")
	fs.IntVar(&lit.LogLevel, "log-level", 0, "LogLevel (0-3)")
	if err := ff.Parse(fs, os.Args[1:], ff.WithEnvVarPrefix("FACT")); err != nil {
		lit.Error("could not parse flags: %v", err)
		return
	}

	session, err := discordgo.New("Bot " + *token)
	if err != nil {
		fmt.Fprintf(fs.Output(), "Usage of %s:\n", fs.Name())
		fs.PrintDefaults()
		log.Println("You must provide a Discord authentication token.")
		return
	}

	session.Identify.Intents = discordgo.IntentsAllWithoutPrivileged |
		discordgo.IntentsGuildMembers |
		discordgo.IntentsGuildMessages

	session.AddHandler(command.OnInteractionCommand)
	session.AddHandler(command.OnAutocomplete)
	session.AddHandler(command.OnModalSubmit)

	if err := session.Open(); err != nil {
		log.Fatalf("error opening connection to Discord: %v", err)
	}
	defer session.Close()

	command.Register(session, session.State.User.ID)

	log.Println(`Now running. Press CTRL-C to exit.`)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, syscall.SIGTERM)
	<-sc

	log.Println("Removing commands...")
	command.Unregister(session, nil)

}
