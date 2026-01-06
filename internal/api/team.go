package api

import (
	"context"
	"net/http"
	"strings"
	"time"

	abstractions "github.com/microsoft/kiota-abstractions-go"
	graph "github.com/microsoftgraph/msgraph-sdk-go"
	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	graphteams "github.com/microsoftgraph/msgraph-sdk-go/teams"

	"github.com/pzsp-teams/lib/config"
	"github.com/pzsp-teams/lib/internal/sender"
	"github.com/pzsp-teams/lib/models"
)

type TeamAPI interface {
	CreateFromTemplate(ctx context.Context, displayName, description string, owners, members []string, visibility string) (string, *sender.RequestError)
	CreateViaGroup(ctx context.Context, displayName, mailNickname, visibility string) (string, *sender.RequestError)
	Get(ctx context.Context, teamID string) (msmodels.Teamable, *sender.RequestError)
	ListMyJoined(ctx context.Context) (msmodels.TeamCollectionResponseable, *sender.RequestError)
	Archive(ctx context.Context, teamID string, spoReadOnlyForMembers *bool) *sender.RequestError
	Unarchive(ctx context.Context, teamID string) *sender.RequestError
	Delete(ctx context.Context, teamID string) *sender.RequestError
	RestoreDeleted(ctx context.Context, deletedGroupID string) (msmodels.DirectoryObjectable, *sender.RequestError)
	UpdateTeam(ctx context.Context, teamID string, update *models.TeamUpdate) (msmodels.Teamable, *sender.RequestError)
	ListMembers(ctx context.Context, teamID string) (msmodels.ConversationMemberCollectionResponseable, *sender.RequestError)
	GetMember(ctx context.Context, teamID, memberID string) (msmodels.ConversationMemberable, *sender.RequestError)
	AddMember(ctx context.Context, teamID, userRef string, roles []string) (msmodels.ConversationMemberable, *sender.RequestError)
	RemoveMember(ctx context.Context, teamID, memberID string) *sender.RequestError
	UpdateMemberRoles(ctx context.Context, teamID, memberID string, roles []string) (msmodels.ConversationMemberable, *sender.RequestError)
}

type teamAPI struct {
	client    *graph.GraphServiceClient
	senderCfg *config.SenderConfig
}

func NewTeams(client *graph.GraphServiceClient, senderCfg *config.SenderConfig) TeamAPI {
	return &teamAPI{client, senderCfg}
}

func (t *teamAPI) CreateFromTemplate(ctx context.Context, displayName, description string, owners, members []string, visibility string) (string, *sender.RequestError) {
	if strings.TrimSpace(displayName) == "" {
		return "", &sender.RequestError{Code: http.StatusBadRequest, Message: "displayName cannot be empty"}
	}
	if len(owners) == 0 {
		return "", &sender.RequestError{Code: http.StatusBadRequest, Message: "at least one owner is required"}
	}

	body := msmodels.NewTeam()
	body.SetDisplayName(&displayName)
	if strings.TrimSpace(description) != "" {
		body.SetDescription(&description)
	}

	teamVisibility := msmodels.PUBLIC_TEAMVISIBILITYTYPE
	if strings.ToLower(strings.TrimSpace(visibility)) == "private" {
		teamVisibility = msmodels.PRIVATE_TEAMVISIBILITYTYPE
	}
	body.SetVisibility(&teamVisibility)

	first := "General"
	body.SetFirstChannelName(&first)
	body.SetAdditionalData(map[string]any{
		templateBindKey: templateBindValue,
	})

	var convMembers []msmodels.ConversationMemberable
	addToMembers(&convMembers, owners, []string{roleOwner})
	addToMembers(&convMembers, members, []string{})
	if len(convMembers) > 0 {
		body.SetMembers(convMembers)
	}
	var loc, contentLoc string
	responseHandler := func(resp any, _ abstractions.ErrorMappings) (any, error) {
		httpResp, ok := resp.(*http.Response)
		if !ok || httpResp == nil {
			return nil, nil
		}
		loc = httpResp.Header.Get("Location")
		contentLoc = httpResp.Header.Get("Content-Location")
		return nil, nil
	}

	ctx = context.WithValue(ctx, abstractions.ResponseHandlerOptionKey, responseHandler)

	call := func(ctx context.Context) (sender.Response, error) {
		return t.client.Teams().Post(ctx, body, nil)
	}
	if _, err := sender.SendRequest(ctx, call, t.senderCfg); err != nil {
		return "", err
	}
	teamID, ok := parseTeamIDFromHeaders(contentLoc, loc)
	if !ok || teamID == "" {
		return "", &sender.RequestError{Code: http.StatusInternalServerError, Message: "unable to parse team ID from response headers"}
	}
	if err := t.waitTeamReady(ctx, teamID, 30*time.Second); err != nil {
		return "", err
	}
	return teamID, nil
}

func (t *teamAPI) CreateViaGroup(ctx context.Context, displayName, mailNickname, visibility string) (string, *sender.RequestError) {
	grp := msmodels.NewGroup()
	grp.SetDisplayName(&displayName)
	grp.SetDescription(&displayName)
	grp.SetGroupTypes([]string{"Unified"})
	mailEnabled := true
	grp.SetMailEnabled(&mailEnabled)
	grp.SetMailNickname(&mailNickname)
	securityEnabled := false
	grp.SetSecurityEnabled(&securityEnabled)

	vis := normalizeVisibilityForGroup(visibility)
	grp.SetVisibility(&vis)

	createGroup := func(ctx context.Context) (sender.Response, error) {
		return t.client.Groups().Post(ctx, grp, nil)
	}

	gresp, gerr := sender.SendRequest(ctx, createGroup, t.senderCfg)
	if gerr != nil {
		return "", gerr
	}
	group, ok := gresp.(msmodels.Groupable)
	if !ok || group.GetId() == nil {
		return "", newTypeError("Groupable")
	}
	groupID := *group.GetId()

	body := msmodels.NewTeam()
	putTeam := func(ctx context.Context) (sender.Response, error) {
		return t.client.Groups().ByGroupId(groupID).Team().Put(ctx, body, nil)
	}
	if _, err := sender.SendRequest(ctx, putTeam, t.senderCfg); err != nil {
		return "", err
	}
	if err := t.waitTeamReady(ctx, groupID, 30*time.Second); err != nil {
		return "", err
	}
	return groupID, nil
}

func (t *teamAPI) Get(ctx context.Context, teamID string) (msmodels.Teamable, *sender.RequestError) {
	call := func(ctx context.Context) (sender.Response, error) {
		return t.client.Teams().ByTeamId(teamID).Get(ctx, nil)
	}
	resp, err := sender.SendRequest(ctx, call, t.senderCfg)
	if err != nil {
		return nil, err
	}
	out, ok := resp.(msmodels.Teamable)
	if !ok {
		return nil, newTypeError("Teamable")
	}
	return out, nil
}

func (t *teamAPI) ListMyJoined(ctx context.Context) (msmodels.TeamCollectionResponseable, *sender.RequestError) {
	call := func(ctx context.Context) (sender.Response, error) {
		return t.client.Me().JoinedTeams().Get(ctx, nil)
	}
	resp, err := sender.SendRequest(ctx, call, t.senderCfg)
	if err != nil {
		return nil, err
	}
	out, ok := resp.(msmodels.TeamCollectionResponseable)
	if !ok {
		return nil, newTypeError("TeamCollectionResponseable")
	}
	return out, nil
}

func (t *teamAPI) Archive(ctx context.Context, teamID string, spoReadOnlyForMembers *bool) *sender.RequestError {
	body := graphteams.NewItemArchivePostRequestBody()
	if spoReadOnlyForMembers != nil {
		body.SetShouldSetSpoSiteReadOnlyForMembers(spoReadOnlyForMembers)
	}
	call := func(ctx context.Context) (sender.Response, error) {
		return nil, t.client.
			Teams().
			ByTeamId(teamID).
			Archive().
			Post(ctx, body, nil)
	}
	_, err := sender.SendRequest(ctx, call, t.senderCfg)
	return err
}

func (t *teamAPI) Unarchive(ctx context.Context, teamID string) *sender.RequestError {
	call := func(ctx context.Context) (sender.Response, error) {
		return nil, t.client.
			Teams().
			ByTeamId(teamID).
			Unarchive().
			Post(ctx, nil)
	}
	_, err := sender.SendRequest(ctx, call, t.senderCfg)
	return err
}

func (t *teamAPI) Delete(ctx context.Context, teamID string) *sender.RequestError {
	call := func(ctx context.Context) (sender.Response, error) {
		return nil, t.client.
			Groups().
			ByGroupId(teamID).
			Delete(ctx, nil)
	}
	_, err := sender.SendRequest(ctx, call, t.senderCfg)
	return err
}

func (t *teamAPI) RestoreDeleted(ctx context.Context, deletedGroupID string) (msmodels.DirectoryObjectable, *sender.RequestError) {
	call := func(ctx context.Context) (sender.Response, error) {
		return t.client.
			Directory().
			DeletedItems().
			ByDirectoryObjectId(deletedGroupID).
			Restore().
			Post(ctx, nil)
	}
	resp, err := sender.SendRequest(ctx, call, t.senderCfg)
	if err != nil {
		return nil, err
	}
	out, ok := resp.(msmodels.DirectoryObjectable)
	if !ok {
		return nil, newTypeError("DirectoryObjectable")
	}
	return out, nil
}

func (t *teamAPI) UpdateTeam(ctx context.Context, teamID string, update *models.TeamUpdate) (msmodels.Teamable, *sender.RequestError) {
	if update == nil {
		return t.Get(ctx, teamID)
	}
	patch := msmodels.NewGroup()
	changed := false

	if update.DisplayName != nil {
		patch.SetDisplayName(update.DisplayName)
		changed = true
	}
	if update.Description != nil {
		patch.SetDescription(update.Description)
		changed = true
	}
	if update.Visibility != nil {
		vis := normalizeVisibilityForGroup(*update.Visibility)
		patch.SetVisibility(&vis)
		changed = true
	}
	if changed {
		call := func(ctx context.Context) (sender.Response, error) {
			return t.client.
				Groups().
				ByGroupId(teamID).
				Patch(ctx, patch, nil)
		}
		if _, err := sender.SendRequest(ctx, call, t.senderCfg); err != nil {
			return nil, err
		}
	}
	return t.Get(ctx, teamID)
}

func (t *teamAPI) ListMembers(ctx context.Context, teamID string) (msmodels.ConversationMemberCollectionResponseable, *sender.RequestError) {
	call := func(ctx context.Context) (sender.Response, error) {
		return t.client.
			Teams().
			ByTeamId(teamID).
			Members().
			Get(ctx, nil)
	}
	resp, err := sender.SendRequest(ctx, call, t.senderCfg)
	if err != nil {
		return nil, err
	}
	out, ok := resp.(msmodels.ConversationMemberCollectionResponseable)
	if !ok {
		return nil, newTypeError("ConversationMemberCollectionResponseable")
	}
	return out, nil
}

func (t *teamAPI) GetMember(ctx context.Context, teamID, memberID string) (msmodels.ConversationMemberable, *sender.RequestError) {
	call := func(ctx context.Context) (sender.Response, error) {
		return t.client.
			Teams().
			ByTeamId(teamID).
			Members().
			ByConversationMemberId(memberID).
			Get(ctx, nil)
	}
	resp, err := sender.SendRequest(ctx, call, t.senderCfg)
	if err != nil {
		return nil, err
	}
	out, ok := resp.(msmodels.ConversationMemberable)
	if !ok {
		return nil, newTypeError("ConversationMemberable")
	}
	return out, nil
}

func (t *teamAPI) AddMember(ctx context.Context, teamID, userRef string, roles []string) (msmodels.ConversationMemberable, *sender.RequestError) {
	member := newAadUserMemberBody(userRef, roles)
	call := func(ctx context.Context) (sender.Response, error) {
		return t.client.
			Teams().
			ByTeamId(teamID).
			Members().
			Post(ctx, member, nil)
	}
	resp, err := sender.SendRequest(ctx, call, t.senderCfg)
	if err != nil {
		return nil, err
	}
	out, ok := resp.(msmodels.ConversationMemberable)
	if !ok {
		return nil, newTypeError("ConversationMemberable")
	}
	return out, nil
}

func (t *teamAPI) RemoveMember(ctx context.Context, teamID, memberID string) *sender.RequestError {
	call := func(ctx context.Context) (sender.Response, error) {
		return nil, t.client.
			Teams().
			ByTeamId(teamID).
			Members().
			ByConversationMemberId(memberID).
			Delete(ctx, nil)
	}
	_, err := sender.SendRequest(ctx, call, t.senderCfg)
	return err
}

// Roles can be ["owner"] or [] (member)
func (t *teamAPI) UpdateMemberRoles(ctx context.Context, teamID, memberID string, roles []string) (msmodels.ConversationMemberable, *sender.RequestError) {
	patch := newRolesPatchBody(roles)
	call := func(ctx context.Context) (sender.Response, error) {
		return t.client.
			Teams().
			ByTeamId(teamID).
			Members().
			ByConversationMemberId(memberID).
			Patch(ctx, patch, nil)
	}
	resp, err := sender.SendRequest(ctx, call, t.senderCfg)
	if err != nil {
		return nil, err
	}
	out, ok := resp.(msmodels.ConversationMemberable)
	if !ok {
		return nil, newTypeError("ConversationMemberable")
	}
	return out, nil
}
