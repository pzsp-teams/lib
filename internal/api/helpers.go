package api

import (
	"context"
	"fmt"
	"net/http"

	graph "github.com/microsoftgraph/msgraph-sdk-go"
	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/config"
	"github.com/pzsp-teams/lib/internal/sender"
)

const (
	graphUserBindFmt    = "https://graph.microsoft.com/v1.0/users('%s')"
	graphUserBindKey    = "user@odata.bind"
	templateBindValue   = "https://graph.microsoft.com/v1.0/teamsTemplates('standard')"
	templateBindKey     = "template@odata.bind"
	graphMessageBindFmt = "https://graph.microsoft.com/v1.0/chats/%s/messages/%s"
	graphMessageBindKey = "message@odata.bind"
)

func addToMembers(members *[]msmodels.ConversationMemberable, userRefs []string, role string) {
	for _, userRef := range userRefs {
		member := msmodels.NewAadUserConversationMember()
		member.SetRoles([]string{role})
		member.SetAdditionalData(map[string]any{
			graphUserBindKey: fmt.Sprintf(graphUserBindFmt, userRef),
		})
		*members = append(*members, member)
	}
}

func messageToGraph(content, contentType string) msmodels.ItemBodyable {
	body := msmodels.NewItemBody()
	body.SetContent(&content)
	ct := msmodels.TEXT_BODYTYPE
	if contentType == "html" {
		ct = msmodels.HTML_BODYTYPE
	}
	body.SetContentType(&ct)
	return body
}

func getMe(ctx context.Context, client *graph.GraphServiceClient, senderCfg *config.SenderConfig) (msmodels.Userable, *sender.RequestError) {
	call := func(ctx context.Context) (sender.Response, error) {
		return client.Me().Get(ctx, nil)
	}

	resp, err := sender.SendRequest(ctx, call, senderCfg)
	if err != nil {
		return nil, err
	}

	user, ok := resp.(msmodels.Userable)
	if !ok {
		return nil, newTypeError("msmodels.Userable")
	}
	return user, nil
}

func newTypeError(expected string) *sender.RequestError {
	return &sender.RequestError{
		Code:    http.StatusUnprocessableEntity,
		Message: "Expected " + expected,
	}
}
