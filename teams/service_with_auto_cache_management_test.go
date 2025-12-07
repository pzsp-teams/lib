package teams

import (
	"context"
	"strings"
	"testing"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/cacher"
	snd "github.com/pzsp-teams/lib/internal/sender"
)

// --- FAKES / HELPERKI -------------------------------------------------------

type fakeCacher struct {
	setCalls        int
	invalidateCalls int

	setKeys        []string
	setValues      []any
	invalidateKeys []string
}

func (f *fakeCacher) Get(key string) (any, bool, error) {
	// Dekorator tego nie używa – cache jest tylko Set/Invalidate.
	return nil, false, nil
}

func (f *fakeCacher) Set(key string, value any) error {
	f.setCalls++
	f.setKeys = append(f.setKeys, key)
	f.setValues = append(f.setValues, value)
	return nil
}

func (f *fakeCacher) Invalidate(key string) error {
	f.invalidateCalls++
	f.invalidateKeys = append(f.invalidateKeys, key)
	return nil
}

func (f *fakeCacher) Clear() error {
	return nil
}

// fakeTeamResolver implementuje resolver.TeamResolver (pole Service.teamResolver)
type fakeTeamResolver struct {
	resolveFunc func(ctx context.Context, teamRef string) (string, error)
	lastRef     string
	calls       int
}

func (f *fakeTeamResolver) ResolveTeamRefToID(ctx context.Context, teamRef string) (string, error) {
	f.calls++
	f.lastRef = teamRef
	if f.resolveFunc != nil {
		return f.resolveFunc(ctx, teamRef)
	}
	return "", nil
}

// fakeTeamAPI implementuje api.TeamAPI (pole Service.teamAPI)
// Podpinamy funkcje, które wykorzystują metody Service.
type fakeTeamAPI struct {
	getFunc              func(ctx context.Context, teamID string) (msmodels.Teamable, *snd.RequestError)
	listMyJoinedFunc     func(ctx context.Context) (msmodels.TeamCollectionResponseable, *snd.RequestError)
	updateFunc           func(ctx context.Context, teamID string, patch *msmodels.Team) (msmodels.Teamable, *snd.RequestError)
	createFromTemplateFn func(ctx context.Context, displayName, description string, owners []string) (string, *snd.RequestError)
	createViaGroupFn     func(ctx context.Context, displayName, mailNickname, visibility string) (string, *snd.RequestError)
	archiveFunc          func(ctx context.Context, teamID string, spoReadOnlyForMembers *bool) *snd.RequestError
	unarchiveFunc        func(ctx context.Context, teamID string) *snd.RequestError
	deleteFunc           func(ctx context.Context, teamID string) *snd.RequestError
	restoreDeletedFunc   func(ctx context.Context, deletedGroupID string) (msmodels.DirectoryObjectable, *snd.RequestError)
}

func (f *fakeTeamAPI) Get(ctx context.Context, teamID string) (msmodels.Teamable, *snd.RequestError) {
	if f.getFunc != nil {
		return f.getFunc(ctx, teamID)
	}
	return nil, nil
}

func (f *fakeTeamAPI) ListMyJoined(ctx context.Context) (msmodels.TeamCollectionResponseable, *snd.RequestError) {
	if f.listMyJoinedFunc != nil {
		return f.listMyJoinedFunc(ctx)
	}
	return nil, nil
}

func (f *fakeTeamAPI) Update(ctx context.Context, teamID string, patch *msmodels.Team) (msmodels.Teamable, *snd.RequestError) {
	if f.updateFunc != nil {
		return f.updateFunc(ctx, teamID, patch)
	}
	return nil, nil
}

func (f *fakeTeamAPI) CreateViaGroup(ctx context.Context, displayName, mailNickname, visibility string) (string, *snd.RequestError) {
	if f.createViaGroupFn != nil {
		return f.createViaGroupFn(ctx, displayName, mailNickname, visibility)
	}
	return "", nil
}

func (f *fakeTeamAPI) CreateFromTemplate(ctx context.Context, displayName, description string, owners []string) (string, *snd.RequestError) {
	if f.createFromTemplateFn != nil {
		return f.createFromTemplateFn(ctx, displayName, description, owners)
	}
	return "", nil
}

func (f *fakeTeamAPI) Archive(ctx context.Context, teamID string, spoReadOnlyForMembers *bool) *snd.RequestError {
	if f.archiveFunc != nil {
		return f.archiveFunc(ctx, teamID, spoReadOnlyForMembers)
	}
	return nil
}

func (f *fakeTeamAPI) Unarchive(ctx context.Context, teamID string) *snd.RequestError {
	if f.unarchiveFunc != nil {
		return f.unarchiveFunc(ctx, teamID)
	}
	return nil
}

func (f *fakeTeamAPI) Delete(ctx context.Context, teamID string) *snd.RequestError {
	if f.deleteFunc != nil {
		return f.deleteFunc(ctx, teamID)
	}
	return nil
}

func (f *fakeTeamAPI) RestoreDeleted(ctx context.Context, deletedGroupID string) (msmodels.DirectoryObjectable, *snd.RequestError) {
	if f.restoreDeletedFunc != nil {
		return f.restoreDeletedFunc(ctx, deletedGroupID)
	}
	return nil, nil
}

// helper: tworzenie grafowego Team
func newGraphTeam(id, name string) msmodels.Teamable {
	t := msmodels.NewTeam()
	t.SetId(&id)
	t.SetDisplayName(&name)
	return t
}

// helper: kolekcja teamów
func newTeamCollection(teams ...msmodels.Teamable) msmodels.TeamCollectionResponseable {
	resp := msmodels.NewTeamCollectionResponse()
	resp.SetValue(teams)
	return resp
}

// helper: DirectoryObject z ID
func newDirectoryObject(id string) msmodels.DirectoryObjectable {
	obj := msmodels.NewDirectoryObject()
	obj.SetId(&id)
	return obj
}

// --- TESTY ------------------------------------------------------------------

// Get: po udanym wywołaniu powinniśmy dodać team do cache (displayName -> ID)
func TestServiceWithAutoCacheManagement_Get_AddsTeamToCacheOnSuccess(t *testing.T) {
	ctx := context.Background()

	fc := &fakeCacher{}
	fr := &fakeTeamResolver{
		resolveFunc: func(ctx context.Context, teamRef string) (string, error) {
			if teamRef != "my-ref" {
				t.Errorf("expected teamRef=my-ref, got %q", teamRef)
			}
			return "team-id-123", nil
		},
	}
	fapi := &fakeTeamAPI{
		getFunc: func(ctx context.Context, teamID string) (msmodels.Teamable, *snd.RequestError) {
			if teamID != "team-id-123" {
				t.Errorf("expected teamID=team-id-123, got %q", teamID)
			}
			return newGraphTeam("team-id-123", "  My Team  "), nil
		},
	}

	svc := &Service{
		teamAPI:      fapi,
		teamResolver: fr,
	}
	decor := NewServiceWithAutoCacheManagement(svc, fc)

	team, err := decor.Get(ctx, "my-ref")
	if err != nil {
		t.Fatalf("unexpected error from Get: %v", err)
	}
	if team == nil {
		t.Fatalf("expected non-nil team")
	}
	// tu wystarczy sprawdzić ID (albo trymowaną nazwę, jeśli chcesz)
	if team.ID != "team-id-123" {
		t.Fatalf("unexpected mapped team: %#v", team)
	}
	// opcjonalnie:
	// if strings.TrimSpace(team.DisplayName) != "My Team" {
	//     t.Fatalf("unexpected mapped team: %#v", team)
	// }

	if fc.setCalls != 1 {
		t.Fatalf("expected 1 Set call, got %d", fc.setCalls)
	}
	expectedKey := cacher.NewTeamKeyBuilder("My Team").ToString()
	if fc.setKeys[0] != expectedKey {
		t.Errorf("expected cache key %q, got %q", expectedKey, fc.setKeys[0])
	}
	if v, ok := fc.setValues[0].(string); !ok || v != "team-id-123" {
		t.Errorf("expected cached value 'team-id-123', got %#v", fc.setValues[0])
	}
}


// ListMyJoined: każdy zwrócony team powinien być dogrzany do cache.
func TestServiceWithAutoCacheManagement_ListMyJoined_WarmsCache(t *testing.T) {
	ctx := context.Background()

	fc := &fakeCacher{}
	fapi := &fakeTeamAPI{
		listMyJoinedFunc: func(ctx context.Context) (msmodels.TeamCollectionResponseable, *snd.RequestError) {
			return newTeamCollection(
				newGraphTeam("id1", " Team One "),
				newGraphTeam("id2", "Team Two"),
			), nil
		},
	}

	svc := &Service{
		teamAPI: fapi,
	}
	decor := NewServiceWithAutoCacheManagement(svc, fc)

	teams, err := decor.ListMyJoined(ctx)
	if err != nil {
		t.Fatalf("unexpected error from ListMyJoined: %v", err)
	}
	if len(teams) != 2 {
		t.Fatalf("expected 2 teams, got %d", len(teams))
	}

	if fc.setCalls != 2 {
		t.Fatalf("expected 2 Set calls, got %d", fc.setCalls)
	}

	expectedKeys := []string{
		cacher.NewTeamKeyBuilder("Team One").ToString(),
		cacher.NewTeamKeyBuilder("Team Two").ToString(),
	}
	if len(fc.setKeys) != len(expectedKeys) {
		t.Fatalf("expected %d keys, got %d", len(expectedKeys), len(fc.setKeys))
	}
	for i, k := range expectedKeys {
		if fc.setKeys[i] != k {
			t.Errorf("at index %d expected key %q, got %q", i, k, fc.setKeys[i])
		}
	}
}

// Update: dla nazwy (nie GUID) – invaliduje stary ref + dodaje nową nazwę do cache.
func TestServiceWithAutoCacheManagement_Update_InvalidatesOldAndCachesNew(t *testing.T) {
	ctx := context.Background()

	fc := &fakeCacher{}
	fr := &fakeTeamResolver{
		resolveFunc: func(ctx context.Context, teamRef string) (string, error) {
			if strings.TrimSpace(teamRef) != "Old Name" {
				t.Errorf("expected teamRef Old Name, got %q", teamRef)
			}
			return "team-id-xyz", nil
		},
	}
	fapi := &fakeTeamAPI{
		updateFunc: func(ctx context.Context, teamID string, patch *msmodels.Team) (msmodels.Teamable, *snd.RequestError) {
			if teamID != "team-id-xyz" {
				t.Errorf("expected teamID team-id-xyz, got %q", teamID)
			}
			return newGraphTeam("team-id-xyz", "  New Name  "), nil
		},
	}

	svc := &Service{
		teamAPI:      fapi,
		teamResolver: fr,
	}
	decor := NewServiceWithAutoCacheManagement(svc, fc)

	team, err := decor.Update(ctx, "  Old Name  ", nil)
	if err != nil {
		t.Fatalf("unexpected error from Update: %v", err)
	}
	if team == nil || team.ID != "team-id-xyz" || strings.TrimSpace(team.DisplayName) != "New Name" {
		t.Fatalf("unexpected team returned: %#v", team)
	}


	if fc.invalidateCalls != 1 {
		t.Fatalf("expected 1 Invalidate call, got %d", fc.invalidateCalls)
	}
	expectedInvalidKey := cacher.NewTeamKeyBuilder("Old Name").ToString()
	if fc.invalidateKeys[0] != expectedInvalidKey {
		t.Errorf("expected invalidate key %q, got %q", expectedInvalidKey, fc.invalidateKeys[0])
	}

	if fc.setCalls != 1 {
		t.Fatalf("expected 1 Set call, got %d", fc.setCalls)
	}
	expectedSetKey := cacher.NewTeamKeyBuilder("New Name").ToString()
	if fc.setKeys[0] != expectedSetKey {
		t.Errorf("expected Set key %q, got %q", expectedSetKey, fc.setKeys[0])
	}
	if v, ok := fc.setValues[0].(string); !ok || v != "team-id-xyz" {
		t.Errorf("expected Set value 'team-id-xyz', got %#v", fc.setValues[0])
	}
}

// Update: gdy teamRef jest GUID–em, nie powinno być invalidacji, tylko Set.
func TestServiceWithAutoCacheManagement_Update_DoesNotInvalidateForGUID(t *testing.T) {
	ctx := context.Background()

	guidRef := "123e4567-e89b-12d3-a456-426614174000"

	fc := &fakeCacher{}
	fr := &fakeTeamResolver{
		resolveFunc: func(ctx context.Context, teamRef string) (string, error) {
			return "team-id-guid", nil
		},
	}
	fapi := &fakeTeamAPI{
		updateFunc: func(ctx context.Context, teamID string, patch *msmodels.Team) (msmodels.Teamable, *snd.RequestError) {
			return newGraphTeam("team-id-guid", "Name From API"), nil
		},
	}

	svc := &Service{
		teamAPI:      fapi,
		teamResolver: fr,
	}
	decor := NewServiceWithAutoCacheManagement(svc, fc)

	_, err := decor.Update(ctx, guidRef, nil)
	if err != nil {
		t.Fatalf("unexpected error from Update: %v", err)
	}

	if fc.invalidateCalls != 0 {
		t.Fatalf("expected 0 Invalidate calls for GUID ref, got %d", fc.invalidateCalls)
	}
	if fc.setCalls != 1 {
		t.Fatalf("expected 1 Set call, got %d", fc.setCalls)
	}
}

// CreateFromTemplate: po sukcesie tylko Invalidate(displayName), bez Set.
func TestServiceWithAutoCacheManagement_CreateFromTemplate_InvalidatesByName(t *testing.T) {
	ctx := context.Background()

	fc := &fakeCacher{}
	fapi := &fakeTeamAPI{
		createFromTemplateFn: func(ctx context.Context, displayName, description string, owners []string) (string, *snd.RequestError) {
			if strings.TrimSpace(displayName) != "My Team" {
				t.Errorf("expected displayName My Team, got %q", displayName)
			}
			return "new-team-id", nil
		},
	}

	svc := &Service{
		teamAPI: fapi,
	}
	decor := NewServiceWithAutoCacheManagement(svc, fc)

	id, err := decor.CreateFromTemplate(ctx, "  My Team  ", "desc", []string{"owner"})
	if err != nil {
		t.Fatalf("unexpected error from CreateFromTemplate: %v", err)
	}
	if id != "new-team-id" {
		t.Fatalf("expected id=new-team-id, got %q", id)
	}

	if fc.invalidateCalls != 1 {
		t.Fatalf("expected 1 Invalidate call, got %d", fc.invalidateCalls)
	}
	expectedKey := cacher.NewTeamKeyBuilder("My Team").ToString()
	if fc.invalidateKeys[0] != expectedKey {
		t.Errorf("expected invalidate key %q, got %q", expectedKey, fc.invalidateKeys[0])
	}
	if fc.setCalls != 0 {
		t.Fatalf("expected 0 Set calls, got %d", fc.setCalls)
	}
}

// CreateViaGroup: po sukcesie tylko Invalidate(displayName), bez Set.
func TestServiceWithAutoCacheManagement_CreateViaGroup_InvalidatesByName(t *testing.T) {
	ctx := context.Background()

	fc := &fakeCacher{}
	fapi := &fakeTeamAPI{
		createViaGroupFn: func(ctx context.Context, displayName, mailNickname, visibility string) (string, *snd.RequestError) {
			if strings.TrimSpace(displayName) != "My Team" {
				t.Errorf("expected displayName My Team, got %q", displayName)
			}
			return "created-id", nil
		},
		getFunc: func(ctx context.Context, teamID string) (msmodels.Teamable, *snd.RequestError) {
			if teamID != "created-id" {
				t.Errorf("expected Get with created-id, got %q", teamID)
			}
			return newGraphTeam("created-id", "My Team"), nil
		},
	}

	svc := &Service{
		teamAPI: fapi,
	}
	decor := NewServiceWithAutoCacheManagement(svc, fc)

	team, err := decor.CreateViaGroup(ctx, "  My Team  ", "nick", "Public")
	if err != nil {
		t.Fatalf("unexpected error from CreateViaGroup: %v", err)
	}
	if team == nil || team.ID != "created-id" {
		t.Fatalf("unexpected team: %#v", team)
	}

	if fc.invalidateCalls != 1 {
		t.Fatalf("expected 1 Invalidate call, got %d", fc.invalidateCalls)
	}
	expectedKey := cacher.NewTeamKeyBuilder("My Team").ToString()
	if fc.invalidateKeys[0] != expectedKey {
		t.Errorf("expected invalidate key %q, got %q", expectedKey, fc.invalidateKeys[0])
	}
	if fc.setCalls != 0 {
		t.Fatalf("expected 0 Set calls, got %d", fc.setCalls)
	}
}

// Archive: dla nazwy (nie GUID) – po sukcesie invalidacja nazwy.
func TestServiceWithAutoCacheManagement_Archive_InvalidatesByName(t *testing.T) {
	ctx := context.Background()

	fc := &fakeCacher{}
	fr := &fakeTeamResolver{
		resolveFunc: func(ctx context.Context, teamRef string) (string, error) {
			return "team-id-1", nil
		},
	}
	fapi := &fakeTeamAPI{
		archiveFunc: func(ctx context.Context, teamID string, spoReadOnlyForMembers *bool) *snd.RequestError {
			if teamID != "team-id-1" {
				t.Errorf("expected teamID team-id-1, got %q", teamID)
			}
			return nil
		},
	}

	svc := &Service{
		teamAPI:      fapi,
		teamResolver: fr,
	}
	decor := NewServiceWithAutoCacheManagement(svc, fc)

	if err := decor.Archive(ctx, "  My Team  ", nil); err != nil {
		t.Fatalf("unexpected error from Archive: %v", err)
	}

	if fc.invalidateCalls != 1 {
		t.Fatalf("expected 1 Invalidate call, got %d", fc.invalidateCalls)
	}
	expectedKey := cacher.NewTeamKeyBuilder("My Team").ToString()
	if fc.invalidateKeys[0] != expectedKey {
		t.Errorf("expected invalidate key %q, got %q", expectedKey, fc.invalidateKeys[0])
	}
	if fc.setCalls != 0 {
		t.Fatalf("expected 0 Set calls, got %d", fc.setCalls)
	}
}

// Delete: gdy ref to GUID, nie powinno być invalidacji.
func TestServiceWithAutoCacheManagement_Delete_DoesNotInvalidateForGUID(t *testing.T) {
	ctx := context.Background()

	guidRef := "123e4567-e89b-12d3-a456-426614174000"

	fc := &fakeCacher{}
	fr := &fakeTeamResolver{
		resolveFunc: func(ctx context.Context, teamRef string) (string, error) {
			return "team-id-guid", nil
		},
	}
	fapi := &fakeTeamAPI{
		deleteFunc: func(ctx context.Context, teamID string) *snd.RequestError {
			if teamID != "team-id-guid" {
				t.Errorf("expected teamID team-id-guid, got %q", teamID)
			}
			return nil
		},
	}

	svc := &Service{
		teamAPI:      fapi,
		teamResolver: fr,
	}
	decor := NewServiceWithAutoCacheManagement(svc, fc)

	if err := decor.Delete(ctx, guidRef); err != nil {
		t.Fatalf("unexpected error from Delete: %v", err)
	}

	if fc.invalidateCalls != 0 {
		t.Fatalf("expected 0 Invalidate calls for GUID ref, got %d", fc.invalidateCalls)
	}
	if fc.setCalls != 0 {
		t.Fatalf("expected 0 Set calls, got %d", fc.setCalls)
	}
}

// RestoreDeleted: dekorator tylko deleguje – nie powinien dotykać cache.
func TestServiceWithAutoCacheManagement_RestoreDeleted_DoesNotTouchCache(t *testing.T) {
	ctx := context.Background()

	fc := &fakeCacher{}
	fapi := &fakeTeamAPI{
		restoreDeletedFunc: func(ctx context.Context, deletedGroupID string) (msmodels.DirectoryObjectable, *snd.RequestError) {
			return newDirectoryObject("restored-id"), nil
		},
	}

	svc := &Service{
		teamAPI: fapi,
	}
	decor := NewServiceWithAutoCacheManagement(svc, fc)

	id, err := decor.RestoreDeleted(ctx, "deleted-123")
	if err != nil {
		t.Fatalf("unexpected error from RestoreDeleted: %v", err)
	}
	if id != "restored-id" {
		t.Fatalf("expected id=restored-id, got %q", id)
	}

	if fc.setCalls != 0 || fc.invalidateCalls != 0 {
		t.Fatalf("expected no cache operations in RestoreDeleted, got set=%d invalidates=%d",
			fc.setCalls, fc.invalidateCalls)
	}
}
