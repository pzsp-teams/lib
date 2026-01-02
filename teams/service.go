package teams

import (
	"context"
	"fmt"

	"github.com/pzsp-teams/lib/internal/resolver"
	"github.com/pzsp-teams/lib/internal/resources"
	snd "github.com/pzsp-teams/lib/internal/sender"
	"github.com/pzsp-teams/lib/models"
)

type service struct {
	teamOps    teamsOps
	teamResolver resolver.TeamResolver
}

// NewService creates a new Service instance.
func NewService(teamOps teamsOps, tr resolver.TeamResolver) Service {
	return &service{teamOps: teamOps, teamResolver: tr}
}

func (s *service) Get(ctx context.Context, teamRef string) (*models.Team, error) {
	teamID, err := s.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return nil, err
	}

	resp, requestErr := s.teamOps.GetTeamByID(ctx, teamID)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(resources.Team, teamRef))
	}

	return resp, nil
}

func (s *service) ListMyJoined(ctx context.Context) ([]*models.Team, error) {
	resp, requestErr := s.teamOps.ListMyJoinedTeams(ctx)
	if requestErr != nil {
		return nil, snd.MapError(requestErr)
	}
	return resp, nil
}

func (s *service) CreateViaGroup(ctx context.Context, displayName, mailNickname, visibility string) (*models.Team, error) {
	t, requestErr := s.teamOps.CreateViaGroup(ctx, displayName, mailNickname, visibility)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(resources.Team, displayName))
	}

	return t, nil
}

func (s *service) CreateFromTemplate(ctx context.Context, displayName, description string, owners []string) (string, error) {
	id, requestErr := s.teamOps.CreateFromTemplate(ctx, displayName, description, owners)
	if requestErr != nil {
		return "", snd.MapError(requestErr, snd.WithResources(resources.User, owners))
	}

	return id, nil
}

func (s *service) Archive(ctx context.Context, teamRef string, spoReadOnlyForMembers *bool) error {
	teamID, err := s.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return err
	}

	if e := s.teamOps.Archive(ctx, teamID, teamRef, spoReadOnlyForMembers); e != nil {
		return snd.MapError(e, snd.WithResource(resources.Team, teamRef))
	}

	return nil
}

func (s *service) Unarchive(ctx context.Context, teamRef string) error {
	teamID, err := s.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return err
	}

	if requestErr := s.teamOps.Unarchive(ctx, teamID); requestErr != nil {
		return snd.MapError(requestErr, snd.WithResource(resources.Team, teamRef))
	}

	return nil
}

func (s *service) Delete(ctx context.Context, teamRef string) error {
	teamID, err := s.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return err
	}

	if requestErr := s.teamOps.DeleteTeam(ctx, teamID, teamRef); requestErr != nil {
		return snd.MapError(requestErr, snd.WithResource(resources.Team, teamRef))
	}

	return nil
}

func (s *service) RestoreDeleted(ctx context.Context, deletedGroupID string) (string, error) {
	id, err := s.teamOps.RestoreDeletedTeam(ctx, deletedGroupID)
	if err != nil {
		return "", snd.MapError(err, snd.WithResource(resources.Team, deletedGroupID))
	}

	if id == "" {
		return "", fmt.Errorf("restored object has empty id")
	}

	return id, nil
}

func (s *service) ListMembers(ctx context.Context, teamRef string) ([]*models.Member, error) {
	teamID, err := s.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return nil, err
	}

	resp, requestErr := s.teamOps.ListMembers(ctx, teamID)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(resources.Team, teamRef))
	}

	return resp, nil
}

func (s *service) AddMember(ctx context.Context, teamRef, userRef string, isOwner bool) (*models.Member, error) {
	teamID, err := s.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return nil, err
	}

	resp, requestErr := s.teamOps.AddMember(ctx, teamID, userRef, isOwner)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(resources.Team, teamRef), snd.WithResource(resources.User, userRef))
	}
	return resp, nil
}

func (s *service) GetMember(ctx context.Context, teamRef, userRef string) (*models.Member, error) {
	teamID, err := s.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return nil, err
	}

	memberID, err := s.teamResolver.ResolveTeamMemberRefToID(ctx, teamID, userRef)
	if err != nil {
		return nil, err
	}

	resp, requestErr := s.teamOps.GetMemberByID(ctx, teamID, memberID)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(resources.Team, teamRef), snd.WithResource(resources.User, userRef))
	}

	return resp, nil
}

func (s *service) RemoveMember(ctx context.Context, teamRef, userRef string) error {
	teamID, err := s.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return err
	}

	memberID, err := s.teamResolver.ResolveTeamMemberRefToID(ctx, teamID, userRef)
	if err != nil {
		return err
	}

	if requestErr := s.teamOps.RemoveMember(ctx, teamID, memberID, userRef); requestErr != nil {
		return snd.MapError(requestErr, snd.WithResource(resources.Team, teamRef), snd.WithResource(resources.User, userRef))
	}

	return nil
}

func (s *service) UpdateMemberRoles(ctx context.Context, teamRef, userRef string, isOwner bool) (*models.Member, error) {
	teamID, err := s.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return nil, err
	}

	memberID, err := s.teamResolver.ResolveTeamMemberRefToID(ctx, teamID, userRef)
	if err != nil {
		return nil, err
	}

	updated, requestErr := s.teamOps.UpdateMemberRoles(ctx, teamID, memberID, isOwner)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(resources.Team, teamRef), snd.WithResource(resources.User, userRef))
	}
	return updated, nil
}
