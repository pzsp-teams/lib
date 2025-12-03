package teams

import (
	"context"
	"fmt"
	"time"

	graph "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	graphteams "github.com/microsoftgraph/msgraph-sdk-go/teams"

	"github.com/pzsp-teams/lib/internal/sender"
)

type API struct {
	client     *graph.GraphServiceClient
	techParams sender.RequestTechParams
}

func NewTeamsAPI(client *graph.GraphServiceClient, techParams sender.RequestTechParams) *API {
	return &API{client, techParams}
}

func (api *API) CreateFromTemplate(ctx context.Context, displayName, description string, owners []string) (string, *sender.RequestError) {
	body := models.NewTeam()
	body.SetDisplayName(&displayName)
	body.SetDescription(&description)
	first := "General"
	body.SetFirstChannelName(&first)
	body.SetAdditionalData(map[string]interface{}{
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
	return "", &sender.RequestError{Code: "AsyncOperation", Message: "Team creation started (202). Consider group->team flow to know id immediately."}
}

func (api *API) CreateViaGroup(ctx context.Context, displayName, mailNickname, visibility string) (string, *sender.RequestError) {
	grp := models.NewGroup()
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
	group, ok := gresp.(models.Groupable)
	if !ok || group.GetId() == nil {
		return "", &sender.RequestError{Code: "TypeCastError", Message: "Expected Groupable"}
	}
	groupID := *group.GetId()
	body := models.NewTeam()
	body.SetAdditionalData(map[string]interface{}{
		"template@odata.bind": "https://graph.microsoft.com/v1.0/teamsTemplates('standard')",
	})
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

func (api *API) waitTeamReady(ctx context.Context, teamID string, timeout time.Duration) *sender.RequestError {
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

func (api *API) Get(ctx context.Context, teamID string) (models.Teamable, *sender.RequestError) {
	call := func(ctx context.Context) (sender.Response, error) {
		return api.client.Teams().ByTeamId(teamID).Get(ctx, nil)
	}
	resp, err := sender.SendRequest(ctx, call, api.techParams)
	if err != nil {
		return nil, err
	}
	out, ok := resp.(models.Teamable)
	if !ok {
		return nil, &sender.RequestError{Code: "TypeCastError", Message: "Expected Teamable"}
	}
	return out, nil
}

func (api *API) ListMyJoined(ctx context.Context) (models.TeamCollectionResponseable, *sender.RequestError) {
	call := func(ctx context.Context) (sender.Response, error) {
		return api.client.Me().JoinedTeams().Get(ctx, nil)
	}
	resp, err := sender.SendRequest(ctx, call, api.techParams)
	if err != nil {
		return nil, err
	}
	out, ok := resp.(models.TeamCollectionResponseable)
	if !ok {
		return nil, &sender.RequestError{Code: "TypeCastError", Message: "Expected TeamCollectionResponseable"}
	}
	return out, nil
}

func (api *API) Update(ctx context.Context, teamID string, patch *models.Team) (models.Teamable, *sender.RequestError) {
	call := func(ctx context.Context) (sender.Response, error) {
		return api.client.Teams().ByTeamId(teamID).Patch(ctx, patch, nil)
	}
	resp, err := sender.SendRequest(ctx, call, api.techParams)
	if err != nil {
		return nil, err
	}
	out, ok := resp.(models.Teamable)
	if !ok {
		return nil, &sender.RequestError{Code: "TypeCastError", Message: "Expected Teamable"}
	}
	return out, nil
}

func (api *API) Archive(ctx context.Context, teamID string, spoReadOnlyForMembers *bool) *sender.RequestError {
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

func (api *API) Unarchive(ctx context.Context, teamID string) *sender.RequestError {
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

func (api *API) Delete(ctx context.Context, teamID string) *sender.RequestError {
    call := func(ctx context.Context) (sender.Response, error) {
        return nil, api.client.
            Groups().
            ByGroupId(teamID).
            Delete(ctx, nil)
    }
    _, err := sender.SendRequest(ctx, call, api.techParams)
    return err
}

func (api *API) RestoreDeleted(ctx context.Context, deletedGroupID string) (models.DirectoryObjectable, *sender.RequestError) {
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
    out, ok := resp.(models.DirectoryObjectable)
    if !ok {
        return nil, &sender.RequestError{Code: "TypeCastError", Message: fmt.Sprintf("Expected DirectoryObjectable, got %T", resp)}
    }
    return out, nil
}
