package mentions

import (
	"strings"
	"testing"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/models"
	"github.com/stretchr/testify/require"
)

const (
	okGUID   = "00000000-0000-0000-0000-000000000001"
	okGUID2  = "00000000-0000-0000-0000-000000000002"
	okTeamID = "00000000-0000-0000-0000-0000000000aa"
	okThread = "19:thread@thread.tacv2"
)

func requireMention(t *testing.T, m msmodels.ChatMessageMentionable) *msmodels.ChatMessageMention {
	t.Helper()
	cm, ok := m.(*msmodels.ChatMessageMention)
	require.True(t, ok, "expected *msmodels.ChatMessageMention, got %T", m)
	return cm
}

func requireMentionBasics(t *testing.T, m *msmodels.ChatMessageMention, id int32, text string) {
	t.Helper()
	require.NotNil(t, m.GetId())
	require.Equal(t, id, *m.GetId())
	require.NotNil(t, m.GetMentionText())
	require.Equal(t, text, *m.GetMentionText())
}

func requireMentionedUser(t *testing.T, m *msmodels.ChatMessageMention, wantID, wantDisplay string) msmodels.Identityable {
	t.Helper()
	mentioned := m.GetMentioned()
	require.NotNil(t, mentioned)
	u := mentioned.GetUser()
	require.NotNil(t, u)

	require.NotNil(t, u.GetId())
	require.Equal(t, wantID, *u.GetId())
	require.NotNil(t, u.GetDisplayName())
	require.Equal(t, wantDisplay, *u.GetDisplayName())

	return u
}

func requireMentionedConversation(
	t *testing.T,
	m *msmodels.ChatMessageMention,
	wantID string,
	wantType msmodels.TeamworkConversationIdentityType,
) {
	t.Helper()

	mentioned := m.GetMentioned()
	require.NotNil(t, mentioned)

	conv := mentioned.GetConversation()
	require.NotNil(t, conv)

	require.NotNil(t, conv.GetId())
	require.Equal(t, wantID, *conv.GetId())

	require.NotNil(t, conv.GetConversationIdentityType())
	require.Equal(t, wantType, *conv.GetConversationIdentityType())
}

func Test_mapMentions(t *testing.T) {
	t.Parallel()

	t.Run("empty input -> nil, nil", func(t *testing.T) {
		t.Parallel()

		got, err := mapMentions(nil)
		require.NoError(t, err)
		require.Nil(t, got)

		got, err = mapMentions([]models.Mention{})
		require.NoError(t, err)
		require.Nil(t, got)
	})

	t.Run("maps user mention", func(t *testing.T) {
		t.Parallel()

		in := []models.Mention{
			{Kind: models.MentionUser, AtID: 0, Text: "Alice", TargetID: okGUID},
		}

		got, err := mapMentions(in)
		require.NoError(t, err)
		require.Len(t, got, 1)

		m := requireMention(t, got[0])
		requireMentionBasics(t, m, 0, "Alice")

		u := requireMentionedUser(t, m, okGUID, "Alice")
		ad := u.GetAdditionalData()
		require.NotNil(t, ad)
		require.Equal(t, "aadUser", ad["userIdentityType"])
	})

	t.Run("maps conversation mentions (channel/team/everyone)", func(t *testing.T) {
		t.Parallel()

		in := []models.Mention{
			{Kind: models.MentionChannel, AtID: 0, Text: "Channel", TargetID: okThread},
			{Kind: models.MentionTeam, AtID: 1, Text: "Team", TargetID: okGUID},
			{Kind: models.MentionEveryone, AtID: 2, Text: "Everyone", TargetID: okThread},
		}

		got, err := mapMentions(in)
		require.NoError(t, err)
		require.Len(t, got, 3)

		m0 := requireMention(t, got[0])
		requireMentionedConversation(t, m0, okThread, msmodels.CHANNEL_TEAMWORKCONVERSATIONIDENTITYTYPE)

		m1 := requireMention(t, got[1])
		requireMentionedConversation(t, m1, okGUID, msmodels.TEAM_TEAMWORKCONVERSATIONIDENTITYTYPE)

		m2 := requireMention(t, got[2])
		requireMentionedConversation(t, m2, okThread, msmodels.CHAT_TEAMWORKCONVERSATIONIDENTITYTYPE)
	})

	t.Run("error is wrapped with index", func(t *testing.T) {
		t.Parallel()

		in := []models.Mention{
			{Kind: models.MentionUser, AtID: 0, Text: "Alice", TargetID: "not-a-guid"},
		}

		_, err := mapMentions(in)
		require.Error(t, err)
		require.Contains(t, err.Error(), "mention[0]:")
	})
}

func Test_validateAtTags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body *models.MessageBody
		want string
	}{
		{
			name: "nil body -> ok",
			body: nil,
			want: "",
		},
		{
			name: "non-HTML -> ok (even if content contains <at>)",
			body: &models.MessageBody{
				ContentType: models.MessageContentTypeText,
				Content:     `<at id="0">Alice</at>`,
				Mentions:    []models.Mention{{Kind: models.MentionUser, AtID: 0, Text: "Alice", TargetID: okGUID}},
			},
			want: "",
		},
		{
			name: "content contains <at but mentions empty -> error",
			body: &models.MessageBody{
				ContentType: models.MessageContentTypeHTML,
				Content:     `<at id="0">Alice</at>`,
				Mentions:    nil,
			},
			want: "mentions list is empty",
		},
		{
			name: "duplicate AtID -> error",
			body: &models.MessageBody{
				ContentType: models.MessageContentTypeHTML,
				Content:     `<at id="0">A</at> <at id="0">B</at>`,
				Mentions: []models.Mention{
					{Kind: models.MentionUser, AtID: 0, Text: "A", TargetID: okGUID},
					{Kind: models.MentionUser, AtID: 0, Text: "B", TargetID: okGUID2},
				},
			},
			want: "duplicate at-id 0",
		},
		{
			name: "missing specific <at id=\"X\" -> error",
			body: &models.MessageBody{
				ContentType: models.MessageContentTypeHTML,
				Content:     `<at id="0">A</at>`,
				Mentions: []models.Mention{
					{Kind: models.MentionUser, AtID: 0, Text: "A", TargetID: okGUID},
					{Kind: models.MentionUser, AtID: 1, Text: "B", TargetID: okGUID2},
				},
			},
			want: `missing <at id="1"`,
		},
		{
			name: "happy path -> ok",
			body: &models.MessageBody{
				ContentType: models.MessageContentTypeHTML,
				Content:     `Hello <at id="0">Alice</at> and <at id="1">Bob</at>`,
				Mentions: []models.Mention{
					{Kind: models.MentionUser, AtID: 0, Text: "Alice", TargetID: okGUID},
					{Kind: models.MentionUser, AtID: 1, Text: "Bob", TargetID: okGUID2},
				},
			},
			want: "",
		},
		{
			name: "case sensitive for id attribute -> error",
			body: &models.MessageBody{
				ContentType: models.MessageContentTypeHTML,
				Content:     `<div><at ID="0">Alice</at></div>`,
				Mentions: []models.Mention{
					{Kind: models.MentionUser, AtID: 0, Text: "Alice", TargetID: okGUID},
				},
			},
			want: `missing <at id="0"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateAtTags(tt.body)
			if tt.want == "" {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.want)
		})
	}
}

func Test_mapToGraphMention(t *testing.T) {
	t.Parallel()

	t.Run("validation errors", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name string
			in   models.Mention
			want string
		}{
			{"missing Text", models.Mention{Kind: models.MentionUser, AtID: 0, Text: "", TargetID: okGUID}, "missing Text"},
			{"invalid AtID", models.Mention{Kind: models.MentionUser, AtID: -1, Text: "A", TargetID: okGUID}, "invalid AtID"},
			{"missing TargetID", models.Mention{Kind: models.MentionUser, AtID: 0, Text: "A", TargetID: ""}, "missing TargetID"},
			{"unsupported kind", models.Mention{Kind: models.MentionKind("weird"), AtID: 0, Text: "A", TargetID: okGUID}, "unsupported MentionKind"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				_, err := mapToGraphMention(tt.in)
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.want)
			})
		}
	})

	t.Run("user mention -> sets mentioned.user", func(t *testing.T) {
		t.Parallel()

		in := models.Mention{Kind: models.MentionUser, AtID: 7, Text: "Alice", TargetID: okGUID}

		out, err := mapToGraphMention(in)
		require.NoError(t, err)

		m := requireMention(t, out)
		requireMentionBasics(t, m, 7, "Alice")

		u := requireMentionedUser(t, m, okGUID, "Alice")
		ad := u.GetAdditionalData()
		require.NotNil(t, ad)
		require.Equal(t, "aadUser", ad["userIdentityType"])
	})

	t.Run("conversation mention kinds -> set mentioned.conversation (type-specific)", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name     string
			kind     models.MentionKind
			targetID string
			wantType msmodels.TeamworkConversationIdentityType
		}{
			{"channel", models.MentionChannel, okThread, msmodels.CHANNEL_TEAMWORKCONVERSATIONIDENTITYTYPE},
			{"team", models.MentionTeam, okTeamID, msmodels.TEAM_TEAMWORKCONVERSATIONIDENTITYTYPE},
			{"everyone", models.MentionEveryone, okThread, msmodels.CHAT_TEAMWORKCONVERSATIONIDENTITYTYPE},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				in := models.Mention{Kind: tt.kind, AtID: 1, Text: "X", TargetID: tt.targetID}
				out, err := mapToGraphMention(in)
				require.NoError(t, err)

				m := requireMention(t, out)
				requireMentionedConversation(t, m, tt.targetID, tt.wantType)
			})
		}
	})

	t.Run("conversation mention invalid TargetID -> error", func(t *testing.T) {
		t.Parallel()

		_, err := mapToGraphMention(models.Mention{
			Kind:     models.MentionTeam,
			AtID:     0,
			Text:     "Team",
			TargetID: "not-a-guid",
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid TargetID for team mention")
	})
}

func Test_buildUserMentioned(t *testing.T) {
	t.Parallel()

	t.Run("invalid targetID -> error", func(t *testing.T) {
		t.Parallel()

		_, err := buildUserMentioned("not-a-guid", "Alice")
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid TargetID for user mention")
	})

	t.Run("sets fields", func(t *testing.T) {
		t.Parallel()

		u, err := buildUserMentioned(okGUID, "Alice")
		require.NoError(t, err)

		require.NotNil(t, u.GetId())
		require.Equal(t, okGUID, *u.GetId())
		require.NotNil(t, u.GetDisplayName())
		require.Equal(t, "Alice", *u.GetDisplayName())

		ad := u.GetAdditionalData()
		require.NotNil(t, ad)
		require.Equal(t, "aadUser", ad["userIdentityType"])
	})
}

func Test_buildConversationMentioned(t *testing.T) {
	t.Parallel()

	t.Run("validate returns false -> error", func(t *testing.T) {
		t.Parallel()

		validate := func(string) bool { return false }
		_, err := buildConversationMentioned("x", "Name", msmodels.CHAT_TEAMWORKCONVERSATIONIDENTITYTYPE, validate, "label")
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid TargetID for label")
	})

	t.Run("sets fields and type", func(t *testing.T) {
		t.Parallel()

		validate := func(string) bool { return true }
		typ := msmodels.CHANNEL_TEAMWORKCONVERSATIONIDENTITYTYPE

		conv, err := buildConversationMentioned("conv-id", "Conv", typ, validate, "channel mention")
		require.NoError(t, err)

		require.NotNil(t, conv.GetId())
		require.Equal(t, "conv-id", *conv.GetId())
		require.NotNil(t, conv.GetDisplayName())
		require.Equal(t, "Conv", *conv.GetDisplayName())
		require.NotNil(t, conv.GetConversationIdentityType())
		require.Equal(t, typ, *conv.GetConversationIdentityType())
	})
}

func Test_PrepareMentions(t *testing.T) {
	t.Parallel()

	t.Run("no mentions -> nil, nil", func(t *testing.T) {
		t.Parallel()

		body := &models.MessageBody{
			Content:     "hello",
			ContentType: models.MessageContentTypeText,
			Mentions:    nil,
		}

		got, err := PrepareMentions(body)
		require.NoError(t, err)
		require.Nil(t, got)
	})

	t.Run("mentions with non-HTML -> error", func(t *testing.T) {
		t.Parallel()

		body := &models.MessageBody{
			Content:     "hello",
			ContentType: models.MessageContentTypeText,
			Mentions: []models.Mention{
				{Kind: models.MentionUser, AtID: 0, Text: "Alice", TargetID: okGUID},
			},
		}

		_, err := PrepareMentions(body)
		require.Error(t, err)
		require.Contains(t, err.Error(), "mentions can only be used with HTML content type")
	})

	t.Run("at-tags present but mentions empty -> error (nil or empty slice)", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			name     string
			mentions []models.Mention
		}{
			{"nil", nil},
			{"empty", []models.Mention{}},
		}

		for _, tt := range cases {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				body := &models.MessageBody{
					Content:     `<div>Hello <at id="0">X</at></div>`,
					ContentType: models.MessageContentTypeHTML,
					Mentions:    tt.mentions,
				}

				_, err := PrepareMentions(body)
				require.Error(t, err)
				require.Contains(t, err.Error(), "mentions list is empty")
			})
		}
	})

	t.Run("missing <at id=\"X\" tag -> error", func(t *testing.T) {
		t.Parallel()

		body := &models.MessageBody{
			Content:     `<div>Hello</div>`,
			ContentType: models.MessageContentTypeHTML,
			Mentions: []models.Mention{
				{Kind: models.MentionUser, AtID: 0, Text: "Alice", TargetID: okGUID},
			},
		}

		_, err := PrepareMentions(body)
		require.Error(t, err)
		require.Contains(t, err.Error(), `missing <at id="0"`)
	})

	t.Run("duplicate AtID -> error", func(t *testing.T) {
		t.Parallel()

		body := &models.MessageBody{
			Content:     `<div><at id="0">Alice</at> and <at id="0">Bob</at></div>`,
			ContentType: models.MessageContentTypeHTML,
			Mentions: []models.Mention{
				{Kind: models.MentionUser, AtID: 0, Text: "Alice", TargetID: okGUID},
				{Kind: models.MentionUser, AtID: 0, Text: "Bob", TargetID: okGUID2},
			},
		}

		_, err := PrepareMentions(body)
		require.Error(t, err)
		require.Contains(t, err.Error(), "duplicate at-id 0")
	})

	t.Run("mapMentions wraps index error", func(t *testing.T) {
		t.Parallel()

		body := &models.MessageBody{
			Content:     `<div><at id="-1">Alice</at></div>`,
			ContentType: models.MessageContentTypeHTML,
			Mentions: []models.Mention{
				{Kind: models.MentionUser, AtID: -1, Text: "Alice", TargetID: okGUID},
			},
		}

		_, err := PrepareMentions(body)
		require.Error(t, err)
		require.Contains(t, err.Error(), "mention[0]:")
		require.Contains(t, err.Error(), "invalid AtID")
	})

	t.Run("success paths (user/channel/team/everyone)", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name     string
			mention  models.Mention
			content  string
			check    func(t *testing.T, out []msmodels.ChatMessageMentionable)
			wantKind models.MentionKind
		}{
			{
				name: "user",
				mention: models.Mention{
					Kind: models.MentionUser, AtID: 0, Text: "Alice", TargetID: okGUID,
				},
				content: `<div>Hello <at id="0">Alice</at></div>`,
				check: func(t *testing.T, out []msmodels.ChatMessageMentionable) {
					require.Len(t, out, 1)
					m := requireMention(t, out[0])
					requireMentionBasics(t, m, 0, "Alice")
					u := requireMentionedUser(t, m, okGUID, "Alice")
					require.Equal(t, "aadUser", u.GetAdditionalData()["userIdentityType"])
				},
			},
			{
				name: "channel",
				mention: models.Mention{
					Kind: models.MentionChannel, AtID: 0, Text: "General", TargetID: okThread,
				},
				content: `<div>Hello <at id="0">General</at></div>`,
				check: func(t *testing.T, out []msmodels.ChatMessageMentionable) {
					require.Len(t, out, 1)
					m := requireMention(t, out[0])
					requireMentionedConversation(t, m, okThread, msmodels.CHANNEL_TEAMWORKCONVERSATIONIDENTITYTYPE)
				},
			},
			{
				name: "team",
				mention: models.Mention{
					Kind: models.MentionTeam, AtID: 0, Text: "My Team", TargetID: okTeamID,
				},
				content: `<div>Hello <at id="0">My Team</at></div>`,
				check: func(t *testing.T, out []msmodels.ChatMessageMentionable) {
					require.Len(t, out, 1)
					m := requireMention(t, out[0])
					requireMentionedConversation(t, m, okTeamID, msmodels.TEAM_TEAMWORKCONVERSATIONIDENTITYTYPE)
				},
			},
			{
				name: "everyone",
				mention: models.Mention{
					Kind: models.MentionEveryone, AtID: 0, Text: "Everyone", TargetID: okThread,
				},
				content: `<div>Hello <at id="0">Everyone</at></div>`,
				check: func(t *testing.T, out []msmodels.ChatMessageMentionable) {
					require.Len(t, out, 1)
					m := requireMention(t, out[0])
					requireMentionedConversation(t, m, okThread, msmodels.CHAT_TEAMWORKCONVERSATIONIDENTITYTYPE)
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				body := &models.MessageBody{
					Content:     tt.content,
					ContentType: models.MessageContentTypeHTML,
					Mentions:    []models.Mention{tt.mention},
				}

				got, err := PrepareMentions(body)
				require.NoError(t, err)
				tt.check(t, got)
			})
		}
	})

	t.Run("unsupported mention kind -> error", func(t *testing.T) {
		t.Parallel()

		body := &models.MessageBody{
			Content:     `<div>Hello <at id="0">X</at></div>`,
			ContentType: models.MessageContentTypeHTML,
			Mentions: []models.Mention{
				{Kind: models.MentionKind("weird"), AtID: 0, Text: "X", TargetID: "whatever"},
			},
		}

		_, err := PrepareMentions(body)
		require.Error(t, err)
		require.Contains(t, err.Error(), "unsupported MentionKind")
	})

	t.Run("error messages are stable enough (spot-check)", func(t *testing.T) {
		t.Parallel()

		body := &models.MessageBody{
			Content:     `<div>Hello <at id="0">Alice</at></div>`,
			ContentType: models.MessageContentTypeHTML,
			Mentions: []models.Mention{
				{Kind: models.MentionUser, AtID: 0, Text: "Alice", TargetID: "not-a-guid"},
			},
		}

		_, err := PrepareMentions(body)
		require.Error(t, err)

		require.True(t, strings.Contains(err.Error(), "mention[0]:"), err.Error())
	})
}
