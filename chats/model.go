package chats

import (
	"time"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
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
		ID:       deref(graphChat.GetId()),
		Type:     deref(graphChat.GetChatType()).String(),
		IsHidden: deref(graphChat.GetIsHiddenForAllMembers()),
		Topic:    graphChat.GetTopic(),
	}
}

func mapGraphChatMember(graphChatMember msmodels.ConversationMemberable) *ChatMember {
	if aadMember, ok := graphChatMember.(msmodels.AadUserConversationMemberable); ok {
		return &ChatMember{
			ID:            deref(aadMember.GetId()),
			Email:         deref(aadMember.GetEmail()),
			DisplayedName: deref(aadMember.GetDisplayName()),
			Roles:         aadMember.GetRoles(),
		}
	}
	return nil
}
func mapGraphChatMessage(graphMsg msmodels.ChatMessageable) *ChatMessage {
	var from string

	if fromIdentity := graphMsg.GetFrom(); fromIdentity != nil {
		if user := fromIdentity.GetUser(); user != nil {
			from = deref(user.GetDisplayName())
		}
	}

	return &ChatMessage{
		ID:     deref(graphMsg.GetId()),
		From:   from,
		SendAt: deref(graphMsg.GetCreatedDateTime()),
	}
}
