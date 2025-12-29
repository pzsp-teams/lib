package api

import (
	"context"
	"fmt"
	"time"

	graph "github.com/microsoftgraph/msgraph-sdk-go"
	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	graphusers "github.com/microsoftgraph/msgraph-sdk-go/users"
	"github.com/pzsp-teams/lib/internal/sender"
)

type OneOnOneChatAPI interface {
	CreateOneOnOneChat(ctx context.Context, recipientRef string) (msmodels.Chatable, *sender.RequestError)
}

type GroupChatAPI interface {
	CreateGroupChat(ctx context.Context, recipientRefs []string, topic string, includeMe bool) (msmodels.Chatable, *sender.RequestError)
	AddMemberToGroupChat(ctx context.Context, chatID, userRef string) (msmodels.ConversationMemberable, *sender.RequestError)
	RemoveMemberFromGroupChat(ctx context.Context, chatID, userRef string) *sender.RequestError
	ListGroupChatMembers(ctx context.Context, chatID string) (msmodels.ConversationMemberCollectionResponseable, *sender.RequestError)
	UpdateGroupChatTopic(ctx context.Context, chatID, topic string) (msmodels.Chatable, *sender.RequestError)
}

type ChatAPI interface {
	OneOnOneChatAPI
	GroupChatAPI
	ListMessages(ctx context.Context, chatID string) (msmodels.ChatMessageCollectionResponseable, *sender.RequestError)
	ListChats(ctx context.Context, chatType string) (msmodels.ChatCollectionResponseable, *sender.RequestError)
	SendMessage(ctx context.Context, chatID, content, contentType string) (msmodels.ChatMessageable, *sender.RequestError)
	DeleteMessage(ctx context.Context, chatID, messageID string) *sender.RequestError
	GetMessage(ctx context.Context, chatID, messageID string) (msmodels.ChatMessageable, *sender.RequestError)
	ListAllMessages(ctx context.Context, startTime, endTime *time.Time, top *int32) (msmodels.ChatMessageCollectionResponseable, *sender.RequestError)
	ListPinnedMessages(ctx context.Context, chatID string) (msmodels.ChatMessageCollectionResponseable, *sender.RequestError)
	PinMessage(ctx context.Context, chatID, messageID string) *sender.RequestError
	UnpinMessage(ctx context.Context, chatID, pinnedID string) *sender.RequestError
}

type chatsAPI struct {
	client     *graph.GraphServiceClient
	techParams sender.RequestTechParams
}

func NewChat(client *graph.GraphServiceClient, techParams sender.RequestTechParams) ChatAPI {
	return &chatsAPI{client, techParams}
}

func (c *chatsAPI) CreateOneOnOneChat(ctx context.Context, userRef string) (msmodels.Chatable, *sender.RequestError) {
	body := msmodels.NewChat()
	chatType := msmodels.ONEONONE_CHATTYPE
	body.SetChatType(&chatType)

	me, err := getMe(ctx, c.client, c.techParams)
	if err != nil {
		return nil, err
	}

	userRefs := []string{*me.GetId(), userRef}
	members := make([]msmodels.ConversationMemberable, 0, len(userRefs))
	addToMembers(&members, userRefs, "owner")
	body.SetMembers(members)

	call := func(ctx context.Context) (sender.Response, error) {
		return c.client.Chats().Post(ctx, body, nil)
	}

	resp, err := sender.SendRequest(ctx, call, c.techParams)
	if err != nil {
		return nil, err
	}

	out, ok := resp.(msmodels.Chatable)
	if !ok {
		return nil, newTypeError("Chatable")
	}
	return out, nil
}

func (c *chatsAPI) ListChats(ctx context.Context, chatType string) (msmodels.ChatCollectionResponseable, *sender.RequestError) {
	filter := fmt.Sprintf("chatType eq '%s'", chatType)
	requestParameters := &graphusers.ItemChatsRequestBuilderGetQueryParameters{
		Filter:  &filter,
		Expand:  []string{"members"},
		Orderby: []string{"lastMessagePreview/createdDateTime desc"},
	}

	configuration := &graphusers.ItemChatsRequestBuilderGetRequestConfiguration{
		QueryParameters: requestParameters,
	}

	call := func(ctx context.Context) (sender.Response, error) {
		return c.client.
			Me().
			Chats().
			Get(ctx, configuration)
	}

	resp, err := sender.SendRequest(ctx, call, c.techParams)
	if err != nil {
		return nil, err
	}

	out, ok := resp.(msmodels.ChatCollectionResponseable)
	if !ok {
		return nil, newTypeError("ChatCollectionResponseable")
	}

	return out, nil
}

func (c *chatsAPI) CreateGroupChat(ctx context.Context, userRefs []string, topic string, includeMe bool) (msmodels.Chatable, *sender.RequestError) {
	body := msmodels.NewChat()
	chatType := msmodels.GROUP_CHATTYPE
	body.SetChatType(&chatType)
	body.SetTopic(&topic)

	if includeMe {
		me, err := getMe(ctx, c.client, c.techParams)
		if err != nil {
			return nil, err
		}
		userRefs = append(userRefs, *me.GetId())
	}

	members := make([]msmodels.ConversationMemberable, 0, len(userRefs))
	addToMembers(&members, userRefs, "owner")
	body.SetMembers(members)

	call := func(ctx context.Context) (sender.Response, error) {
		return c.client.Chats().Post(ctx, body, nil)
	}

	resp, err := sender.SendRequest(ctx, call, c.techParams)
	if err != nil {
		return nil, err
	}

	out, ok := resp.(msmodels.Chatable)
	if !ok {
		return nil, newTypeError("Chatable")
	}
	return out, nil
}

// ListGroupChats removed; use ListChats with "group" as chatType instead.

func (c *chatsAPI) AddMemberToGroupChat(ctx context.Context, chatID, userRef string) (msmodels.ConversationMemberable, *sender.RequestError) {
	chatMember := msmodels.NewAadUserConversationMember()
	chatMember.SetRoles([]string{"owner"})
	chatMember.SetAdditionalData(map[string]any{
		graphUserBindKey: fmt.Sprintf(graphUserBindFmt, userRef),
	})

	call := func(ctx context.Context) (sender.Response, error) {
		return c.client.Chats().ByChatId(chatID).Members().Post(ctx, chatMember, nil)
	}

	resp, err := sender.SendRequest(ctx, call, c.techParams)
	if err != nil {
		return nil, err
	}

	out, ok := resp.(msmodels.ConversationMemberable)
	if !ok {
		return nil, newTypeError("ConversationMemberable")
	}

	return out, nil
}

func (c *chatsAPI) RemoveMemberFromGroupChat(ctx context.Context, chatID, userRef string) *sender.RequestError {
	call := func(ctx context.Context) (sender.Response, error) {
		return nil, c.client.Chats().
			ByChatId(chatID).
			Members().
			ByConversationMemberId(userRef).
			Delete(ctx, nil)
	}

	_, err := sender.SendRequest(ctx, call, c.techParams)
	return err
}

func (c *chatsAPI) ListGroupChatMembers(ctx context.Context, chatID string) (msmodels.ConversationMemberCollectionResponseable, *sender.RequestError) {
	call := func(ctx context.Context) (sender.Response, error) {
		return c.client.Chats().ByChatId(chatID).Members().Get(ctx, nil)
	}

	resp, err := sender.SendRequest(ctx, call, c.techParams)
	if err != nil {
		return nil, err
	}

	out, ok := resp.(msmodels.ConversationMemberCollectionResponseable)
	if !ok {
		return nil, newTypeError("ConversationMemberCollectionResponseable")
	}
	return out, nil
}

func (c *chatsAPI) UpdateGroupChatTopic(ctx context.Context, chatID, topic string) (msmodels.Chatable, *sender.RequestError) {
	body := msmodels.NewChat()
	body.SetTopic(&topic)

	call := func(ctx context.Context) (sender.Response, error) {
		return c.client.Chats().ByChatId(chatID).Patch(ctx, body, nil)
	}

	resp, err := sender.SendRequest(ctx, call, c.techParams)
	if err != nil {
		return nil, err
	}

	out, ok := resp.(msmodels.Chatable)
	if !ok {
		return nil, newTypeError("Chatable")
	}
	return out, nil
}

func (c *chatsAPI) ListMessages(ctx context.Context, chatID string) (msmodels.ChatMessageCollectionResponseable, *sender.RequestError) {
	call := func(ctx context.Context) (sender.Response, error) {
		return c.client.Chats().ByChatId(chatID).Messages().Get(ctx, nil)
	}

	resp, err := sender.SendRequest(ctx, call, c.techParams)
	if err != nil {
		return nil, err
	}

	out, ok := resp.(msmodels.ChatMessageCollectionResponseable)
	if !ok {
		return nil, newTypeError("ChatCollectionResponseable")
	}

	return out, nil
}

func (c *chatsAPI) SendMessage(ctx context.Context, chatID, content, contentType string) (msmodels.ChatMessageable, *sender.RequestError) {
	requestBody := msmodels.NewChatMessage()
	body := msmodels.NewItemBody()
	bodyType := msmodels.TEXT_BODYTYPE
	body.SetContentType(&bodyType)
	body.SetContent(&content)
	requestBody.SetBody(body)

	call := func(ctx context.Context) (sender.Response, error) {
		return c.client.Chats().ByChatId(chatID).Messages().Post(ctx, requestBody, nil)
	}

	resp, err := sender.SendRequest(ctx, call, c.techParams)
	if err != nil {
		return nil, err
	}

	out, ok := resp.(msmodels.ChatMessageable)
	if !ok {
		return nil, newTypeError("ChatMessageable")
	}
	return out, nil
}

func (c *chatsAPI) DeleteMessage(ctx context.Context, chatID, messageID string) *sender.RequestError {
	call := func(ctx context.Context) (sender.Response, error) {
		return nil, c.client.
			Me().
			Chats().
			ByChatId(chatID).
			Messages().
			ByChatMessageId(messageID).
			SoftDelete().
			Post(ctx, nil)
	}

	_, err := sender.SendRequest(ctx, call, c.techParams)
	return err
}

func (c *chatsAPI) GetMessage(ctx context.Context, chatID, messageID string) (msmodels.ChatMessageable, *sender.RequestError) {
	call := func(ctx context.Context) (sender.Response, error) {
		return c.client.
			Chats().
			ByChatId(chatID).
			Messages().
			ByChatMessageId(messageID).
			Get(ctx, nil)
	}

	resp, err := sender.SendRequest(ctx, call, c.techParams)
	if err != nil {
		return nil, err
	}

	out, ok := resp.(msmodels.ChatMessageable)
	if !ok {
		return nil, newTypeError("ChatMessageable")
	}

	return out, nil
}

func (c *chatsAPI) ListAllMessages(ctx context.Context, startTime, endTime *time.Time, top *int32) (msmodels.ChatMessageCollectionResponseable, *sender.RequestError) {
	requestParameters := &graphusers.ItemChatsGetAllMessagesRequestBuilderGetQueryParameters{
		Top: top,
	}

	if startTime != nil && endTime != nil {
		filter := fmt.Sprintf(
			"lastModifiedDateTime gt %s and lastModifiedDateTime lt %s",
			startTime.UTC().Format(time.RFC3339),
			endTime.UTC().Format(time.RFC3339),
		)
		requestParameters.Filter = &filter
	}

	configuration := &graphusers.ItemChatsGetAllMessagesRequestBuilderGetRequestConfiguration{
		QueryParameters: requestParameters,
	}

	call := func(ctx context.Context) (sender.Response, error) {
		return c.client.Me().Chats().GetAllMessages().GetAsGetAllMessagesGetResponse(ctx, configuration)
	}

	resp, err := sender.SendRequest(ctx, call, c.techParams)
	if err != nil {
		return nil, err
	}

	out, ok := resp.(msmodels.ChatMessageCollectionResponseable)
	if !ok {
		return nil, newTypeError("ChatMessageCollectionResponseable")
	}

	return out, nil
}

func (c *chatsAPI) ListPinnedMessages(ctx context.Context, chatID string) (msmodels.ChatMessageCollectionResponseable, *sender.RequestError) {
	call := func(ctx context.Context) (sender.Response, error) {
		return c.client.Chats().ByChatId(chatID).PinnedMessages().Get(ctx, nil)
	}

	resp, err := sender.SendRequest(ctx, call, c.techParams)
	if err != nil {
		return nil, err
	}

	out, ok := resp.(msmodels.ChatMessageCollectionResponseable)
	if !ok {
		return nil, newTypeError("ChatMessageCollectionResponseable")
	}
	return out, nil
}

func (c *chatsAPI) PinMessage(ctx context.Context, chatID, messageID string) *sender.RequestError {
	pinned := msmodels.NewPinnedChatMessageInfo()
	body := map[string]any{
		graphMessageBindKey: fmt.Sprintf(graphMessageBindFmt, chatID, messageID),
	}
	pinned.SetAdditionalData(body)

	call := func(ctx context.Context) (sender.Response, error) {
		return c.client.
			Chats().
			ByChatId(chatID).
			PinnedMessages().
			Post(ctx, pinned, nil)
	}

	_, err := sender.SendRequest(ctx, call, c.techParams)
	if err != nil {
		return err
	}

	return nil
}

func (c *chatsAPI) UnpinMessage(ctx context.Context, chatID, pinnedID string) *sender.RequestError {
	call := func(ctx context.Context) (sender.Response, error) {
		return nil, c.client.
			Chats().
			ByChatId(chatID).
			PinnedMessages().
			ByPinnedChatMessageInfoId(pinnedID).
			Delete(ctx, nil)
	}

	_, err := sender.SendRequest(ctx, call, c.techParams)
	return err
}
