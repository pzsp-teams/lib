package teams

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


type sutDepsWithCache struct {
	teamOps *testutil.MockteamsOps
	cacher  *testutil.MockCacher
	runner  *testutil.MockTaskRunner
}

func newSUTWithCache(t *testing.T) (teamsOps, sutDepsWithCache) {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	deps := sutDepsWithCache{
		teamOps: testutil.NewMockteamsOps(ctrl),
		cacher:  testutil.NewMockCacher(ctrl),
		runner:  testutil.NewMockTaskRunner(ctrl),
	}

	cacheHandler := &cacher.CacheHandler{
		Cacher: deps.cacher,
		Runner: deps.runner,
	}

	sut := NewOpsWithCache(deps.teamOps, cacheHandler)
	return sut, deps
}

func expectRunNow(r *testutil.MockTaskRunner) {
	r.EXPECT().Run(gomock.Any()).DoAndReturn(func(fn func()) {
		fn()
	})
}

func reqErr() *snd.RequestError {
	return &snd.RequestError{Code: 400}
}


func TestNewOpsWithCache_WhenCacheNil_ReturnsOriginalOps(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	teamOps := testutil.NewMockteamsOps(ctrl)

	got := NewOpsWithCache(teamOps, nil)
	require.Same(t, teamOps, got)
}

func TestOpsWithCache_Wait(t *testing.T) {
	sut, d := newSUTWithCache(t)
	d.runner.EXPECT().Wait().Times(1)
	sut.Wait()
}

func TestOpsWithCache_GetTeamByID(t *testing.T) {
	team := &models.Team{ID: "t1", DisplayName: "Team A"}

	tests := []struct {
		name    string
		setup   func(d sutDepsWithCache)
		want    *models.Team
		wantErr *snd.RequestError
	}{
		{
			name: "Success - Caches team",
			setup: func(d sutDepsWithCache) {
				d.teamOps.EXPECT().GetTeamByID(gomock.Any(), "id").Return(team, nil)
				expectRunNow(d.runner)
				d.cacher.EXPECT().Set(cacher.NewTeamKey(team.DisplayName), team.ID).Return(nil)
			},
			want: team,
		},
		{
			name: "Success - Nil team - Does not cache",
			setup: func(d sutDepsWithCache) {
				d.teamOps.EXPECT().GetTeamByID(gomock.Any(), "id").Return(nil, nil)
			},
			want: nil,
		},
		{
			name: "Error - Clears cache",
			setup: func(d sutDepsWithCache) {
				d.teamOps.EXPECT().GetTeamByID(gomock.Any(), "id").Return(nil, reqErr())
				expectRunNow(d.runner)
				d.cacher.EXPECT().Clear().Return(nil)
			},
			wantErr: reqErr(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sut, d := newSUTWithCache(t)
			tc.setup(d)
			got, err := sut.GetTeamByID(context.Background(), "id")

			if tc.wantErr != nil {
				require.Equal(t, tc.wantErr, err)
			} else {
				require.Nil(t, err)
			}
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestOpsWithCache_ListMyJoinedTeams(t *testing.T) {
	out := []*models.Team{{ID: "t1", DisplayName: "A"}, nil, {ID: "t2", DisplayName: " "}}

	tests := []struct {
		name    string
		setup   func(d sutDepsWithCache)
		want    []*models.Team
		wantErr *snd.RequestError
	}{
		{
			name: "Success - Caches valid teams",
			setup: func(d sutDepsWithCache) {
				d.teamOps.EXPECT().ListMyJoinedTeams(gomock.Any()).Return(out, nil)
				expectRunNow(d.runner)
				d.cacher.EXPECT().Set(cacher.NewTeamKey("A"), "t1").Return(nil)
			},
			want: out,
		},
		{
			name: "Error - Clears cache",
			setup: func(d sutDepsWithCache) {
				d.teamOps.EXPECT().ListMyJoinedTeams(gomock.Any()).Return(nil, reqErr())
				expectRunNow(d.runner)
				d.cacher.EXPECT().Clear().Return(nil)
			},
			wantErr: reqErr(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sut, d := newSUTWithCache(t)
			tc.setup(d)
			got, err := sut.ListMyJoinedTeams(context.Background())

			if tc.wantErr != nil {
				require.Equal(t, tc.wantErr, err)
			} else {
				require.Nil(t, err)
			}
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestOpsWithCache_CreateFromTemplate(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(d sutDepsWithCache)
		wantID  string
		wantErr *snd.RequestError
	}{
		{
			name: "Success - Invalidates team key",
			setup: func(d sutDepsWithCache) {
				d.teamOps.EXPECT().CreateFromTemplate(gomock.Any(), "Team A", "d", gomock.Any()).Return("id", nil)
				expectRunNow(d.runner)
				d.cacher.EXPECT().Invalidate(cacher.NewTeamKey("Team A")).Return(nil)
			},
			wantID: "id",
		},
		{
			name: "Error - Clears cache",
			setup: func(d sutDepsWithCache) {
				d.teamOps.EXPECT().CreateFromTemplate(gomock.Any(), "Team A", "d", gomock.Any()).Return("id", reqErr())
				expectRunNow(d.runner)
				d.cacher.EXPECT().Clear().Return(nil)
			},
			wantID:  "id",
			wantErr: reqErr(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sut, d := newSUTWithCache(t)
			tc.setup(d)
			id, err := sut.CreateFromTemplate(context.Background(), "Team A", "d", []string{"u1"})

			if tc.wantErr != nil {
				require.Equal(t, tc.wantErr, err)
			} else {
				require.Nil(t, err)
			}
			assert.Equal(t, tc.wantID, id)
		})
	}
}

func TestOpsWithCache_CreateViaGroup(t *testing.T) {
	team := &models.Team{ID: "t1", DisplayName: "Team A"}

	tests := []struct {
		name    string
		setup   func(d sutDepsWithCache)
		want    *models.Team
		wantErr *snd.RequestError
	}{
		{
			name: "Success - Caches team",
			setup: func(d sutDepsWithCache) {
				d.teamOps.EXPECT().CreateViaGroup(gomock.Any(), "Team A", "n", "p").Return(team, nil)
				expectRunNow(d.runner)
				d.cacher.EXPECT().Set(cacher.NewTeamKey(team.DisplayName), team.ID).Return(nil)
			},
			want: team,
		},
		{
			name: "Error - Clears cache",
			setup: func(d sutDepsWithCache) {
				d.teamOps.EXPECT().CreateViaGroup(gomock.Any(), "Team A", "n", "p").Return(nil, reqErr())
				expectRunNow(d.runner)
				d.cacher.EXPECT().Clear().Return(nil)
			},
			wantErr: reqErr(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sut, d := newSUTWithCache(t)
			tc.setup(d)
			got, err := sut.CreateViaGroup(context.Background(), "Team A", "n", "p")

			if tc.wantErr != nil {
				require.Equal(t, tc.wantErr, err)
			} else {
				require.Nil(t, err)
			}
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestOpsWithCache_Archive(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(d sutDepsWithCache)
		wantErr *snd.RequestError
	}{
		{
			name: "Success - Invalidates team key",
			setup: func(d sutDepsWithCache) {
				d.teamOps.EXPECT().Archive(gomock.Any(), "tid", "Team A", nil).Return(nil)
				expectRunNow(d.runner)
				d.cacher.EXPECT().Invalidate(cacher.NewTeamKey("Team A")).Return(nil)
			},
		},
		{
			name: "Error - Clears cache",
			setup: func(d sutDepsWithCache) {
				d.teamOps.EXPECT().Archive(gomock.Any(), "tid", "Team A", nil).Return(reqErr())
				expectRunNow(d.runner)
				d.cacher.EXPECT().Clear().Return(nil)
			},
			wantErr: reqErr(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sut, d := newSUTWithCache(t)
			tc.setup(d)
			err := sut.Archive(context.Background(), "tid", "Team A", nil)
			require.Equal(t, tc.wantErr, err)
		})
	}
}

func TestOpsWithCache_Unarchive(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(d sutDepsWithCache)
		wantErr *snd.RequestError
	}{
		{
			name: "Success - No cache ops",
			setup: func(d sutDepsWithCache) {
				d.teamOps.EXPECT().Unarchive(gomock.Any(), "tid").Return(nil)
			},
		},
		{
			name: "Error - Clears cache",
			setup: func(d sutDepsWithCache) {
				d.teamOps.EXPECT().Unarchive(gomock.Any(), "tid").Return(reqErr())
				expectRunNow(d.runner)
				d.cacher.EXPECT().Clear().Return(nil)
			},
			wantErr: reqErr(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sut, d := newSUTWithCache(t)
			tc.setup(d)
			err := sut.Unarchive(context.Background(), "tid")
			require.Equal(t, tc.wantErr, err)
		})
	}
}

func TestOpsWithCache_DeleteTeam(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(d sutDepsWithCache)
		wantErr *snd.RequestError
	}{
		{
			name: "Success - Invalidates team key",
			setup: func(d sutDepsWithCache) {
				d.teamOps.EXPECT().DeleteTeam(gomock.Any(), "tid", "Team A").Return(nil)
				expectRunNow(d.runner)
				d.cacher.EXPECT().Invalidate(cacher.NewTeamKey("Team A")).Return(nil)
			},
		},
		{
			name: "Error - Clears cache",
			setup: func(d sutDepsWithCache) {
				d.teamOps.EXPECT().DeleteTeam(gomock.Any(), "tid", "Team A").Return(reqErr())
				expectRunNow(d.runner)
				d.cacher.EXPECT().Clear().Return(nil)
			},
			wantErr: reqErr(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sut, d := newSUTWithCache(t)
			tc.setup(d)
			err := sut.DeleteTeam(context.Background(), "tid", "Team A")
			require.Equal(t, tc.wantErr, err)
		})
	}
}

func TestOpsWithCache_RestoreDeletedTeam(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(d sutDepsWithCache)
		wantID  string
		wantErr *snd.RequestError
	}{
		{
			name: "Success - No cache ops",
			setup: func(d sutDepsWithCache) {
				d.teamOps.EXPECT().RestoreDeletedTeam(gomock.Any(), "did").Return("rid", nil)
			},
			wantID: "rid",
		},
		{
			name: "Error - Clears cache",
			setup: func(d sutDepsWithCache) {
				d.teamOps.EXPECT().RestoreDeletedTeam(gomock.Any(), "did").Return("", reqErr())
				expectRunNow(d.runner)
				d.cacher.EXPECT().Clear().Return(nil)
			},
			wantErr: reqErr(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sut, d := newSUTWithCache(t)
			tc.setup(d)
			id, err := sut.RestoreDeletedTeam(context.Background(), "did")
			assert.Equal(t, tc.wantID, id)
			require.Equal(t, tc.wantErr, err)
		})
	}
}

func TestOpsWithCache_ListMembers(t *testing.T) {
	members := []*models.Member{{ID: "m1", Email: "a@b.com"}, nil, {ID: "m2", Email: ""}}

	tests := []struct {
		name    string
		setup   func(d sutDepsWithCache)
		want    []*models.Member
		wantErr *snd.RequestError
	}{
		{
			name: "Success - Caches valid members",
			setup: func(d sutDepsWithCache) {
				d.teamOps.EXPECT().ListMembers(gomock.Any(), "tid").Return(members, nil)
				expectRunNow(d.runner)
				d.cacher.EXPECT().Set(cacher.NewTeamMemberKey("tid", "a@b.com", nil), "m1").Return(nil)
			},
			want: members,
		},
		{
			name: "Error - Clears cache",
			setup: func(d sutDepsWithCache) {
				d.teamOps.EXPECT().ListMembers(gomock.Any(), "tid").Return(nil, reqErr())
				expectRunNow(d.runner)
				d.cacher.EXPECT().Clear().Return(nil)
			},
			wantErr: reqErr(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sut, d := newSUTWithCache(t)
			tc.setup(d)
			got, err := sut.ListMembers(context.Background(), "tid")
			require.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestOpsWithCache_GetMemberByID(t *testing.T) {
	member := &models.Member{ID: "m1", Email: "a@b.com"}

	tests := []struct {
		name    string
		setup   func(d sutDepsWithCache)
		want    *models.Member
		wantErr *snd.RequestError
	}{
		{
			name: "Success - Caches member",
			setup: func(d sutDepsWithCache) {
				d.teamOps.EXPECT().GetMemberByID(gomock.Any(), "tid", "mid").Return(member, nil)
				expectRunNow(d.runner)
				d.cacher.EXPECT().Set(cacher.NewTeamMemberKey("tid", "a@b.com", nil), "m1").Return(nil)
			},
			want: member,
		},
		{
			name: "Success - Nil member - No cache",
			setup: func(d sutDepsWithCache) {
				d.teamOps.EXPECT().GetMemberByID(gomock.Any(), "tid", "mid").Return(nil, nil)
			},
			want: nil,
		},
		{
			name: "Error - Clears cache",
			setup: func(d sutDepsWithCache) {
				d.teamOps.EXPECT().GetMemberByID(gomock.Any(), "tid", "mid").Return(nil, reqErr())
				expectRunNow(d.runner)
				d.cacher.EXPECT().Clear().Return(nil)
			},
			wantErr: reqErr(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sut, d := newSUTWithCache(t)
			tc.setup(d)
			got, err := sut.GetMemberByID(context.Background(), "tid", "mid")
			require.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestOpsWithCache_AddMember(t *testing.T) {
	member := &models.Member{ID: "m1", Email: "a@b.com"}

	tests := []struct {
		name    string
		setup   func(d sutDepsWithCache)
		want    *models.Member
		wantErr *snd.RequestError
	}{
		{
			name: "Success - Caches member",
			setup: func(d sutDepsWithCache) {
				d.teamOps.EXPECT().AddMember(gomock.Any(), "tid", "uid", true).Return(member, nil)
				expectRunNow(d.runner)
				d.cacher.EXPECT().Set(cacher.NewTeamMemberKey("tid", "a@b.com", nil), "m1").Return(nil)
			},
			want: member,
		},
		{
			name: "Error - Clears cache",
			setup: func(d sutDepsWithCache) {
				d.teamOps.EXPECT().AddMember(gomock.Any(), "tid", "uid", true).Return(nil, reqErr())
				expectRunNow(d.runner)
				d.cacher.EXPECT().Clear().Return(nil)
			},
			wantErr: reqErr(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sut, d := newSUTWithCache(t)
			tc.setup(d)
			got, err := sut.AddMember(context.Background(), "tid", "uid", true)
			require.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestOpsWithCache_UpdateMemberRoles(t *testing.T) {
	member := &models.Member{ID: "m1"}

	tests := []struct {
		name    string
		setup   func(d sutDepsWithCache)
		want    *models.Member
		wantErr *snd.RequestError
	}{
		{
			name: "Success - No cache ops",
			setup: func(d sutDepsWithCache) {
				d.teamOps.EXPECT().UpdateMemberRoles(gomock.Any(), "tid", "mid", true).Return(member, nil)
			},
			want: member,
		},
		{
			name: "Error - Clears cache",
			setup: func(d sutDepsWithCache) {
				d.teamOps.EXPECT().UpdateMemberRoles(gomock.Any(), "tid", "mid", true).Return(nil, reqErr())
				expectRunNow(d.runner)
				d.cacher.EXPECT().Clear().Return(nil)
			},
			wantErr: reqErr(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sut, d := newSUTWithCache(t)
			tc.setup(d)
			got, err := sut.UpdateMemberRoles(context.Background(), "tid", "mid", true)
			require.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestOpsWithCache_RemoveMember(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(d sutDepsWithCache)
		wantErr *snd.RequestError
	}{
		{
			name: "Success - Invalidates member key",
			setup: func(d sutDepsWithCache) {
				d.teamOps.EXPECT().RemoveMember(gomock.Any(), "tid", "mid", "a@b.com").Return(nil)
				expectRunNow(d.runner)
				d.cacher.EXPECT().Invalidate(cacher.NewTeamMemberKey("tid", "a@b.com", nil)).Return(nil)
			},
		},
		{
			name: "Error - Clears cache",
			setup: func(d sutDepsWithCache) {
				d.teamOps.EXPECT().RemoveMember(gomock.Any(), "tid", "mid", "a@b.com").Return(reqErr())
				expectRunNow(d.runner)
				d.cacher.EXPECT().Clear().Return(nil)
			},
			wantErr: reqErr(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sut, d := newSUTWithCache(t)
			tc.setup(d)
			err := sut.RemoveMember(context.Background(), "tid", "mid", "a@b.com")
			require.Equal(t, tc.wantErr, err)
		})
	}
}
