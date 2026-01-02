package teams

import (
	"context"
	"fmt"

	"github.com/pzsp-teams/lib/internal/resolver"
	"github.com/pzsp-teams/lib/models"
)

type service struct {
	teamOps      teamsOps
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

	resp, err := s.teamOps.GetTeamByID(ctx, teamID)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *service) ListMyJoined(ctx context.Context) ([]*models.Team, error) {
	resp, err := s.teamOps.ListMyJoinedTeams(ctx)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *service) CreateViaGroup(ctx context.Context, displayName, mailNickname, visibility string) (*models.Team, error) {
	t, err := s.teamOps.CreateViaGroup(ctx, displayName, mailNickname, visibility)
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (s *service) CreateFromTemplate(ctx context.Context, displayName, description string, owners []string) (string, error) {
	id, err := s.teamOps.CreateFromTemplate(ctx, displayName, description, owners)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (s *service) Archive(ctx context.Context, teamRef string, spoReadOnlyForMembers *bool) error {
	teamID, err := s.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return err
	}
	
	if err = s.teamOps.Archive(ctx, teamID, teamRef, spoReadOnlyForMembers); err != nil {
		return err
	}

	return nil
}

func (s *service) Unarchive(ctx context.Context, teamRef string) error {
	teamID, err := s.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return err
	}

	if err := s.teamOps.Unarchive(ctx, teamID); err != nil {
		return err
	}

	return nil
}

func (s *service) Delete(ctx context.Context, teamRef string) error {
	teamID, err := s.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return err
	}

	if err := s.teamOps.DeleteTeam(ctx, teamID, teamRef); err != nil {
		return err
	}

	return nil
}

func (s *service) RestoreDeleted(ctx context.Context, deletedGroupID string) (string, error) {
	id, err := s.teamOps.RestoreDeletedTeam(ctx, deletedGroupID)
	if err != nil {
		return "", err
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

	resp, err := s.teamOps.ListMembers(ctx, teamID)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *service) AddMember(ctx context.Context, teamRef, userRef string, isOwner bool) (*models.Member, error) {
	teamID, err := s.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return nil, err
	}

	resp, err := s.teamOps.AddMember(ctx, teamID, userRef, isOwner)
	if err != nil {
		return nil, err
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

	resp, err := s.teamOps.GetMemberByID(ctx, teamID, memberID)
	if err != nil {
		return nil, err
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

	if err := s.teamOps.RemoveMember(ctx, teamID, memberID, userRef); err != nil {
		return err
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

	updated, err := s.teamOps.UpdateMemberRoles(ctx, teamID, memberID, isOwner)
	if err != nil {
		return nil, err
	}
	return updated, nil
}
