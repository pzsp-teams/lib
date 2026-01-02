package teams

import (
	"context"

	"github.com/pzsp-teams/lib/internal/adapter"
	"github.com/pzsp-teams/lib/internal/api"
	snd "github.com/pzsp-teams/lib/internal/sender"
	"github.com/pzsp-teams/lib/internal/util"
	"github.com/pzsp-teams/lib/models"
)

type ops struct {
	teamAPI api.TeamAPI
}

func NewOps(teamAPI api.TeamAPI) teamsOps {
	return &ops{
		teamAPI: teamAPI,
	}
}

func (o *ops) Wait() {}

func (o *ops) GetTeamByID(ctx context.Context, teamID string) (*models.Team, *snd.RequestError) {
	resp, requestErr := o.teamAPI.Get(ctx, teamID)
	if requestErr != nil {
		return nil, requestErr
	}
	return adapter.MapGraphTeam(resp), nil
}

func (o *ops) ListMyJoinedTeams(ctx context.Context) ([]*models.Team, *snd.RequestError) {
	resp, requestErr := o.teamAPI.ListMyJoined(ctx)
	if requestErr != nil {
		return nil, requestErr
	}
	return util.MapSlices(resp.GetValue(), adapter.MapGraphTeam), nil
}

func (o *ops) CreateViaGroup(ctx context.Context, displayName, mailNickname, visibility string) (*models.Team, *snd.RequestError) {
	id, requestErr := o.teamAPI.CreateViaGroup(ctx, displayName, mailNickname, visibility)
	if requestErr != nil {
		return nil, requestErr
	}

	t, ge := o.teamAPI.Get(ctx, id)
	if ge != nil {
		return nil, ge
	}

	return adapter.MapGraphTeam(t), nil
}

func (o *ops) CreateFromTemplate(ctx context.Context, displayName, description string, ownerIDs []string) (string, *snd.RequestError) {
	id, requestErr := o.teamAPI.CreateFromTemplate(ctx, displayName, description, ownerIDs)
	if requestErr != nil {
		return "", requestErr
	}
	return id, nil
}

func (o *ops) Archive(ctx context.Context, teamID, teamRef string, spoReadOnlyForMembers *bool) *snd.RequestError {
	return o.teamAPI.Archive(ctx, teamID, spoReadOnlyForMembers)
}

func (o *ops) Unarchive(ctx context.Context, teamID string) *snd.RequestError {
	return o.teamAPI.Unarchive(ctx, teamID)
}

func (o *ops) DeleteTeam(ctx context.Context, teamID, teamRef string) *snd.RequestError {
	return o.teamAPI.Delete(ctx, teamID)
}

func (o *ops) RestoreDeletedTeam(ctx context.Context, deletedGroupID string) (string, *snd.RequestError) {
	obj, requestErr := o.teamAPI.RestoreDeleted(ctx, deletedGroupID)
	if requestErr != nil {
		return "", requestErr
	}
	if obj == nil {
		return "", nil
	}
	id := util.Deref(obj.GetId())
	return id, nil
}

func (o *ops) ListMembers(ctx context.Context, teamID string) ([]*models.Member, *snd.RequestError) {
	resp, requestErr := o.teamAPI.ListMembers(ctx, teamID)
	if requestErr != nil {
		return nil, requestErr
	}
	return util.MapSlices(resp.GetValue(), adapter.MapGraphMember), nil
}

func (o *ops) AddMember(ctx context.Context, teamID, userID string, isOwner bool) (*models.Member, *snd.RequestError) {
	roles := util.MemberRole(isOwner)
	resp, requestErr := o.teamAPI.AddMember(ctx, teamID, userID, roles)
	if requestErr != nil {
		return nil, requestErr
	}
	return adapter.MapGraphMember(resp), nil
}

func (o *ops) GetMemberByID(ctx context.Context, teamID, memberID string) (*models.Member, *snd.RequestError) {
	resp, requestErr := o.teamAPI.GetMember(ctx, teamID, memberID)
	if requestErr != nil {
		return nil, requestErr
	}
	return adapter.MapGraphMember(resp), nil
}

func (o *ops) RemoveMember(ctx context.Context, teamID, memberID, userRef string) *snd.RequestError {
	return o.teamAPI.RemoveMember(ctx, teamID, memberID)
}

func (o *ops) UpdateMemberRoles(ctx context.Context, teamID, memberID string, isOwner bool) (*models.Member, *snd.RequestError) {
	roles := util.MemberRole(isOwner)
	resp, requestErr := o.teamAPI.UpdateMemberRoles(ctx, teamID, memberID, roles)
	if requestErr != nil {
		return nil, requestErr
	}
	return adapter.MapGraphMember(resp), nil
}