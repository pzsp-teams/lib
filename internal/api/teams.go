package api

import (
	"context"
	"fmt"
	"time"

	graph "github.com/microsoftgraph/msgraph-sdk-go"
	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	graphteams "github.com/microsoftgraph/msgraph-sdk-go/teams"

	"github.com/pzsp-teams/lib/internal/sender"
)

type Teams interface {
	CreateFromTemplate(ctx context.Context, displayName, description string, owners []string) (string, *sender.RequestError)
	CreateViaGroup(ctx context.Context, displayName, mailNickname, visibility string) (string, *sender.RequestError)
	Get(ctx context.Context, teamID string) (msmodels.Teamable, *sender.RequestError)
	ListMyJoined(ctx context.Context) (msmodels.TeamCollectionResponseable, *sender.RequestError)
	Update(ctx context.Context, teamID string, patch *msmodels.Team) (msmodels.Teamable, *sender.RequestError)
	Archive(ctx context.Context, teamID string, spoReadOnlyForMembers *bool) *sender.RequestError
	Unarchive(ctx context.Context, teamID string) *sender.RequestError
	Delete(ctx context.Context, teamID string) *sender.RequestError
	RestoreDeleted(ctx context.Context, deletedGroupID string) (msmodels.DirectoryObjectable, *sender.RequestError)
}

type teams struct {
	client     *graph.GraphServiceClient
	techParams sender.RequestTechParams
}

// NewTeams will be used later
func NewTeams(client *graph.GraphServiceClient, techParams sender.RequestTechParams) *teams {
	return &teams{client, techParams}
}

// CreateFromTemplate will be used later
func (api *teams) CreateFromTemplate(ctx context.Context, displayName, description string, owners []string) (string, *sender.RequestError) {
	body := msmodels.NewTeam()
	body.SetDisplayName(&displayName)
	body.SetDescription(&description)
	first := "General"
	body.SetFirstChannelName(&first)
	body.SetAdditionalData(map[string]any{
		"template@odata.bind": "https://graph.microsoft.com/v1.0/teamsTemplates('standard')",
	})
	call := func(ctx context.Context) (sender.Response, error) {
		return api.client.Teams().Post(ctx, body, nil)
	}
	resp, err := sender.SendRequest(ctx, call, api.techParams)
	if err != nil {
		return "", err
	}
	_ = resp
	return "id will be given later (async)", nil
}

// CreateViaGroup will be used later
func (api *teams) CreateViaGroup(ctx context.Context, displayName, mailNickname, visibility string) (string, *sender.RequestError) {
	grp := msmodels.NewGroup()
	grp.SetDisplayName(&displayName)
	grp.SetDescription(&displayName)
	grp.SetGroupTypes([]string{"Unified"})
	mailEnabled := true
	grp.SetMailEnabled(&mailEnabled)
	grp.SetMailNickname(&mailNickname)
	securityEnabled := false
	grp.SetSecurityEnabled(&securityEnabled)
	grp.SetVisibility(&visibility)
	createGroup := func(ctx context.Context) (sender.Response, error) {
		return api.client.Groups().Post(ctx, grp, nil)
	}
	gresp, gerr := sender.SendRequest(ctx, createGroup, api.techParams)
	if gerr != nil {
		return "", gerr
	}
	group, ok := gresp.(msmodels.Groupable)
	if !ok || group.GetId() == nil {
		return "", &sender.RequestError{Code: "TypeCastError", Message: "Expected Groupable"}
	}
	groupID := *group.GetId()
	body := msmodels.NewTeam()
	putTeam := func(ctx context.Context) (sender.Response, error) {
		return api.client.Groups().ByGroupId(groupID).Team().Put(ctx, body, nil)
	}
	if _, err := sender.SendRequest(ctx, putTeam, api.techParams); err != nil {
		return "", err
	}
	if err := api.waitTeamReady(ctx, groupID, 30*time.Second); err != nil {
		return "", err
	}
	return groupID, nil
}

func (api *teams) waitTeamReady(ctx context.Context, teamID string, timeout time.Duration) *sender.RequestError {
	deadline := time.Now().Add(timeout)
	for {
		call := func(ctx context.Context) (sender.Response, error) {
			return api.client.Teams().ByTeamId(teamID).Get(ctx, nil)
		}
		if _, err := sender.SendRequest(ctx, call, api.techParams); err == nil {
			return nil
		}
		if time.Now().After(deadline) {
			return &sender.RequestError{Code: "Timeout", Message: "Team not ready within timeout"}
		}
		time.Sleep(2 * time.Second)
	}
}

// Get will be used later
func (api *teams) Get(ctx context.Context, teamID string) (msmodels.Teamable, *sender.RequestError) {
	call := func(ctx context.Context) (sender.Response, error) {
		return api.client.Teams().ByTeamId(teamID).Get(ctx, nil)
	}
	resp, err := sender.SendRequest(ctx, call, api.techParams)
	if err != nil {
		return nil, err
	}
	out, ok := resp.(msmodels.Teamable)
	if !ok {
		return nil, &sender.RequestError{Code: "TypeCastError", Message: "Expected Teamable"}
	}
	return out, nil
}

// ListMyJoined will be used later
func (api *teams) ListMyJoined(ctx context.Context) (msmodels.TeamCollectionResponseable, *sender.RequestError) {
	call := func(ctx context.Context) (sender.Response, error) {
		return api.client.Me().JoinedTeams().Get(ctx, nil)
	}
	resp, err := sender.SendRequest(ctx, call, api.techParams)
	if err != nil {
		return nil, err
	}
	out, ok := resp.(msmodels.TeamCollectionResponseable)
	if !ok {
		return nil, &sender.RequestError{Code: "TypeCastError", Message: "Expected TeamCollectionResponseable"}
	}
	return out, nil
}

// Update will be used later
func (api *teams) Update(ctx context.Context, teamID string, patch *msmodels.Team) (msmodels.Teamable, *sender.RequestError) {
	call := func(ctx context.Context) (sender.Response, error) {
		return api.client.Teams().ByTeamId(teamID).Patch(ctx, patch, nil)
	}
	resp, err := sender.SendRequest(ctx, call, api.techParams)
	if err != nil {
		return nil, err
	}
	out, ok := resp.(msmodels.Teamable)
	if !ok {
		return nil, &sender.RequestError{Code: "TypeCastError", Message: "Expected Teamable"}
	}
	return out, nil
}

// Archive will be used later
func (api *teams) Archive(ctx context.Context, teamID string, spoReadOnlyForMembers *bool) *sender.RequestError {
	body := graphteams.NewItemArchivePostRequestBody()
	if spoReadOnlyForMembers != nil {
		body.SetShouldSetSpoSiteReadOnlyForMembers(spoReadOnlyForMembers)
	}
	call := func(ctx context.Context) (sender.Response, error) {
		return nil, api.client.
			Teams().
			ByTeamId(teamID).
			Archive().
			Post(ctx, body, nil)
	}
	_, err := sender.SendRequest(ctx, call, api.techParams)
	return err
}

// Unarchive will be used later
func (api *teams) Unarchive(ctx context.Context, teamID string) *sender.RequestError {
	call := func(ctx context.Context) (sender.Response, error) {
		return nil, api.client.
			Teams().
			ByTeamId(teamID).
			Unarchive().
			Post(ctx, nil)
	}
	_, err := sender.SendRequest(ctx, call, api.techParams)
	return err
}

// Delete will be used later
func (api *teams) Delete(ctx context.Context, teamID string) *sender.RequestError {
	call := func(ctx context.Context) (sender.Response, error) {
		return nil, api.client.
			Groups().
			ByGroupId(teamID).
			Delete(ctx, nil)
	}
	_, err := sender.SendRequest(ctx, call, api.techParams)
	return err
}

// RestoreDeleted will be used later
func (api *teams) RestoreDeleted(ctx context.Context, deletedGroupID string) (msmodels.DirectoryObjectable, *sender.RequestError) {
	call := func(ctx context.Context) (sender.Response, error) {
		return api.client.
			Directory().
			DeletedItems().
			ByDirectoryObjectId(deletedGroupID).
			Restore().
			Post(ctx, nil)
	}
	resp, err := sender.SendRequest(ctx, call, api.techParams)
	if err != nil {
		return nil, err
	}
	out, ok := resp.(msmodels.DirectoryObjectable)
	if !ok {
		return nil, &sender.RequestError{Code: "TypeCastError", Message: fmt.Sprintf("Expected DirectoryObjectable, got %T", resp)}
	}
	return out, nil
}
