package channels

import (
	"context"
	"net/http"
	"testing"

	"github.com/pzsp-teams/lib/internal/cacher"
	"github.com/pzsp-teams/lib/internal/testutil"
	"github.com/pzsp-teams/lib/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
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

func expectClearNow(d opsWithCacheSUTDeps) {
	testutil.ExpectRunNow(d.runner)
	d.cacher.EXPECT().Clear().Return(nil).Times(1)
}

func TestNewOpsWithCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockOps := testutil.NewMockchannelOps(ctrl)

	t.Run("returns underlying ops when cache is nil", func(t *testing.T) {
		sut := NewOpsWithCache(mockOps, nil)

		got := sut.(*testutil.MockchannelOps)
		require.Same(t, mockOps, got)
	})

	t.Run("wraps ops when cache is provided", func(t *testing.T) {
		sut := NewOpsWithCache(
			mockOps,
			&cacher.CacheHandler{
				Cacher: testutil.NewMockCacher(ctrl),
				Runner: testutil.NewMockTaskRunner(ctrl),
			},
		)
		_, ok := sut.(*opsWithCache)
		require.True(t, ok)
	})
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
			d.chanOps.EXPECT().ListChannelsByTeamID(gomock.Any(), teamID).Return(out, nil).Times(1)

			testutil.ExpectRunNow(d.runner)
			d.cacher.EXPECT().Set(cacher.NewChannelKey(teamID, "General"), "c1").Return(nil).Times(1)
			d.cacher.EXPECT().Set(cacher.NewChannelKey(teamID, "Dev"), "c3").Return(nil).Times(1)
		})

		got, err := sut.ListChannelsByTeamID(ctx, teamID)
		require.NoError(t, err)
		assert.Equal(t, out, got)
	})

	t.Run("error triggers cache clear for 400", func(t *testing.T) {
		err400 := testutil.ReqErr(http.StatusBadRequest)

		sut, ctx := newOpsWithCacheSUT(t, func(_ context.Context, d opsWithCacheSUTDeps) {
			d.chanOps.EXPECT().ListChannelsByTeamID(gomock.Any(), teamID).Return(nil, err400).Times(1)
			expectClearNow(d)
		})

		got, err := sut.ListChannelsByTeamID(ctx, teamID)
		require.Nil(t, got)
		require.Error(t, err)
		require.True(t, err == err400)
	})
}

func TestOpsWithCache_GetChannelByID(t *testing.T) {
	teamID := "team-1"
	channelID := "channel-1"

	t.Run("nil channel does not schedule cache update", func(t *testing.T) {
		sut, ctx := newOpsWithCacheSUT(t, func(_ context.Context, d opsWithCacheSUTDeps) {
			d.chanOps.EXPECT().GetChannelByID(gomock.Any(), teamID, channelID).Return(nil, nil).Times(1)
		})

		ch, err := sut.GetChannelByID(ctx, teamID, channelID)
		require.NoError(t, err)
		require.Nil(t, ch)
	})

	t.Run("success caches channel", func(t *testing.T) {
		out := &models.Channel{ID: "c1", Name: "General"}

		sut, ctx := newOpsWithCacheSUT(t, func(_ context.Context, d opsWithCacheSUTDeps) {
			d.chanOps.EXPECT().GetChannelByID(gomock.Any(), teamID, channelID).Return(out, nil).Times(1)

			testutil.ExpectRunNow(d.runner)
			d.cacher.EXPECT().Set(cacher.NewChannelKey(teamID, "General"), "c1").Return(nil).Times(1)
		})

		ch, err := sut.GetChannelByID(ctx, teamID, channelID)
		require.NoError(t, err)
		require.Equal(t, out, ch)
	})

	t.Run("error triggers cache clear", func(t *testing.T) {
		err404 := testutil.ReqErr(http.StatusNotFound)

		sut, ctx := newOpsWithCacheSUT(t, func(_ context.Context, d opsWithCacheSUTDeps) {
			d.chanOps.EXPECT().GetChannelByID(gomock.Any(), teamID, channelID).Return(nil, err404).Times(1)
			expectClearNow(d)
		})

		ch, err := sut.GetChannelByID(ctx, teamID, channelID)
		require.Nil(t, ch)
		require.True(t, err == err404)
	})
}

func TestOpsWithCache_CreateStandardChannel(t *testing.T) {
	teamID := "team-1"

	t.Run("nil channel does not schedule cache update", func(t *testing.T) {
		sut, ctx := newOpsWithCacheSUT(t, func(_ context.Context, d opsWithCacheSUTDeps) {
			d.chanOps.EXPECT().CreateStandardChannel(gomock.Any(), teamID, "X").Return(nil, nil).Times(1)
		})

		ch, err := sut.CreateStandardChannel(ctx, teamID, "X")
		require.NoError(t, err)
		require.Nil(t, ch)
	})

	t.Run("success caches channel", func(t *testing.T) {
		out := &models.Channel{ID: "c1", Name: "General"}

		sut, ctx := newOpsWithCacheSUT(t, func(_ context.Context, d opsWithCacheSUTDeps) {
			d.chanOps.EXPECT().CreateStandardChannel(gomock.Any(), teamID, "General").Return(out, nil).Times(1)

			testutil.ExpectRunNow(d.runner)
			d.cacher.EXPECT().Set(cacher.NewChannelKey(teamID, "General"), "c1").Return(nil).Times(1)
		})

		ch, err := sut.CreateStandardChannel(ctx, teamID, "General")
		require.NoError(t, err)
		require.Equal(t, out, ch)
	})

	t.Run("error triggers cache clear", func(t *testing.T) {
		err404 := testutil.ReqErr(http.StatusNotFound)

		sut, ctx := newOpsWithCacheSUT(t, func(_ context.Context, d opsWithCacheSUTDeps) {
			d.chanOps.EXPECT().CreateStandardChannel(gomock.Any(), teamID, "X").Return(nil, err404).Times(1)
			expectClearNow(d)
		})

		ch, err := sut.CreateStandardChannel(ctx, teamID, "X")
		require.Nil(t, ch)
		require.True(t, err == err404)
	})
}

func TestOpsWithCache_CreatePrivateChannel(t *testing.T) {
	teamID := "team-1"
	members := []string{"u1", "u2"}
	owners := []string{"o1"}

	t.Run("nil channel does not schedule cache update", func(t *testing.T) {
		sut, ctx := newOpsWithCacheSUT(t, func(_ context.Context, d opsWithCacheSUTDeps) {
			d.chanOps.EXPECT().
				CreatePrivateChannel(gomock.Any(), teamID, "X", members, owners).
				Return(nil, nil).Times(1)
		})

		ch, err := sut.CreatePrivateChannel(ctx, teamID, "X", members, owners)
		require.NoError(t, err)
		require.Nil(t, ch)
	})

	t.Run("success caches channel", func(t *testing.T) {
		out := &models.Channel{ID: "c9", Name: "Secret"}

		sut, ctx := newOpsWithCacheSUT(t, func(_ context.Context, d opsWithCacheSUTDeps) {
			d.chanOps.EXPECT().
				CreatePrivateChannel(gomock.Any(), teamID, "Secret", members, owners).
				Return(out, nil).Times(1)

			testutil.ExpectRunNow(d.runner)
			d.cacher.EXPECT().Set(cacher.NewChannelKey(teamID, "Secret"), "c9").Return(nil).Times(1)
		})

		ch, err := sut.CreatePrivateChannel(ctx, teamID, "Secret", members, owners)
		require.NoError(t, err)
		require.Equal(t, out, ch)
	})

	t.Run("error triggers cache clear", func(t *testing.T) {
		err400 := testutil.ReqErr(http.StatusBadRequest)

		sut, ctx := newOpsWithCacheSUT(t, func(_ context.Context, d opsWithCacheSUTDeps) {
			d.chanOps.EXPECT().
				CreatePrivateChannel(gomock.Any(), teamID, "X", members, owners).
				Return(nil, err400).Times(1)
			expectClearNow(d)
		})

		ch, err := sut.CreatePrivateChannel(ctx, teamID, "X", members, owners)
		require.Nil(t, ch)
		require.True(t, err == err400)
	})
}

func TestOpsWithCache_DeleteChannel(t *testing.T) {
	teamID := "team-1"
	channelID := "channel-1"
	channelRef := "General"

	t.Run("success invalidates channel ref", func(t *testing.T) {
		sut, ctx := newOpsWithCacheSUT(t, func(_ context.Context, d opsWithCacheSUTDeps) {
			d.chanOps.EXPECT().DeleteChannel(gomock.Any(), teamID, channelID, channelRef).Return(nil).Times(1)

			testutil.ExpectRunNow(d.runner)
			d.cacher.EXPECT().Invalidate(cacher.NewChannelKey(teamID, channelRef)).Return(nil).Times(1)
		})

		err := sut.DeleteChannel(ctx, teamID, channelID, channelRef)
		require.NoError(t, err)
	})

	t.Run("error triggers cache clear and does not invalidate", func(t *testing.T) {
		err404 := testutil.ReqErr(http.StatusNotFound)

		sut, ctx := newOpsWithCacheSUT(t, func(_ context.Context, d opsWithCacheSUTDeps) {
			d.chanOps.EXPECT().DeleteChannel(gomock.Any(), teamID, channelID, channelRef).Return(err404).Times(1)
			expectClearNow(d)

			d.cacher.EXPECT().Invalidate(gomock.Any()).Times(0)
		})

		err := sut.DeleteChannel(ctx, teamID, channelID, channelRef)
		require.True(t, err == err404)
	})
}

func TestOpsWithCache_ListMembers(t *testing.T) {
	teamID := "team-1"
	channelID := "channel-1"

	t.Run("success caches non-nil, non-blank emails", func(t *testing.T) {
		out := []*models.Member{
			{ID: "m1", Email: "a@b.com"},
			nil,
			{ID: "m2", Email: "   "},
			{ID: "m3", Email: "c@d.com"},
		}

		sut, ctx := newOpsWithCacheSUT(t, func(_ context.Context, d opsWithCacheSUTDeps) {
			d.chanOps.EXPECT().ListMembers(gomock.Any(), teamID, channelID).Return(out, nil).Times(1)

			testutil.ExpectRunNow(d.runner)
			d.cacher.EXPECT().Set(cacher.NewChannelMemberKey(teamID, channelID, "a@b.com", nil), "m1").Return(nil).Times(1)
			d.cacher.EXPECT().Set(cacher.NewChannelMemberKey(teamID, channelID, "c@d.com", nil), "m3").Return(nil).Times(1)
		})

		got, err := sut.ListMembers(ctx, teamID, channelID)
		require.NoError(t, err)
		assert.Equal(t, out, got)
	})

	t.Run("error triggers cache clear", func(t *testing.T) {
		err404 := testutil.ReqErr(http.StatusNotFound)

		sut, ctx := newOpsWithCacheSUT(t, func(_ context.Context, d opsWithCacheSUTDeps) {
			d.chanOps.EXPECT().ListMembers(gomock.Any(), teamID, channelID).Return(nil, err404).Times(1)
			expectClearNow(d)
		})

		got, err := sut.ListMembers(ctx, teamID, channelID)
		require.Nil(t, got)
		require.True(t, err == err404)
	})
}

func TestOpsWithCache_AddMember(t *testing.T) {
	teamID := "team-1"
	channelID := "channel-1"

	t.Run("nil member does not schedule cache update", func(t *testing.T) {
		sut, ctx := newOpsWithCacheSUT(t, func(_ context.Context, d opsWithCacheSUTDeps) {
			d.chanOps.EXPECT().AddMember(gomock.Any(), teamID, channelID, "u1", true).Return(nil, nil).Times(1)
		})

		m, err := sut.AddMember(ctx, teamID, channelID, "u1", true)
		require.NoError(t, err)
		require.Nil(t, m)
	})

	t.Run("success caches member by email", func(t *testing.T) {
		out := &models.Member{ID: "m1", Email: "a@b.com"}

		sut, ctx := newOpsWithCacheSUT(t, func(_ context.Context, d opsWithCacheSUTDeps) {
			d.chanOps.EXPECT().AddMember(gomock.Any(), teamID, channelID, "u1", true).Return(out, nil).Times(1)

			testutil.ExpectRunNow(d.runner)
			d.cacher.EXPECT().Set(cacher.NewChannelMemberKey(teamID, channelID, "a@b.com", nil), "m1").Return(nil).Times(1)
		})

		m, err := sut.AddMember(ctx, teamID, channelID, "u1", true)
		require.NoError(t, err)
		require.Equal(t, out, m)
	})

	t.Run("error triggers cache clear", func(t *testing.T) {
		err400 := testutil.ReqErr(http.StatusBadRequest)

		sut, ctx := newOpsWithCacheSUT(t, func(_ context.Context, d opsWithCacheSUTDeps) {
			d.chanOps.EXPECT().AddMember(gomock.Any(), teamID, channelID, "u1", false).Return(nil, err400).Times(1)
			expectClearNow(d)
		})

		m, err := sut.AddMember(ctx, teamID, channelID, "u1", false)
		require.Nil(t, m)
		require.True(t, err == err400)
	})
}

func TestOpsWithCache_RemoveMember(t *testing.T) {
	teamID := "team-1"
	channelID := "channel-1"
	memberID := "member-1"
	userRef := "user@example.com"

	t.Run("success invalidates member ref", func(t *testing.T) {
		sut, ctx := newOpsWithCacheSUT(t, func(_ context.Context, d opsWithCacheSUTDeps) {
			d.chanOps.EXPECT().RemoveMember(gomock.Any(), teamID, channelID, memberID, userRef).Return(nil).Times(1)

			testutil.ExpectRunNow(d.runner)
			d.cacher.EXPECT().Invalidate(cacher.NewChannelMemberKey(teamID, channelID, userRef, nil)).Return(nil).Times(1)
		})

		err := sut.RemoveMember(ctx, teamID, channelID, memberID, userRef)
		require.NoError(t, err)
	})

	t.Run("error triggers cache clear and does not invalidate", func(t *testing.T) {
		err404 := testutil.ReqErr(http.StatusNotFound)

		sut, ctx := newOpsWithCacheSUT(t, func(_ context.Context, d opsWithCacheSUTDeps) {
			d.chanOps.EXPECT().RemoveMember(gomock.Any(), teamID, channelID, memberID, userRef).Return(err404).Times(1)
			expectClearNow(d)

			d.cacher.EXPECT().Invalidate(gomock.Any()).Times(0)
		})

		err := sut.RemoveMember(ctx, teamID, channelID, memberID, userRef)
		require.True(t, err == err404)
	})
}

func TestOpsWithCache_WithErrorClearMethods_ClearCacheOnError(t *testing.T) {
	err400 := testutil.ReqErr(http.StatusBadRequest)

	type testCase struct {
		name   string
		expect func(d opsWithCacheSUTDeps)
		call   func(sut channelOps, ctx context.Context) (isNil bool, err error)
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
				d.chanOps.EXPECT().SendMessage(gomock.Any(), teamID, channelID, gomock.Any()).Return(nil, err400).Times(1)
			},
			call: func(sut channelOps, ctx context.Context) (bool, error) {
				msg, err := sut.SendMessage(ctx, teamID, channelID, models.MessageBody{})
				return msg == nil, err
			},
		},
		{
			name: "SendReply",
			expect: func(d opsWithCacheSUTDeps) {
				d.chanOps.EXPECT().SendReply(gomock.Any(), teamID, channelID, messageID, gomock.Any()).Return(nil, err400).Times(1)
			},
			call: func(sut channelOps, ctx context.Context) (bool, error) {
				msg, err := sut.SendReply(ctx, teamID, channelID, messageID, models.MessageBody{})
				return msg == nil, err
			},
		},
		{
			name: "ListMessages",
			expect: func(d opsWithCacheSUTDeps) {
				d.chanOps.EXPECT().ListMessages(gomock.Any(), teamID, channelID, gomock.Any(), false).Return(nil, err400).Times(1)
			},
			call: func(sut channelOps, ctx context.Context) (bool, error) {
				msgs, err := sut.ListMessages(ctx, teamID, channelID, nil, false)
				return msgs == nil, err
			},
		},
		{
			name: "ListReplies",
			expect: func(d opsWithCacheSUTDeps) {
				d.chanOps.EXPECT().ListReplies(gomock.Any(), teamID, channelID, messageID, gomock.Any(), false).Return(nil, err400).Times(1)
			},
			call: func(sut channelOps, ctx context.Context) (bool, error) {
				msgs, err := sut.ListReplies(ctx, teamID, channelID, messageID, nil, false)
				return msgs == nil, err
			},
		},
		{
			name: "GetMessage",
			expect: func(d opsWithCacheSUTDeps) {
				d.chanOps.EXPECT().GetMessage(gomock.Any(), teamID, channelID, messageID).Return(nil, err400).Times(1)
			},
			call: func(sut channelOps, ctx context.Context) (bool, error) {
				msg, err := sut.GetMessage(ctx, teamID, channelID, messageID)
				return msg == nil, err
			},
		},
		{
			name: "GetReply",
			expect: func(d opsWithCacheSUTDeps) {
				d.chanOps.EXPECT().GetReply(gomock.Any(), teamID, channelID, messageID, replyID).Return(nil, err400).Times(1)
			},
			call: func(sut channelOps, ctx context.Context) (bool, error) {
				msg, err := sut.GetReply(ctx, teamID, channelID, messageID, replyID)
				return msg == nil, err
			},
		},
		{
			name: "UpdateMemberRoles",
			expect: func(d opsWithCacheSUTDeps) {
				d.chanOps.EXPECT().UpdateMemberRoles(gomock.Any(), teamID, channelID, memberID, true).Return(nil, err400).Times(1)
			},
			call: func(sut channelOps, ctx context.Context) (bool, error) {
				m, err := sut.UpdateMemberRoles(ctx, teamID, channelID, memberID, true)
				return m == nil, err
			},
		},
		{
			name: "GetMentions",
			expect: func(d opsWithCacheSUTDeps) {
				d.chanOps.EXPECT().
					GetMentions(gomock.Any(), teamID, "TeamRef", "ChanRef", channelID, gomock.Any()).
					Return(nil, err400).Times(1)
			},
			call: func(sut channelOps, ctx context.Context) (bool, error) {
				ments, err := sut.GetMentions(ctx, teamID, "TeamRef", "ChanRef", channelID, []string{"@x"})
				return ments == nil, err
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sut, ctx := newOpsWithCacheSUT(t, func(_ context.Context, d opsWithCacheSUTDeps) {
				tc.expect(d)
				expectClearNow(d)
			})

			isNil, err := tc.call(sut, ctx)
			require.True(t, isNil)
			require.Error(t, err)
			require.True(t, err == err400)
		})
	}
}

func TestOpsWithCache_GetMentions(t *testing.T) {
	teamID := "team-1"
	channelID := "channel-1"

	t.Run("success passes through result without touching cache", func(t *testing.T) {
		out := []models.Mention{
			{Kind: models.MentionTeam, Text: "team", AtID: 0},
		}

		sut, ctx := newOpsWithCacheSUT(t, func(_ context.Context, d opsWithCacheSUTDeps) {
			d.chanOps.EXPECT().
				GetMentions(gomock.Any(), teamID, "TeamRef", "ChanRef", channelID, []string{"@team"}).
				Return(out, nil).Times(1)

			d.runner.EXPECT().Run(gomock.Any()).Times(0)
			d.cacher.EXPECT().Clear().Times(0)
			d.cacher.EXPECT().Set(gomock.Any(), gomock.Any()).Times(0)
			d.cacher.EXPECT().Invalidate(gomock.Any()).Times(0)
		})

		got, err := sut.GetMentions(ctx, teamID, "TeamRef", "ChanRef", channelID, []string{"@team"})
		require.NoError(t, err)
		require.Equal(t, out, got)
	})
}
