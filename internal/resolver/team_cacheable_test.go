package resolver

import (
	"context"
	"net/http"
	"testing"

	"github.com/pzsp-teams/lib/internal/cacher"
	sender "github.com/pzsp-teams/lib/internal/sender"
	testutil "github.com/pzsp-teams/lib/internal/testutil"
	"github.com/pzsp-teams/lib/internal/util"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestTeamResolverCacheable_ResolveTeamRefToID(t *testing.T) {
	type testCase struct {
		name         string
		teamRef      string
		cacheEnabled bool
		setupMocks   func(api *testutil.MockTeamAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner)
		expectedID   string
		expectError  bool
	}

	testCases := []testCase{
		{
			name:         "Empty team reference",
			teamRef:      "   ",
			cacheEnabled: true,
			setupMocks: func(api *testutil.MockTeamAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner) {
				api.EXPECT().ListMyJoined(gomock.Any()).Times(0)
				c.EXPECT().Get(gomock.Any()).Times(0)
				tr.EXPECT().Run(gomock.Any()).Times(0)
			},
			expectError: true,
		},
		{
			name:         "GUID short circuit",
			teamRef:      "123e4567-e89b-12d3-a456-426614174000",
			cacheEnabled: true,
			setupMocks: func(api *testutil.MockTeamAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner) {
				api.EXPECT().ListMyJoined(gomock.Any()).Times(0)
				c.EXPECT().Get(gomock.Any()).Times(0)
				tr.EXPECT().Run(gomock.Any()).Times(0)
			},
			expectedID: "123e4567-e89b-12d3-a456-426614174000",
		},
		{
			name:         "Trims whitespace before cache lookup",
			teamRef:      "  My Team   ",
			cacheEnabled: true,
			setupMocks: func(api *testutil.MockTeamAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner) {
				api.EXPECT().ListMyJoined(gomock.Any()).Times(0)
				c.EXPECT().
					Get(cacher.NewTeamKey("My Team")).
					Return([]string{"team-id-123"}, true, nil).
					Times(1)
				tr.EXPECT().Run(gomock.Any()).Times(0)
			},
			expectedID: "team-id-123",
		},
		{
			name:         "Cache single hit",
			teamRef:      "My Team",
			cacheEnabled: true,
			setupMocks: func(api *testutil.MockTeamAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner) {
				api.EXPECT().ListMyJoined(gomock.Any()).Times(0)
				c.EXPECT().
					Get(cacher.NewTeamKey("My Team")).
					Return([]string{"team-id-123"}, true, nil).
					Times(1)
				tr.EXPECT().Run(gomock.Any()).Times(0)
			},
			expectedID: "team-id-123",
		},
		{
			name:         "Cache miss uses API and caches result",
			teamRef:      "My Team",
			cacheEnabled: true,
			setupMocks: func(api *testutil.MockTeamAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner) {
				collection := testutil.NewTeamCollection(testutil.NewGraphTeam(
					&testutil.NewTeamParams{
						ID:          util.Ptr("team-id-123"),
						DisplayName: util.Ptr("My Team"),
					},
				))

				api.EXPECT().ListMyJoined(gomock.Any()).Return(collection, nil).Times(1)
				c.EXPECT().Get(cacher.NewTeamKey("My Team")).Return(nil, false, nil).Times(1)
				tr.EXPECT().Run(gomock.Any()).Times(1)
			},
			expectedID: "team-id-123",
		},
		{
			name:         "Cache disabled",
			teamRef:      "My Team",
			cacheEnabled: false,
			setupMocks: func(api *testutil.MockTeamAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner) {
				collection := testutil.NewTeamCollection(testutil.NewGraphTeam(
					&testutil.NewTeamParams{
						ID:          util.Ptr("team-id-123"),
						DisplayName: util.Ptr("My Team"),
					},
				))

				api.EXPECT().ListMyJoined(gomock.Any()).Return(collection, nil).Times(1)
				c.EXPECT().Get(gomock.Any()).Times(0)
				tr.EXPECT().Run(gomock.Any()).Times(0)
			},
			expectedID: "team-id-123",
		},
		{
			name:         "Fetch from API fails",
			teamRef:      "My Team",
			cacheEnabled: true,
			setupMocks: func(api *testutil.MockTeamAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner) {
				apiErr := &sender.RequestError{
					Code:    http.StatusInternalServerError,
					Message: "Internal Server Error",
				}
				api.EXPECT().ListMyJoined(gomock.Any()).Return(nil, apiErr).Times(1)
				c.EXPECT().Get(cacher.NewTeamKey("My Team")).Return(nil, false, nil).Times(1)
				tr.EXPECT().Run(gomock.Any()).Times(0)
			},
			expectError: true,
		},
		{
			name:         "Team not found in API response",
			teamRef:      "My Team",
			cacheEnabled: true,
			setupMocks: func(api *testutil.MockTeamAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner) {
				collection := testutil.NewTeamCollection(testutil.NewGraphTeam(
					&testutil.NewTeamParams{
						ID:          util.Ptr("team-id-999"),
						DisplayName: util.Ptr("Other Team"),
					},
				))

				api.EXPECT().ListMyJoined(gomock.Any()).Return(collection, nil).Times(1)
				c.EXPECT().Get(cacher.NewTeamKey("My Team")).Return(nil, false, nil).Times(1)
				tr.EXPECT().Run(gomock.Any()).Times(0)
			},
			expectError: true,
		},
		{
			name:         "Team name ambiguous",
			teamRef:      "My Team",
			cacheEnabled: true,
			setupMocks: func(api *testutil.MockTeamAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner) {
				collection := testutil.NewTeamCollection(
					testutil.NewGraphTeam(
						&testutil.NewTeamParams{
							ID:          util.Ptr("team-id-1"),
							DisplayName: util.Ptr("My Team"),
						},
					),
					testutil.NewGraphTeam(
						&testutil.NewTeamParams{
							ID:          util.Ptr("team-id-2"),
							DisplayName: util.Ptr("My Team"),
						},
					),
				)

				api.EXPECT().ListMyJoined(gomock.Any()).Return(collection, nil).Times(1)
				c.EXPECT().Get(cacher.NewTeamKey("My Team")).Return(nil, false, nil).Times(1)
				tr.EXPECT().Run(gomock.Any()).Times(0)
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAPI := testutil.NewMockTeamAPI(ctrl)
			mockCacher := testutil.NewMockCacher(ctrl)
			mockTaskRunner := testutil.NewMockTaskRunner(ctrl)

			tc.setupMocks(mockAPI, mockCacher, mockTaskRunner)

			var cacherArg *cacher.CacheHandler
			if tc.cacheEnabled {
				cacherArg = &cacher.CacheHandler{Cacher: mockCacher, Runner: mockTaskRunner}
			}

			res := NewTeamResolverCacheable(mockAPI, cacherArg)

			id, err := res.ResolveTeamRefToID(context.Background(), tc.teamRef)

			if tc.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedID, id)
		})
	}
}

func TestTeamResolverCacheable_ResolveTeamMemberRefToID(t *testing.T) {
	type testCase struct {
		name         string
		teamID       string
		userRef      string
		cacheEnabled bool
		setupMocks   func(api *testutil.MockTeamAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner)
		expectedID   string
		expectError  bool
	}

	const teamID = "team-id-123"

	testCases := []testCase{
		{
			name:         "Empty user reference",
			teamID:       teamID,
			userRef:      "   ",
			cacheEnabled: true,
			setupMocks: func(api *testutil.MockTeamAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner) {
				api.EXPECT().ListMembers(gomock.Any(), gomock.Any()).Times(0)
				c.EXPECT().Get(gomock.Any()).Times(0)
				tr.EXPECT().Run(gomock.Any()).Times(0)
			},
			expectError: true,
		},
		{
			name:         "GUID short circuit (also trims)",
			teamID:       teamID,
			userRef:      "  123e4567-e89b-12d3-a456-426614174000   ",
			cacheEnabled: true,
			setupMocks: func(api *testutil.MockTeamAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner) {
				api.EXPECT().ListMembers(gomock.Any(), gomock.Any()).Times(0)
				c.EXPECT().Get(gomock.Any()).Times(0)
				tr.EXPECT().Run(gomock.Any()).Times(0)
			},
			expectedID: "123e4567-e89b-12d3-a456-426614174000",
		},
		{
			name:         "Cache hit (email, trims whitespace before key build)",
			teamID:       teamID,
			userRef:      "  user@example.com  ",
			cacheEnabled: true,
			setupMocks: func(api *testutil.MockTeamAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner) {
				api.EXPECT().ListMembers(gomock.Any(), gomock.Any()).Times(0)
				c.EXPECT().
					Get(cacher.NewTeamMemberKey(teamID, "user@example.com", nil)).
					Return([]string{"member-id-123"}, true, nil).
					Times(1)
				tr.EXPECT().Run(gomock.Any()).Times(0)
			},
			expectedID: "member-id-123",
		},
		{
			name:         "Cache miss uses API and caches result (email -> member ID)",
			teamID:       teamID,
			userRef:      "user@example.com",
			cacheEnabled: true,
			setupMocks: func(api *testutil.MockTeamAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner) {
				collection := testutil.NewMemberCollection(
					testutil.NewGraphMember(
						&testutil.NewMemberParams{
							ID:    util.Ptr("member-id-123"),
							Email: util.Ptr("user@example.com"),
						},
					),
				)

				api.EXPECT().ListMembers(gomock.Any(), teamID).Return(collection, nil).Times(1)
				c.EXPECT().Get(cacher.NewTeamMemberKey(teamID, "user@example.com", nil)).Return(nil, false, nil).Times(1)
				tr.EXPECT().Run(gomock.Any()).Times(1)
			},
			expectedID: "member-id-123",
		},
		{
			name:         "Cache disabled",
			teamID:       teamID,
			userRef:      "user@example.com",
			cacheEnabled: false,
			setupMocks: func(api *testutil.MockTeamAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner) {
				collection := testutil.NewMemberCollection(
					testutil.NewGraphMember(
						&testutil.NewMemberParams{
							ID:    util.Ptr("member-id-123"),
							Email: util.Ptr("user@example.com"),
						},
					),
				)

				api.EXPECT().ListMembers(gomock.Any(), teamID).Return(collection, nil).Times(1)
				c.EXPECT().Get(gomock.Any()).Times(0)
				tr.EXPECT().Run(gomock.Any()).Times(0)
			},
			expectedID: "member-id-123",
		},
		{
			name:         "Fetch from API fails",
			teamID:       teamID,
			userRef:      "user@example.com",
			cacheEnabled: true,
			setupMocks: func(api *testutil.MockTeamAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner) {
				apiErr := &sender.RequestError{
					Code:    http.StatusInternalServerError,
					Message: "Internal Server Error",
				}

				api.EXPECT().ListMembers(gomock.Any(), teamID).Return(nil, apiErr).Times(1)
				c.EXPECT().Get(cacher.NewTeamMemberKey(teamID, "user@example.com", nil)).Return(nil, false, nil).Times(1)
				tr.EXPECT().Run(gomock.Any()).Times(0)
			},
			expectError: true,
		},
		{
			name:         "Member not found in API response",
			teamID:       teamID,
			userRef:      "user@example.com",
			cacheEnabled: true,
			setupMocks: func(api *testutil.MockTeamAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner) {
				collection := testutil.NewMemberCollection(
					testutil.NewGraphMember(
						&testutil.NewMemberParams{
							ID:    util.Ptr("member-id-999"),
							Email: util.Ptr("other@example.com"),
						},
					),
				)

				api.EXPECT().ListMembers(gomock.Any(), teamID).Return(collection, nil).Times(1)
				c.EXPECT().Get(cacher.NewTeamMemberKey(teamID, "user@example.com", nil)).Return(nil, false, nil).Times(1)
				tr.EXPECT().Run(gomock.Any()).Times(0)
			},
			expectError: true,
		},
		{
			name:         "Member email duplicates - returns first match and caches",
			teamID:       teamID,
			userRef:      "user@example.com",
			cacheEnabled: true,
			setupMocks: func(api *testutil.MockTeamAPI, c *testutil.MockCacher, tr *testutil.MockTaskRunner) {
				collection := testutil.NewMemberCollection(
					testutil.NewGraphMember(&testutil.NewMemberParams{
						ID:    util.Ptr("member-id-1"),
						Email: util.Ptr("user@example.com"),
					}),
					testutil.NewGraphMember(&testutil.NewMemberParams{
						ID:    util.Ptr("member-id-2"),
						Email: util.Ptr("user@example.com"),
					}),
				)

				api.EXPECT().ListMembers(gomock.Any(), teamID).Return(collection, nil).Times(1)
				c.EXPECT().Get(cacher.NewTeamMemberKey(teamID, "user@example.com", nil)).Return(nil, false, nil).Times(1)

				tr.EXPECT().Run(gomock.Any()).Times(1)
			},
			expectedID:  "member-id-1", 
			expectError: false,
		},

	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAPI := testutil.NewMockTeamAPI(ctrl)
			mockCacher := testutil.NewMockCacher(ctrl)
			mockTaskRunner := testutil.NewMockTaskRunner(ctrl)

			tc.setupMocks(mockAPI, mockCacher, mockTaskRunner)

			var cacherArg *cacher.CacheHandler
			if tc.cacheEnabled {
				cacherArg = &cacher.CacheHandler{Cacher: mockCacher, Runner: mockTaskRunner}
			}

			res := NewTeamResolverCacheable(mockAPI, cacherArg)

			id, err := res.ResolveTeamMemberRefToID(context.Background(), tc.teamID, tc.userRef)

			if tc.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedID, id)
		})
	}
}
