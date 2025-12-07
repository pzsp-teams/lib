package resolver

import (
	"context"
	"errors"
	"testing"
)

type fakeChannelResolver struct {
	channelIDResult string
	channelErr      error
	channelCalls    int
	lastTeamID      string
	lastChannelRef  string

	memberIDResult   string
	memberErr        error
	memberCalls      int
	lastTeamIDMember string
	lastChannelID    string
	lastUserRef      string
}

func (f *fakeChannelResolver) ResolveChannelRefToID(ctx context.Context, teamID, channelName string) (string, error) {
	f.channelCalls++
	f.lastTeamID = teamID
	f.lastChannelRef = channelName
	return f.channelIDResult, f.channelErr
}

func (f *fakeChannelResolver) ResolveUserRefToMemberID(ctx context.Context, teamID, channelID, userRef string) (string, error) {
	f.memberCalls++
	f.lastTeamIDMember = teamID
	f.lastChannelID = channelID
	f.lastUserRef = userRef
	return f.memberIDResult, f.memberErr
}

func TestChannelResolverCacheable_ResolveChannelRefToID_EmptyRef(t *testing.T) {
	ctx := context.Background()
	res := NewChannelResolverCacheable(nil, nil)

	_, err := res.ResolveChannelRefToID(ctx, "team-1", "   ")
	if err == nil {
		t.Fatalf("expected error for empty channel reference, got nil")
	}
}

func TestChannelResolverCacheable_ResolveChannelRefToID_GUIDShortCircuit(t *testing.T) {
	ctx := context.Background()

	channelID := "19:abc123@thread.tacv2"

	c := &fakeCacher{}
	r := &fakeChannelResolver{}
	res := NewChannelResolverCacheable(c, r)

	id, err := res.ResolveChannelRefToID(ctx, "team-1", channelID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != channelID {
		t.Fatalf("expected id %q, got %q", channelID, id)
	}

	if c.getCalls != 0 || c.setCalls != 0 {
		t.Errorf("expected no cache calls for channel ID, got get=%d set=%d", c.getCalls, c.setCalls)
	}
	if r.channelCalls != 0 {
		t.Errorf("expected resolver not called for channel ID, got %d calls", r.channelCalls)
	}
}

func TestChannelResolverCacheable_ResolveChannelRefToID_UsesCacheOnHit(t *testing.T) {
	ctx := context.Background()

	fc := &fakeCacher{
		getValue: "chan-id-123",
		getFound: true,
		getErr:   nil,
	}
	fr := &fakeChannelResolver{
		channelIDResult: "should-not-be-used",
	}
	res := NewChannelResolverCacheable(fc, fr)

	id, err := res.ResolveChannelRefToID(ctx, "team-1", "general")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "chan-id-123" {
		t.Fatalf("expected id chan-id-123 from cache, got %q", id)
	}

	if fc.getCalls != 1 {
		t.Errorf("expected 1 cache Get call, got %d", fc.getCalls)
	}
	if fc.lastGetKey != "$channel$:team-1:general" {
		t.Errorf("expected cache key $channel$:team-1:general, got %q", fc.lastGetKey)
	}
	if fr.channelCalls != 0 {
		t.Errorf("expected resolver not called on cache hit, got %d", fr.channelCalls)
	}
	if fc.setCalls != 0 {
		t.Errorf("expected no cache Set on cache hit, got %d", fc.setCalls)
	}
}

func TestChannelResolverCacheable_ResolveChannelRefToID_ResolverCalledOnMissAndCaches(t *testing.T) {
	ctx := context.Background()

	fc := &fakeCacher{
		getFound: false,
	}
	fr := &fakeChannelResolver{
		channelIDResult: "chan-id-xyz",
	}
	res := NewChannelResolverCacheable(fc, fr)

	id, err := res.ResolveChannelRefToID(ctx, "team-42", " general ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "chan-id-xyz" {
		t.Fatalf("expected id chan-id-xyz from resolver, got %q", id)
	}

	if fc.getCalls != 1 {
		t.Errorf("expected 1 cache Get call, got %d", fc.getCalls)
	}
	if fc.lastGetKey != "$channel$:team-42:general" {
		t.Errorf("expected cache key $channel$:team-42:general, got %q", fc.lastGetKey)
	}

	if fr.channelCalls != 1 {
		t.Errorf("expected resolver called once, got %d", fr.channelCalls)
	}
	if fr.lastTeamID != "team-42" || fr.lastChannelRef != "general" {
		t.Errorf("expected resolver called with team-42/general, got team=%q chan=%q", fr.lastTeamID, fr.lastChannelRef)
	}

	if fc.setCalls != 1 {
		t.Errorf("expected 1 cache Set call, got %d", fc.setCalls)
	}
	if fc.lastSetKey != "$channel$:team-42:general" {
		t.Errorf("expected cache Set key $channel$:team-42:general, got %q", fc.lastSetKey)
	}
	if v, ok := fc.lastSetValue.(string); !ok || v != "chan-id-xyz" {
		t.Errorf("expected cache value 'chan-id-xyz', got %#v", fc.lastSetValue)
	}
}

func TestChannelResolverCacheable_ResolveChannelRefToID_ResolverErrorPropagated(t *testing.T) {
	ctx := context.Background()

	fc := &fakeCacher{
		getFound: false,
	}
	wantErr := errors.New("channel resolver failed")
	fr := &fakeChannelResolver{
		channelErr: wantErr,
	}
	res := NewChannelResolverCacheable(fc, fr)

	_, err := res.ResolveChannelRefToID(ctx, "team-1", "chan-x")
	if err == nil {
		t.Fatalf("expected error from resolver, got nil")
	}
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected resolver error %v, got %v", wantErr, err)
	}

	if fr.channelCalls != 1 {
		t.Errorf("expected resolver called once, got %d", fr.channelCalls)
	}
	if fc.setCalls != 0 {
		t.Errorf("expected no cache Set when resolver fails, got %d", fc.setCalls)
	}
}

func TestChannelResolverCacheable_ResolveChannelRefToID_WrongTypeInCacheFallsBack(t *testing.T) {
	ctx := context.Background()

	fc := &fakeCacher{
		getValue: 123, 
		getFound: true,
		getErr:   nil,
	}
	fr := &fakeChannelResolver{
		channelIDResult: "chan-id-from-resolver",
	}
	res := NewChannelResolverCacheable(fc, fr)

	id, err := res.ResolveChannelRefToID(ctx, "team-1", "chan-z")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "chan-id-from-resolver" {
		t.Fatalf("expected id chan-id-from-resolver, got %q", id)
	}

	if fr.channelCalls != 1 {
		t.Errorf("expected resolver called once due to wrong cache type, got %d", fr.channelCalls)
	}
	if fc.setCalls != 1 {
		t.Errorf("expected cache Set called once to overwrite bad value, got %d", fc.setCalls)
	}
	if fc.lastSetKey != "$channel$:team-1:chan-z" {
		t.Errorf("expected cache Set key $channel$:team-1:chan-z, got %q", fc.lastSetKey)
	}
}

func TestChannelResolverCacheable_ResolveUserRefToMemberID_EmptyRef(t *testing.T) {
	ctx := context.Background()
	res := NewChannelResolverCacheable(nil, nil)

	_, err := res.ResolveUserRefToMemberID(ctx, "team-1", "chan-1", "   ")
	if err == nil {
		t.Fatalf("expected error for empty user reference, got nil")
	}
}

func TestChannelResolverCacheable_ResolveUserRefToMemberID_UsesCacheOnHit(t *testing.T) {
	ctx := context.Background()

	fc := &fakeCacher{
		getValue: "member-id-123",
		getFound: true,
		getErr:   nil,
	}
	fr := &fakeChannelResolver{
		memberIDResult: "should-not-be-used",
	}
	res := NewChannelResolverCacheable(fc, fr)

	id, err := res.ResolveUserRefToMemberID(ctx, "team-1", "chan-1", "user-ref")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "member-id-123" {
		t.Fatalf("expected id member-id-123 from cache, got %q", id)
	}

	if fc.getCalls != 1 {
		t.Errorf("expected 1 cache Get call, got %d", fc.getCalls)
	}
	if fc.lastGetKey != "$member$:team-1:chan-1:user-ref" {
		t.Errorf("expected cache key $member$:team-1:chan-1:user-ref, got %q", fc.lastGetKey)
	}
	if fr.memberCalls != 0 {
		t.Errorf("expected resolver not called on cache hit, got %d", fr.memberCalls)
	}
	if fc.setCalls != 0 {
		t.Errorf("expected no cache Set on cache hit, got %d", fc.setCalls)
	}
}

func TestChannelResolverCacheable_ResolveUserRefToMemberID_ResolverCalledOnMissAndCaches(t *testing.T) {
	ctx := context.Background()

	fc := &fakeCacher{
		getFound: false,
	}
	fr := &fakeChannelResolver{
		memberIDResult: "member-id-xyz",
	}
	res := NewChannelResolverCacheable(fc, fr)

	id, err := res.ResolveUserRefToMemberID(ctx, "team-42", "chan-7", " user@example.com ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "member-id-xyz" {
		t.Fatalf("expected id member-id-xyz from resolver, got %q", id)
	}

	if fc.getCalls != 1 {
		t.Errorf("expected 1 cache Get call, got %d", fc.getCalls)
	}
	if fc.lastGetKey != "$member$:team-42:chan-7:user@example.com" {
		t.Errorf("expected cache key $member$:team-42:chan-7:user@example.com, got %q", fc.lastGetKey)
	}

	if fr.memberCalls != 1 {
		t.Errorf("expected resolver called once, got %d", fr.memberCalls)
	}
	if fr.lastTeamIDMember != "team-42" || fr.lastChannelID != "chan-7" || fr.lastUserRef != "user@example.com" {
		t.Errorf("expected resolver called with team-42/chan-7/user@example.com, got team=%q chan=%q user=%q",
			fr.lastTeamIDMember, fr.lastChannelID, fr.lastUserRef)
	}

	if fc.setCalls != 1 {
		t.Errorf("expected 1 cache Set call, got %d", fc.setCalls)
	}
	if fc.lastSetKey != "$member$:team-42:chan-7:user@example.com" {
		t.Errorf("expected cache Set key $member$:team-42:chan-7:user@example.com, got %q", fc.lastSetKey)
	}
	if v, ok := fc.lastSetValue.(string); !ok || v != "member-id-xyz" {
		t.Errorf("expected cache value 'member-id-xyz', got %#v", fc.lastSetValue)
	}
}

func TestChannelResolverCacheable_ResolveUserRefToMemberID_ResolverErrorPropagated(t *testing.T) {
	ctx := context.Background()

	fc := &fakeCacher{
		getFound: false,
	}
	wantErr := errors.New("member resolver failed")
	fr := &fakeChannelResolver{
		memberErr: wantErr,
	}
	res := NewChannelResolverCacheable(fc, fr)

	_, err := res.ResolveUserRefToMemberID(ctx, "team-1", "chan-1", "user-ref")
	if err == nil {
		t.Fatalf("expected error from resolver, got nil")
	}
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected resolver error %v, got %v", wantErr, err)
	}

	if fr.memberCalls != 1 {
		t.Errorf("expected resolver called once, got %d", fr.memberCalls)
	}
	if fc.setCalls != 0 {
		t.Errorf("expected no cache Set when resolver fails, got %d", fc.setCalls)
	}
}

func TestChannelResolverCacheable_ResolveUserRefToMemberID_WrongTypeInCacheFallsBack(t *testing.T) {
	ctx := context.Background()

	fc := &fakeCacher{
		getValue: 123, 
		getFound: true,
		getErr:   nil,
	}
	fr := &fakeChannelResolver{
		memberIDResult: "member-id-from-resolver",
	}
	res := NewChannelResolverCacheable(fc, fr)

	id, err := res.ResolveUserRefToMemberID(ctx, "team-1", "chan-1", "user-ref")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "member-id-from-resolver" {
		t.Fatalf("expected id member-id-from-resolver, got %q", id)
	}

	if fr.memberCalls != 1 {
		t.Errorf("expected resolver called once due to wrong cache type, got %d", fr.memberCalls)
	}
	if fc.setCalls != 1 {
		t.Errorf("expected cache Set called once to overwrite bad value, got %d", fc.setCalls)
	}
	if fc.lastSetKey != "$member$:team-1:chan-1:user-ref" {
		t.Errorf("expected cache Set key $member$:team-1:chan-1:user-ref, got %q", fc.lastSetKey)
	}
}
