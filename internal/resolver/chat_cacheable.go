package resolver

import (
	"context"
	"strings"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/cacher"
	"github.com/pzsp-teams/lib/internal/api"
	"github.com/pzsp-teams/lib/internal/sender"
	"github.com/pzsp-teams/lib/internal/util"
)

// ChatResolver defines methods to resolve chat and chat-member references
// into their corresponding Microsoft Graph IDs.
//
// It supports resolving both one-on-one chats and group chats.
type ChatResolver interface {
	// ResolveOneOnOneChatRefToID resolves a one-on-one chat reference (user email or ID)
	// to a chat ID.
	//
	// If the reference already appears to be an one-on-one chat ID,
	// it may be returned directly.
	ResolveOneOnOneChatRefToID(ctx context.Context, userRef string) (string, error)

	// ResolveGroupChatRefToID resolves a group chat reference (topic or ID)
	// to a chat ID.
	//
	// If the reference already appears to be an group chat ID,
	// it may be returned directly.
	ResolveGroupChatRefToID(ctx context.Context, topic string) (string, error)

	// ResolveChatMemberRefToID resolves a user reference (email or ID)
	// to a chat member ID within the specified chat.
	ResolveChatMemberRefToID(ctx context.Context, chatID, userRef string) (string, error)
}

// ChatResolverCacheable resolves chat references using the graph API
// and optionally caches successful resolutions.
type ChatResolverCacheable struct {
	chatsAPI     api.ChatAPI
	cacher       cacher.Cacher
	cacheEnabled bool
}

// NewChatResolverCacheable creates a new ChatResolverCacheable.
func NewChatResolverCacheable(chatsAPI api.ChatAPI, c cacher.Cacher, cacheEnabled bool) ChatResolver {
	return &ChatResolverCacheable{
		chatsAPI:     chatsAPI,
		cacher:       c,
		cacheEnabled: cacheEnabled,
	}
}

// ResolveOneOnOneChatRefToID implements ChatResolver.
func (m *ChatResolverCacheable) ResolveOneOnOneChatRefToID(ctx context.Context, userRef string) (string, error) {
	rCtx := m.newOneOnOneResolveContext(userRef)
	return rCtx.resolveWithCache(ctx, m.cacher, m.cacheEnabled)
}

// ResolveChatMemberRefToID implements ChatResolver.
func (m *ChatResolverCacheable) ResolveChatMemberRefToID(ctx context.Context, chatID, userRef string) (string, error) {
	rCtx := m.newChatMemberResolveContext(chatID, userRef)
	return rCtx.resolveWithCache(ctx, m.cacher, m.cacheEnabled)
}

// ResolveGroupChatRefToID implements ChatResolver.
func (m *ChatResolverCacheable) ResolveGroupChatRefToID(ctx context.Context, chatRef string) (string, error) {
	rCtx := m.newGroupChatResolveContext(chatRef)
	return rCtx.resolveWithCache(ctx, m.cacher, m.cacheEnabled)
}

func (m *ChatResolverCacheable) newOneOnOneResolveContext(userRef string) ResolverContext[msmodels.ChatCollectionResponseable] {
	ref := strings.TrimSpace(userRef)
	return ResolverContext[msmodels.ChatCollectionResponseable]{
		cacheKey:    cacher.NewOneOnOneChatKey(ref, nil),
		keyType:     cacher.DirectChat,
		ref:         ref,
		isAlreadyID: func() bool { return util.IsLikelyGUID(ref) },
		fetch: func(ctx context.Context) (msmodels.ChatCollectionResponseable, *sender.RequestError) {
			oneOnOneChat := "oneOnOne"
			return m.chatsAPI.ListChats(ctx, &oneOnOneChat)
		},
		extract: func(data msmodels.ChatCollectionResponseable) (string, error) {
			return resolveOneOnOneChatIDByUserRef(data, ref)
		},
	}
}

func (m *ChatResolverCacheable) newGroupChatResolveContext(topic string) ResolverContext[msmodels.ChatCollectionResponseable] {
	ref := strings.TrimSpace(topic)
	return ResolverContext[msmodels.ChatCollectionResponseable]{
		cacheKey:    cacher.NewGroupChatKey(ref),
		keyType:     cacher.GroupChat,
		ref:         ref,
		isAlreadyID: func() bool { return util.IsLikelyThreadConversationID(ref) },
		fetch: func(ctx context.Context) (msmodels.ChatCollectionResponseable, *sender.RequestError) {
			groupChat := "group"
			return m.chatsAPI.ListChats(ctx, &groupChat)
		},
		extract: func(data msmodels.ChatCollectionResponseable) (string, error) {
			return resolveGroupChatIDByTopic(data, ref)
		},
	}
}

func (m *ChatResolverCacheable) newChatMemberResolveContext(chatID, userRef string) ResolverContext[msmodels.ConversationMemberCollectionResponseable] {
	ref := strings.TrimSpace(userRef)
	return ResolverContext[msmodels.ConversationMemberCollectionResponseable]{
		cacheKey:    cacher.NewGroupChatMemberKey(chatID, ref, nil),
		keyType:     cacher.GroupChatMember,
		ref:         ref,
		isAlreadyID: func() bool { return util.IsLikelyGUID(ref) },
		fetch: func(ctx context.Context) (msmodels.ConversationMemberCollectionResponseable, *sender.RequestError) {
			return m.chatsAPI.ListGroupChatMembers(ctx, chatID)
		},
		extract: func(data msmodels.ConversationMemberCollectionResponseable) (string, error) {
			return resolveMemberID(data, ref)
		},
	}
}
