package resolver

import (
	"errors"
	"testing"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
)

func TestResolveTeamIDByName_NoTeamsAvailable(t *testing.T) {
	col := msmodels.NewTeamCollectionResponse()
	col.SetValue(nil)

	_, err := resolveTeamIDByName(col, "X")
	if err == nil {
		t.Fatalf("expected error for no teams, got nil")
	}

	var rnErr *resourcesNotAvailableError
	if !errors.As(err, &rnErr) {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestResolveTeamIDByName_NoMatch(t *testing.T) {
	t1 := newGraphTeam("1", "Alpha")
	col := newTeamCollection(t1)

	_, err := resolveTeamIDByName(col, "Beta")
	if err == nil {
		t.Fatalf("expected error for missing team, got nil")
	}

	var rnErr *resourceNotFoundError
	if !errors.As(err, &rnErr) {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestResolveTeamIDByName_SingleMatch(t *testing.T) {
	t1 := newGraphTeam("1", "Alpha")
	t2 := newGraphTeam("2", "Beta")
	col := newTeamCollection(t1, t2)

	id, err := resolveTeamIDByName(col, "Beta")
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

	_, err := resolveTeamIDByName(col, "Alpha")
	if err == nil {
		t.Fatalf("expected error for multiple matches, got nil")
	}

	var raErr *resourceAmbiguousError
	if !errors.As(err, &raErr) {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestResolveChannelIDByName_NoChannelsAvailable(t *testing.T) {
	col := msmodels.NewChannelCollectionResponse()
	col.SetValue(nil)

	_, err := resolveChannelIDByName(col, "team-1", "X")
	if err == nil {
		t.Fatalf("expected error for no channels, got nil")
	}

	var rnErr *resourcesNotAvailableError
	if !errors.As(err, &rnErr) {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestResolveChannelIDByName_NoMatch(t *testing.T) {
	ch1 := newGraphChannel("c1", "General")
	col := newChannelCollection(ch1)

	_, err := resolveChannelIDByName(col, "team-1", "Random")
	if err == nil {
		t.Fatalf("expected error for missing channel, got nil")
	}

	var rnErr *resourceNotFoundError
	if !errors.As(err, &rnErr) {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestResolveChannelIDByName_SingleMatch(t *testing.T) {
	ch1 := newGraphChannel("c1", "General")
	ch2 := newGraphChannel("c2", "Random")
	col := newChannelCollection(ch1, ch2)

	id, err := resolveChannelIDByName(col, "team-1", "Random")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "c2" {
		t.Fatalf("expected id=c2, got %q", id)
	}
}

func TestResolveChannelIDByName_MultipleMatches(t *testing.T) {
	ch1 := newGraphChannel("c1", "General")
	ch2 := newGraphChannel("c2", "General")
	col := newChannelCollection(ch1, ch2)

	_, err := resolveChannelIDByName(col, "team-1", "General")
	if err == nil {
		t.Fatalf("expected error for multiple matches, got nil")
	}

	var raErr *resourceAmbiguousError
	if !errors.As(err, &raErr) {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestResolveOneOnOneChatIDByUserRef_NoChatsAvailable(t *testing.T) {
	col := msmodels.NewChatCollectionResponse()
	col.SetValue(nil)

	_, err := resolveOneOnOneChatIDByUserRef(col, "u")
	if err == nil {
		t.Fatalf("expected error for no one-on-one chats, got nil")
	}

	var rnErr *resourcesNotAvailableError
	if !errors.As(err, &rnErr) {
		t.Errorf("unexpected error: %v", err)
	}
}
func TestResolveOneOnOneChatIDByUserRef_NoMatch(t *testing.T) {
	m := newAadUserMember("m-1", "usr-1", "jane@example.com")
	chat := newOneOnOneChat("chat-1", m)
	col := newChatCollection(chat)

	_, err := resolveOneOnOneChatIDByUserRef(col, "other-user")
	if err == nil {
		t.Fatalf("expected error for missing user, got nil")
	}

	var rnErr *resourceNotFoundError
	if !errors.As(err, &rnErr) {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestResolveOneOnOneChatIDByUserRef_IDMatch(t *testing.T) {
	m := newAadUserMember("m-1", "usr-1", "jane@example.com")
	chat := newOneOnOneChat("chat-1", m)
	col := newChatCollection(chat)

	id, err := resolveOneOnOneChatIDByUserRef(col, "usr-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id != "chat-1" {
		t.Fatalf("expected chat id chat-1, got %q", id)
	}
}

func TestResolveOneOnOneChatIDByUserRef_EmailMatch(t *testing.T) {
	m := newAadUserMember("m-1", "usr-1", "jane@example.com")
	chat := newOneOnOneChat("chat-1", m)
	col := newChatCollection(chat)

	id, err := resolveOneOnOneChatIDByUserRef(col, "jane@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id != "chat-1" {
		t.Fatalf("expected chat id chat-1, got %q", id)
	}
}

func TestResolveGroupChatIDByTopic_NoChatsAvailable(t *testing.T) {
	col := msmodels.NewChatCollectionResponse()
	col.SetValue(nil)

	_, err := resolveGroupChatIDByTopic(col, "Topic")
	if err == nil {
		t.Fatalf("expected error for no group chats, got nil")
	}

	var rnErr *resourcesNotAvailableError
	if !errors.As(err, &rnErr) {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestResolveGroupChatIDByTopic_NoMatch(t *testing.T) {
	c := newGroupChat("c1", "X")
	col := newChatCollection(c)

	_, err := resolveGroupChatIDByTopic(col, "Y")
	if err == nil {
		t.Fatalf("expected error for missing topic, got nil")
	}
	var rnErr *resourceNotFoundError
	if !errors.As(err, &rnErr) {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestResolveGroupChatIDByTopic_SingleMatch(t *testing.T) {
	c1 := newGroupChat("c1", "Project")
	col := newChatCollection(c1)

	id, err := resolveGroupChatIDByTopic(col, "Project")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "c1" {
		t.Fatalf("expected c1, got %q", id)
	}
}

func TestResolveGroupChatIDByTopic_MultipleMatches(t *testing.T) {
	c1 := newGroupChat("c1", "Project")
	c2 := newGroupChat("c2", "Project")
	col := newChatCollection(c1, c2)

	_, err := resolveGroupChatIDByTopic(col, "Project")
	if err == nil {
		t.Fatalf("expected error for multiple chats, got nil")
	}

	var raErr *resourceAmbiguousError
	if !errors.As(err, &raErr) {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestResolveMemberID_NoMembersAvailable(t *testing.T) {
	colEmpty := msmodels.NewConversationMemberCollectionResponse()
	colEmpty.SetValue(nil)

	_, err := resolveMemberID(colEmpty, "any")
	if err == nil {
		t.Fatalf("expected error for no members, got nil")
	}

	var rnErr *resourcesNotAvailableError
	if !errors.As(err, &rnErr) {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestResolveMemberID_NoMatch(t *testing.T) {
	m := newAadUserMember("m-1", "usr-1", "jane@example.com")
	col := newMemberCollection(m)

	_, err := resolveMemberID(col, "missing")
	if err == nil {
		t.Fatalf("expected error for missing member, got nil")
	}

	var rnErr *resourceNotFoundError
	if !errors.As(err, &rnErr) {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestResolveMemberID_UserIDMatch(t *testing.T) {
	m := newAadUserMember("m-1", "usr-1", "jane@example.com")
	col2 := newMemberCollection(m)

	id, err := resolveMemberID(col2, "usr-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id != "m-1" {
		t.Fatalf("expected m-1, got %q", id)
	}
}

func TestResolveMemberID_UserEmailMatch(t *testing.T) {
	m := newAadUserMember("m-1", "usr-1", "jane@example.com")
	col2 := newMemberCollection(m)

	id2, err := resolveMemberID(col2, "jane@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id2 != "m-1" {
		t.Fatalf("expected m-1 by display name, got %q", id2)
	}
}

// ----------
// TEST UTILS
// ----------

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

func newGraphChannel(id, name string) msmodels.Channelable {
	ch := msmodels.NewChannel()
	ch.SetId(&id)
	ch.SetDisplayName(&name)
	return ch
}

func newChannelCollection(channels ...msmodels.Channelable) msmodels.ChannelCollectionResponseable {
	col := msmodels.NewChannelCollectionResponse()
	col.SetValue(channels)
	return col
}

func newAadUserMember(id, userID, email string) msmodels.ConversationMemberable {
	m := msmodels.NewAadUserConversationMember()
	m.SetId(&id)
	m.SetUserId(&userID)
	_ = m.GetBackingStore().Set("email", &email)
	return m
}

func newMemberCollection(members ...msmodels.ConversationMemberable) msmodels.ConversationMemberCollectionResponseable {
	col := msmodels.NewConversationMemberCollectionResponse()
	col.SetValue(members)
	return col
}

func newOneOnOneChat(chatID string, member msmodels.ConversationMemberable) msmodels.Chatable {
	c := msmodels.NewChat()
	c.SetId(&chatID)
	chatType := msmodels.ONEONONE_CHATTYPE
	c.SetChatType(&chatType)
	c.SetMembers([]msmodels.ConversationMemberable{member})
	return c
}

func newGroupChat(chatID, topic string) msmodels.Chatable {
	c := msmodels.NewChat()
	c.SetId(&chatID)
	chatType := msmodels.GROUP_CHATTYPE
	c.SetChatType(&chatType)
	c.SetTopic(&topic)
	return c
}

func newChatCollection(chats ...msmodels.Chatable) msmodels.ChatCollectionResponseable {
	col := msmodels.NewChatCollectionResponse()
	col.SetValue(chats)
	return col
}
