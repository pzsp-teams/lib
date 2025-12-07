package resolver

import (
	"context"
	"strings"
	"testing"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/cacher"
	sender "github.com/pzsp-teams/lib/internal/sender"
)

type fakeTeamAPI struct {
	listResp  msmodels.TeamCollectionResponseable
	listErr   *sender.RequestError
	listCalls int
}

func (f *fakeTeamAPI) CreateFromTemplate(ctx context.Context, displayName, description string, owners []string) (string, *sender.RequestError) {
	return "", nil
}

func (f *fakeTeamAPI) CreateViaGroup(ctx context.Context, displayName, mailNickname, visibility string) (string, *sender.RequestError) {
	return "", nil
}

func (f *fakeTeamAPI) Get(ctx context.Context, teamID string) (msmodels.Teamable, *sender.RequestError) {
	return nil, nil
}

func (f *fakeTeamAPI) ListMyJoined(ctx context.Context) (msmodels.TeamCollectionResponseable, *sender.RequestError) {
	f.listCalls++
	return f.listResp, f.listErr
}

func (f *fakeTeamAPI) Update(ctx context.Context, teamID string, patch *msmodels.Team) (msmodels.Teamable, *sender.RequestError) {
	return nil, nil
}

func (f *fakeTeamAPI) Archive(ctx context.Context, teamID string, spoReadOnlyForMembers *bool) *sender.RequestError {
	return nil
}

func (f *fakeTeamAPI) Unarchive(ctx context.Context, teamID string) *sender.RequestError {
	return nil
}

func (f *fakeTeamAPI) Delete(ctx context.Context, teamID string) *sender.RequestError {
	return nil
}

func (f *fakeTeamAPI) RestoreDeleted(ctx context.Context, deletedGroupID string) (msmodels.DirectoryObjectable, *sender.RequestError) {
	return nil, nil
}

func newGraphTeam(id, name string) msmodels.Teamable {
	t := msmodels.NewTeam()
	t.SetId(&id)
	t.SetDisplayName(&name)
	return t
}

func newTeamCollection(teams ...msmodels.Teamable) msmodels.TeamCollectionResponseable {
	col := msmodels.NewTeamCollectionResponse()
	col.SetValue(teams)
	return col
}

func TestTeamResolverCacheable_ResolveTeamRefToID_EmptyRef(t *testing.T) {
	ctx := context.Background()
	apiFake := &fakeTeamAPI{}
	res := NewTeamResolverCacheable(apiFake, nil, true)

	_, err := res.ResolveTeamRefToID(ctx, "   ")
	if err == nil {
		t.Fatalf("expected error for empty team reference, got nil")
	}
	if apiFake.listCalls != 0 {
		t.Errorf("expected no ListMyJoined calls on empty ref, got %d", apiFake.listCalls)
	}
}

func TestTeamResolverCacheable_ResolveTeamRefToID_GUIDShortCircuit(t *testing.T) {
	ctx := context.Background()
	apiFake := &fakeTeamAPI{}
	fc := &fakeCacher{}

	res := NewTeamResolverCacheable(apiFake, fc, true)

	guid := "123e4567-e89b-12d3-a456-426614174000"

	id, err := res.ResolveTeamRefToID(ctx, guid)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != guid {
		t.Fatalf("expected %q, got %q", guid, id)
	}
	if apiFake.listCalls != 0 {
		t.Errorf("expected no ListMyJoined calls for GUID, got %d", apiFake.listCalls)
	}
	if fc.getCalls != 0 || fc.setCalls != 0 {
		t.Errorf("expected no cache calls for GUID, got get=%d set=%d", fc.getCalls, fc.setCalls)
	}
}

func TestTeamResolverCacheable_ResolveTeamRefToID_CacheHitSingleID(t *testing.T) {
	ctx := context.Background()

	fc := &fakeCacher{
		getValue: []string{"team-id-123"},
		getFound: true,
	}
	apiFake := &fakeTeamAPI{}

	res := NewTeamResolverCacheable(apiFake, fc, true)

	id, err := res.ResolveTeamRefToID(ctx, "My Team")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "team-id-123" {
		t.Fatalf("expected team-id-123 from cache, got %q", id)
	}

	if fc.getCalls != 1 {
		t.Errorf("expected 1 cache Get call, got %d", fc.getCalls)
	}
	expectedKey := cacher.NewTeamKeyBuilder("My Team").ToString()
	if fc.lastGetKey != expectedKey {
		t.Errorf("expected cache key %q, got %q", expectedKey, fc.lastGetKey)
	}
	if fc.setCalls != 0 {
		t.Errorf("expected no cache Set on cache hit, got %d", fc.setCalls)
	}
	if apiFake.listCalls != 0 {
		t.Errorf("expected no ListMyJoined on cache hit, got %d", apiFake.listCalls)
	}
}

func TestTeamResolverCacheable_ResolveTeamRefToID_CacheMiss_UsesAPIAndCaches(t *testing.T) {
	ctx := context.Background()

	fc := &fakeCacher{
		getFound: false,
	}
	team := newGraphTeam("team-id-xyz", "My Team")
	apiFake := &fakeTeamAPI{
		listResp: newTeamCollection(team),
	}

	res := NewTeamResolverCacheable(apiFake, fc, true)

	id, err := res.ResolveTeamRefToID(ctx, "  My Team  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "team-id-xyz" {
		t.Fatalf("expected team-id-xyz from API, got %q", id)
	}

	if fc.getCalls != 1 {
		t.Errorf("expected 1 Get call, got %d", fc.getCalls)
	}
	expectedKey := cacher.NewTeamKeyBuilder("My Team").ToString()
	if fc.lastGetKey != expectedKey {
		t.Errorf("expected cache key %q, got %q", expectedKey, fc.lastGetKey)
	}

	if apiFake.listCalls != 1 {
		t.Errorf("expected 1 ListMyJoined call, got %d", apiFake.listCalls)
	}

	if fc.setCalls != 1 {
		t.Errorf("expected 1 Set call, got %d", fc.setCalls)
	}
	if fc.lastSetKey != expectedKey {
		t.Errorf("expected cache Set key %q, got %q", expectedKey, fc.lastSetKey)
	}
	if v, ok := fc.lastSetValue.(string); !ok || v != "team-id-xyz" {
		t.Errorf("expected cached value 'team-id-xyz', got %#v", fc.lastSetValue)
	}
}

func TestTeamResolverCacheable_ResolveTeamRefToID_CacheDisabled_SkipsCache(t *testing.T) {
	ctx := context.Background()

	fc := &fakeCacher{
		getValue: []string{"team-id-cache"},
		getFound: true,
	}
	team := newGraphTeam("team-id-api", "My Team")
	apiFake := &fakeTeamAPI{
		listResp: newTeamCollection(team),
	}

	res := NewTeamResolverCacheable(apiFake, fc, false)

	id, err := res.ResolveTeamRefToID(ctx, "My Team")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "team-id-api" {
		t.Fatalf("expected team-id-api from API, got %q", id)
	}

	if fc.getCalls != 0 || fc.setCalls != 0 {
		t.Errorf("expected no cache calls when cache disabled, got get=%d set=%d", fc.getCalls, fc.setCalls)
	}
	if apiFake.listCalls != 1 {
		t.Errorf("expected 1 ListMyJoined call, got %d", apiFake.listCalls)
	}
}

func TestTeamResolverCacheable_ResolveTeamRefToID_ListMyJoinedErrorPropagated(t *testing.T) {
	ctx := context.Background()

	apiErr := &sender.RequestError{
		Code:    500,
		Message: "boom",
	}
	apiFake := &fakeTeamAPI{
		listErr: apiErr,
	}
	fc := &fakeCacher{
		getFound: false,
	}

	res := NewTeamResolverCacheable(apiFake, fc, true)

	_, err := res.ResolveTeamRefToID(ctx, "My Team")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "boom") {
		t.Errorf("expected error containing 'boom', got %v", err)
	}
	if apiFake.listCalls != 1 {
		t.Errorf("expected 1 ListMyJoined call, got %d", apiFake.listCalls)
	}
	if fc.setCalls != 0 {
		t.Errorf("expected no cache Set when API fails, got %d", fc.setCalls)
	}
}

func TestResolveTeamIDByName_NoTeamsAvailable(t *testing.T) {
	col := msmodels.NewTeamCollectionResponse()
	col.SetValue(nil)

	_, err := resolveTeamIDByName("X", col)
	if err == nil {
		t.Fatalf("expected error for no teams, got nil")
	}
	if !strings.Contains(err.Error(), "no teams available") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestResolveTeamIDByName_NoMatch(t *testing.T) {
	t1 := newGraphTeam("1", "Alpha")
	col := newTeamCollection(t1)

	_, err := resolveTeamIDByName("Beta", col)
	if err == nil {
		t.Fatalf("expected error for missing team, got nil")
	}
	if !strings.Contains(err.Error(), "team with name") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestResolveTeamIDByName_SingleMatch(t *testing.T) {
	t1 := newGraphTeam("1", "Alpha")
	t2 := newGraphTeam("2", "Beta")
	col := newTeamCollection(t1, t2)

	id, err := resolveTeamIDByName("Beta", col)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "2" {
		t.Fatalf("expected id=2, got %q", id)
	}
}

func TestResolveTeamIDByName_MultipleMatches(t *testing.T) {
	t1 := newGraphTeam("1", "Alpha")
	t2 := newGraphTeam("2", "Alpha")
	col := newTeamCollection(t1, t2)

	_, err := resolveTeamIDByName("Alpha", col)
	if err == nil {
		t.Fatalf("expected error for multiple matches, got nil")
	}
	if !strings.Contains(err.Error(), "multiple teams named") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestIsLikelyGUID_Positive(t *testing.T) {
	s := "123e4567-e89b-12d3-a456-426614174000"
	if !isLikelyGUID(s) {
		t.Fatalf("expected isLikelyGUID(%q)=true, got false", s)
	}
}

func TestIsLikelyGUID_Negative(t *testing.T) {
	for _, s := range []string{
		"", "not-a-guid", "123e4567-e89b-12d3-a456-42661417400", "zzze4567-e89b-12d3-a456-426614174000",
	} {
		if isLikelyGUID(s) {
			t.Fatalf("expected isLikelyGUID(%q)=false, got true", s)
		}
	}
}
