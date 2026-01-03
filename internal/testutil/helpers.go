package testutil

import (
	"errors"
	"testing"
	"time"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	sender "github.com/pzsp-teams/lib/internal/sender"
	"github.com/stretchr/testify/require"
)

// TEAM UTILS

type NewTeamParams struct {
	ID          *string
	DisplayName *string
	Description *string
	IsArchived  *bool
	Visibility  *msmodels.TeamVisibilityType
}

func NewGraphTeam(params *NewTeamParams) msmodels.Teamable {
	if params == nil {
		return nil
	}
	graphTeam := msmodels.NewTeam()

	graphTeam.SetId(params.ID)
	graphTeam.SetDisplayName(params.DisplayName)
	graphTeam.SetDescription(params.Description)
	graphTeam.SetIsArchived(params.IsArchived)
	graphTeam.SetVisibility(params.Visibility)

	return graphTeam
}

func NewTeamCollection(teams ...msmodels.Teamable) msmodels.TeamCollectionResponseable {
	col := msmodels.NewTeamCollectionResponse()
	col.SetValue(teams)
	return col
}

// CHANNEL UTILS

type NewChannelParams struct {
	ID   *string
	Name *string
}

func NewGraphChannel(params *NewChannelParams) msmodels.Channelable {
	if params == nil {
		return nil
	}
	graphChannel := msmodels.NewChannel()
	graphChannel.SetId(params.ID)
	graphChannel.SetDisplayName(params.Name)
	return graphChannel
}

func NewChannelCollection(channels ...msmodels.Channelable) msmodels.ChannelCollectionResponseable {
	col := msmodels.NewChannelCollectionResponse()
	col.SetValue(channels)
	return col
}

// MEMBER UTILS

type NewMemberParams struct {
	ID          *string
	UserID      *string
	DisplayName *string
	Roles       []string
	Email       *string
}

func NewGraphMember(params *NewMemberParams) msmodels.ConversationMemberable {
	if params == nil {
		return nil
	}
	member := msmodels.NewAadUserConversationMember()

	member.SetId(params.ID)

	member.SetUserId(params.UserID)
	member.SetDisplayName(params.DisplayName)
	_ = member.GetBackingStore().Set("email", params.Email)

	if params.Roles != nil {
		member.SetRoles(params.Roles)
	}

	return member
}

func NewMemberCollection(
	members ...msmodels.ConversationMemberable,
) msmodels.ConversationMemberCollectionResponseable {
	col := msmodels.NewConversationMemberCollectionResponse()
	col.SetValue(members)
	return col
}

// CHAT UTILS

type NewChatParams struct {
	ID       *string
	Type     *msmodels.ChatType
	Members  []msmodels.ConversationMemberable
	IsHidden *bool
	Topic    *string
}

func NewGraphChat(params *NewChatParams) msmodels.Chatable {
	if params == nil {
		return nil
	}
	chat := msmodels.NewChat()

	chat.SetId(params.ID)
	chat.SetChatType(params.Type)
	chat.SetMembers(params.Members)
	chat.SetIsHiddenForAllMembers(params.IsHidden)
	chat.SetTopic(params.Topic)

	return chat
}

func NewChatCollection(chats ...msmodels.Chatable) msmodels.ChatCollectionResponseable {
	col := msmodels.NewChatCollectionResponse()
	col.SetValue(chats)
	return col
}

// MESSAGE

type NewMessageParams struct {
	ID              *string
	Content         *string
	ContentType     *msmodels.BodyType
	CreatedDateTime *time.Time
	FromUserID      *string
	FromDisplayName *string
	ReplyCount      *int
}

func NewGraphMessage(params *NewMessageParams) msmodels.ChatMessageable {
	if params == nil {
		return nil
	}
	graphMessage := msmodels.NewChatMessage()
	graphMessage.SetId(params.ID)

	if params.Content != nil || params.ContentType != nil {
		body := msmodels.NewItemBody()
		body.SetContent(params.Content)
		body.SetContentType(params.ContentType)
		graphMessage.SetBody(body)
	}

	graphMessage.SetCreatedDateTime(params.CreatedDateTime)

	if params.FromUserID != nil || params.FromDisplayName != nil {
		from := msmodels.NewChatMessageFromIdentitySet()

		user := msmodels.NewUser()
		user.SetId(params.FromUserID)
		user.SetDisplayName(params.FromDisplayName)

		from.SetUser(user)
		graphMessage.SetFrom(from)
	}

	if params.ReplyCount != nil {
		messages := make([]msmodels.ChatMessageable, *params.ReplyCount)
		for i := range messages {
			messages[i] = msmodels.NewChatMessage()
		}
		graphMessage.SetReplies(messages)
	}

	return graphMessage
}

// USER
type NewUserParams struct {
	ID          *string
	DisplayName *string
}

func NewGraphUser(params *NewUserParams) msmodels.Userable {
	if params == nil {
		return nil
	}
	u := msmodels.NewUser()
	if params.ID != nil && *params.ID != "" {
		u.SetId(params.ID)
	}
	if params.DisplayName != nil && *params.DisplayName != "" {
		u.SetDisplayName(params.DisplayName)
	}
	return u
}

func RequireWrapped(t *testing.T, err error) *sender.OpError {
	t.Helper()
	require.Error(t, err)

	var opErr *sender.OpError
	require.ErrorAs(t, err, &opErr)
	return opErr
}

func RequireReqErrCode(t *testing.T, err error, wantCode int) {
	t.Helper()
	_ = RequireWrapped(t, err)

	var re *sender.RequestError
	if errors.As(err, &re) {
		require.Equal(t, wantCode, re.Code)
		return
	}

	var forbidden *sender.ErrAccessForbidden
	if errors.As(err, &forbidden) {
		require.Equal(t, wantCode, 403)
		return
	}

	var notFound *sender.ErrResourceNotFound
	if errors.As(err, &notFound) {
		require.Equal(t, wantCode, 404)
		return
	}

	t.Fatalf("expected error with code=%d (RequestError/ErrAccessForbidden/ErrResourceNotFound), got: %T: %v",
		wantCode, err, err)
}
