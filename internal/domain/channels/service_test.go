package channels

import (
	"context"
	"errors"
	"testing"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	sender "github.com/pzsp-teams/lib/internal/sender"
)

type fakeChannelAPI struct {
	listResp   msmodels.ChannelCollectionResponseable
	listErr    *sender.RequestError
	getResp    msmodels.Channelable
	getErr     *sender.RequestError
	createResp msmodels.Channelable
	createErr  *sender.RequestError
	deleteErr  *sender.RequestError
	lastCreate msmodels.Channelable
	lastTeamID string
	lastChanID string
}

func (f *fakeChannelAPI) ListChannels(ctx context.Context, teamID string) (msmodels.ChannelCollectionResponseable, *sender.RequestError) {
	f.lastTeamID = teamID
	return f.listResp, f.listErr
}

func (f *fakeChannelAPI) GetChannel(ctx context.Context, teamID, channelID string) (msmodels.Channelable, *sender.RequestError) {
	f.lastTeamID = teamID
	f.lastChanID = channelID
	return f.getResp, f.getErr
}

func (f *fakeChannelAPI) CreateChannel(ctx context.Context, teamID string, channel msmodels.Channelable) (msmodels.Channelable, *sender.RequestError) {
	f.lastTeamID = teamID
	f.lastCreate = channel
	return f.createResp, f.createErr
}

func (f *fakeChannelAPI) DeleteChannel(ctx context.Context, teamID, channelID string) *sender.RequestError {
	f.lastTeamID = teamID
	f.lastChanID = channelID
	return f.deleteErr
}

func newGraphChannel(id, name string) msmodels.Channelable {
	channel := msmodels.NewChannel()
	channel.SetId(&id)
	channel.SetDisplayName(&name)
	return channel
}

func TestService_ListChannels_MapsFieldsAndGeneralFlag(t *testing.T) {
	ctx := context.Background()
	col := msmodels.NewChannelCollectionResponse()

	ch1 := newGraphChannel("1", "General")
	ch2 := newGraphChannel("2", "Random")

	col.SetValue([]msmodels.Channelable{ch1, ch2})

	api := &fakeChannelAPI{listResp: col}
	svc := NewService(api)

	got, err := svc.ListChannels(ctx, "team-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 channels, got %d", len(got))
	}

	if got[0].ID != "1" || got[0].Name != "General" || !got[0].IsGeneral {
		t.Errorf("unexpected first channel: %+v", got[0])
	}
	if got[1].ID != "2" || got[1].Name != "Random" || got[1].IsGeneral {
		t.Errorf("unexpected second channel: %+v", got[1])
	}
}

func TestService_ListChannels_MapsErrors(t *testing.T) {
	ctx := context.Background()
	api := &fakeChannelAPI{
		listErr: &sender.RequestError{
			Code:    "ResourceNotFound",
			Message: "team not found",
		},
	}
	svc := NewService(api)

	_, err := svc.ListChannels(ctx, "non-existing-team")
	if !errors.Is(err, ErrChannelNotFound) {
		t.Fatalf("expected ErrChannelNotFound, got %v", err)
	}
}

func TestService_Get_MapsSingleChannel(t *testing.T) {
	ctx := context.Background()
	ch := newGraphChannel("42", "General")
	api := &fakeChannelAPI{getResp: ch}
	svc := NewService(api)

	got, err := svc.Get(ctx, "team-1", "42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.ID != "42" || got.Name != "General" || !got.IsGeneral {
		t.Errorf("unexpected channel: %+v", got)
	}
}

func TestService_Create_SetsNameAndMapsResult(t *testing.T) {
	ctx := context.Background()
	created := newGraphChannel("123", "my-channel")

	api := &fakeChannelAPI{
		createResp: created,
	}
	svc := NewService(api)

	got, err := svc.Create(ctx, "team-1", "my-channel")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.ID != "123" || got.Name != "my-channel" {
		t.Errorf("unexpected result: %+v", got)
	}
	if got.IsGeneral {
		t.Errorf("expected IsGeneral=false for created channel, got true")
	}

	dn := api.lastCreate.GetDisplayName()
	if dn == nil || *dn != "my-channel" {
		t.Errorf("expected displayName 'my-channel', got %#v", dn)
	}
}

func TestService_Delete_MapsError(t *testing.T) {
	ctx := context.Background()
	api := &fakeChannelAPI{
		deleteErr: &sender.RequestError{
			Code:    "AccessDenied",
			Message: "nope",
		},
	}
	svc := NewService(api)

	err := svc.Delete(ctx, "team-1", "chan-1")
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestDeref_NilReturnsEmpty(t *testing.T) {
	if got := deref(nil); got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
}

func TestDeref_NonNil(t *testing.T) {
	s := "hello"
	if got := deref(&s); got != "hello" {
		t.Fatalf("expected 'hello', got %q", got)
	}
}
