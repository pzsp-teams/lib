package resolver

import (
	"context"
	"fmt"
	"strings"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/cacher"
	"github.com/pzsp-teams/lib/internal/api"
	"github.com/pzsp-teams/lib/internal/sender"
	"github.com/pzsp-teams/lib/internal/util"
)

type ChatResolver interface {
	ResolveOneOnOneChatRefToID(ctx context.Context, userRef string) (string, error)
	ResolveGroupChatRefToID(ctx context.Context, topic string) (string, error)
	ResolveChatMemberRefToID(ctx context.Context, chatID, userRef string) (string, error)
}

type ChatResolverCacheable struct {
	chatsAPI     api.ChatAPI
	cacher       cacher.Cacher
	cacheEnabled bool
}

func NewChatResolverCacheable(chatsAPI api.ChatAPI, cacher cacher.Cacher, cacheEnabled bool) ChatResolver {
	return &ChatResolverCacheable{
		chatsAPI:     chatsAPI,
		cacher:       cacher,
		cacheEnabled: cacheEnabled,
	}
}

func (m *ChatResolverCacheable) ResolveOneOnOneChatRefToID(ctx context.Context, userRef string) (string, error) {
	rCtx := m.newOneOnOneResolveContext(userRef)
	return rCtx.resolveWithCache(ctx, m.cacher, m.cacheEnabled)
}
func (m *ChatResolverCacheable) ResolveChatMemberRefToID(ctx context.Context, chatID, userRef string) (string, error) {
	rCtx := m.newChatMemberResolveContext(chatID, userRef)
	return rCtx.resolveWithCache(ctx, m.cacher, m.cacheEnabled)
}

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
			return m.chatsAPI.ListChats(ctx, "oneOnOne")
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
			return m.chatsAPI.ListChats(ctx, "group")
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

func resolveOneOnOneChatIDByUserRef(chats msmodels.ChatCollectionResponseable, userRef string) (string, error) {
	if chats == nil || chats.GetValue() == nil || len(chats.GetValue()) == 0 {
		return "", fmt.Errorf("no one-on-one chats avaliable")
	}

	for _, chat := range chats.GetValue() {
		if chat == nil {
			continue
		}
		members := chat.GetMembers()
		for _, member := range members {
			um, ok := member.(msmodels.AadUserConversationMemberable)
			if !ok {
				continue
			}
			if matchesUserRef(um, userRef) {
				return util.Deref(chat.GetId()), nil
			}
		}
	}
	return "", fmt.Errorf("chat with given user %q not found", userRef)
}

func resolveGroupChatIDByTopic(chats msmodels.ChatCollectionResponseable, topic string) (string, error) {
	if chats == nil || chats.GetValue() == nil || len(chats.GetValue()) == 0 {
		return "", fmt.Errorf("no group chats avaliable")
	}

	matches := make([]msmodels.Chatable, 0, len(chats.GetValue()))
	for _, chat := range chats.GetValue() {
		if chat == nil {
			continue
		}
		if util.Deref(chat.GetTopic()) == topic {
			matches = append(matches, chat)
		}
	}
	switch len(matches) {
	case 0:
		return "", fmt.Errorf("chat with given topic %q not found", topic)
	case 1:
		return util.Deref(matches[0].GetId()), nil
	default:
		var options []string
		for _, c := range matches {
			options = append(options,
				fmt.Sprintf("%s (ID: %s)", util.Deref(c.GetTopic()), util.Deref(c.GetId())))
		}
		return "", fmt.Errorf("multiple chats with given topic %q found: \n%s. \nPlease use one of the IDs to resolve the chat",
			topic, strings.Join(options, ";\n"))
	}
}
