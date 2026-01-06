package teams

import (
	"context"
	"errors"
	"reflect"
	"testing"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/internal/resources"
	snd "github.com/pzsp-teams/lib/internal/sender"
	"github.com/pzsp-teams/lib/internal/testutil"
	"github.com/pzsp-teams/lib/internal/util"
	"github.com/pzsp-teams/lib/models"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type opsSUTDeps struct {
	teamAPI *testutil.MockTeamAPI
}

func newOpsSUT(t *testing.T, setup func(ctx context.Context, d *opsSUTDeps)) (teamsOps, context.Context) {
	t.Helper()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	ctx := context.Background()
	d := &opsSUTDeps{
		teamAPI: testutil.NewMockTeamAPI(ctrl),
	}
	if setup != nil {
		setup(ctx, d)
	}

	return NewOps(d.teamAPI), ctx
}

func requireStatus(t *testing.T, err error, want int) {
	t.Helper()
	got, ok := snd.StatusCode(err)
	if ok {
		require.Equal(t, want, got)
		return
	}
	var re *snd.RequestError
	require.ErrorAs(t, err, &re, "expected snd.StatusCode() or *snd.RequestError for: %T (%v)", err, err)
	require.Equal(t, want, re.Code)
}

func extractResourceRefs(err error) (map[resources.Resource][]string, bool) {
	for e := err; e != nil; e = errors.Unwrap(e) {
		v := reflect.ValueOf(e)
		if v.Kind() == reflect.Pointer {
			if v.IsNil() {
				continue
			}
			v = v.Elem()
		}
		if v.Kind() != reflect.Struct {
			continue
		}

		f := v.FieldByName("ResourceRefs")
		if !f.IsValid() || f.Kind() != reflect.Map || !f.CanInterface() {
			continue
		}

		if m, ok := f.Interface().(map[resources.Resource][]string); ok {
			return m, true
		}
	}
	return nil, false
}

func requireErrDataHas(t *testing.T, err error, res resources.Resource, want string) {
	t.Helper()

	refs, ok := extractResourceRefs(err)
	require.True(t, ok, "could not extract ResourceRefs from error type %T (%v)", err, err)

	require.Contains(t, refs[res], want, "expected ResourceRefs[%v] to contain %q; got: %#v", res, want, refs[res])
}

func TestOps_GetTeamByID(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name   string
		teamID string
		setup  func(ctx context.Context, d *opsSUTDeps)
		check  func(t *testing.T, got *models.Team, err error)
	}

	testCases := []testCase{
		{
			name:   "error is mapped (status + resource)",
			teamID: "team-1",
			setup: func(ctx context.Context, d *opsSUTDeps) {
				d.teamAPI.EXPECT().
					Get(ctx, "team-1").
					Return(nil, &snd.RequestError{Code: 500, Message: "boom"})
			},
			check: func(t *testing.T, got *models.Team, err error) {
				require.Nil(t, got)
				require.Error(t, err)
				requireStatus(t, err, 500)
			},
		},
		{
			name:   "maps team",
			teamID: "team-2",
			setup: func(ctx context.Context, d *opsSUTDeps) {
				d.teamAPI.EXPECT().
					Get(ctx, "team-2").
					Return(testutil.NewGraphTeam(&testutil.NewTeamParams{
						ID:          util.Ptr("id-2"),
						DisplayName: util.Ptr("Team Two"),
					}), nil)
			},
			check: func(t *testing.T, got *models.Team, err error) {
				require.NoError(t, err)
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
			tc.check(t, out, err)
		})
	}
}

func TestOps_ListMyJoinedTeams(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name    string
		setup   func(ctx context.Context, d *opsSUTDeps)
		wantErr int
		wantIDs []string
	}

	testCases := []testCase{
		{
			name: "error is mapped",
			setup: func(ctx context.Context, d *opsSUTDeps) {
				d.teamAPI.EXPECT().
					ListMyJoined(ctx).
					Return(nil, &snd.RequestError{Code: 401, Message: "nope"})
			},
			wantErr: 401,
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

			if tc.wantErr != 0 {
				require.Error(t, err)
				requireStatus(t, err, tc.wantErr)
				require.Nil(t, out)
				return
			}

			require.NoError(t, err)
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
		check      func(t *testing.T, got *models.Team, err error)
	}

	testCases := []testCase{
		{
			name:       "create error is mapped (team=displayName)",
			display:    "X",
			mailNick:   "x",
			visibility: "private",
			setup: func(ctx context.Context, d *opsSUTDeps) {
				d.teamAPI.EXPECT().
					CreateViaGroup(ctx, "X", "x", "private").
					Return("", &snd.RequestError{Code: 400, Message: "bad"})
			},
			check: func(t *testing.T, got *models.Team, err error) {
				require.Nil(t, got)
				require.Error(t, err)
				requireStatus(t, err, 400)
			},
		},
		{
			name:       "get after create error is NOT mapped (as implemented)",
			display:    "Y",
			mailNick:   "y",
			visibility: "public",
			setup: func(ctx context.Context, d *opsSUTDeps) {
				d.teamAPI.EXPECT().CreateViaGroup(ctx, "Y", "y", "public").Return("new-id", nil)
				d.teamAPI.EXPECT().Get(ctx, "new-id").Return(nil, &snd.RequestError{Code: 503, Message: "down"})
			},
			check: func(t *testing.T, got *models.Team, err error) {
				require.Nil(t, got)
				require.Error(t, err)
				var re *snd.RequestError
				require.ErrorAs(t, err, &re)
				require.Equal(t, 503, re.Code)
			},
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
			check: func(t *testing.T, got *models.Team, err error) {
				require.NoError(t, err)
				require.NotNil(t, got)
				require.Equal(t, "tid", got.ID)
				require.Equal(t, "Z", got.DisplayName)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sut, ctx := newOpsSUT(t, tc.setup)

			out, err := sut.CreateViaGroup(ctx, tc.display, tc.mailNick, tc.visibility)
			tc.check(t, out, err)
		})
	}
}

func TestOps_CreateFromTemplate(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name      string
		setup     func(ctx context.Context, d *opsSUTDeps)
		wantID    string
		wantErr   int
		wantUsers []string
	}

	testCases := []testCase{
		{
			name: "success",
			setup: func(ctx context.Context, d *opsSUTDeps) {
				d.teamAPI.EXPECT().
					CreateFromTemplate(ctx, "N", "D", []string{"o1"}, nil, "").
					Return("id-1", nil)
			},
			wantID: "id-1",
		},
		{
			name: "error returns id and is mapped with users",
			setup: func(ctx context.Context, d *opsSUTDeps) {
				d.teamAPI.EXPECT().
					CreateFromTemplate(ctx, "N", "D", []string{"o1"}, nil, "").
					Return("id-partial", &snd.RequestError{Code: 409, Message: "conflict"})
			},
			wantID:    "id-partial",
			wantErr:   409,
			wantUsers: []string{"o1"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sut, ctx := newOpsSUT(t, tc.setup)

			id, err := sut.CreateFromTemplate(ctx, "N", "D", []string{"o1"}, nil, "")

			require.Equal(t, tc.wantID, id)
			if tc.wantErr != 0 {
				require.Error(t, err)
				requireStatus(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
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
	require.NoError(t, sut.Archive(ctx, "team-1", "ignored-ref", &readOnly))

	err := sut.Unarchive(ctx, "team-2")
	require.Error(t, err)
	requireStatus(t, err, 500)

	require.NoError(t, sut.DeleteTeam(ctx, "team-3", "ignored-ref"))
	require.NoError(t, sut.RemoveMember(ctx, "team-4", "mem-1", "ignored-user-ref"))
}

func TestOps_RestoreDeletedTeam(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name          string
		deletedGroup  string
		setup         func(ctx context.Context, d *opsSUTDeps)
		wantID        string
		wantStatus    int
		wantPlainErr  bool
		wantErrSubstr string
	}

	testCases := []testCase{
		{
			name:         "error is mapped with team=deletedGroupID",
			deletedGroup: "dg1",
			setup: func(ctx context.Context, d *opsSUTDeps) {
				d.teamAPI.EXPECT().
					RestoreDeleted(ctx, "dg1").
					Return(nil, &snd.RequestError{Code: 404, Message: "nope"})
			},
			wantStatus: 404,
		},
		{
			name:         "nil object => empty id, no error (as implemented)",
			deletedGroup: "dg2",
			setup: func(ctx context.Context, d *opsSUTDeps) {
				d.teamAPI.EXPECT().RestoreDeleted(ctx, "dg2").Return(nil, nil)
			},
			wantID: "",
		},
		{
			name:         "object with id",
			deletedGroup: "dg3",
			setup: func(ctx context.Context, d *opsSUTDeps) {
				obj := msmodels.NewDirectoryObject()
				obj.SetId(util.Ptr("restored-id"))
				d.teamAPI.EXPECT().RestoreDeleted(ctx, "dg3").Return(obj, nil)
			},
			wantID: "restored-id",
		},
		{
			name:         "object with empty id => plain error",
			deletedGroup: "dg4",
			setup: func(ctx context.Context, d *opsSUTDeps) {
				obj := msmodels.NewDirectoryObject()
				obj.SetId(util.Ptr(""))
				d.teamAPI.EXPECT().RestoreDeleted(ctx, "dg4").Return(obj, nil)
			},
			wantPlainErr:  true,
			wantErrSubstr: "restored object has empty id",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sut, ctx := newOpsSUT(t, tc.setup)

			id, err := sut.RestoreDeletedTeam(ctx, tc.deletedGroup)

			if tc.wantStatus != 0 {
				require.Error(t, err)
				require.Equal(t, "", id)
				requireStatus(t, err, tc.wantStatus)
				requireErrDataHas(t, err, resources.Team, tc.deletedGroup)
				return
			}

			if tc.wantPlainErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErrSubstr)
				require.Equal(t, "", id)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.wantID, id)
		})
	}
}

func TestOps_ListMembers(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name      string
		setup     func(ctx context.Context, d *opsSUTDeps)
		wantErr   int
		wantTeam  string
		wantIDs   []string
		checkRefs bool
	}

	testCases := []testCase{
		{
			name:     "error is mapped (team resource)",
			wantErr:  403,
			wantTeam: "team-1",
			setup: func(ctx context.Context, d *opsSUTDeps) {
				d.teamAPI.EXPECT().
					ListMembers(ctx, "team-1").
					Return(nil, &snd.RequestError{Code: 403, Message: "no"})
			},
			checkRefs: true,
		},
		{
			name: "maps members",
			setup: func(ctx context.Context, d *opsSUTDeps) {
				resp := msmodels.NewConversationMemberCollectionResponse()
				resp.SetValue([]msmodels.ConversationMemberable{
					testutil.NewGraphMember(&testutil.NewMemberParams{
						ID:    util.Ptr("m1"),
						Email: util.Ptr("a@b.com"),
					}),
					testutil.NewGraphMember(&testutil.NewMemberParams{
						ID:    util.Ptr("m2"),
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

			if tc.wantErr != 0 {
				require.Error(t, err)
				requireStatus(t, err, tc.wantErr)
				require.Nil(t, out)
				if tc.checkRefs {
					requireErrDataHas(t, err, resources.Team, tc.wantTeam)
				}
				return
			}

			require.NoError(t, err)
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
		name     string
		setup    func(ctx context.Context, d *opsSUTDeps)
		wantErr  int
		wantTeam string
		wantUser string
		wantID   string
	}

	testCases := []testCase{
		{
			name:     "error is mapped (team+user resources)",
			wantErr:  404,
			wantTeam: "t1",
			wantUser: "m1",
			setup: func(ctx context.Context, d *opsSUTDeps) {
				d.teamAPI.EXPECT().
					GetMember(ctx, "t1", "m1").
					Return(nil, &snd.RequestError{Code: 404, Message: "nope"})
			},
		},
		{
			name: "maps member",
			setup: func(ctx context.Context, d *opsSUTDeps) {
				d.teamAPI.EXPECT().
					GetMember(ctx, "t1", "m1").
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

			if tc.wantErr != 0 {
				require.Error(t, err)
				requireStatus(t, err, tc.wantErr)
				require.Nil(t, out)
				requireErrDataHas(t, err, resources.Team, tc.wantTeam)
				requireErrDataHas(t, err, resources.User, tc.wantUser)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, out)
			require.Equal(t, tc.wantID, out.ID)
		})
	}
}

func TestOps_AddMember(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name     string
		isOwner  bool
		setup    func(ctx context.Context, d *opsSUTDeps)
		wantErr  int
		wantTeam string
		wantUser string
		wantID   string
	}

	testCases := []testCase{
		{
			name:     "error is mapped (team+user resources)",
			isOwner:  false,
			wantErr:  400,
			wantTeam: "t1",
			wantUser: "u1",
			setup: func(ctx context.Context, d *opsSUTDeps) {
				roles := util.MemberRole(false)
				d.teamAPI.EXPECT().
					AddMember(ctx, "t1", "u1", roles).
					Return(nil, &snd.RequestError{Code: 400, Message: "bad"})
			},
		},
		{
			name:    "maps member",
			isOwner: true,
			setup: func(ctx context.Context, d *opsSUTDeps) {
				roles := util.MemberRole(true)
				d.teamAPI.EXPECT().
					AddMember(ctx, "t1", "u1", roles).
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

			if tc.wantErr != 0 {
				require.Error(t, err)
				requireStatus(t, err, tc.wantErr)
				require.Nil(t, out)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, out)
			require.Equal(t, tc.wantID, out.ID)
		})
	}
}

func TestOps_UpdateMemberRoles(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name     string
		isOwner  bool
		setup    func(ctx context.Context, d *opsSUTDeps)
		wantErr  int
		wantTeam string
		wantUser string
		wantID   string
	}

	testCases := []testCase{
		{
			name:     "error is mapped (team+user resources)",
			isOwner:  true,
			wantErr:  409,
			wantTeam: "t1",
			wantUser: "m1",
			setup: func(ctx context.Context, d *opsSUTDeps) {
				roles := util.MemberRole(true)
				d.teamAPI.EXPECT().
					UpdateMemberRoles(ctx, "t1", "m1", roles).
					Return(nil, &snd.RequestError{Code: 409, Message: "conflict"})
			},
		},
		{
			name:    "maps updated member",
			isOwner: false,
			setup: func(ctx context.Context, d *opsSUTDeps) {
				roles := util.MemberRole(false)
				d.teamAPI.EXPECT().
					UpdateMemberRoles(ctx, "t1", "m1", roles).
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

			if tc.wantErr != 0 {
				require.Error(t, err)
				requireStatus(t, err, tc.wantErr)
				require.Nil(t, out)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, out)
			require.Equal(t, tc.wantID, out.ID)
		})
	}
}
