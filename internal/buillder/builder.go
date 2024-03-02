package builder

import "github.com/bwmarrin/discordgo"

type OptionBuilder struct {
	option *discordgo.ApplicationCommandOption
}

func NewOptionBuilder() *OptionBuilder {
	return &OptionBuilder{
		option: &discordgo.ApplicationCommandOption{},
	}
}

func (b *OptionBuilder) SetType(t discordgo.ApplicationCommandOptionType) *OptionBuilder {
	b.option.Type = t
	return b
}

func (b *OptionBuilder) SetName(name string) *OptionBuilder {
	b.option.Name = name
	return b
}

func (b *OptionBuilder) SetDescription(description string) *OptionBuilder {
	b.option.Description = description
	return b
}

func (b *OptionBuilder) SetRequired(required bool) *OptionBuilder {
	b.option.Required = required
	return b
}

func (b *OptionBuilder) Build() *discordgo.ApplicationCommandOption {
	return b.option
}
