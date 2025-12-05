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
	listResp    msmodels.ChannelCollectionResponseable
	listErr     *sender.RequestError
	membersResp msmodels.ConversationMemberCollectionResponseable
	membersErr  *sender.RequestError
}

func (f *fakeChannelAPI) ListChannels(ctx context.Context, teamID string) (msmodels.ChannelCollectionResponseable, *sender.RequestError) {
	return f.listResp, f.listErr
}

func (f *fakeChannelAPI) CreateStandardChannel(ctx context.Context, teamID string, channel msmodels.Channelable) (msmodels.Channelable, *sender.RequestError) {
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

func (f *fakeChannelAPI) AddMember(ctx context.Context, teamID, channelID, memberID, ownerID string) (msmodels.ConversationMemberable, *sender.RequestError) {
	return nil, nil
}

func (f *fakeChannelAPI) ListMembers(ctx context.Context, teamID, channelID string) (msmodels.ConversationMemberCollectionResponseable, *sender.RequestError) {
	return f.membersResp, f.membersErr
}

func (f *fakeChannelAPI) RemoveMember(ctx context.Context, teamID, channelID, memberID string) *sender.RequestError {
	return nil
}

func (f *fakeChannelAPI) UpdateMemberRole(ctx context.Context, teamID, channelID, memberID, role string) (msmodels.ConversationMemberable, *sender.RequestError) {
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

	id, err := m.MapTeamRefToTeamID(ctx, "Beta")
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

	_, err := m.MapTeamRefToTeamID(ctx, "Beta")
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

	id, err := m.MapChannelRefToChannelID(ctx, "team-123", "Random")
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

	_, err := m.MapChannelRefToChannelID(ctx, "team-123", "Random")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "channel with name") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func newAadMember(memberID, userID, displayName string, add map[string]any) msmodels.ConversationMemberable {
	m := msmodels.NewAadUserConversationMember()
	if memberID != "" {
		m.SetId(&memberID)
	}
	if userID != "" {
		m.SetUserId(&userID)
	}
	if displayName != "" {
		m.SetDisplayName(&displayName)
	}
	if add != nil {
		m.SetAdditionalData(add)
	}
	return m
}

func TestMapper_MapUserRefToMemberID_MatchByUserID(t *testing.T) {
	ctx := context.Background()

	col := msmodels.NewConversationMemberCollectionResponse()
	m1 := newAadMember("m-1", "user-1", "Alice", nil)
	col.SetValue([]msmodels.ConversationMemberable{m1})

	chFake := &fakeChannelAPI{membersResp: col}
	teamsFake := &fakeTeamsAPI{}

	m := &mapper{
		teamsAPI:    teamsFake,
		channelsAPI: chFake,
	}

	id, err := m.MapUserRefToMemberID(ctx, "team-1", "chan-1", "user-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id != "m-1" {
		t.Fatalf("expected member id 'm-1', got %q", id)
	}
}

func TestMapper_MapUserRefToMemberID_MatchByDisplayName(t *testing.T) {
	ctx := context.Background()

	col := msmodels.NewConversationMemberCollectionResponse()
	m1 := newAadMember("m-2", "user-2", "Bob", nil)
	col.SetValue([]msmodels.ConversationMemberable{m1})

	chFake := &fakeChannelAPI{membersResp: col}
	teamsFake := &fakeTeamsAPI{}

	m := &mapper{
		teamsAPI:    teamsFake,
		channelsAPI: chFake,
	}

	id, err := m.MapUserRefToMemberID(ctx, "team-1", "chan-1", "Bob")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id != "m-2" {
		t.Fatalf("expected member id 'm-2', got %q", id)
	}
}

func TestMapper_MapUserRefToMemberID_MatchByUPNOrMail(t *testing.T) {
	ctx := context.Background()

	col := msmodels.NewConversationMemberCollectionResponse()
	m1 := newAadMember("m-3", "user-3", "Charlie", map[string]any{
		"userPrincipalName": "charlie@contoso.com",
	})
	m2 := newAadMember("m-4", "user-4", "Dave", map[string]any{
		"mail": "dave@contoso.com",
	})
	col.SetValue([]msmodels.ConversationMemberable{m1, m2})

	chFake := &fakeChannelAPI{membersResp: col}
	teamsFake := &fakeTeamsAPI{}
	m := &mapper{
		teamsAPI:    teamsFake,
		channelsAPI: chFake,
	}

	id, err := m.MapUserRefToMemberID(ctx, "team-1", "chan-1", "charlie@contoso.com")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id != "m-3" {
		t.Fatalf("expected member id 'm-3', got %q", id)
	}

	id2, err := m.MapUserRefToMemberID(ctx, "team-1", "chan-1", "dave@contoso.com")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id2 != "m-4" {
		t.Fatalf("expected member id 'm-4', got %q", id2)
	}
}

func TestMapper_MapUserRefToMemberID_NoMembers(t *testing.T) {
	ctx := context.Background()

	chFake := &fakeChannelAPI{membersResp: nil}
	teamsFake := &fakeTeamsAPI{}
	m := &mapper{
		teamsAPI:    teamsFake,
		channelsAPI: chFake,
	}

	_, err := m.MapUserRefToMemberID(ctx, "team-1", "chan-1", "someone@contoso.com")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "no members found in channel") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestMapper_MapUserRefToMemberID_NotFound(t *testing.T) {
	ctx := context.Background()

	col := msmodels.NewConversationMemberCollectionResponse()
	m1 := newAadMember("m-1", "user-1", "Alice", nil)
	col.SetValue([]msmodels.ConversationMemberable{m1})

	chFake := &fakeChannelAPI{membersResp: col}
	teamsFake := &fakeTeamsAPI{}
	m := &mapper{
		teamsAPI:    teamsFake,
		channelsAPI: chFake,
	}

	_, err := m.MapUserRefToMemberID(ctx, "team-1", "chan-1", "nonexistent@contoso.com")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "member \"nonexistent@contoso.com\" not found") {
		t.Errorf("unexpected error message: %v", err)
	}
}
