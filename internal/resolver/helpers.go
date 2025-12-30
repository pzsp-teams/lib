package resolver

import (
	"fmt"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/internal/resources"
	"github.com/pzsp-teams/lib/internal/util"
)

func resolveTeamIDByName(list msmodels.TeamCollectionResponseable, ref string) (string, error) {
	if list == nil || list.GetValue() == nil || len(list.GetValue()) == 0 {
		return "", &resourcesNotAvailableError{resourceType: resources.Team}
	}
	matches := make([]msmodels.Teamable, 0, len(list.GetValue()))
	for _, t := range list.GetValue() {
		if t == nil {
			continue
		}
		if util.Deref(t.GetDisplayName()) == ref {
			matches = append(matches, t)
		}
	}
	switch len(matches) {
	case 0:
		return "", &resourceNotFoundError{resourceType: resources.Team, ref: ref}
	case 1:
		id := util.Deref(matches[0].GetId())
		if id == "" {
			return "", &resourceEmptyIDError{resourceType: resources.Team, ref: ref}
		}
		return id, nil
	default:
		var options []string
		for _, t := range matches {
			options = append(options,
				fmt.Sprintf("%s (ID: %s)", util.Deref(t.GetDisplayName()), util.Deref(t.GetId())))
		}
		return "", &resourceAmbiguousError{
			resourceType: resources.Team,
			ref:          ref,
			options:      options,
		}
	}
}

func resolveChannelIDByName(chans msmodels.ChannelCollectionResponseable, teamID, ref string) (string, error) {
	if chans == nil || chans.GetValue() == nil || len(chans.GetValue()) == 0 {
		return "", &resourcesNotAvailableError{resourceType: resources.Channel}
	}
	matches := make([]msmodels.Channelable, 0, len(chans.GetValue()))
	for _, c := range chans.GetValue() {
		if c == nil {
			continue
		}
		if util.Deref(c.GetDisplayName()) == ref {
			matches = append(matches, c)
		}
	}
	switch len(matches) {
	case 0:
		return "", &resourceNotFoundError{resourceType: resources.Channel, ref: ref}
	case 1:
		id := util.Deref(matches[0].GetId())
		if id == "" {
			return "", &resourceEmptyIDError{resourceType: resources.Channel, ref: ref}
		}
		return id, nil
	default:
		var options []string
		for _, c := range matches {
			options = append(options,
				fmt.Sprintf("%s (ID: %s)", util.Deref(c.GetDisplayName()), util.Deref(c.GetId())))
		}
		return "", &resourceAmbiguousError{
			resourceType: resources.Channel,
			ref:          ref,
			options:      options,
		}
	}
}

func resolveOneOnOneChatIDByUserRef(chats msmodels.ChatCollectionResponseable, userRef string) (string, error) {
	if chats == nil || chats.GetValue() == nil || len(chats.GetValue()) == 0 {
		return "", &resourcesNotAvailableError{resourceType: resources.OneOnOneChat}
	}

	for _, chat := range chats.GetValue() {
		if chat == nil {
			continue
		}
		members := chat.GetMembers()
		for _, member := range members {
			um, ok := member.(msmodels.AadUserConversationMemberable)
			if !ok {
				continue
			}
			if matchesUserRef(um, userRef) {
				return util.Deref(chat.GetId()), nil
			}
		}
	}
	return "", &resourceNotFoundError{resourceType: resources.OneOnOneChat, ref: userRef}
}

func resolveGroupChatIDByTopic(chats msmodels.ChatCollectionResponseable, topic string) (string, error) {
	if chats == nil || chats.GetValue() == nil || len(chats.GetValue()) == 0 {
		return "", &resourcesNotAvailableError{resourceType: resources.GroupChat}
	}

	matches := make([]msmodels.Chatable, 0, len(chats.GetValue()))
	for _, chat := range chats.GetValue() {
		if chat == nil {
			continue
		}
		if util.Deref(chat.GetTopic()) == topic {
			matches = append(matches, chat)
		}
	}
	switch len(matches) {
	case 0:
		return "", &resourceNotFoundError{resourceType: resources.GroupChat, ref: topic}
	case 1:
		id := util.Deref(matches[0].GetId())
		if id == "" {
			return "", &resourceEmptyIDError{resourceType: resources.GroupChat, ref: topic}
		}
		return util.Deref(matches[0].GetId()), nil
	default:
		var options []string
		for _, c := range matches {
			options = append(options,
				fmt.Sprintf("%s (ID: %s)", util.Deref(c.GetTopic()), util.Deref(c.GetId())))
		}
		return "", &resourceAmbiguousError{
			resourceType: resources.GroupChat,
			ref:          topic,
			options:      options,
		}
	}
}

func resolveMemberID(members msmodels.ConversationMemberCollectionResponseable, ref string) (string, error) {
	if members == nil || members.GetValue() == nil || len(members.GetValue()) == 0 {
		return "", &resourcesNotAvailableError{resourceType: resources.User}
	}
	for _, member := range members.GetValue() {
		if member == nil {
			continue
		}
		um, ok := member.(msmodels.AadUserConversationMemberable)
		if !ok {
			continue
		}
		if matchesUserRef(um, ref) {
			return util.Deref(member.GetId()), nil
		}
	}
	return "", &resourceNotFoundError{resourceType: resources.User, ref: ref}
}

func matchesUserRef(um msmodels.AadUserConversationMemberable, userRef string) bool {
	if userRef == "" {
		return false
	}
	if util.Deref(um.GetUserId()) == userRef {
		return true
	}
	email, err := um.GetBackingStore().Get("email")
	if err == nil {
		if emailStr, ok := email.(*string); ok {
			if util.Deref(emailStr) == userRef {
				return true
			}
		}
	}
	return false
}
