package events

import "github.com/andersfylling/disgord"

func Register(client *disgord.Client) {
	go client.Gateway().MessageCreate(func(s disgord.Session, h *disgord.MessageCreate) {
		if h.Message.Content == "hi" {
			s.SendMsg(h.Message.ChannelID, "hey")
		}
	})

}
