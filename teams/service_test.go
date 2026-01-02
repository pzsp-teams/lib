package teams

import (
	"context"
	"errors"
	"testing"

	gomock "github.com/golang/mock/gomock"
	sender "github.com/pzsp-teams/lib/internal/sender"
	"github.com/pzsp-teams/lib/internal/testutil"
	"github.com/pzsp-teams/lib/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type sutDeps struct {
	ops      *testutil.MockteamsOps
	resolver *testutil.MockTeamResolver
}

func newSUT(t *testing.T, setup func(d sutDeps)) (Service, context.Context) {
	t.Helper()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	opsMock := testutil.NewMockteamsOps(ctrl)
	resolverMock := testutil.NewMockTeamResolver(ctrl)

	if setup != nil {
		setup(sutDeps{ops: opsMock, resolver: resolverMock})
	}

	return NewService(opsMock, resolverMock), context.Background()
}

func TestService_ListMyJoined(t *testing.T) {
	type testCase struct {
		name       string
		setupMocks func(d sutDeps)
		assertFn   func(t *testing.T, got []*models.Team, err error)
	}

	testCases := []testCase{
		{
			name: "maps teams",
			setupMocks: func(d sutDeps) {
				teams := []*models.Team{
					{ID: "1", DisplayName: "Alpha"},
					{ID: "2", DisplayName: "Beta"},
				}

				d.ops.EXPECT().
					ListMyJoinedTeams(gomock.Any()).
					Return(teams, nil).
					Times(1)
			},
			assertFn: func(t *testing.T, got []*models.Team, err error) {
				require.NoError(t, err)
				require.Len(t, got, 2)
				assert.Equal(t, "1", got[0].ID)
				assert.Equal(t, "Beta", got[1].DisplayName)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc, ctx := newSUT(t, tc.setupMocks)
			got, err := svc.ListMyJoined(ctx)
			tc.assertFn(t, got, err)
		})
	}

	t.Run("maps api error", func(t *testing.T) {
		svc, ctx := newSUT(t, func(d sutDeps) {
			d.ops.EXPECT().
				ListMyJoinedTeams(gomock.Any()).
				Return(nil, &sender.RequestError{Code: 403, Message: "nope"}).
				Times(1)
		})

		_, err := svc.ListMyJoined(ctx)
		require.Error(t, err)

		var forbidden sender.ErrAccessForbidden
		require.ErrorAs(t, err, &forbidden)
	})
}

func TestService_Get(t *testing.T) {
	type testCase struct {
		name       string
		teamRef    string
		setupMocks func(d sutDeps)
		wantID     string
		wantName   string
		wantErrAs  any
	}

	testCases := []testCase{
		{
			name:    "maps team and calls resolver",
			teamRef: "team-name-42",
			setupMocks: func(d sutDeps) {
				d.resolver.EXPECT().
					ResolveTeamRefToID(gomock.Any(), "team-name-42").
					Return("resolved-42", nil).
					Times(1)

				d.ops.EXPECT().
					GetTeamByID(gomock.Any(), "resolved-42").
					Return(&models.Team{ID: "42", DisplayName: "X"}, nil).
					Times(1)
			},
			wantID:   "42",
			wantName: "X",
		},
		{
			name:    "resolver error is propagated",
			teamRef: "team-x",
			setupMocks: func(d sutDeps) {
				d.resolver.EXPECT().
					ResolveTeamRefToID(gomock.Any(), "team-x").
					Return("", errors.New("boom")).
					Times(1)
			},
			wantErrAs: new(error),
		},
		{
			name:    "api error is mapped",
			teamRef: "missing-team",
			setupMocks: func(d sutDeps) {
				d.resolver.EXPECT().
					ResolveTeamRefToID(gomock.Any(), "missing-team").
					Return("missing-id", nil).
					Times(1)

				d.ops.EXPECT().
					GetTeamByID(gomock.Any(), "missing-id").
					Return(nil, &sender.RequestError{Code: 404, Message: "no such team"}).
					Times(1)
			},
			wantErrAs: new(sender.ErrResourceNotFound),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc, ctx := newSUT(t, tc.setupMocks)

			got, err := svc.Get(ctx, tc.teamRef)

			if tc.wantErrAs != nil {
				require.Error(t, err)
				switch target := tc.wantErrAs.(type) {
				case *sender.ErrResourceNotFound:
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
			assert.Equal(t, tc.wantName, got.DisplayName)
		})
	}
}

func TestService_Delete(t *testing.T) {
	type testCase struct {
		name       string
		teamRef    string
		setupMocks func(d sutDeps)
		wantErrAs  any
	}

	testCases := []testCase{
		{
			name:    "success",
			teamRef: "MyTeam",
			setupMocks: func(d sutDeps) {
				d.resolver.EXPECT().
					ResolveTeamRefToID(gomock.Any(), "MyTeam").
					Return("team-id", nil).
					Times(1)

				d.ops.EXPECT().
					DeleteTeam(gomock.Any(), "team-id", "MyTeam").
					Return(nil).
					Times(1)
			},
		},
		{
			name:    "maps forbidden",
			teamRef: "MyTeam",
			setupMocks: func(d sutDeps) {
				d.resolver.EXPECT().
					ResolveTeamRefToID(gomock.Any(), "MyTeam").
					Return("team-id", nil).
					Times(1)

				d.ops.EXPECT().
					DeleteTeam(gomock.Any(), "team-id", "MyTeam").
					Return(&sender.RequestError{Code: 403, Message: "nope"}).
					Times(1)
			},
			wantErrAs: new(sender.ErrAccessForbidden),
		},
		{
			name:    "resolver error is propagated",
			teamRef: "MyTeam",
			setupMocks: func(d sutDeps) {
				d.resolver.EXPECT().
					ResolveTeamRefToID(gomock.Any(), "MyTeam").
					Return("", errors.New("resolver boom")).
					Times(1)
			},
			wantErrAs: new(error),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc, ctx := newSUT(t, tc.setupMocks)

			err := svc.Delete(ctx, tc.teamRef)

			if tc.wantErrAs == nil {
				require.NoError(t, err)
				return
			}

			require.Error(t, err)
			switch target := tc.wantErrAs.(type) {
			case *sender.ErrAccessForbidden:
				require.ErrorAs(t, err, target)
			case *error:
			default:
				t.Fatalf("unsupported wantErrAs type: %T", tc.wantErrAs)
			}
		})
	}
}

func TestService_CreateViaGroup(t *testing.T) {
	type testCase struct {
		name       string
		setupMocks func(d sutDeps)
		wantErrAs  any
		wantID     string
	}

	testCases := []testCase{
		{
			name: "maps create error",
			setupMocks: func(d sutDeps) {
				d.ops.EXPECT().
					CreateViaGroup(gomock.Any(), "X", "x", "public").
					Return(nil, &sender.RequestError{Code: 403, Message: "nope"}).
					Times(1)
			},
			wantErrAs: new(sender.ErrAccessForbidden),
		},
		{
			name: "success maps team",
			setupMocks: func(d sutDeps) {
				d.ops.EXPECT().
					CreateViaGroup(gomock.Any(), "X", "x", "public").
					Return(&models.Team{ID: "team-xyz", DisplayName: "X"}, nil).
					Times(1)
			},
			wantID: "team-xyz",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc, ctx := newSUT(t, tc.setupMocks)

			got, err := svc.CreateViaGroup(ctx, "X", "x", "public")

			if tc.wantErrAs != nil {
				require.Error(t, err)
				switch target := tc.wantErrAs.(type) {
				case *sender.ErrAccessForbidden:
					require.ErrorAs(t, err, target)
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

func TestService_CreateFromTemplate(t *testing.T) {
	type testCase struct {
		name       string
		setupMocks func(d sutDeps)
		wantID     string
		wantErrAs  any
	}

	testCases := []testCase{
		{
			name: "returns id",
			setupMocks: func(d sutDeps) {
				d.ops.EXPECT().
					CreateFromTemplate(gomock.Any(), "Tpl", "Desc", gomock.Any()).
					Return("tmpl-123", nil).
					Times(1)
			},
			wantID: "tmpl-123",
		},
		{
			name: "maps error",
			setupMocks: func(d sutDeps) {
				d.ops.EXPECT().
					CreateFromTemplate(gomock.Any(), "Tpl", "Desc", gomock.Any()).
					Return("", &sender.RequestError{Code: 403, Message: "nope"}).
					Times(1)
			},
			wantErrAs: new(sender.ErrAccessForbidden),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc, ctx := newSUT(t, tc.setupMocks)

			got, err := svc.CreateFromTemplate(ctx, "Tpl", "Desc", nil)

			if tc.wantErrAs != nil {
				require.Error(t, err)
				switch target := tc.wantErrAs.(type) {
				case *sender.ErrAccessForbidden:
					require.ErrorAs(t, err, target)
				default:
					t.Fatalf("unsupported wantErrAs type: %T", tc.wantErrAs)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.wantID, got)
		})
	}
}

func TestService_Archive(t *testing.T) {
	type testCase struct {
		name       string
		setupMocks func(d sutDeps)
		wantErrAs  any
	}

	readOnly := false

	testCases := []testCase{
		{
			name: "success",
			setupMocks: func(d sutDeps) {
				d.resolver.EXPECT().
					ResolveTeamRefToID(gomock.Any(), "T1").
					Return("team-id", nil).
					Times(1)
				d.ops.EXPECT().
					Archive(gomock.Any(), "team-id", "T1", &readOnly).
					Return(nil).
					Times(1)
			},
		},
		{
			name: "maps forbidden",
			setupMocks: func(d sutDeps) {
				d.resolver.EXPECT().
					ResolveTeamRefToID(gomock.Any(), "T1").
					Return("team-id", nil).
					Times(1)
				d.ops.EXPECT().
					Archive(gomock.Any(), "team-id", "T1", &readOnly).
					Return(&sender.RequestError{Code: 403, Message: "nope"}).
					Times(1)
			},
			wantErrAs: new(sender.ErrAccessForbidden),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc, ctx := newSUT(t, tc.setupMocks)

			err := svc.Archive(ctx, "T1", &readOnly)

			if tc.wantErrAs == nil {
				require.NoError(t, err)
				return
			}

			require.Error(t, err)
			switch target := tc.wantErrAs.(type) {
			case *sender.ErrAccessForbidden:
				require.ErrorAs(t, err, target)
			default:
				t.Fatalf("unsupported wantErrAs type: %T", tc.wantErrAs)
			}
		})
	}
}

func TestService_Unarchive(t *testing.T) {
	type testCase struct {
		name       string
		setupMocks func(d sutDeps)
		wantErrAs  any
	}

	testCases := []testCase{
		{
			name: "success",
			setupMocks: func(d sutDeps) {
				d.resolver.EXPECT().
					ResolveTeamRefToID(gomock.Any(), "T1").
					Return("team-id", nil).
					Times(1)
				d.ops.EXPECT().
					Unarchive(gomock.Any(), "team-id").
					Return(nil).
					Times(1)
			},
		},
		{
			name: "maps forbidden",
			setupMocks: func(d sutDeps) {
				d.resolver.EXPECT().
					ResolveTeamRefToID(gomock.Any(), "T1").
					Return("team-id", nil).
					Times(1)
				d.ops.EXPECT().
					Unarchive(gomock.Any(), "team-id").
					Return(&sender.RequestError{Code: 403, Message: "nope"}).
					Times(1)
			},
			wantErrAs: new(sender.ErrAccessForbidden),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc, ctx := newSUT(t, tc.setupMocks)

			err := svc.Unarchive(ctx, "T1")

			if tc.wantErrAs == nil {
				require.NoError(t, err)
				return
			}

			require.Error(t, err)
			switch target := tc.wantErrAs.(type) {
			case *sender.ErrAccessForbidden:
				require.ErrorAs(t, err, target)
			default:
				t.Fatalf("unsupported wantErrAs type: %T", tc.wantErrAs)
			}
		})
	}
}

func TestService_RestoreDeleted(t *testing.T) {
	type testCase struct {
		name       string
		setupMocks func(d sutDeps)
		wantID     string
		wantErrAs  any
		wantErr    bool
	}

	testCases := []testCase{
		{
			name: "returns id",
			setupMocks: func(d sutDeps) {
				d.ops.EXPECT().
					RestoreDeletedTeam(gomock.Any(), "deleted-1").
					Return("restored-id", nil).
					Times(1)
			},
			wantID: "restored-id",
		},
		{
			name: "maps not found",
			setupMocks: func(d sutDeps) {
				d.ops.EXPECT().
					RestoreDeletedTeam(gomock.Any(), "deleted-1").
					Return("", &sender.RequestError{Code: 404, Message: "missing"}).
					Times(1)
			},
			wantErrAs: new(sender.ErrResourceNotFound),
		},
		{
			name: "empty id returns error",
			setupMocks: func(d sutDeps) {
				d.ops.EXPECT().
					RestoreDeletedTeam(gomock.Any(), "deleted-1").
					Return("", nil).
					Times(1)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc, ctx := newSUT(t, tc.setupMocks)

			got, err := svc.RestoreDeleted(ctx, "deleted-1")

			if tc.wantErrAs != nil {
				require.Error(t, err)
				switch target := tc.wantErrAs.(type) {
				case *sender.ErrResourceNotFound:
					require.ErrorAs(t, err, target)
				default:
					t.Fatalf("unsupported wantErrAs type: %T", tc.wantErrAs)
				}
				return
			}
			if tc.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.wantID, got)
		})
	}
}

func TestService_ListMembers(t *testing.T) {
	type testCase struct {
		name       string
		teamRef    string
		setupMocks func(d sutDeps)
		wantIDs    []string
		wantErrAs  any
	}

	testCases := []testCase{
		{
			name:    "success maps members",
			teamRef: "TeamX",
			setupMocks: func(d sutDeps) {
				d.resolver.EXPECT().
					ResolveTeamRefToID(gomock.Any(), "TeamX").
					Return("team-id", nil).
					Times(1)

				members := []*models.Member{
					{ID: "m1"},
					{ID: "m2"},
				}

				d.ops.EXPECT().
					ListMembers(gomock.Any(), "team-id").
					Return(members, nil).
					Times(1)
			},
			wantIDs: []string{"m1", "m2"},
		},
		{
			name:    "resolver error is propagated",
			teamRef: "TeamX",
			setupMocks: func(d sutDeps) {
				d.resolver.EXPECT().
					ResolveTeamRefToID(gomock.Any(), "TeamX").
					Return("", errors.New("boom")).
					Times(1)
			},
			wantErrAs: new(error),
		},
		{
			name:    "api error is mapped",
			teamRef: "TeamX",
			setupMocks: func(d sutDeps) {
				d.resolver.EXPECT().
					ResolveTeamRefToID(gomock.Any(), "TeamX").
					Return("team-id", nil).
					Times(1)

				d.ops.EXPECT().
					ListMembers(gomock.Any(), "team-id").
					Return(nil, &sender.RequestError{Code: 403, Message: "nope"}).
					Times(1)
			},
			wantErrAs: new(sender.ErrAccessForbidden),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc, ctx := newSUT(t, tc.setupMocks)

			got, err := svc.ListMembers(ctx, tc.teamRef)

			if tc.wantErrAs != nil {
				require.Error(t, err)
				switch target := tc.wantErrAs.(type) {
				case *sender.ErrAccessForbidden:
					require.ErrorAs(t, err, target)
				case *error:
				default:
					t.Fatalf("unsupported wantErrAs type: %T", tc.wantErrAs)
				}
				return
			}

			require.NoError(t, err)
			require.Len(t, got, len(tc.wantIDs))
			for i, id := range tc.wantIDs {
				assert.Equal(t, id, got[i].ID)
			}
		})
	}
}

func TestService_AddMember(t *testing.T) {
	type testCase struct {
		name       string
		teamRef    string
		userRef    string
		isOwner    bool
		setupMocks func(d sutDeps)
		wantID     string
		wantErrAs  any
	}

	testCases := []testCase{
		{
			name:    "success adds owner",
			teamRef: "TeamX",
			userRef: "user@x.com",
			isOwner: true,
			setupMocks: func(d sutDeps) {
				d.resolver.EXPECT().
					ResolveTeamRefToID(gomock.Any(), "TeamX").
					Return("team-id", nil).
					Times(1)

				d.ops.EXPECT().
					AddMember(gomock.Any(), "team-id", "user@x.com", true).
					Return(&models.Member{ID: "m1"}, nil).
					Times(1)
			},
			wantID: "m1",
		},
		{
			name:    "success adds member (no owner role)",
			teamRef: "TeamX",
			userRef: "user@x.com",
			isOwner: false,
			setupMocks: func(d sutDeps) {
				d.resolver.EXPECT().
					ResolveTeamRefToID(gomock.Any(), "TeamX").
					Return("team-id", nil).
					Times(1)

				d.ops.EXPECT().
					AddMember(gomock.Any(), "team-id", "user@x.com", false).
					Return(&models.Member{ID: "m2"}, nil).
					Times(1)
			},
			wantID: "m2",
		},
		{
			name:    "resolver error is propagated",
			teamRef: "TeamX",
			userRef: "user@x.com",
			isOwner: false,
			setupMocks: func(d sutDeps) {
				d.resolver.EXPECT().
					ResolveTeamRefToID(gomock.Any(), "TeamX").
					Return("", errors.New("boom")).
					Times(1)
			},
			wantErrAs: new(error),
		},
		{
			name:    "api error is mapped",
			teamRef: "TeamX",
			userRef: "user@x.com",
			isOwner: false,
			setupMocks: func(d sutDeps) {
				d.resolver.EXPECT().
					ResolveTeamRefToID(gomock.Any(), "TeamX").
					Return("team-id", nil).
					Times(1)

				d.ops.EXPECT().
					AddMember(gomock.Any(), "team-id", "user@x.com", false).
					Return(nil, &sender.RequestError{Code: 403, Message: "nope"}).
					Times(1)
			},
			wantErrAs: new(sender.ErrAccessForbidden),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc, ctx := newSUT(t, tc.setupMocks)

			got, err := svc.AddMember(ctx, tc.teamRef, tc.userRef, tc.isOwner)

			if tc.wantErrAs != nil {
				require.Error(t, err)
				switch target := tc.wantErrAs.(type) {
				case *sender.ErrAccessForbidden:
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

func TestService_GetMember(t *testing.T) {
	type testCase struct {
		name       string
		teamRef    string
		userRef    string
		setupMocks func(d sutDeps)
		wantID     string
		wantErrAs  any
	}

	testCases := []testCase{
		{
			name:    "success resolves memberID and gets member",
			teamRef: "TeamX",
			userRef: "user@x.com",
			setupMocks: func(d sutDeps) {
				d.resolver.EXPECT().
					ResolveTeamRefToID(gomock.Any(), "TeamX").
					Return("team-id", nil).
					Times(1)

				d.resolver.EXPECT().
					ResolveTeamMemberRefToID(gomock.Any(), "team-id", "user@x.com").
					Return("member-id", nil).
					Times(1)

				d.ops.EXPECT().
					GetMemberByID(gomock.Any(), "team-id", "member-id").
					Return(&models.Member{ID: "member-id"}, nil).
					Times(1)
			},
			wantID: "member-id",
		},
		{
			name:    "team resolver error is propagated",
			teamRef: "TeamX",
			userRef: "user@x.com",
			setupMocks: func(d sutDeps) {
				d.resolver.EXPECT().
					ResolveTeamRefToID(gomock.Any(), "TeamX").
					Return("", errors.New("boom")).
					Times(1)
			},
			wantErrAs: new(error),
		},
		{
			name:    "member resolver error is propagated",
			teamRef: "TeamX",
			userRef: "user@x.com",
			setupMocks: func(d sutDeps) {
				d.resolver.EXPECT().
					ResolveTeamRefToID(gomock.Any(), "TeamX").
					Return("team-id", nil).
					Times(1)

				d.resolver.EXPECT().
					ResolveTeamMemberRefToID(gomock.Any(), "team-id", "user@x.com").
					Return("", errors.New("resolve member boom")).
					Times(1)
			},
			wantErrAs: new(error),
		},
		{
			name:    "api error is mapped",
			teamRef: "TeamX",
			userRef: "user@x.com",
			setupMocks: func(d sutDeps) {
				d.resolver.EXPECT().
					ResolveTeamRefToID(gomock.Any(), "TeamX").
					Return("team-id", nil).
					Times(1)

				d.resolver.EXPECT().
					ResolveTeamMemberRefToID(gomock.Any(), "team-id", "user@x.com").
					Return("member-id", nil).
					Times(1)

				d.ops.EXPECT().
					GetMemberByID(gomock.Any(), "team-id", "member-id").
					Return(nil, &sender.RequestError{Code: 404, Message: "missing"}).
					Times(1)
			},
			wantErrAs: new(sender.ErrResourceNotFound),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc, ctx := newSUT(t, tc.setupMocks)

			got, err := svc.GetMember(ctx, tc.teamRef, tc.userRef)

			if tc.wantErrAs != nil {
				require.Error(t, err)
				switch target := tc.wantErrAs.(type) {
				case *sender.ErrResourceNotFound:
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

func TestService_RemoveMember(t *testing.T) {
	type testCase struct {
		name       string
		teamRef    string
		userRef    string
		setupMocks func(d sutDeps)
		wantErrAs  any
	}

	testCases := []testCase{
		{
			name:    "success resolves memberID and removes member",
			teamRef: "TeamX",
			userRef: "user@x.com",
			setupMocks: func(d sutDeps) {
				d.resolver.EXPECT().
					ResolveTeamRefToID(gomock.Any(), "TeamX").
					Return("team-id", nil).
					Times(1)

				d.resolver.EXPECT().
					ResolveTeamMemberRefToID(gomock.Any(), "team-id", "user@x.com").
					Return("member-id", nil).
					Times(1)

				d.ops.EXPECT().
					RemoveMember(gomock.Any(), "team-id", "member-id", "user@x.com").
					Return(nil).
					Times(1)
			},
		},
		{
			name:    "member resolver error is propagated",
			teamRef: "TeamX",
			userRef: "user@x.com",
			setupMocks: func(d sutDeps) {
				d.resolver.EXPECT().
					ResolveTeamRefToID(gomock.Any(), "TeamX").
					Return("team-id", nil).
					Times(1)

				d.resolver.EXPECT().
					ResolveTeamMemberRefToID(gomock.Any(), "team-id", "user@x.com").
					Return("", errors.New("boom")).
					Times(1)
			},
			wantErrAs: new(error),
		},
		{
			name:    "api error is mapped",
			teamRef: "TeamX",
			userRef: "user@x.com",
			setupMocks: func(d sutDeps) {
				d.resolver.EXPECT().
					ResolveTeamRefToID(gomock.Any(), "TeamX").
					Return("team-id", nil).
					Times(1)

				d.resolver.EXPECT().
					ResolveTeamMemberRefToID(gomock.Any(), "team-id", "user@x.com").
					Return("member-id", nil).
					Times(1)

				d.ops.EXPECT().
					RemoveMember(gomock.Any(), "team-id", "member-id", "user@x.com").
					Return(&sender.RequestError{Code: 403, Message: "nope"}).
					Times(1)
			},
			wantErrAs: new(sender.ErrAccessForbidden),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc, ctx := newSUT(t, tc.setupMocks)

			err := svc.RemoveMember(ctx, tc.teamRef, tc.userRef)

			if tc.wantErrAs == nil {
				require.NoError(t, err)
				return
			}

			require.Error(t, err)
			switch target := tc.wantErrAs.(type) {
			case *sender.ErrAccessForbidden:
				require.ErrorAs(t, err, target)
			case *error:
			default:
				t.Fatalf("unsupported wantErrAs type: %T", tc.wantErrAs)
			}
		})
	}
}

func TestService_UpdateMemberRoles(t *testing.T) {
	type testCase struct {
		name       string
		teamRef    string
		userRef    string
		isOwner    bool
		setupMocks func(d sutDeps)
		wantID     string
		wantErrAs  any
	}

	testCases := []testCase{
		{
			name:    "success promotes to owner",
			teamRef: "TeamX",
			userRef: "user@x.com",
			isOwner: true,
			setupMocks: func(d sutDeps) {
				d.resolver.EXPECT().
					ResolveTeamRefToID(gomock.Any(), "TeamX").
					Return("team-id", nil).
					Times(1)

				d.resolver.EXPECT().
					ResolveTeamMemberRefToID(gomock.Any(), "team-id", "user@x.com").
					Return("member-id", nil).
					Times(1)

				d.ops.EXPECT().
					UpdateMemberRoles(gomock.Any(), "team-id", "member-id", true).
					Return(&models.Member{ID: "member-id"}, nil).
					Times(1)
			},
			wantID: "member-id",
		},
		{
			name:    "success demotes to member",
			teamRef: "TeamX",
			userRef: "user@x.com",
			isOwner: false,
			setupMocks: func(d sutDeps) {
				d.resolver.EXPECT().
					ResolveTeamRefToID(gomock.Any(), "TeamX").
					Return("team-id", nil).
					Times(1)

				d.resolver.EXPECT().
					ResolveTeamMemberRefToID(gomock.Any(), "team-id", "user@x.com").
					Return("member-id", nil).
					Times(1)

				d.ops.EXPECT().
					UpdateMemberRoles(gomock.Any(), "team-id", "member-id", false).
					Return(&models.Member{ID: "member-id"}, nil).
					Times(1)
			},
			wantID: "member-id",
		},
		{
			name:    "api error is mapped",
			teamRef: "TeamX",
			userRef: "user@x.com",
			isOwner: true,
			setupMocks: func(d sutDeps) {
				d.resolver.EXPECT().
					ResolveTeamRefToID(gomock.Any(), "TeamX").
					Return("team-id", nil).
					Times(1)

				d.resolver.EXPECT().
					ResolveTeamMemberRefToID(gomock.Any(), "team-id", "user@x.com").
					Return("member-id", nil).
					Times(1)

				d.ops.EXPECT().
					UpdateMemberRoles(gomock.Any(), "team-id", "member-id", gomock.Any()).
					Return(nil, &sender.RequestError{Code: 403, Message: "nope"}).
					Times(1)
			},
			wantErrAs: new(sender.ErrAccessForbidden),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc, ctx := newSUT(t, tc.setupMocks)

			got, err := svc.UpdateMemberRoles(ctx, tc.teamRef, tc.userRef, tc.isOwner)

			if tc.wantErrAs != nil {
				require.Error(t, err)
				switch target := tc.wantErrAs.(type) {
				case *sender.ErrAccessForbidden:
					require.ErrorAs(t, err, target)
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
