package channels

import (
	"context"
	"errors"
	"testing"

	snd "github.com/pzsp-teams/lib/internal/sender"
	"github.com/pzsp-teams/lib/internal/testutil"
	"github.com/pzsp-teams/lib/internal/util"
	"github.com/pzsp-teams/lib/models"
	"github.com/pzsp-teams/lib/search"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type sutDeps struct {
	ops             *testutil.MockchannelOps
	teamResolver    *testutil.MockTeamResolver
	channelResolver *testutil.MockChannelResolver
}

func newSUT(t *testing.T, setup func(d sutDeps)) (Service, context.Context) {
	t.Helper()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	d := sutDeps{
		ops:             testutil.NewMockchannelOps(ctrl),
		teamResolver:    testutil.NewMockTeamResolver(ctrl),
		channelResolver: testutil.NewMockChannelResolver(ctrl),
	}

	if setup != nil {
		setup(d)
	}

	return NewService(d.ops, d.teamResolver, d.channelResolver), context.Background()
}

const (
	defaultTeamRef    = "TeamA"
	defaultTeamID     = "team-id"
	defaultChannelRef = "ChanA"
	defaultChannelID  = "chan-id"
)

func expectResolveTeam(t *testing.T, d sutDeps) {
	t.Helper()
	d.teamResolver.EXPECT().
		ResolveTeamRefToID(gomock.Any(), defaultTeamRef).
		Return(defaultTeamID, nil).
		Times(1)
}

func expectResolveTeamAndChannel(t *testing.T, d sutDeps) {
	t.Helper()

	expectResolveTeam(t, d)

	d.channelResolver.EXPECT().
		ResolveChannelRefToID(gomock.Any(), defaultTeamID, defaultChannelRef).
		Return(defaultChannelID, nil).
		Times(1)
}

func TestService_ListChannels(t *testing.T) {
	type testCase struct {
		name       string
		teamRef    string
		setupMocks func(d sutDeps)
		wantLen    int
		assertErr  func(t *testing.T, err error)
	}

	testCases := []testCase{
		{
			name:    "success returns channels",
			teamRef: "TeamA",
			setupMocks: func(d sutDeps) {
				expectResolveTeam(t, d)
				d.ops.EXPECT().
					ListChannelsByTeamID(gomock.Any(), "team-id").
					Return([]*models.Channel{
						{ID: "1", Name: "General", IsGeneral: true},
						{ID: "2", Name: "Random"},
					}, nil).
					Times(1)
			},
			wantLen: 2,
		},
		{
			name:    "team resolver error is propagated",
			teamRef: "TeamA",
			setupMocks: func(d sutDeps) {
				d.teamResolver.EXPECT().
					ResolveTeamRefToID(gomock.Any(), "TeamA").
					Return("", errors.New("boom")).
					Times(1)
			},
			assertErr: func(t *testing.T, err error) { require.Error(t, err) },
		},
		{
			name:    "ops error is propagated",
			teamRef: "TeamA",
			setupMocks: func(d sutDeps) {
				expectResolveTeam(t, d)
				d.ops.EXPECT().
					ListChannelsByTeamID(gomock.Any(), "team-id").
					Return(nil, &snd.ErrAccessForbidden{Code: 403, OriginalMessage: "nope"}).
					Times(1)
			},
			assertErr: func(t *testing.T, err error) { testutil.RequireReqErrCode(t, err, 403) },
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc, ctx := newSUT(t, tc.setupMocks)

			got, err := svc.ListChannels(ctx, tc.teamRef)

			if tc.assertErr != nil {
				tc.assertErr(t, err)
				return
			}

			require.NoError(t, err)
			require.Len(t, got, tc.wantLen)
			assert.Equal(t, "1", got[0].ID)
		})
	}
}

func TestService_Get(t *testing.T) {
	type testCase struct {
		name       string
		setupMocks func(d sutDeps)
		wantID     string
		assertErr  func(t *testing.T, err error)
	}

	testCases := []testCase{
		{
			name: "success resolves team+channel and gets channel",
			setupMocks: func(d sutDeps) {
				expectResolveTeamAndChannel(t, d)
				d.ops.EXPECT().
					GetChannelByID(gomock.Any(), "team-id", "chan-id").
					Return(&models.Channel{ID: "chan-id", Name: "ChanA"}, nil).
					Times(1)
			},
			wantID: "chan-id",
		},
		{
			name: "team resolver error is propagated",
			setupMocks: func(d sutDeps) {
				d.teamResolver.EXPECT().
					ResolveTeamRefToID(gomock.Any(), "TeamA").
					Return("", errors.New("boom")).
					Times(1)
			},
			assertErr: func(t *testing.T, err error) { require.Error(t, err) },
		},
		{
			name: "channel resolver error is propagated",
			setupMocks: func(d sutDeps) {
				expectResolveTeam(t, d)
				d.channelResolver.EXPECT().
					ResolveChannelRefToID(gomock.Any(), "team-id", "ChanA").
					Return("", errors.New("boom")).
					Times(1)
			},
			assertErr: func(t *testing.T, err error) { require.Error(t, err) },
		},
		{
			name: "ops error is propagated",
			setupMocks: func(d sutDeps) {
				expectResolveTeamAndChannel(t, d)
				d.ops.EXPECT().
					GetChannelByID(gomock.Any(), "team-id", "chan-id").
					Return(nil, &snd.ErrResourceNotFound{Code: 404, OriginalMessage: "missing"}).
					Times(1)
			},
			assertErr: func(t *testing.T, err error) { testutil.RequireReqErrCode(t, err, 404) },
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc, ctx := newSUT(t, tc.setupMocks)

			got, err := svc.Get(ctx, "TeamA", "ChanA")

			if tc.assertErr != nil {
				tc.assertErr(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, tc.wantID, got.ID)
		})
	}
}

func TestService_CreateStandardChannel(t *testing.T) {
	type testCase struct {
		name       string
		setupMocks func(d sutDeps)
		assertErr  func(t *testing.T, err error)
	}

	testCases := []testCase{
		{
			name: "success",
			setupMocks: func(d sutDeps) {
				expectResolveTeam(t, d)
				d.ops.EXPECT().
					CreateStandardChannel(gomock.Any(), "team-id", "NewChan").
					Return(&models.Channel{ID: "c1", Name: "NewChan"}, nil).
					Times(1)
			},
		},
		{
			name: "team resolver error",
			setupMocks: func(d sutDeps) {
				d.teamResolver.EXPECT().
					ResolveTeamRefToID(gomock.Any(), "TeamA").
					Return("", errors.New("boom")).
					Times(1)
			},
			assertErr: func(t *testing.T, err error) { require.Error(t, err) },
		},
		{
			name: "ops error propagated",
			setupMocks: func(d sutDeps) {
				expectResolveTeam(t, d)
				d.ops.EXPECT().
					CreateStandardChannel(gomock.Any(), "team-id", "NewChan").
					Return(nil, &snd.ErrAccessForbidden{Code: 403, OriginalMessage: "nope"}).
					Times(1)
			},
			assertErr: func(t *testing.T, err error) { testutil.RequireReqErrCode(t, err, 403) },
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc, ctx := newSUT(t, tc.setupMocks)

			got, err := svc.CreateStandardChannel(ctx, "TeamA", "NewChan")

			if tc.assertErr != nil {
				tc.assertErr(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, "c1", got.ID)
			assert.Equal(t, "NewChan", got.Name)
		})
	}
}

func TestService_CreatePrivateChannel(t *testing.T) {
	type testCase struct {
		name       string
		setupMocks func(d sutDeps)
		assertErr  func(t *testing.T, err error)
	}

	members := []string{"u1", "u2"}
	owners := []string{"o1"}

	testCases := []testCase{
		{
			name: "success",
			setupMocks: func(d sutDeps) {
				expectResolveTeam(t, d)
				d.ops.EXPECT().
					CreatePrivateChannel(gomock.Any(), "team-id", "Secret", members, owners).
					Return(&models.Channel{ID: "pc1", Name: "Secret"}, nil).
					Times(1)
			},
		},
		{
			name: "team resolver error",
			setupMocks: func(d sutDeps) {
				d.teamResolver.EXPECT().
					ResolveTeamRefToID(gomock.Any(), "TeamA").
					Return("", errors.New("boom")).
					Times(1)
			},
			assertErr: func(t *testing.T, err error) { require.Error(t, err) },
		},
		{
			name: "ops error propagated",
			setupMocks: func(d sutDeps) {
				expectResolveTeam(t, d)
				d.ops.EXPECT().
					CreatePrivateChannel(gomock.Any(), "team-id", "Secret", members, owners).
					Return(nil, &snd.ErrResourceNotFound{Code: 404, OriginalMessage: "missing"}).
					Times(1)
			},
			assertErr: func(t *testing.T, err error) { testutil.RequireReqErrCode(t, err, 404) },
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc, ctx := newSUT(t, tc.setupMocks)

			got, err := svc.CreatePrivateChannel(ctx, "TeamA", "Secret", members, owners)

			if tc.assertErr != nil {
				tc.assertErr(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, "pc1", got.ID)
		})
	}
}

func TestService_Delete(t *testing.T) {
	type testCase struct {
		name       string
		setupMocks func(d sutDeps)
		assertErr  func(t *testing.T, err error)
	}

	testCases := []testCase{
		{
			name: "success passes channelRef to ops",
			setupMocks: func(d sutDeps) {
				expectResolveTeamAndChannel(t, d)
				d.ops.EXPECT().
					DeleteChannel(gomock.Any(), "team-id", "chan-id", "ChanA").
					Return(nil).
					Times(1)
			},
		},
		{
			name: "ops error propagated",
			setupMocks: func(d sutDeps) {
				expectResolveTeamAndChannel(t, d)
				d.ops.EXPECT().
					DeleteChannel(gomock.Any(), "team-id", "chan-id", "ChanA").
					Return(&snd.ErrAccessForbidden{Code: 403, OriginalMessage: "nope"}).
					Times(1)
			},
			assertErr: func(t *testing.T, err error) { testutil.RequireReqErrCode(t, err, 403) },
		},
		{
			name: "team resolver error",
			setupMocks: func(d sutDeps) {
				d.teamResolver.EXPECT().
					ResolveTeamRefToID(gomock.Any(), "TeamA").
					Return("", errors.New("boom")).
					Times(1)
			},
			assertErr: func(t *testing.T, err error) { require.Error(t, err) },
		},
		{
			name: "channel resolver error",
			setupMocks: func(d sutDeps) {
				expectResolveTeam(t, d)
				d.channelResolver.EXPECT().
					ResolveChannelRefToID(gomock.Any(), "team-id", "ChanA").
					Return("", errors.New("boom")).
					Times(1)
			},
			assertErr: func(t *testing.T, err error) { require.Error(t, err) },
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc, ctx := newSUT(t, tc.setupMocks)
			err := svc.Delete(ctx, "TeamA", "ChanA")
			if tc.assertErr != nil {
				tc.assertErr(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestService_SendMessage_SendReply(t *testing.T) {
	type testCase struct {
		name       string
		setupMocks func(d sutDeps)
		call       func(svc Service, ctx context.Context) error
		assertErr  func(t *testing.T, err error)
	}

	bodyMsg := models.MessageBody{Content: "hi", ContentType: models.MessageContentTypeText}
	bodyReply := models.MessageBody{Content: "reply", ContentType: models.MessageContentTypeText}

	testCases := []testCase{
		{
			name: "SendMessage success",
			setupMocks: func(d sutDeps) {
				expectResolveTeamAndChannel(t, d)
				d.ops.EXPECT().
					SendMessage(gomock.Any(), "team-id", "chan-id", bodyMsg).
					Return(&models.Message{ID: "m1"}, nil).
					Times(1)
			},
			call: func(svc Service, ctx context.Context) error {
				_, err := svc.SendMessage(ctx, "TeamA", "ChanA", bodyMsg)
				return err
			},
		},
		{
			name: "SendReply success",
			setupMocks: func(d sutDeps) {
				expectResolveTeamAndChannel(t, d)
				d.ops.EXPECT().
					SendReply(gomock.Any(), "team-id", "chan-id", "msg-1", bodyReply).
					Return(&models.Message{ID: "r1"}, nil).
					Times(1)
			},
			call: func(svc Service, ctx context.Context) error {
				_, err := svc.SendReply(ctx, "TeamA", "ChanA", "msg-1", bodyReply)
				return err
			},
		},
		{
			name: "ops error propagated",
			setupMocks: func(d sutDeps) {
				expectResolveTeamAndChannel(t, d)
				d.ops.EXPECT().
					SendMessage(gomock.Any(), "team-id", "chan-id", bodyMsg).
					Return(nil, &snd.ErrAccessForbidden{Code: 403, OriginalMessage: "nope"}).
					Times(1)
			},
			call: func(svc Service, ctx context.Context) error {
				_, err := svc.SendMessage(ctx, "TeamA", "ChanA", bodyMsg)
				return err
			},
			assertErr: func(t *testing.T, err error) { testutil.RequireReqErrCode(t, err, 403) },
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc, ctx := newSUT(t, tc.setupMocks)
			err := tc.call(svc, ctx)
			if tc.assertErr != nil {
				tc.assertErr(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestService_ListMessages_ListReplies(t *testing.T) {
	t.Run("ListMessages passes opts through", func(t *testing.T) {
		top := int32(5)
		opts := &models.ListMessagesOptions{Top: &top}

		svc, ctx := newSUT(t, func(d sutDeps) {
			expectResolveTeamAndChannel(t, d)
			d.ops.EXPECT().
				ListMessages(gomock.Any(), "team-id", "chan-id", opts, false).
				Return(&models.MessageCollection{Messages: []*models.Message{{ID: "m1"}, {ID: "m2"}}}, nil).
				Times(1)
		})

		got, err := svc.ListMessages(ctx, "TeamA", "ChanA", opts, false, nil)
		require.NoError(t, err)
		require.Len(t, got.Messages, 2)
	})

	t.Run("ListReplies builds opts from top pointer and passes it", func(t *testing.T) {
		top := int32(10)

		svc, ctx := newSUT(t, func(d sutDeps) {
			expectResolveTeamAndChannel(t, d)
			d.ops.EXPECT().
				ListReplies(gomock.Any(), "team-id", "chan-id", "msg-1", gomock.Any(), false).
				DoAndReturn(func(_ context.Context, teamID, channelID, msgID string, opts *models.ListMessagesOptions, includeSystemEvents bool) (*models.MessageCollection, error) {
					require.Equal(t, "team-id", teamID)
					require.Equal(t, "chan-id", channelID)
					require.Equal(t, "msg-1", msgID)
					require.NotNil(t, opts)
					require.NotNil(t, opts.Top)
					assert.Equal(t, top, *opts.Top)
					return &models.MessageCollection{Messages: []*models.Message{{ID: "r1"}}}, nil
				}).
				Times(1)
		})

		got, err := svc.ListReplies(ctx, "TeamA", "ChanA", "msg-1", &top, false, nil)
		require.NoError(t, err)
		require.Len(t, got.Messages, 1)
	})
}

func TestService_ListMessages_WithNextLink(t *testing.T) {
	next := "next-link-1"

	svc, ctx := newSUT(t, func(d sutDeps) {
		expectResolveTeamAndChannel(t, d)
		d.ops.EXPECT().
			ListMessagesNext(gomock.Any(), defaultTeamID, defaultChannelID, next, true).
			Return(&models.MessageCollection{Messages: []*models.Message{{ID: "m1"}}}, nil).
			Times(1)

		d.ops.EXPECT().
			ListMessages(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Times(0)
	})

	got, err := svc.ListMessages(ctx, defaultTeamRef, defaultChannelRef, &models.ListMessagesOptions{}, true, &next)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Len(t, got.Messages, 1)
	assert.Equal(t, "m1", got.Messages[0].ID)
}

func TestService_ListMessages_WithNextLink_Error(t *testing.T) {
	next := "next-link-1"

	svc, ctx := newSUT(t, func(d sutDeps) {
		expectResolveTeamAndChannel(t, d)

		d.ops.EXPECT().
			ListMessagesNext(gomock.Any(), defaultTeamID, defaultChannelID, next, false).
			Return(nil, &snd.ErrAccessForbidden{Code: 403, OriginalMessage: "nope"}).
			Times(1)

		d.ops.EXPECT().
			ListMessages(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Times(0)
	})

	_, err := svc.ListMessages(ctx, defaultTeamRef, defaultChannelRef, nil, false, &next)
	testutil.RequireReqErrCode(t, err, 403)
}

func TestService_GetMessage_GetReply(t *testing.T) {
	t.Run("GetMessage delegates", func(t *testing.T) {
		svc, ctx := newSUT(t, func(d sutDeps) {
			expectResolveTeamAndChannel(t, d)
			d.ops.EXPECT().
				GetMessage(gomock.Any(), "team-id", "chan-id", "m1").
				Return(&models.Message{ID: "m1"}, nil).
				Times(1)
		})

		got, err := svc.GetMessage(ctx, "TeamA", "ChanA", "m1")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "m1", got.ID)
	})

	t.Run("GetReply delegates", func(t *testing.T) {
		svc, ctx := newSUT(t, func(d sutDeps) {
			expectResolveTeamAndChannel(t, d)
			d.ops.EXPECT().
				GetReply(gomock.Any(), "team-id", "chan-id", "m1", "r1").
				Return(&models.Message{ID: "r1"}, nil).
				Times(1)
		})

		got, err := svc.GetReply(ctx, "TeamA", "ChanA", "m1", "r1")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "r1", got.ID)
	})
}

func TestService_ListMembers_AddMember(t *testing.T) {
	t.Run("ListMembers delegates", func(t *testing.T) {
		svc, ctx := newSUT(t, func(d sutDeps) {
			expectResolveTeamAndChannel(t, d)
			d.ops.EXPECT().
				ListMembers(gomock.Any(), "team-id", "chan-id").
				Return([]*models.Member{{ID: "m1"}, {ID: "m2"}}, nil).
				Times(1)
		})

		got, err := svc.ListMembers(ctx, "TeamA", "ChanA")
		require.NoError(t, err)
		require.Len(t, got, 2)
	})

	t.Run("AddMember delegates", func(t *testing.T) {
		svc, ctx := newSUT(t, func(d sutDeps) {
			expectResolveTeamAndChannel(t, d)
			d.ops.EXPECT().
				AddMember(gomock.Any(), "team-id", "chan-id", "user@x.com", true).
				Return(&models.Member{ID: "mem-1"}, nil).
				Times(1)
		})

		got, err := svc.AddMember(ctx, "TeamA", "ChanA", "user@x.com", true)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "mem-1", got.ID)
	})
}

func TestService_UpdateMemberRoles_RemoveMember(t *testing.T) {
	t.Run("UpdateMemberRoles resolves memberID and delegates", func(t *testing.T) {
		svc, ctx := newSUT(t, func(d sutDeps) {
			expectResolveTeamAndChannel(t, d)

			d.channelResolver.EXPECT().
				ResolveChannelMemberRefToID(gomock.Any(), "team-id", "chan-id", "user@x.com").
				Return("member-id", nil).
				Times(1)

			d.ops.EXPECT().
				UpdateMemberRoles(gomock.Any(), "team-id", "chan-id", "member-id", true).
				Return(&models.Member{ID: "member-id"}, nil).
				Times(1)
		})

		got, err := svc.UpdateMemberRoles(ctx, "TeamA", "ChanA", "user@x.com", true)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "member-id", got.ID)
	})

	t.Run("RemoveMember resolves memberID and passes userRef to ops", func(t *testing.T) {
		svc, ctx := newSUT(t, func(d sutDeps) {
			expectResolveTeamAndChannel(t, d)

			d.channelResolver.EXPECT().
				ResolveChannelMemberRefToID(gomock.Any(), "team-id", "chan-id", "user@x.com").
				Return("member-id", nil).
				Times(1)

			d.ops.EXPECT().
				RemoveMember(gomock.Any(), "team-id", "chan-id", "member-id", "user@x.com").
				Return(nil).
				Times(1)
		})

		err := svc.RemoveMember(ctx, "TeamA", "ChanA", "user@x.com")
		require.NoError(t, err)
	})
}

func TestService_GetMentions(t *testing.T) {
	type testCase struct {
		name       string
		raw        []string
		setupMocks func(d sutDeps)
		wantLen    int
		assertErr  func(t *testing.T, err error)
	}

	testCases := []testCase{
		{
			name: "delegates to ops with resolved ids and original refs",
			raw:  []string{"  ", "alice@example.com", "team", "channel"},
			setupMocks: func(d sutDeps) {
				expectResolveTeamAndChannel(t, d)

				d.ops.EXPECT().
					GetMentions(gomock.Any(), "team-id", "TeamA", "ChanA", "chan-id", gomock.Any()).
					DoAndReturn(func(
						_ context.Context,
						teamID, teamRef, channelRef, channelID string,
						raw []string,
					) ([]models.Mention, error) {
						require.Equal(t, "team-id", teamID)
						require.Equal(t, "TeamA", teamRef)
						require.Equal(t, "ChanA", channelRef)
						require.Equal(t, "chan-id", channelID)
						require.Equal(t, []string{"  ", "alice@example.com", "team", "channel"}, raw)
						return []models.Mention{
							{Kind: models.MentionUser, TargetID: "u-1", Text: "Alice", AtID: 0},
							{Kind: models.MentionTeam, TargetID: "team-id", Text: "TeamA", AtID: 1},
						}, nil
					}).
					Times(1)
			},
			wantLen: 2,
		},
		{
			name: "team resolver error is propagated",
			raw:  []string{"x"},
			setupMocks: func(d sutDeps) {
				d.teamResolver.EXPECT().
					ResolveTeamRefToID(gomock.Any(), "TeamA").
					Return("", errors.New("boom")).
					Times(1)
			},
			assertErr: func(t *testing.T, err error) { require.Error(t, err) },
		},
		{
			name: "channel resolver error is propagated",
			raw:  []string{"x"},
			setupMocks: func(d sutDeps) {
				expectResolveTeam(t, d)
				d.channelResolver.EXPECT().
					ResolveChannelRefToID(gomock.Any(), "team-id", "ChanA").
					Return("", errors.New("boom")).
					Times(1)
			},
			assertErr: func(t *testing.T, err error) { require.Error(t, err) },
		},
		{
			name: "ops error is propagated",
			raw:  []string{"x"},
			setupMocks: func(d sutDeps) {
				expectResolveTeamAndChannel(t, d)
				d.ops.EXPECT().
					GetMentions(gomock.Any(), "team-id", "TeamA", "ChanA", "chan-id", []string{"x"}).
					Return(nil, &snd.ErrAccessForbidden{Code: 403, OriginalMessage: "nope"}).
					Times(1)
			},
			assertErr: func(t *testing.T, err error) { testutil.RequireReqErrCode(t, err, 403) },
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc, ctx := newSUT(t, tc.setupMocks)

			got, err := svc.GetMentions(ctx, "TeamA", "ChanA", tc.raw)

			if tc.assertErr != nil {
				tc.assertErr(t, err)
				return
			}

			require.NoError(t, err)
			require.Len(t, got, tc.wantLen)
		})
	}
}

func TestService_SendMessage_Errors(t *testing.T) {
	type testCase struct {
		name       string
		setupMocks func(d sutDeps)
		call       func(svc Service, ctx context.Context) error
		assertErr  func(t *testing.T, err error)
	}

	body := models.MessageBody{Content: "hi", ContentType: models.MessageContentTypeText}

	testCases := []testCase{
		{
			name: "team resolver error",
			setupMocks: func(d sutDeps) {
				d.teamResolver.EXPECT().
					ResolveTeamRefToID(gomock.Any(), defaultTeamRef).
					Return("", errors.New("boom")).
					Times(1)
			},
			call: func(svc Service, ctx context.Context) error {
				_, err := svc.SendMessage(ctx, defaultTeamRef, defaultChannelRef, body)
				return err
			},
			assertErr: func(t *testing.T, err error) { require.Error(t, err) },
		},
		{
			name: "channel resolver error",
			setupMocks: func(d sutDeps) {
				expectResolveTeam(t, d)
				d.channelResolver.EXPECT().
					ResolveChannelRefToID(gomock.Any(), defaultTeamID, defaultChannelRef).
					Return("", errors.New("boom")).
					Times(1)
			},
			call: func(svc Service, ctx context.Context) error {
				_, err := svc.SendMessage(ctx, defaultTeamRef, defaultChannelRef, body)
				return err
			},
			assertErr: func(t *testing.T, err error) { require.Error(t, err) },
		},
		{
			name: "ops error is wrapped",
			setupMocks: func(d sutDeps) {
				expectResolveTeamAndChannel(t, d)
				d.ops.EXPECT().
					SendMessage(gomock.Any(), defaultTeamID, defaultChannelID, body).
					Return(nil, &snd.ErrAccessForbidden{Code: 403, OriginalMessage: "nope"}).
					Times(1)
			},
			call: func(svc Service, ctx context.Context) error {
				_, err := svc.SendMessage(ctx, defaultTeamRef, defaultChannelRef, body)
				return err
			},
			assertErr: func(t *testing.T, err error) { testutil.RequireReqErrCode(t, err, 403) },
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc, ctx := newSUT(t, tc.setupMocks)
			err := tc.call(svc, ctx)
			tc.assertErr(t, err)
		})
	}
}

func TestService_SendReply_Errors(t *testing.T) {
	type testCase struct {
		name       string
		setupMocks func(d sutDeps)
		call       func(svc Service, ctx context.Context) error
		assertErr  func(t *testing.T, err error)
	}

	body := models.MessageBody{Content: "reply", ContentType: models.MessageContentTypeText}

	testCases := []testCase{
		{
			name: "team resolver error",
			setupMocks: func(d sutDeps) {
				d.teamResolver.EXPECT().
					ResolveTeamRefToID(gomock.Any(), defaultTeamRef).
					Return("", errors.New("boom")).
					Times(1)
			},
			call: func(svc Service, ctx context.Context) error {
				_, err := svc.SendReply(ctx, defaultTeamRef, defaultChannelRef, "msg-1", body)
				return err
			},
			assertErr: func(t *testing.T, err error) { require.Error(t, err) },
		},
		{
			name: "channel resolver error",
			setupMocks: func(d sutDeps) {
				expectResolveTeam(t, d)
				d.channelResolver.EXPECT().
					ResolveChannelRefToID(gomock.Any(), defaultTeamID, defaultChannelRef).
					Return("", errors.New("boom")).
					Times(1)
			},
			call: func(svc Service, ctx context.Context) error {
				_, err := svc.SendReply(ctx, defaultTeamRef, defaultChannelRef, "msg-1", body)
				return err
			},
			assertErr: func(t *testing.T, err error) { require.Error(t, err) },
		},
		{
			name: "ops error is wrapped",
			setupMocks: func(d sutDeps) {
				expectResolveTeamAndChannel(t, d)
				d.ops.EXPECT().
					SendReply(gomock.Any(), defaultTeamID, defaultChannelID, "msg-1", body).
					Return(nil, &snd.ErrResourceNotFound{Code: 404, OriginalMessage: "missing"}).
					Times(1)
			},
			call: func(svc Service, ctx context.Context) error {
				_, err := svc.SendReply(ctx, defaultTeamRef, defaultChannelRef, "msg-1", body)
				return err
			},
			assertErr: func(t *testing.T, err error) { testutil.RequireReqErrCode(t, err, 404) },
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc, ctx := newSUT(t, tc.setupMocks)
			err := tc.call(svc, ctx)
			tc.assertErr(t, err)
		})
	}
}

func TestService_ListMessages_Errors(t *testing.T) {
	type testCase struct {
		name       string
		setupMocks func(d sutDeps)
		assertErr  func(t *testing.T, err error)
	}

	opts := &models.ListMessagesOptions{}

	testCases := []testCase{
		{
			name: "team resolver error",
			setupMocks: func(d sutDeps) {
				d.teamResolver.EXPECT().
					ResolveTeamRefToID(gomock.Any(), defaultTeamRef).
					Return("", errors.New("boom")).
					Times(1)
			},
			assertErr: func(t *testing.T, err error) { require.Error(t, err) },
		},
		{
			name: "channel resolver error",
			setupMocks: func(d sutDeps) {
				expectResolveTeam(t, d)
				d.channelResolver.EXPECT().
					ResolveChannelRefToID(gomock.Any(), defaultTeamID, defaultChannelRef).
					Return("", errors.New("boom")).
					Times(1)
			},
			assertErr: func(t *testing.T, err error) { require.Error(t, err) },
		},
		{
			name: "ops error is wrapped",
			setupMocks: func(d sutDeps) {
				expectResolveTeamAndChannel(t, d)
				d.ops.EXPECT().
					ListMessages(gomock.Any(), defaultTeamID, defaultChannelID, opts, false).
					Return(nil, &snd.ErrAccessForbidden{Code: 403, OriginalMessage: "nope"}).
					Times(1)
			},
			assertErr: func(t *testing.T, err error) { testutil.RequireReqErrCode(t, err, 403) },
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc, ctx := newSUT(t, tc.setupMocks)
			_, err := svc.ListMessages(ctx, defaultTeamRef, defaultChannelRef, opts, false, nil)
			tc.assertErr(t, err)
		})
	}
}

func TestService_ListReplies_Errors_AndNilTop(t *testing.T) {
	t.Run("ListReplies passes nil Top when input top is nil", func(t *testing.T) {
		svc, ctx := newSUT(t, func(d sutDeps) {
			expectResolveTeamAndChannel(t, d)

			d.ops.EXPECT().
				ListReplies(gomock.Any(), defaultTeamID, defaultChannelID, "msg-1", gomock.Any(), false).
				DoAndReturn(func(_ context.Context, teamID, channelID, messageID string, opts *models.ListMessagesOptions, includeSystemEvents bool) ([]*models.Message, error) {
					require.NotNil(t, opts)
					require.Nil(t, opts.Top)
					return []*models.Message{{ID: "r1"}}, nil
				}).
				Times(1)
		})

		_, err := svc.ListReplies(ctx, defaultTeamRef, defaultChannelRef, "msg-1", nil, false, nil)
		require.NoError(t, err)
	})

	type testCase struct {
		name       string
		setupMocks func(d sutDeps)
		assertErr  func(t *testing.T, err error)
	}

	top := int32(1)

	testCases := []testCase{
		{
			name: "team resolver error",
			setupMocks: func(d sutDeps) {
				d.teamResolver.EXPECT().
					ResolveTeamRefToID(gomock.Any(), defaultTeamRef).
					Return("", errors.New("boom")).
					Times(1)
			},
			assertErr: func(t *testing.T, err error) { require.Error(t, err) },
		},
		{
			name: "channel resolver error",
			setupMocks: func(d sutDeps) {
				expectResolveTeam(t, d)
				d.channelResolver.EXPECT().
					ResolveChannelRefToID(gomock.Any(), defaultTeamID, defaultChannelRef).
					Return("", errors.New("boom")).
					Times(1)
			},
			assertErr: func(t *testing.T, err error) { require.Error(t, err) },
		},
		{
			name: "ops error is wrapped",
			setupMocks: func(d sutDeps) {
				expectResolveTeamAndChannel(t, d)
				d.ops.EXPECT().
					ListReplies(gomock.Any(), defaultTeamID, defaultChannelID, "msg-1", gomock.Any(), false).
					Return(nil, &snd.ErrAccessForbidden{Code: 403, OriginalMessage: "nope"}).
					Times(1)
			},
			assertErr: func(t *testing.T, err error) { testutil.RequireReqErrCode(t, err, 403) },
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc, ctx := newSUT(t, tc.setupMocks)
			_, err := svc.ListReplies(ctx, defaultTeamRef, defaultChannelRef, "msg-1", &top, false, nil)
			tc.assertErr(t, err)
		})
	}
}

func TestService_ListReplies_WithNextLink(t *testing.T) {
	next := "next-link-2"

	svc, ctx := newSUT(t, func(d sutDeps) {
		expectResolveTeamAndChannel(t, d)

		d.ops.EXPECT().
			ListRepliesNext(gomock.Any(), defaultTeamID, defaultChannelID, "msg-1", next, true).
			Return(&models.MessageCollection{Messages: []*models.Message{{ID: "r1"}}}, nil).
			Times(1)

		d.ops.EXPECT().
			ListReplies(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Times(0)
	})

	got, err := svc.ListReplies(ctx, defaultTeamRef, defaultChannelRef, "msg-1", util.Ptr[int32](1), true, &next)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Len(t, got.Messages, 1)
	assert.Equal(t, "r1", got.Messages[0].ID)
}

func TestService_ListReplies_WithNextLink_Error(t *testing.T) {
	next := "next-link-2"

	svc, ctx := newSUT(t, func(d sutDeps) {
		expectResolveTeamAndChannel(t, d)

		d.ops.EXPECT().
			ListRepliesNext(gomock.Any(), defaultTeamID, defaultChannelID, "msg-1", next, false).
			Return(nil, &snd.ErrResourceNotFound{Code: 404, OriginalMessage: "missing"}).
			Times(1)

		d.ops.EXPECT().
			ListReplies(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Times(0)
	})

	_, err := svc.ListReplies(ctx, defaultTeamRef, defaultChannelRef, "msg-1", nil, false, &next)
	testutil.RequireReqErrCode(t, err, 404)
}

func TestService_ListReplies_NilTop_PassesNilTopInOpts(t *testing.T) {
	svc, ctx := newSUT(t, func(d sutDeps) {
		expectResolveTeamAndChannel(t, d)

		d.ops.EXPECT().
			ListReplies(gomock.Any(), defaultTeamID, defaultChannelID, "msg-1", gomock.Any(), false).
			DoAndReturn(func(_ context.Context, teamID, channelID, messageID string, opts *models.ListMessagesOptions, includeSystem bool) (*models.MessageCollection, error) {
				require.Equal(t, defaultTeamID, teamID)
				require.Equal(t, defaultChannelID, channelID)
				require.Equal(t, "msg-1", messageID)

				require.NotNil(t, opts)
				require.Nil(t, opts.Top)

				return &models.MessageCollection{Messages: []*models.Message{{ID: "r1"}}}, nil
			}).
			Times(1)
	})

	_, err := svc.ListReplies(ctx, defaultTeamRef, defaultChannelRef, "msg-1", nil, false, nil)
	require.NoError(t, err)
}

func TestService_GetMessage_Errors(t *testing.T) {
	type testCase struct {
		name       string
		setupMocks func(d sutDeps)
		assertErr  func(t *testing.T, err error)
	}

	testCases := []testCase{
		{
			name: "team resolver error",
			setupMocks: func(d sutDeps) {
				d.teamResolver.EXPECT().
					ResolveTeamRefToID(gomock.Any(), defaultTeamRef).
					Return("", errors.New("boom")).
					Times(1)
			},
			assertErr: func(t *testing.T, err error) { require.Error(t, err) },
		},
		{
			name: "channel resolver error",
			setupMocks: func(d sutDeps) {
				expectResolveTeam(t, d)
				d.channelResolver.EXPECT().
					ResolveChannelRefToID(gomock.Any(), defaultTeamID, defaultChannelRef).
					Return("", errors.New("boom")).
					Times(1)
			},
			assertErr: func(t *testing.T, err error) { require.Error(t, err) },
		},
		{
			name: "ops error wrapped",
			setupMocks: func(d sutDeps) {
				expectResolveTeamAndChannel(t, d)
				d.ops.EXPECT().
					GetMessage(gomock.Any(), defaultTeamID, defaultChannelID, "m1").
					Return(nil, &snd.ErrResourceNotFound{Code: 404, OriginalMessage: "missing"}).
					Times(1)
			},
			assertErr: func(t *testing.T, err error) { testutil.RequireReqErrCode(t, err, 404) },
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc, ctx := newSUT(t, tc.setupMocks)
			_, err := svc.GetMessage(ctx, defaultTeamRef, defaultChannelRef, "m1")
			tc.assertErr(t, err)
		})
	}
}

func TestService_GetReply_Errors(t *testing.T) {
	type testCase struct {
		name       string
		setupMocks func(d sutDeps)
		assertErr  func(t *testing.T, err error)
	}

	testCases := []testCase{
		{
			name: "team resolver error",
			setupMocks: func(d sutDeps) {
				d.teamResolver.EXPECT().
					ResolveTeamRefToID(gomock.Any(), defaultTeamRef).
					Return("", errors.New("boom")).
					Times(1)
			},
			assertErr: func(t *testing.T, err error) { require.Error(t, err) },
		},
		{
			name: "channel resolver error",
			setupMocks: func(d sutDeps) {
				expectResolveTeam(t, d)
				d.channelResolver.EXPECT().
					ResolveChannelRefToID(gomock.Any(), defaultTeamID, defaultChannelRef).
					Return("", errors.New("boom")).
					Times(1)
			},
			assertErr: func(t *testing.T, err error) { require.Error(t, err) },
		},
		{
			name: "ops error wrapped",
			setupMocks: func(d sutDeps) {
				expectResolveTeamAndChannel(t, d)
				d.ops.EXPECT().
					GetReply(gomock.Any(), defaultTeamID, defaultChannelID, "m1", "r1").
					Return(nil, &snd.ErrAccessForbidden{Code: 403, OriginalMessage: "nope"}).
					Times(1)
			},
			assertErr: func(t *testing.T, err error) { testutil.RequireReqErrCode(t, err, 403) },
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc, ctx := newSUT(t, tc.setupMocks)
			_, err := svc.GetReply(ctx, defaultTeamRef, defaultChannelRef, "m1", "r1")
			tc.assertErr(t, err)
		})
	}
}

func TestService_ListMembers_Errors(t *testing.T) {
	type testCase struct {
		name       string
		setupMocks func(d sutDeps)
		assertErr  func(t *testing.T, err error)
	}

	testCases := []testCase{
		{
			name: "team resolver error",
			setupMocks: func(d sutDeps) {
				d.teamResolver.EXPECT().
					ResolveTeamRefToID(gomock.Any(), defaultTeamRef).
					Return("", errors.New("boom")).
					Times(1)
			},
			assertErr: func(t *testing.T, err error) { require.Error(t, err) },
		},
		{
			name: "channel resolver error",
			setupMocks: func(d sutDeps) {
				expectResolveTeam(t, d)
				d.channelResolver.EXPECT().
					ResolveChannelRefToID(gomock.Any(), defaultTeamID, defaultChannelRef).
					Return("", errors.New("boom")).
					Times(1)
			},
			assertErr: func(t *testing.T, err error) { require.Error(t, err) },
		},
		{
			name: "ops error wrapped",
			setupMocks: func(d sutDeps) {
				expectResolveTeamAndChannel(t, d)
				d.ops.EXPECT().
					ListMembers(gomock.Any(), defaultTeamID, defaultChannelID).
					Return(nil, &snd.ErrAccessForbidden{Code: 403, OriginalMessage: "nope"}).
					Times(1)
			},
			assertErr: func(t *testing.T, err error) { testutil.RequireReqErrCode(t, err, 403) },
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc, ctx := newSUT(t, tc.setupMocks)
			_, err := svc.ListMembers(ctx, defaultTeamRef, defaultChannelRef)
			tc.assertErr(t, err)
		})
	}
}

func TestService_AddMember_Errors(t *testing.T) {
	type testCase struct {
		name       string
		setupMocks func(d sutDeps)
		assertErr  func(t *testing.T, err error)
	}

	userRef := "user@x.com"

	testCases := []testCase{
		{
			name: "team resolver error",
			setupMocks: func(d sutDeps) {
				d.teamResolver.EXPECT().
					ResolveTeamRefToID(gomock.Any(), defaultTeamRef).
					Return("", errors.New("boom")).
					Times(1)
			},
			assertErr: func(t *testing.T, err error) { require.Error(t, err) },
		},
		{
			name: "channel resolver error",
			setupMocks: func(d sutDeps) {
				expectResolveTeam(t, d)
				d.channelResolver.EXPECT().
					ResolveChannelRefToID(gomock.Any(), defaultTeamID, defaultChannelRef).
					Return("", errors.New("boom")).
					Times(1)
			},
			assertErr: func(t *testing.T, err error) { require.Error(t, err) },
		},
		{
			name: "ops error wrapped",
			setupMocks: func(d sutDeps) {
				expectResolveTeamAndChannel(t, d)
				d.ops.EXPECT().
					AddMember(gomock.Any(), defaultTeamID, defaultChannelID, userRef, false).
					Return(nil, &snd.ErrAccessForbidden{Code: 403, OriginalMessage: "nope"}).
					Times(1)
			},
			assertErr: func(t *testing.T, err error) { testutil.RequireReqErrCode(t, err, 403) },
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc, ctx := newSUT(t, tc.setupMocks)
			_, err := svc.AddMember(ctx, defaultTeamRef, defaultChannelRef, userRef, false)
			tc.assertErr(t, err)
		})
	}
}

func TestService_UpdateMemberRoles_Errors(t *testing.T) {
	type testCase struct {
		name       string
		setupMocks func(d sutDeps)
		assertErr  func(t *testing.T, err error)
	}

	userRef := "user@x.com"

	testCases := []testCase{
		{
			name: "resolve team+channel fails at team resolver",
			setupMocks: func(d sutDeps) {
				d.teamResolver.EXPECT().
					ResolveTeamRefToID(gomock.Any(), defaultTeamRef).
					Return("", errors.New("boom")).
					Times(1)
			},
			assertErr: func(t *testing.T, err error) { require.Error(t, err) },
		},
		{
			name: "resolve team+channel fails at channel resolver",
			setupMocks: func(d sutDeps) {
				expectResolveTeam(t, d)
				d.channelResolver.EXPECT().
					ResolveChannelRefToID(gomock.Any(), defaultTeamID, defaultChannelRef).
					Return("", errors.New("boom")).
					Times(1)
			},
			assertErr: func(t *testing.T, err error) { require.Error(t, err) },
		},
		{
			name: "member resolver error is wrapped",
			setupMocks: func(d sutDeps) {
				expectResolveTeamAndChannel(t, d)
				d.channelResolver.EXPECT().
					ResolveChannelMemberRefToID(gomock.Any(), defaultTeamID, defaultChannelID, userRef).
					Return("", errors.New("boom")).
					Times(1)
			},
			assertErr: func(t *testing.T, err error) { require.Error(t, err) },
		},
		{
			name: "ops error wrapped",
			setupMocks: func(d sutDeps) {
				expectResolveTeamAndChannel(t, d)
				d.channelResolver.EXPECT().
					ResolveChannelMemberRefToID(gomock.Any(), defaultTeamID, defaultChannelID, userRef).
					Return("member-id", nil).
					Times(1)

				d.ops.EXPECT().
					UpdateMemberRoles(gomock.Any(), defaultTeamID, defaultChannelID, "member-id", true).
					Return(nil, &snd.ErrAccessForbidden{Code: 403, OriginalMessage: "nope"}).
					Times(1)
			},
			assertErr: func(t *testing.T, err error) { testutil.RequireReqErrCode(t, err, 403) },
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc, ctx := newSUT(t, tc.setupMocks)
			_, err := svc.UpdateMemberRoles(ctx, defaultTeamRef, defaultChannelRef, userRef, true)
			tc.assertErr(t, err)
		})
	}
}

func TestService_RemoveMember_Errors(t *testing.T) {
	type testCase struct {
		name       string
		setupMocks func(d sutDeps)
		assertErr  func(t *testing.T, err error)
	}

	userRef := "user@x.com"

	testCases := []testCase{
		{
			name: "resolve team+channel fails at team resolver",
			setupMocks: func(d sutDeps) {
				d.teamResolver.EXPECT().
					ResolveTeamRefToID(gomock.Any(), defaultTeamRef).
					Return("", errors.New("boom")).
					Times(1)
			},
			assertErr: func(t *testing.T, err error) { require.Error(t, err) },
		},
		{
			name: "resolve team+channel fails at channel resolver",
			setupMocks: func(d sutDeps) {
				expectResolveTeam(t, d)
				d.channelResolver.EXPECT().
					ResolveChannelRefToID(gomock.Any(), defaultTeamID, defaultChannelRef).
					Return("", errors.New("boom")).
					Times(1)
			},
			assertErr: func(t *testing.T, err error) { require.Error(t, err) },
		},
		{
			name: "member resolver error is wrapped",
			setupMocks: func(d sutDeps) {
				expectResolveTeamAndChannel(t, d)
				d.channelResolver.EXPECT().
					ResolveChannelMemberRefToID(gomock.Any(), defaultTeamID, defaultChannelID, userRef).
					Return("", errors.New("boom")).
					Times(1)
			},
			assertErr: func(t *testing.T, err error) { require.Error(t, err) },
		},
		{
			name: "ops error wrapped",
			setupMocks: func(d sutDeps) {
				expectResolveTeamAndChannel(t, d)
				d.channelResolver.EXPECT().
					ResolveChannelMemberRefToID(gomock.Any(), defaultTeamID, defaultChannelID, userRef).
					Return("member-id", nil).
					Times(1)

				d.ops.EXPECT().
					RemoveMember(gomock.Any(), defaultTeamID, defaultChannelID, "member-id", userRef).
					Return(&snd.ErrAccessForbidden{Code: 403, OriginalMessage: "nope"}).
					Times(1)
			},
			assertErr: func(t *testing.T, err error) { testutil.RequireReqErrCode(t, err, 403) },
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc, ctx := newSUT(t, tc.setupMocks)
			err := svc.RemoveMember(ctx, defaultTeamRef, defaultChannelRef, userRef)
			tc.assertErr(t, err)
		})
	}
}

func TestService_SearchMessages(t *testing.T) {
	opts := &search.SearchMessagesOptions{}
	def := search.DefaultSearchConfig()

	t.Run("when teamRef and channelRef are nil -> does not resolve and passes nil IDs; uses default config", func(t *testing.T) {
		want := &search.SearchResults{}

		svc, ctx := newSUT(t, func(d sutDeps) {
			d.teamResolver.EXPECT().ResolveTeamRefToID(gomock.Any(), gomock.Any()).Times(0)
			d.channelResolver.EXPECT().ResolveChannelRefToID(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

			d.ops.EXPECT().
				SearchChannelMessages(gomock.Any(), (*string)(nil), (*string)(nil), opts, gomock.Any()).
				DoAndReturn(func(_ context.Context, teamIDptr, channelIDptr *string, gotOpts *search.SearchMessagesOptions, cfg *search.SearchConfig) (*search.SearchResults, error) {
					require.Nil(t, teamIDptr)
					require.Nil(t, channelIDptr)
					require.Same(t, opts, gotOpts)

					require.NotNil(t, cfg)
					require.Equal(t, *def, *cfg)

					return want, nil
				}).
				Times(1)
		})

		got, err := svc.SearchMessages(ctx, nil, nil, opts, nil)
		require.NoError(t, err)
		require.Same(t, want, got)
	})

	t.Run("when only teamRef is provided -> still resolve team id; pass team id", func(t *testing.T) {
		teamRef := defaultTeamRef
		want := &search.SearchResults{}

		svc, ctx := newSUT(t, func(d sutDeps) {
			expectResolveTeam(t, d)
			d.channelResolver.EXPECT().ResolveChannelRefToID(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

			d.ops.EXPECT().
				SearchChannelMessages(gomock.Any(), gomock.Any(), (*string)(nil), opts, gomock.Any()).
				DoAndReturn(func(_ context.Context, teamIDptr, channelIDptr *string, gotOpts *search.SearchMessagesOptions, cfg *search.SearchConfig) (*search.SearchResults, error) {
					require.NotNil(t, teamIDptr)
					require.Nil(t, channelIDptr)
					require.Same(t, opts, gotOpts)
					require.Equal(t, defaultTeamID, *teamIDptr)

					require.NotNil(t, cfg)
					require.Equal(t, *def, *cfg)

					return want, nil
				}).
				Times(1)
		})

		got, err := svc.SearchMessages(ctx, &teamRef, nil, opts, nil)
		require.NoError(t, err)
		require.Same(t, want, got)
	})

	t.Run("when only channelRef is provided -> error", func(t *testing.T) {
		channelRef := defaultChannelRef

		svc, ctx := newSUT(t, func(d sutDeps) {})

		got, err := svc.SearchMessages(ctx, nil, &channelRef, opts, nil)
		require.Error(t, err)
		require.Nil(t, got)
	})

	t.Run("when both refs are provided -> resolves IDs and passes pointers; uses default config", func(t *testing.T) {
		teamRef := defaultTeamRef
		channelRef := defaultChannelRef
		want := &search.SearchResults{}

		svc, ctx := newSUT(t, func(d sutDeps) {
			expectResolveTeamAndChannel(t, d)

			d.ops.EXPECT().
				SearchChannelMessages(gomock.Any(), gomock.Any(), gomock.Any(), opts, gomock.Any()).
				DoAndReturn(func(_ context.Context, teamIDptr, channelIDptr *string, gotOpts *search.SearchMessagesOptions, cfg *search.SearchConfig) (*search.SearchResults, error) {
					require.NotNil(t, teamIDptr)
					require.NotNil(t, channelIDptr)
					require.Equal(t, defaultTeamID, *teamIDptr)
					require.Equal(t, defaultChannelID, *channelIDptr)
					require.Same(t, opts, gotOpts)

					require.NotNil(t, cfg)
					require.Equal(t, *def, *cfg)

					return want, nil
				}).
				Times(1)
		})

		got, err := svc.SearchMessages(ctx, &teamRef, &channelRef, opts, nil)
		require.NoError(t, err)
		require.Same(t, want, got)
	})

	t.Run("when custom searchConfig is provided -> passes the same pointer through", func(t *testing.T) {
		teamRef := defaultTeamRef
		channelRef := defaultChannelRef
		want := &search.SearchResults{}
		customCfg := search.DefaultSearchConfig()

		svc, ctx := newSUT(t, func(d sutDeps) {
			expectResolveTeamAndChannel(t, d)

			d.ops.EXPECT().
				SearchChannelMessages(gomock.Any(), gomock.Any(), gomock.Any(), opts, customCfg).
				Return(want, nil).
				Times(1)
		})

		got, err := svc.SearchMessages(ctx, &teamRef, &channelRef, opts, customCfg)
		require.NoError(t, err)
		require.Same(t, want, got)
	})

	t.Run("when resolver fails -> returns wrapped error and does not call ops", func(t *testing.T) {
		teamRef := defaultTeamRef
		channelRef := defaultChannelRef

		svc, ctx := newSUT(t, func(d sutDeps) {
			d.teamResolver.EXPECT().
				ResolveTeamRefToID(gomock.Any(), defaultTeamRef).
				Return("", errors.New("boom")).
				Times(1)

			d.ops.EXPECT().
				SearchChannelMessages(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Times(0)
		})

		_, err := svc.SearchMessages(ctx, &teamRef, &channelRef, opts, nil)
		require.Error(t, err)
	})

	t.Run("ops error wrapped; with refs provided -> request error code preserved", func(t *testing.T) {
		teamRef := defaultTeamRef
		channelRef := defaultChannelRef

		svc, ctx := newSUT(t, func(d sutDeps) {
			expectResolveTeamAndChannel(t, d)

			d.ops.EXPECT().
				SearchChannelMessages(gomock.Any(), gomock.Any(), gomock.Any(), opts, gomock.Any()).
				Return(nil, &snd.ErrAccessForbidden{Code: 403, OriginalMessage: "nope"}).
				Times(1)
		})

		_, err := svc.SearchMessages(ctx, &teamRef, &channelRef, opts, nil)
		testutil.RequireReqErrCode(t, err, 403)
	})

	t.Run("ops error wrapped; with nil refs -> request error code preserved", func(t *testing.T) {
		svc, ctx := newSUT(t, func(d sutDeps) {
			d.ops.EXPECT().
				SearchChannelMessages(gomock.Any(), (*string)(nil), (*string)(nil), opts, gomock.Any()).
				Return(nil, &snd.ErrAccessForbidden{Code: 403, OriginalMessage: "nope"}).
				Times(1)
		})

		_, err := svc.SearchMessages(ctx, nil, nil, opts, nil)
		testutil.RequireReqErrCode(t, err, 403)
	})
}
