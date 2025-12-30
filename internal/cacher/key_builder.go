package cacher

import (
	"fmt"
	"strings"

	"github.com/pzsp-teams/lib/internal/pepper"
	"github.com/pzsp-teams/lib/internal/util"
)

type KeyType string

const (
	Team            KeyType = "team"
	Channel         KeyType = "channel"
	ChannelMember   KeyType = "channel-member"
	GroupChat       KeyType = "group-chat"
	DirectChat      KeyType = "direct-chat"
	GroupChatMember KeyType = "group-chat-member"
	User            KeyType = "user"
)

func formatKey(t KeyType, parts ...string) string {
	for i, p := range parts {
		parts[i] = strings.TrimSpace(p)
	}
	return fmt.Sprintf("$%s$:%s", t, strings.Join(parts, ":"))
}

func NewTeamKey(name string) string {
	return formatKey(Team, name)
}

func NewChannelKey(teamID, name string) string {
	return formatKey(Channel, teamID, name)
}

func NewOneOnOneChatKey(userRef string, pep *string) string {
	return formatKey(DirectChat, hashRef(userRef, pep))
}

func NewGroupChatKey(topic string) string {
	return formatKey(GroupChat, topic)
}

func NewGroupChatMemberKey(chatID, userRef string, pep *string) string {
	return formatKey(GroupChatMember, chatID, hashRef(userRef, pep))
}

func NewChannelMemberKey(teamID, channelID, userRef string, pep *string) string {
	return formatKey(ChannelMember, teamID, channelID, hashRef(userRef, pep))
}

func NewUserKey(userRef string, pep *string) string {
	return formatKey(User, hashRef(userRef, pep))
}

func hashRef(ref string, pep *string) string {
	if pep == nil {
		p, err := pepper.GetOrAskPepper()
		if err != nil {
			p = "default-pepper"
		}
		pep = &p
	}
	return util.HashWithPepper(*pep, strings.TrimSpace(ref))
}
