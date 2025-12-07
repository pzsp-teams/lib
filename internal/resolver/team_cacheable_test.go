package resolver

import (
	"context"
	"errors"
	"testing"
)

type fakeCacher struct {
	getValue any
	getFound bool
	getErr   error

	lastGetKey string
	getCalls   int

	lastSetKey   string
	lastSetValue any
	setCalls     int
}

func (f *fakeCacher) Get(key string) (any, bool, error) {
	f.getCalls++
	f.lastGetKey = key
	return f.getValue, f.getFound, f.getErr
}

func (f *fakeCacher) Set(key string, value any) error {
	f.setCalls++
	f.lastSetKey = key
	f.lastSetValue = value
	return nil
}

func (f *fakeCacher) Invalidate(key string) error { return nil }
func (f *fakeCacher) Clear() error               { return nil }


type fakeTeamResolver struct {
	result  string
	err     error
	calls   int
	lastRef string
}

func (f *fakeTeamResolver) ResolveTeamRefToID(ctx context.Context, teamRef string) (string, error) {
	f.calls++
	f.lastRef = teamRef
	return f.result, f.err
}

func TestTeamResolverCacheable_EmptyRef(t *testing.T) {
	ctx := context.Background()
	res := NewTeamResolverCacheable(nil, nil)

	_, err := res.ResolveTeamRefToID(ctx, "   ")
	if err == nil {
		t.Fatalf("expected error for empty team reference, got nil")
	}
}

func TestTeamResolverCacheable_GUIDShortCircuit(t *testing.T) {
	ctx := context.Background()

	guid := "123e4567-e89b-12d3-a456-426614174000"

	c := &fakeCacher{}
	r := &fakeTeamResolver{}
	res := NewTeamResolverCacheable(c, r)

	id, err := res.ResolveTeamRefToID(ctx, guid)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != guid {
		t.Fatalf("expected id %q, got %q", guid, id)
	}
	if c.getCalls != 0 || c.setCalls != 0 {
		t.Errorf("expected no cache calls for GUID, got get=%d set=%d", c.getCalls, c.setCalls)
	}
	if r.calls != 0 {
		t.Errorf("expected no resolver calls for GUID, got %d", r.calls)
	}
}

func TestTeamResolverCacheable_UsesCacheOnHit(t *testing.T) {
	ctx := context.Background()

	fc := &fakeCacher{
		getValue: "team-id-123",
		getFound: true,
		getErr:   nil,
	}
	fr := &fakeTeamResolver{
		result: "should-not-be-used",
	}
	res := NewTeamResolverCacheable(fc, fr)

	id, err := res.ResolveTeamRefToID(ctx, "my-team")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "team-id-123" {
		t.Fatalf("expected id team-id-123 from cache, got %q", id)
	}

	if fc.getCalls != 1 {
		t.Errorf("expected 1 cache Get call, got %d", fc.getCalls)
	}
	if fc.lastGetKey != "$team$:my-team" {
		t.Errorf("expected cache key $team$:my-team, got %q", fc.lastGetKey)
	}
	if fr.calls != 0 {
		t.Errorf("expected resolver not to be called on cache hit, got %d calls", fr.calls)
	}
	if fc.setCalls != 0 {
		t.Errorf("expected no cache Set on cache hit, got %d", fc.setCalls)
	}
}

func TestTeamResolverCacheable_ResolverCalledOnMissAndCaches(t *testing.T) {
	ctx := context.Background()

	fc := &fakeCacher{
		getFound: false, 
		getErr:   nil,
	}
	fr := &fakeTeamResolver{
		result: "team-id-xyz",
	}
	res := NewTeamResolverCacheable(fc, fr)

	id, err := res.ResolveTeamRefToID(ctx, " my-team ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "team-id-xyz" {
		t.Fatalf("expected id team-id-xyz from resolver, got %q", id)
	}

	if fc.getCalls != 1 {
		t.Errorf("expected 1 cache Get call, got %d", fc.getCalls)
	}
	if fc.lastGetKey != "$team$:my-team" {
		t.Errorf("expected cache key $team$:my-team, got %q", fc.lastGetKey)
	}

	if fr.calls != 1 {
		t.Errorf("expected resolver called once, got %d", fr.calls)
	}
	if fr.lastRef != "my-team" {
		t.Errorf("expected resolver called with 'my-team', got %q", fr.lastRef)
	}

	if fc.setCalls != 1 {
		t.Errorf("expected 1 cache Set call, got %d", fc.setCalls)
	}
	if fc.lastSetKey != "$team$:my-team" {
		t.Errorf("expected cache Set key $team$:my-team, got %q", fc.lastSetKey)
	}
	if v, ok := fc.lastSetValue.(string); !ok || v != "team-id-xyz" {
		t.Errorf("expected cache value 'team-id-xyz', got %#v", fc.lastSetValue)
	}
}

func TestTeamResolverCacheable_ResolverErrorPropagated(t *testing.T) {
	ctx := context.Background()

	fc := &fakeCacher{
		getFound: false,
	}
	expectedErr := errors.New("resolver failed")
	fr := &fakeTeamResolver{
		err: expectedErr,
	}
	res := NewTeamResolverCacheable(fc, fr)

	_, err := res.ResolveTeamRefToID(ctx, "team-x")
	if err == nil {
		t.Fatalf("expected error from resolver, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected resolver error %v, got %v", expectedErr, err)
	}

	if fr.calls != 1 {
		t.Errorf("expected resolver called once, got %d", fr.calls)
	}
	if fc.setCalls != 0 {
		t.Errorf("expected no cache Set when resolver fails, got %d", fc.setCalls)
	}
}

func TestTeamResolverCacheable_WrongTypeInCacheFallsBack(t *testing.T) {
	ctx := context.Background()

	fc := &fakeCacher{
		getValue: 123,
		getFound: true,
		getErr:   nil,
	}
	fr := &fakeTeamResolver{
		result: "team-id-from-resolver",
	}
	res := NewTeamResolverCacheable(fc, fr)

	id, err := res.ResolveTeamRefToID(ctx, "team-z")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "team-id-from-resolver" {
		t.Fatalf("expected id team-id-from-resolver, got %q", id)
	}

	if fr.calls != 1 {
		t.Errorf("expected resolver called once due to wrong cache type, got %d", fr.calls)
	}
	if fc.setCalls != 1 {
		t.Errorf("expected cache Set called once to overwrite bad value, got %d", fc.setCalls)
	}
	if fc.lastSetKey != "$team$:team-z" {
		t.Errorf("expected cache Set key $team$:team-z, got %q", fc.lastSetKey)
	}
}
