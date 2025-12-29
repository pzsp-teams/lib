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

func TestMemberResolverCacheable_ResolveUserRefToMemberID_EmptyRef(t *testing.T) {
	ctx := context.Background()
	fc := &fakeCacher{}
	apiFake := &fakeChannelAPI{}
	chatFake := &fakeChatAPI{}
	res := NewMemberResolverCacheable(apiFake, chatFake, fc, true)

	memberCtx := res.NewChannelMemberContext("team-1", "chan-1", "   ")
	_, err := res.ResolveUserRefToMemberID(ctx, memberCtx)
	if err == nil {
		t.Fatalf("expected error for empty user reference, got nil")
	}
}

func TestMemberResolverCacheable_ResolveUserRefToMemberID_CacheHitSingleID(t *testing.T) {
	ctx := context.Background()
	fc := &fakeCacher{
		getValue: []string{"member-id-123"},
		getFound: true,
	}
	apiFake := &fakeChannelAPI{}
	res := NewMemberResolverCacheable(apiFake, &fakeChatAPI{}, fc, true)

	memberCtx := res.NewChannelMemberContext("team-1", "chan-1", "user-ref")
	id, err := res.ResolveUserRefToMemberID(ctx, memberCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "member-id-123" {
		t.Fatalf("expected member-id-123 from cache, got %q", id)
	}

	if fc.getCalls != 1 {
		t.Errorf("expected 1 Get call, got %d", fc.getCalls)
	}
	if fc.lastGetKey != cacher.NewChannelMemberKey("team-1", "chan-1", "user-ref", nil) {
		t.Errorf("unexpected cache key, got %q", fc.lastGetKey)
	}
	if fc.setCalls != 0 {
		t.Errorf("expected no Set on cache hit, got %d", fc.setCalls)
	}
	if apiFake.listMembersCalls != 0 {
		t.Errorf("expected no ListMembers on cache hit, got %d", apiFake.listMembersCalls)
	}
}

func TestMemberResolverCacheable_ResolveUserRefToMemberID_CacheMiss_UsesAPIAndCaches(t *testing.T) {
	ctx := context.Background()
	fc := &fakeCacher{
		getFound: false,
	}
	member := newAadUserMember("m-1", "u-1", "Alice")
	apiFake := &fakeChannelAPI{
		membersResp: newMemberCollection(member),
	}
	res := NewMemberResolverCacheable(apiFake, &fakeChatAPI{}, fc, true)

	memberCtx := res.NewChannelMemberContext("team-42", "chan-7", " u-1 ")
	id, err := res.ResolveUserRefToMemberID(ctx, memberCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "m-1" {
		t.Fatalf("expected m-1 from API, got %q", id)
	}

	if fc.getCalls != 1 {
		t.Errorf("expected 1 Get call, got %d", fc.getCalls)
	}
	if fc.lastGetKey != cacher.NewChannelMemberKey("team-42", "chan-7", "u-1", nil) {
		t.Errorf("unexpected cache key, got %q", fc.lastGetKey)
	}

	if apiFake.listMembersCalls != 1 {
		t.Errorf("expected 1 ListMembers call, got %d", apiFake.listMembersCalls)
	}
	if apiFake.lastMembersTeamID != "team-42" || apiFake.lastMembersChanID != "chan-7" {
		t.Errorf("expected ListMembers for team-42/chan-7, got team=%q chan=%q", apiFake.lastMembersTeamID, apiFake.lastMembersChanID)
	}

	if fc.setCalls != 1 {
		t.Errorf("expected 1 Set call, got %d", fc.setCalls)
	}
	if fc.lastSetKey != cacher.NewChannelMemberKey("team-42", "chan-7", "u-1", nil) {
		t.Errorf("unexpected Set key, got %q", fc.lastSetKey)
	}
	if v, ok := fc.lastSetValue.(string); !ok || v != "m-1" {
		t.Errorf("expected cached value 'm-1', got %#v", fc.lastSetValue)
	}
}

func TestMemberResolverCacheable_ResolveUserRefToMemberID_CacheDisabled_SkipsCache(t *testing.T) {
	ctx := context.Background()
	fc := &fakeCacher{
		getValue: []string{"member-id-cache"},
		getFound: true,
	}
	member := newAadUserMember("m-api", "u-1", "Alice")
	apiFake := &fakeChannelAPI{
		membersResp: newMemberCollection(member),
	}
	res := NewMemberResolverCacheable(apiFake, &fakeChatAPI{}, fc, false)

	memberCtx := res.NewChannelMemberContext("team-1", "chan-1", "u-1")
	id, err := res.ResolveUserRefToMemberID(ctx, memberCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "m-api" {
		t.Fatalf("expected m-api from API, got %q", id)
	}

	if fc.getCalls != 0 && fc.setCalls != 0 {
		t.Errorf("expected no cache calls when cache disabled, got get=%d set=%d", fc.getCalls, fc.setCalls)
	}
	if apiFake.listMembersCalls != 1 {
		t.Errorf("expected 1 ListMembers call, got %d", apiFake.listMembersCalls)
	}
}

func TestMemberResolverCacheable_ResolveUserRefToMemberID_ResolverErrorPropagated(t *testing.T) {
	ctx := context.Background()
	fc := &fakeCacher{
		getFound: false,
	}
	wantErr := &sender.RequestError{
		Message: "nope",
	}
	apiFake := &fakeChannelAPI{
		membersErr: wantErr,
	}
	res := NewMemberResolverCacheable(apiFake, &fakeChatAPI{}, fc, true)

	memberCtx := res.NewChannelMemberContext("team-1", "chan-1", "user-ref")
	_, err := res.ResolveUserRefToMemberID(ctx, memberCtx)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	var reqErr *sender.RequestError
	if !errors.As(err, &reqErr) {
		t.Fatalf("expected RequestError, got %T %v", err, err)
	}
}
