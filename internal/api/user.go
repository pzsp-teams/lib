package api

import (
	"context"
	"fmt"
	"strings"

	graph "github.com/microsoftgraph/msgraph-sdk-go"
	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	graphusers "github.com/microsoftgraph/msgraph-sdk-go/users"

	"github.com/pzsp-teams/lib/internal/sender"
)

type UsersAPI interface {
	GetUserIDByEmailOrUPN(ctx context.Context, emailOrUPN string) (string, *sender.RequestError)
}

type usersAPI struct {
	client     *graph.GraphServiceClient
	techParams sender.RequestTechParams
}

func NewUsers(client *graph.GraphServiceClient, techParams sender.RequestTechParams) UsersAPI {
	return &usersAPI{client: client, techParams: techParams}
}

func (u *usersAPI) GetUserIDByEmailOrUPN(ctx context.Context, emailOrUPN string) (string, *sender.RequestError) {
	key := strings.TrimSpace(emailOrUPN)
	if key == "" {
		return "", &sender.RequestError{Message: "emailOrUPN is empty"}
	}
	id, reqErr := u.getUserIDByKey(ctx, key)
	if reqErr == nil {
		return id, nil
	}
	if strings.Contains(key, "@") {
		id2, reqErr2 := u.findUserIDByEmail(ctx, key)
		if reqErr2 == nil {
			return id2, nil
		}
		return "", reqErr2
	}

	return "", reqErr
}

func (u *usersAPI) getUserIDByKey(ctx context.Context, key string) (string, *sender.RequestError) {
	call := func(ctx context.Context) (sender.Response, error) {
		return u.client.Users().ByUserId(key).Get(ctx, nil)
	}

	resp, err := sender.SendRequest(ctx, call, u.techParams)
	if err != nil {
		return "", err
	}

	userResp, ok := resp.(msmodels.Userable)
	if !ok {
		return "", newTypeError("Userable")
	}

	userIDPtr := userResp.GetId()
	if userIDPtr == nil || strings.TrimSpace(*userIDPtr) == "" {
		return "", &sender.RequestError{Message: fmt.Sprintf("user id is empty for key=%q", key)}
	}

	return *userIDPtr, nil
}

func (u *usersAPI) findUserIDByEmail(ctx context.Context, email string) (string, *sender.RequestError) {
	escaped := strings.ReplaceAll(strings.TrimSpace(email), "'", "''")
	filter := fmt.Sprintf(
		"mail eq '%[1]s' or userPrincipalName eq '%[1]s' or otherMails/any(x:x eq '%[1]s') or "+
			"proxyAddresses/any(p:p eq 'SMTP:%[1]s') or proxyAddresses/any(p:p eq 'smtp:%[1]s')",
		escaped,
	)
	top := int32(2)
	cfg := &graphusers.UsersRequestBuilderGetRequestConfiguration{
		QueryParameters: &graphusers.UsersRequestBuilderGetQueryParameters{
			Filter: &filter,
			Select: []string{"id"},
			Top:    &top,
		},
	}
	call := func(ctx context.Context) (sender.Response, error) {
		return u.client.Users().Get(ctx, cfg)
	}
	resp, err := sender.SendRequest(ctx, call, u.techParams)
	if err != nil {
		return "", err
	}
	col, ok := resp.(msmodels.UserCollectionResponseable)
	if !ok {
		return "", newTypeError("UserCollectionResponseable")
	}
	values := col.GetValue()
	if len(values) == 0 {
		return "", &sender.RequestError{Message: fmt.Sprintf("user not found by email=%q", email)}
	}
	if len(values) > 1 {
		return "", &sender.RequestError{Message: fmt.Sprintf("email=%q is ambiguous (multiple users matched)", email)}
	}

	idPtr := values[0].GetId()
	if idPtr == nil || strings.TrimSpace(*idPtr) == "" {
		return "", &sender.RequestError{Message: fmt.Sprintf("matched user has empty id for email=%q", email)}
	}

	return *idPtr, nil
}
