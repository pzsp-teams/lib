package teams

import (
	"context"
	"strings"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/cacher"
	"github.com/pzsp-teams/lib/internal/util"
	"github.com/pzsp-teams/lib/models"
)

type ServiceWithAutoCacheManagement struct {
	svc   *Service
	cache cacher.Cacher
	run   func(func())
}

func NewServiceWithAutoCacheManagement(svc *Service, cache cacher.Cacher) *ServiceWithAutoCacheManagement {
	return &ServiceWithAutoCacheManagement{
		svc:   svc,
		cache: cache,
		run:   func(fn func()) { go fn() },
	}
}

func (s *ServiceWithAutoCacheManagement) Get(ctx context.Context, teamRef string) (*models.Team, error) {
	team, err := s.svc.Get(ctx, teamRef)
	if err != nil {
		return nil, err
	}
	s.run(func() {
		s.addTeamsToCache(*team)
	})
	return team, nil
}

func (s *ServiceWithAutoCacheManagement) ListMyJoined(ctx context.Context) ([]*models.Team, error) {
	teams, err := s.svc.ListMyJoined(ctx)
	if err != nil {
		return nil, err
	}
	vals := make([]models.Team, 0, len(teams))
	for _, t := range teams {
		if t != nil {
			vals = append(vals, *t)
		}
	}
	s.run(func() {
		s.addTeamsToCache(vals...)
	})
	return teams, nil
}

func (s *ServiceWithAutoCacheManagement) Update(ctx context.Context, teamRef string, patch *msmodels.Team) (*models.Team, error) {
	team, err := s.svc.Update(ctx, teamRef, patch)
	if err != nil {
		return nil, err
	}
	s.run(func() {
		s.removeTeamsFromCache(teamRef)
		s.addTeamsToCache(*team)
	})
	return team, nil
}

func (s *ServiceWithAutoCacheManagement) CreateFromTemplate(ctx context.Context, displayName, description string, owners []string) (string, error) {
	id, err := s.svc.CreateFromTemplate(ctx, displayName, description, owners)
	if err != nil {
		return id, err
	}
	s.run(func() {
		s.removeTeamsFromCache(displayName)
	})
	return id, err
}

func (s *ServiceWithAutoCacheManagement) CreateViaGroup(ctx context.Context, displayName, mailNickname, visibility string) (*models.Team, error) {
	team, err := s.svc.CreateViaGroup(ctx, displayName, mailNickname, visibility)
	if err != nil {
		return nil, err
	}
	s.run(func() {
		s.removeTeamsFromCache(displayName)
	})
	return team, nil
}

func (s *ServiceWithAutoCacheManagement) Archive(ctx context.Context, teamRef string, spoReadOnlyForMembers *bool) error {
	err := s.svc.Archive(ctx, teamRef, spoReadOnlyForMembers)
	if err != nil {
		return err
	}
	s.run(func() {
		s.removeTeamsFromCache(teamRef)
	})
	return nil
}

func (s *ServiceWithAutoCacheManagement) Unarchive(ctx context.Context, teamRef string) error {
	err := s.svc.Unarchive(ctx, teamRef)
	if err != nil {
		return err
	}
	s.run(func() {
		s.removeTeamsFromCache(teamRef)
	})
	return nil
}

func (s *ServiceWithAutoCacheManagement) Delete(ctx context.Context, teamRef string) error {
	err := s.svc.Delete(ctx, teamRef)
	if err != nil {
		return err
	}
	s.run(func() {
		s.removeTeamsFromCache(teamRef)
	})
	return nil
}

func (s *ServiceWithAutoCacheManagement) RestoreDeleted(ctx context.Context, deletedGroupID string) (string, error) {
	return s.svc.RestoreDeleted(ctx, deletedGroupID)
}

func (s *ServiceWithAutoCacheManagement) addTeamsToCache(teams ...models.Team) {
	if s.cache == nil || teams == nil {
		return
	}
	for _, team := range teams {
		keyBuilder := cacher.NewTeamKeyBuilder(strings.TrimSpace(team.DisplayName))
		_ = s.cache.Set(keyBuilder.ToString(), team.ID)
	}
}

func (s *ServiceWithAutoCacheManagement) removeTeamsFromCache(teamRefs ...string) {
	if s.cache == nil {
		return
	}
	for _, teamRef := range teamRefs {
		if util.IsLikelyGUID(teamRef) {
			continue
		}
		keyBuilder := cacher.NewTeamKeyBuilder(strings.TrimSpace(teamRef))
		_ = s.cache.Invalidate(keyBuilder.ToString())
	}
}
