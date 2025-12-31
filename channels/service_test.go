package channels

import (
	"context"
	"errors"
	"testing"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/internal/adapter"
	sender "github.com/pzsp-teams/lib/internal/sender"
	"github.com/pzsp-teams/lib/models"
)

type fakeTeamRes struct {
	resolveTeamErr error
	lastTeamName   string
}

func (m *fakeTeamRes) ResolveTeamRefToID(ctx context.Context, teamName string) (string, error) {
	m.lastTeamName = teamName
	if m.resolveTeamErr != nil {
		return "", m.resolveTeamErr
	}
	return teamName, nil
}

type fakeChannelRes struct {
	resChanErr           error
	lastChannelName      string
	lastTeamIDForChannel string
	resUserRefErr        error
	lastUserRef          string
	lastTeamIDForUser    string
	lastChannelIDForUser string
}

func (m *fakeChannelRes) ResolveChannelRefToID(ctx context.Context, teamID, channelName string) (string, error) {
	m.lastTeamIDForChannel = teamID
	m.lastChannelName = channelName
	if m.resChanErr != nil {
		return "", m.resChanErr
	}
	return channelName, nil
}

func (m *fakeChannelRes) ResolveChannelMemberRefToID(ctx context.Context, teamID, channelID, userRef string) (string, error) {
	m.lastTeamIDForUser = teamID
	m.lastChannelIDForUser = channelID
	m.lastUserRef = userRef
	if m.resUserRefErr != nil {
		return "", m.resUserRefErr
	}
	return userRef, nil
}

type fakeChanAPI struct {
	listResp        msmodels.ChannelCollectionResponseable
	listErr         *sender.RequestError
	getResp         msmodels.Channelable
	getErr          *sender.RequestError
	createResp      msmodels.Channelable
	createErr       *sender.RequestError
	deleteErr       *sender.RequestError
	lastCreate      msmodels.Channelable
	lastTeamID      string
	lastChanID      string
	sendMsgResp     msmodels.ChatMessageable
	sendMsgErr      *sender.RequestError
	listMsgsResp    msmodels.ChatMessageCollectionResponseable
	listMsgsErr     *sender.RequestError
	getMsgResp      msmodels.ChatMessageable
	getMsgErr       *sender.RequestError
	listRepliesResp msmodels.ChatMessageCollectionResponseable
	listRepliesErr  *sender.RequestError
	getReplyResp    msmodels.ChatMessageable
	getReplyErr     *sender.RequestError
	lastContent     string
	lastContentType string
	lastMessageID   string
	lastReplyID     string

	membersResp msmodels.ConversationMemberCollectionResponseable
	membersErr  *sender.RequestError

	addMemberResp  msmodels.ConversationMemberable
	addMemberErr   *sender.RequestError
	lastAddUserRef string
	lastAddRole    string

	updateMemberResp   msmodels.ConversationMemberable
	updateMemberErr    *sender.RequestError
	lastUpdateMemberID string
	lastUpdateRole     string

	removeMemberErr    *sender.RequestError
	lastRemoveMemberID string
}

func (f *fakeChanAPI) ListChannels(ctx context.Context, teamID string) (msmodels.ChannelCollectionResponseable, *sender.RequestError) {
	f.lastTeamID = teamID
	return f.listResp, f.listErr
}

func (f *fakeChanAPI) GetChannel(ctx context.Context, teamID, channelID string) (msmodels.Channelable, *sender.RequestError) {
	f.lastTeamID = teamID
	f.lastChanID = channelID
	return f.getResp, f.getErr
}

func (f *fakeChanAPI) CreateStandardChannel(ctx context.Context, teamID string, channel msmodels.Channelable) (msmodels.Channelable, *sender.RequestError) {
	f.lastTeamID = teamID
	f.lastCreate = channel
	return f.createResp, f.createErr
}

func (f *fakeChanAPI) CreatePrivateChannelWithMembers(ctx context.Context, teamID, displayName string, memberIDs, ownerIDs []string) (msmodels.Channelable, *sender.RequestError) {
	f.lastTeamID = teamID
	return f.createResp, f.createErr
}

func (f *fakeChanAPI) DeleteChannel(ctx context.Context, teamID, channelID string) *sender.RequestError {
	f.lastTeamID = teamID
	f.lastChanID = channelID
	return f.deleteErr
}

func (f *fakeChanAPI) SendMessage(ctx context.Context, teamID, channelID, content, contentType string, mentions []msmodels.ChatMessageMentionable) (msmodels.ChatMessageable, *sender.RequestError) {
	f.lastTeamID = teamID
	f.lastChanID = channelID
	f.lastContent = content
	f.lastContentType = contentType
	return f.sendMsgResp, f.sendMsgErr
}

func (f *fakeChanAPI) SendReply(ctx context.Context, teamID, channelID, messageID, content, contentType string, mentions []msmodels.ChatMessageMentionable) (msmodels.ChatMessageable, *sender.RequestError) {
	f.lastTeamID = teamID
	f.lastChanID = channelID
	f.lastMessageID = messageID
	f.lastContent = content
	f.lastContentType = contentType
	return f.sendMsgResp, f.sendMsgErr
}

func (f *fakeChanAPI) ListMessages(ctx context.Context, teamID, channelID string, top *int32) (msmodels.ChatMessageCollectionResponseable, *sender.RequestError) {
	f.lastTeamID = teamID
	f.lastChanID = channelID
	return f.listMsgsResp, f.listMsgsErr
}

func (f *fakeChanAPI) GetMessage(ctx context.Context, teamID, channelID, messageID string) (msmodels.ChatMessageable, *sender.RequestError) {
	f.lastTeamID = teamID
	f.lastChanID = channelID
	f.lastMessageID = messageID
	return f.getMsgResp, f.getMsgErr
}

func (f *fakeChanAPI) ListReplies(ctx context.Context, teamID, channelID, messageID string, top *int32) (msmodels.ChatMessageCollectionResponseable, *sender.RequestError) {
	f.lastTeamID = teamID
	f.lastChanID = channelID
	f.lastMessageID = messageID
	return f.listRepliesResp, f.listRepliesErr
}

func (f *fakeChanAPI) GetReply(ctx context.Context, teamID, channelID, messageID, replyID string) (msmodels.ChatMessageable, *sender.RequestError) {
	f.lastTeamID = teamID
	f.lastChanID = channelID
	f.lastMessageID = messageID
	f.lastReplyID = replyID
	return f.getReplyResp, f.getReplyErr
}

func (f *fakeChanAPI) ListMembers(ctx context.Context, teamID, channelID string) (msmodels.ConversationMemberCollectionResponseable, *sender.RequestError) {
	f.lastTeamID = teamID
	f.lastChanID = channelID
	return f.membersResp, f.membersErr
}

func (f *fakeChanAPI) AddMember(ctx context.Context, teamID, channelID, userID, role string) (msmodels.ConversationMemberable, *sender.RequestError) {
	f.lastTeamID = teamID
	f.lastChanID = channelID
	f.lastAddUserRef = userID
	f.lastAddRole = role
	return f.addMemberResp, f.addMemberErr
}

func (f *fakeChanAPI) RemoveMember(ctx context.Context, teamID, channelID, memberID string) *sender.RequestError {
	f.lastTeamID = teamID
	f.lastChanID = channelID
	f.lastRemoveMemberID = memberID
	return f.removeMemberErr
}

func (f *fakeChanAPI) UpdateMemberRole(ctx context.Context, teamID, channelID, memberID, role string) (msmodels.ConversationMemberable, *sender.RequestError) {
	f.lastTeamID = teamID
	f.lastChanID = channelID
	f.lastUpdateMemberID = memberID
	f.lastUpdateRole = role
	return f.updateMemberResp, f.updateMemberErr
}

func newGraphChan(id, name string) msmodels.Channelable {
	channel := msmodels.NewChannel()
	channel.SetId(&id)
	channel.SetDisplayName(&name)
	return channel
}

func newChatMessage(id, content string) msmodels.ChatMessageable {
	msg := msmodels.NewChatMessage()
	msg.SetId(&id)
	body := msmodels.NewItemBody()
	body.SetContent(&content)
	msg.SetBody(body)
	return msg
}

func newAadUserMember(id, userID, displayName string, roles []string) msmodels.ConversationMemberable {
	m := msmodels.NewAadUserConversationMember()
	if id != "" {
		m.SetId(&id)
	}
	if userID != "" {
		m.SetUserId(&userID)
	}
	if displayName != "" {
		m.SetDisplayName(&displayName)
	}
	if roles != nil {
		m.SetRoles(roles)
	}
	return m
}

type fakeUsersAPI struct {
	byKey   map[string]msmodels.Userable
	lastKey string
	calls   int
	err     *sender.RequestError
}

func (f *fakeUsersAPI) GetUserByEmailOrUPN(ctx context.Context, emailOrUPN string) (msmodels.Userable, *sender.RequestError) {
	f.lastKey = emailOrUPN
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	if f.byKey == nil {
		return nil, &sender.RequestError{Message: "no users configured"}
	}
	u, ok := f.byKey[emailOrUPN]
	if !ok {
		return nil, &sender.RequestError{Message: "user not found"}
	}
	return u, nil
}

func newGraphUser(id, displayName string) msmodels.Userable {
	u := msmodels.NewUser()
	if id != "" {
		u.SetId(&id)
	}
	if displayName != "" {
		u.SetDisplayName(&displayName)
	}
	return u
}

func TestService_ListChannels_MapsFieldsAndGeneralFlag(t *testing.T) {
	ctx := context.Background()
	col := msmodels.NewChannelCollectionResponse()

	ch1 := newGraphChan("1", "General")
	ch2 := newGraphChan("2", "Random")

	col.SetValue([]msmodels.Channelable{ch1, ch2})

	api := &fakeChanAPI{listResp: col}
	m := &fakeTeamRes{}
	svc := NewService(api, m, &fakeChannelRes{}, nil)

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

	if m.lastTeamName != "team-1" {
		t.Errorf("expected mapper to be called with team-1, got %q", m.lastTeamName)
	}

	if api.lastTeamID != "team-1" {
		t.Errorf("expected api to be called with team-1, got %q", api.lastTeamID)
	}
}

func TestService_ListChannels_MapsErrors(t *testing.T) {
	ctx := context.Background()
	api := &fakeChanAPI{
		listErr: &sender.RequestError{
			Code:    404, // ResourceNotFound
			Message: "team not found",
		},
	}
	svc := NewService(api, &fakeTeamRes{}, &fakeChannelRes{}, nil)

	_, err := svc.ListChannels(ctx, "non-existing-team")
	var notFound sender.ErrResourceNotFound
	if !errors.As(err, &notFound) {
		t.Fatalf("expected ErrResourceNotFound, got %v", err)
	}
}

func TestService_Get_MapsSingleChannel(t *testing.T) {
	ctx := context.Background()
	ch := newGraphChan("42", "General")
	api := &fakeChanAPI{getResp: ch}
	svc := NewService(api, &fakeTeamRes{}, &fakeChannelRes{}, nil)

	got, err := svc.Get(ctx, "team-1", "42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.ID != "42" || got.Name != "General" || !got.IsGeneral {
		t.Errorf("unexpected channel: %+v", got)
	}
	if api.lastTeamID != "team-1" || api.lastChanID != "42" {
		t.Errorf("expected api called with team-1/42, got team=%q, chan=%q", api.lastTeamID, api.lastChanID)
	}
}

func TestService_Create_SetsNameAndMapsResult(t *testing.T) {
	ctx := context.Background()
	created := newGraphChan("123", "my-channel")

	api := &fakeChanAPI{
		createResp: created,
	}
	svc := NewService(api, &fakeTeamRes{}, &fakeChannelRes{}, nil)

	got, err := svc.CreateStandardChannel(ctx, "team-1", "my-channel")
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
	if api.lastTeamID != "team-1" {
		t.Errorf("expected team ID 'team-1', got %q", api.lastTeamID)
	}
}

func TestService_Delete_MapsError(t *testing.T) {
	ctx := context.Background()
	api := &fakeChanAPI{
		deleteErr: &sender.RequestError{
			Code:    403,
			Message: "nope",
		},
	}
	svc := NewService(api, &fakeTeamRes{}, &fakeChannelRes{}, nil)

	err := svc.Delete(ctx, "team-1", "chan-1")
	var forbidden sender.ErrAccessForbidden
	if !errors.As(err, &forbidden) {
		t.Fatalf("expected ErrAccessForbidden, got %v", err)
	}
}

func TestService_SendMessage_CreatesMessageAndMapsResult(t *testing.T) {
	ctx := context.Background()
	msgID := "msg-123"
	msgContent := "Hello, Teams!"

	respMsg := newChatMessage(msgID, msgContent)
	api := &fakeChanAPI{sendMsgResp: respMsg}
	svc := NewService(api, &fakeTeamRes{}, &fakeChannelRes{}, nil)

	body := models.MessageBody{Content: msgContent, ContentType: models.MessageContentTypeText}
	got, err := svc.SendMessage(ctx, "team-1", "chan-1", body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.ID != msgID {
		t.Errorf("expected ID %q, got %q", msgID, got.ID)
	}
	if got.Content != msgContent {
		t.Errorf("expected content %q, got %q", msgContent, got.Content)
	}

	if api.lastTeamID != "team-1" {
		t.Errorf("expected team ID 'team-1', got %q", api.lastTeamID)
	}
	if api.lastChanID != "chan-1" {
		t.Errorf("expected channel ID 'chan-1', got %q", api.lastChanID)
	}
	if api.lastContent != msgContent {
		t.Errorf("expected content %q, got %q", msgContent, api.lastContent)
	}
	if api.lastContentType != "text" {
		t.Errorf("expected content type 'text', got %q", api.lastContentType)
	}
}

func TestService_SendMessage_MapsError(t *testing.T) {
	ctx := context.Background()
	api := &fakeChanAPI{
		sendMsgErr: &sender.RequestError{
			Code:    403,
			Message: "not allowed",
		},
	}
	svc := NewService(api, &fakeTeamRes{}, &fakeChannelRes{}, nil)

	body := models.MessageBody{Content: "test"}
	_, err := svc.SendMessage(ctx, "team-1", "chan-1", body)
	var forbidden sender.ErrAccessForbidden
	if !errors.As(err, &forbidden) {
		t.Fatalf("expected ErrAccessForbidden, got %v", err)
	}
}

func TestService_ListMessages_MapsMultipleMessages(t *testing.T) {
	ctx := context.Background()
	col := msmodels.NewChatMessageCollectionResponse()

	msg1 := newChatMessage("msg-1", "First message")
	msg2 := newChatMessage("msg-2", "Second message")
	col.SetValue([]msmodels.ChatMessageable{msg1, msg2})

	api := &fakeChanAPI{listMsgsResp: col}
	svc := NewService(api, &fakeTeamRes{}, &fakeChannelRes{}, nil)

	got, err := svc.ListMessages(ctx, "team-1", "chan-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(got))
	}

	if got[0].ID != "msg-1" || got[0].Content != "First message" {
		t.Errorf("unexpected first message: %+v", got[0])
	}
	if got[1].ID != "msg-2" || got[1].Content != "Second message" {
		t.Errorf("unexpected second message: %+v", got[1])
	}
}

func TestService_ListMessages_WithTopOption(t *testing.T) {
	ctx := context.Background()
	col := msmodels.NewChatMessageCollectionResponse()
	api := &fakeChanAPI{listMsgsResp: col}
	svc := NewService(api, &fakeTeamRes{}, &fakeChannelRes{}, nil)

	var top int32 = 10
	opts := &models.ListMessagesOptions{Top: &top}
	_, err := svc.ListMessages(ctx, "team-1", "chan-1", opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_GetMessage_ReturnsMessage(t *testing.T) {
	ctx := context.Background()
	msg := newChatMessage("msg-42", "Test message")

	api := &fakeChanAPI{getMsgResp: msg}
	svc := NewService(api, &fakeTeamRes{}, &fakeChannelRes{}, nil)

	got, err := svc.GetMessage(ctx, "team-1", "chan-1", "msg-42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.ID != "msg-42" || got.Content != "Test message" {
		t.Errorf("unexpected message: %+v", got)
	}

	if api.lastMessageID != "msg-42" {
		t.Errorf("expected message ID 'msg-42', got %q", api.lastMessageID)
	}
}

func TestService_ListReplies_MapsReplies(t *testing.T) {
	ctx := context.Background()
	col := msmodels.NewChatMessageCollectionResponse()

	reply1 := newChatMessage("reply-1", "First reply")
	reply2 := newChatMessage("reply-2", "Second reply")
	col.SetValue([]msmodels.ChatMessageable{reply1, reply2})

	api := &fakeChanAPI{listRepliesResp: col}
	svc := NewService(api, &fakeTeamRes{}, &fakeChannelRes{}, nil)

	got, err := svc.ListReplies(ctx, "team-1", "chan-1", "msg-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 replies, got %d", len(got))
	}

	if got[0].ID != "reply-1" || got[0].Content != "First reply" {
		t.Errorf("unexpected first reply: %+v", got[0])
	}

	if api.lastMessageID != "msg-1" {
		t.Errorf("expected message ID 'msg-1', got %q", api.lastMessageID)
	}
}

func TestService_GetReply_ReturnsReply(t *testing.T) {
	ctx := context.Background()
	reply := newChatMessage("reply-42", "Test reply")

	api := &fakeChanAPI{getReplyResp: reply}
	svc := NewService(api, &fakeTeamRes{}, &fakeChannelRes{}, nil)

	got, err := svc.GetReply(ctx, "team-1", "chan-1", "msg-1", "reply-42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.ID != "reply-42" || got.Content != "Test reply" {
		t.Errorf("unexpected reply: %+v", got)
	}

	if api.lastMessageID != "msg-1" {
		t.Errorf("expected message ID 'msg-1', got %q", api.lastMessageID)
	}
	if api.lastReplyID != "reply-42" {
		t.Errorf("expected reply ID 'reply-42', got %q", api.lastReplyID)
	}
}

func TestMapChatMessageToMessage_NilInput(t *testing.T) {
	got := adapter.MapGraphMessage(nil)
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

func TestService_CreatePrivateChannel_Success(t *testing.T) {
	ctx := context.Background()
	created := newGraphChan("pc-1", "Secret channel")

	api := &fakeChanAPI{
		createResp: created,
	}
	tm := &fakeTeamRes{}
	cm := &fakeChannelRes{}
	svc := NewService(api, tm, cm, nil)

	memberRefs := []string{"user1", "user2"}
	ownerRefs := []string{"leader1"}

	got, err := svc.CreatePrivateChannel(ctx, "team-priv", "Secret channel", memberRefs, ownerRefs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got == nil {
		t.Fatalf("expected non-nil channel, got nil")
	}

	if got.ID != "pc-1" || got.Name != "Secret channel" {
		t.Errorf("unexpected result: %+v", got)
	}
	if got.IsGeneral {
		t.Errorf("expected IsGeneral=false for private channel, got true")
	}
	if tm.lastTeamName != "team-priv" {
		t.Errorf("expected mapper to be called with team-priv, got %q", tm.lastTeamName)
	}
	if api.lastTeamID != "team-priv" {
		t.Errorf("expected api to be called with team-priv, got %q", api.lastTeamID)
	}
}

func TestService_CreatePrivateChannel_MapperError(t *testing.T) {
	ctx := context.Background()
	mapErr := errors.New("mapper failed")

	api := &fakeChanAPI{}

	tm := &fakeTeamRes{resolveTeamErr: mapErr}
	cm := &fakeChannelRes{}

	svc := NewService(api, tm, cm, nil)

	_, err := svc.CreatePrivateChannel(ctx, "some-team", "Secret", nil, nil)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, mapErr) {
		t.Fatalf("expected mapper error %v, got %v", mapErr, err)
	}
}

func TestService_CreatePrivateChannel_MapsError(t *testing.T) {
	ctx := context.Background()
	api := &fakeChanAPI{
		createErr: &sender.RequestError{
			Code:    403,
			Message: "nope",
		},
	}

	tm := &fakeTeamRes{}
	cm := &fakeChannelRes{}

	svc := NewService(api, tm, cm, nil)

	_, err := svc.CreatePrivateChannel(ctx, "team-1", "Secret", nil, nil)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	var forbidden sender.ErrAccessForbidden
	if !errors.As(err, &forbidden) {
		t.Fatalf("expected ErrAccessForbidden, got %v", err)
	}
}

func TestService_ListMembers_MapsMembers(t *testing.T) {
	ctx := context.Background()

	col := msmodels.NewConversationMemberCollectionResponse()
	m1 := newAadUserMember("m-1", "u-1", "Alice", []string{"owner"})
	m2 := newAadUserMember("m-2", "u-2", "Bob", []string{"member"})
	col.SetValue([]msmodels.ConversationMemberable{m1, m2})

	api := &fakeChanAPI{membersResp: col}

	tm := &fakeTeamRes{}
	cm := &fakeChannelRes{}

	svc := NewService(api, tm, cm, nil)

	got, err := svc.ListMembers(ctx, "team-1", "chan-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 members, got %d", len(got))
	}

	if got[0].ID != "m-1" || got[0].UserID != "u-1" || got[0].DisplayName != "Alice" || got[0].Role != "owner" {
		t.Errorf("unexpected first member: %+v", got[0])
	}
	if got[1].ID != "m-2" || got[1].UserID != "u-2" || got[1].DisplayName != "Bob" || got[1].Role != "member" {
		t.Errorf("unexpected second member: %+v", got[1])
	}

	if tm.lastTeamName != "team-1" || cm.lastChannelName != "chan-1" {
		t.Errorf("expected mapper called with team-1/chan-1, got team=%q, chan=%q", tm.lastTeamName, cm.lastChannelName)
	}
	if api.lastTeamID != "team-1" || api.lastChanID != "chan-1" {
		t.Errorf("expected api called with team-1/chan-1, got team=%q, chan=%q", api.lastTeamID, api.lastChanID)
	}
}

func TestService_ListMembers_MapsError(t *testing.T) {
	ctx := context.Background()
	api := &fakeChanAPI{
		membersErr: &sender.RequestError{
			Code:    403,
			Message: "nope",
		},
	}

	tm := &fakeTeamRes{}
	cm := &fakeChannelRes{}

	svc := NewService(api, tm, cm, nil)

	_, err := svc.ListMembers(ctx, "team-1", "chan-1")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	var forbidden sender.ErrAccessForbidden
	if !errors.As(err, &forbidden) {
		t.Fatalf("expected ErrAccessForbidden, got %v", err)
	}
}

func TestService_AddMember_OwnerRole(t *testing.T) {
	ctx := context.Background()

	member := newAadUserMember("m-10", "u-10", "OwnerUser", []string{"owner"})
	api := &fakeChanAPI{addMemberResp: member}

	tm := &fakeTeamRes{}
	cm := &fakeChannelRes{}

	svc := NewService(api, tm, cm, nil)

	got, err := svc.AddMember(ctx, "team-1", "chan-1", "user-ref", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.ID != "m-10" || got.UserID != "u-10" || got.DisplayName != "OwnerUser" || got.Role != "owner" {
		t.Errorf("unexpected mapped member: %+v", got)
	}

	if api.lastAddUserRef != "user-ref" || api.lastAddRole != "owner" {
		t.Errorf("expected AddMember called with user-ref/owner, got user=%q role=%q", api.lastAddUserRef, api.lastAddRole)
	}
	if api.lastTeamID != "team-1" || api.lastChanID != "chan-1" {
		t.Errorf("expected api called with team-1/chan-1, got team=%q, chan=%q", api.lastTeamID, api.lastChanID)
	}
}

func TestService_AddMember_MapsError(t *testing.T) {
	ctx := context.Background()
	api := &fakeChanAPI{
		addMemberErr: &sender.RequestError{
			Code:    403,
			Message: "nope",
		},
	}

	tm := &fakeTeamRes{}
	cm := &fakeChannelRes{}

	svc := NewService(api, tm, cm, nil)

	_, err := svc.AddMember(ctx, "team-1", "chan-1", "user-ref", false)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	var forbidden sender.ErrAccessForbidden
	if !errors.As(err, &forbidden) {
		t.Fatalf("expected ErrAccessForbidden, got %v", err)
	}
}

func TestService_UpdateMemberRole_OwnerRole(t *testing.T) {
	ctx := context.Background()

	member := newAadUserMember("m-20", "u-20", "PromotedUser", []string{"owner"})
	api := &fakeChanAPI{updateMemberResp: member}

	tm := &fakeTeamRes{}
	cm := &fakeChannelRes{}

	svc := NewService(api, tm, cm, nil)

	got, err := svc.UpdateMemberRole(ctx, "team-1", "chan-1", "user-ref", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.ID != "m-20" || got.UserID != "u-20" || got.DisplayName != "PromotedUser" || got.Role != "owner" {
		t.Errorf("unexpected mapped member: %+v", got)
	}

	if cm.lastUserRef != "user-ref" || cm.lastTeamIDForUser != "team-1" || cm.lastChannelIDForUser != "chan-1" {
		t.Errorf("expected mapper called with team-1/chan-1/user-ref, got team=%q chan=%q user=%q",
			cm.lastTeamIDForUser, cm.lastChannelIDForUser, cm.lastUserRef)
	}

	if api.lastUpdateMemberID != "user-ref" || api.lastUpdateRole != "owner" {
		t.Errorf("expected UpdateMemberRole called with memberID=user-ref role=owner, got id=%q role=%q",
			api.lastUpdateMemberID, api.lastUpdateRole)
	}
}

func TestService_UpdateMemberRole_MapperError(t *testing.T) {
	ctx := context.Background()
	mapErr := errors.New("map user failed")
	api := &fakeChanAPI{}

	tm := &fakeTeamRes{}
	cm := &fakeChannelRes{resUserRefErr: mapErr}

	svc := NewService(api, tm, cm, nil)

	_, err := svc.UpdateMemberRole(ctx, "team-1", "chan-1", "user-ref", true)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, mapErr) {
		t.Fatalf("expected mapper error, got %v", err)
	}
}

func TestService_UpdateMemberRole_MapsError(t *testing.T) {
	ctx := context.Background()
	api := &fakeChanAPI{
		updateMemberErr: &sender.RequestError{
			Code:    403,
			Message: "nope",
		},
	}

	tm := &fakeTeamRes{}
	cm := &fakeChannelRes{}

	svc := NewService(api, tm, cm, nil)

	_, err := svc.UpdateMemberRole(ctx, "team-1", "chan-1", "user-ref", false)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	var forbidden sender.ErrAccessForbidden
	if !errors.As(err, &forbidden) {
		t.Fatalf("expected ErrAccessForbidden, got %v", err)
	}
}

func TestService_RemoveMember_Success(t *testing.T) {
	ctx := context.Background()

	api := &fakeChanAPI{}

	tm := &fakeTeamRes{}
	cm := &fakeChannelRes{}

	svc := NewService(api, tm, cm, nil)

	err := svc.RemoveMember(ctx, "team-1", "chan-1", "user-ref")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if api.lastRemoveMemberID != "user-ref" {
		t.Errorf("expected RemoveMember called with memberID=user-ref, got %q", api.lastRemoveMemberID)
	}
	if api.lastTeamID != "team-1" || api.lastChanID != "chan-1" {
		t.Errorf("expected api called with team-1/chan-1, got team=%q chan=%q", api.lastTeamID, api.lastChanID)
	}
}

func TestService_RemoveMember_MapsError(t *testing.T) {
	ctx := context.Background()
	api := &fakeChanAPI{
		removeMemberErr: &sender.RequestError{
			Code:    403,
			Message: "nope",
		},
	}

	tm := &fakeTeamRes{}
	cm := &fakeChannelRes{}

	svc := NewService(api, tm, cm, nil)

	err := svc.RemoveMember(ctx, "team-1", "chan-1", "user-ref")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	var forbidden sender.ErrAccessForbidden
	if !errors.As(err, &forbidden) {
		t.Fatalf("expected ErrAccessForbidden, got %v", err)
	}
}

func TestMapConversationMemberToChannelMember_NilInput(t *testing.T) {
	got := adapter.MapGraphMember(nil)
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

func TestMapConversationMemberToChannelMember_UserMember(t *testing.T) {
	member := newAadUserMember("m-99", "u-99", "Some User", []string{"owner"})
	got := adapter.MapGraphMember(member)
	if got == nil {
		t.Fatalf("expected non-nil, got nil")
	}
	if got.ID != "m-99" || got.UserID != "u-99" || got.DisplayName != "Some User" || got.Role != "owner" {
		t.Errorf("unexpected mapped member: %+v", got)
	}
}

func TestService_GetMentions_ResolvesUserTeamChannelDupsAllowed(t *testing.T) {
	ctx := context.Background()

	apiChan := &fakeChanAPI{}
	tm := &fakeTeamRes{}
	cr := &fakeChannelRes{}
	users := &fakeUsersAPI{
		byKey: map[string]msmodels.Userable{
			"alice@example.com": newGraphUser("u-1", "Alice A."),
		},
	}

	svc := NewService(apiChan, tm, cr, users)
	raw := []string{
		" alice@example.com ",
		"TEAM",
		"channel",
		"alice@example.com",
		"team",
	}

	got, err := svc.GetMentions(ctx, "team-1", "chan-1", raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 5 {
		t.Fatalf("expected 5 mentions, got %d: %#v", len(got), got)
	}

	if got[0].AtID != 0 || got[1].AtID != 1 || got[2].AtID != 2 {
		t.Fatalf("unexpected AtIDs: %#v", got)
	}

	if got[0].Kind != models.MentionUser || got[0].TargetID != "u-1" || got[0].Text != "Alice A." {
		t.Errorf("unexpected user mention: %+v", got[0])
	}

	if got[1].Kind != models.MentionTeam || got[1].TargetID != "team-1" || got[1].Text != "team-1" {
		t.Errorf("unexpected team mention: %+v", got[1])
	}

	if got[2].Kind != models.MentionChannel || got[2].TargetID != "chan-1" || got[2].Text != "chan-1" {
		t.Errorf("unexpected channel mention: %+v", got[2])
	}
}

func TestService_GetMentions_UnknownRefReturnsError(t *testing.T) {
	ctx := context.Background()

	svc := NewService(&fakeChanAPI{}, &fakeTeamRes{}, &fakeChannelRes{}, &fakeUsersAPI{
		byKey: map[string]msmodels.Userable{},
	})

	_, err := svc.GetMentions(ctx, "team-1", "chan-1", []string{"something-else"})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestService_GetMentions_UserAPIErrorPropagates(t *testing.T) {
	ctx := context.Background()

	users := &fakeUsersAPI{
		err: &sender.RequestError{Code: 500, Message: "boom"},
	}
	svc := NewService(&fakeChanAPI{}, &fakeTeamRes{}, &fakeChannelRes{}, users)

	_, err := svc.GetMentions(ctx, "team-1", "chan-1", []string{"alice@example.com"})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestService_GetMentions_UserMissingDisplayNameReturnsError(t *testing.T) {
	ctx := context.Background()

	users := &fakeUsersAPI{
		byKey: map[string]msmodels.Userable{
			"alice@example.com": newGraphUser("u-1", ""),
		},
	}
	svc := NewService(&fakeChanAPI{}, &fakeTeamRes{}, &fakeChannelRes{}, users)

	_, err := svc.GetMentions(ctx, "team-1", "chan-1", []string{"alice@example.com"})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestService_SendReply_Success(t *testing.T) {
	ctx := context.Background()
	replyID := "reply-99"
	content := "Reply content"

	respMsg := newChatMessage(replyID, content)
	api := &fakeChanAPI{sendMsgResp: respMsg}

	svc := NewService(api, &fakeTeamRes{}, &fakeChannelRes{}, nil)

	body := models.MessageBody{Content: content, ContentType: models.MessageContentTypeText}

	got, err := svc.SendReply(ctx, "team-1", "chan-1", "msg-parent", body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.ID != replyID {
		t.Errorf("expected ID %q, got %q", replyID, got.ID)
	}
	if api.lastMessageID != "msg-parent" {
		t.Errorf("expected parent message ID 'msg-parent', got %q", api.lastMessageID)
	}
	if api.lastContent != content {
		t.Errorf("expected content %q, got %q", content, api.lastContent)
	}
}

func TestService_Delete_Success(t *testing.T) {
	ctx := context.Background()
	api := &fakeChanAPI{}
	svc := NewService(api, &fakeTeamRes{}, &fakeChannelRes{}, nil)

	err := svc.Delete(ctx, "team-1", "chan-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if api.lastTeamID != "team-1" || api.lastChanID != "chan-1" {
		t.Errorf("expected delete called with team-1/chan-1, got %q/%q", api.lastTeamID, api.lastChanID)
	}
}

func TestService_Get_MapsError(t *testing.T) {
	ctx := context.Background()
	api := &fakeChanAPI{
		getErr: &sender.RequestError{Code: 404, Message: "Not Found"},
	}
	svc := NewService(api, &fakeTeamRes{}, &fakeChannelRes{}, nil)

	_, err := svc.Get(ctx, "team-1", "chan-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
