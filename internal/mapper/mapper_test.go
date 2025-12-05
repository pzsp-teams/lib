package mapper

import (
	"context"
	"strings"
	"testing"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/internal/sender"
)

type fakeTeamsAPI struct {
	listResp msmodels.TeamCollectionResponseable
	listErr  *sender.RequestError
}

func (f *fakeTeamsAPI) ListMyJoined(ctx context.Context) (msmodels.TeamCollectionResponseable, *sender.RequestError) {
	return f.listResp, f.listErr
}

func (f *fakeTeamsAPI) Archive(ctx context.Context, teamID string, shouldSetSpoSiteReadOnlyForMembers *bool) *sender.RequestError {
	return nil
}

func (f *fakeTeamsAPI) Unarchive(ctx context.Context, teamID string) *sender.RequestError {
	return nil
}

func (f *fakeTeamsAPI) CreateFromTemplate(ctx context.Context, displayName, description string, template []string) (string, *sender.RequestError) {
	return "", nil
}

func (f *fakeTeamsAPI) CreateViaGroup(ctx context.Context, groupID, displayName, description string) (string, *sender.RequestError) {
	return "", nil
}

func (f *fakeTeamsAPI) Delete(ctx context.Context, teamID string) *sender.RequestError {
	return nil
}

func (f *fakeTeamsAPI) Get(ctx context.Context, teamID string) (msmodels.Teamable, *sender.RequestError) {
	return nil, nil
}

func (f *fakeTeamsAPI) RestoreDeleted(ctx context.Context, teamID string) (msmodels.DirectoryObjectable, *sender.RequestError) {
	return nil, nil
}

func (f *fakeTeamsAPI) Update(ctx context.Context, teamID string, team *msmodels.Team) (msmodels.Teamable, *sender.RequestError) {
	return nil, nil
}

type fakeChannelAPI struct {
	listResp msmodels.ChannelCollectionResponseable
	listErr  *sender.RequestError
}

func (f *fakeChannelAPI) ListChannels(ctx context.Context, teamID string) (msmodels.ChannelCollectionResponseable, *sender.RequestError) {
	return f.listResp, f.listErr
}

func (f *fakeChannelAPI) CreateChannel(ctx context.Context, teamID string, channel msmodels.Channelable) (msmodels.Channelable, *sender.RequestError) {
	return nil, nil
}

func (f *fakeChannelAPI) CreatePrivateChannelWithMembers(ctx context.Context, teamID, displayName string, memberIDs, ownerIDs []string) (msmodels.Channelable, *sender.RequestError) {
	return nil, nil
}

func (f *fakeChannelAPI) DeleteChannel(ctx context.Context, teamID, channelID string) *sender.RequestError {
	return nil
}

func (f *fakeChannelAPI) GetChannel(ctx context.Context, teamID, channelID string) (msmodels.Channelable, *sender.RequestError) {
	return nil, nil
}

func (f *fakeChannelAPI) GetMessage(ctx context.Context, teamID, channelID, messageID string) (msmodels.ChatMessageable, *sender.RequestError) {
	return nil, nil
}

func (f *fakeChannelAPI) GetReply(ctx context.Context, teamID, channelID, messageID, replyID string) (msmodels.ChatMessageable, *sender.RequestError) {
	return nil, nil
}

func (f *fakeChannelAPI) ListMessages(ctx context.Context, teamID, channelID string, top *int32) (msmodels.ChatMessageCollectionResponseable, *sender.RequestError) {
	return nil, nil
}

func (f *fakeChannelAPI) ListReplies(ctx context.Context, teamID, channelID, messageID string, top *int32) (msmodels.ChatMessageCollectionResponseable, *sender.RequestError) {
	return nil, nil
}

func (f *fakeChannelAPI) SendMessage(ctx context.Context, teamID, channelID string, message msmodels.ChatMessageable) (msmodels.ChatMessageable, *sender.RequestError) {
	return nil, nil
}

func newGraphTeam(id, name string) msmodels.Teamable {
	t := msmodels.NewTeam()
	t.SetId(&id)
	t.SetDisplayName(&name)
	return t
}

func newGraphChannel(id, name string) msmodels.Channelable {
	ch := msmodels.NewChannel()
	ch.SetId(&id)
	ch.SetDisplayName(&name)
	return ch
}

func TestMapper_MapTeamNameToTeamID_Found(t *testing.T) {
	ctx := context.Background()

	col := msmodels.NewTeamCollectionResponse()
	a := newGraphTeam("1", "Alpha")
	b := newGraphTeam("2", "Beta")
	col.SetValue([]msmodels.Teamable{a, b})

	teamsFake := &fakeTeamsAPI{listResp: col}
	channelsFake := &fakeChannelAPI{}

	m := &mapper{
		teamsAPI:    teamsFake,
		channelsAPI: channelsFake,
	}

	id, err := m.MapTeamNameToTeamID(ctx, "Beta")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id != "2" {
		t.Fatalf("expected id '2', got %q", id)
	}
}

func TestMapper_MapTeamNameToTeamID_NotFound(t *testing.T) {
	ctx := context.Background()

	col := msmodels.NewTeamCollectionResponse()
	col.SetValue([]msmodels.Teamable{
		newGraphTeam("1", "Alpha"),
	})

	teamsFake := &fakeTeamsAPI{listResp: col}
	channelsFake := &fakeChannelAPI{}

	m := &mapper{
		teamsAPI:    teamsFake,
		channelsAPI: channelsFake,
	}

	_, err := m.MapTeamNameToTeamID(ctx, "Beta")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "team with name") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestMapper_MapChannelNameToChannelID_Found(t *testing.T) {
	ctx := context.Background()

	col := msmodels.NewChannelCollectionResponse()
	c1 := newGraphChannel("10", "General")
	c2 := newGraphChannel("11", "Random")
	col.SetValue([]msmodels.Channelable{c1, c2})

	chFake := &fakeChannelAPI{listResp: col}
	teamsFake := &fakeTeamsAPI{}

	m := &mapper{
		teamsAPI:    teamsFake,
		channelsAPI: chFake,
	}

	id, err := m.MapChannelNameToChannelID(ctx, "team-123", "Random")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id != "11" {
		t.Fatalf("expected id '11', got %q", id)
	}
}

func TestMapper_MapChannelNameToChannelID_NotFound(t *testing.T) {
	ctx := context.Background()

	col := msmodels.NewChannelCollectionResponse()
	col.SetValue([]msmodels.Channelable{
		newGraphChannel("10", "General"),
	})

	chFake := &fakeChannelAPI{listResp: col}
	teamsFake := &fakeTeamsAPI{}

	m := &mapper{
		teamsAPI:    teamsFake,
		channelsAPI: chFake,
	}

	_, err := m.MapChannelNameToChannelID(ctx, "team-123", "Random")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "channel with name") {
		t.Errorf("unexpected error message: %v", err)
	}
}
