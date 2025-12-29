package resolver

import (
	"context"
	"errors"
	"testing"
	"time"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/cacher"
	sender "github.com/pzsp-teams/lib/internal/sender"
)

type fakeChatAPI struct {
	membersResp msmodels.ConversationMemberCollectionResponseable
	membersErr  *sender.RequestError

	listGroupCalls int
	lastChatID     string
}

func (f *fakeChatAPI) ListGroupChatMembers(ctx context.Context, chatID string) (msmodels.ConversationMemberCollectionResponseable, *sender.RequestError) {
	f.listGroupCalls++
	f.lastChatID = chatID
	return f.membersResp, f.membersErr
}

func (f *fakeChatAPI) ListChats(ctx context.Context, chatType string) (msmodels.ChatCollectionResponseable, *sender.RequestError) {
	return nil, nil
}
func (f *fakeChatAPI) ListMessages(ctx context.Context, chatID string) (msmodels.ChatMessageCollectionResponseable, *sender.RequestError) {
	return nil, nil
}
func (f *fakeChatAPI) SendMessage(ctx context.Context, chatID, content, contentType string) (msmodels.ChatMessageable, *sender.RequestError) {
	return nil, nil
}
func (f *fakeChatAPI) DeleteMessage(ctx context.Context, chatID, messageID string) *sender.RequestError {
	return nil
}
func (f *fakeChatAPI) GetMessage(ctx context.Context, chatID, messageID string) (msmodels.ChatMessageable, *sender.RequestError) {
	return nil, nil
}
func (f *fakeChatAPI) ListAllMessages(ctx context.Context, startTime, endTime *time.Time, top *int32) (msmodels.ChatMessageCollectionResponseable, *sender.RequestError) {
	return nil, nil
}
func (f *fakeChatAPI) ListPinnedMessages(ctx context.Context, chatID string) (msmodels.ChatMessageCollectionResponseable, *sender.RequestError) {
	return nil, nil
}
func (f *fakeChatAPI) PinMessage(ctx context.Context, chatID, messageID string) *sender.RequestError {
	return nil
}
func (f *fakeChatAPI) UnpinMessage(ctx context.Context, chatID, pinnedID string) *sender.RequestError {
	return nil
}
func (f *fakeChatAPI) CreateOneOnOneChat(ctx context.Context, recipientRef string) (msmodels.Chatable, *sender.RequestError) {
	return nil, nil
}
func (f *fakeChatAPI) CreateGroupChat(ctx context.Context, recipientRefs []string, topic string, includeMe bool) (msmodels.Chatable, *sender.RequestError) {
	return nil, nil
}
func (f *fakeChatAPI) AddMemberToGroupChat(ctx context.Context, chatID, userRef string) (msmodels.ConversationMemberable, *sender.RequestError) {
	return nil, nil
}
func (f *fakeChatAPI) RemoveMemberFromGroupChat(ctx context.Context, chatID, userRef string) *sender.RequestError {
	return nil
}
func (f *fakeChatAPI) UpdateGroupChatTopic(ctx context.Context, chatID, topic string) (msmodels.Chatable, *sender.RequestError) {
	return nil, nil
}

func TestMemberResolverCacheable_ResolveUserRefToMemberID_EmptyRef_Chat(t *testing.T) {
	ctx := context.Background()
	fc := &fakeCacher{}
	apiFake := &fakeChatAPI{}
	res := NewChatResolverCacheable(apiFake, fc, true)

	_, err := res.ResolveChatMemberRefToID(ctx, "chat-1", "   ")
	if err == nil {
		t.Fatalf("expected error for empty user reference, got nil")
	}
}

func TestMemberResolverCacheable_ResolveUserRefToMemberID_CacheHitSingleID_Chat(t *testing.T) {
	ctx := context.Background()
	fc := &fakeCacher{
		getValue: []string{"member-id-123"},
		getFound: true,
	}
	apiFake := &fakeChatAPI{}
	res := NewChatResolverCacheable(apiFake, fc, true)

	id, err := res.ResolveChatMemberRefToID(ctx, "chat-1", "user-ref")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "member-id-123" {
		t.Fatalf("expected member-id-123 from cache, got %q", id)
	}

	if fc.getCalls != 1 {
		t.Errorf("expected 1 Get call, got %d", fc.getCalls)
	}
	if fc.lastGetKey != cacher.NewGroupChatMemberKey("chat-1", "user-ref", nil) {
		t.Errorf("unexpected cache key, got %q", fc.lastGetKey)
	}
	if fc.setCalls != 0 {
		t.Errorf("expected no Set on cache hit, got %d", fc.setCalls)
	}
	if apiFake.listGroupCalls != 0 {
		t.Errorf("expected no ListGroupChatMembers on cache hit, got %d", apiFake.listGroupCalls)
	}
}

func TestMemberResolverCacheable_ResolveUserRefToMemberID_CacheMiss_UsesAPIAndCaches_Chat(t *testing.T) {
	ctx := context.Background()
	fc := &fakeCacher{
		getFound: false,
	}
	member := newAadUserMember("m-1", "u-1", "Alice")
	apiFake := &fakeChatAPI{
		membersResp: newMemberCollection(member),
	}
	res := NewChatResolverCacheable(apiFake, fc, true)

	id, err := res.ResolveChatMemberRefToID(ctx, "chat-42", " u-1 ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "m-1" {
		t.Fatalf("expected m-1 from API, got %q", id)
	}

	if fc.getCalls != 1 {
		t.Errorf("expected 1 Get call, got %d", fc.getCalls)
	}
	if fc.lastGetKey != cacher.NewGroupChatMemberKey("chat-42", "u-1", nil) {
		t.Errorf("unexpected cache key, got %q", fc.lastGetKey)
	}

	if apiFake.listGroupCalls != 1 {
		t.Errorf("expected 1 ListGroupChatMembers call, got %d", apiFake.listGroupCalls)
	}
	if apiFake.lastChatID != "chat-42" {
		t.Errorf("expected ListGroupChatMembers for chat-42, got chat=%q", apiFake.lastChatID)
	}

	if fc.setCalls != 1 {
		t.Errorf("expected 1 Set call, got %d", fc.setCalls)
	}
	if fc.lastSetKey != cacher.NewGroupChatMemberKey("chat-42", "u-1", nil) {
		t.Errorf("unexpected Set key, got %q", fc.lastSetKey)
	}
	if v, ok := fc.lastSetValue.(string); !ok || v != "m-1" {
		t.Errorf("expected cached value 'm-1', got %#v", fc.lastSetValue)
	}
}

func TestMemberResolverCacheable_ResolveUserRefToMemberID_CacheDisabled_SkipsCache_Chat(t *testing.T) {
	ctx := context.Background()
	fc := &fakeCacher{
		getFound: true,
	}
	member := newAadUserMember("m-api", "u-1", "Alice")
	apiFake := &fakeChatAPI{
		membersResp: newMemberCollection(member),
	}
	res := NewChatResolverCacheable(apiFake, fc, false)

	id, err := res.ResolveChatMemberRefToID(ctx, "chat-1", "u-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "m-api" {
		t.Fatalf("expected m-api from API, got %q", id)
	}

	if fc.getCalls != 0 && fc.setCalls != 0 {
		t.Errorf("expected no cache calls when cache disabled, got get=%d set=%d", fc.getCalls, fc.setCalls)
	}
	if apiFake.listGroupCalls != 1 {
		t.Errorf("expected 1 ListGroupChatMembers call, got %d", apiFake.listGroupCalls)
	}
}

func TestMemberResolverCacheable_ResolveUserRefToMemberID_ResolverErrorPropagated_Chat(t *testing.T) {
	ctx := context.Background()
	fc := &fakeCacher{
		getFound: false,
	}
	wantErr := &sender.RequestError{
		Message: "nope",
	}
	apiFake := &fakeChatAPI{
		membersErr: wantErr,
	}
	res := NewChatResolverCacheable(apiFake, fc, true)

	_, err := res.ResolveChatMemberRefToID(ctx, "chat-1", "user-ref")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	var reqErr *sender.RequestError
	if !errors.As(err, &reqErr) {
		t.Fatalf("expected RequestError, got %T %v", err, err)
	}
}
