package resolver

import (
	"context"
	"testing"

	"github.com/pzsp-teams/lib/internal/cacher"
	sender "github.com/pzsp-teams/lib/internal/sender"
	testutil "github.com/pzsp-teams/lib/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestTeamResolverCacheable_ResolveTeamRefToID(t *testing.T) {
	ctx := context.Background()

	type testCase struct {
		name         string
		teamRef      string
		cacheEnabled bool
		setupMocks   func(api *testutil.MockTeamAPI, c *testutil.MockCacher)
		expectedID   string
		expectError  bool
	}

	testCases := []testCase{
		{
			name:         "Empty team reference",
			teamRef:      "   ",
			cacheEnabled: true,
			setupMocks: func(api *testutil.MockTeamAPI, c *testutil.MockCacher) {
				api.EXPECT().ListMyJoined(gomock.Any()).Times(0)
				c.EXPECT().Get(gomock.Any()).Times(0)
			},
			expectError: true,
		},
		{
			name:         "GUID short circuit",
			teamRef:      "123e4567-e89b-12d3-a456-426614174000",
			cacheEnabled: true,
			setupMocks: func(api *testutil.MockTeamAPI, c *testutil.MockCacher) {
				api.EXPECT().ListMyJoined(gomock.Any()).Times(0)
				c.EXPECT().Get(gomock.Any()).Times(0)
			},
			expectedID: "123e4567-e89b-12d3-a456-426614174000",
		},
		{
			name:         "Cache single hit",
			teamRef:      "My Team",
			cacheEnabled: true,
			setupMocks: func(api *testutil.MockTeamAPI, c *testutil.MockCacher) {
				api.EXPECT().ListMyJoined(gomock.Any()).Times(0)
				c.EXPECT().Get(cacher.NewTeamKey("My Team")).Return([]string{"team-id-123"}, true, nil).Times(1)
				c.EXPECT().Set(gomock.Any(), gomock.Any()).Times(0)
			},
			expectedID: "team-id-123",
		},
		{
			name:         "Cache miss uses API and caches result",
			teamRef:      "My Team",
			cacheEnabled: true,
			setupMocks: func(api *testutil.MockTeamAPI, c *testutil.MockCacher) {
				team := newGraphTeam("team-id-123", "My Team")
				collection := newTeamCollection(team)
				api.EXPECT().ListMyJoined(gomock.Any()).Return(collection, nil).Times(1)
				c.EXPECT().Get(cacher.NewTeamKey("My Team")).Return(nil, false, nil).Times(1)
				c.EXPECT().Set(cacher.NewTeamKey("My Team"), "team-id-123").Return(nil).Times(1)
			},
			expectedID: "team-id-123",
		},
		{
			name:         "Cache disabled",
			teamRef:      "My Team",
			cacheEnabled: false,
			setupMocks: func(api *testutil.MockTeamAPI, c *testutil.MockCacher) {
				team := newGraphTeam("team-id-123", "My Team")
				collection := newTeamCollection(team)
				api.EXPECT().ListMyJoined(gomock.Any()).Return(collection, nil).Times(1)
				c.EXPECT().Get(gomock.Any()).Times(0)
				c.EXPECT().Set(gomock.Any(), gomock.Any()).Times(0)
			},
			expectedID: "team-id-123",
		},
		{
			name:         "Fetch from API fails",
			teamRef:      "My Team",
			cacheEnabled: true,
			setupMocks: func(api *testutil.MockTeamAPI, c *testutil.MockCacher) {
				apiErr := &sender.RequestError{
					Code:    500,
					Message: "Internal Server Error",
				}
				api.EXPECT().ListMyJoined(gomock.Any()).Return(nil, apiErr).Times(1)
				c.EXPECT().Get(cacher.NewTeamKey("My Team")).Return(nil, false, nil).Times(1)
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
			if tc.setupMocks != nil {
				tc.setupMocks(mockAPI, mockCacher)
			}

			res := NewTeamResolverCacheable(mockAPI, mockCacher, tc.cacheEnabled)

			id, err := res.ResolveTeamRefToID(ctx, tc.teamRef)

			if tc.expectError {
				require.Error(t, err, "expected error but got none")
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expectedID, id, "resolved team ID does not match expected")
		})
	}
}
