package teams

import (
	"context"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
)

// Service will be used later
type Service struct {
	api APIInterface
}

// NewService will be used later
func NewService(api APIInterface) *Service {
	return &Service{api: api}
}

// Get will be used later
func (s *Service) Get(ctx context.Context, teamID string) (*Team, error) {
	resp, err := s.api.Get(ctx, teamID)
	if err != nil {
		return nil, mapError(err)
	}
	return mapGraphTeam(resp), nil
}

// ListMyJoined will be used later
func (s *Service) ListMyJoined(ctx context.Context) ([]*Team, error) {
	resp, err := s.api.ListMyJoined(ctx)
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
func (s *Service) Update(ctx context.Context, teamID string, patch *msmodels.Team) (*Team, error) {
	resp, err := s.api.Update(ctx, teamID, patch)
	if err != nil {
		return nil, mapError(err)
	}
	return mapGraphTeam(resp), nil
}

// CreateViaGroup will be used later
func (s *Service) CreateViaGroup(ctx context.Context, displayName, mailNickname, visibility string) (*Team, error) {
	id, err := s.api.CreateViaGroup(ctx, displayName, mailNickname, visibility)
	if err != nil {
		return nil, mapError(err)
	}
	t, ge := s.api.Get(ctx, id)
	if ge != nil {
		return nil, mapError(ge)
	}
	return mapGraphTeam(t), nil
}

// CreateFromTemplate will be used later
func (s *Service) CreateFromTemplate(ctx context.Context, displayName, description string, owners []string) (string, error) {
	id, err := s.api.CreateFromTemplate(ctx, displayName, description, owners)
	if err != nil {
		return "", mapError(err)
	}
	return id, nil
}

// Archive will be used later
func (s *Service) Archive(ctx context.Context, teamID string, spoReadOnlyForMembers *bool) error {
	if e := s.api.Archive(ctx, teamID, spoReadOnlyForMembers); e != nil {
		return mapError(e)
	}
	return nil
}

// Unarchive will be used later
func (s *Service) Unarchive(ctx context.Context, teamID string) error {
	if e := s.api.Unarchive(ctx, teamID); e != nil {
		return mapError(e)
	}
	return nil
}

// Delete will be used later
func (s *Service) Delete(ctx context.Context, teamID string) error {
	if e := s.api.Delete(ctx, teamID); e != nil {
		return mapError(e)
	}
	return nil
}

// RestoreDeleted will be used later
func (s *Service) RestoreDeleted(ctx context.Context, deletedGroupID string) (string, error) {
	obj, err := s.api.RestoreDeleted(ctx, deletedGroupID)
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
