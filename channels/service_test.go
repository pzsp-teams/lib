package channels

import (
	"context"
	"errors"
	"testing"

	gomock "github.com/golang/mock/gomock"
	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	snd "github.com/pzsp-teams/lib/internal/sender"
	"github.com/pzsp-teams/lib/internal/testutil"
	"github.com/pzsp-teams/lib/internal/util"
	"github.com/pzsp-teams/lib/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type sutDeps struct {
	ops             *testutil.MockchannelOps
	teamResolver    *testutil.MockTeamResolver
	channelResolver *testutil.MockChannelResolver
	usersAPI        *testutil.MockUsersAPI
}

func newSUT(t *testing.T, setup func(d sutDeps)) (Service, context.Context) {
	t.Helper()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	opsMock := testutil.NewMockchannelOps(ctrl)
	trMock := testutil.NewMockTeamResolver(ctrl)
	crMock := testutil.NewMockChannelResolver(ctrl)
	usersMock := testutil.NewMockUsersAPI(ctrl)

	if setup != nil {
		setup(sutDeps{
			ops:             opsMock,
			teamResolver:    trMock,
			channelResolver: crMock,
			usersAPI:        usersMock,
		})
	}

	return NewService(opsMock, trMock, crMock, usersMock), context.Background()
}

func expectResolveTeamAndChannel(t *testing.T, d sutDeps) {
	t.Helper()

	const (
		teamRef    = "TeamA"
		teamID     = "team-id"
		channelRef = "ChanA"
		channelID  = "chan-id"
	)

	d.teamResolver.EXPECT().
		ResolveTeamRefToID(gomock.Any(), teamRef).
		Return(teamID, nil).
		Times(1)

	d.channelResolver.EXPECT().
		ResolveChannelRefToID(gomock.Any(), teamID, channelRef).
		Return(channelID, nil).
		Times(1)
}

func TestService_ListChannels(t *testing.T) {
	type testCase struct {
		name       string
		teamRef    string
		setupMocks func(d sutDeps)
		wantLen    int
		wantErrAs  any
	}

	testCases := []testCase{
		{
			name:    "success returns channels",
			teamRef: "TeamA",
			setupMocks: func(d sutDeps) {
				d.teamResolver.EXPECT().
					ResolveTeamRefToID(gomock.Any(), "TeamA").
					Return("team-id", nil).
					Times(1)

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
			wantErrAs: new(error),
		},
		{
			name:    "maps ops error",
			teamRef: "TeamA",
			setupMocks: func(d sutDeps) {
				d.teamResolver.EXPECT().
					ResolveTeamRefToID(gomock.Any(), "TeamA").
					Return("team-id", nil).
					Times(1)

				d.ops.EXPECT().
					ListChannelsByTeamID(gomock.Any(), "team-id").
					Return(nil, &snd.RequestError{Code: 403, Message: "nope"}).
					Times(1)
			},
			wantErrAs: new(snd.ErrAccessForbidden),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc, ctx := newSUT(t, tc.setupMocks)

			got, err := svc.ListChannels(ctx, tc.teamRef)

			if tc.wantErrAs != nil {
				require.Error(t, err)
				switch target := tc.wantErrAs.(type) {
				case *snd.ErrAccessForbidden:
					require.ErrorAs(t, err, target)
				case *error:
					// any error OK
				default:
					t.Fatalf("unsupported wantErrAs type: %T", tc.wantErrAs)
				}
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
		wantErrAs  any
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
			wantErrAs: new(error),
		},
		{
			name: "channel resolver error is propagated",
			setupMocks: func(d sutDeps) {
				d.teamResolver.EXPECT().
					ResolveTeamRefToID(gomock.Any(), "TeamA").
					Return("team-id", nil).
					Times(1)
				d.channelResolver.EXPECT().
					ResolveChannelRefToID(gomock.Any(), "team-id", "ChanA").
					Return("", errors.New("boom")).
					Times(1)
			},
			wantErrAs: new(error),
		},
		{
			name: "maps ops error",
			setupMocks: func(d sutDeps) {
				expectResolveTeamAndChannel(t, d)

				d.ops.EXPECT().
					GetChannelByID(gomock.Any(), "team-id", "chan-id").
					Return(nil, &snd.RequestError{Code: 404, Message: "missing"}).
					Times(1)
			},
			wantErrAs: new(snd.ErrResourceNotFound),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc, ctx := newSUT(t, tc.setupMocks)

			got, err := svc.Get(ctx, "TeamA", "ChanA")

			if tc.wantErrAs != nil {
				require.Error(t, err)
				switch target := tc.wantErrAs.(type) {
				case *snd.ErrResourceNotFound:
					require.ErrorAs(t, err, target)
				case *error:
				default:
					t.Fatalf("unsupported wantErrAs type: %T", tc.wantErrAs)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, tc.wantID, got.ID)
		})
	}
}

func TestService_CreateStandardChannel(t *testing.T) {
	svc, ctx := newSUT(t, func(d sutDeps) {
		d.teamResolver.EXPECT().
			ResolveTeamRefToID(gomock.Any(), "TeamA").
			Return("team-id", nil).
			Times(1)

		d.ops.EXPECT().
			CreateStandardChannel(gomock.Any(), "team-id", "NewChan").
			Return(&models.Channel{ID: "c1", Name: "NewChan"}, nil).
			Times(1)
	})

	got, err := svc.CreateStandardChannel(ctx, "TeamA", "NewChan")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "c1", got.ID)
	assert.Equal(t, "NewChan", got.Name)
}

func TestService_CreatePrivateChannel(t *testing.T) {
	svc, ctx := newSUT(t, func(d sutDeps) {
		d.teamResolver.EXPECT().
			ResolveTeamRefToID(gomock.Any(), "TeamA").
			Return("team-id", nil).
			Times(1)

		d.ops.EXPECT().
			CreatePrivateChannel(gomock.Any(), "team-id", "Secret", []string{"u1", "u2"}, []string{"o1"}).
			Return(&models.Channel{ID: "pc1", Name: "Secret"}, nil).
			Times(1)
	})

	got, err := svc.CreatePrivateChannel(ctx, "TeamA", "Secret", []string{"u1", "u2"}, []string{"o1"})
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "pc1", got.ID)
}

func TestService_Delete(t *testing.T) {
	type testCase struct {
		name      string
		setup     func(d sutDeps)
		wantErrAs any
	}

	testCases := []testCase{
		{
			name: "success passes channelRef to ops",
			setup: func(d sutDeps) {
				expectResolveTeamAndChannel(t, d)

				d.ops.EXPECT().
					DeleteChannel(gomock.Any(), "team-id", "chan-id", "ChanA").
					Return(nil).
					Times(1)
			},
		},
		{
			name: "maps ops error",
			setup: func(d sutDeps) {
				expectResolveTeamAndChannel(t, d)

				d.ops.EXPECT().
					DeleteChannel(gomock.Any(), "team-id", "chan-id", "ChanA").
					Return(&snd.RequestError{Code: 403, Message: "nope"}).
					Times(1)
			},
			wantErrAs: new(snd.ErrAccessForbidden),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc, ctx := newSUT(t, tc.setup)
			err := svc.Delete(ctx, "TeamA", "ChanA")

			if tc.wantErrAs == nil {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			switch target := tc.wantErrAs.(type) {
			case *snd.ErrAccessForbidden:
				require.ErrorAs(t, err, target)
			default:
				t.Fatalf("unsupported wantErrAs type: %T", tc.wantErrAs)
			}
		})
	}
}

func TestService_SendMessage(t *testing.T) {
	body := models.MessageBody{Content: "hi", ContentType: models.MessageContentTypeText}

	svc, ctx := newSUT(t, func(d sutDeps) {
		expectResolveTeamAndChannel(t, d)

		d.ops.EXPECT().
			SendMessage(gomock.Any(), "team-id", "chan-id", body).
			Return(&models.Message{ID: "m1", Content: "hi"}, nil).
			Times(1)
	})

	got, err := svc.SendMessage(ctx, "TeamA", "ChanA", body)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "m1", got.ID)
}

func TestService_SendReply(t *testing.T) {
	body := models.MessageBody{Content: "reply", ContentType: models.MessageContentTypeText}

	svc, ctx := newSUT(t, func(d sutDeps) {
		expectResolveTeamAndChannel(t, d)

		d.ops.EXPECT().
			SendReply(gomock.Any(), "team-id", "chan-id", "msg-1", body).
			Return(&models.Message{ID: "r1", Content: "reply"}, nil).
			Times(1)
	})

	got, err := svc.SendReply(ctx, "TeamA", "ChanA", "msg-1", body)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "r1", got.ID)
}

func TestService_ListMessages(t *testing.T) {
	top := int32(5)
	opts := &models.ListMessagesOptions{Top: &top}

	svc, ctx := newSUT(t, func(d sutDeps) {
		expectResolveTeamAndChannel(t, d)

		d.ops.EXPECT().
			ListMessages(gomock.Any(), "team-id", "chan-id", opts).
			Return([]*models.Message{{ID: "m1"}, {ID: "m2"}}, nil).
			Times(1)
	})

	got, err := svc.ListMessages(ctx, "TeamA", "ChanA", opts)
	require.NoError(t, err)
	require.Len(t, got, 2)
}

func TestService_ListReplies_PassesTopInOptions(t *testing.T) {
	top := int32(10)

	svc, ctx := newSUT(t, func(d sutDeps) {
		expectResolveTeamAndChannel(t, d)

		d.ops.EXPECT().
			ListReplies(gomock.Any(), "team-id", "chan-id", "msg-1", gomock.Any()).
			DoAndReturn(func(_ context.Context, _ string, _ string, _ string, opts *models.ListMessagesOptions) ([]*models.Message, *snd.RequestError) {
				require.NotNil(t, opts)
				require.NotNil(t, opts.Top)
				assert.Equal(t, top, *opts.Top)
				return []*models.Message{{ID: "r1"}}, nil
			}).
			Times(1)
	})

	got, err := svc.ListReplies(ctx, "TeamA", "ChanA", "msg-1", &top)
	require.NoError(t, err)
	require.Len(t, got, 1)
}

func TestService_GetMessage(t *testing.T) {
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
}

func TestService_GetReply(t *testing.T) {
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
}

func TestService_ListMembers(t *testing.T) {
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
}

func TestService_AddMember(t *testing.T) {
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
}

func TestService_UpdateMemberRoles_ResolvesMemberID(t *testing.T) {
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
}

func TestService_RemoveMember_ResolvesMemberIDAndPassesUserRef(t *testing.T) {
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
}

func TestService_GetMentions(t *testing.T) {
	type testCase struct {
		name       string
		raw        []string
		setupMocks func(d sutDeps)
		wantLen    int
		wantErr    bool
		assertFn   func(t *testing.T, got []models.Mention)
	}

	testCases := []testCase{
		{
			name: "resolves user, team, channel and skips blanks",
			raw:  []string{"  ", "alice@example.com", "team", "channel", "alice@example.com"},
			setupMocks: func(d sutDeps) {
				expectResolveTeamAndChannel(t, d)

				u := msmodels.NewUser()
				u.SetId(util.Ptr("u-1"))
				u.SetDisplayName(util.Ptr("Alice"))

				d.usersAPI.EXPECT().
					GetUserByEmailOrUPN(gomock.Any(), "alice@example.com").
					Return(u, nil).
					Times(2)
			},
			wantLen: 4,
			assertFn: func(t *testing.T, got []models.Mention) {
				require.Len(t, got, 4)

				// user mention
				assert.Equal(t, models.MentionUser, got[0].Kind)
				assert.Equal(t, "u-1", got[0].TargetID)
				assert.Equal(t, "Alice", got[0].Text)
				assert.Equal(t, 0, int(got[0].AtID))

				// team mention (text uses teamRef, target uses teamID)
				assert.Equal(t, models.MentionTeam, got[1].Kind)
				assert.Equal(t, "team-id", got[1].TargetID)
				assert.NotEmpty(t, got[1].Text)
				assert.Equal(t, 1, int(got[1].AtID))

				// channel mention (text uses channelRef, target uses channelID)
				assert.Equal(t, models.MentionChannel, got[2].Kind)
				assert.Equal(t, "chan-id", got[2].TargetID)
				assert.NotEmpty(t, got[2].Text)
				assert.Equal(t, 2, int(got[2].AtID))

				// second user mention
				assert.Equal(t, models.MentionUser, got[3].Kind)
				assert.Equal(t, "u-1", got[3].TargetID)
				assert.Equal(t, "Alice", got[3].Text)
				assert.Equal(t, 3, int(got[3].AtID))
			},
		},
		{
			name: "unknown ref returns error",
			raw:  []string{"something-else"},
			setupMocks: func(d sutDeps) {
				expectResolveTeamAndChannel(t, d)
			},
			wantErr: true,
		},
		{
			name: "user api error is propagated",
			raw:  []string{"alice@example.com"},
			setupMocks: func(d sutDeps) {
				expectResolveTeamAndChannel(t, d)

				d.usersAPI.EXPECT().
					GetUserByEmailOrUPN(gomock.Any(), "alice@example.com").
					Return(nil, &snd.RequestError{Code: 500, Message: "boom"}).
					Times(1)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc, ctx := newSUT(t, tc.setupMocks)

			got, err := svc.GetMentions(ctx, "TeamA", "ChanA", tc.raw)

			if tc.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Len(t, got, tc.wantLen)
			if tc.assertFn != nil {
				tc.assertFn(t, got)
			}
		})
	}
}
