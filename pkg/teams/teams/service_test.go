package teams

import (
	"context"
	"errors"
	"testing"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	sender "github.com/pzsp-teams/lib/internal/sender"
)


type fakeTeamsAPI struct {
	getResp    msmodels.Teamable
	getErr     *sender.RequestError
	listResp   msmodels.TeamCollectionResponseable
	listErr    *sender.RequestError
	updateResp msmodels.Teamable
	updateErr  *sender.RequestError

	createViaGroupID string
	createViaGroupErr *sender.RequestError

	createFromTemplateID string
	createFromTemplateErr *sender.RequestError

	archiveErr   *sender.RequestError
	unarchiveErr *sender.RequestError
	deleteErr    *sender.RequestError

	restoreObj   msmodels.DirectoryObjectable
	restoreErr   *sender.RequestError
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
	svc := NewService(api)

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
	svc := NewService(api)
	got, err := svc.Get(ctx, "42")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got.ID != "42" || got.DisplayName != "X" {
		t.Fatalf("bad mapping: %#v", got)
	}
}

func TestService_Delete_MapsError(t *testing.T) {
	ctx := context.Background()
	api := &fakeTeamsAPI{deleteErr: &sender.RequestError{Code: "AccessDenied", Message: "nope"}}
	svc := NewService(api)
	if err := svc.Delete(ctx, "T1"); !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}
