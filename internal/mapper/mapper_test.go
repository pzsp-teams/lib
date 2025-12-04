package utils

import (
	"context"
	"strings"
	"testing"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	sender "github.com/pzsp-teams/lib/internal/sender"
	channelspkg "github.com/pzsp-teams/lib/pkg/teams/channels"
	teamspkg "github.com/pzsp-teams/lib/pkg/teams/teams"
)

type fakeTeamsAPI struct {
	listResp msmodels.TeamCollectionResponseable
	listErr  *sender.RequestError
}

func (f *fakeTeamsAPI) CreateFromTemplate(ctx context.Context, displayName, description string, owners []string) (string, *sender.RequestError) {
	return "", nil
}
func (f *fakeTeamsAPI) CreateViaGroup(ctx context.Context, displayName, mailNickname, visibility string) (string, *sender.RequestError) {
	return "", nil
}
func (f *fakeTeamsAPI) Get(ctx context.Context, teamID string) (msmodels.Teamable, *sender.RequestError) {
	return nil, nil
}
func (f *fakeTeamsAPI) ListMyJoined(ctx context.Context) (msmodels.TeamCollectionResponseable, *sender.RequestError) {
	return f.listResp, f.listErr
}
func (f *fakeTeamsAPI) Update(ctx context.Context, teamID string, patch *msmodels.Team) (msmodels.Teamable, *sender.RequestError) {
	return nil, nil
}
func (f *fakeTeamsAPI) Archive(ctx context.Context, teamID string, spoReadOnlyForMembers *bool) *sender.RequestError {
	return nil
}
func (f *fakeTeamsAPI) Unarchive(ctx context.Context, teamID string) *sender.RequestError {
	return nil
}
func (f *fakeTeamsAPI) Delete(ctx context.Context, teamID string) *sender.RequestError {
	return nil
}
func (f *fakeTeamsAPI) RestoreDeleted(ctx context.Context, deletedGroupID string) (msmodels.DirectoryObjectable, *sender.RequestError) {
	return nil, nil
}

type fakeChannelAPI struct {
	listResp msmodels.ChannelCollectionResponseable
	listErr  *sender.RequestError
}

func (f *fakeChannelAPI) ListChannels(ctx context.Context, teamID string) (msmodels.ChannelCollectionResponseable, *sender.RequestError) {
	return f.listResp, f.listErr
}
func (f *fakeChannelAPI) GetChannel(ctx context.Context, teamID, channelID string) (msmodels.Channelable, *sender.RequestError) {
	return nil, nil
}
func (f *fakeChannelAPI) CreateChannel(ctx context.Context, teamID string, channel msmodels.Channelable) (msmodels.Channelable, *sender.RequestError) {
	return nil, nil
}
func (f *fakeChannelAPI) DeleteChannel(ctx context.Context, teamID, channelID string) *sender.RequestError {
	return nil
}
func (f *fakeChannelAPI) SendMessage(ctx context.Context, teamID, channelID string, message msmodels.ChatMessageable) (msmodels.ChatMessageable, *sender.RequestError) {
	return nil, nil
}
func (f *fakeChannelAPI) ListMessages(ctx context.Context, teamID, channelID string, top *int32) (msmodels.ChatMessageCollectionResponseable, *sender.RequestError) {
	return nil, nil
}
func (f *fakeChannelAPI) GetMessage(ctx context.Context, teamID, channelID, messageID string) (msmodels.ChatMessageable, *sender.RequestError) {
	return nil, nil
}
func (f *fakeChannelAPI) ListReplies(ctx context.Context, teamID, channelID, messageID string, top *int32) (msmodels.ChatMessageCollectionResponseable, *sender.RequestError) {
	return nil, nil
}
func (f *fakeChannelAPI) GetReply(ctx context.Context, teamID, channelID, messageID, replyID string) (msmodels.ChatMessageable, *sender.RequestError) {
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

	teamsAPI := &fakeTeamsAPI{listResp: col}
	teamSvc := teamspkg.NewService(teamsAPI)

	channelsAPI := &fakeChannelAPI{}
	channelSvc := channelspkg.NewService(channelsAPI)

	m := NewMapper(teamSvc, channelSvc)

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

	teamsAPI := &fakeTeamsAPI{listResp: col}
	teamSvc := teamspkg.NewService(teamsAPI)
	channelSvc := channelspkg.NewService(&fakeChannelAPI{})

	m := NewMapper(teamSvc, channelSvc)

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

	chAPI := &fakeChannelAPI{listResp: col}
	channelSvc := channelspkg.NewService(chAPI)

	teamsAPI := &fakeTeamsAPI{}
	teamSvc := teamspkg.NewService(teamsAPI)

	m := NewMapper(teamSvc, channelSvc)

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

	chAPI := &fakeChannelAPI{listResp: col}
	channelSvc := channelspkg.NewService(chAPI)

	teamsAPI := &fakeTeamsAPI{}
	teamSvc := teamspkg.NewService(teamsAPI)

	m := NewMapper(teamSvc, channelSvc)

	_, err := m.MapChannelNameToChannelID(ctx, "team-123", "Random")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "channel with name") {
		t.Errorf("unexpected error message: %v", err)
	}
}
