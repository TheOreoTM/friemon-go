package command

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/lit"
)

const (
	// Version is a constant that stores the dgobot version information.
	Version                  = "v0.1.0-rewrite"
	RESPONSE_FLAGS_EPHEMERAL = 64
)

type Color = int

// Colors
const (
	ColorSuccess = 0x46b485
	ColorError   = 0xf05050
	ColorWarn    = 0xfee65c
	ColorInfo    = 0x297bd1
	ColorLoading = 0x23272a
	ColorDefault = 0x2b2d31
)

var (
	AdminUserID  string
	HerderRoleID string
	Commands     = make(map[string]*Command)
)

type Command struct {
	*discordgo.ApplicationCommand
	Handler      func(*discordgo.Session, *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error)
	Autocomplete func(*discordgo.Session, *discordgo.InteractionCreate) ([]*discordgo.ApplicationCommandOptionChoice, error)
}

func OnAutocomplete(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	if interaction.Type != discordgo.InteractionApplicationCommandAutocomplete {
		return
	}

	data := interaction.ApplicationCommandData()
	cmd, ok := Commands[data.Name]
	if !ok || cmd.Autocomplete == nil {
		return
	}

	choices, err := cmd.Autocomplete(session, interaction)
	if err != nil {
		return
	}

	err = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{
			Choices: choices,
		},
	})
	if err != nil {
		lit.Error("responding to autocomplete: %v", err)
	}
}

func Register(s *discordgo.Session, appid string) error {
	cmds := make([]*discordgo.ApplicationCommand, 0, len(Commands))
	for _, cmd := range Commands {
		fmt.Println("Registering", cmd.Name)
		cmds = append(cmds, cmd.ApplicationCommand)
	}
	_, err := s.ApplicationCommandBulkOverwrite(appid, "1138806085352951950", cmds)
	return err
}

func Unregister(s *discordgo.Session, guildID *string) {
	cmds, err := s.ApplicationCommands("1109691974648332378", "1138806085352951950")
	if err != nil {
		lit.Error(err.Error())
		return
	}
	for _, command := range cmds {
		err := s.ApplicationCommandDelete("1109691974648332378", "1138806085352951950", command.ID)
		if err != nil {
			lit.Error("Cannot delete '%v' command: %v", command.Name, err)
		}
	}
}

func OnInteractionCommand(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	if interaction.Type != discordgo.InteractionApplicationCommand {
		return
	}

	data := interaction.ApplicationCommandData()
	cmd, ok := Commands[data.Name]
	if !ok {
		return
	}

	res, err := cmd.Handler(session, interaction)
	if err != nil {
		res = &discordgo.InteractionResponseData{
			Flags: RESPONSE_FLAGS_EPHEMERAL,
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Error",
					Description: err.Error(),
					Color:       ColorError,
				},
			},
		}
	}

	typ := discordgo.InteractionResponseChannelMessageWithSource
	if res.Title != "" {
		typ = discordgo.InteractionResponseModal
	}
	err = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: typ,
		Data: res,
	})
	if err != nil {
		lit.Error("responding to interaction %s: %v", data.Name, err)
	}
}

// OnModalSubmit routes modal submit interactions to the appropriate handler.
// it uses `prefix:` from the custom ID to determine which handler to use.
func OnModalSubmit(ds *discordgo.Session, ic *discordgo.InteractionCreate) {
	if ic.Type != discordgo.InteractionModalSubmit {
		return
	}

	data := ic.ModalSubmitData()
	prefix, _, ok := strings.Cut(data.CustomID, ":")
	if !ok {
		lit.Error("Invalid custom ID: %s", data.CustomID)
		EphemeralResponse("Invalid modal submit.")
		return
	}

	cmd := Commands[prefix]
	res, err := cmd.Handler(ds, ic)
	if err != nil {
		res = &discordgo.InteractionResponseData{
			Flags: RESPONSE_FLAGS_EPHEMERAL,
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Error",
					Description: err.Error(),
					Color:       ColorError,
				},
			},
		}
	}

	err = ds.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: res,
	})
	if err != nil {
		lit.Error("responding to modal submit %s: %v", data.CustomID, err)
	}
}

func ContentResponse(c string) *discordgo.InteractionResponseData {
	return &discordgo.InteractionResponseData{
		Content: c,
	}
}

func EphemeralResponse(c string) *discordgo.InteractionResponseData {
	return &discordgo.InteractionResponseData{
		Flags:   RESPONSE_FLAGS_EPHEMERAL,
		Content: c,
	}
}

func SimpleEmbedResponse(title, desc string, rawColor *Color) *discordgo.InteractionResponseData {
	color := ColorDefault

	if rawColor != nil {
		color = *rawColor
	}

	return &discordgo.InteractionResponseData{
		Embeds: []*discordgo.MessageEmbed{
			{
				Color:       color,
				Title:       title,
				Description: desc,
			},
		},
	}
}

func EmbedResponse(e discordgo.MessageEmbed) *discordgo.InteractionResponseData {
	return &discordgo.InteractionResponseData{
		Embeds: []*discordgo.MessageEmbed{&e},
	}
}

func FileResponse(f discordgo.File) *discordgo.InteractionResponseData {
	return &discordgo.InteractionResponseData{
		Files: []*discordgo.File{&f},
	}
}

func Autocomplete(options ...string) []*discordgo.ApplicationCommandOptionChoice {
	if len(options) > 25 {
		options = options[:25]
	}

	var choices []*discordgo.ApplicationCommandOptionChoice
	for _, opt := range options {
		if len(opt) > 100 {
			opt = opt[:100]
		}

		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
			Name:  opt,
			Value: opt,
		})
	}
	return choices
}
