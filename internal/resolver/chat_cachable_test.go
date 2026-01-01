package resolver

import (
	"context"
	"testing"

	gomock "github.com/golang/mock/gomock"
	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/internal/cacher"
	sender "github.com/pzsp-teams/lib/internal/sender"
	"github.com/pzsp-teams/lib/internal/testutil"
	"github.com/pzsp-teams/lib/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChatResolverCachable_ResolveOneOnOneRef(t *testing.T) {
	type testCase struct {
		name         string
		chatRef      string
		cacheEnabled bool
		setupMocks   func(api *testutil.MockChatAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner)
		expectedID   string
		expectError  bool
	}

	testCases := []testCase{
		{
			name:         "Empty one-on-one chat reference",
			chatRef:      "   ",
			cacheEnabled: true,
			setupMocks: func(api *testutil.MockChatAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner) {
				api.EXPECT().ListChats(gomock.Any(), gomock.Any()).Times(0)
				c.EXPECT().Get(gomock.Any()).Times(0)
				tr.EXPECT().Run(gomock.Any()).Times(0)
			},
			expectError: true,
		},
		{
			name:         "Chat ID short circuit",
			chatRef:      "19:3A8b081ef6-4792-4def-b2c9-c363a1bf41d5_877192bd-9183-47d3-a74c-8aa0426716cf@unq.gbl.spaces",
			cacheEnabled: true,
			setupMocks: func(api *testutil.MockChatAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner) {
				api.EXPECT().ListChats(gomock.Any(), gomock.Any()).Times(0)
				c.EXPECT().Get(gomock.Any()).Times(0)
				tr.EXPECT().Run(gomock.Any()).Times(0)
			},
			expectedID: "19:3A8b081ef6-4792-4def-b2c9-c363a1bf41d5_877192bd-9183-47d3-a74c-8aa0426716cf@unq.gbl.spaces",
		},
		{
			name:         "Cache single hit",
			chatRef:      "jane@example.com",
			cacheEnabled: true,
			setupMocks: func(api *testutil.MockChatAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner) {
				api.EXPECT().ListChats(gomock.Any(), gomock.Any()).Times(0)
				c.EXPECT().
					Get(cacher.NewOneOnOneChatKey("jane@example.com", nil)).
					Return([]string{"chat-id-123"}, true, nil).
					Times(1)
				tr.EXPECT().Run(gomock.Any()).Times(0)
			},
			expectedID: "chat-id-123",
		},
		{
			name:         "Cache miss uses API and caches result",
			chatRef:      "jane@example.com",
			cacheEnabled: true,
			setupMocks: func(api *testutil.MockChatAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner) {
				m := testutil.NewGraphMember(
					&testutil.NewMemberParams{
						Email: util.Ptr("jane@example.com"),
					},
				)
				chat := testutil.NewGraphChat(
					&testutil.NewChatParams{
						ID:      util.Ptr("chat-id-123"),
						Type:    util.Ptr(msmodels.ONEONONE_CHATTYPE),
						Members: []msmodels.ConversationMemberable{m},
					},
				)
				collection := testutil.NewChatCollection(chat)
				api.EXPECT().ListChats(gomock.Any(), gomock.Any()).Return(collection, nil).Times(1)
				c.EXPECT().
					Get(cacher.NewOneOnOneChatKey("jane@example.com", nil)).
					Return(nil, false, nil).
					Times(1)
				tr.EXPECT().Run(gomock.Any()).Times(1)
			},
			expectedID: "chat-id-123",
		},
		{
			name:         "Cache disabled skips cache",
			chatRef:      "jane@example.com",
			cacheEnabled: false,
			setupMocks: func(api *testutil.MockChatAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner) {
				m := testutil.NewGraphMember(
					&testutil.NewMemberParams{
						Email: util.Ptr("jane@example.com"),
					},
				)
				chat := testutil.NewGraphChat(
					&testutil.NewChatParams{
						ID:      util.Ptr("chat-id-123"),
						Type:    util.Ptr(msmodels.ONEONONE_CHATTYPE),
						Members: []msmodels.ConversationMemberable{m},
					},
				)
				collection := testutil.NewChatCollection(chat)
				api.EXPECT().ListChats(gomock.Any(), gomock.Any()).Return(collection, nil).Times(1)
				c.EXPECT().Get(gomock.Any()).Times(0)
				tr.EXPECT().Run(gomock.Any()).Times(0)
			},
			expectedID: "chat-id-123",
		},
		{
			name:         "Fetch from API fails",
			chatRef:      "jane@example.com",
			cacheEnabled: true,
			setupMocks: func(api *testutil.MockChatAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner) {
				wantErr := &sender.RequestError{Message: "nope"}
				api.EXPECT().ListChats(gomock.Any(), gomock.Any()).Return(nil, wantErr).Times(1)
				c.EXPECT().Get(gomock.Any()).Return(nil, false, nil).Times(1)
				tr.EXPECT().Run(gomock.Any()).Times(0)
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			apiMock := testutil.NewMockChatAPI(ctrl)
			cacherMock := testutil.NewMockCacher(ctrl)
			taskRunnerMock := testutil.NewMockTaskRunner(ctrl)

			tc.setupMocks(apiMock, cacherMock, taskRunnerMock)

			var cacherArg *cacher.CacheHandler = nil
			if tc.cacheEnabled {
				cacherArg = &cacher.CacheHandler{Cacher: cacherMock, Runner: taskRunnerMock}
			}

			resolver := NewChatResolverCacheable(apiMock, cacherArg)

			ctx := context.Background()
			id, err := resolver.ResolveOneOnOneChatRefToID(ctx, tc.chatRef)

			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.expectedID, id)
		})
	}
}

func TestChatResolverCachable_ResolveGroupChatRefToID(t *testing.T) {
	type testCase struct {
		name         string
		chatRef      string
		cacheEnabled bool
		setupMocks   func(api *testutil.MockChatAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner)
		expectedID   string
		expectError  bool
	}

	testCases := []testCase{
		{
			name:         "Empty group chat reference",
			chatRef:      "   ",
			cacheEnabled: true,
			setupMocks: func(api *testutil.MockChatAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner) {
				api.EXPECT().ListChats(gomock.Any(), gomock.Any()).Times(0)
				c.EXPECT().Get(gomock.Any()).Times(0)
				tr.EXPECT().Run(gomock.Any()).Times(0)
			},
			expectError: true,
		},
		{
			name:         "Thread ID short circuit",
			chatRef:      "19:abc123@thread.tacv2",
			cacheEnabled: true,
			setupMocks: func(api *testutil.MockChatAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner) {
				api.EXPECT().ListChats(gomock.Any(), gomock.Any()).Times(0)
				c.EXPECT().Get(gomock.Any()).Times(0)
				tr.EXPECT().Run(gomock.Any()).Times(0)
			},
			expectedID: "19:abc123@thread.tacv2",
		},
		{
			name:         "Cache single hit",
			chatRef:      "My Topic",
			cacheEnabled: true,
			setupMocks: func(api *testutil.MockChatAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner) {
				api.EXPECT().ListChats(gomock.Any(), gomock.Any()).Times(0)
				c.EXPECT().
					Get(cacher.NewGroupChatKey("My Topic")).
					Return([]string{"c-id-123"}, true, nil).
					Times(1)
				tr.EXPECT().Run(gomock.Any()).Times(0)
			},
			expectedID: "c-id-123",
		},
		{
			name:         "Cache miss uses API and caches result",
			chatRef:      "My Topic",
			cacheEnabled: true,
			setupMocks: func(api *testutil.MockChatAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner) {
				chat := testutil.NewGraphChat(
					&testutil.NewChatParams{
						ID:    util.Ptr("gc-1"),
						Type:  util.Ptr(msmodels.GROUP_CHATTYPE),
						Topic: util.Ptr("My Topic"),
					},
				)
				collection := testutil.NewChatCollection(chat)
				api.EXPECT().ListChats(gomock.Any(), gomock.Any()).Return(collection, nil).Times(1)
				c.EXPECT().Get(cacher.NewGroupChatKey("My Topic")).Return(nil, false, nil).Times(1)
				tr.EXPECT().Run(gomock.Any()).Times(1)
			},
			expectedID: "gc-1",
		},
		{
			name:         "Cache disabled skips cache",
			chatRef:      "Topic",
			cacheEnabled: false,
			setupMocks: func(api *testutil.MockChatAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner) {
				chat := testutil.NewGraphChat(
					&testutil.NewChatParams{
						ID:    util.Ptr("gc-api"),
						Type:  util.Ptr(msmodels.GROUP_CHATTYPE),
						Topic: util.Ptr("Topic"),
					},
				)
				collection := testutil.NewChatCollection(chat)
				api.EXPECT().ListChats(gomock.Any(), gomock.Any()).Return(collection, nil).Times(1)
				c.EXPECT().Get(gomock.Any()).Times(0)
				tr.EXPECT().Run(gomock.Any()).Times(0)
			},
			expectedID: "gc-api",
		},
		{
			name:         "Fetch from API fails",
			chatRef:      "Topic",
			cacheEnabled: true,
			setupMocks: func(api *testutil.MockChatAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner) {
				apiErr := &sender.RequestError{Code: 500, Message: "api error"}
				api.EXPECT().ListChats(gomock.Any(), gomock.Any()).Return(nil, apiErr).Times(1)
				c.EXPECT().Get(cacher.NewGroupChatKey("Topic")).Return(nil, false, nil).Times(1)
				tr.EXPECT().Run(gomock.Any()).Times(0)
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			apiMock := testutil.NewMockChatAPI(ctrl)
			cacherMock := testutil.NewMockCacher(ctrl)
			taskRunnerMock := testutil.NewMockTaskRunner(ctrl)

			tc.setupMocks(apiMock, cacherMock, taskRunnerMock)

			var cacherArg *cacher.CacheHandler = nil
			if tc.cacheEnabled {
				cacherArg = &cacher.CacheHandler{Cacher: cacherMock, Runner: taskRunnerMock}
			}

			resolver := NewChatResolverCacheable(apiMock, cacherArg)

			ctx := context.Background()
			id, err := resolver.ResolveGroupChatRefToID(ctx, tc.chatRef)

			if tc.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedID, id)
		})
	}
}
