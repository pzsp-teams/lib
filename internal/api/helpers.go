package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	abstractions "github.com/microsoft/kiota-abstractions-go"
	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	graphteams "github.com/microsoftgraph/msgraph-sdk-go/teams"
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

func buildTeamFromTemplateBody(displayName, description, visibility, primaryOwner string) msmodels.Teamable {
	body := msmodels.NewTeam()
	body.SetDisplayName(&displayName)

	if description != "" {
		body.SetDescription(&description)
	}

	teamVisibility := msmodels.PUBLIC_TEAMVISIBILITYTYPE
	if strings.EqualFold(strings.TrimSpace(visibility), "private") {
		teamVisibility = msmodels.PRIVATE_TEAMVISIBILITYTYPE
	}
	body.SetVisibility(&teamVisibility)

	first := "General"
	body.SetFirstChannelName(&first)
	body.SetAdditionalData(map[string]any{
		templateBindKey: templateBindValue,
	})

	var convMembers []msmodels.ConversationMemberable
	addToMembers(&convMembers, []string{primaryOwner}, []string{roleOwner})
	body.SetMembers(convMembers)

	return body
}

func (t *teamAPI) normalizeOwners(ctx context.Context, owners []string, includeMe bool) ([]string, *sender.RequestError) {
	owners = filterTrimNonEmpty(owners)

	if includeMe {
		me, err := getMe(ctx, t.client, t.senderCfg)
		if err != nil {
			return nil, err
		}
		if me.GetId() != nil && strings.TrimSpace(*me.GetId()) != "" {
			owners = append(owners, *me.GetId())
		}
	}

	if len(owners) == 0 {
		return nil, &sender.RequestError{Code: http.StatusBadRequest, Message: "at least one owner is required"}
	}

	return owners, nil
}

func validateCreateFromTemplate(displayName string) *sender.RequestError {
	if displayName == "" {
		return &sender.RequestError{Code: http.StatusBadRequest, Message: "displayName cannot be empty"}
	}
	return nil
}

func (t *teamAPI) postTeamAndExtractID(ctx context.Context, body msmodels.Teamable) (string, *sender.RequestError) {
	var loc, contentLoc string

	responseHandler := func(resp any, _ abstractions.ErrorMappings) (any, error) {
		httpResp, ok := resp.(*http.Response)
		if !ok || httpResp == nil {
			return nil, nil
		}
		if httpResp.StatusCode >= 400 {
			msg := httpResp.Status
			if httpResp.Body != nil {
				b, _ := io.ReadAll(httpResp.Body)
				_ = httpResp.Body.Close()
				if s := strings.TrimSpace(string(b)); s != "" {
					msg = s
				}
			}
			return nil, &sender.RequestError{
				Code:    httpResp.StatusCode,
				Message: msg,
			}
		}
		loc = httpResp.Header.Get("Location")
		contentLoc = httpResp.Header.Get("Content-Location")
		return nil, nil
	}

	handlerOpt := abstractions.NewRequestHandlerOption()
	handlerOpt.SetResponseHandler(responseHandler)

	reqCfg := &graphteams.TeamsRequestBuilderPostRequestConfiguration{
		Options: []abstractions.RequestOption{handlerOpt},
	}

	call := func(_ context.Context) (sender.Response, error) {
		return t.client.Teams().Post(ctx, body, reqCfg)
	}
	if _, err := sender.SendRequest(ctx, call, t.senderCfg); err != nil {
		return "", err
	}

	teamID, ok := parseTeamIDFromHeaders(contentLoc, loc)
	if !ok || strings.TrimSpace(teamID) == "" {
		return "", &sender.RequestError{
			Code:    http.StatusInternalServerError,
			Message: "unable to parse team ID from response headers",
		}
	}

	return teamID, nil
}

func (t *teamAPI) addPostCreateMembersAndOwners(ctx context.Context, teamID string, members, remainingOwners []string) *sender.RequestError {
	members = filterTrimNonEmpty(members)
	remainingOwners = filterTrimNonEmpty(remainingOwners)

	if err := t.addMembersInBulk(ctx, teamID, members); err != nil {
		return err
	}

	for _, ownerID := range remainingOwners {
		if err := t.addOwnerWithRetry(ctx, teamID, ownerID); err != nil {
			return err
		}
	}

	return nil
}

func filterTrimNonEmpty(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		out = append(out, s)
	}
	return out
}

func (t *teamAPI) addMembersInBulk(ctx context.Context, teamID string, members []string) *sender.RequestError {
	if len(members) == 0 {
		return nil
	}
	requestBody := graphteams.NewItemMembersAddPostRequestBody()
	var membersToAdd []msmodels.ConversationMemberable
	addToMembers(&membersToAdd, members, []string{})
	requestBody.SetValues(membersToAdd)
	addMembersCall := func(ctx context.Context) (sender.Response, error) {
		return t.client.
			Teams().
			ByTeamId(teamID).
			Members().
			Add().
			PostAsAddPostResponse(ctx, requestBody, nil)
	}
	if _, err := sender.SendRequest(ctx, addMembersCall, t.senderCfg); err != nil {
		return err
	}
	return nil
}
