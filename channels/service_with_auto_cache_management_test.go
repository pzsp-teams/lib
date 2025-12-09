package channels

import (
	"context"
	"strings"
	"testing"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/cacher"
	snd "github.com/pzsp-teams/lib/internal/sender"
)

type fakeCacher struct {
	setCalls        int
	invalidateCalls int

	setKeys        []string
	setValues      []any
	invalidateKeys []string
}

func (f *fakeCacher) Get(key string) (value any, found bool, err error) {
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

type fakeTeamResolver struct {
	calls       int
	lastTeamRef string
	resolveFunc func(ctx context.Context, teamRef string) (string, error)
}

func (f *fakeTeamResolver) ResolveTeamRefToID(ctx context.Context, teamRef string) (string, error) {
	f.calls++
	f.lastTeamRef = teamRef
	if f.resolveFunc != nil {
		return f.resolveFunc(ctx, teamRef)
	}
	return "team-id-default", nil
}

type fakeChannelResolver struct {
	resolveChannelFunc func(ctx context.Context, teamID, channelRef string) (string, error)
	resolveUserFunc    func(ctx context.Context, teamID, channelID, userRef string) (string, error)

	lastTeamID     string
	lastChannelID  string
	lastChannelRef string
	lastUserRef    string
}

func (f *fakeChannelResolver) ResolveChannelRefToID(ctx context.Context, teamID, channelRef string) (string, error) {
	f.lastTeamID = teamID
	f.lastChannelRef = channelRef
	if f.resolveChannelFunc != nil {
		return f.resolveChannelFunc(ctx, teamID, channelRef)
	}
	return "channel-id-default", nil
}

func (f *fakeChannelResolver) ResolveUserRefToMemberID(ctx context.Context, teamID, channelID, userRef string) (string, error) {
	f.lastTeamID = teamID
	f.lastChannelID = channelID
	f.lastUserRef = userRef
	if f.resolveUserFunc != nil {
		return f.resolveUserFunc(ctx, teamID, channelID, userRef)
	}
	return "member-id-default", nil
}

type fakeChannelAPI struct {
	listChannelsFunc                func(ctx context.Context, teamID string) (msmodels.ChannelCollectionResponseable, *snd.RequestError)
	getChannelFunc                  func(ctx context.Context, teamID, channelID string) (msmodels.Channelable, *snd.RequestError)
	createStandardChannelFunc       func(ctx context.Context, teamID string, channel msmodels.Channelable) (msmodels.Channelable, *snd.RequestError)
	createPrivateChannelWithMembers func(ctx context.Context, teamID, name string, memberRefs, ownerRefs []string) (msmodels.Channelable, *snd.RequestError)
	deleteChannelFunc               func(ctx context.Context, teamID, channelID string) *snd.RequestError
	addMemberFunc                   func(ctx context.Context, teamID, channelID, userRef, role string) (msmodels.ConversationMemberable, *snd.RequestError)
	removeMemberFunc                func(ctx context.Context, teamID, channelID, memberID string) *snd.RequestError
}

func (f *fakeChannelAPI) ListChannels(ctx context.Context, teamID string) (msmodels.ChannelCollectionResponseable, *snd.RequestError) {
	if f.listChannelsFunc != nil {
		return f.listChannelsFunc(ctx, teamID)
	}
	return nil, nil
}

func (f *fakeChannelAPI) GetChannel(ctx context.Context, teamID, channelID string) (msmodels.Channelable, *snd.RequestError) {
	if f.getChannelFunc != nil {
		return f.getChannelFunc(ctx, teamID, channelID)
	}
	return nil, nil
}

func (f *fakeChannelAPI) CreateStandardChannel(ctx context.Context, teamID string, channel msmodels.Channelable) (msmodels.Channelable, *snd.RequestError) {
	if f.createStandardChannelFunc != nil {
		return f.createStandardChannelFunc(ctx, teamID, channel)
	}
	return nil, nil
}

func (f *fakeChannelAPI) CreatePrivateChannelWithMembers(ctx context.Context, teamID, name string, memberRefs, ownerRefs []string) (msmodels.Channelable, *snd.RequestError) {
	if f.createPrivateChannelWithMembers != nil {
		return f.createPrivateChannelWithMembers(ctx, teamID, name, memberRefs, ownerRefs)
	}
	return nil, nil
}

func (f *fakeChannelAPI) DeleteChannel(ctx context.Context, teamID, channelID string) *snd.RequestError {
	if f.deleteChannelFunc != nil {
		return f.deleteChannelFunc(ctx, teamID, channelID)
	}
	return nil
}

func (f *fakeChannelAPI) SendMessage(ctx context.Context, teamID, channelID string, msg msmodels.ChatMessageable) (msmodels.ChatMessageable, *snd.RequestError) {
	return nil, nil
}

func (f *fakeChannelAPI) ListMessages(ctx context.Context, teamID, channelID string, top *int32) (msmodels.ChatMessageCollectionResponseable, *snd.RequestError) {
	return nil, nil
}

func (f *fakeChannelAPI) GetMessage(ctx context.Context, teamID, channelID, messageID string) (msmodels.ChatMessageable, *snd.RequestError) {
	return nil, nil
}

func (f *fakeChannelAPI) ListReplies(ctx context.Context, teamID, channelID, messageID string, top *int32) (msmodels.ChatMessageCollectionResponseable, *snd.RequestError) {
	return nil, nil
}

func (f *fakeChannelAPI) GetReply(ctx context.Context, teamID, channelID, messageID, replyID string) (msmodels.ChatMessageable, *snd.RequestError) {
	return nil, nil
}

func (f *fakeChannelAPI) ListMembers(ctx context.Context, teamID, channelID string) (msmodels.ConversationMemberCollectionResponseable, *snd.RequestError) {
	return nil, nil
}

func (f *fakeChannelAPI) AddMember(ctx context.Context, teamID, channelID, userRef, role string) (msmodels.ConversationMemberable, *snd.RequestError) {
	if f.addMemberFunc != nil {
		return f.addMemberFunc(ctx, teamID, channelID, userRef, role)
	}
	return nil, nil
}

func (f *fakeChannelAPI) UpdateMemberRole(ctx context.Context, teamID, channelID, memberID, role string) (msmodels.ConversationMemberable, *snd.RequestError) {
	return nil, nil
}

func (f *fakeChannelAPI) RemoveMember(ctx context.Context, teamID, channelID, memberID string) *snd.RequestError {
	if f.removeMemberFunc != nil {
		return f.removeMemberFunc(ctx, teamID, channelID, memberID)
	}
	return nil
}

func newChannelCollection(chans ...msmodels.Channelable) msmodels.ChannelCollectionResponseable {
	resp := msmodels.NewChannelCollectionResponse()
	resp.SetValue(chans)
	return resp
}

func newConversationMember(id, userID, displayName string, roles []string) msmodels.ConversationMemberable {
	m := msmodels.NewAadUserConversationMember()
	m.SetId(&id)
	m.SetRoles(roles)
	m.SetUserId(&userID)
	m.SetDisplayName(&displayName)
	return m
}

func TestServiceWithAutoCacheManagement_ListChannels_WarmsCache(t *testing.T) {
	ctx := context.Background()

	fc := &fakeCacher{}
	fr := &fakeTeamResolver{
		resolveFunc: func(ctx context.Context, teamRef string) (string, error) {
			if teamRef != "my-team" {
				t.Errorf("expected teamRef=my-team, got %q", teamRef)
			}
			return "team-id-123", nil
		},
	}
	cr := &fakeChannelRes{}
	fapi := &fakeChannelAPI{
		listChannelsFunc: func(ctx context.Context, teamID string) (msmodels.ChannelCollectionResponseable, *snd.RequestError) {
			if teamID != "team-id-123" {
				t.Errorf("expected teamID=team-id-123, got %q", teamID)
			}
			return newChannelCollection(
				newGraphChan("chan-1", "  General  "),
				newGraphChan("chan-2", "My Channel"),
			), nil
		},
	}

	svc := &service{
		channelAPI:      fapi,
		teamResolver:    fr,
		channelResolver: cr,
	}
	decor := &serviceWithAutoCacheManagement{
		svc:   svc,
		cache: fc,
		run:   func(fn func()) { fn() },
	}

	chans, err := decor.ListChannels(ctx, "my-team")
	if err != nil {
		t.Fatalf("unexpected error from ListChannels: %v", err)
	}
	if len(chans) != 2 {
		t.Fatalf("expected 2 channels, got %d", len(chans))
	}

	if fc.setCalls != 2 {
		t.Fatalf("expected 2 Set calls, got %d", fc.setCalls)
	}

	expectedKeys := []string{
		cacher.NewChannelKeyBuilder("team-id-123", "General").ToString(),
		cacher.NewChannelKeyBuilder("team-id-123", "My Channel").ToString(),
	}
	if len(fc.setKeys) != len(expectedKeys) {
		t.Fatalf("expected %d cache keys, got %d", len(expectedKeys), len(fc.setKeys))
	}
	for i, exp := range expectedKeys {
		if fc.setKeys[i] != exp {
			t.Errorf("at index %d expected cache key %q, got %q", i, exp, fc.setKeys[i])
		}
	}
	if v, ok := fc.setValues[0].(string); !ok || v != "chan-1" {
		t.Errorf("expected cached value chan-1, got %#v", fc.setValues[0])
	}
	if v, ok := fc.setValues[1].(string); !ok || v != "chan-2" {
		t.Errorf("expected cached value chan-2, got %#v", fc.setValues[1])
	}
}

func TestServiceWithAutoCacheManagement_Get_WarmsCache(t *testing.T) {
	ctx := context.Background()

	fc := &fakeCacher{}
	fr := &fakeTeamResolver{
		resolveFunc: func(ctx context.Context, teamRef string) (string, error) {
			if teamRef != "team-ref" {
				t.Errorf("expected teamRef=team-ref, got %q", teamRef)
			}
			return "team-id-xyz", nil
		},
	}
	cr := &fakeChannelResolver{
		resolveChannelFunc: func(ctx context.Context, teamID, channelRef string) (string, error) {
			if teamID != "team-id-xyz" {
				t.Errorf("expected teamID=team-id-xyz, got %q", teamID)
			}
			if strings.TrimSpace(channelRef) != "my-channel" {
				t.Errorf("expected channelRef=my-channel, got %q", channelRef)
			}
			return "channel-id-1", nil
		},
	}
	fapi := &fakeChannelAPI{
		getChannelFunc: func(ctx context.Context, teamID, channelID string) (msmodels.Channelable, *snd.RequestError) {
			if teamID != "team-id-xyz" {
				t.Errorf("expected teamID=team-id-xyz, got %q", teamID)
			}
			if channelID != "channel-id-1" {
				t.Errorf("expected channelID=channel-id-1, got %q", channelID)
			}
			return newGraphChan("channel-id-1", "  My Channel  "), nil
		},
	}

	svc := &service{
		channelAPI:      fapi,
		teamResolver:    fr,
		channelResolver: cr,
	}
	decor := &serviceWithAutoCacheManagement{
		svc:   svc,
		cache: fc,
		run:   func(fn func()) { fn() },
	}

	ch, err := decor.Get(ctx, "team-ref", "my-channel")
	if err != nil {
		t.Fatalf("unexpected error from Get: %v", err)
	}
	if ch == nil || ch.ID != "channel-id-1" || strings.TrimSpace(ch.Name) != "My Channel" {
		t.Fatalf("unexpected channel returned: %#v", ch)
	}

	if fc.setCalls != 1 {
		t.Fatalf("expected 1 Set call, got %d", fc.setCalls)
	}
	expectedKey := cacher.NewChannelKeyBuilder("team-id-xyz", "My Channel").ToString()
	if fc.setKeys[0] != expectedKey {
		t.Errorf("expected cache key %q, got %q", expectedKey, fc.setKeys[0])
	}
	if v, ok := fc.setValues[0].(string); !ok || v != "channel-id-1" {
		t.Errorf("expected cached value channel-id-1, got %#v", fc.setValues[0])
	}
}

func TestServiceWithAutoCacheManagement_CreateStandardChannel_InvalidatesAndCaches(t *testing.T) {
	ctx := context.Background()

	fc := &fakeCacher{}
	fr := &fakeTeamResolver{
		resolveFunc: func(ctx context.Context, teamRef string) (string, error) {
			if strings.TrimSpace(teamRef) != "Team Name" {
				t.Errorf("expected teamRef Team Name, got %q", teamRef)
			}
			return "team-id-1", nil
		},
	}
	cr := &fakeChannelRes{}
	fapi := &fakeChannelAPI{
		createStandardChannelFunc: func(ctx context.Context, teamID string, ch msmodels.Channelable) (msmodels.Channelable, *snd.RequestError) {
			if teamID != "team-id-1" {
				t.Errorf("expected teamID team-id-1, got %q", teamID)
			}
			return newGraphChan("new-channel-id", "New Channel"), nil
		},
	}

	svc := &service{
		channelAPI:      fapi,
		teamResolver:    fr,
		channelResolver: cr,
	}
	decor := &serviceWithAutoCacheManagement{
		svc:   svc,
		cache: fc,
		run:   func(fn func()) { fn() },
	}

	ch, err := decor.CreateStandardChannel(ctx, "  Team Name  ", "New Channel")
	if err != nil {
		t.Fatalf("unexpected error from CreateStandardChannel: %v", err)
	}
	if ch == nil || ch.ID != "new-channel-id" || strings.TrimSpace(ch.Name) != "New Channel" {
		t.Fatalf("unexpected channel: %#v", ch)
	}

	if fc.invalidateCalls != 1 {
		t.Fatalf("expected 1 Invalidate call, got %d", fc.invalidateCalls)
	}
	invalidateKey := cacher.NewChannelKeyBuilder("team-id-1", "New Channel").ToString()
	if fc.invalidateKeys[0] != invalidateKey {
		t.Errorf("expected invalidate key %q, got %q", invalidateKey, fc.invalidateKeys[0])
	}

	if fc.setCalls != 1 {
		t.Fatalf("expected 1 Set call, got %d", fc.setCalls)
	}
	setKey := cacher.NewChannelKeyBuilder("team-id-1", "New Channel").ToString()
	if fc.setKeys[0] != setKey {
		t.Errorf("expected Set key %q, got %q", setKey, fc.setKeys[0])
	}
	if v, ok := fc.setValues[0].(string); !ok || v != "new-channel-id" {
		t.Errorf("expected cached value new-channel-id, got %#v", fc.setValues[0])
	}
}

func TestServiceWithAutoCacheManagement_Delete_InvalidatesCache(t *testing.T) {
	ctx := context.Background()

	fc := &fakeCacher{}
	fr := &fakeTeamResolver{
		resolveFunc: func(ctx context.Context, teamRef string) (string, error) {
			if teamRef != "my-team" {
				t.Errorf("expected teamRef my-team, got %q", teamRef)
			}
			return "team-id-del", nil
		},
	}
	cr := &fakeChannelResolver{
		resolveChannelFunc: func(ctx context.Context, teamID, channelRef string) (string, error) {
			if teamID != "team-id-del" {
				t.Errorf("expected teamID team-id-del, got %q", teamID)
			}
			if channelRef != "Channel To Delete" {
				t.Errorf("expected channelRef Channel To Delete, got %q", channelRef)
			}
			return "channel-id-del", nil
		},
	}
	fapi := &fakeChannelAPI{
		deleteChannelFunc: func(ctx context.Context, teamID, channelID string) *snd.RequestError {
			if teamID != "team-id-del" {
				t.Errorf("expected teamID team-id-del, got %q", teamID)
			}
			if channelID != "channel-id-del" {
				t.Errorf("expected channelID channel-id-del, got %q", channelID)
			}
			return nil
		},
	}

	svc := &service{
		channelAPI:      fapi,
		teamResolver:    fr,
		channelResolver: cr,
	}
	decor := &serviceWithAutoCacheManagement{
		svc:   svc,
		cache: fc,
		run:   func(fn func()) { fn() },
	}

	if err := decor.Delete(ctx, "my-team", "Channel To Delete"); err != nil {
		t.Fatalf("unexpected error from Delete: %v", err)
	}

	if fc.invalidateCalls != 1 {
		t.Fatalf("expected 1 Invalidate call, got %d", fc.invalidateCalls)
	}
	expectedKey := cacher.NewChannelKeyBuilder("team-id-del", "Channel To Delete").ToString()
	if fc.invalidateKeys[0] != expectedKey {
		t.Errorf("expected invalidate key %q, got %q", expectedKey, fc.invalidateKeys[0])
	}
	if fc.setCalls != 0 {
		t.Fatalf("expected 0 Set calls, got %d", fc.setCalls)
	}
}

func TestServiceWithAutoCacheManagement_AddMember_CachesMemberMapping(t *testing.T) {
	ctx := context.Background()

	fc := &fakeCacher{}
	fr := &fakeTeamResolver{
		resolveFunc: func(ctx context.Context, teamRef string) (string, error) {
			if teamRef != "team-ref" {
				t.Errorf("expected teamRef team-ref, got %q", teamRef)
			}
			return "team-id-1", nil
		},
	}
	cr := &fakeChannelResolver{
		resolveChannelFunc: func(ctx context.Context, teamID, channelRef string) (string, error) {
			if teamID != "team-id-1" {
				t.Errorf("expected teamID team-id-1, got %q", teamID)
			}
			if channelRef != "channel-ref" {
				t.Errorf("expected channelRef channel-ref, got %q", channelRef)
			}
			return "channel-id-1", nil
		},
	}
	fapi := &fakeChannelAPI{
		addMemberFunc: func(ctx context.Context, teamID, channelID, userRef, role string) (msmodels.ConversationMemberable, *snd.RequestError) {
			if teamID != "team-id-1" {
				t.Errorf("expected teamID team-id-1, got %q", teamID)
			}
			if channelID != "channel-id-1" {
				t.Errorf("expected channelID channel-id-1, got %q", channelID)
			}
			if strings.TrimSpace(userRef) != "user@example.com" {
				t.Errorf("expected userRef user@example.com, got %q", userRef)
			}
			return newConversationMember("member-id-1", "user-id-42", "User Name", []string{"member"}), nil
		},
	}

	svc := &service{
		channelAPI:      fapi,
		teamResolver:    fr,
		channelResolver: cr,
	}
	decor := &serviceWithAutoCacheManagement{
		svc:   svc,
		cache: fc,
		run:   func(fn func()) { fn() },
	}

	member, err := decor.AddMember(ctx, "team-ref", "channel-ref", "  user@example.com  ", false)
	if err != nil {
		t.Fatalf("unexpected error from AddMember: %v", err)
	}
	if member == nil || member.ID == "" {
		t.Fatalf("expected non-nil member with ID, got %#v", member)
	}

	if fc.setCalls != 1 {
		t.Fatalf("expected 1 Set call, got %d", fc.setCalls)
	}
	expectedKey := cacher.NewMemberKeyBuilder("user@example.com", "team-id-1", "channel-id-1").ToString()
	if fc.setKeys[0] != expectedKey {
		t.Errorf("expected member cache key %q, got %q", expectedKey, fc.setKeys[0])
	}
	if v, ok := fc.setValues[0].(string); !ok || v != member.ID {
		t.Errorf("expected cached value %q, got %#v", member.ID, fc.setValues[0])
	}
}

func TestServiceWithAutoCacheManagement_RemoveMember_InvalidatesMemberMapping(t *testing.T) {
	ctx := context.Background()

	fc := &fakeCacher{}
	fr := &fakeTeamResolver{
		resolveFunc: func(ctx context.Context, teamRef string) (string, error) {
			return "team-id-1", nil
		},
	}
	cr := &fakeChannelResolver{
		resolveChannelFunc: func(ctx context.Context, teamID, channelRef string) (string, error) {
			return "channel-id-1", nil
		},
		resolveUserFunc: func(ctx context.Context, teamID, channelID, userRef string) (string, error) {
			return "member-id-1", nil
		},
	}
	fapi := &fakeChannelAPI{
		removeMemberFunc: func(ctx context.Context, teamID, channelID, memberID string) *snd.RequestError {
			if teamID != "team-id-1" || channelID != "channel-id-1" || memberID != "member-id-1" {
				t.Errorf("unexpected RemoveMember args: teamID=%q channelID=%q memberID=%q", teamID, channelID, memberID)
			}
			return nil
		},
	}

	svc := &service{
		channelAPI:      fapi,
		teamResolver:    fr,
		channelResolver: cr,
	}
	decor := &serviceWithAutoCacheManagement{
		svc:   svc,
		cache: fc,
		run:   func(fn func()) { fn() },
	}

	if err := decor.RemoveMember(ctx, "team-ref", "channel-ref", "user@example.com"); err != nil {
		t.Fatalf("unexpected error from RemoveMember: %v", err)
	}

	if fc.invalidateCalls != 1 {
		t.Fatalf("expected 1 Invalidate call, got %d", fc.invalidateCalls)
	}
	expectedKey := cacher.NewMemberKeyBuilder("user@example.com", "team-id-1", "channel-id-1").ToString()
	if fc.invalidateKeys[0] != expectedKey {
		t.Errorf("expected invalidate key %q, got %q", expectedKey, fc.invalidateKeys[0])
	}
	if fc.setCalls != 0 {
		t.Fatalf("expected 0 Set calls, got %d", fc.setCalls)
	}
}
