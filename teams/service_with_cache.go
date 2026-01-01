package teams

import (
	"context"
	"strings"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/internal/cacher"
	"github.com/pzsp-teams/lib/internal/resolver"
	"github.com/pzsp-teams/lib/internal/util"
	"github.com/pzsp-teams/lib/models"
)

type serviceWithCache struct {
	svc          Service
	teamResolver resolver.TeamResolver
	cacheHandler *cacher.CacheHandler
}

func NewServiceWithCache(svc Service, cacheHandler *cacher.CacheHandler, teamResolver resolver.TeamResolver) Service {
	if cacheHandler == nil {
		return svc
	}
	return &serviceWithCache{svc, teamResolver, cacheHandler}
}

func (s *serviceWithCache) Wait() {
	s.cacheHandler.Runner.Wait()
}

func (s *serviceWithCache) Get(ctx context.Context, teamRef string) (*models.Team, error) {
	team, err := s.svc.Get(ctx, teamRef)
	if err != nil {
		s.onError()
		return nil, err
	}
	s.cacheHandler.Runner.Run(func() {
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
	s.cacheHandler.Runner.Run(func() {
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
	s.cacheHandler.Runner.Run(func() {
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
	s.cacheHandler.Runner.Run(func() {
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
	s.cacheHandler.Runner.Run(func() {
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
	s.cacheHandler.Runner.Run(func() {
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
	s.cacheHandler.Runner.Run(func() {
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
	s.cacheHandler.Runner.Run(func() {
		s.removeTeamsFromCache(teamRef)
	})
	return nil
}

func (s *serviceWithCache) RestoreDeleted(ctx context.Context, deletedGroupID string) (string, error) {
	return withErrorClear(func() (string, error) {
		return s.svc.RestoreDeleted(ctx, deletedGroupID)
	}, s)
}

func (s *serviceWithCache) ListMembers(ctx context.Context, teamRef string) ([]*models.Member, error) {
	members, err := s.svc.ListMembers(ctx, teamRef)
	if err != nil {
		s.onError()
		return nil, err
	}
	local := util.CopyNonNil(members)
	s.cacheHandler.Runner.Run(func() {
		s.addMembersToCache(teamRef, local...)
	})
	return members, nil
}

func (s *serviceWithCache) GetMember(ctx context.Context, teamRef, userRef string) (*models.Member, error) {
	member, err := s.svc.GetMember(ctx, teamRef, userRef)
	if err != nil {
		s.onError()
		return nil, err
	}
	if member != nil {
		local := *member
		s.cacheHandler.Runner.Run(func() {
			s.addMembersToCache(teamRef, local)
		})
	}
	return member, nil
}

func (s *serviceWithCache) AddMember(ctx context.Context, teamRef, userRef string, isOwner bool) (*models.Member, error) {
	member, err := s.svc.AddMember(ctx, teamRef, userRef, isOwner)
	if err != nil {
		s.onError()
		return nil, err
	}
	if member != nil {
		local := *member
		s.cacheHandler.Runner.Run(func() {
			s.addMembersToCache(teamRef, local)
		})
	}
	return member, nil
}

func (s *serviceWithCache) UpdateMemberRoles(ctx context.Context, teamRef, userRef string, isOwner bool) (*models.Member, error) {
	return withErrorClear(func() (*models.Member, error) {
		return s.svc.UpdateMemberRoles(ctx, teamRef, userRef, isOwner)
	}, s)
}

func (s *serviceWithCache) RemoveMember(ctx context.Context, teamRef, userRef string) error {
	if err := s.svc.RemoveMember(ctx, teamRef, userRef); err != nil {
		s.onError()
		return err
	}
	s.cacheHandler.Runner.Run(func() {
		s.invalidateMemberCache(teamRef, userRef)
	})
	return nil
}

func (s *serviceWithCache) addTeamsToCache(teams ...models.Team) {
	for _, team := range teams {
		key := cacher.NewTeamKey(strings.TrimSpace(team.DisplayName))
		_ = s.cacheHandler.Cacher.Set(key, team.ID)
	}
}

func (s *serviceWithCache) removeTeamsFromCache(teamRefs ...string) {
	for _, teamRef := range teamRefs {
		if util.IsLikelyGUID(teamRef) {
			continue
		}
		key := cacher.NewTeamKey(strings.TrimSpace(teamRef))
		_ = s.cacheHandler.Cacher.Invalidate(key)
	}
}

func (s *serviceWithCache) onError() {
	s.cacheHandler.Runner.Run(func() {
		_ = s.cacheHandler.Cacher.Clear()
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

func (s *serviceWithCache) addMembersToCache(teamRef string, members ...models.Member) {
	for _, member := range members {
		ref := strings.TrimSpace(strings.TrimSpace(member.Email))
		if ref == "" {
			continue
		}
		ctx := context.Background()
		teamID, err := s.teamResolver.ResolveTeamRefToID(ctx, teamRef)
		if err != nil {
			continue
		}
		key := cacher.NewTeamMemberKey(teamID, ref, nil)
		_ = s.cacheHandler.Cacher.Set(key, member.ID)
	}
}

func (s *serviceWithCache) invalidateMemberCache(teamRef, userRef string) {
	ref := strings.TrimSpace(userRef)
	if ref == "" {
		return
	}

	ctx := context.Background()
	teamID, err := s.teamResolver.ResolveTeamRefToID(ctx, teamRef)
	if err != nil {
		return
	}

	key := cacher.NewTeamMemberKey(teamID, ref, nil)
	_ = s.cacheHandler.Cacher.Invalidate(key)
}
