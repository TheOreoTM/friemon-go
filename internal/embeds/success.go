package embeds

import (
	"github.com/andersfylling/disgord"
)

// Info instantiates an success embed.
func Success(title, description, footer string, fields ...*disgord.EmbedField) *disgord.Embed {
	return &disgord.Embed{
		Title:       title,
		Description: description,
		Footer:      &disgord.EmbedFooter{Text: footer},
		Color:       0xf05050,
		Fields:      fields,
	}
}

func SuccessImage(title, description, footer string, url string, fields ...*disgord.EmbedField) *disgord.Embed {
	base := Info(title, description, footer, fields...)
	base.Image = &disgord.EmbedImage{URL: url}
	return base
}
