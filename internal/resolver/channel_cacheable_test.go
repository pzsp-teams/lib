package resolver

import (
	"context"
	"errors"
	"testing"

	"github.com/pzsp-teams/lib/internal/cacher"
	sender "github.com/pzsp-teams/lib/internal/sender"
	"github.com/pzsp-teams/lib/internal/testutil"
	"github.com/pzsp-teams/lib/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestChannelResolverCacheable_ResolveChannelRefToID(t *testing.T) {
	type testCase struct {
		name         string
		teamID       string
		channelRef   string
		cacheEnabled bool
		setupMocks   func(api *testutil.MockChannelAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner)
		expectedID   string
		expectError  bool
	}

	testCases := []testCase{
		{
			name:         "Empty channel reference",
			teamID:       "team-1",
			channelRef:   "   ",
			cacheEnabled: true,
			setupMocks: func(api *testutil.MockChannelAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner) {
				api.EXPECT().ListChannels(gomock.Any(), gomock.Any()).Times(0)
				c.EXPECT().Get(gomock.Any()).Times(0)
				tr.EXPECT().Run(gomock.Any()).Times(0)
			},
			expectError: true,
		},
		{
			name:         "Direct channel ID short circuit",
			teamID:       "team-1",
			channelRef:   "19:abc123@thread.tacv2",
			cacheEnabled: true,
			setupMocks: func(api *testutil.MockChannelAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner) {
				api.EXPECT().ListChannels(gomock.Any(), gomock.Any()).Times(0)
				c.EXPECT().Get(gomock.Any()).Times(0)
				tr.EXPECT().Run(gomock.Any()).Times(0)
			},
			expectedID: "19:abc123@thread.tacv2",
		},
		{
			name:         "Cache single hit",
			teamID:       "team-1",
			channelRef:   "General",
			cacheEnabled: true,
			setupMocks: func(api *testutil.MockChannelAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner) {
				api.EXPECT().ListChannels(gomock.Any(), gomock.Any()).Times(0)
				c.EXPECT().
					Get(cacher.NewChannelKey("team-1", "General")).
					Return([]string{"chan-id-123"}, true, nil).
					Times(1)
				tr.EXPECT().Run(gomock.Any()).Times(0)
			},
			expectedID: "chan-id-123",
		},
		{
			name:         "Cache miss uses API and caches result",
			teamID:       "team-42",
			channelRef:   "  General ",
			cacheEnabled: true,
			setupMocks: func(api *testutil.MockChannelAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner) {
				ch := testutil.NewGraphChannel(
					&testutil.NewChannelParams{
						ID:   util.Ptr("chan-id-xyz"),
						Name: util.Ptr("General"),
					},
				)
				collection := testutil.NewChannelCollection(ch)

				api.EXPECT().ListChannels(gomock.Any(), "team-42").Return(collection, nil).Times(1)
				c.EXPECT().
					Get(cacher.NewChannelKey("team-42", "General")).
					Return(nil, false, nil).
					Times(1)
				tr.EXPECT().Run(gomock.Any()).Times(1)
			},
			expectedID: "chan-id-xyz",
		},
		{
			name:         "Cache disabled skips cache",
			teamID:       "team-1",
			channelRef:   "General",
			cacheEnabled: false,
			setupMocks: func(api *testutil.MockChannelAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner) {
				ch := testutil.NewGraphChannel(
					&testutil.NewChannelParams{
						ID:   util.Ptr("chan-id-api"),
						Name: util.Ptr("General"),
					},
				)
				collection := testutil.NewChannelCollection(ch)

				api.EXPECT().ListChannels(gomock.Any(), "team-1").Return(collection, nil).Times(1)
				c.EXPECT().Get(gomock.Any()).Times(0)
				tr.EXPECT().Run(gomock.Any()).Times(0)
			},
			expectedID: "chan-id-api",
		},
		{
			name:         "Fetch from API fails",
			teamID:       "team-1",
			channelRef:   "General",
			cacheEnabled: true,
			setupMocks: func(api *testutil.MockChannelAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner) {
				wantErr := &sender.RequestError{Message: "boom"}

				api.EXPECT().ListChannels(gomock.Any(), "team-1").Return(nil, wantErr).Times(1)
				c.EXPECT().
					Get(cacher.NewChannelKey("team-1", "General")).
					Return(nil, false, nil).
					Times(1)
				tr.EXPECT().Run(gomock.Any()).Times(0)
			},
			expectError: true,
		},

		{
			name:         "Cache hit multiple IDs triggers invalidation and falls back to API",
			teamID:       "team-1",
			channelRef:   "General",
			cacheEnabled: true,
			setupMocks: func(api *testutil.MockChannelAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner) {
				ch := testutil.NewGraphChannel(
					&testutil.NewChannelParams{
						ID:   util.Ptr("chan-id-api"),
						Name: util.Ptr("General"),
					},
				)
				collection := testutil.NewChannelCollection(ch)

				c.EXPECT().
					Get(cacher.NewChannelKey("team-1", "General")).
					Return([]string{"id-1", "id-2"}, true, nil).
					Times(1)

				api.EXPECT().ListChannels(gomock.Any(), "team-1").Return(collection, nil).Times(1)

				tr.EXPECT().Run(gomock.Any()).Times(2)
			},
			expectedID: "chan-id-api",
		},
		{
			name:         "Cache Get error is ignored and resolver falls back to API",
			teamID:       "team-1",
			channelRef:   "General",
			cacheEnabled: true,
			setupMocks: func(api *testutil.MockChannelAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner) {
				ch := testutil.NewGraphChannel(
					&testutil.NewChannelParams{
						ID:   util.Ptr("chan-id-api"),
						Name: util.Ptr("General"),
					},
				)
				collection := testutil.NewChannelCollection(ch)

				c.EXPECT().
					Get(cacher.NewChannelKey("team-1", "General")).
					Return(nil, false, errors.New("cache down")).
					Times(1)

				api.EXPECT().ListChannels(gomock.Any(), "team-1").Return(collection, nil).Times(1)
				tr.EXPECT().Run(gomock.Any()).Times(1)
			},
			expectedID: "chan-id-api",
		},
		{
			name:         "Cache hit with wrong type is ignored and resolver falls back to API",
			teamID:       "team-1",
			channelRef:   "General",
			cacheEnabled: true,
			setupMocks: func(api *testutil.MockChannelAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner) {
				ch := testutil.NewGraphChannel(
					&testutil.NewChannelParams{
						ID:   util.Ptr("chan-id-api"),
						Name: util.Ptr("General"),
					},
				)
				collection := testutil.NewChannelCollection(ch)

				c.EXPECT().
					Get(cacher.NewChannelKey("team-1", "General")).
					Return(123, true, nil).
					Times(1)

				api.EXPECT().ListChannels(gomock.Any(), "team-1").Return(collection, nil).Times(1)
				tr.EXPECT().Run(gomock.Any()).Times(1)
			},
			expectedID: "chan-id-api",
		},
		{
			name:         "Cache hit with empty slice is ignored and resolver falls back to API",
			teamID:       "team-1",
			channelRef:   "General",
			cacheEnabled: true,
			setupMocks: func(api *testutil.MockChannelAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner) {
				ch := testutil.NewGraphChannel(
					&testutil.NewChannelParams{
						ID:   util.Ptr("chan-id-api"),
						Name: util.Ptr("General"),
					},
				)
				collection := testutil.NewChannelCollection(ch)

				c.EXPECT().
					Get(cacher.NewChannelKey("team-1", "General")).
					Return([]string{}, true, nil).
					Times(1)

				api.EXPECT().ListChannels(gomock.Any(), "team-1").Return(collection, nil).Times(1)
				tr.EXPECT().Run(gomock.Any()).Times(1)
			},
			expectedID: "chan-id-api",
		},
		{
			name:         "Extract fails (channel not found / ambiguous) - error returned and nothing is cached",
			teamID:       "team-1",
			channelRef:   "General",
			cacheEnabled: true,
			setupMocks: func(api *testutil.MockChannelAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner) {
				ch := testutil.NewGraphChannel(
					&testutil.NewChannelParams{
						ID:   util.Ptr("chan-id-other"),
						Name: util.Ptr("Other"),
					},
				)
				collection := testutil.NewChannelCollection(ch)

				c.EXPECT().
					Get(cacher.NewChannelKey("team-1", "General")).
					Return(nil, false, nil).
					Times(1)

				api.EXPECT().ListChannels(gomock.Any(), "team-1").Return(collection, nil).Times(1)

				tr.EXPECT().Run(gomock.Any()).Times(0)
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			apiMock := testutil.NewMockChannelAPI(ctrl)
			cacherMock := testutil.NewMockCacher(ctrl)
			taskRunnerMock := testutil.NewMockTaskRunner(ctrl)

			tc.setupMocks(apiMock, cacherMock, taskRunnerMock)

			var cacherArg *cacher.CacheHandler
			if tc.cacheEnabled {
				cacherArg = &cacher.CacheHandler{Cacher: cacherMock, Runner: taskRunnerMock}
			}

			resolver := NewChannelResolverCacheable(apiMock, cacherArg)

			id, err := resolver.ResolveChannelRefToID(context.Background(), tc.teamID, tc.channelRef)

			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.expectedID, id)
		})
	}
}

func TestChannelResolverCacheable_ResolveChannelMemberRefToID(t *testing.T) {
	type testCase struct {
		name         string
		teamID       string
		channelID    string
		userRef      string
		cacheEnabled bool
		setupMocks   func(api *testutil.MockChannelAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner)
		expectedID   string
		expectError  bool
	}

	testCases := []testCase{
		{
			name:         "Empty user reference",
			teamID:       "team-1",
			channelID:    "chan-1",
			userRef:      " ",
			cacheEnabled: true,
			setupMocks: func(api *testutil.MockChannelAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner) {
				api.EXPECT().ListMembers(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
				c.EXPECT().Get(gomock.Any()).Times(0)
				tr.EXPECT().Run(gomock.Any()).Times(0)
			},
			expectError: true,
		},
		{
			name:         "Direct member ID (GUID) short circuit",
			teamID:       "team-1",
			channelID:    "chan-1",
			userRef:      "d94f3f01-0c1f-4aac-9c8a-1fb3f3f1e3de",
			cacheEnabled: true,
			setupMocks: func(api *testutil.MockChannelAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner) {
				api.EXPECT().ListMembers(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
				c.EXPECT().Get(gomock.Any()).Times(0)
				tr.EXPECT().Run(gomock.Any()).Times(0)
			},
			expectedID: "d94f3f01-0c1f-4aac-9c8a-1fb3f3f1e3de",
		},
		{
			name:         "Cache hit single ID",
			teamID:       "team-1",
			channelID:    "chan-1",
			userRef:      "user-ref",
			cacheEnabled: true,
			setupMocks: func(api *testutil.MockChannelAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner) {
				api.EXPECT().ListMembers(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
				c.EXPECT().
					Get(cacher.NewChannelMemberKey("team-1", "chan-1", "user-ref", nil)).
					Return([]string{"member-id-123"}, true, nil).
					Times(1)
				tr.EXPECT().Run(gomock.Any()).Times(0)
			},
			expectedID: "member-id-123",
		},
		{
			name:         "Cache miss uses API and caches",
			teamID:       "team-42",
			channelID:    "chan-7",
			cacheEnabled: true,
			userRef:      " u-1 ",
			setupMocks: func(api *testutil.MockChannelAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner) {
				member := testutil.NewGraphMember(
					&testutil.NewMemberParams{
						ID:          util.Ptr("m-1"),
						UserID:      util.Ptr("u-1"),
						DisplayName: util.Ptr("Alice"),
					},
				)
				collection := testutil.NewMemberCollection(member)

				api.EXPECT().
					ListMembers(gomock.Any(), "team-42", "chan-7").
					Return(collection, nil).
					Times(1)
				c.EXPECT().
					Get(cacher.NewChannelMemberKey("team-42", "chan-7", "u-1", nil)).
					Return(nil, false, nil).
					Times(1)
				tr.EXPECT().Run(gomock.Any()).Times(1)
			},
			expectedID: "m-1",
		},
		{
			name:         "Cache disabled skips cache",
			teamID:       "team-1",
			channelID:    "chan-1",
			userRef:      "u-1",
			cacheEnabled: false,
			setupMocks: func(api *testutil.MockChannelAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner) {
				member := testutil.NewGraphMember(
					&testutil.NewMemberParams{
						ID:          util.Ptr("m-api"),
						UserID:      util.Ptr("u-1"),
						DisplayName: util.Ptr("Alice"),
					},
				)
				collection := testutil.NewMemberCollection(member)

				api.EXPECT().
					ListMembers(gomock.Any(), "team-1", "chan-1").
					Return(collection, nil).
					Times(1)
				c.EXPECT().Get(gomock.Any()).Times(0)
				tr.EXPECT().Run(gomock.Any()).Times(0)
			},
			expectedID: "m-api",
		},
		{
			name:         "Resolver error propagated",
			teamID:       "team-1",
			channelID:    "chan-1",
			userRef:      "user-ref",
			cacheEnabled: true,
			setupMocks: func(api *testutil.MockChannelAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner) {
				apiErr := &sender.RequestError{Code: 500, Message: "api error"}

				api.EXPECT().
					ListMembers(gomock.Any(), "team-1", "chan-1").
					Return(nil, apiErr).
					Times(1)
				c.EXPECT().
					Get(cacher.NewChannelMemberKey("team-1", "chan-1", "user-ref", nil)).
					Return(nil, false, nil).
					Times(1)
				tr.EXPECT().Run(gomock.Any()).Times(0)
			},
			expectError: true,
		},

		{
			name:         "Cache hit multiple IDs triggers invalidation and falls back to API",
			teamID:       "team-42",
			channelID:    "chan-7",
			userRef:      "u-1",
			cacheEnabled: true,
			setupMocks: func(api *testutil.MockChannelAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner) {
				member := testutil.NewGraphMember(
					&testutil.NewMemberParams{
						ID:          util.Ptr("m-1"),
						UserID:      util.Ptr("u-1"),
						DisplayName: util.Ptr("Alice"),
					},
				)
				collection := testutil.NewMemberCollection(member)

				c.EXPECT().
					Get(cacher.NewChannelMemberKey("team-42", "chan-7", "u-1", nil)).
					Return([]string{"m-x", "m-y"}, true, nil).
					Times(1)

				api.EXPECT().
					ListMembers(gomock.Any(), "team-42", "chan-7").
					Return(collection, nil).
					Times(1)

				tr.EXPECT().Run(gomock.Any()).Times(2)
			},
			expectedID: "m-1",
		},
		{
			name:         "Cache Get error is ignored and resolver falls back to API",
			teamID:       "team-42",
			channelID:    "chan-7",
			userRef:      "u-1",
			cacheEnabled: true,
			setupMocks: func(api *testutil.MockChannelAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner) {
				member := testutil.NewGraphMember(
					&testutil.NewMemberParams{
						ID:          util.Ptr("m-1"),
						UserID:      util.Ptr("u-1"),
						DisplayName: util.Ptr("Alice"),
					},
				)
				collection := testutil.NewMemberCollection(member)

				c.EXPECT().
					Get(cacher.NewChannelMemberKey("team-42", "chan-7", "u-1", nil)).
					Return(nil, false, errors.New("cache down")).
					Times(1)

				api.EXPECT().
					ListMembers(gomock.Any(), "team-42", "chan-7").
					Return(collection, nil).
					Times(1)

				tr.EXPECT().Run(gomock.Any()).Times(1)
			},
			expectedID: "m-1",
		},
		{
			name:         "Cache hit with wrong type is ignored and resolver falls back to API",
			teamID:       "team-42",
			channelID:    "chan-7",
			userRef:      "u-1",
			cacheEnabled: true,
			setupMocks: func(api *testutil.MockChannelAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner) {
				member := testutil.NewGraphMember(
					&testutil.NewMemberParams{
						ID:          util.Ptr("m-1"),
						UserID:      util.Ptr("u-1"),
						DisplayName: util.Ptr("Alice"),
					},
				)
				collection := testutil.NewMemberCollection(member)

				c.EXPECT().
					Get(cacher.NewChannelMemberKey("team-42", "chan-7", "u-1", nil)).
					Return("nope", true, nil).
					Times(1)

				api.EXPECT().
					ListMembers(gomock.Any(), "team-42", "chan-7").
					Return(collection, nil).
					Times(1)

				tr.EXPECT().Run(gomock.Any()).Times(1)
			},
			expectedID: "m-1",
		},
		{
			name:         "Extract fails (member not found / ambiguous) - error returned and nothing is cached",
			teamID:       "team-42",
			channelID:    "chan-7",
			userRef:      "u-x",
			cacheEnabled: true,
			setupMocks: func(api *testutil.MockChannelAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner) {
				member := testutil.NewGraphMember(
					&testutil.NewMemberParams{
						ID:          util.Ptr("m-1"),
						UserID:      util.Ptr("u-1"),
						DisplayName: util.Ptr("Alice"),
					},
				)
				collection := testutil.NewMemberCollection(member)

				c.EXPECT().
					Get(cacher.NewChannelMemberKey("team-42", "chan-7", "u-x", nil)).
					Return(nil, false, nil).
					Times(1)

				api.EXPECT().
					ListMembers(gomock.Any(), "team-42", "chan-7").
					Return(collection, nil).
					Times(1)

				tr.EXPECT().Run(gomock.Any()).Times(0)
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			apiMock := testutil.NewMockChannelAPI(ctrl)
			cacherMock := testutil.NewMockCacher(ctrl)
			taskRunnerMock := testutil.NewMockTaskRunner(ctrl)

			tc.setupMocks(apiMock, cacherMock, taskRunnerMock)

			var cacherArg *cacher.CacheHandler
			if tc.cacheEnabled {
				cacherArg = &cacher.CacheHandler{Cacher: cacherMock, Runner: taskRunnerMock}
			}

			resolver := NewChannelResolverCacheable(apiMock, cacherArg)

			id, err := resolver.ResolveChannelMemberRefToID(
				context.Background(),
				tc.teamID,
				tc.channelID,
				tc.userRef,
			)

			if tc.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedID, id)
		})
	}
}
