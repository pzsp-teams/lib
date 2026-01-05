package api

import (
	"fmt"
	"net/http"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/internal/sender"
)

const (
	graphUserBindFmt    = "https://graph.microsoft.com/v1.0/users('%s')"
	graphUserBindKey    = "user@odata.bind"
	templateBindValue   = "https://graph.microsoft.com/v1.0/teamsTemplates('standard')"
	templateBindKey     = "template@odata.bind"
	graphMessageBindFmt = "https://graph.microsoft.com/v1.0/chats/%s/messages/%s"
	graphMessageBindKey = "message@odata.bind"
	roleOwner           = "owner"
)

func newAadUserMemberBody(userRef string, roles []string) msmodels.ConversationMemberable {
	m := msmodels.NewAadUserConversationMember()
	m.SetRoles(roles)
	m.SetAdditionalData(map[string]any{
		graphUserBindKey: fmt.Sprintf(graphUserBindFmt, userRef),
	})
	return m
}

func newRolesPatchBody(roles []string) msmodels.ConversationMemberable {
	patch := msmodels.NewAadUserConversationMember()
	patch.SetRoles(roles)
	return patch
}

func addToMembers(members *[]msmodels.ConversationMemberable, userRefs, roles []string) {
	for _, userRef := range userRefs {
		*members = append(*members, newAadUserMemberBody(userRef, roles))
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

func newTypeError(expected string) *sender.RequestError {
	return &sender.RequestError{
		Code:    http.StatusUnprocessableEntity,
		Message: "Expected " + expected,
	}
}

func isSystemEvent(m msmodels.ChatMessageable) bool {
	if m.GetEventDetail() != nil {
		return true
	}
	if mt := m.GetMessageType(); mt != nil && *mt == msmodels.CHATEVENT_CHATMESSAGETYPE {
		return true
	}
	return false
}

func filterOutSystemEvents(messages msmodels.ChatMessageCollectionResponseable) []msmodels.ChatMessageable {
	vals := messages.GetValue()
	if vals == nil {
		return nil
	}
	filtered := make([]msmodels.ChatMessageable, 0, len(vals))
	for _, v := range vals {
		if v == nil || isSystemEvent(v) {
			continue
		}
		filtered = append(filtered, v)
	}
	return filtered
}
