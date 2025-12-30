package resolver

import (
	"testing"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/internal/testutil"
	"github.com/pzsp-teams/lib/internal/util"
	"github.com/stretchr/testify/require"
)

func TestResolveTeamIDByName(t *testing.T) {
	type testCase struct {
		name        string
		setupTeams  func() msmodels.TeamCollectionResponseable
		teamName    string
		expectedID  string
		expectError bool
		errorType   error
	}

	testCases := []testCase{
		{
			name: "No teams available",
			setupTeams: func() msmodels.TeamCollectionResponseable {
				col := msmodels.NewTeamCollectionResponse()
				col.SetValue(nil)
				return col
			},
			teamName:    "Alpha",
			expectError: true,
			errorType:   &resourcesNotAvailableError{},
		},
		{
			name: "No match",
			setupTeams: func() msmodels.TeamCollectionResponseable {
				t1 := testutil.NewGraphTeam(
					&testutil.NewTeamParams{ID: util.Ptr("1"), DisplayName: util.Ptr("Alpha")},
				)
				col := testutil.NewTeamCollection(t1)
				return col
			},
			teamName:    "Beta",
			expectError: true,
			errorType:   &resourceNotFoundError{},
		},
		{
			name: "Single match",
			setupTeams: func() msmodels.TeamCollectionResponseable {
				t1 := testutil.NewGraphTeam(
					&testutil.NewTeamParams{ID: util.Ptr("1"), DisplayName: util.Ptr("Alpha")},
				)
				col := testutil.NewTeamCollection(t1)
				return col
			},
			teamName:   "Alpha",
			expectedID: "1",
		},
		{
			name: "Multiple matches",
			setupTeams: func() msmodels.TeamCollectionResponseable {
				t1 := testutil.NewGraphTeam(
					&testutil.NewTeamParams{ID: util.Ptr("1"), DisplayName: util.Ptr("Alpha")},
				)
				t2 := testutil.NewGraphTeam(
					&testutil.NewTeamParams{ID: util.Ptr("2"), DisplayName: util.Ptr("Alpha")},
				)
				col := testutil.NewTeamCollection(t1, t2)
				return col
			},
			teamName:    "Alpha",
			expectError: true,
			errorType:   &resourceAmbiguousError{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			teams := tc.setupTeams()
			id, err := resolveTeamIDByName(teams, tc.teamName)

			if tc.expectError {
				require.Error(t, err)
				require.ErrorAs(t, err, &tc.errorType)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.expectedID, id, "resolved team ID does not match expected")
		})
	}
}

func TestResolveChannelIDByName(t *testing.T) {
	type testCase struct {
		name          string
		setupChannels func() msmodels.ChannelCollectionResponseable
		teamID        string
		channelName   string
		expectedID    string
		expectError   bool
		errorType     error
	}

	testCases := []testCase{
		{
			name: "No channels available",
			setupChannels: func() msmodels.ChannelCollectionResponseable {
				col := msmodels.NewChannelCollectionResponse()
				col.SetValue(nil)
				return col
			},
			teamID:      "team-1",
			channelName: "Alpha",
			expectError: true,
			errorType:   &resourcesNotAvailableError{},
		},
		{
			name: "No match",
			setupChannels: func() msmodels.ChannelCollectionResponseable {
				ch1 := testutil.NewGraphChannel(
					&testutil.NewChannelParams{ID: util.Ptr("c1"), Name: util.Ptr("Alpha")},
				)
				col := testutil.NewChannelCollection(ch1)
				return col
			},
			teamID:      "team-1",
			channelName: "Beta",
			expectError: true,
			errorType:   &resourceNotFoundError{},
		},
		{
			name: "Single match",
			setupChannels: func() msmodels.ChannelCollectionResponseable {
				ch1 := testutil.NewGraphChannel(
					&testutil.NewChannelParams{ID: util.Ptr("c1"), Name: util.Ptr("Alpha")},
				)
				col := testutil.NewChannelCollection(ch1)
				return col
			},
			teamID:      "team-1",
			channelName: "Alpha",
			expectedID:  "c1",
		},
		{
			name: "Multiple matches",
			setupChannels: func() msmodels.ChannelCollectionResponseable {
				ch1 := testutil.NewGraphChannel(
					&testutil.NewChannelParams{ID: util.Ptr("c1"), Name: util.Ptr("Alpha")},
				)
				ch2 := testutil.NewGraphChannel(
					&testutil.NewChannelParams{ID: util.Ptr("c2"), Name: util.Ptr("Alpha")},
				)
				col := testutil.NewChannelCollection(ch1, ch2)
				return col
			},
			teamID:      "team-1",
			channelName: "Alpha",
			expectError: true,
			errorType:   &resourceAmbiguousError{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			channels := tc.setupChannels()
			id, err := resolveChannelIDByName(channels, tc.teamID, tc.channelName)

			if tc.expectError {
				require.Error(t, err)
				require.ErrorAs(t, err, &tc.errorType)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.expectedID, id, "resolved channel ID does not match expected")
		})
	}
}

func TestResolveOneonOneChatIDByUserRef(t *testing.T) {
	type testCase struct {
		name        string
		setupChats  func() msmodels.ChatCollectionResponseable
		userRef     string
		expectedID  string
		expectError bool
		errorType   error
	}

	testCases := []testCase{
		{
			name: "No chats available",
			setupChats: func() msmodels.ChatCollectionResponseable {
				col := msmodels.NewChatCollectionResponse()
				col.SetValue(nil)
				return col
			},
			userRef:     "usr-1",
			expectError: true,
			errorType:   &resourcesNotAvailableError{},
		},
		{
			name: "No match",
			setupChats: func() msmodels.ChatCollectionResponseable {
				m := testutil.NewGraphMember(&testutil.NewMemberParams{
					ID:     util.Ptr("m-1"),
					UserID: util.Ptr("usr-1"),
					Email:  util.Ptr("jane@example.com"),
				})
				chat := testutil.NewGraphChat(&testutil.NewChatParams{
					ID:      util.Ptr("chat-1"),
					Type:    util.Ptr(msmodels.ONEONONE_CHATTYPE),
					Members: []msmodels.ConversationMemberable{m},
				})
				col := testutil.NewChatCollection(chat)
				return col
			},
			userRef:     "other-user",
			expectError: true,
			errorType:   &resourceNotFoundError{},
		},
		{
			name: "ID match",
			setupChats: func() msmodels.ChatCollectionResponseable {
				m := testutil.NewGraphMember(&testutil.NewMemberParams{
					ID:     util.Ptr("m-1"),
					UserID: util.Ptr("usr-1"),
					Email:  util.Ptr("jane@example.com"),
				})
				chat := testutil.NewGraphChat(&testutil.NewChatParams{
					ID:      util.Ptr("chat-1"),
					Type:    util.Ptr(msmodels.ONEONONE_CHATTYPE),
					Members: []msmodels.ConversationMemberable{m},
				})
				col := testutil.NewChatCollection(chat)
				return col
			},
			userRef:    "usr-1",
			expectedID: "chat-1",
		},
		{
			name: "Email match",
			setupChats: func() msmodels.ChatCollectionResponseable {
				m := testutil.NewGraphMember(&testutil.NewMemberParams{
					ID:     util.Ptr("m-1"),
					UserID: util.Ptr("usr-1"),
					Email:  util.Ptr("jane@example.com"),
				})
				chat := testutil.NewGraphChat(&testutil.NewChatParams{
					ID:      util.Ptr("chat-1"),
					Type:    util.Ptr(msmodels.ONEONONE_CHATTYPE),
					Members: []msmodels.ConversationMemberable{m},
				})
				col := testutil.NewChatCollection(chat)
				return col
			},
			userRef:    "jane@example.com",
			expectedID: "chat-1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			chats := tc.setupChats()
			id, err := resolveOneOnOneChatIDByUserRef(chats, tc.userRef)
			if tc.expectError {
				require.Error(t, err)
				require.ErrorAs(t, err, &tc.errorType)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.expectedID, id, "resolved chat ID does not match expected")
		})
	}
}

func TestResolveGroupChatID(t *testing.T) {
	type testCase struct {
		name        string
		setupChats  func() msmodels.ChatCollectionResponseable
		topic       string
		expectedID  string
		expectError bool
		errorType   error
	}

	testCases := []testCase{
		{
			name: "No chats available",
			setupChats: func() msmodels.ChatCollectionResponseable {
				col := msmodels.NewChatCollectionResponse()
				col.SetValue(nil)
				return col
			},
			topic:       "Project X",
			expectError: true,
			errorType:   &resourcesNotAvailableError{},
		},
		{
			name: "No match",
			setupChats: func() msmodels.ChatCollectionResponseable {
				chat := testutil.NewGraphChat(&testutil.NewChatParams{
					ID:    util.Ptr("chat-1"),
					Type:  util.Ptr(msmodels.GROUP_CHATTYPE),
					Topic: util.Ptr("Project A"),
				})
				col := testutil.NewChatCollection(chat)
				return col
			},
			topic:       "Project X",
			expectError: true,
			errorType:   &resourceNotFoundError{},
		},
		{
			name: "Single match",
			setupChats: func() msmodels.ChatCollectionResponseable {
				chat := testutil.NewGraphChat(&testutil.NewChatParams{
					ID:    util.Ptr("chat-1"),
					Type:  util.Ptr(msmodels.GROUP_CHATTYPE),
					Topic: util.Ptr("Project X"),
				})
				col := testutil.NewChatCollection(chat)
				return col
			},
			topic:      "Project X",
			expectedID: "chat-1",
		},
		{
			name: "Multiple matches",
			setupChats: func() msmodels.ChatCollectionResponseable {
				chat1 := testutil.NewGraphChat(&testutil.NewChatParams{
					ID:    util.Ptr("chat-1"),
					Type:  util.Ptr(msmodels.GROUP_CHATTYPE),
					Topic: util.Ptr("Project X"),
				})
				chat2 := testutil.NewGraphChat(&testutil.NewChatParams{
					ID:    util.Ptr("chat-2"),
					Type:  util.Ptr(msmodels.GROUP_CHATTYPE),
					Topic: util.Ptr("Project X"),
				})
				col := testutil.NewChatCollection(chat1, chat2)
				return col
			},
			topic:       "Project X",
			expectError: true,
			errorType:   &resourceAmbiguousError{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			chats := tc.setupChats()
			id, err := resolveGroupChatIDByTopic(chats, tc.topic)

			if tc.expectError {
				require.Error(t, err)
				require.ErrorAs(t, err, &tc.errorType)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.expectedID, id, "resolved chat ID does not match expected")
		})
	}
}

func TestResolveMemberID(t *testing.T) {
	type testCase struct {
		name         string
		setupMembers func() msmodels.ConversationMemberCollectionResponseable
		userRef      string
		expectedID   string
		expectError  bool
		errorType    error
	}

	testCases := []testCase{
		{
			name: "No members available",
			setupMembers: func() msmodels.ConversationMemberCollectionResponseable {
				col := msmodels.NewConversationMemberCollectionResponse()
				col.SetValue(nil)
				return col
			},
			userRef:     "any",
			expectError: true,
			errorType:   &resourcesNotAvailableError{},
		},
		{
			name: "No match",
			setupMembers: func() msmodels.ConversationMemberCollectionResponseable {
				m := testutil.NewGraphMember(&testutil.NewMemberParams{
					ID:     util.Ptr("m-1"),
					UserID: util.Ptr("usr-1"),
					Email:  util.Ptr("jane@emxaple.com"),
				})
				col := testutil.NewMemberCollection(m)
				return col
			},
			userRef:     "missing",
			expectError: true,
			errorType:   &resourceNotFoundError{},
		},
		{
			name: "UserID match",
			setupMembers: func() msmodels.ConversationMemberCollectionResponseable {
				m := testutil.NewGraphMember(&testutil.NewMemberParams{
					ID:     util.Ptr("m-1"),
					UserID: util.Ptr("usr-1"),
					Email:  util.Ptr("jane@example.com"),
				})
				col := testutil.NewMemberCollection(m)
				return col
			},
			userRef:    "usr-1",
			expectedID: "m-1",
		},
		{
			name: "UserEmail match",
			setupMembers: func() msmodels.ConversationMemberCollectionResponseable {
				m := testutil.NewGraphMember(&testutil.NewMemberParams{
					ID:     util.Ptr("m-1"),
					UserID: util.Ptr("usr-1"),
					Email:  util.Ptr("jane@example.com"),
				})
				col := testutil.NewMemberCollection(m)
				return col
			},
			userRef:    "jane@example.com",
			expectedID: "m-1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			members := tc.setupMembers()
			id, err := resolveMemberID(members, tc.userRef)

			if tc.expectError {
				require.Error(t, err)
				require.ErrorAs(t, err, &tc.errorType)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.expectedID, id, "resolved member ID does not match expected")
		})
	}
}
