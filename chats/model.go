package chats

import (
	"time"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/internal/util"
)

type Chat struct {
	ID       string
	Type     string
	IsHidden bool
	Topic    *string
}

type ChatMember struct {
	ID            string
	Email         string
	DisplayedName string
	Roles         []string
}

type ChatMessage struct {
	ID     string
	From   string
	SendAt time.Time
}

func mapGraphChat(graphChat msmodels.Chatable) *Chat {
	return &Chat{
		ID:       util.Deref(graphChat.GetId()),
		Type:     util.Deref(graphChat.GetChatType()).String(),
		IsHidden: util.Deref(graphChat.GetIsHiddenForAllMembers()),
		Topic:    graphChat.GetTopic(),
	}
}

func mapGraphChatMember(graphChatMember msmodels.ConversationMemberable) *ChatMember {
	if aadMember, ok := graphChatMember.(msmodels.AadUserConversationMemberable); ok {
		return &ChatMember{
			ID:            util.Deref(aadMember.GetId()),
			Email:         util.Deref(aadMember.GetEmail()),
			DisplayedName: util.Deref(aadMember.GetDisplayName()),
			Roles:         aadMember.GetRoles(),
		}
	}
	return nil
}
func mapGraphChatMessage(graphMsg msmodels.ChatMessageable) *ChatMessage {
	var from string

	if fromIdentity := graphMsg.GetFrom(); fromIdentity != nil {
		if user := fromIdentity.GetUser(); user != nil {
			from = util.Deref(user.GetDisplayName())
		}
	}

	return &ChatMessage{
		ID:     util.Deref(graphMsg.GetId()),
		From:   from,
		SendAt: util.Deref(graphMsg.GetCreatedDateTime()),
	}
}
