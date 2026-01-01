package resolver

import (
	"context"
	"net/http"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/pzsp-teams/lib/internal/cacher"
	sender "github.com/pzsp-teams/lib/internal/sender"
	testutil "github.com/pzsp-teams/lib/internal/testutil"
	"github.com/pzsp-teams/lib/internal/util"
	"github.com/stretchr/testify/assert"
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
				team := testutil.NewGraphTeam(
					&testutil.NewTeamParams{
						ID:          util.Ptr("team-id-123"),
						DisplayName: util.Ptr("My Team"),
					},
				)
				collection := testutil.NewTeamCollection(team)
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
				team := testutil.NewGraphTeam(
					&testutil.NewTeamParams{
						ID:          util.Ptr("team-id-123"),
						DisplayName: util.Ptr("My Team"),
					},
				)
				collection := testutil.NewTeamCollection(team)
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
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAPI := testutil.NewMockTeamAPI(ctrl)
			mockCacher := testutil.NewMockCacher(ctrl)
			mockTaskRunner := testutil.NewMockTaskRunner(ctrl)

			tc.setupMocks(mockAPI, mockCacher, mockTaskRunner)

			var cacherArg *cacher.CacheHandler = nil
			if tc.cacheEnabled {
				cacherArg = &cacher.CacheHandler{Cacher: mockCacher, Runner: mockTaskRunner}
			}

			res := NewTeamResolverCacheable(mockAPI, cacherArg)

			ctx := context.Background()
			id, err := res.ResolveTeamRefToID(ctx, tc.teamRef)

			if tc.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedID, id)
		})
	}
}
