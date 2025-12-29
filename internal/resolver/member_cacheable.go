package resolver

import (
	"context"
	"fmt"
	"strings"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/cacher"
	"github.com/pzsp-teams/lib/internal/api"
	"github.com/pzsp-teams/lib/internal/util"
)

type MemberResolver interface {
	ResolveUserRefToMemberID(ctx context.Context, memberCtx *MemberContext) (string, error)
	NewGroupChatMemberContext(chatID string, userRef string) *MemberContext
	NewChannelMemberContext(teamID, channelID, userRef string) *MemberContext
}

type MemberContext struct {
	cacheKey         string
	userRef          string
	containerID      string
	containerName    string
	fetchMembersFunc func(ctx context.Context) (msmodels.ConversationMemberCollectionResponseable, error)
}

type MemberResolverCacheable struct {
	channelsAPI  api.ChannelAPI
	chatsAPI     api.ChatAPI
	cacher       cacher.Cacher
	cacheEnabled bool
}

func NewMemberResolverCacheable(channelsAPI api.ChannelAPI, chatsAPI api.ChatAPI, cacher cacher.Cacher, cacheEnabled bool) *MemberResolverCacheable {
	return &MemberResolverCacheable{
		channelsAPI:  channelsAPI,
		chatsAPI:     chatsAPI,
		cacher:       cacher,
		cacheEnabled: cacheEnabled,
	}
}

func (res *MemberResolverCacheable) ResolveUserRefToMemberID(ctx context.Context, memberCtx *MemberContext) (string, error) {
	if memberCtx.userRef == "" {
		return "", fmt.Errorf("empty user reference")
	}
	if res.cacheEnabled {
		value, found, err := res.cacher.Get(memberCtx.cacheKey)
		if err == nil && found {
			if ids, ok := value.([]string); ok && len(ids) == 1 {
				return ids[0], nil
			}
		}
	}
	resp, apiErr := memberCtx.fetchMembersFunc(ctx)
	if apiErr != nil {
		return "", apiErr
	}
	if resp == nil || resp.GetValue() == nil || len(resp.GetValue()) == 0 {
		return "", fmt.Errorf("no members found in %s %q", memberCtx.containerName, memberCtx.containerID)
	}
	id := findMemberID(resp.GetValue(), memberCtx.userRef)
	if id == "" {
		return "", fmt.Errorf("user %q not found in %s %q", memberCtx.userRef, memberCtx.containerName, memberCtx.containerID)
	}
	if res.cacheEnabled {
		_ = res.cacher.Set(memberCtx.cacheKey, id)
	}
	return id, nil
}

func (res *MemberResolverCacheable) NewGroupChatMemberContext(chatID string, userRef string) *MemberContext {
	return &MemberContext{
		cacheKey:      cacher.NewGroupChatMemberKey(chatID, userRef, nil),
		userRef:       strings.TrimSpace(userRef),
		containerID:   chatID,
		containerName: "group-chat",
		fetchMembersFunc: func(ctx context.Context) (msmodels.ConversationMemberCollectionResponseable, error) {
			return res.chatsAPI.ListGroupChatMembers(ctx, chatID)
		},
	}
}

func (res *MemberResolverCacheable) NewChannelMemberContext(teamID, channelID, userRef string) *MemberContext {
	return &MemberContext{
		cacheKey:      cacher.NewChannelMemberKey(teamID, channelID, userRef, nil),
		userRef:       strings.TrimSpace(userRef),
		containerID:   channelID,
		containerName: "channel",
		fetchMembersFunc: func(ctx context.Context) (msmodels.ConversationMemberCollectionResponseable, error) {
			return res.channelsAPI.ListMembers(ctx, teamID, channelID)
		},
	}
}

func findMemberID(members []msmodels.ConversationMemberable, ref string) string {
	for _, member := range members {
		if member == nil {
			continue
		}
		um, ok := member.(msmodels.AadUserConversationMemberable)
		if !ok {
			continue
		}
		if matchesUserRef(um, ref) {
			return util.Deref(member.GetId())
		}
	}
	return ""
}

func matchesUserRef(um msmodels.AadUserConversationMemberable, userRef string) bool {
	if userRef == "" {
		return false
	}
	if util.Deref(um.GetUserId()) == userRef {
		return true
	}
	if util.Deref(um.GetDisplayName()) == userRef {
		return true
	}
	email, err := um.GetBackingStore().Get("email")
	if err == nil {
		if emailStr, ok := email.(*string); ok {
			if util.Deref(emailStr) == userRef {
				return true
			}
		}
	}
	return false
}
