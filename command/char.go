package command

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/theoreotm/gordinal/model"
)

func init() {
	Commands[cmdChar.Name] = &Command{
		ApplicationCommand: cmdChar,
		Handler:            handleChar,
	}
}

var cmdChar = &discordgo.ApplicationCommand{
	Type:        discordgo.ChatApplicationCommand,
	Name:        "char",
	Description: "Get a random character",
}

func handleChar(_ *discordgo.Session, ic *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
	randomCharacter := model.NewCharacter(ic.Member.User.ID)

	return ContentResponse(fmt.Sprintf("IV percent %.2f for %v", randomCharacter.IvPercentage()*100, randomCharacter.IvTotal)), nil
}
