package teams

import (
	"context"
	"errors"
	"testing"

	gomock "github.com/golang/mock/gomock"
	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	sender "github.com/pzsp-teams/lib/internal/sender"
	"github.com/pzsp-teams/lib/internal/testutil"
	"github.com/pzsp-teams/lib/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type sutDeps struct {
	api      *testutil.MockTeamAPI
	resolver *testutil.MockTeamResolver
}

func newSUT(t *testing.T, setup func(d sutDeps)) (Service, context.Context) {
	t.Helper()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	apiMock := testutil.NewMockTeamAPI(ctrl)
	resolverMock := testutil.NewMockTeamResolver(ctrl)

	if setup != nil {
		setup(sutDeps{api: apiMock, resolver: resolverMock})
	}

	return NewService(apiMock, resolverMock), context.Background()
}

func TestService_ListMyJoined(t *testing.T) {
	type testCase struct {
		name       string
		setupMocks func(d sutDeps)
		assertFn   func(t *testing.T, got any, err error)
	}

	testCases := []testCase{
		{
			name: "maps teams",
			setupMocks: func(d sutDeps) {
				col := msmodels.NewTeamCollectionResponse()
				a := testutil.NewGraphTeam(&testutil.NewTeamParams{ID: util.Ptr("1"), DisplayName: util.Ptr("Alpha")})
				b := testutil.NewGraphTeam(&testutil.NewTeamParams{ID: util.Ptr("2"), DisplayName: util.Ptr("Beta")})
				col.SetValue([]msmodels.Teamable{a, b})

				d.api.EXPECT().
					ListMyJoined(gomock.Any()).
					Return(col, nil).
					Times(1)
			},
			assertFn: func(t *testing.T, got any, err error) {
				require.NoError(t, err)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc, ctx := newSUT(t, tc.setupMocks)

			got, err := svc.ListMyJoined(ctx)

			require.NoError(t, err)
			require.Len(t, got, 2)
			assert.Equal(t, "1", got[0].ID)
			assert.Equal(t, "Beta", got[1].DisplayName)
		})
	}

	t.Run("maps api error", func(t *testing.T) {
		svc, ctx := newSUT(t, func(d sutDeps) {
			d.api.EXPECT().
				ListMyJoined(gomock.Any()).
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

				d.api.EXPECT().
					Get(gomock.Any(), "resolved-42").
					Return(testutil.NewGraphTeam(&testutil.NewTeamParams{ID: util.Ptr("42"), DisplayName: util.Ptr("X")}), nil).
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

				d.api.EXPECT().
					Get(gomock.Any(), "missing-id").
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

				d.api.EXPECT().
					Delete(gomock.Any(), "team-id").
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

				d.api.EXPECT().
					Delete(gomock.Any(), "team-id").
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
				d.api.EXPECT().
					CreateViaGroup(gomock.Any(), "X", "x", "public").
					Return("", &sender.RequestError{Code: 403, Message: "nope"}).
					Times(1)
			},
			wantErrAs: new(sender.ErrAccessForbidden),
		},
		{
			name: "maps get error",
			setupMocks: func(d sutDeps) {
				d.api.EXPECT().
					CreateViaGroup(gomock.Any(), "X", "x", "public").
					Return("team-xyz", nil).
					Times(1)

				d.api.EXPECT().
					Get(gomock.Any(), "team-xyz").
					Return(nil, &sender.RequestError{Code: 404, Message: "not ready"}).
					Times(1)
			},
			wantErrAs: new(sender.ErrResourceNotFound),
		},
		{
			name: "success maps team",
			setupMocks: func(d sutDeps) {
				d.api.EXPECT().
					CreateViaGroup(gomock.Any(), "X", "x", "public").
					Return("team-xyz", nil).
					Times(1)

				d.api.EXPECT().
					Get(gomock.Any(), "team-xyz").
					Return(testutil.NewGraphTeam(&testutil.NewTeamParams{ID: util.Ptr("team-xyz"), DisplayName: util.Ptr("X")}), nil).
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
				case *sender.ErrResourceNotFound:
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
				d.api.EXPECT().
					CreateFromTemplate(gomock.Any(), "Tpl", "Desc", gomock.Any()).
					Return("tmpl-123", nil).
					Times(1)
			},
			wantID: "tmpl-123",
		},
		{
			name: "treats 201 in RequestError as success",
			setupMocks: func(d sutDeps) {
				d.api.EXPECT().
					CreateFromTemplate(gomock.Any(), "Tpl", "Desc", gomock.Any()).
					Return("tmpl-201", &sender.RequestError{Code: 201, Message: "created"}).
					Times(1)
			},
			wantID: "tmpl-201",
		},
		{
			name: "maps error",
			setupMocks: func(d sutDeps) {
				d.api.EXPECT().
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
				d.api.EXPECT().
					Archive(gomock.Any(), "team-id", &readOnly).
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
				d.api.EXPECT().
					Archive(gomock.Any(), "team-id", &readOnly).
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
				d.api.EXPECT().
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
				d.api.EXPECT().
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
				obj := msmodels.NewDirectoryObject()
				id := "restored-id"
				obj.SetId(&id)

				d.api.EXPECT().
					RestoreDeleted(gomock.Any(), "deleted-1").
					Return(obj, nil).
					Times(1)
			},
			wantID: "restored-id",
		},
		{
			name: "maps not found",
			setupMocks: func(d sutDeps) {
				d.api.EXPECT().
					RestoreDeleted(gomock.Any(), "deleted-1").
					Return(nil, &sender.RequestError{Code: 404, Message: "missing"}).
					Times(1)
			},
			wantErrAs: new(sender.ErrResourceNotFound),
		},
		{
			name: "empty object returns error",
			setupMocks: func(d sutDeps) {
				obj := msmodels.NewDirectoryObject()
				d.api.EXPECT().
					RestoreDeleted(gomock.Any(), "deleted-1").
					Return(obj, nil).
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

				col := msmodels.NewConversationMemberCollectionResponse()
				col.SetValue([]msmodels.ConversationMemberable{
					testutil.NewGraphMember(&testutil.NewMemberParams{
						ID: util.Ptr("m1"),
					}),
					testutil.NewGraphMember(&testutil.NewMemberParams{
						ID: util.Ptr("m2"),
					}),
				})

				d.api.EXPECT().
					ListMembers(gomock.Any(), "team-id").
					Return(col, nil).
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

				d.api.EXPECT().
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
					// any error ok
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

				d.api.EXPECT().
					AddMember(gomock.Any(), "team-id", "user@x.com", gomock.Any()).
					DoAndReturn(func(_ context.Context, _ string, _ string, roles []string) (msmodels.ConversationMemberable, *sender.RequestError) {
						require.Equal(t, []string{"owner"}, roles)
						return testutil.NewGraphMember(&testutil.NewMemberParams{
							ID: util.Ptr("m1"),
						}), nil
					}).
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

				d.api.EXPECT().
					AddMember(gomock.Any(), "team-id", "user@x.com", gomock.Any()).
					DoAndReturn(func(_ context.Context, _ string, _ string, roles []string) (msmodels.ConversationMemberable, *sender.RequestError) {
						require.Len(t, roles, 0)
						return testutil.NewGraphMember(&testutil.NewMemberParams{
							ID: util.Ptr("m2"),
						}), nil
					}).
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

				d.api.EXPECT().
					AddMember(gomock.Any(), "team-id", "user@x.com", gomock.Any()).
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

				d.api.EXPECT().
					GetMember(gomock.Any(), "team-id", "member-id").
					Return(testutil.NewGraphMember(&testutil.NewMemberParams{
						ID: util.Ptr("member-id"),
					}), nil).
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

				d.api.EXPECT().
					GetMember(gomock.Any(), "team-id", "member-id").
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

				d.api.EXPECT().
					RemoveMember(gomock.Any(), "team-id", "member-id").
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

				d.api.EXPECT().
					RemoveMember(gomock.Any(), "team-id", "member-id").
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

				d.api.EXPECT().
					UpdateMemberRoles(gomock.Any(), "team-id", "member-id", gomock.Any()).
					DoAndReturn(func(_ context.Context, _ string, _ string, roles []string) (msmodels.ConversationMemberable, *sender.RequestError) {
						require.Equal(t, []string{"owner"}, roles)
						return testutil.NewGraphMember(&testutil.NewMemberParams{
							ID: util.Ptr("member-id"),
						}), nil
					}).
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

				d.api.EXPECT().
					UpdateMemberRoles(gomock.Any(), "team-id", "member-id", gomock.Any()).
					DoAndReturn(func(_ context.Context, _ string, _ string, roles []string) (msmodels.ConversationMemberable, *sender.RequestError) {
						require.Len(t, roles, 0)
						return testutil.NewGraphMember(&testutil.NewMemberParams{
							ID: util.Ptr("member-id"),
						}), nil
					}).
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

				d.api.EXPECT().
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
