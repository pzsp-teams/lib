package resolver

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/pzsp-teams/lib/cacher"
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


func TestTeamResolverCacheable_JSONFileCacher_MissThenHitSameInstance(t *testing.T) {
	t.Helper()
	ctx := context.Background()

	tmpDir := t.TempDir()
	cacheFile := filepath.Join(tmpDir, "teams-cache.json")

	base := &fakeTeamResolver{
		result: "team-id-1",
	}
	cache := cacher.NewJSONFileCacher(cacheFile)

	res := NewTeamResolverCacheable(cache, base)

	id1, err := res.ResolveTeamRefToID(ctx, "MyTeam")
	if err != nil {
		t.Fatalf("unexpected error on first call: %v", err)
	}
	if id1 != "team-id-1" {
		t.Fatalf("expected id 'team-id-1', got %q", id1)
	}
	if base.calls != 1 {
		t.Fatalf("expected resolver to be called once, got %d", base.calls)
	}

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		t.Fatalf("failed to read cache file: %v", err)
	}
	if len(data) == 0 {
		t.Fatalf("expected non-empty cache file")
	}

	var stored map[string]json.RawMessage
	if err := json.Unmarshal(data, &stored); err != nil {
		t.Fatalf("failed to unmarshal cache JSON: %v", err)
	}

	key := cacher.NewTeamKeyBuilder("MyTeam").ToString()
	raw, ok := stored[key]
	if !ok {
		t.Fatalf("expected key %q in cache file, got keys: %#v", key, stored)
	}

	var storedID string
	if err := json.Unmarshal(raw, &storedID); err != nil {
		t.Fatalf("failed to unmarshal cached ID: %v", err)
	}
	if storedID != "team-id-1" {
		t.Fatalf("expected cached ID 'team-id-1', got %q", storedID)
	}

	base.result = "team-id-2"

	id2, err := res.ResolveTeamRefToID(ctx, "MyTeam")
	if err != nil {
		t.Fatalf("unexpected error on second call: %v", err)
	}
	if id2 != "team-id-1" {
		t.Fatalf("expected id from cache 'team-id-1', got %q", id2)
	}
	if base.calls != 1 {
		t.Fatalf("expected resolver to still be called only once, got %d", base.calls)
	}
}

func TestTeamResolverCacheable_JSONFileCacher_LoadsExistingFile(t *testing.T) {
	t.Helper()
	ctx := context.Background()

	tmpDir := t.TempDir()
	cacheFile := filepath.Join(tmpDir, "teams-cache.json")

	popCache := cacher.NewJSONFileCacher(cacheFile)
	key := cacher.NewTeamKeyBuilder("MyTeam").ToString()
	if err := popCache.Set(key, "team-id-from-file"); err != nil {
		t.Fatalf("failed to pre-populate cache file: %v", err)
	}

	base := &fakeTeamResolver{
		result: "should-not-be-called",
	}
	cache := cacher.NewJSONFileCacher(cacheFile)
	res := NewTeamResolverCacheable(cache, base)

	id, err := res.ResolveTeamRefToID(ctx, "MyTeam")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "team-id-from-file" {
		t.Fatalf("expected id 'team-id-from-file' loaded from cache file, got %q", id)
	}
	if base.calls != 0 {
		t.Fatalf("expected resolver not to be called, got %d calls", base.calls)
	}
}

func TestTeamResolverCacheable_JSONFileCacher_CorruptedFileFallsBackToResolver(t *testing.T) {
	t.Helper()
	ctx := context.Background()

	tmpDir := t.TempDir()
	cacheFile := filepath.Join(tmpDir, "teams-cache.json")

	if err := os.WriteFile(cacheFile, []byte("{ this is not valid json"), 0o644); err != nil {
		t.Fatalf("failed to write corrupted cache file: %v", err)
	}

	base := &fakeTeamResolver{
		result: "team-id-from-resolver",
	}
	cache := cacher.NewJSONFileCacher(cacheFile)
	res := NewTeamResolverCacheable(cache, base)

	id, err := res.ResolveTeamRefToID(ctx, "MyTeam")
	if err != nil {
		t.Fatalf("unexpected error (should fallback to resolver): %v", err)
	}
	if id != "team-id-from-resolver" {
		t.Fatalf("expected id 'team-id-from-resolver' from resolver fallback, got %q", id)
	}
	if base.calls != 1 {
		t.Fatalf("expected resolver called once on corrupted cache, got %d", base.calls)
	}
}