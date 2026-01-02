package teams

import (
	"context"
	"fmt"

	"github.com/pzsp-teams/lib/internal/adapter"
	"github.com/pzsp-teams/lib/internal/api"
	"github.com/pzsp-teams/lib/internal/resources"
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


func (o *ops) GetTeamByID(ctx context.Context, teamID string) (*models.Team, error) {
	resp, requestErr := o.teamAPI.Get(ctx, teamID)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(resources.Team, teamID))
	}
	return adapter.MapGraphTeam(resp), nil
}

func (o *ops) ListMyJoinedTeams(ctx context.Context) ([]*models.Team, error) {
	resp, requestErr := o.teamAPI.ListMyJoined(ctx)
	if requestErr != nil {
		return nil, snd.MapError(requestErr)
	}
	return util.MapSlices(resp.GetValue(), adapter.MapGraphTeam), nil
}

func (o *ops) CreateViaGroup(ctx context.Context, displayName, mailNickname, visibility string) (*models.Team, error) {
	id, requestErr := o.teamAPI.CreateViaGroup(ctx, displayName, mailNickname, visibility)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(resources.Team, displayName))
	}

	t, ge := o.teamAPI.Get(ctx, id)
	if ge != nil {
		return nil, ge
	}

	return adapter.MapGraphTeam(t), nil
}

func (o *ops) CreateFromTemplate(ctx context.Context, displayName, description string, ownerIDs []string) (string, error) {
	id, err := o.teamAPI.CreateFromTemplate(ctx, displayName, description, ownerIDs)
	if err != nil {
		return id, snd.MapError(err, snd.WithResources(resources.User, ownerIDs))
	}
	return id, nil
}

func (o *ops) Archive(ctx context.Context, teamID, teamRef string, spoReadOnlyForMembers *bool) error {
	return snd.MapError(o.teamAPI.Archive(ctx, teamID, spoReadOnlyForMembers), snd.WithResource(resources.Team, teamID))
}

func (o *ops) Unarchive(ctx context.Context, teamID string) error {
	return snd.MapError(o.teamAPI.Unarchive(ctx, teamID), snd.WithResource(resources.Team, teamID))
}

func (o *ops) DeleteTeam(ctx context.Context, teamID, teamRef string) error {
	return snd.MapError(o.teamAPI.Delete(ctx, teamID), snd.WithResource(resources.Team, teamID))
}

func (o *ops) RestoreDeletedTeam(ctx context.Context, deletedGroupID string) (string, error) {
	obj, err := o.teamAPI.RestoreDeleted(ctx, deletedGroupID)
	if err != nil {
		return "", snd.MapError(err, snd.WithResource(resources.Team, deletedGroupID))
	}
	if obj == nil {
		return "", nil
	}
	id := util.Deref(obj.GetId())
	if id == "" {
		return "", fmt.Errorf("restored object has empty id")
	}
	return id, nil
}

func (o *ops) ListMembers(ctx context.Context, teamID string) ([]*models.Member, error) {
	resp, err := o.teamAPI.ListMembers(ctx, teamID)
	if err != nil {
		return nil, snd.MapError(err, snd.WithResource(resources.Team, teamID))
	}
	return util.MapSlices(resp.GetValue(), adapter.MapGraphMember), nil
}

func (o *ops) AddMember(ctx context.Context, teamID, userID string, isOwner bool) (*models.Member, error) {
	roles := util.MemberRole(isOwner)
	resp, err := o.teamAPI.AddMember(ctx, teamID, userID, roles)
	if err != nil {
		return nil, snd.MapError(err, snd.WithResource(resources.Team, teamID), snd.WithResource(resources.User, userID))
	}
	return adapter.MapGraphMember(resp), nil
}

func (o *ops) GetMemberByID(ctx context.Context, teamID, memberID string) (*models.Member, error) {
	resp, err := o.teamAPI.GetMember(ctx, teamID, memberID)
	if err != nil {
		return nil, snd.MapError(err, snd.WithResource(resources.Team, teamID), snd.WithResource(resources.User, memberID))
	}
	return adapter.MapGraphMember(resp), nil
}

func (o *ops) RemoveMember(ctx context.Context, teamID, memberID, userRef string) error {
	return snd.MapError(o.teamAPI.RemoveMember(ctx, teamID, memberID), snd.WithResource(resources.Team, teamID), snd.WithResource(resources.User, userRef))
}

func (o *ops) UpdateMemberRoles(ctx context.Context, teamID, memberID string, isOwner bool) (*models.Member, error) {
	roles := util.MemberRole(isOwner)
	resp, err := o.teamAPI.UpdateMemberRoles(ctx, teamID, memberID, roles)
	if err != nil {
		return nil, snd.MapError(err, snd.WithResource(resources.Team, teamID), snd.WithResource(resources.User, memberID))
	}
	return adapter.MapGraphMember(resp), nil
}
