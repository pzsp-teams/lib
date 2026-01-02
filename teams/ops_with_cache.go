package teams

import (
	"context"
	"strings"

	"github.com/pzsp-teams/lib/internal/cacher"
	snd "github.com/pzsp-teams/lib/internal/sender"
	"github.com/pzsp-teams/lib/internal/util"
	"github.com/pzsp-teams/lib/models"
)

type opsWithCache struct {
	teamOps      teamsOps
	cacheHandler *cacher.CacheHandler
}

func NewOpsWithCache(teamOps teamsOps, cache *cacher.CacheHandler) teamsOps {
	if cache == nil {
		return teamOps
	}
	return &opsWithCache{
		teamOps:      teamOps,
		cacheHandler: cache,
	}
}

func (o *opsWithCache) Wait() {
	o.cacheHandler.Runner.Wait()
}

func (o *opsWithCache) GetTeamByID(ctx context.Context, teamID string) (*models.Team, *snd.RequestError) {
	team, requestErr := o.teamOps.GetTeamByID(ctx, teamID)
	if requestErr != nil {
		o.cacheHandler.OnError(requestErr)
		return nil, requestErr
	}
	if team != nil {
		local := *team
		o.cacheHandler.Runner.Run(func() {
			o.addTeamsToCache(local)
		})
	}
	return team, nil
}

func (o *opsWithCache) ListMyJoinedTeams(ctx context.Context) ([]*models.Team, *snd.RequestError) {
	out, requestErr := o.teamOps.ListMyJoinedTeams(ctx)
	if requestErr != nil {
		o.cacheHandler.OnError(requestErr)
		return nil, requestErr
	}
	local := util.CopyNonNil(out)
	o.cacheHandler.Runner.Run(func() {
		o.addTeamsToCache(local...)
	})
	return out, nil
}

func (o *opsWithCache) CreateFromTemplate(ctx context.Context, displayName, description string, owners []string) (string, *snd.RequestError) {
	id, requestErr := o.teamOps.CreateFromTemplate(ctx, displayName, description, owners)
	if requestErr != nil {
		o.cacheHandler.OnError(requestErr)
		return id, requestErr
	}
	o.cacheHandler.Runner.Run(func() {
		o.removeTeamFromCache(displayName)
	})
	return id, nil
}

func (o *opsWithCache) CreateViaGroup(ctx context.Context, displayName, mailNickname, visibility string) (*models.Team, *snd.RequestError) {
	team, requestErr := o.teamOps.CreateViaGroup(ctx, displayName, mailNickname, visibility)
	if requestErr != nil {
		o.cacheHandler.OnError(requestErr)
		return nil, requestErr
	}
	if team != nil {
		local := *team
		o.cacheHandler.Runner.Run(func() {
			o.addTeamsToCache(local)
		})
	}
	return team, nil
}

func (o *opsWithCache) Archive(ctx context.Context, teamID, teamRef string, spoReadOnlyForMembers *bool) *snd.RequestError {
	requestErr := o.teamOps.Archive(ctx, teamID, teamRef,spoReadOnlyForMembers)
	if requestErr != nil {
		o.cacheHandler.OnError(requestErr)
		return requestErr
	}
	o.cacheHandler.Runner.Run(func() {
		o.removeTeamFromCache(teamRef)
	})
	return nil
}

func (o *opsWithCache) Unarchive(ctx context.Context, teamID string) *snd.RequestError {
	requestErr := o.teamOps.Unarchive(ctx, teamID)
	if requestErr != nil {
		o.cacheHandler.OnError(requestErr)
		return requestErr
	}
	return nil
}

func (o *opsWithCache) DeleteTeam(ctx context.Context, teamID, teamRef string) *snd.RequestError {
	requestErr := o.teamOps.DeleteTeam(ctx, teamID, teamRef)
	if requestErr != nil {
		o.cacheHandler.OnError(requestErr)
		return requestErr
	}
	o.cacheHandler.Runner.Run(func() {
		o.removeTeamFromCache(teamRef)
	})
	return nil
}

func (o *opsWithCache) RestoreDeletedTeam(ctx context.Context, deletedGroupID string) (string, *snd.RequestError) {
	return cacher.WithErrorClear(func() (string, *snd.RequestError) {
		return o.teamOps.RestoreDeletedTeam(ctx, deletedGroupID)
	}, o.cacheHandler)
}

func (o *opsWithCache) ListMembers(ctx context.Context, teamID string) ([]*models.Member, *snd.RequestError) {
	members, requestErr := o.teamOps.ListMembers(ctx, teamID)
	if requestErr != nil {
		o.cacheHandler.OnError(requestErr)
		return nil, requestErr
	}
	local := util.CopyNonNil(members)
	o.cacheHandler.Runner.Run(func() {
		o.addMembersToCache(teamID, local...)
	})
	return members, nil
}

func (o *opsWithCache) GetMemberByID(ctx context.Context, teamID, memberID string) (*models.Member, *snd.RequestError) {
	member, requestErr := o.teamOps.GetMemberByID(ctx, teamID, memberID)
	if requestErr != nil {
		o.cacheHandler.OnError(requestErr)
		return nil, requestErr
	}
	if member != nil {
		local := *member
		o.cacheHandler.Runner.Run(func() {
			o.addMembersToCache(teamID, local)
		})
	}
	return member, nil
}

func (o *opsWithCache) AddMember(ctx context.Context, teamID, userID string, isOwner bool) (*models.Member, *snd.RequestError) {
	member, requestErr := o.teamOps.AddMember(ctx, teamID, userID, isOwner)
	if requestErr != nil {
		o.cacheHandler.OnError(requestErr)
		return nil, requestErr
	}
	if member != nil {
		local := *member
		o.cacheHandler.Runner.Run(func() {
			o.addMembersToCache(teamID, local)
		})
	}
	return member, nil
}

func (o *opsWithCache) UpdateMemberRoles(ctx context.Context, teamID, memberID string, isOwner bool) (*models.Member, *snd.RequestError) {
	return cacher.WithErrorClear(func() (*models.Member, *snd.RequestError) {
		return o.teamOps.UpdateMemberRoles(ctx, teamID, memberID, isOwner)
	}, o.cacheHandler)
}

func (o *opsWithCache) RemoveMember(ctx context.Context, teamID, memberID, userRef string) *snd.RequestError {
	requestErr := o.teamOps.RemoveMember(ctx, teamID, memberID, userRef)
	if requestErr != nil {
		o.cacheHandler.OnError(requestErr)
		return requestErr
	}
	o.cacheHandler.Runner.Run(func() {
		o.removeMemberFromCache(teamID, userRef)
	})
	return nil
}

func (o *opsWithCache) addTeamsToCache(teams ...models.Team) {
	for _, team := range teams {
		name := strings.TrimSpace(team.DisplayName)
		if name == "" {
			continue
		}
		key := cacher.NewTeamKey(name)
		_ = o.cacheHandler.Cacher.Set(key, team.ID)
	}
}

func (o *opsWithCache) removeTeamFromCache(teamRef string) {
	teamRef = strings.TrimSpace(teamRef)
	if teamRef == "" {
		return
	}
	key := cacher.NewTeamKey(teamRef)
	_ = o.cacheHandler.Cacher.Invalidate(key)
}

func (o *opsWithCache) addMembersToCache(teamID string, members ...models.Member) {
	teamID = strings.TrimSpace(teamID)
	if teamID == "" {
		return
	}
	for _, member := range members {
		email := strings.TrimSpace(member.Email)
		if email == "" {
			continue
		}
		key := cacher.NewTeamMemberKey(teamID, email, nil)
		_ = o.cacheHandler.Cacher.Set(key, member.ID)
	}
}

func (o *opsWithCache) removeMemberFromCache(teamID, userRef string) {
	key := cacher.NewTeamMemberKey(teamID, userRef, nil)
	_ = o.cacheHandler.Cacher.Invalidate(key)
}

