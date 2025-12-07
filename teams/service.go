package teams

import (
	"context"
	"fmt"
	"net/http"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/internal/api"
	"github.com/pzsp-teams/lib/internal/resolver"
	snd "github.com/pzsp-teams/lib/internal/sender"
	"github.com/pzsp-teams/lib/internal/util"
)

// Service will be used later
type Service struct {
	teamAPI      api.TeamAPI
	teamResolver resolver.TeamResolver
}

// NewService will be used later
func NewService(teamsAPI api.TeamAPI, m resolver.TeamResolver) *Service {
	return &Service{teamAPI: teamsAPI, teamResolver: m}
}

// Get will be used later
func (s *Service) Get(ctx context.Context, teamRef string) (*Team, error) {
	teamID, err := s.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return nil, err
	}
	resp, requestErr := s.teamAPI.Get(ctx, teamID)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.Team, teamRef))
	}
	return mapGraphTeam(resp), nil
}

// ListMyJoined will be used later
func (s *Service) ListMyJoined(ctx context.Context) ([]*Team, error) {
	resp, requestErr := s.teamAPI.ListMyJoined(ctx)
	if requestErr != nil {
		return nil, snd.MapError(requestErr)
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
	teamID, err := s.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return nil, err
	}
	resp, requestErr := s.teamAPI.Update(ctx, teamID, patch)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(snd.Team, teamID))
	}
	return mapGraphTeam(resp), nil
}

// CreateViaGroup will be used later
func (s *Service) CreateViaGroup(ctx context.Context, displayName, mailNickname, visibility string) (*Team, error) {
	id, requestErr := s.teamAPI.CreateViaGroup(ctx, displayName, mailNickname, visibility)
	if requestErr != nil {
		return nil, snd.MapError(requestErr)
	}
	t, ge := s.teamAPI.Get(ctx, id)
	if ge != nil {
		return nil, snd.MapError(ge)
	}
	return mapGraphTeam(t), nil
}

// CreateFromTemplate will be used later
func (s *Service) CreateFromTemplate(ctx context.Context, displayName, description string, owners []string) (string, error) {
	id, requestErr := s.teamAPI.CreateFromTemplate(ctx, displayName, description, owners)
	if requestErr != nil {
		if requestErr.Code == http.StatusCreated {
			return id, nil
		}
		return "", snd.MapError(requestErr, snd.WithResources(snd.User, owners))
	}
	return id, nil
}

// Archive will be used later
func (s *Service) Archive(ctx context.Context, teamRef string, spoReadOnlyForMembers *bool) error {
	teamID, err := s.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return err
	}
	if e := s.teamAPI.Archive(ctx, teamID, spoReadOnlyForMembers); e != nil {
		return snd.MapError(e, snd.WithResource(snd.Team, teamRef))
	}
	return nil
}

// Unarchive will be used later
func (s *Service) Unarchive(ctx context.Context, teamRef string) error {
	teamID, err := s.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return err
	}
	if requestErr := s.teamAPI.Unarchive(ctx, teamID); requestErr != nil {
		return snd.MapError(requestErr, snd.WithResource(snd.Team, teamRef))
	}
	return nil
}

// Delete will be used later
func (s *Service) Delete(ctx context.Context, teamRef string) error {
	teamID, err := s.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return err
	}
	if requestErr := s.teamAPI.Delete(ctx, teamID); requestErr != nil {
		return snd.MapError(requestErr, snd.WithResource(snd.Team, teamRef))
	}
	return nil
}

// RestoreDeleted will be used later
func (s *Service) RestoreDeleted(ctx context.Context, deletedGroupID string) (string, error) {
	obj, err := s.teamAPI.RestoreDeleted(ctx, deletedGroupID)
	if err != nil {
		return "", snd.MapError(err, snd.WithResource(snd.Team, deletedGroupID))
	}
	if obj == nil {
		return "", fmt.Errorf("restored object is nil")
	}
	id := util.Deref((obj.GetId()))
	if id == "" {
		return "", fmt.Errorf("restored object has empty id")
	}
	return id, nil
}

func mapGraphTeam(t msmodels.Teamable) *Team {
	if t == nil {
		return nil
	}
	out := &Team{
		ID:          util.Deref(t.GetId()),
		DisplayName: util.Deref(t.GetDisplayName()),
		Description: util.Deref(t.GetDescription()),
	}
	if v := t.GetVisibility(); v != nil {
		out.Visibility = v.String()
	}
	if t.GetIsArchived() != nil {
		out.IsArchived = *t.GetIsArchived()
	}
	return out
}
