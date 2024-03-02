package bot

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	"github.com/theoreotm/gordinal/internal/commands"
	"github.com/theoreotm/gordinal/internal/handler"
)

var (
	GuildID        = flag.String("guild", "", "Test guild ID. If not passed - bot registers commands globally")
	BotToken       = flag.String("token", "", "Bot access token")
	RemoveCommands = flag.Bool("rmcmd", false, "Remove all commands after shutdowning or not")
	Prefix         = flag.String("prefix", "!", "Bot command prefix")
	CaseSensitive  = flag.Bool("case", false, "Bot command case sensitivity")
)

var s *discordgo.Session

func init() {
	flag.Parse()
}

func init() {
	var err error
	s, err = discordgo.New("Bot " + *BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
}

func Start() {
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)

		log.Println("Adding commands...")
		cmds, _ := commands.Register(s, GuildID)

		h := handler.Setup(s, cmds, &handler.SetupOptions{
			Prefix:        *Prefix,
			CaseSensitive: *CaseSensitive,
		})

		h.LoadCommands()
	})

	err := s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	defer s.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop

	log.Println("Gracefully shutting down.")

	if *RemoveCommands {
		log.Println("Removing commands...")
		commands.Unregister(s, GuildID)
	}
}
