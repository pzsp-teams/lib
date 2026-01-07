package api

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

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

var (
	reTeamIDQuoted = regexp.MustCompile(`/teams\('([^']+)'\)`)
	reTeamIDSlash  = regexp.MustCompile(`/teams/([^/?]+)`)
)

func normalizeVisibilityForGroup(v string) string {
	ve := strings.TrimSpace(strings.ToLower(v))
	switch ve {
	case "private":
		return "Private"
	case "public":
		return "Public"
	default:
		if ve == "" {
			return "Public"
		}
		return v
	}
}

func parseTeamIDFromHeaders(contentLocation, location string) (string, bool) {
	for _, h := range []string{contentLocation, location} {
		h = strings.TrimSpace(h)
		if h == "" {
			continue
		}
		if m := reTeamIDQuoted.FindStringSubmatch(h); len(m) == 2 {
			return m[1], true
		}
		if m := reTeamIDSlash.FindStringSubmatch(h); len(m) == 2 {
			return m[1], true
		}
		if idx := strings.Index(h, "/operations"); idx > 0 {
			part := h[:idx]
			if m := reTeamIDQuoted.FindStringSubmatch(part); len(m) == 2 {
				return m[1], true
			}
			if m := reTeamIDSlash.FindStringSubmatch(part); len(m) == 2 {
				return m[1], true
			}
		}
	}
	return "", false
}

func (t *teamAPI) waitTeamReady(ctx context.Context, teamID string, timeout time.Duration) *sender.RequestError {
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	tryOnce := func() bool {
		call := func(ctx context.Context) (sender.Response, error) {
			return t.client.Teams().ByTeamId(teamID).Get(ctx, nil)
		}
		_, err := sender.SendRequest(ctx, call, t.senderCfg)
		return err == nil
	}

	if tryOnce() {
		return nil
	}

	for {
		select {
		case <-ctx.Done():
			return &sender.RequestError{
				Code:    http.StatusRequestTimeout,
				Message: ctx.Err().Error(),
			}
		case <-deadline.C:
			return &sender.RequestError{
				Code:    http.StatusRequestTimeout,
				Message: "Team not ready within timeout",
			}
		case <-ticker.C:
			if tryOnce() {
				return nil
			}
		}
	}
}

func (t *teamAPI) addOwnerWithRetry(ctx context.Context, teamID, ownerID string) *sender.RequestError {
	const (
		totalTimeout = 90 * time.Second
		maxBackoff   = 8 * time.Second
	)
	deadline := time.Now().Add(totalTimeout)
	backoff := 500 * time.Millisecond

	for {
		_, err := t.AddMember(ctx, teamID, ownerID, []string{roleOwner})
		if err == nil {
			return nil
		}
		retryable := err.Code == http.StatusNotFound ||
			err.Code == http.StatusTooManyRequests ||
			err.Code == http.StatusServiceUnavailable

		if !retryable || time.Now().After(deadline) {
			return err
		}

		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			return &sender.RequestError{Code: http.StatusRequestTimeout, Message: ctx.Err().Error()}
		}

		if backoff < maxBackoff {
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		}
	}
}
