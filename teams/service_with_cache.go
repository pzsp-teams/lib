package teams

import (
	"context"
	"strings"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/internal/cacher"
	"github.com/pzsp-teams/lib/internal/util"
	"github.com/pzsp-teams/lib/models"
)

type serviceWithCache struct {
	svc    Service
	cache  cacher.Cacher
	runner util.TaskRunner
}

func NewSyncServiceWithCache(svc Service, cache cacher.Cacher) *serviceWithCache {
	return &serviceWithCache{
		svc:    svc,
		cache:  cache,
		runner: &util.SyncRunner{},
	}
}

func NewAsyncServiceWithCache(svc Service, cache cacher.Cacher) *serviceWithCache {
	return &serviceWithCache{
		svc:    svc,
		cache:  cache,
		runner: &util.AsyncRunner{},
	}
}

func (s *serviceWithCache) Wait() {
	s.runner.Wait()
}

func (s *serviceWithCache) Get(ctx context.Context, teamRef string) (*models.Team, error) {
	team, err := s.svc.Get(ctx, teamRef)
	if err != nil {
		s.onError()
		return nil, err
	}
	s.runner.Run(func() {
		s.addTeamsToCache(*team)
	})
	return team, nil
}

func (s *serviceWithCache) ListMyJoined(ctx context.Context) ([]*models.Team, error) {
	teams, err := s.svc.ListMyJoined(ctx)
	if err != nil {
		s.onError()
		return nil, err
	}
	vals := util.CopyNonNil(teams)
	s.runner.Run(func() {
		s.addTeamsToCache(vals...)
	})
	return teams, nil
}

func (s *serviceWithCache) Update(ctx context.Context, teamRef string, patch *msmodels.Team) (*models.Team, error) {
	team, err := s.svc.Update(ctx, teamRef, patch)
	if err != nil {
		s.onError()
		return nil, err
	}
	s.runner.Run(func() {
		s.removeTeamsFromCache(teamRef)
		s.addTeamsToCache(*team)
	})
	return team, nil
}

func (s *serviceWithCache) CreateFromTemplate(ctx context.Context, displayName, description string, owners []string) (string, error) {
	id, err := s.svc.CreateFromTemplate(ctx, displayName, description, owners)
	if err != nil {
		s.onError()
		return id, err
	}
	s.runner.Run(func() {
		s.removeTeamsFromCache(displayName)
	})
	return id, err
}

func (s *serviceWithCache) CreateViaGroup(ctx context.Context, displayName, mailNickname, visibility string) (*models.Team, error) {
	team, err := s.svc.CreateViaGroup(ctx, displayName, mailNickname, visibility)
	if err != nil {
		s.onError()
		return nil, err
	}
	s.runner.Run(func() {
		s.removeTeamsFromCache(displayName)
	})
	return team, nil
}

func (s *serviceWithCache) Archive(ctx context.Context, teamRef string, spoReadOnlyForMembers *bool) error {
	err := s.svc.Archive(ctx, teamRef, spoReadOnlyForMembers)
	if err != nil {
		s.onError()
		return err
	}
	s.runner.Run(func() {
		s.removeTeamsFromCache(teamRef)
	})
	return nil
}

func (s *serviceWithCache) Unarchive(ctx context.Context, teamRef string) error {
	err := s.svc.Unarchive(ctx, teamRef)
	if err != nil {
		s.onError()
		return err
	}
	s.runner.Run(func() {
		s.removeTeamsFromCache(teamRef)
	})
	return nil
}

func (s *serviceWithCache) Delete(ctx context.Context, teamRef string) error {
	err := s.svc.Delete(ctx, teamRef)
	if err != nil {
		s.onError()
		return err
	}
	s.runner.Run(func() {
		s.removeTeamsFromCache(teamRef)
	})
	return nil
}

func (s *serviceWithCache) RestoreDeleted(ctx context.Context, deletedGroupID string) (string, error) {
	return withErrorClear(func() (string, error) {
		return s.svc.RestoreDeleted(ctx, deletedGroupID)
	}, s)
}

func (s *serviceWithCache) addTeamsToCache(teams ...models.Team) {
	if s.cache == nil || teams == nil {
		return
	}
	for _, team := range teams {
		key := cacher.NewTeamKey(strings.TrimSpace(team.DisplayName))
		_ = s.cache.Set(key, team.ID)
	}
}

func (s *serviceWithCache) removeTeamsFromCache(teamRefs ...string) {
	if s.cache == nil {
		return
	}
	for _, teamRef := range teamRefs {
		if util.IsLikelyGUID(teamRef) {
			continue
		}
		key := cacher.NewTeamKey(strings.TrimSpace(teamRef))
		_ = s.cache.Invalidate(key)
	}
}

func (s *serviceWithCache) onError() {
	if s.cache == nil {
		return
	}
	s.runner.Run(func() {
		_ = s.cache.Clear()
	})
}

func withErrorClear[T any](
	fn func() (T, error), s *serviceWithCache,
) (T, error) {
	res, err := fn()
	if err != nil {
		s.onError()
		var zero T
		return zero, err
	}
	return res, nil
}
