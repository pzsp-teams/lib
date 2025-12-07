package teams

import (
	"context"
	"regexp"
	"strings"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/cacher"
)

type ServiceWithAutoCacheManagement struct {
	svc   *Service
	cache cacher.Cacher
}

func NewServiceWithAutoCacheManagement(svc *Service, cache cacher.Cacher) *ServiceWithAutoCacheManagement {
	return &ServiceWithAutoCacheManagement{
		svc:   svc,
		cache: cache,
	}
}

func (s *ServiceWithAutoCacheManagement) Get(ctx context.Context, teamRef string) (*Team, error) {
	team, err := s.svc.Get(ctx, teamRef)
	if err != nil {
		return nil, err
	}
	s.addTeamToCache(team)
	return team, nil
}

func (s *ServiceWithAutoCacheManagement) ListMyJoined(ctx context.Context) ([]*Team, error) {
	teams, err := s.svc.ListMyJoined(ctx)
	if err != nil {
		return nil, err
	}
	for _, team := range teams {
		s.addTeamToCache(team)
	}
	return teams, nil
}

func (s *ServiceWithAutoCacheManagement) Update(ctx context.Context, teamRef string, patch *msmodels.Team) (*Team, error) {
	team, err := s.svc.Update(ctx, teamRef, patch)
	if err != nil {
		return nil, err
	}
	s.removeTeamFromCache(teamRef)
	s.addTeamToCache(team)
	return team, nil
}

func (s *ServiceWithAutoCacheManagement) CreateFromTemplate(ctx context.Context, displayName, description string, owners []string) (string, error) {
	id, err := s.svc.CreateFromTemplate(ctx, displayName, description, owners)
	if err != nil {
		return id, err
	}
	s.removeTeamFromCache(displayName)
	return id, err
}

func (s *ServiceWithAutoCacheManagement) CreateViaGroup(ctx context.Context, displayName, mailNickname, visibility string) (*Team, error) {
	team, err := s.svc.CreateViaGroup(ctx, displayName, mailNickname, visibility)
	if err != nil {
		return nil, err
	}
	s.removeTeamFromCache(displayName)
	return team, nil
}

func (s *ServiceWithAutoCacheManagement) Archive(ctx context.Context, teamRef string, spoReadOnlyForMembers *bool) error {
	err := s.svc.Archive(ctx, teamRef, spoReadOnlyForMembers)
	if err != nil {
		return err
	}
	s.removeTeamFromCache(teamRef)
	return nil
}

func (s *ServiceWithAutoCacheManagement) Unarchive(ctx context.Context, teamRef string) error {
	err := s.svc.Unarchive(ctx, teamRef)
	if err != nil {
		return err
	}
	s.removeTeamFromCache(teamRef)
	return nil
}

func (s *ServiceWithAutoCacheManagement) Delete(ctx context.Context, teamRef string) error {
	err := s.svc.Delete(ctx, teamRef)
	if err != nil {
		return err
	}
	s.removeTeamFromCache(teamRef)
	return nil
}

func (s *ServiceWithAutoCacheManagement) RestoreDeleted(ctx context.Context, deletedGroupID string) (string, error) {
	return s.svc.RestoreDeleted(ctx, deletedGroupID)
}

func (s *ServiceWithAutoCacheManagement) addTeamToCache(team *Team) {
	keyBuilder := cacher.NewTeamKeyBuilder(strings.TrimSpace(team.DisplayName))
	_ = s.cache.Set(keyBuilder.ToString(), team.ID)
}

func (s *ServiceWithAutoCacheManagement) removeTeamFromCache(teamRef string) {
	if isLikelyGUID(teamRef) {
		return
	}
	keyBuilder := cacher.NewTeamKeyBuilder(strings.TrimSpace(teamRef))
	_ = s.cache.Invalidate(keyBuilder.ToString())
}

func isLikelyGUID(s string) bool {
	var guidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	return guidRegex.MatchString(s)
}
