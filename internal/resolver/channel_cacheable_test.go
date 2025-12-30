package resolver

import (
	"context"
	"errors"
	"testing"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/cacher"
	sender "github.com/pzsp-teams/lib/internal/sender"
)

type fakeCacher struct {
	getValue     any
	getFound     bool
	getErr       error
	setErr       error
	getCalls     int
	setCalls     int
	lastGetKey   string
	lastSetKey   string
	lastSetValue any
}

func (f *fakeCacher) Get(key string) (valeu any, found bool, err error) {
	f.getCalls++
	f.lastGetKey = key
	return f.getValue, f.getFound, f.getErr
}

func (f *fakeCacher) Set(key string, value any) error {
	f.setCalls++
	f.lastSetKey = key
	f.lastSetValue = value
	return f.setErr
}

func (f *fakeCacher) Invalidate(key string) error {
	return nil
}

func (f *fakeCacher) Clear() error {
	return nil
}

type fakeChannelAPI struct {
	listResp       msmodels.ChannelCollectionResponseable
	listErr        *sender.RequestError
	listCalls      int
	lastListTeamID string

	membersResp       msmodels.ConversationMemberCollectionResponseable
	membersErr        *sender.RequestError
	listMembersCalls  int
	lastMembersTeamID string
	lastMembersChanID string
}

func (f *fakeChannelAPI) ListChannels(ctx context.Context, teamID string) (msmodels.ChannelCollectionResponseable, *sender.RequestError) {
	f.listCalls++
	f.lastListTeamID = teamID
	return f.listResp, f.listErr
}

func (f *fakeChannelAPI) ListMembers(ctx context.Context, teamID, channelID string) (msmodels.ConversationMemberCollectionResponseable, *sender.RequestError) {
	f.listMembersCalls++
	f.lastMembersTeamID = teamID
	f.lastMembersChanID = channelID
	return f.membersResp, f.membersErr
}

func (f *fakeChannelAPI) GetChannel(ctx context.Context, teamID, channelID string) (msmodels.Channelable, *sender.RequestError) {
	return nil, nil
}

func (f *fakeChannelAPI) CreateStandardChannel(ctx context.Context, teamID string, channel msmodels.Channelable) (msmodels.Channelable, *sender.RequestError) {
	return nil, nil
}

func (f *fakeChannelAPI) CreatePrivateChannelWithMembers(ctx context.Context, teamID, displayName string, memberIDs, ownerIDs []string) (msmodels.Channelable, *sender.RequestError) {
	return nil, nil
}

func (f *fakeChannelAPI) DeleteChannel(ctx context.Context, teamID, channelID string) *sender.RequestError {
	return nil
}

func (f *fakeChannelAPI) SendMessage(ctx context.Context, teamID, channelID, content, contentType string, mentions []msmodels.ChatMessageMentionable) (msmodels.ChatMessageable, *sender.RequestError) {
	return nil, nil
}

func (f *fakeChannelAPI) SendReply(ctx context.Context, teamID, channelID, messageID, content, contentType string, mentions []msmodels.ChatMessageMentionable) (msmodels.ChatMessageable, *sender.RequestError) {
	return nil, nil
}

func (f *fakeChannelAPI) ListMessages(ctx context.Context, teamID, channelID string, top *int32) (msmodels.ChatMessageCollectionResponseable, *sender.RequestError) {
	return nil, nil
}

func (f *fakeChannelAPI) GetMessage(ctx context.Context, teamID, channelID, messageID string) (msmodels.ChatMessageable, *sender.RequestError) {
	return nil, nil
}

func (f *fakeChannelAPI) ListReplies(ctx context.Context, teamID, channelID, messageID string, top *int32) (msmodels.ChatMessageCollectionResponseable, *sender.RequestError) {
	return nil, nil
}

func (f *fakeChannelAPI) GetReply(ctx context.Context, teamID, channelID, messageID, replyID string) (msmodels.ChatMessageable, *sender.RequestError) {
	return nil, nil
}

func (f *fakeChannelAPI) AddMember(ctx context.Context, teamID, channelID, userID, role string) (msmodels.ConversationMemberable, *sender.RequestError) {
	return nil, nil
}

func (f *fakeChannelAPI) RemoveMember(ctx context.Context, teamID, channelID, memberID string) *sender.RequestError {
	return nil
}

func (f *fakeChannelAPI) UpdateMemberRole(ctx context.Context, teamID, channelID, memberID, role string) (msmodels.ConversationMemberable, *sender.RequestError) {
	return nil, nil
}

func TestChannelResolverCacheable_ResolveChannelRefToID_EmptyRef(t *testing.T) {
	ctx := context.Background()
	fc := &fakeCacher{}
	apiFake := &fakeChannelAPI{}

	res := NewChannelResolverCacheable(apiFake, fc, true)

	_, err := res.ResolveChannelRefToID(ctx, "team-1", "   ")
	if err == nil {
		t.Fatalf("expected error for empty channel reference, got nil")
	}
}

func TestChannelResolverCacheable_ResolveChannelRefToID_DirectID_ShortCircuit(t *testing.T) {
	ctx := context.Background()
	fc := &fakeCacher{}
	apiFake := &fakeChannelAPI{}
	res := NewChannelResolverCacheable(apiFake, fc, true)

	chID := "19:abc123@thread.tacv2"

	id, err := res.ResolveChannelRefToID(ctx, "team-1", chID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != chID {
		t.Fatalf("expected %q, got %q", chID, id)
	}

	if fc.getCalls != 0 || fc.setCalls != 0 {
		t.Errorf("expected no cache calls for direct channel ID, got get=%d set=%d", fc.getCalls, fc.setCalls)
	}
	if apiFake.listCalls != 0 {
		t.Errorf("expected no ListChannels calls for direct ID, got %d", apiFake.listCalls)
	}
}

func TestChannelResolverCacheable_ResolveChannelRefToID_CacheHitSingleID(t *testing.T) {
	ctx := context.Background()
	fc := &fakeCacher{
		getValue: []string{"chan-id-123"},
		getFound: true,
	}

	apiFake := &fakeChannelAPI{}
	res := NewChannelResolverCacheable(apiFake, fc, true)

	id, err := res.ResolveChannelRefToID(ctx, "team-1", "General")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "chan-id-123" {
		t.Fatalf("expected chan-id-123 from cache, got %q", id)
	}

	if fc.getCalls != 1 {
		t.Errorf("expected 1 cache Get call, got %d", fc.getCalls)
	}
	if fc.lastGetKey != cacher.NewChannelKey("team-1", "General") {
		t.Errorf("unexpected cache key, got %q", fc.lastGetKey)
	}
	if fc.setCalls != 0 {
		t.Errorf("expected no cache Set on hit, got %d", fc.setCalls)
	}
	if apiFake.listCalls != 0 {
		t.Errorf("expected no ListChannels on cache hit, got %d", apiFake.listCalls)
	}
}

func TestChannelResolverCacheable_ResolveChannelRefToID_CacheMiss_UsesAPIAndCaches(t *testing.T) {
	ctx := context.Background()
	fc := &fakeCacher{
		getFound: false,
	}
	ch := newGraphChannel("chan-id-xyz", "General")
	apiFake := &fakeChannelAPI{
		listResp: newChannelCollection(ch),
	}
	res := NewChannelResolverCacheable(apiFake, fc, true)

	id, err := res.ResolveChannelRefToID(ctx, "team-42", "  General ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "chan-id-xyz" {
		t.Fatalf("expected chan-id-xyz, got %q", id)
	}

	if fc.getCalls != 1 {
		t.Errorf("expected 1 Get call, got %d", fc.getCalls)
	}
	if fc.lastGetKey != cacher.NewChannelKey("team-42", "General") {
		t.Errorf("unexpected cache key, got %q", fc.lastGetKey)
	}

	if apiFake.listCalls != 1 {
		t.Errorf("expected 1 ListChannels call, got %d", apiFake.listCalls)
	}
	if apiFake.lastListTeamID != "team-42" {
		t.Errorf("expected ListChannels for team-42, got %q", apiFake.lastListTeamID)
	}

	if fc.setCalls != 1 {
		t.Errorf("expected 1 Set call, got %d", fc.setCalls)
	}
	if fc.lastSetKey != cacher.NewChannelKey("team-42", "General") {
		t.Errorf("unexpected Set key, got %q", fc.lastSetKey)
	}
	if v, ok := fc.lastSetValue.(string); !ok || v != "chan-id-xyz" {
		t.Errorf("expected cached value 'chan-id-xyz', got %#v", fc.lastSetValue)
	}
}

func TestChannelResolverCacheable_ResolveChannelRefToID_CacheDisabled_SkipsCache(t *testing.T) {
	ctx := context.Background()
	fc := &fakeCacher{
		getValue: []string{"chan-id-from-cache"},
		getFound: true,
	}
	ch := newGraphChannel("chan-id-api", "General")
	apiFake := &fakeChannelAPI{
		listResp: newChannelCollection(ch),
	}

	res := NewChannelResolverCacheable(apiFake, fc, false)

	id, err := res.ResolveChannelRefToID(ctx, "team-1", "General")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "chan-id-api" {
		t.Fatalf("expected chan-id-api from API, got %q", id)
	}

	if fc.getCalls != 0 && fc.setCalls != 0 {
		t.Errorf("expected no cache calls when cache disabled, got get=%d set=%d", fc.getCalls, fc.setCalls)
	}
	if apiFake.listCalls != 1 {
		t.Errorf("expected 1 ListChannels call, got %d", apiFake.listCalls)
	}
}

func TestChannelResolverCacheable_ResolveChannelRefToID_ResolverErrorPropagated(t *testing.T) {
	ctx := context.Background()
	fc := &fakeCacher{
		getFound: false,
	}
	apiErr := &sender.RequestError{
		Message: "boom",
	}
	apiFake := &fakeChannelAPI{
		listErr: apiErr,
	}
	res := NewChannelResolverCacheable(apiFake, fc, true)

	_, err := res.ResolveChannelRefToID(ctx, "team-1", "General")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestChannelResolverCacheable_ResolveUserRefToMemberID_EmptyRef(t *testing.T) {
	ctx := context.Background()
	fc := &fakeCacher{}
	apiFake := &fakeChannelAPI{}
	res := NewChannelResolverCacheable(apiFake, fc, true)

	_, err := res.ResolveChannelMemberRefToID(ctx, "team-1", "chan-1", " ")
	if err == nil {
		t.Fatalf("expected error for empty user reference, got nil")
	}
}

func TestChannelResolverCacheable_ResolveUserRefToMemberID_CacheHitSingleID(t *testing.T) {
	ctx := context.Background()
	fc := &fakeCacher{
		getValue: []string{"member-id-123"},
		getFound: true,
	}
	apiFake := &fakeChannelAPI{}
	res := NewChannelResolverCacheable(apiFake, fc, true)

	id, err := res.ResolveChannelMemberRefToID(ctx, "team-1", "chan-1", "user-ref")
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

func TestChannelResolverCacheable_ResolveUserRefToMemberID_CacheMiss_UsesAPIAndCaches(t *testing.T) {
	ctx := context.Background()
	fc := &fakeCacher{
		getFound: false,
	}
	member := newAadUserMember("m-1", "u-1", "Alice")
	apiFake := &fakeChannelAPI{
		membersResp: newMemberCollection(member),
	}
	res := NewChannelResolverCacheable(apiFake, fc, true)
	id, err := res.ResolveChannelMemberRefToID(ctx, "team-42", "chan-7", " u-1 ")
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

func TestChannelResolverCacheable_ResolveUserRefToMemberID_CacheDisabled_SkipsCache(t *testing.T) {
	ctx := context.Background()
	fc := &fakeCacher{
		getFound: true,
	}
	member := newAadUserMember("m-api", "u-1", "Alice")
	apiFake := &fakeChannelAPI{
		membersResp: newMemberCollection(member),
	}
	res := NewChannelResolverCacheable(apiFake, fc, false)

	id, err := res.ResolveChannelMemberRefToID(ctx, "team-1", "chan-1", "u-1")
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

func TestChannelResolverCacheable_ResolveUserRefToMemberID_ResolverErrorPropagated(t *testing.T) {
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
	res := NewChannelResolverCacheable(apiFake, fc, true)

	_, err := res.ResolveChannelMemberRefToID(ctx, "team-1", "chan-1", "user-ref")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	var reqErr *sender.RequestError
	if !errors.As(err, &reqErr) {
		t.Fatalf("expected RequestError, got %T %v", err, err)
	}
}
