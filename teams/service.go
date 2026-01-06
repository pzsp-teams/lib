package teams

import (
	"context"

	"github.com/pzsp-teams/lib/internal/resolver"
	"github.com/pzsp-teams/lib/internal/resources"
	"github.com/pzsp-teams/lib/internal/sender"
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
		return nil, sender.Wrap("Get", err,
			sender.NewParam(resources.TeamRef, teamRef),
		)
	}

	resp, err := s.teamOps.GetTeamByID(ctx, teamID)
	if err != nil {
		return nil, sender.Wrap("Get", err,
			sender.NewParam(resources.TeamRef, teamRef),
		)
	}

	return resp, nil
}

func (s *service) ListMyJoined(ctx context.Context) ([]*models.Team, error) {
	resp, err := s.teamOps.ListMyJoinedTeams(ctx)
	if err != nil {
		return nil, sender.Wrap("ListMyJoined", err)
	}
	return resp, nil
}

func (s *service) CreateViaGroup(ctx context.Context, displayName, mailNickname, visibility string) (*models.Team, error) {
	t, err := s.teamOps.CreateViaGroup(ctx, displayName, mailNickname, visibility)
	if err != nil {
		return nil, sender.Wrap("CreateViaGroup", err)
	}

	return t, nil
}

func (s *service) CreateFromTemplate(ctx context.Context, displayName, description string, owners, members []string, visibility string, includeMe bool) (string, error) {
	id, err := s.teamOps.CreateFromTemplate(ctx, displayName, description, owners, members, visibility, includeMe)
	if err != nil {
		return "", sender.Wrap("CreateFromTemplate", err)
	}

	return id, nil
}

func (s *service) Archive(ctx context.Context, teamRef string, spoReadOnlyForMembers *bool) error {
	teamID, err := s.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return sender.Wrap("Archive", err,
			sender.NewParam(resources.TeamRef, teamRef),
		)
	}

	if err = s.teamOps.Archive(ctx, teamID, teamRef, spoReadOnlyForMembers); err != nil {
		return sender.Wrap("Archive", err,
			sender.NewParam(resources.TeamRef, teamRef),
		)
	}

	return nil
}

func (s *service) Unarchive(ctx context.Context, teamRef string) error {
	teamID, err := s.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return sender.Wrap("Unarchive", err,
			sender.NewParam(resources.TeamRef, teamRef),
		)
	}

	if err := s.teamOps.Unarchive(ctx, teamID); err != nil {
		return sender.Wrap("Unarchive", err,
			sender.NewParam(resources.TeamRef, teamRef),
		)
	}

	return nil
}

func (s *service) Delete(ctx context.Context, teamRef string) error {
	teamID, err := s.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return sender.Wrap("Delete", err,
			sender.NewParam(resources.TeamRef, teamRef),
		)
	}

	if err := s.teamOps.DeleteTeam(ctx, teamID, teamRef); err != nil {
		return sender.Wrap("Delete", err,
			sender.NewParam(resources.TeamRef, teamRef),
		)
	}

	return nil
}

func (s *service) RestoreDeleted(ctx context.Context, deletedGroupID string) (string, error) {
	id, err := s.teamOps.RestoreDeletedTeam(ctx, deletedGroupID)
	if err != nil {
		return "", sender.Wrap("RestoreDeleted", err)
	}

	return id, nil
}

func (s *service) ListMembers(ctx context.Context, teamRef string) ([]*models.Member, error) {
	teamID, err := s.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return nil, sender.Wrap("ListMembers", err,
			sender.NewParam(resources.TeamRef, teamRef),
		)
	}

	resp, err := s.teamOps.ListMembers(ctx, teamID)
	if err != nil {
		return nil, sender.Wrap("ListMembers", err,
			sender.NewParam(resources.TeamRef, teamRef),
		)
	}

	return resp, nil
}

func (s *service) AddMember(ctx context.Context, teamRef, userRef string, isOwner bool) (*models.Member, error) {
	teamID, err := s.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return nil, sender.Wrap("AddMember", err,
			sender.NewParam(resources.TeamRef, teamRef),
			sender.NewParam(resources.UserRef, userRef),
		)
	}

	resp, err := s.teamOps.AddMember(ctx, teamID, userRef, isOwner)
	if err != nil {
		return nil, sender.Wrap("AddMember", err,
			sender.NewParam(resources.TeamRef, teamRef),
			sender.NewParam(resources.UserRef, userRef),
		)
	}
	return resp, nil
}

func (s *service) GetMember(ctx context.Context, teamRef, userRef string) (*models.Member, error) {
	teamID, err := s.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return nil, sender.Wrap("GetMember", err,
			sender.NewParam(resources.TeamRef, teamRef),
			sender.NewParam(resources.UserRef, userRef),
		)
	}

	memberID, err := s.teamResolver.ResolveTeamMemberRefToID(ctx, teamID, userRef)
	if err != nil {
		return nil, sender.Wrap("GetMember", err,
			sender.NewParam(resources.TeamRef, teamRef),
			sender.NewParam(resources.UserRef, userRef),
		)
	}

	resp, err := s.teamOps.GetMemberByID(ctx, teamID, memberID)
	if err != nil {
		return nil, sender.Wrap("GetMember", err,
			sender.NewParam(resources.TeamRef, teamRef),
			sender.NewParam(resources.UserRef, userRef),
		)
	}

	return resp, nil
}

func (s *service) RemoveMember(ctx context.Context, teamRef, userRef string) error {
	teamID, err := s.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return sender.Wrap("RemoveMember", err,
			sender.NewParam(resources.TeamRef, teamRef),
			sender.NewParam(resources.UserRef, userRef),
		)
	}

	memberID, err := s.teamResolver.ResolveTeamMemberRefToID(ctx, teamID, userRef)
	if err != nil {
		return sender.Wrap("RemoveMember", err,
			sender.NewParam(resources.TeamRef, teamRef),
			sender.NewParam(resources.UserRef, userRef),
		)
	}

	if err := s.teamOps.RemoveMember(ctx, teamID, memberID, userRef); err != nil {
		return sender.Wrap("RemoveMember", err,
			sender.NewParam(resources.TeamRef, teamRef),
			sender.NewParam(resources.UserRef, userRef),
		)
	}

	return nil
}

func (s *service) UpdateMemberRoles(ctx context.Context, teamRef, userRef string, isOwner bool) (*models.Member, error) {
	teamID, err := s.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return nil, sender.Wrap("UpdateMemberRoles", err,
			sender.NewParam(resources.TeamRef, teamRef),
			sender.NewParam(resources.UserRef, userRef),
		)
	}

	memberID, err := s.teamResolver.ResolveTeamMemberRefToID(ctx, teamID, userRef)
	if err != nil {
		return nil, sender.Wrap("UpdateMemberRoles", err,
			sender.NewParam(resources.TeamRef, teamRef),
			sender.NewParam(resources.UserRef, userRef),
		)
	}

	updated, err := s.teamOps.UpdateMemberRoles(ctx, teamID, memberID, isOwner)
	if err != nil {
		return nil, sender.Wrap("UpdateMemberRoles", err,
			sender.NewParam(resources.TeamRef, teamRef),
			sender.NewParam(resources.UserRef, userRef),
		)
	}
	return updated, nil
}
