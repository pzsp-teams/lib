package api

import (
	"context"
	"fmt"

	graph "github.com/microsoftgraph/msgraph-sdk-go"
	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/internal/sender"
)

type ChatsAPI interface {
	CreateChat(ctx context.Context, guestNames []string, guestRole string) (msmodels.Chatable, *sender.RequestError)
	// CreateGroupChat(ctx context.Context)
	// GetChat()
	// DeleteChat()
	// GetMember()
	// ListMembers()
	// AddMember()
	// RemoveMember()
	// ListMessagesInChat()
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

func (c *chatsAPI) CreateChat(ctx context.Context, guestNames []string, guestRole string) (msmodels.Chatable, *sender.RequestError) {
	chatType := msmodels.GROUP_CHATTYPE
	if len(guestNames) == 1 {
		chatType = msmodels.ONEONONE_CHATTYPE
	}
	body := msmodels.NewChat()
	body.SetChatType(&chatType)

	me, _ := c.client.Me().Get(ctx, nil)

	ids := []string{*me.GetId()}
	ids = append(ids, guestNames...)
	members := make([]msmodels.ConversationMemberable, len(ids))

	for i, id := range ids {
		chatMember := msmodels.NewAadUserConversationMember()
		roles := []string{guestRole}
		if i == 0 {
			roles = []string{"owner"}
		}
		chatMember.SetRoles(roles)
		data := map[string]any{
			"user@odata.bind": fmt.Sprintf("https://graph.microsoft.com/v1.0/users('%s')", id),
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
