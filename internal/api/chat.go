package api

import (
	"context"
	"fmt"

	graph "github.com/microsoftgraph/msgraph-sdk-go"
	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	graphusers "github.com/microsoftgraph/msgraph-sdk-go/users"
	"github.com/pzsp-teams/lib/internal/sender"
)

type ChatsAPI interface {
	Create(ctx context.Context, emails []string, topic string) (msmodels.Chatable, *sender.RequestError)
	SendMessage(ctx context.Context, chatID, content string) (msmodels.ChatMessageable, *sender.RequestError)
	ListMyJoined(ctx context.Context) (msmodels.ChatCollectionResponseable, *sender.RequestError)
	ListMembers(ctx context.Context, chatID string) (msmodels.ConversationMemberCollectionResponseable, *sender.RequestError)
	AddMember(ctx context.Context, chatID, email string) (msmodels.ConversationMemberable, *sender.RequestError)
	// RemoveMember()
	// ListMessages()
	// ListPinnedMessages()
	// PinMessage()
	// UnpinMessage()
}

type chatsAPI struct {
	client     *graph.GraphServiceClient
	techParams sender.RequestTechParams
}

func NewChat(client *graph.GraphServiceClient, techParams sender.RequestTechParams) ChatsAPI {
	return &chatsAPI{client, techParams}
}

func (c *chatsAPI) Create(ctx context.Context, emails []string, topic string) (msmodels.Chatable, *sender.RequestError) {
	body := msmodels.NewChat()
	chatType := msmodels.ONEONONE_CHATTYPE
	if len(emails) > 2 {
		chatType = msmodels.GROUP_CHATTYPE
		body.SetTopic(&topic)
	}
	body.SetChatType(&chatType)

	members := make([]msmodels.ConversationMemberable, len(emails))
	for i, email := range emails {
		chatMember := msmodels.NewAadUserConversationMember()
		roles := []string{"owner"}
		chatMember.SetRoles(roles)
		data := map[string]any{
			graphUserBindKey: fmt.Sprintf(graphUserBindFmt, email),
		}

		chatMember.SetAdditionalData(data)
		members[i] = chatMember
	}
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

func (c *chatsAPI) SendMessage(ctx context.Context, chatID, content string) (msmodels.ChatMessageable, *sender.RequestError) {
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

func (c *chatsAPI) ListMyJoined(ctx context.Context) (msmodels.ChatCollectionResponseable, *sender.RequestError) {
	requestParameters := &graphusers.ItemChatsRequestBuilderGetQueryParameters{
		Expand:  []string{"members"},
		Orderby: []string{"lastMessagePreview/createdDateTime desc"},
	}
	configuration := &graphusers.ItemChatsRequestBuilderGetRequestConfiguration{
		QueryParameters: requestParameters,
	}
	call := func(ctx context.Context) (sender.Response, error) {
		return c.client.Me().Chats().Get(ctx, configuration)
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

func (c *chatsAPI) ListMembers(ctx context.Context, chatID string) (msmodels.ConversationMemberCollectionResponseable, *sender.RequestError) {
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

func (c *chatsAPI) AddMember(ctx context.Context, chatID, email string) (msmodels.ConversationMemberable, *sender.RequestError) {
	chatMember := msmodels.NewAadUserConversationMember()
	chatMember.SetRoles([]string{"owner"})
	data := map[string]any{
		graphUserBindKey: fmt.Sprintf(graphUserBindFmt, email),
	}
	chatMember.SetAdditionalData(data)

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
