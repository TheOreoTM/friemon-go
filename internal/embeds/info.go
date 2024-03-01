package embeds

import (
	"github.com/andersfylling/disgord"
)

// Info instantiates an informational embed.
func Info(title, description, footer string, fields ...*disgord.EmbedField) *disgord.Embed {
	return &disgord.Embed{
		Title:       title,
		Description: description,
		Footer:      &disgord.EmbedFooter{Text: footer},
		Color:       0x297bd1,
		Fields:      fields,
	}
}

func InfoImage(title, description, footer string, url string, fields ...*disgord.EmbedField) *disgord.Embed {
	base := Info(title, description, footer, fields...)
	base.Image = &disgord.EmbedImage{URL: url}
	return base
}
