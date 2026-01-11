package adapter

import (
	"testing"
	"time"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/internal/testutil"
	"github.com/pzsp-teams/lib/internal/util"
	"github.com/pzsp-teams/lib/models"
	"github.com/stretchr/testify/assert"
)

func TestMapGraphTeam(t *testing.T) {
	type testCase struct {
		name   string
		params *testutil.NewTeamParams
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
			&testutil.NewTeamParams{
				ID:          util.Ptr("team-id"),
				DisplayName: util.Ptr("Team name"),
				Description: util.Ptr("A sample team"),
				IsArchived:  util.Ptr(true),
				Visibility:  util.Ptr(msmodels.PRIVATE_TEAMVISIBILITYTYPE),
			},
			&models.Team{
				ID:          "team-id",
				DisplayName: "Team name",
				Description: "A sample team",
				IsArchived:  true,
				Visibility:  util.Ptr("private"),
			},
		},
		{
			"Public team",
			&testutil.NewTeamParams{
				ID:          util.Ptr("team-id"),
				DisplayName: util.Ptr("Team name"),
				Description: util.Ptr("A sample team"),
				IsArchived:  util.Ptr(false),
				Visibility:  util.Ptr(msmodels.PUBLIC_TEAMVISIBILITYTYPE),
			},
			&models.Team{
				ID:          "team-id",
				DisplayName: "Team name",
				Description: "A sample team",
				IsArchived:  false,
				Visibility:  util.Ptr("public"),
			},
		},
		{
			"Missing fields",
			&testutil.NewTeamParams{},
			&models.Team{
				ID:          "",
				DisplayName: "",
				Description: "",
				IsArchived:  false,
				Visibility:  nil,
			},
		},
	}

	for _, tc := range testCases {
		graphTeam := testutil.NewGraphTeam(tc.params)
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

func TestMapGraphChannel(t *testing.T) {
	type testCase struct {
		name   string
		params *testutil.NewChannelParams
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
			&testutil.NewChannelParams{ID: util.Ptr("channel-id"), Name: util.Ptr("General")},
			&models.Channel{ID: "channel-id", Name: "General", IsGeneral: true},
		},
		{
			"Standard channel",
			&testutil.NewChannelParams{
				ID:   util.Ptr("channel-id"),
				Name: util.Ptr("channel-name"),
			},
			&models.Channel{ID: "channel-id", Name: "channel-name", IsGeneral: false},
		},
		{
			"Missing fields",
			&testutil.NewChannelParams{},
			&models.Channel{ID: "", Name: "", IsGeneral: false},
		},
	}

	for _, tc := range testCases {
		graphChannel := testutil.NewGraphChannel(tc.params)
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

func TestMapGraphMessage(t *testing.T) {
	type testCase struct {
		name   string
		params *testutil.NewMessageParams
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
			&testutil.NewMessageParams{
				ID:              util.Ptr("message-id"),
				Content:         util.Ptr("Hello, world!"),
				ContentType:     util.Ptr(msmodels.TEXT_BODYTYPE),
				CreatedDateTime: util.Ptr(time.Date(2024, 1, 2, 15, 4, 5, 0, time.UTC)),
				FromUserID:      util.Ptr("user-id"),
				FromDisplayName: util.Ptr("John Doe"),
				ReplyCount:      util.Ptr(3),
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
			&testutil.NewMessageParams{
				ID:              util.Ptr("message-id"),
				Content:         util.Ptr("<p>Hello, <b>world</b>!</p>"),
				ContentType:     util.Ptr(msmodels.HTML_BODYTYPE),
				CreatedDateTime: util.Ptr(time.Date(2024, 1, 2, 15, 4, 5, 0, time.UTC)),
				FromUserID:      util.Ptr("user-id"),
				FromDisplayName: util.Ptr("John Doe"),
				ReplyCount:      util.Ptr(0),
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
			&testutil.NewMessageParams{},
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
		graphMessage := testutil.NewGraphMessage(tc.params)
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

func TestMapGraphPinnedMessage(t *testing.T) {
	type testCase struct {
		name   string
		params *testutil.NewPinnedMessageParams
		result *models.Message
	}

	testCases := []testCase{
		{
			"Nil pinned message",
			nil,
			nil,
		},
		{
			"Complete pinned message",
			&testutil.NewPinnedMessageParams{
				ID: util.Ptr("message-id"),
				NewMsgParams: &testutil.NewMessageParams{
					ID:              util.Ptr("message-id"),
					Content:         util.Ptr("Pinned message content"),
					ContentType:     util.Ptr(msmodels.TEXT_BODYTYPE),
					CreatedDateTime: util.Ptr(time.Date(2024, 2, 3, 10, 20, 30, 0, time.UTC)),
					FromUserID:      util.Ptr("user-id"),
					FromDisplayName: util.Ptr("Alice Johnson"),
					ReplyCount:      util.Ptr(5),
				},
			},
			&models.Message{
				ID:              "message-id",
				Content:         "Pinned message content",
				ContentType:     models.MessageContentTypeText,
				CreatedDateTime: time.Date(2024, 2, 3, 10, 20, 30, 0, time.UTC),
				From:            &models.MessageFrom{UserID: "user-id", DisplayName: "Alice Johnson"},
				ReplyCount:      5,
			},
		},
		{
			"Missing fields",
			&testutil.NewPinnedMessageParams{
				NewMsgParams: &testutil.NewMessageParams{},
			},
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
		graphPinnedMessage := testutil.NewGraphPinnedMessage(tc.params)
		pinnedMessage := MapGraphPinnedMessage(graphPinnedMessage)
		if pinnedMessage == nil || tc.result == nil {
			assert.Equal(t, tc.result, pinnedMessage)
			continue
		}

		assert.Equal(t, tc.result.ID, pinnedMessage.ID)
		assert.Equal(t, tc.result.Content, pinnedMessage.Content)
		assert.Equal(t, tc.result.ContentType, pinnedMessage.ContentType)
		assert.Equal(t, tc.result.CreatedDateTime, pinnedMessage.CreatedDateTime)

		if tc.result.From == nil || pinnedMessage.From == nil {
			assert.Equal(t, tc.result.From, pinnedMessage.From)
		} else {
			assert.Equal(t, tc.result.From.UserID, pinnedMessage.From.UserID)
			assert.Equal(t, tc.result.From.DisplayName, pinnedMessage.From.DisplayName)
		}

		assert.Equal(t, tc.result.ReplyCount, pinnedMessage.ReplyCount)
	}
}

func TestMapGraphMember(t *testing.T) {
	type testCase struct {
		name   string
		params *testutil.NewMemberParams
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
			&testutil.NewMemberParams{
				ID:          util.Ptr("member-id"),
				UserID:      util.Ptr("user-id"),
				DisplayName: util.Ptr("Jane Smith"),
				Roles:       []string{"owner"},
				Email:       util.Ptr("jane.smith@example.com"),
			},
			&models.Member{
				ID:          "member-id",
				UserID:      "user-id",
				DisplayName: "Jane Smith",
				Role:        "owner",
				Email:       "jane.smith@example.com",
			},
		},
		{
			"Missing fields",
			&testutil.NewMemberParams{},
			&models.Member{ID: "", UserID: "", DisplayName: "", Role: "", Email: ""},
		},
	}

	for _, tc := range testCases {
		graphMember := testutil.NewGraphMember(tc.params)
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

func TestMapGraphChat(t *testing.T) {
	type testCase struct {
		name   string
		params *testutil.NewChatParams
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
			&testutil.NewChatParams{
				ID:       util.Ptr("chat-id"),
				Type:     util.Ptr(msmodels.ONEONONE_CHATTYPE),
				IsHidden: util.Ptr(false),
				Topic:    nil,
			},
			&models.Chat{ID: "chat-id", Type: models.ChatTypeOneOnOne, IsHidden: false, Topic: nil},
		},
		{
			"Group chat with topic",
			&testutil.NewChatParams{
				ID:       util.Ptr("chat-id"),
				Type:     util.Ptr(msmodels.GROUP_CHATTYPE),
				IsHidden: util.Ptr(true),
				Topic:    util.Ptr("Project Discussion"),
			},
			&models.Chat{
				ID:       "chat-id",
				Type:     models.ChatTypeGroup,
				IsHidden: true,
				Topic:    util.Ptr("Project Discussion"),
			},
		},
		{
			"Missing fields",
			&testutil.NewChatParams{},
			&models.Chat{ID: "", Type: "", IsHidden: false, Topic: nil},
		},
	}

	for _, tc := range testCases {
		graphChat := testutil.NewGraphChat(tc.params)
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
