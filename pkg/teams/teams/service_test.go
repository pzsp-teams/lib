package teams

import (
	"context"
	"errors"
	"testing"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	sender "github.com/pzsp-teams/lib/internal/sender"
)

type fakeMapper struct {
	lastTeamName string
	mapErr       error
}

func (m *fakeMapper) MapTeamNameToTeamID(ctx context.Context, teamName string) (string, error) {
	m.lastTeamName = teamName
	if m.mapErr != nil {
		return "", m.mapErr
	}
	return teamName, nil
}

func (m *fakeMapper) MapChannelNameToChannelID(ctx context.Context, teamID, channelName string) (string, error) {
	return channelName, nil
}

type fakeTeamsAPI struct {
	getResp    msmodels.Teamable
	getErr     *sender.RequestError
	listResp   msmodels.TeamCollectionResponseable
	listErr    *sender.RequestError
	updateResp msmodels.Teamable
	updateErr  *sender.RequestError

	createViaGroupID  string
	createViaGroupErr *sender.RequestError

	createFromTemplateID  string
	createFromTemplateErr *sender.RequestError

	archiveErr   *sender.RequestError
	unarchiveErr *sender.RequestError
	deleteErr    *sender.RequestError

	restoreObj msmodels.DirectoryObjectable
	restoreErr *sender.RequestError
}

func (f *fakeTeamsAPI) CreateFromTemplate(ctx context.Context, displayName, description string, owners []string) (string, *sender.RequestError) {
	return f.createFromTemplateID, f.createFromTemplateErr
}

func (f *fakeTeamsAPI) CreateViaGroup(ctx context.Context, displayName, mailNickname, visibility string) (string, *sender.RequestError) {
	return f.createViaGroupID, f.createViaGroupErr
}

func (f *fakeTeamsAPI) Get(ctx context.Context, teamID string) (msmodels.Teamable, *sender.RequestError) {
	return f.getResp, f.getErr
}

func (f *fakeTeamsAPI) ListMyJoined(ctx context.Context) (msmodels.TeamCollectionResponseable, *sender.RequestError) {
	return f.listResp, f.listErr
}

func (f *fakeTeamsAPI) Update(ctx context.Context, teamID string, patch *msmodels.Team) (msmodels.Teamable, *sender.RequestError) {
	return f.updateResp, f.updateErr
}

func (f *fakeTeamsAPI) Archive(ctx context.Context, teamID string, spoReadOnlyForMembers *bool) *sender.RequestError {
	return f.archiveErr
}

func (f *fakeTeamsAPI) Unarchive(ctx context.Context, teamID string) *sender.RequestError {
	return f.unarchiveErr
}

func (f *fakeTeamsAPI) Delete(ctx context.Context, teamID string) *sender.RequestError {
	return f.deleteErr
}

func (f *fakeTeamsAPI) RestoreDeleted(ctx context.Context, deletedGroupID string) (msmodels.DirectoryObjectable, *sender.RequestError) {
	return f.restoreObj, f.restoreErr
}

func newGraphTeam(id, name string) msmodels.Teamable {
	t := msmodels.NewTeam()
	t.SetId(&id)
	t.SetDisplayName(&name)
	return t
}

func TestService_ListMyJoined_MapsTeams(t *testing.T) {
	ctx := context.Background()
	col := msmodels.NewTeamCollectionResponse()
	a := newGraphTeam("1", "Alpha")
	b := newGraphTeam("2", "Beta")
	col.SetValue([]msmodels.Teamable{a, b})

	api := &fakeTeamsAPI{listResp: col}
	svc := NewService(api, &fakeMapper{})

	got, err := svc.ListMyJoined(ctx)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(got) != 2 || got[0].ID != "1" || got[1].DisplayName != "Beta" {
		t.Fatalf("bad mapping: %#v", got)
	}
}

func TestService_Get_MapsTeam(t *testing.T) {
	ctx := context.Background()
	api := &fakeTeamsAPI{getResp: newGraphTeam("42", "X")}
	m := &fakeMapper{}
	svc := NewService(api, m)

	got, err := svc.Get(ctx, "team-name-42")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got.ID != "42" || got.DisplayName != "X" {
		t.Fatalf("bad mapping: %#v", got)
	}
	if m.lastTeamName != "team-name-42" {
		t.Errorf("expected mapper called with 'team-name-42', got %q", m.lastTeamName)
	}
}

func TestService_Delete_MapsError(t *testing.T) {
	ctx := context.Background()
	api := &fakeTeamsAPI{deleteErr: &sender.RequestError{Code: "AccessDenied", Message: "nope"}}
	svc := NewService(api, &fakeMapper{})

	if err := svc.Delete(ctx, "MyTeam"); !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected errForbidden, got %v", err)
	}
}

func TestService_Update_MapsTeam(t *testing.T) {
	ctx := context.Background()
	api := &fakeTeamsAPI{updateResp: newGraphTeam("7", "Updated")}
	svc := NewService(api, &fakeMapper{})

	patch := msmodels.NewTeam()
	got, err := svc.Update(ctx, "MyTeam", patch)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got.ID != "7" || got.DisplayName != "Updated" {
		t.Fatalf("bad mapping: %#v", got)
	}
}

func TestService_Update_MapsError(t *testing.T) {
	ctx := context.Background()
	api := &fakeTeamsAPI{
		updateErr: &sender.RequestError{
			Code:    "ResourceNotFound",
			Message: "no such team",
		},
	}
	svc := NewService(api, &fakeMapper{})

	_, err := svc.Update(ctx, "missing-team", msmodels.NewTeam())
	if !errors.Is(err, ErrTeamNotFound) {
		t.Fatalf("expected errTeamNotFound, got %v", err)
	}
}

func TestService_CreateViaGroup_MapsTeamAfterCreation(t *testing.T) {
	ctx := context.Background()
	api := &fakeTeamsAPI{
		createViaGroupID: "team-123",
		getResp:          newGraphTeam("team-123", "My team"),
	}
	svc := NewService(api, &fakeMapper{})

	got, err := svc.CreateViaGroup(ctx, "My team", "myteam", "public")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got.ID != "team-123" || got.DisplayName != "My team" {
		t.Fatalf("bad mapping: %#v", got)
	}
}

func TestService_CreateViaGroup_MapsCreateError(t *testing.T) {
	ctx := context.Background()
	api := &fakeTeamsAPI{
		createViaGroupErr: &sender.RequestError{
			Code:    "AccessDenied",
			Message: "nope",
		},
	}
	svc := NewService(api, &fakeMapper{})

	_, err := svc.CreateViaGroup(ctx, "X", "x", "public")
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected errForbidden, got %v", err)
	}
}

func TestService_CreateViaGroup_MapsGetError(t *testing.T) {
	ctx := context.Background()
	api := &fakeTeamsAPI{
		createViaGroupID: "team-xyz",
		getErr: &sender.RequestError{
			Code:    "ResourceNotFound",
			Message: "not ready",
		},
	}
	svc := NewService(api, &fakeMapper{})

	_, err := svc.CreateViaGroup(ctx, "X", "x", "public")
	if !errors.Is(err, ErrTeamNotFound) {
		t.Fatalf("expected errTeamNotFound, got %v", err)
	}
}

func TestService_CreateFromTemplate_ReturnsID(t *testing.T) {
	ctx := context.Background()
	api := &fakeTeamsAPI{
		createFromTemplateID: "tmpl-123",
	}
	svc := NewService(api, &fakeMapper{})

	got, err := svc.CreateFromTemplate(ctx, "Tpl", "Desc", nil)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got != "tmpl-123" {
		t.Fatalf("expected id tmpl-123, got %q", got)
	}
}

func TestService_CreateFromTemplate_MapsError(t *testing.T) {
	ctx := context.Background()
	api := &fakeTeamsAPI{
		createFromTemplateErr: &sender.RequestError{
			Code:    "AccessDenied",
			Message: "nope",
		},
	}
	svc := NewService(api, &fakeMapper{})

	_, err := svc.CreateFromTemplate(ctx, "Tpl", "Desc", nil)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected errForbidden, got %v", err)
	}
}

func TestService_Archive_Success(t *testing.T) {
	ctx := context.Background()
	api := &fakeTeamsAPI{}
	svc := NewService(api, &fakeMapper{})
	readOnlyForMembers := false
	if err := svc.Archive(ctx, "T1", &readOnlyForMembers); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestService_Archive_MapsError(t *testing.T) {
	ctx := context.Background()
	api := &fakeTeamsAPI{
		archiveErr: &sender.RequestError{
			Code:    "AccessDenied",
			Message: "nope",
		},
	}
	svc := NewService(api, &fakeMapper{})

	readOnlyForMembers := false
	if err := svc.Archive(ctx, "T1", &readOnlyForMembers); !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected errForbidden, got %v", err)
	}
}

func TestService_Unarchive_Success(t *testing.T) {
	ctx := context.Background()
	api := &fakeTeamsAPI{}
	svc := NewService(api, &fakeMapper{})

	if err := svc.Unarchive(ctx, "T1"); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestService_Unarchive_MapsError(t *testing.T) {
	ctx := context.Background()
	api := &fakeTeamsAPI{
		unarchiveErr: &sender.RequestError{
			Code:    "AccessDenied",
			Message: "nope",
		},
	}
	svc := NewService(api, &fakeMapper{})

	if err := svc.Unarchive(ctx, "T1"); !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected errForbidden, got %v", err)
	}
}

func TestService_RestoreDeleted_ReturnsID(t *testing.T) {
	ctx := context.Background()
	obj := msmodels.NewDirectoryObject()
	id := "restored-id"
	obj.SetId(&id)

	api := &fakeTeamsAPI{restoreObj: obj}
	svc := NewService(api, &fakeMapper{})

	got, err := svc.RestoreDeleted(ctx, "deleted-1")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got != id {
		t.Fatalf("expected %q, got %q", id, got)
	}
}

func TestService_RestoreDeleted_MapsError(t *testing.T) {
	ctx := context.Background()
	api := &fakeTeamsAPI{
		restoreErr: &sender.RequestError{
			Code:    "NotFound",
			Message: "missing",
		},
	}
	svc := NewService(api, &fakeMapper{})

	_, err := svc.RestoreDeleted(ctx, "deleted-1")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected errNotFound, got %v", err)
	}
}

func TestService_RestoreDeleted_EmptyObjectReturnsUnknown(t *testing.T) {
	ctx := context.Background()
	obj := msmodels.NewDirectoryObject()
	api := &fakeTeamsAPI{restoreObj: obj}
	svc := NewService(api, &fakeMapper{})

	_, err := svc.RestoreDeleted(ctx, "deleted-1")
	if !errors.Is(err, ErrUnknown) {
		t.Fatalf("expected errUnknown, got %v", err)
	}
}

func TestDeref_NilReturnsEmpty_Teams(t *testing.T) {
	if got := deref(nil); got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
}

func TestDeref_NonNil_Teams(t *testing.T) {
	s := "hello"
	if got := deref(&s); got != "hello" {
		t.Fatalf("expected 'hello', got %q", got)
	}
}
