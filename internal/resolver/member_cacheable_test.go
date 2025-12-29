package resolver

import (
	"context"
	"testing"
)

func TestChannelResolverCacheable_ResolveUserRefToMemberID_EmptyRef(t *testing.T) {
	ctx := context.Background()
	fc := &fakeCacher{}
	apiFake := &fakeChannelAPI{}
	res := NewMemberResolverCacheable(apiFake, fc, true)

	_, err := res.ResolveUserRefToMemberID(ctx, "team-1", "chan-1", "   ")
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

	id, err := res.ResolveUserRefToMemberID(ctx, "team-1", "chan-1", "user-ref")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "member-id-123" {
		t.Fatalf("expected member-id-123 from cache, got %q", id)
	}

	if fc.getCalls != 1 {
		t.Errorf("expected 1 Get call, got %d", fc.getCalls)
	}
	if fc.lastGetKey != cacher.NewMemberKey("user-ref", "team-1", "chan-1", nil) {
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

	id, err := res.ResolveUserRefToMemberID(ctx, "team-42", "chan-7", " u-1 ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "m-1" {
		t.Fatalf("expected m-1 from API, got %q", id)
	}

	if fc.getCalls != 1 {
		t.Errorf("expected 1 Get call, got %d", fc.getCalls)
	}
	if fc.lastGetKey != cacher.NewMemberKey("u-1", "team-42", "chan-7", nil) {
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
	if fc.lastSetKey != cacher.NewMemberKey("u-1", "team-42", "chan-7", nil) {
		t.Errorf("unexpected Set key, got %q", fc.lastSetKey)
	}
	if v, ok := fc.lastSetValue.(string); !ok || v != "m-1" {
		t.Errorf("expected cached value 'm-1', got %#v", fc.lastSetValue)
	}
}

func TestChannelResolverCacheable_ResolveUserRefToMemberID_CacheDisabled_SkipsCache(t *testing.T) {
	ctx := context.Background()
	fc := &fakeCacher{
		getValue: []string{"member-id-cache"},
		getFound: true,
	}
	member := newAadUserMember("m-api", "u-1", "Alice")
	apiFake := &fakeChannelAPI{
		membersResp: newMemberCollection(member),
	}

	res := NewChannelResolverCacheable(apiFake, fc, false)

	id, err := res.ResolveUserRefToMemberID(ctx, "team-1", "chan-1", "u-1")
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

	_, err := res.ResolveUserRefToMemberID(ctx, "team-1", "chan-1", "user-ref")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	var reqErr *sender.RequestError
	if !errors.As(err, &reqErr) {
		t.Fatalf("expected RequestError, got %T %v", err, err)
	}
}
