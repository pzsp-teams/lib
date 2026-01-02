package mentions

import (
	"strings"
	"testing"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/models"
	"github.com/stretchr/testify/require"
)

const (
	okGUID   = "00000000-0000-0000-0000-000000000001"
	okThread = "19:thread@thread.tacv2"
)

func TestMapMentions_EmptyInput_ReturnsNil(t *testing.T) {
	got, err := mapMentions(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil, got %#v", got)
	}
}

func TestMapMentions_MapsUserMention(t *testing.T) {
	in := []models.Mention{
		{Kind: models.MentionUser, AtID: 0, Text: "Alice", TargetID: okGUID},
	}

	got, err := mapMentions(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 mention, got %d", len(got))
	}

	m := got[0].(*msmodels.ChatMessageMention)
	if m.GetId() == nil || *m.GetId() != 0 {
		t.Fatalf("expected id=0, got %#v", m.GetId())
	}
	if m.GetMentionText() == nil || *m.GetMentionText() != "Alice" {
		t.Fatalf("expected mentionText=Alice, got %#v", m.GetMentionText())
	}

	mentioned := m.GetMentioned()
	if mentioned == nil || mentioned.GetUser() == nil {
		t.Fatalf("expected mentioned.user to be set")
	}
	u := mentioned.GetUser()
	if u.GetId() == nil || *u.GetId() != okGUID {
		t.Fatalf("expected user.id=%q, got %#v", okGUID, u.GetId())
	}
}

func TestMapMentions_MapsConversationMentions_ChannelTeamEveryone(t *testing.T) {
	in := []models.Mention{
		{Kind: models.MentionChannel, AtID: 0, Text: "Channel", TargetID: okThread},
		{Kind: models.MentionTeam, AtID: 1, Text: "Team", TargetID: okGUID},
		{Kind: models.MentionEveryone, AtID: 2, Text: "Everyone", TargetID: okThread},
	}

	got, err := mapMentions(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 mentions, got %d", len(got))
	}

	m0 := got[0].(*msmodels.ChatMessageMention)
	if m0.GetMentioned() == nil || m0.GetMentioned().GetConversation() == nil {
		t.Fatalf("expected channel mention to have mentioned.conversation")
	}
	c0 := m0.GetMentioned().GetConversation()
	if c0.GetId() == nil || *c0.GetId() != okThread {
		t.Fatalf("expected channel conversation.id=%q, got %#v", okThread, c0.GetId())
	}

	m1 := got[1].(*msmodels.ChatMessageMention)
	c1 := m1.GetMentioned().GetConversation()
	if c1 == nil {
		t.Fatalf("expected team mention to have mentioned.conversation")
	}
	if c1.GetId() == nil || *c1.GetId() != okGUID {
		t.Fatalf("expected team conversation.id=%q, got %#v", okGUID, c1.GetId())
	}

	m2 := got[2].(*msmodels.ChatMessageMention)
	c2 := m2.GetMentioned().GetConversation()
	if c2 == nil {
		t.Fatalf("expected everyone mention to have mentioned.conversation")
	}
	if c2.GetId() == nil || *c2.GetId() != okThread {
		t.Fatalf("expected everyone conversation.id=%q, got %#v", okThread, c2.GetId())
	}
}

func TestMapMentions_ErrorIsWrappedWithIndex(t *testing.T) {
	in := []models.Mention{
		{Kind: models.MentionUser, AtID: 0, Text: "Alice", TargetID: "not-a-guid"},
	}

	_, err := mapMentions(in)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "mention[0]:") {
		t.Fatalf("expected error to contain mention[0]: prefix, got %v", err)
	}
}

func TestValidateAtTags_NilBody_OK(t *testing.T) {
	if err := validateAtTags(nil); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestValidateAtTags_NonHTML_OK(t *testing.T) {
	b := &models.MessageBody{
		ContentType: models.MessageContentTypeText,
		Content:     `<at id="0">Alice</at>`,
		Mentions:    []models.Mention{{Kind: models.MentionUser, AtID: 0, Text: "Alice", TargetID: okGUID}},
	}
	if err := validateAtTags(b); err != nil {
		t.Fatalf("expected nil error for non-HTML, got %v", err)
	}
}

func TestValidateAtTags_AtTagButNoMentions_Error(t *testing.T) {
	b := &models.MessageBody{
		ContentType: models.MessageContentTypeHTML,
		Content:     `<at id="0">Alice</at>`,
		Mentions:    nil,
	}
	err := validateAtTags(b)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "mentions list is empty") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateAtTags_DuplicateAtID_Error(t *testing.T) {
	b := &models.MessageBody{
		ContentType: models.MessageContentTypeHTML,
		Content:     `<at id="0">A</at> <at id="0">B</at>`,
		Mentions: []models.Mention{
			{Kind: models.MentionUser, AtID: 0, Text: "A", TargetID: okGUID},
			{Kind: models.MentionUser, AtID: 0, Text: "B", TargetID: okGUID},
		},
	}
	err := validateAtTags(b)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "duplicate at-id 0") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateAtTags_MissingSpecificAtIDTag_Error(t *testing.T) {
	b := &models.MessageBody{
		ContentType: models.MessageContentTypeHTML,
		Content:     `<at id="0">A</at>`,
		Mentions: []models.Mention{
			{Kind: models.MentionUser, AtID: 0, Text: "A", TargetID: okGUID},
			{Kind: models.MentionUser, AtID: 1, Text: "B", TargetID: okGUID},
		},
	}
	err := validateAtTags(b)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), `missing <at id="1"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateAtTags_HappyPath_OK(t *testing.T) {
	b := &models.MessageBody{
		ContentType: models.MessageContentTypeHTML,
		Content:     `Hello <at id="0">Alice</at> and <at id="1">Bob</at>`,
		Mentions: []models.Mention{
			{Kind: models.MentionUser, AtID: 0, Text: "Alice", TargetID: okGUID},
			{Kind: models.MentionUser, AtID: 1, Text: "Bob", TargetID: okGUID},
		},
	}
	if err := validateAtTags(b); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMapToGraphMention_ValidationErrors(t *testing.T) {
	tests := []struct {
		name string
		in   models.Mention
		want string
	}{
		{"missing Text", models.Mention{Kind: models.MentionUser, AtID: 0, Text: "", TargetID: okGUID}, "missing Text"},
		{"invalid AtID", models.Mention{Kind: models.MentionUser, AtID: -1, Text: "A", TargetID: okGUID}, "invalid AtID"},
		{"missing TargetID", models.Mention{Kind: models.MentionUser, AtID: 0, Text: "A", TargetID: ""}, "missing TargetID"},
		{"unsupported kind", models.Mention{Kind: models.MentionKind("weird"), AtID: 0, Text: "A", TargetID: okGUID}, "unsupported MentionKind"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := mapToGraphMention(tt.in)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected error to contain %q, got %v", tt.want, err)
			}
		})
	}
}

func TestBuildUserMentioned_InvalidTargetID(t *testing.T) {
	_, err := buildUserMentioned("not-a-guid", "Alice")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "invalid TargetID for user mention") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBuildUserMentioned_SetsFields(t *testing.T) {
	u, err := buildUserMentioned(okGUID, "Alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.GetId() == nil || *u.GetId() != okGUID {
		t.Fatalf("expected id=%q, got %#v", okGUID, u.GetId())
	}
	if u.GetDisplayName() == nil || *u.GetDisplayName() != "Alice" {
		t.Fatalf("expected displayName=Alice, got %#v", u.GetDisplayName())
	}
	ad := u.GetAdditionalData()
	if ad == nil || ad["userIdentityType"] != "aadUser" {
		t.Fatalf("expected additionalData.userIdentityType=aadUser, got %#v", ad)
	}
}

func TestBuildConversationMentioned_ValidateFalse_Error(t *testing.T) {
	validate := func(string) bool { return false }
	_, err := buildConversationMentioned("x", "Name", msmodels.CHAT_TEAMWORKCONVERSATIONIDENTITYTYPE, validate, "label")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "invalid TargetID for label") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBuildConversationMentioned_SetsFieldsAndType(t *testing.T) {
	validate := func(string) bool { return true }
	typ := msmodels.CHANNEL_TEAMWORKCONVERSATIONIDENTITYTYPE

	conv, err := buildConversationMentioned("conv-id", "Conv", typ, validate, "channel mention")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conv.GetId() == nil || *conv.GetId() != "conv-id" {
		t.Fatalf("expected id=conv-id, got %#v", conv.GetId())
	}
	if conv.GetDisplayName() == nil || *conv.GetDisplayName() != "Conv" {
		t.Fatalf("expected displayName=Conv, got %#v", conv.GetDisplayName())
	}
	if conv.GetConversationIdentityType() == nil || *conv.GetConversationIdentityType() != typ {
		t.Fatalf("expected conversationIdentityType=%v, got %#v", typ, conv.GetConversationIdentityType())
	}
}

func TestPrepareMentions_NoMentions_ReturnsNilNil(t *testing.T) {
	body := &models.MessageBody{
		Content:     "hello",
		ContentType: models.MessageContentTypeText,
		Mentions:    nil,
	}

	got, err := PrepareMentions(body)
	require.NoError(t, err)
	require.Nil(t, got)
}

func TestPrepareMentions_MentionsWithNonHTML_ReturnsError(t *testing.T) {
	body := &models.MessageBody{
		Content:     "hello",
		ContentType: models.MessageContentTypeText,
		Mentions: []models.Mention{
			{Kind: models.MentionUser, AtID: 0, Text: "Alice", TargetID: "00000000-0000-0000-0000-000000000001"},
		},
	}

	_, err := PrepareMentions(body)
	require.Error(t, err)
	require.Contains(t, err.Error(), "mentions can only be used with HTML content type")
}

func TestPrepareMentions_AtTagsPresentButMentionsEmpty_ReturnsError(t *testing.T) {
	body := &models.MessageBody{
		Content:     `<div>Hello <at id="0">Alice</at></div>`,
		ContentType: models.MessageContentTypeHTML,
		Mentions:    nil,
	}

	_, err := PrepareMentions(body)
	require.Error(t, err)
	require.Contains(t, err.Error(), "content contains <at> tags but mentions list is empty")
}

func TestPrepareMentions_MentionMissingAtTag_ReturnsError(t *testing.T) {
	body := &models.MessageBody{
		Content:     `<div>Hello world</div>`,
		ContentType: models.MessageContentTypeHTML,
		Mentions: []models.Mention{
			{Kind: models.MentionUser, AtID: 0, Text: "Alice", TargetID: "00000000-0000-0000-0000-000000000001"},
		},
	}

	_, err := PrepareMentions(body)
	require.Error(t, err)
	require.Contains(t, err.Error(), `missing <at id="0"`)
}

func TestPrepareMentions_DuplicateAtID_ReturnsError(t *testing.T) {
	body := &models.MessageBody{
		Content:     `<div><at id="0">Alice</at> and <at id="0">Bob</at></div>`,
		ContentType: models.MessageContentTypeHTML,
		Mentions: []models.Mention{
			{Kind: models.MentionUser, AtID: 0, Text: "Alice", TargetID: "00000000-0000-0000-0000-000000000001"},
			{Kind: models.MentionUser, AtID: 0, Text: "Bob", TargetID: "00000000-0000-0000-0000-000000000002"},
		},
	}

	_, err := PrepareMentions(body)
	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicate at-id 0")
}

func TestPrepareMentions_MapMentionsWrapsIndexError(t *testing.T) {
	body := &models.MessageBody{
		Content:     `<div><at id="-1">Alice</at></div>`,
		ContentType: models.MessageContentTypeHTML,
		Mentions: []models.Mention{
			{Kind: models.MentionUser, AtID: -1, Text: "Alice", TargetID: "00000000-0000-0000-0000-000000000001"},
		},
	}

	_, err := PrepareMentions(body)
	require.Error(t, err)
	require.Contains(t, err.Error(), "mention[0]:")
	require.Contains(t, err.Error(), "invalid AtID")
}

func TestPrepareMentions_UserMention_Success(t *testing.T) {
	userID := "00000000-0000-0000-0000-000000000001"

	body := &models.MessageBody{
		Content:     `<div>Hello <at id="0">Alice</at></div>`,
		ContentType: models.MessageContentTypeHTML,
		Mentions: []models.Mention{
			{Kind: models.MentionUser, AtID: 0, Text: "Alice", TargetID: userID},
		},
	}

	got, err := PrepareMentions(body)
	require.NoError(t, err)
	require.Len(t, got, 1)

	m0 := got[0]
	require.NotNil(t, m0.GetId())
	require.Equal(t, int32(0), *m0.GetId())
	require.NotNil(t, m0.GetMentionText())
	require.Equal(t, "Alice", *m0.GetMentionText())

	mentioned := m0.GetMentioned()
	require.NotNil(t, mentioned)
	require.NotNil(t, mentioned.GetUser())

	u := mentioned.GetUser()
	require.NotNil(t, u.GetId())
	require.Equal(t, userID, *u.GetId())

	ad := u.GetAdditionalData()
	require.NotNil(t, ad)
	require.Equal(t, "aadUser", ad["userIdentityType"])
}

func TestPrepareMentions_ChannelMention_Success(t *testing.T) {
	channelThreadID := "19:abc123def456@thread.tacv2"

	body := &models.MessageBody{
		Content:     `<div>Hello <at id="0">General</at></div>`,
		ContentType: models.MessageContentTypeHTML,
		Mentions: []models.Mention{
			{Kind: models.MentionChannel, AtID: 0, Text: "General", TargetID: channelThreadID},
		},
	}

	got, err := PrepareMentions(body)
	require.NoError(t, err)
	require.Len(t, got, 1)

	m0 := got[0]
	mentioned := m0.GetMentioned()
	require.NotNil(t, mentioned)
	require.NotNil(t, mentioned.GetConversation())

	conv := mentioned.GetConversation()
	require.NotNil(t, conv.GetId())
	require.Equal(t, channelThreadID, *conv.GetId())
	require.NotNil(t, conv.GetConversationIdentityType())
	require.Equal(t, msmodels.CHANNEL_TEAMWORKCONVERSATIONIDENTITYTYPE, *conv.GetConversationIdentityType())
}

func TestPrepareMentions_TeamMention_Success(t *testing.T) {
	teamID := "00000000-0000-0000-0000-0000000000aa"

	body := &models.MessageBody{
		Content:     `<div>Hello <at id="0">My Team</at></div>`,
		ContentType: models.MessageContentTypeHTML,
		Mentions: []models.Mention{
			{Kind: models.MentionTeam, AtID: 0, Text: "My Team", TargetID: teamID},
		},
	}

	got, err := PrepareMentions(body)
	require.NoError(t, err)
	require.Len(t, got, 1)

	m0 := got[0]
	require.NotNil(t, m0.GetMentioned())
	require.NotNil(t, m0.GetMentioned().GetConversation())

	conv := m0.GetMentioned().GetConversation()
	require.NotNil(t, conv.GetId())
	require.Equal(t, teamID, *conv.GetId())
	require.NotNil(t, conv.GetConversationIdentityType())
	require.Equal(t, msmodels.TEAM_TEAMWORKCONVERSATIONIDENTITYTYPE, *conv.GetConversationIdentityType())
}

func TestPrepareMentions_EveryoneMention_Success(t *testing.T) {
	chatThreadID := "19:abcdef123456@thread.tacv2"

	body := &models.MessageBody{
		Content:     `<div>Hello <at id="0">Everyone</at></div>`,
		ContentType: models.MessageContentTypeHTML,
		Mentions: []models.Mention{
			{Kind: models.MentionEveryone, AtID: 0, Text: "Everyone", TargetID: chatThreadID},
		},
	}

	got, err := PrepareMentions(body)
	require.NoError(t, err)
	require.Len(t, got, 1)

	m0 := got[0]
	mentioned := m0.GetMentioned()
	require.NotNil(t, mentioned)
	require.NotNil(t, mentioned.GetConversation())

	conv := mentioned.GetConversation()
	require.NotNil(t, conv.GetId())
	require.Equal(t, chatThreadID, *conv.GetId())
	require.NotNil(t, conv.GetConversationIdentityType())
	require.Equal(t, msmodels.CHAT_TEAMWORKCONVERSATIONIDENTITYTYPE, *conv.GetConversationIdentityType())
}

func TestPrepareMentions_AtTagCheckIsCaseSensitiveOnIdAttribute(t *testing.T) {
	body := &models.MessageBody{
		Content:     `<div><at ID="0">Alice</at></div>`,
		ContentType: models.MessageContentTypeHTML,
		Mentions: []models.Mention{
			{Kind: models.MentionUser, AtID: 0, Text: "Alice", TargetID: "00000000-0000-0000-0000-000000000001"},
		},
	}

	_, err := PrepareMentions(body)
	require.Error(t, err)
	require.Contains(t, err.Error(), `missing <at id="0"`)
}

func TestPrepareMentions_AtTagPresentButMentionsIsEmptySlice_ReturnsError(t *testing.T) {
	body := &models.MessageBody{
		Content:     `<div>Hello <at id="0">X</at></div>`,
		ContentType: models.MessageContentTypeHTML,
		Mentions:    []models.Mention{},
	}

	_, err := PrepareMentions(body)
	require.Error(t, err)
	require.Contains(t, err.Error(), "content contains <at> tags but mentions list is empty")
}

func TestPrepareMentions_NoAtTagsButMentionsPresent_ReturnsError(t *testing.T) {
	body := &models.MessageBody{
		Content:     `<div>Hello</div>`,
		ContentType: models.MessageContentTypeHTML,
		Mentions: []models.Mention{
			{Kind: models.MentionTeam, AtID: 0, Text: "My Team", TargetID: "00000000-0000-0000-0000-0000000000aa"},
		},
	}

	_, err := PrepareMentions(body)
	require.Error(t, err)
	require.Contains(t, err.Error(), `missing <at id="0"`)
}

func TestPrepareMentions_UnsupportedMentionKind_ReturnsError(t *testing.T) {
	body := &models.MessageBody{
		Content:     `<div>Hello <at id="0">X</at></div>`,
		ContentType: models.MessageContentTypeHTML,
		Mentions: []models.Mention{
			{Kind: models.MentionKind("weird"), AtID: 0, Text: "X", TargetID: "whatever"},
		},
	}

	_, err := PrepareMentions(body)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported MentionKind")
}
