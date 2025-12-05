package api

import (
	"context"
	"fmt"

	graph "github.com/microsoftgraph/msgraph-sdk-go"
	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/internal/sender"
)

type ChatsAPI interface {
	Create(ctx context.Context, emails []string, topic string) (msmodels.Chatable, *sender.RequestError)
	//	ListMyJoined(ctx context.Context) (msmodels.ChatCollectionResponseable, *sender.RequestError)
	// Get()
	// Delete()
	// GetChatMember()
	// ListChatMembers()
	// AddMember()
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
	for i, name := range emails {
		chatMember := msmodels.NewAadUserConversationMember()
		roles := []string{"owner"}
		chatMember.SetRoles(roles)
		data := map[string]any{
			"user@odata.bind": fmt.Sprintf("https://graph.microsoft.com/v1.0/users('%s')", name),
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
		return nil, &sender.RequestError{Code: "TypeCastError", Message: "Expected Chatable"}
	}
	return out, nil
}

func (c *chatsAPI) ListMyJoined(ctx context.Context) (msmodels.ChatCollectionResponseable, *sender.RequestError) {
	call := func(ctx context.Context) (sender.Response, error) {
		return c.client.Me().Chats().Get(ctx, nil)
	}
	resp, err := sender.SendRequest(ctx, call, c.techParams)
	if err != nil {
		return nil, err
	}
	out, ok := resp.(msmodels.ChatCollectionResponseable)
	if !ok {
		return nil, &sender.RequestError{Code: "TypeCastError", Message: "Expected ChatCollectionResponseable"}
	}
	return out, nil
}
