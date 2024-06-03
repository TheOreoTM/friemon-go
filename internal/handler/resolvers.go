package handler

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var (
	SnowflakeRegex           = regexp.MustCompile(`^(\d{17,19})$`)
	UserOrMemberMentionRegex = regexp.MustCompile(`^<@!?(\d{17,19})>$`)
)

func resolveUser(session *discordgo.Session, parameter string) *discordgo.User {
	userId := strings.Trim(parameter, "<@!>")

	user, err := session.User(userId)
	if err != nil {
		return nil
	}

	return user
}

func resolveMember(session *discordgo.Session, guild *discordgo.Guild, parameter string) (*discordgo.Member, error) {
	member, _ := resolveById(parameter, guild, session)
	if member != nil {
		return member, nil
	}

	member, err := resolveByQuery(parameter, session, guild)
	if member != nil {
		return member, nil
	}

	if err != nil {
		return nil, err
	}

	return nil, fmt.Errorf("ArgumentMemberError")
}

func resolveById(id string, guild *discordgo.Guild, session *discordgo.Session) (*discordgo.Member, error) {
	memberId := strings.Trim(id, "<@!>")

	if len(memberId) > 0 {
		member, err := session.GuildMember(guild.ID, memberId)
		if err != nil {
			return nil, err
		}
		return member, nil
	}

	return nil, nil
}

func resolveByQuery(argument string, session *discordgo.Session, guild *discordgo.Guild) (*discordgo.Member, error) {
	if len(argument) > 5 && strings.Contains(argument, "#") {
		argument = argument[:len(argument)-5]
	}

	members, err := session.GuildMembersSearch(guild.ID, argument, 1)
	if err != nil {
		return nil, err
	}

	member := members[0]

	if member != nil {
		return member, nil
	}

	return nil, nil
}
