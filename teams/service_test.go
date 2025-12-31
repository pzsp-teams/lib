package teams

import (
	"context"
	"errors"
	"testing"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	sender "github.com/pzsp-teams/lib/internal/sender"
)

type fakeResolver struct {
	lastTeamName string
	resolverErr  error
}

func (m *fakeResolver) ResolveTeamRefToID(ctx context.Context, teamName string) (string, error) {
	m.lastTeamName = teamName
	if m.resolverErr != nil {
		return "", m.resolverErr
	}
	return teamName, nil
}

func (m *fakeResolver) MapChannelRefToChannelID(ctx context.Context, teamID, channelName string) (string, error) {
	return channelName, nil
}

func (m *fakeResolver) MapUserRefToMemberID(ctx context.Context, userRef, teamID, channelID string) (string, error) {
	return userRef, nil
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

func TestService_Delete_MapsError(t *testing.T) {
	ctx := context.Background()
	api := &fakeTeamsAPI{
		deleteErr: &sender.RequestError{
			Code:    403,
			Message: "nope",
		},
	}
	svc := NewService(api, &fakeResolver{})

	err := svc.Delete(ctx, "MyTeam")
	var forbidden sender.ErrAccessForbidden
	if !errors.As(err, &forbidden) {
		t.Fatalf("expected ErrAccessForbidden, got %v", err)
	}
}

func TestService_Update_MapsError(t *testing.T) {
	ctx := context.Background()
	api := &fakeTeamsAPI{
		updateErr: &sender.RequestError{
			Code:    404,
			Message: "no such team",
		},
	}
	svc := NewService(api, &fakeResolver{})

	_, err := svc.Update(ctx, "missing-team", msmodels.NewTeam())
	var notFound sender.ErrResourceNotFound
	if !errors.As(err, &notFound) {
		t.Fatalf("expected ErrResourceNotFound, got %v", err)
	}
}

func TestService_CreateViaGroup_MapsCreateError(t *testing.T) {
	ctx := context.Background()
	api := &fakeTeamsAPI{
		createViaGroupErr: &sender.RequestError{
			Code:    403,
			Message: "nope",
		},
	}
	svc := NewService(api, &fakeResolver{})

	_, err := svc.CreateViaGroup(ctx, "X", "x", "public")
	var forbidden sender.ErrAccessForbidden
	if !errors.As(err, &forbidden) {
		t.Fatalf("expected ErrAccessForbidden, got %v", err)
	}
}

func TestService_CreateViaGroup_MapsGetError(t *testing.T) {
	ctx := context.Background()
	api := &fakeTeamsAPI{
		createViaGroupID: "team-xyz",
		getErr: &sender.RequestError{
			Code:    404,
			Message: "not ready",
		},
	}
	svc := NewService(api, &fakeResolver{})

	_, err := svc.CreateViaGroup(ctx, "X", "x", "public")
	var notFound sender.ErrResourceNotFound
	if !errors.As(err, &notFound) {
		t.Fatalf("expected ErrResourceNotFound, got %v", err)
	}
}

func TestService_CreateFromTemplate_ReturnsID(t *testing.T) {
	ctx := context.Background()
	api := &fakeTeamsAPI{
		createFromTemplateID: "tmpl-123",
	}
	svc := NewService(api, &fakeResolver{})

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
			Code:    403,
			Message: "nope",
		},
	}
	svc := NewService(api, &fakeResolver{})

	_, err := svc.CreateFromTemplate(ctx, "Tpl", "Desc", nil)
	var forbidden sender.ErrAccessForbidden
	if !errors.As(err, &forbidden) {
		t.Fatalf("expected ErrAccessForbidden, got %v", err)
	}
}

func TestService_Archive_Success(t *testing.T) {
	ctx := context.Background()
	api := &fakeTeamsAPI{}
	svc := NewService(api, &fakeResolver{})
	readOnlyForMembers := false
	if err := svc.Archive(ctx, "T1", &readOnlyForMembers); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestService_Archive_MapsError(t *testing.T) {
	ctx := context.Background()
	api := &fakeTeamsAPI{
		archiveErr: &sender.RequestError{
			Code:    403,
			Message: "nope",
		},
	}
	svc := NewService(api, &fakeResolver{})

	readOnlyForMembers := false
	err := svc.Archive(ctx, "T1", &readOnlyForMembers)
	var forbidden sender.ErrAccessForbidden
	if !errors.As(err, &forbidden) {
		t.Fatalf("expected ErrAccessForbidden, got %v", err)
	}
}

func TestService_Unarchive_Success(t *testing.T) {
	ctx := context.Background()
	api := &fakeTeamsAPI{}
	svc := NewService(api, &fakeResolver{})

	if err := svc.Unarchive(ctx, "T1"); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestService_Unarchive_MapsError(t *testing.T) {
	ctx := context.Background()
	api := &fakeTeamsAPI{
		unarchiveErr: &sender.RequestError{
			Code:    403,
			Message: "nope",
		},
	}
	svc := NewService(api, &fakeResolver{})

	err := svc.Unarchive(ctx, "T1")
	var forbidden sender.ErrAccessForbidden
	if !errors.As(err, &forbidden) {
		t.Fatalf("expected ErrAccessForbidden, got %v", err)
	}
}

func TestService_RestoreDeleted_ReturnsID(t *testing.T) {
	ctx := context.Background()
	obj := msmodels.NewDirectoryObject()
	id := "restored-id"
	obj.SetId(&id)

	api := &fakeTeamsAPI{restoreObj: obj}
	svc := NewService(api, &fakeResolver{})

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
			Code:    404,
			Message: "missing",
		},
	}
	svc := NewService(api, &fakeResolver{})

	_, err := svc.RestoreDeleted(ctx, "deleted-1")
	var notFound sender.ErrResourceNotFound
	if !errors.As(err, &notFound) {
		t.Fatalf("expected ErrResourceNotFound, got %v", err)
	}
}

func TestService_RestoreDeleted_EmptyObjectReturnsUnknown(t *testing.T) {
	ctx := context.Background()
	obj := msmodels.NewDirectoryObject()
	api := &fakeTeamsAPI{restoreObj: obj}
	svc := NewService(api, &fakeResolver{})

	_, err := svc.RestoreDeleted(ctx, "deleted-1")
	if err == nil {
		t.Fatalf("expected error for empty restored object, got nil")
	}
}
