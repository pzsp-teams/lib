package teams

import (
	"context"
	"fmt"
	"net/http"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/internal/adapter"
	"github.com/pzsp-teams/lib/internal/api"
	"github.com/pzsp-teams/lib/internal/resolver"
	"github.com/pzsp-teams/lib/internal/resources"
	snd "github.com/pzsp-teams/lib/internal/sender"
	"github.com/pzsp-teams/lib/internal/util"
	"github.com/pzsp-teams/lib/models"
)

// Service will be used later
type service struct {
	teamAPI      api.TeamAPI
	teamResolver resolver.TeamResolver
}

// NewService will be used later
func NewService(teamsAPI api.TeamAPI, tr resolver.TeamResolver) Service {
	return &service{teamAPI: teamsAPI, teamResolver: tr}
}

// Get will be used later
func (s *service) Get(ctx context.Context, teamRef string) (*models.Team, error) {
	teamID, err := s.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return nil, err
	}

	resp, requestErr := s.teamAPI.Get(ctx, teamID)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(resources.Team, teamRef))
	}

	return adapter.MapGraphTeam(resp), nil
}

// ListMyJoined will be used later
func (s *service) ListMyJoined(ctx context.Context) ([]*models.Team, error) {
	resp, requestErr := s.teamAPI.ListMyJoined(ctx)
	if requestErr != nil {
		return nil, snd.MapError(requestErr)
	}
	return util.MapSlices(resp.GetValue(), adapter.MapGraphTeam), nil
}

// Update will be used later
func (s *service) Update(ctx context.Context, teamRef string, patch *msmodels.Team) (*models.Team, error) {
	teamID, err := s.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return nil, err
	}

	resp, requestErr := s.teamAPI.Update(ctx, teamID, patch)
	if requestErr != nil {
		return nil, snd.MapError(requestErr, snd.WithResource(resources.Team, teamID))
	}

	return adapter.MapGraphTeam(resp), nil
}

// CreateViaGroup will be used later
func (s *service) CreateViaGroup(ctx context.Context, displayName, mailNickname, visibility string) (*models.Team, error) {
	id, requestErr := s.teamAPI.CreateViaGroup(ctx, displayName, mailNickname, visibility)
	if requestErr != nil {
		return nil, snd.MapError(requestErr)
	}

	t, ge := s.teamAPI.Get(ctx, id)
	if ge != nil {
		return nil, snd.MapError(ge)
	}

	return adapter.MapGraphTeam(t), nil
}

// CreateFromTemplate will be used later
func (s *service) CreateFromTemplate(ctx context.Context, displayName, description string, owners []string) (string, error) {
	id, requestErr := s.teamAPI.CreateFromTemplate(ctx, displayName, description, owners)
	if requestErr != nil {
		if requestErr.Code == http.StatusCreated {
			return id, nil
		}
		return "", snd.MapError(requestErr, snd.WithResources(resources.User, owners))
	}

	return id, nil
}

// Archive will be used later
func (s *service) Archive(ctx context.Context, teamRef string, spoReadOnlyForMembers *bool) error {
	teamID, err := s.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return err
	}

	if e := s.teamAPI.Archive(ctx, teamID, spoReadOnlyForMembers); e != nil {
		return snd.MapError(e, snd.WithResource(resources.Team, teamRef))
	}

	return nil
}

// Unarchive will be used later
func (s *service) Unarchive(ctx context.Context, teamRef string) error {
	teamID, err := s.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return err
	}

	if requestErr := s.teamAPI.Unarchive(ctx, teamID); requestErr != nil {
		return snd.MapError(requestErr, snd.WithResource(resources.Team, teamRef))
	}

	return nil
}

// Delete will be used later
func (s *service) Delete(ctx context.Context, teamRef string) error {
	teamID, err := s.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return err
	}

	if requestErr := s.teamAPI.Delete(ctx, teamID); requestErr != nil {
		return snd.MapError(requestErr, snd.WithResource(resources.Team, teamRef))
	}

	return nil
}

// RestoreDeleted will be used later
func (s *service) RestoreDeleted(ctx context.Context, deletedGroupID string) (string, error) {
	obj, err := s.teamAPI.RestoreDeleted(ctx, deletedGroupID)
	if err != nil {
		return "", snd.MapError(err, snd.WithResource(resources.Team, deletedGroupID))
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
