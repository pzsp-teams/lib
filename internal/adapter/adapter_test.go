package adapter

import (
	"testing"
	"time"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/models"
	"github.com/stretchr/testify/assert"
)

func ptr[T any](v T) *T { return &v }

type newTeamParams struct {
	ID          *string
	DisplayName *string
	Description *string
	IsArchived  *bool
	Visibilitiy *msmodels.TeamVisibilityType
}

func newGraphTeam(params *newTeamParams) msmodels.Teamable {
	if params == nil {
		return nil
	}
	graphTeam := msmodels.NewTeam()

	graphTeam.SetId(params.ID)
	graphTeam.SetDisplayName(params.DisplayName)
	graphTeam.SetDescription(params.Description)
	graphTeam.SetIsArchived(params.IsArchived)
	graphTeam.SetVisibility(params.Visibilitiy)

	return graphTeam
}

func TestMapGraphTeam(t *testing.T) {
	type testCase struct {
		name   string
		params *newTeamParams
		result *models.Team
	}

	testCases := []testCase{
		{
			"Nil team",
			nil,
			nil,
		},
		{
			"Private team",
			&newTeamParams{ptr("team-id"), ptr("Team name"), ptr("A sample team"), ptr(true), ptr(msmodels.PRIVATE_TEAMVISIBILITYTYPE)},
			&models.Team{ID: "team-id", DisplayName: "Team name", Description: "A sample team", IsArchived: true, Visibility: ptr("private")},
		},
		{
			"Public team",
			&newTeamParams{ptr("team-id"), ptr("Team name"), ptr("A sample team"), ptr(false), ptr(msmodels.PUBLIC_TEAMVISIBILITYTYPE)},
			&models.Team{ID: "team-id", DisplayName: "Team name", Description: "A sample team", IsArchived: false, Visibility: ptr("public")},
		},
		{
			"Missing fields",
			&newTeamParams{},
			&models.Team{ID: "", DisplayName: "", Description: "", IsArchived: false, Visibility: nil},
		},
	}

	for _, tc := range testCases {
		graphTeam := newGraphTeam(tc.params)
		team := MapGraphTeam(graphTeam)

		if team == nil || tc.result == nil {
			assert.Equal(t, tc.result, team)
			continue
		}

		assert.Equal(t, tc.result.ID, team.ID)
		assert.Equal(t, tc.result.DisplayName, team.DisplayName)
		assert.Equal(t, tc.result.Description, team.Description)
		assert.Equal(t, tc.result.IsArchived, team.IsArchived)

		if tc.result.Visibility == nil || team.Visibility == nil {
			assert.Equal(t, tc.result.Visibility, team.Visibility)
		} else {
			assert.Equal(t, *tc.result.Visibility, *team.Visibility)
		}
	}
}

type newChannelParams struct {
	ID   *string
	Name *string
}

func newGraphChannel(params *newChannelParams) msmodels.Channelable {
	if params == nil {
		return nil
	}
	graphChannel := msmodels.NewChannel()
	graphChannel.SetId(params.ID)
	graphChannel.SetDisplayName(params.Name)
	return graphChannel
}

func TestMapGraphChannel(t *testing.T) {
	type testCase struct {
		name   string
		params *newChannelParams
		result *models.Channel
	}

	testCases := []testCase{
		{
			"Nil channel",
			nil,
			nil,
		},
		{
			"General channel",
			&newChannelParams{ptr("channel-id"), ptr("General")},
			&models.Channel{ID: "channel-id", Name: "General", IsGeneral: true},
		},
		{
			"Standard channel",
			&newChannelParams{ptr("channel-id"), ptr("channel-name")},
			&models.Channel{ID: "channel-id", Name: "channel-name", IsGeneral: false},
		},
		{
			"Missing fields",
			&newChannelParams{},
			&models.Channel{ID: "", Name: "", IsGeneral: false},
		},
	}

	for _, tc := range testCases {
		graphChannel := newGraphChannel(tc.params)
		channel := MapGraphChannel(graphChannel)

		if channel == nil || tc.result == nil {
			assert.Equal(t, tc.result, channel)
			continue
		}

		assert.Equal(t, tc.result.ID, channel.ID)
		assert.Equal(t, tc.result.Name, channel.Name)
		assert.Equal(t, tc.result.IsGeneral, channel.IsGeneral)
	}
}

type newMessageParams struct {
	ID              *string
	Content         *string
	ContentType     *msmodels.BodyType
	CreatedDateTime *time.Time
	FromUserID      *string
	FromDisplayName *string
	ReplyCount      *int
}

func newGraphMessage(params *newMessageParams) msmodels.ChatMessageable {
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

func TestMapGraphMessage(t *testing.T) {
	type testCase struct {
		name   string
		params *newMessageParams
		result *models.Message
	}

	testCases := []testCase{
		{
			"Nil message",
			nil,
			nil,
		},
		{
			"Text content message",
			&newMessageParams{
				ptr("message-id"),
				ptr("Hello, world!"),
				ptr(msmodels.TEXT_BODYTYPE),
				ptr(time.Date(2024, 1, 2, 15, 4, 5, 0, time.UTC)),
				ptr("user-id"),
				ptr("John Doe"),
				ptr(3),
			},
			&models.Message{
				ID:              "message-id",
				Content:         "Hello, world!",
				ContentType:     models.MessageContentTypeText,
				CreatedDateTime: time.Date(2024, 1, 2, 15, 4, 5, 0, time.UTC),
				From:            &models.MessageFrom{UserID: "user-id", DisplayName: "John Doe"},
				ReplyCount:      3,
			},
		},
		{
			"HTML content message",
			&newMessageParams{
				ptr("message-id"),
				ptr("<p>Hello, <b>world</b>!</p>"),
				ptr(msmodels.HTML_BODYTYPE),
				ptr(time.Date(2024, 1, 2, 15, 4, 5, 0, time.UTC)),
				ptr("user-id"),
				ptr("John Doe"),
				ptr(0),
			},
			&models.Message{
				ID:              "message-id",
				Content:         "<p>Hello, <b>world</b>!</p>",
				ContentType:     models.MessageContentTypeHTML,
				CreatedDateTime: time.Date(2024, 1, 2, 15, 4, 5, 0, time.UTC),
				From:            &models.MessageFrom{UserID: "user-id", DisplayName: "John Doe"},
				ReplyCount:      0,
			},
		},
		{
			"Missing fields",
			&newMessageParams{},
			&models.Message{
				ID:              "",
				Content:         "",
				ContentType:     "",
				CreatedDateTime: time.Time{},
				From:            nil,
				ReplyCount:      0,
			},
		},
	}

	for _, tc := range testCases {
		graphMessage := newGraphMessage(tc.params)
		message := MapGraphMessage(graphMessage)

		if message == nil || tc.result == nil {
			assert.Equal(t, tc.result, message)
			continue
		}

		assert.Equal(t, tc.result.ID, message.ID)
		assert.Equal(t, tc.result.Content, message.Content)
		assert.Equal(t, tc.result.ContentType, message.ContentType)
		assert.Equal(t, tc.result.CreatedDateTime, message.CreatedDateTime)

		if tc.result.From == nil || message.From == nil {
			assert.Equal(t, tc.result.From, message.From)
		} else {
			assert.Equal(t, tc.result.From.UserID, message.From.UserID)
			assert.Equal(t, tc.result.From.DisplayName, message.From.DisplayName)
		}

		assert.Equal(t, tc.result.ReplyCount, message.ReplyCount)
	}
}

type newMemberParams struct {
	ID          *string
	UserID      *string
	DisplayName *string
	Role        *string
	Email       *string
}

func newGraphMember(params *newMemberParams) msmodels.ConversationMemberable {
	if params == nil {
		return nil
	}
	member := msmodels.NewAadUserConversationMember()

	member.SetId(params.ID)

	member.SetUserId(params.UserID)
	member.SetDisplayName(params.DisplayName)

	if params.Role != nil {
		member.SetRoles([]string{*params.Role})
	}

	return member
}

func TestMapGraphMember(t *testing.T) {
	type testCase struct {
		name   string
		params *newMemberParams
		result *models.Member
	}

	testCases := []testCase{
		{
			"Nil member",
			nil,
			nil,
		},
		{
			"Complete member",
			&newMemberParams{ptr("member-id"), ptr("user-id"), ptr("Jane Smith"), ptr("owner"), ptr("jane.smith@example.com")},
			&models.Member{ID: "member-id", UserID: "user-id", DisplayName: "Jane Smith", Role: "owner", Email: "jane.smith@example.com"},
		},
		{
			"Missing fields",
			&newMemberParams{},
			&models.Member{ID: "", UserID: "", DisplayName: "", Role: "", Email: ""},
		},
	}

	for _, tc := range testCases {
		graphMember := newGraphMember(tc.params)
		member := MapGraphMember(graphMember)

		if member == nil || tc.result == nil {
			assert.Equal(t, tc.result, member)
			continue
		}

		assert.Equal(t, tc.result.ID, member.ID)
		assert.Equal(t, tc.result.UserID, member.UserID)
		assert.Equal(t, tc.result.DisplayName, member.DisplayName)
		assert.Equal(t, tc.result.Role, member.Role)
	}
}

type newChatParams struct {
	ID       *string
	Type     *msmodels.ChatType
	IsHidden *bool
	Topic    *string
}

func newGraphChat(params *newChatParams) msmodels.Chatable {
	if params == nil {
		return nil
	}
	chat := msmodels.NewChat()

	chat.SetId(params.ID)
	chat.SetChatType(params.Type)
	chat.SetIsHiddenForAllMembers(params.IsHidden)
	chat.SetTopic(params.Topic)

	return chat
}

func TestMapGraphChat(t *testing.T) {
	type testCase struct {
		name   string
		params *newChatParams
		result *models.Chat
	}

	testCases := []testCase{
		{
			"Nil chat",
			nil,
			nil,
		},
		{
			"One-on-one chat",
			&newChatParams{ptr("chat-id"), ptr(msmodels.ONEONONE_CHATTYPE), ptr(false), nil},
			&models.Chat{ID: "chat-id", Type: models.ChatTypeOneOnOne, IsHidden: false, Topic: nil},
		},
		{
			"Group chat with topic",
			&newChatParams{ptr("chat-id"), ptr(msmodels.GROUP_CHATTYPE), ptr(true), ptr("Project Discussion")},
			&models.Chat{ID: "chat-id", Type: models.ChatTypeGroup, IsHidden: true, Topic: ptr("Project Discussion")},
		},
		{
			"Missing fields",
			&newChatParams{},
			&models.Chat{ID: "", Type: "", IsHidden: false, Topic: nil},
		},
	}

	for _, tc := range testCases {
		graphChat := newGraphChat(tc.params)
		chat := MapGraphChat(graphChat)

		if chat == nil || tc.result == nil {
			assert.Equal(t, tc.result, chat)
			continue
		}

		assert.Equal(t, tc.result.ID, chat.ID)
		assert.Equal(t, tc.result.Type, chat.Type)
		assert.Equal(t, tc.result.IsHidden, chat.IsHidden)

		if tc.result.Topic == nil || chat.Topic == nil {
			assert.Equal(t, tc.result.Topic, chat.Topic)
		} else {
			assert.Equal(t, *tc.result.Topic, *chat.Topic)
		}
	}
}
