package adapter

import (
	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/internal/util"
	"github.com/pzsp-teams/lib/models"
)

// MapGraphTeam maps a Microsoft Graph Teamable to simplified Team model.
func MapGraphTeam(graphTeam msmodels.Teamable) *models.Team {
	if graphTeam == nil {
		return nil
	}

	var visibility string

	if v := graphTeam.GetVisibility(); v != nil {
		visibility = v.String()
	}

	out := &models.Team{
		ID:          util.Deref(graphTeam.GetId()),
		DisplayName: util.Deref(graphTeam.GetDisplayName()),
		Description: util.Deref(graphTeam.GetDescription()),
		IsArchived:  util.Deref(graphTeam.GetIsArchived()),
		Visibility:  visibility,
	}
	return out
}

// MapGraphChannel maps a Microsoft Graph Channelable to simplified Channel model.
func MapGraphChannel(graphChannel msmodels.Channelable) *models.Channel {
	if graphChannel == nil {
		return nil
	}

	name := util.Deref(graphChannel.GetDisplayName())

	return &models.Channel{
		ID:        util.Deref(graphChannel.GetId()),
		Name:      name,
		IsGeneral: name == "General",
	}
}

// MapGraphMessage maps a Microsoft Graph ChatMessageable to simplified Message model.
func MapGraphMessage(graphMessage msmodels.ChatMessageable) *models.Message {
	if graphMessage == nil {
		return nil
	}

	var content string
	var contentType models.MessageContentType
	var from *models.MessageFrom
	var replyCount int

	if body := graphMessage.GetBody(); body != nil {
		content = util.Deref(body.GetContent())
		contentType = models.MessageContentTypeText
		if ct := body.GetContentType(); ct != nil {
			if *ct == msmodels.HTML_BODYTYPE {
				contentType = models.MessageContentTypeHTML
			}
		}
	}

	if graphFrom := graphMessage.GetFrom(); graphFrom != nil {
		if user := graphFrom.GetUser(); user != nil {
			from = &models.MessageFrom{
				UserID:      util.Deref(user.GetId()),
				DisplayName: util.Deref(user.GetDisplayName()),
			}
		}
	}

	if replies := graphMessage.GetReplies(); replies != nil {
		replyCount = len(replies)
	}

	return &models.Message{
		ID:              util.Deref(graphMessage.GetId()),
		Content:         content,
		ContentType:     contentType,
		CreatedDateTime: util.Deref(graphMessage.GetCreatedDateTime()),
		From:            from,
		ReplyCount:      replyCount,
	}
}

// MapGraphMember maps a Microsoft Graph ConversationMemberable to simplified Member model.
func MapGraphMember(graphMember msmodels.ConversationMemberable) *models.Member {
	if graphMember == nil {
		return nil
	}

	var userID string
	var displayName string

	if userMember, ok := graphMember.(*msmodels.AadUserConversationMember); ok {
		userID = util.Deref(userMember.GetUserId())
		displayName = util.Deref(userMember.GetDisplayName())
	}

	role := ""
	roles := graphMember.GetRoles()
	if len(roles) > 0 {
		role = roles[0]
	}

	return &models.Member{
		ID:          util.Deref(graphMember.GetId()),
		UserID:      userID,
		DisplayName: displayName,
		Role:        role,
	}
}

// MapGraphChat maps a Microsoft Graph Chatable to simplified Chat model.
func MapGraphChat(graphChat msmodels.Chatable) *models.Chat {
	if graphChat == nil {
		return nil
	}

	var chatType models.ChatType

	if grapChatType := graphChat.GetChatType(); grapChatType != nil {
		chatType = models.ChatTypeOneOnOne
		if *grapChatType == msmodels.GROUP_CHATTYPE {
			chatType = models.ChatTypeGroup
		}
	}

	return &models.Chat{
		ID:       util.Deref(graphChat.GetId()),
		Type:     chatType,
		IsHidden: util.Deref(graphChat.GetIsHiddenForAllMembers()),
		Topic:    graphChat.GetTopic(),
	}
}
