package teams

import (
	"context"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/internal/api"
	"github.com/pzsp-teams/lib/internal/mapper"
)

// Service will be used later
type Service struct {
	teamAPI    api.TeamAPI
	teamMapper mapper.TeamMapper
}

// NewService will be used later
func NewService(teamsAPI api.TeamAPI, m mapper.TeamMapper) *Service {
	return &Service{teamAPI: teamsAPI, teamMapper: m}
}

// Get will be used later
func (s *Service) Get(ctx context.Context, teamRef string) (*Team, error) {
	teamID, err := s.teamMapper.MapTeamRefToTeamID(ctx, teamRef)
	if err != nil {
		return nil, err
	}
	resp, senderErr := s.teamAPI.Get(ctx, teamID)
	if senderErr != nil {
		return nil, mapError(senderErr)
	}
	return mapGraphTeam(resp), nil
}

// ListMyJoined will be used later
func (s *Service) ListMyJoined(ctx context.Context) ([]*Team, error) {
	resp, err := s.teamAPI.ListMyJoined(ctx)
	if err != nil {
		return nil, mapError(err)
	}
	var out []*Team
	if resp != nil && resp.GetValue() != nil {
		for _, t := range resp.GetValue() {
			out = append(out, mapGraphTeam(t))
		}
	}
	return out, nil
}

// Update will be used later
func (s *Service) Update(ctx context.Context, teamRef string, patch *msmodels.Team) (*Team, error) {
	teamID, err := s.teamMapper.MapTeamRefToTeamID(ctx, teamRef)
	if err != nil {
		return nil, err
	}
	resp, senderErr := s.teamAPI.Update(ctx, teamID, patch)
	if senderErr != nil {
		return nil, mapError(senderErr)
	}
	return mapGraphTeam(resp), nil
}

// CreateViaGroup will be used later
func (s *Service) CreateViaGroup(ctx context.Context, displayName, mailNickname, visibility string) (*Team, error) {
	id, err := s.teamAPI.CreateViaGroup(ctx, displayName, mailNickname, visibility)
	if err != nil {
		return nil, mapError(err)
	}
	t, ge := s.teamAPI.Get(ctx, id)
	if ge != nil {
		return nil, mapError(ge)
	}
	return mapGraphTeam(t), nil
}

// CreateFromTemplate will be used later
func (s *Service) CreateFromTemplate(ctx context.Context, displayName, description string, owners []string) (string, error) {
	id, err := s.teamAPI.CreateFromTemplate(ctx, displayName, description, owners)
	if err != nil {
		if err.Code == "AsyncOperation" {
			return id, nil
		}
		return "", mapError(err)
	}
	return id, nil
}

// Archive will be used later
func (s *Service) Archive(ctx context.Context, teamRef string, spoReadOnlyForMembers *bool) error {
	teamID, err := s.teamMapper.MapTeamRefToTeamID(ctx, teamRef)
	if err != nil {
		return err
	}
	if e := s.teamAPI.Archive(ctx, teamID, spoReadOnlyForMembers); e != nil {
		return mapError(e)
	}
	return nil
}

// Unarchive will be used later
func (s *Service) Unarchive(ctx context.Context, teamRef string) error {
	teamID, err := s.teamMapper.MapTeamRefToTeamID(ctx, teamRef)
	if err != nil {
		return err
	}
	if e := s.teamAPI.Unarchive(ctx, teamID); e != nil {
		return mapError(e)
	}
	return nil
}

// Delete will be used later
func (s *Service) Delete(ctx context.Context, teamRef string) error {
	teamID, err := s.teamMapper.MapTeamRefToTeamID(ctx, teamRef)
	if err != nil {
		return err
	}
	if e := s.teamAPI.Delete(ctx, teamID); e != nil {
		return mapError(e)
	}
	return nil
}

// RestoreDeleted will be used later
func (s *Service) RestoreDeleted(ctx context.Context, deletedGroupID string) (string, error) {
	obj, err := s.teamAPI.RestoreDeleted(ctx, deletedGroupID)
	if err != nil {
		return "", mapError(err)
	}
	if obj == nil || obj.GetId() == nil {
		return "", ErrUnknown
	}
	return *obj.GetId(), nil
}

func mapGraphTeam(t msmodels.Teamable) *Team {
	if t == nil {
		return nil
	}
	out := &Team{
		ID:          deref(t.GetId()),
		DisplayName: deref(t.GetDisplayName()),
		Description: deref(t.GetDescription()),
	}
	if v := t.GetVisibility(); v != nil {
		out.Visibility = v.String()
	}
	if t.GetIsArchived() != nil {
		out.IsArchived = *t.GetIsArchived()
	}
	return out
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
