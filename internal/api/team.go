package api

import (
	"context"
	"net/http"
	"time"

	graph "github.com/microsoftgraph/msgraph-sdk-go"
	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	graphteams "github.com/microsoftgraph/msgraph-sdk-go/teams"

	"github.com/pzsp-teams/lib/config"
	"github.com/pzsp-teams/lib/internal/sender"
)

type TeamAPI interface {
	CreateFromTemplate(ctx context.Context, displayName, description string, owners []string) (string, *sender.RequestError)
	CreateViaGroup(ctx context.Context, displayName, mailNickname, visibility string) (string, *sender.RequestError)
	Get(ctx context.Context, teamID string) (msmodels.Teamable, *sender.RequestError)
	ListMyJoined(ctx context.Context) (msmodels.TeamCollectionResponseable, *sender.RequestError)
	Update(ctx context.Context, teamID string, patch *msmodels.Team) (msmodels.Teamable, *sender.RequestError)
	Archive(ctx context.Context, teamID string, spoReadOnlyForMembers *bool) *sender.RequestError
	Unarchive(ctx context.Context, teamID string) *sender.RequestError
	Delete(ctx context.Context, teamID string) *sender.RequestError
	RestoreDeleted(ctx context.Context, deletedGroupID string) (msmodels.DirectoryObjectable, *sender.RequestError)
	ListMembers(ctx context.Context, teamID string) (msmodels.ConversationMemberCollectionResponseable, *sender.RequestError)
	GetMember(ctx context.Context, teamID, memberID string) (msmodels.ConversationMemberable, *sender.RequestError)
	AddMember(ctx context.Context, teamID string, member msmodels.ConversationMemberable) (msmodels.ConversationMemberable, *sender.RequestError)
	RemoveMember(ctx context.Context, teamID, memberID string) *sender.RequestError
	UpdateMemberRoles(ctx context.Context, teamID, memberID string, roles []string) *sender.RequestError
}

type teamAPI struct {
	client    *graph.GraphServiceClient
	senderCfg *config.SenderConfig
}

func NewTeams(client *graph.GraphServiceClient, senderCfg *config.SenderConfig) TeamAPI {
	return &teamAPI{client, senderCfg}
}

func (t *teamAPI) CreateFromTemplate(ctx context.Context, displayName, description string, owners []string) (string, *sender.RequestError) {
	body := msmodels.NewTeam()
	body.SetDisplayName(&displayName)
	body.SetDescription(&description)
	first := "General"
	body.SetFirstChannelName(&first)
	body.SetAdditionalData(map[string]any{
		templateBindKey: templateBindValue,
	})
	call := func(ctx context.Context) (sender.Response, error) {
		return t.client.Teams().Post(ctx, body, nil)
	}
	resp, err := sender.SendRequest(ctx, call, t.senderCfg)
	if err != nil {
		return "", err
	}
	_ = resp
	return "id will be given later (async)", nil
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
	grp.SetVisibility(&visibility)
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

func (t *teamAPI) waitTeamReady(ctx context.Context, teamID string, timeout time.Duration) *sender.RequestError {
	deadline := time.Now().Add(timeout)
	for {
		call := func(ctx context.Context) (sender.Response, error) {
			return t.client.Teams().ByTeamId(teamID).Get(ctx, nil)
		}
		if _, err := sender.SendRequest(ctx, call, t.senderCfg); err == nil {
			return nil
		}
		if time.Now().After(deadline) {
			return &sender.RequestError{Code: http.StatusRequestTimeout, Message: "Team not ready within timeout"}
		}
		time.Sleep(2 * time.Second)
	}
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

func (t *teamAPI) Update(ctx context.Context, teamID string, patch *msmodels.Team) (msmodels.Teamable, *sender.RequestError) {
	call := func(ctx context.Context) (sender.Response, error) {
		return t.client.Teams().ByTeamId(teamID).Patch(ctx, patch, nil)
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

func (t *teamAPI) AddMember(ctx context.Context, teamID string, member msmodels.ConversationMemberable) (msmodels.ConversationMemberable, *sender.RequestError) {
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

func (t *teamAPI) UpdateMemberRoles(ctx context.Context, teamID, memberID string, roles []string) *sender.RequestError {
	patch := msmodels.NewConversationMember()
	patch.SetRoles(roles)
	call := func(ctx context.Context) (sender.Response, error) {
		return t.client.
			Teams().
			ByTeamId(teamID).
			Members().
			ByConversationMemberId(memberID).
			Patch(ctx, patch, nil)
	}
	_, err := sender.SendRequest(ctx, call, t.senderCfg)
	return err
}
