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
	userId := UserOrMemberMentionRegex.FindStringSubmatch(parameter)
	if userId == nil {
		return nil
	}

	user, err := session.User(userId[1])
	if err != nil {
		return nil
	}

	return user
}

func resolveMember(session *discordgo.Session, guild *discordgo.Guild, parameter string) (*discordgo.Member, error) {
	member, err := resolveById(parameter, guild, session)
	if err != nil {
		return nil, err
	}

	if member == nil {
		member, err = resolveByQuery(parameter, guild)
		if err != nil {
			return nil, err
		}
	}

	if member != nil {
		return member, nil
	}

	return nil, fmt.Errorf("ArgumentMemberError")
}

func resolveById(id string, guild *discordgo.Guild, session *discordgo.Session) (*discordgo.Member, error) {
	memberId := UserOrMemberMentionRegex.FindStringSubmatch(id)
	if len(memberId) == 0 {
		memberId = SnowflakeRegex.FindStringSubmatch(id)
	}

	if len(memberId) > 0 {
		member, err := session.GuildMember(guild.ID, memberId[1])
		if err != nil {
			return nil, err
		}
		return member, nil
	}

	return nil, nil
}

func resolveByQuery(argument string, guild *discordgo.Guild) (*discordgo.Member, error) {
	if len(argument) > 5 && strings.Contains(argument, "#") {
		argument = argument[:len(argument)-5]
	}

	members := guild.Members
	for _, member := range members {
		if strings.EqualFold(member.User.Username, argument) {
			return member, nil
		}
	}

	return nil, nil
}
