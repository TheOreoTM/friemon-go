package command

import (
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type Arguments struct {
	messageContent []string // Message split into content parts

	PickUser       func() *discordgo.User
	PickMember     func() *discordgo.Member
	PickChannel    func() *discordgo.Channel
	PickString     func() string
	PickStringRest func() string
	PickInt        func() int
}

func NewArguments(parts []string) *Arguments {
	return &Arguments{
		messageContent: parts,
		PickUser:       func() *discordgo.User { return pickUser(&parts) },
		PickMember:     func() *discordgo.Member { return pickMember(&parts) },
		PickChannel:    func() *discordgo.Channel { return pickChannel(&parts) },
		PickString:     func() string { return pickString(&parts) },
		PickStringRest: func() string { return pickStringRest(&parts) },
		PickInt:        func() int { return pickInt(&parts) },
	}
}

func pickMember(parts *[]string) *discordgo.Member {
	for i, part := range *parts {
		// Assuming the member is mentioned like <@!MemberID>
		if isUserMention(part) {
			member := extractMemberFromMention(part)
			*parts = append((*parts)[:i], (*parts)[i+1:]...) // Remove the parsed part
			return member
		}
	}
	return nil // No valid member found
}

func pickUser(parts *[]string) *discordgo.User {
	for i, part := range *parts {
		// Assuming the user is mentioned like <@UserID>
		if isUserMention(part) {
			user := extractUserFromMention(part)
			*parts = append((*parts)[:i], (*parts)[i+1:]...) // Remove the parsed part
			return user
		}
	}
	return nil // No valid user found
}

func isUserMention(s string) bool {
	// Check if the part is a valid user mention
	return strings.HasPrefix(s, "<@") && strings.HasSuffix(s, ">")
}

func extractUserFromMention(s string) *discordgo.User {
	// Extract user ID from the mention and return the user (you can use discordgo's method to fetch the user)
	userID := strings.Trim(s, "<@>")
	return &discordgo.User{ID: userID}
}

func extractMemberFromMention(s string) *discordgo.Member {
	// Extract user ID from the mention and return the member (you can use discordgo's method to fetch the member)
	userID := strings.Trim(s, "<@!")
	return &discordgo.Member{User: &discordgo.User{ID: userID}}
}

func pickChannel(parts *[]string) *discordgo.Channel {
	for i, part := range *parts {
		// Assuming the channel is mentioned like <#ChannelID>
		if isChannelMention(part) {
			channel := extractChannelFromMention(part)
			*parts = append((*parts)[:i], (*parts)[i+1:]...) // Remove the parsed part
			return channel
		}
	}
	return nil // No valid channel found
}

func isChannelMention(s string) bool {
	// Check if the part is a valid channel mention
	return strings.HasPrefix(s, "<#") && strings.HasSuffix(s, ">")
}

func extractChannelFromMention(s string) *discordgo.Channel {
	// Extract channel ID from the mention and return the channel (you can use discordgo's method to fetch the channel)
	channelID := strings.Trim(s, "<#>")
	return &discordgo.Channel{ID: channelID}
}

func pickString(parts *[]string) string {
	if len(*parts) == 0 {
		return ""
	}
	part := (*parts)[0]
	*parts = (*parts)[1:]
	return part
}

func pickStringRest(parts *[]string) string {
	return strings.Join(*parts, " ")
}

func pickInt(parts *[]string) int {
	if len(*parts) == 0 {
		return 0
	}
	part := (*parts)[0]
	*parts = (*parts)[1:]
	i, err := strconv.Atoi(part)
	if err != nil {
		return 0
	}
	return i
}
