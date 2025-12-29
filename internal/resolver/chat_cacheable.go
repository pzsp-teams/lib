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

type ChatResolver interface {
	ResolveOneOnOneChatRefToID(ctx context.Context, userRef string) (string, error)
	ResolveGroupChatRefToID(ctx context.Context, topic string) (string, error)
}

type ChatResolverCacheable struct {
	chatsAPI     api.ChatAPI
	cacher       cacher.Cacher
	cacheEnabled bool
}

func NewChatResolverCacheable(chatsAPI api.ChatAPI, c cacher.Cacher, cacheEnabled bool) ChatResolver {
	return &ChatResolverCacheable{
		chatsAPI:     chatsAPI,
		cacher:       c,
		cacheEnabled: cacheEnabled,
	}
}

func (m *ChatResolverCacheable) ResolveOneOnOneChatRefToID(ctx context.Context, userRef string) (string, error) {
	ref := strings.TrimSpace(userRef)
	if ref == "" {
		return "", fmt.Errorf("empty user ref")
	}

	if m.cacheEnabled && m.cacher != nil {
		key := cacher.NewOneOnOneChatKey(ref, nil)
		value, found, err := m.cacher.Get(key)
		if err == nil && found {
			if ids, ok := value.([]string); ok && len(ids) == 1 && ids[0] != "" {
				return ids[0], nil
			}
		}
	}

	chats, apiErr := m.chatsAPI.ListChats(ctx, "oneOnOne")
	if apiErr != nil {
		return "", apiErr
	}

	idResolved, err := m.resolveOneOnOneChatIDByUserRef(ref, chats)
	if err != nil {
		return "", err
	}

	if m.cacheEnabled && m.cacher != nil {
		key := cacher.NewOneOnOneChatKey(ref, nil)
		_ = m.cacher.Set(key, idResolved)
	}

	return idResolved, nil
}

func (m *ChatResolverCacheable) ResolveGroupChatRefToID(ctx context.Context, chatRef string) (string, error) {
	ref := strings.TrimSpace(chatRef)
	if ref == "" {
		return "", fmt.Errorf("empty group chat ref")
	}

	if m.cacheEnabled && m.cacher != nil {
		key := cacher.NewGroupChatKey(ref)
		value, found, err := m.cacher.Get(key)
		if err == nil && found {
			if ids, ok := value.([]string); ok && len(ids) == 1 && ids[0] != "" {
				return ids[0], nil
			}
		}
	}

	chats, apiErr := m.chatsAPI.ListChats(ctx, "group")
	if apiErr != nil {
		return "", apiErr
	}

	idResolved, err := m.resolveGroupChatIDByTopic(ref, chats)
	if err != nil {
		return "", err
	}

	if m.cacheEnabled && m.cacher != nil {
		key := cacher.NewGroupChatKey(ref)
		_ = m.cacher.Set(key, idResolved)
	}

	return idResolved, nil
}

func (m *ChatResolverCacheable) resolveOneOnOneChatIDByUserRef(userRef string, chats msmodels.ChatCollectionResponseable) (string, error) {
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

func (m *ChatResolverCacheable) resolveGroupChatIDByTopic(topic string, chats msmodels.ChatCollectionResponseable) (string, error) {
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
