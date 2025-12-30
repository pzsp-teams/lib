package resolver

import (
	"context"
	"errors"
	"testing"
	"time"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/internal/cacher"
	sender "github.com/pzsp-teams/lib/internal/sender"
)

type fakeChatAPI struct {
	membersResp    msmodels.ConversationMemberCollectionResponseable
	membersErr     *sender.RequestError
	listGroupCalls int
	lastChatID     string

	listResp  msmodels.ChatCollectionResponseable
	listErr   *sender.RequestError
	listCalls int
	lastType  *string
}

func (f *fakeChatAPI) ListGroupChatMembers(ctx context.Context, chatID string) (msmodels.ConversationMemberCollectionResponseable, *sender.RequestError) {
	f.listGroupCalls++
	f.lastChatID = chatID
	return f.membersResp, f.membersErr
}

func (f *fakeChatAPI) ListChats(ctx context.Context, chatType *string) (msmodels.ChatCollectionResponseable, *sender.RequestError) {
	f.listCalls++
	f.lastType = chatType
	return f.listResp, f.listErr
}
func (f *fakeChatAPI) ListMessages(ctx context.Context, chatID string) (msmodels.ChatMessageCollectionResponseable, *sender.RequestError) {
	return nil, nil
}
func (f *fakeChatAPI) SendMessage(ctx context.Context, chatID, content, contentType string, mentions []msmodels.ChatMessageMentionable) (msmodels.ChatMessageable, *sender.RequestError) {
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

func TestChatResolverCacheable_ResolveOneOnOneRef_EmptyRef(t *testing.T) {
	ctx := context.Background()
	fc := &fakeCacher{}
	apiFake := &fakeChatAPI{}
	res := NewChatResolverCacheable(apiFake, fc, true)

	_, err := res.ResolveOneOnOneChatRefToID(ctx, "   ")
	if err == nil {
		t.Fatalf("expected error for empty one-on-one ref, got nil")
	}
}

func TestChatResolverCacheable_ResolveOneOnOneRef_DirectGUIDShortCircuit(t *testing.T) {
	ctx := context.Background()
	fc := &fakeCacher{}
	apiFake := &fakeChatAPI{}
	res := NewChatResolverCacheable(apiFake, fc, true)

	guid := "123e4567-e89b-12d3-a456-426614174000"
	id, err := res.ResolveOneOnOneChatRefToID(ctx, guid)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != guid {
		t.Fatalf("expected %q, got %q", guid, id)
	}
	if apiFake.listCalls != 0 {
		t.Errorf("expected no ListChats calls for GUID, got %d", apiFake.listCalls)
	}
	if fc.getCalls != 0 || fc.setCalls != 0 {
		t.Errorf("expected no cache calls for GUID, got get=%d set=%d", fc.getCalls, fc.setCalls)
	}
}

func TestChatResolverCacheable_ResolveOneOnOneRef_CacheHitSingleID(t *testing.T) {
	ctx := context.Background()
	fc := &fakeCacher{
		getValue: []string{"chat-id-123"},
		getFound: true,
	}
	apiFake := &fakeChatAPI{}
	res := NewChatResolverCacheable(apiFake, fc, true)

	id, err := res.ResolveOneOnOneChatRefToID(ctx, "user-ref")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "chat-id-123" {
		t.Fatalf("expected chat-id-123 from cache, got %q", id)
	}
	if fc.getCalls != 1 {
		t.Errorf("expected 1 Get call, got %d", fc.getCalls)
	}
	if fc.lastGetKey != cacher.NewOneOnOneChatKey("user-ref", nil) {
		t.Errorf("unexpected cache key, got %q", fc.lastGetKey)
	}
	if apiFake.listCalls != 0 {
		t.Errorf("expected no ListChats on cache hit, got %d", apiFake.listCalls)
	}
}

func TestChatResolverCacheable_ResolveOneOnOneRef_CacheMiss_UsesAPIAndCaches(t *testing.T) {
	ctx := context.Background()
	fc := &fakeCacher{getFound: false}
	m := newAadUserMember("m-1", "usr-1", "jane@example.com")
	chat := newOneOnOneChat("chat-1", m)
	apiFake := &fakeChatAPI{listResp: newChatCollection(chat)}
	res := NewChatResolverCacheable(apiFake, fc, true)

	id, err := res.ResolveOneOnOneChatRefToID(ctx, "  usr-1  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "chat-1" {
		t.Fatalf("expected chat-1 from API, got %q", id)
	}
	if apiFake.listCalls != 1 {
		t.Errorf("expected 1 ListChats call, got %d", apiFake.listCalls)
	}
	if fc.setCalls != 1 {
		t.Errorf("expected 1 Set call, got %d", fc.setCalls)
	}
}

func TestChatResolverCacheable_ResolveOneOnOneRef_CacheDisabled_SkipsCache(t *testing.T) {
	ctx := context.Background()
	fc := &fakeCacher{getFound: true}
	m := newAadUserMember("m-1", "usr-1", "jane@example.com")
	chat := newOneOnOneChat("chat-api", m)
	apiFake := &fakeChatAPI{listResp: newChatCollection(chat)}
	res := NewChatResolverCacheable(apiFake, fc, false)

	id, err := res.ResolveOneOnOneChatRefToID(ctx, "usr-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "chat-api" {
		t.Fatalf("expected chat-api from API, got %q", id)
	}
	if fc.getCalls != 0 || fc.setCalls != 0 {
		t.Errorf("expected no cache calls when cache disabled, got get=%d set=%d", fc.getCalls, fc.setCalls)
	}
	if apiFake.listCalls != 1 {
		t.Errorf("expected 1 ListChats call, got %d", apiFake.listCalls)
	}
}

func TestChatResolverCacheable_ResolveOneOnOneRef_ResolverErrorPropagated(t *testing.T) {
	ctx := context.Background()
	fc := &fakeCacher{getFound: false}
	wantErr := &sender.RequestError{Message: "nope"}
	apiFake := &fakeChatAPI{listErr: wantErr}
	res := NewChatResolverCacheable(apiFake, fc, true)

	_, err := res.ResolveOneOnOneChatRefToID(ctx, "user-ref")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	var reqErr *sender.RequestError
	if !errors.As(err, &reqErr) {
		t.Fatalf("expected RequestError, got %T %v", err, err)
	}
}

func TestChatResolverCacheable_ResolveGroupChatRefToID_EmptyRef(t *testing.T) {
	ctx := context.Background()
	fc := &fakeCacher{}
	apiFake := &fakeChatAPI{}
	res := NewChatResolverCacheable(apiFake, fc, true)

	_, err := res.ResolveGroupChatRefToID(ctx, "   ")
	if err == nil {
		t.Fatalf("expected error for empty group chat ref, got nil")
	}
}

func TestChatResolverCacheable_ResolveGroupChatRefToID_DirectIDShortCircuit(t *testing.T) {
	ctx := context.Background()
	fc := &fakeCacher{}
	apiFake := &fakeChatAPI{}
	res := NewChatResolverCacheable(apiFake, fc, true)

	threadID := "19:abc123@thread.tacv2"
	id, err := res.ResolveGroupChatRefToID(ctx, threadID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != threadID {
		t.Fatalf("expected %q, got %q", threadID, id)
	}
	if apiFake.listCalls != 0 {
		t.Errorf("expected no ListChats calls for direct thread id, got %d", apiFake.listCalls)
	}
}

func TestChatResolverCacheable_ResolveGroupChatRefToID_CacheHitSingleID(t *testing.T) {
	ctx := context.Background()
	fc := &fakeCacher{getValue: []string{"c-id-123"}, getFound: true}
	apiFake := &fakeChatAPI{}
	res := NewChatResolverCacheable(apiFake, fc, true)

	id, err := res.ResolveGroupChatRefToID(ctx, "My Topic")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "c-id-123" {
		t.Fatalf("expected c-id-123 from cache, got %q", id)
	}
	if fc.getCalls != 1 {
		t.Errorf("expected 1 Get call, got %d", fc.getCalls)
	}
}

func TestChatResolverCacheable_ResolveGroupChatRefToID_CacheMiss_UsesAPIAndCaches(t *testing.T) {
	ctx := context.Background()
	fc := &fakeCacher{getFound: false}
	chat := newGroupChat("gc-1", "Topic")
	apiFake := &fakeChatAPI{listResp: newChatCollection(chat)}
	res := NewChatResolverCacheable(apiFake, fc, true)

	id, err := res.ResolveGroupChatRefToID(ctx, "  Topic  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "gc-1" {
		t.Fatalf("expected gc-1 from API, got %q", id)
	}
	if apiFake.listCalls != 1 {
		t.Errorf("expected 1 ListChats call, got %d", apiFake.listCalls)
	}
	if fc.setCalls != 1 {
		t.Errorf("expected 1 Set call, got %d", fc.setCalls)
	}
}

func TestChatResolverCacheable_ResolveGroupChatRefToID_CacheDisabled_SkipsCache(t *testing.T) {
	ctx := context.Background()
	fc := &fakeCacher{getFound: true}
	chat := newGroupChat("gc-api", "Topic")
	apiFake := &fakeChatAPI{listResp: newChatCollection(chat)}
	res := NewChatResolverCacheable(apiFake, fc, false)

	id, err := res.ResolveGroupChatRefToID(ctx, "Topic")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "gc-api" {
		t.Fatalf("expected gc-api from API, got %q", id)
	}
	if fc.getCalls != 0 || fc.setCalls != 0 {
		t.Errorf("expected no cache calls when cache disabled, got get=%d set=%d", fc.getCalls, fc.setCalls)
	}
	if apiFake.listCalls != 1 {
		t.Errorf("expected 1 ListChats call, got %d", apiFake.listCalls)
	}
}

func TestChatResolverCacheable_ResolveGroupChatRefToID_ResolverErrorPropagated(t *testing.T) {
	ctx := context.Background()
	fc := &fakeCacher{getFound: false}
	wantErr := &sender.RequestError{Message: "nope"}
	apiFake := &fakeChatAPI{listErr: wantErr}
	res := NewChatResolverCacheable(apiFake, fc, true)

	_, err := res.ResolveGroupChatRefToID(ctx, "Topic")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	var reqErr *sender.RequestError
	if !errors.As(err, &reqErr) {
		t.Fatalf("expected RequestError, got %T %v", err, err)
	}
}

func TestChatResolverCacheable_ResolveUserRefToMemberID_EmptyRef(t *testing.T) {
	ctx := context.Background()
	fc := &fakeCacher{}
	apiFake := &fakeChatAPI{}
	res := NewChatResolverCacheable(apiFake, fc, true)

	_, err := res.ResolveChatMemberRefToID(ctx, "chat-1", "   ")
	if err == nil {
		t.Fatalf("expected error for empty user reference, got nil")
	}
}

func TestChatResolverCacheable_ResolveUserRefToMemberID_CacheHitSingleID(t *testing.T) {
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

func TestChatResolverCacheable_ResolveUserRefToMemberID_CacheMiss_UsesAPIAndCaches(t *testing.T) {
	ctx := context.Background()
	fc := &fakeCacher{
		getFound: false,
	}
	m := newAadUserMember("m-1", "usr-1", "jane@example.com")
	apiFake := &fakeChatAPI{
		membersResp: newMemberCollection(m),
	}
	res := NewChatResolverCacheable(apiFake, fc, true)

	id, err := res.ResolveChatMemberRefToID(ctx, "chat-42", " usr-1 ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "m-1" {
		t.Fatalf("expected m-1 from API, got %q", id)
	}

	if fc.getCalls != 1 {
		t.Errorf("expected 1 Get call, got %d", fc.getCalls)
	}

	if fc.lastGetKey != cacher.NewGroupChatMemberKey("chat-42", "usr-1", nil) {
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
	if fc.lastSetKey != cacher.NewGroupChatMemberKey("chat-42", "usr-1", nil) {
		t.Errorf("unexpected Set key, got %q", fc.lastSetKey)
	}
	if v, ok := fc.lastSetValue.(string); !ok || v != "m-1" {
		t.Errorf("expected cached value 'm-1', got %#v", fc.lastSetValue)
	}
}

func TestChatResolverCacheable_ResolveUserRefToMemberID_CacheDisabled_SkipsCache(t *testing.T) {
	ctx := context.Background()
	fc := &fakeCacher{
		getFound: true,
	}
	m := newAadUserMember("m-1", "usr-1", "jane@example.com")
	apiFake := &fakeChatAPI{
		membersResp: newMemberCollection(m),
	}
	res := NewChatResolverCacheable(apiFake, fc, false)

	id, err := res.ResolveChatMemberRefToID(ctx, "chat-1", "usr-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "m-1" {
		t.Fatalf("expected m-1 from API, got %q", id)
	}

	if fc.getCalls != 0 && fc.setCalls != 0 {
		t.Errorf("expected no cache calls when cache disabled, got get=%d set=%d", fc.getCalls, fc.setCalls)
	}
	if apiFake.listGroupCalls != 1 {
		t.Errorf("expected 1 ListGroupChatMembers call, got %d", apiFake.listGroupCalls)
	}
}

func TestChatResolverCacheable_ResolveUserRefToMemberID_ResolverErrorPropagated(t *testing.T) {
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
