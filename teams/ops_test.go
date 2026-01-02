package teams

import (
	"context"
	"testing"

	gomock "github.com/golang/mock/gomock"
	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	snd "github.com/pzsp-teams/lib/internal/sender"
	"github.com/pzsp-teams/lib/internal/testutil"
	"github.com/pzsp-teams/lib/internal/util"
	models "github.com/pzsp-teams/lib/models"
	"github.com/stretchr/testify/require"
)


type opsSUTDeps struct {
	teamAPI *testutil.MockTeamAPI
}

func newOpsSUT(t *testing.T, setup func(ctx context.Context, d *opsSUTDeps)) (teamsOps, context.Context) {
	t.Helper()

	ctrl := gomock.NewController(t)
	ctx := context.Background()

	d := &opsSUTDeps{
		teamAPI: testutil.NewMockTeamAPI(ctrl),
	}

	if setup != nil {
		setup(ctx, d)
	}

	return NewOps(d.teamAPI), ctx
}

func TestOps_Wait_DoesNothing(t *testing.T) {
	t.Parallel()

	sut, _ := newOpsSUT(t, nil)
	require.NotPanics(t, func() { sut.Wait() })
}

func TestOps_GetTeamByID(t *testing.T) {
	t.Parallel()

		type testCase struct {
			name    string
			teamID  string
			setup   func(ctx context.Context, d *opsSUTDeps)
			wantErr *snd.RequestError
			assert  func(t *testing.T, got *models.Team, err *snd.RequestError)
		}

	testCases := []testCase{
		{
			name:   "error propagated",
			teamID: "team-1",
			setup: func(ctx context.Context, d *opsSUTDeps) {
				d.teamAPI.EXPECT().Get(ctx, "team-1").Return(nil, &snd.RequestError{Code: 500, Message: "boom"})
			},
			wantErr: &snd.RequestError{Code: 500, Message: "boom"},
			assert: func(t *testing.T, got *models.Team, err *snd.RequestError) {
				require.Nil(t, got)
				require.NotNil(t, err)
				require.Equal(t, 500, err.Code)
			},
		},
		{
			name:   "maps team",
			teamID: "team-2",
			setup: func(ctx context.Context, d *opsSUTDeps) {
				d.teamAPI.EXPECT().Get(ctx, "team-2").
					Return(testutil.NewGraphTeam(&testutil.NewTeamParams{
						ID:          util.Ptr("id-2"),
						DisplayName: util.Ptr("Team Two"),
					}), nil)
			},
				assert: func(t *testing.T, got *models.Team, err *snd.RequestError) {
					require.Nil(t, err)
					require.NotNil(t, got)
					require.Equal(t, "id-2", got.ID)
					require.Equal(t, "Team Two", got.DisplayName)
				},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sut, ctx := newOpsSUT(t, tc.setup)

			out, err := sut.GetTeamByID(ctx, tc.teamID)

			if tc.wantErr != nil {
				require.NotNil(t, err)
				require.Equal(t, tc.wantErr.Code, err.Code)
				return
			}
			if tc.assert != nil {
				tc.assert(t, out, err)
			}
		})
	}
}

func TestOps_ListMyJoinedTeams(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name    string
		setup   func(ctx context.Context, d *opsSUTDeps)
		wantErr *snd.RequestError
		wantIDs []string
	}

	testCases := []testCase{
		{
			name: "error propagated",
			setup: func(ctx context.Context, d *opsSUTDeps) {
				d.teamAPI.EXPECT().ListMyJoined(ctx).Return(nil, &snd.RequestError{Code: 401, Message: "nope"})
			},
			wantErr: &snd.RequestError{Code: 401, Message: "nope"},
		},
		{
			name: "maps list",
			setup: func(ctx context.Context, d *opsSUTDeps) {
				resp := msmodels.NewTeamCollectionResponse()
				resp.SetValue([]msmodels.Teamable{
					testutil.NewGraphTeam(&testutil.NewTeamParams{
						ID:          util.Ptr("t1"),
						DisplayName: util.Ptr("A"),
					}),
					testutil.NewGraphTeam(&testutil.NewTeamParams{
						ID:          util.Ptr("t2"),
						DisplayName: util.Ptr("B"),
					}),
				})
				d.teamAPI.EXPECT().ListMyJoined(ctx).Return(resp, nil)
			},
			wantIDs: []string{"t1", "t2"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sut, ctx := newOpsSUT(t, tc.setup)

			out, err := sut.ListMyJoinedTeams(ctx)

			if tc.wantErr != nil {
				require.NotNil(t, err)
				require.Equal(t, tc.wantErr.Code, err.Code)
				require.Nil(t, out)
				return
			}

			require.Nil(t, err)
			require.Len(t, out, len(tc.wantIDs))
			for i, id := range tc.wantIDs {
				require.NotNil(t, out[i])
				require.Equal(t, id, out[i].ID)
			}
		})
	}
}

func TestOps_CreateViaGroup(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name       string
		display    string
		mailNick   string
		visibility string
		setup      func(ctx context.Context, d *opsSUTDeps)
		wantErr    *snd.RequestError
		wantTeamID string
	}

	testCases := []testCase{
		{
			name:       "create error propagated",
			display:    "X",
			mailNick:   "x",
			visibility: "private",
			setup: func(ctx context.Context, d *opsSUTDeps) {
				d.teamAPI.EXPECT().CreateViaGroup(ctx, "X", "x", "private").
					Return("", &snd.RequestError{Code: 400, Message: "bad"})
			},
			wantErr: &snd.RequestError{Code: 400, Message: "bad"},
		},
		{
			name:       "get after create error propagated",
			display:    "Y",
			mailNick:   "y",
			visibility: "public",
			setup: func(ctx context.Context, d *opsSUTDeps) {
				d.teamAPI.EXPECT().CreateViaGroup(ctx, "Y", "y", "public").Return("new-id", nil)
				d.teamAPI.EXPECT().Get(ctx, "new-id").Return(nil, &snd.RequestError{Code: 503, Message: "down"})
			},
			wantErr: &snd.RequestError{Code: 503, Message: "down"},
		},
		{
			name:       "returns mapped team",
			display:    "Z",
			mailNick:   "z",
			visibility: "public",
			setup: func(ctx context.Context, d *opsSUTDeps) {
				d.teamAPI.EXPECT().CreateViaGroup(ctx, "Z", "z", "public").Return("tid", nil)
				d.teamAPI.EXPECT().Get(ctx, "tid").Return(testutil.NewGraphTeam(&testutil.NewTeamParams{
					ID:          util.Ptr("tid"),
					DisplayName: util.Ptr("Z"),
				}), nil)
			},
			wantTeamID: "tid",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sut, ctx := newOpsSUT(t, tc.setup)

			out, err := sut.CreateViaGroup(ctx, tc.display, tc.mailNick, tc.visibility)

			if tc.wantErr != nil {
				require.NotNil(t, err)
				require.Equal(t, tc.wantErr.Code, err.Code)
				require.Nil(t, out)
				return
			}

			require.Nil(t, err)
			require.NotNil(t, out)
			require.Equal(t, tc.wantTeamID, out.ID)
		})
	}
}

func TestOps_CreateFromTemplate(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name    string
		setup   func(ctx context.Context, d *opsSUTDeps)
		wantID  string
		wantErr *snd.RequestError
	}

	testCases := []testCase{
		{
			name: "success",
			setup: func(ctx context.Context, d *opsSUTDeps) {
				d.teamAPI.EXPECT().CreateFromTemplate(ctx, "N", "D", []string{"o1"}).Return("id-1", nil)
			},
			wantID: "id-1",
		},
		{
			name: "error still returns id (as implemented)",
			setup: func(ctx context.Context, d *opsSUTDeps) {
				d.teamAPI.EXPECT().CreateFromTemplate(ctx, "N", "D", []string{"o1"}).
					Return("id-partial", &snd.RequestError{Code: 409, Message: "conflict"})
			},
			wantID:  "id-partial",
			wantErr: &snd.RequestError{Code: 409, Message: "conflict"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sut, ctx := newOpsSUT(t, tc.setup)

			id, err := sut.CreateFromTemplate(ctx, "N", "D", []string{"o1"})

			require.Equal(t, tc.wantID, id)
			if tc.wantErr != nil {
				require.NotNil(t, err)
				require.Equal(t, tc.wantErr.Code, err.Code)
				return
			}
			require.Nil(t, err)
		})
	}
}

func TestOps_Archive_Unarchive_Delete_RemoveMember_ForwardCalls(t *testing.T) {
	t.Parallel()

	sut, ctx := newOpsSUT(t, func(ctx context.Context, d *opsSUTDeps) {
		readOnly := true

		d.teamAPI.EXPECT().Archive(ctx, "team-1", &readOnly).Return(nil)
		d.teamAPI.EXPECT().Unarchive(ctx, "team-2").Return(&snd.RequestError{Code: 500, Message: "x"})
		d.teamAPI.EXPECT().Delete(ctx, "team-3").Return(nil)
		d.teamAPI.EXPECT().RemoveMember(ctx, "team-4", "mem-1").Return(nil)
	})

	readOnly := true
	require.Nil(t, sut.Archive(ctx, "team-1", "ignored-ref", &readOnly))
	require.NotNil(t, sut.Unarchive(ctx, "team-2"))
	require.Nil(t, sut.DeleteTeam(ctx, "team-3", "ignored-ref"))
	require.Nil(t, sut.RemoveMember(ctx, "team-4", "mem-1", "ignored-user-ref"))
}

func TestOps_RestoreDeletedTeam(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name    string
		setup   func(ctx context.Context, d *opsSUTDeps)
		wantID  string
		wantErr *snd.RequestError
	}

	testCases := []testCase{
		{
			name: "error propagated",
			setup: func(ctx context.Context, d *opsSUTDeps) {
				d.teamAPI.EXPECT().RestoreDeleted(ctx, "dg1").
					Return(nil, &snd.RequestError{Code: 404, Message: "nope"})
			},
			wantErr: &snd.RequestError{Code: 404, Message: "nope"},
		},
		{
			name: "nil object => empty id, no error (as implemented)",
			setup: func(ctx context.Context, d *opsSUTDeps) {
				d.teamAPI.EXPECT().RestoreDeleted(ctx, "dg2").Return(nil, nil)
			},
			wantID: "",
		},
		{
			name: "object with id",
			setup: func(ctx context.Context, d *opsSUTDeps) {
				obj := msmodels.NewDirectoryObject()
				obj.SetId(util.Ptr("restored-id"))
				d.teamAPI.EXPECT().RestoreDeleted(ctx, "dg3").Return(obj, nil)
			},
			wantID: "restored-id",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sut, ctx := newOpsSUT(t, tc.setup)

			id, err := sut.RestoreDeletedTeam(ctx, func() string {
				switch tc.name {
				case "error propagated":
					return "dg1"
				case "nil object => empty id, no error (as implemented)":
					return "dg2"
				default:
					return "dg3"
				}
			}())

			if tc.wantErr != nil {
				require.NotNil(t, err)
				require.Equal(t, tc.wantErr.Code, err.Code)
				require.Equal(t, "", id)
				return
			}

			require.Nil(t, err)
			require.Equal(t, tc.wantID, id)
		})
	}
}

func TestOps_ListMembers(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name    string
		setup   func(ctx context.Context, d *opsSUTDeps)
		wantErr *snd.RequestError
		wantIDs []string
	}

	testCases := []testCase{
		{
			name: "error propagated",
			setup: func(ctx context.Context, d *opsSUTDeps) {
				d.teamAPI.EXPECT().ListMembers(ctx, "team-1").
					Return(nil, &snd.RequestError{Code: 403, Message: "no"})
			},
			wantErr: &snd.RequestError{Code: 403, Message: "no"},
		},
		{
			name: "maps members",
			setup: func(ctx context.Context, d *opsSUTDeps) {
				resp := msmodels.NewConversationMemberCollectionResponse()
				resp.SetValue([]msmodels.ConversationMemberable{
					testutil.NewGraphMember(&testutil.NewMemberParams{
						ID:          util.Ptr("m1"),
						Email: util.Ptr("a@b.com"),
					}),
					testutil.NewGraphMember(&testutil.NewMemberParams{
						ID:          util.Ptr("m2"),
						Email: util.Ptr("c@d.com"),
					}),
				})
				d.teamAPI.EXPECT().ListMembers(ctx, "team-1").Return(resp, nil)
			},
			wantIDs: []string{"m1", "m2"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sut, ctx := newOpsSUT(t, tc.setup)

			out, err := sut.ListMembers(ctx, "team-1")

			if tc.wantErr != nil {
				require.NotNil(t, err)
				require.Equal(t, tc.wantErr.Code, err.Code)
				require.Nil(t, out)
				return
			}

			require.Nil(t, err)
			require.Len(t, out, len(tc.wantIDs))
			for i, id := range tc.wantIDs {
				require.NotNil(t, out[i])
				require.Equal(t, id, out[i].ID)
			}
		})
	}
}

func TestOps_GetMemberByID(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name    string
		setup   func(ctx context.Context, d *opsSUTDeps)
		wantErr *snd.RequestError
		wantID  string
	}

	testCases := []testCase{
		{
			name: "error propagated",
			setup: func(ctx context.Context, d *opsSUTDeps) {
				d.teamAPI.EXPECT().GetMember(ctx, "t1", "m1").
					Return(nil, &snd.RequestError{Code: 404, Message: "nope"})
			},
			wantErr: &snd.RequestError{Code: 404, Message: "nope"},
		},
		{
			name: "maps member",
			setup: func(ctx context.Context, d *opsSUTDeps) {
				d.teamAPI.EXPECT().GetMember(ctx, "t1", "m1").
					Return(testutil.NewGraphMember(&testutil.NewMemberParams{
						ID:    util.Ptr("m1"),
						Email: util.Ptr("x@y.com"),
					}), nil)
			},
			wantID: "m1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sut, ctx := newOpsSUT(t, tc.setup)

			out, err := sut.GetMemberByID(ctx, "t1", "m1")

			if tc.wantErr != nil {
				require.NotNil(t, err)
				require.Equal(t, tc.wantErr.Code, err.Code)
				require.Nil(t, out)
				return
			}

			require.Nil(t, err)
			require.NotNil(t, out)
			require.Equal(t, tc.wantID, out.ID)
		})
	}
}

func TestOps_AddMember(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name    string
		isOwner bool
		setup   func(ctx context.Context, d *opsSUTDeps)
		wantErr *snd.RequestError
		wantID  string
	}

	testCases := []testCase{
		{
			name:    "error propagated",
			isOwner: false,
			setup: func(ctx context.Context, d *opsSUTDeps) {
				roles := util.MemberRole(false)
				d.teamAPI.EXPECT().AddMember(ctx, "t1", "u1", roles).
					Return(nil, &snd.RequestError{Code: 400, Message: "bad"})
			},
			wantErr: &snd.RequestError{Code: 400, Message: "bad"},
		},
		{
			name:    "maps member",
			isOwner: true,
			setup: func(ctx context.Context, d *opsSUTDeps) {
				roles := util.MemberRole(true)
				d.teamAPI.EXPECT().AddMember(ctx, "t1", "u1", roles).
					Return(testutil.NewGraphMember(&testutil.NewMemberParams{
						ID:    util.Ptr("m1"),
						Email: util.Ptr("u1@x.com"),
					}), nil)
			},
			wantID: "m1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sut, ctx := newOpsSUT(t, tc.setup)

			out, err := sut.AddMember(ctx, "t1", "u1", tc.isOwner)

			if tc.wantErr != nil {
				require.NotNil(t, err)
				require.Equal(t, tc.wantErr.Code, err.Code)
				require.Nil(t, out)
				return
			}

			require.Nil(t, err)
			require.NotNil(t, out)
			require.Equal(t, tc.wantID, out.ID)
		})
	}
}

func TestOps_UpdateMemberRoles(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name    string
		isOwner bool
		setup   func(ctx context.Context, d *opsSUTDeps)
		wantErr *snd.RequestError
		wantID  string
	}

	testCases := []testCase{
		{
			name:    "error propagated",
			isOwner: true,
			setup: func(ctx context.Context, d *opsSUTDeps) {
				roles := util.MemberRole(true)
				d.teamAPI.EXPECT().UpdateMemberRoles(ctx, "t1", "m1", roles).
					Return(nil, &snd.RequestError{Code: 409, Message: "conflict"})
			},
			wantErr: &snd.RequestError{Code: 409, Message: "conflict"},
		},
		{
			name:    "maps updated member",
			isOwner: false,
			setup: func(ctx context.Context, d *opsSUTDeps) {
				roles := util.MemberRole(false)
				d.teamAPI.EXPECT().UpdateMemberRoles(ctx, "t1", "m1", roles).
					Return(testutil.NewGraphMember(&testutil.NewMemberParams{
						ID:    util.Ptr("m1"),
						Email: util.Ptr("x@y.com"),
					}), nil)
			},
			wantID: "m1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sut, ctx := newOpsSUT(t, tc.setup)

			out, err := sut.UpdateMemberRoles(ctx, "t1", "m1", tc.isOwner)

			if tc.wantErr != nil {
				require.NotNil(t, err)
				require.Equal(t, tc.wantErr.Code, err.Code)
				require.Nil(t, out)
				return
			}

			require.Nil(t, err)
			require.NotNil(t, out)
			require.Equal(t, tc.wantID, out.ID)
		})
	}
}
