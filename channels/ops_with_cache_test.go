package channels

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pzsp-teams/lib/internal/cacher"
	snd "github.com/pzsp-teams/lib/internal/sender"
	"github.com/pzsp-teams/lib/internal/testutil"
	"github.com/pzsp-teams/lib/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type opsWithCacheSUTDeps struct {
	chanOps *testutil.MockchannelOps
	cacher  *testutil.MockCacher
	runner  *testutil.MockTaskRunner
}

func newOpsWithCacheSUT(
	t *testing.T,
	setup func(ctx context.Context, d opsWithCacheSUTDeps),
) (channelOps, context.Context) {
	t.Helper()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	ctx := context.Background()

	d := opsWithCacheSUTDeps{
		chanOps: testutil.NewMockchannelOps(ctrl),
		cacher:  testutil.NewMockCacher(ctrl),
		runner:  testutil.NewMockTaskRunner(ctrl),
	}

	if setup != nil {
		setup(ctx, d)
	}

	sut := NewOpsWithCache(d.chanOps, &cacher.CacheHandler{Cacher: d.cacher, Runner: d.runner})
	return sut, ctx
}

func expectRunNow(r *testutil.MockTaskRunner) {
	r.EXPECT().Run(gomock.Any()).Do(func(fn func()) { fn() }).Times(1)
}

func reqErr(code int) *snd.RequestError {
	return &snd.RequestError{Code: code, Message: "boom"}
}

func TestNewOpsWithCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockOps := testutil.NewMockchannelOps(ctrl)

	t.Run("returns underlying ops when cache is nil", func(t *testing.T) {
		sut := NewOpsWithCache(mockOps, nil)
		require.Same(t, mockOps, sut.(*testutil.MockchannelOps))
	})

	t.Run("wraps ops when cache is provided", func(t *testing.T) {
		sut := NewOpsWithCache(mockOps, &cacher.CacheHandler{Cacher: testutil.NewMockCacher(ctrl), Runner: testutil.NewMockTaskRunner(ctrl)})
		_, ok := sut.(*opsWithCache)
		require.True(t, ok)
	})
}

func TestOpsWithCache_Wait(t *testing.T) {
	sut, _ := newOpsWithCacheSUT(t, func(_ context.Context, d opsWithCacheSUTDeps) {
		d.runner.EXPECT().Wait().Times(1)
	})

	sut.Wait()
}

func TestOpsWithCache_ListChannelsByTeamID(t *testing.T) {
	teamID := "team-1"

	t.Run("success caches non-nil, non-blank channels", func(t *testing.T) {
		out := []*models.Channel{
			{ID: "c1", Name: "General"},
			nil,
			{ID: "c2", Name: "   "},
			{ID: "c3", Name: "Dev"},
		}

		sut, ctx := newOpsWithCacheSUT(t, func(_ context.Context, d opsWithCacheSUTDeps) {
			d.chanOps.EXPECT().ListChannelsByTeamID(gomock.Any(), teamID).Return(out, nil)

			expectRunNow(d.runner)
			d.cacher.EXPECT().Set(cacher.NewChannelKey(teamID, "General"), "c1").Return(nil)
			d.cacher.EXPECT().Set(cacher.NewChannelKey(teamID, "Dev"), "c3").Return(nil)
		})

		got, err := sut.ListChannelsByTeamID(ctx, teamID)
		require.Nil(t, err)
		assert.Equal(t, out, got)
	})

	t.Run("error triggers cache handler", func(t *testing.T) {
		err400 := reqErr(400)

		sut, ctx := newOpsWithCacheSUT(t, func(_ context.Context, d opsWithCacheSUTDeps) {
			d.chanOps.EXPECT().ListChannelsByTeamID(gomock.Any(), teamID).Return(nil, err400)

			expectRunNow(d.runner)
			d.cacher.EXPECT().Clear().Return(nil)
		})

		got, err := sut.ListChannelsByTeamID(ctx, teamID)
		require.Nil(t, got)
		require.Same(t, err400, err)
	})
}

func TestOpsWithCache_GetChannelByID(t *testing.T) {
	teamID := "team-1"
	channelID := "channel-1"

	t.Run("nil channel does not schedule cache update", func(t *testing.T) {
		sut, ctx := newOpsWithCacheSUT(t, func(_ context.Context, d opsWithCacheSUTDeps) {
			d.chanOps.EXPECT().GetChannelByID(gomock.Any(), teamID, channelID).Return(nil, nil)
			// no runner.Run expected
		})

		ch, err := sut.GetChannelByID(ctx, teamID, channelID)
		require.Nil(t, err)
		require.Nil(t, ch)
	})

	t.Run("success caches channel", func(t *testing.T) {
		out := &models.Channel{ID: "c1", Name: "General"}

		sut, ctx := newOpsWithCacheSUT(t, func(_ context.Context, d opsWithCacheSUTDeps) {
			d.chanOps.EXPECT().GetChannelByID(gomock.Any(), teamID, channelID).Return(out, nil)

			expectRunNow(d.runner)
			d.cacher.EXPECT().Set(cacher.NewChannelKey(teamID, "General"), "c1").Return(nil)
		})

		ch, err := sut.GetChannelByID(ctx, teamID, channelID)
		require.Nil(t, err)
		require.Equal(t, out, ch)
	})
}

func TestOpsWithCache_DeleteChannel(t *testing.T) {
	teamID := "team-1"
	channelID := "channel-1"
	channelRef := "General"

	sut, ctx := newOpsWithCacheSUT(t, func(_ context.Context, d opsWithCacheSUTDeps) {
		d.chanOps.EXPECT().DeleteChannel(gomock.Any(), teamID, channelID, channelRef).Return(nil)

		expectRunNow(d.runner)
		d.cacher.EXPECT().Invalidate(cacher.NewChannelKey(teamID, channelRef)).Return(nil)
	})

	err := sut.DeleteChannel(ctx, teamID, channelID, channelRef)
	require.Nil(t, err)
}

func TestOpsWithCache_ListMembers(t *testing.T) {
	teamID := "team-1"
	channelID := "channel-1"

	out := []*models.Member{
		{ID: "m1", Email: "a@b.com"},
		nil,
		{ID: "m2", Email: "   "},
		{ID: "m3", Email: "c@d.com"},
	}

	sut, ctx := newOpsWithCacheSUT(t, func(_ context.Context, d opsWithCacheSUTDeps) {
		d.chanOps.EXPECT().ListMembers(gomock.Any(), teamID, channelID).Return(out, nil)

		expectRunNow(d.runner)
		d.cacher.EXPECT().Set(cacher.NewChannelMemberKey(teamID, channelID, "a@b.com", nil), "m1").Return(nil)
		d.cacher.EXPECT().Set(cacher.NewChannelMemberKey(teamID, channelID, "c@d.com", nil), "m3").Return(nil)
	})

	got, err := sut.ListMembers(ctx, teamID, channelID)
	require.Nil(t, err)
	assert.Equal(t, out, got)
}

func TestOpsWithCache_RemoveMember(t *testing.T) {
	teamID := "team-1"
	channelID := "channel-1"
	memberID := "member-1"
	userRef := "user@example.com"

	sut, ctx := newOpsWithCacheSUT(t, func(_ context.Context, d opsWithCacheSUTDeps) {
		d.chanOps.EXPECT().RemoveMember(gomock.Any(), teamID, channelID, memberID, userRef).Return(nil)

		expectRunNow(d.runner)
		d.cacher.EXPECT().Invalidate(cacher.NewChannelMemberKey(teamID, channelID, userRef, nil)).Return(nil)
	})

	err := sut.RemoveMember(ctx, teamID, channelID, memberID, userRef)
	require.Nil(t, err)
}

func TestOpsWithCache_WithErrorClearMethods_ClearCacheOnError(t *testing.T) {
	err400 := reqErr(400)

	type testCase struct {
		name   string
		expect func(d opsWithCacheSUTDeps)
		call   func(sut channelOps, ctx context.Context) (isNil bool, err *snd.RequestError)
	}

	teamID := "team-1"
	channelID := "channel-1"
	messageID := "msg-1"
	replyID := "reply-1"
	memberID := "member-1"

	cases := []testCase{
		{
			name: "SendMessage",
			expect: func(d opsWithCacheSUTDeps) {
				d.chanOps.EXPECT().SendMessage(gomock.Any(), teamID, channelID, gomock.Any()).Return(nil, err400)
			},
			call: func(sut channelOps, ctx context.Context) (bool, *snd.RequestError) {
				msg, err := sut.SendMessage(ctx, teamID, channelID, models.MessageBody{})
				return msg == nil, err
			},
		},
		{
			name: "SendReply",
			expect: func(d opsWithCacheSUTDeps) {
				d.chanOps.EXPECT().SendReply(gomock.Any(), teamID, channelID, messageID, gomock.Any()).Return(nil, err400)
			},
			call: func(sut channelOps, ctx context.Context) (bool, *snd.RequestError) {
				msg, err := sut.SendReply(ctx, teamID, channelID, messageID, models.MessageBody{})
				return msg == nil, err
			},
		},
		{
			name: "ListMessages",
			expect: func(d opsWithCacheSUTDeps) {
				d.chanOps.EXPECT().ListMessages(gomock.Any(), teamID, channelID, gomock.Any()).Return(nil, err400)
			},
			call: func(sut channelOps, ctx context.Context) (bool, *snd.RequestError) {
				msgs, err := sut.ListMessages(ctx, teamID, channelID, nil)
				return msgs == nil, err
			},
		},
		{
			name: "ListReplies",
			expect: func(d opsWithCacheSUTDeps) {
				d.chanOps.EXPECT().ListReplies(gomock.Any(), teamID, channelID, messageID, gomock.Any()).Return(nil, err400)
			},
			call: func(sut channelOps, ctx context.Context) (bool, *snd.RequestError) {
				msgs, err := sut.ListReplies(ctx, teamID, channelID, messageID, nil)
				return msgs == nil, err
			},
		},
		{
			name: "GetMessage",
			expect: func(d opsWithCacheSUTDeps) {
				d.chanOps.EXPECT().GetMessage(gomock.Any(), teamID, channelID, messageID).Return(nil, err400)
			},
			call: func(sut channelOps, ctx context.Context) (bool, *snd.RequestError) {
				msg, err := sut.GetMessage(ctx, teamID, channelID, messageID)
				return msg == nil, err
			},
		},
		{
			name: "GetReply",
			expect: func(d opsWithCacheSUTDeps) {
				d.chanOps.EXPECT().GetReply(gomock.Any(), teamID, channelID, messageID, replyID).Return(nil, err400)
			},
			call: func(sut channelOps, ctx context.Context) (bool, *snd.RequestError) {
				msg, err := sut.GetReply(ctx, teamID, channelID, messageID, replyID)
				return msg == nil, err
			},
		},
		{
			name: "UpdateMemberRoles",
			expect: func(d opsWithCacheSUTDeps) {
				d.chanOps.EXPECT().UpdateMemberRoles(gomock.Any(), teamID, channelID, memberID, true).Return(nil, err400)
			},
			call: func(sut channelOps, ctx context.Context) (bool, *snd.RequestError) {
				m, err := sut.UpdateMemberRoles(ctx, teamID, channelID, memberID, true)
				return m == nil, err
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sut, ctx := newOpsWithCacheSUT(t, func(_ context.Context, d opsWithCacheSUTDeps) {
				tc.expect(d)
				expectRunNow(d.runner)
				d.cacher.EXPECT().Clear().Return(nil)
			})

			isNil, err := tc.call(sut, ctx)
			require.True(t, isNil)
			require.Same(t, err400, err)
		})
	}
}
