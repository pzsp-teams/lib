package api

import (
	"context"
	"fmt"
	"strings"

	graph "github.com/microsoftgraph/msgraph-sdk-go"
	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	graphusers "github.com/microsoftgraph/msgraph-sdk-go/users"

	"github.com/pzsp-teams/lib/config"
	"github.com/pzsp-teams/lib/internal/sender"
	"github.com/pzsp-teams/lib/internal/util"
)

type UserAPI interface {
	GetUserByEmailOrUPN(ctx context.Context, emailOrUPN string) (msmodels.Userable, *sender.RequestError)
}

type userAPI struct {
	client    *graph.GraphServiceClient
	snederCfg *config.SenderConfig
}

func NewUser(client *graph.GraphServiceClient, senderCfg *config.SenderConfig) UserAPI {
	return &userAPI{client, senderCfg}
}

func (u *userAPI) GetUserByEmailOrUPN(ctx context.Context, emailOrUPN string) (msmodels.Userable, *sender.RequestError) {
	key := strings.TrimSpace(emailOrUPN)
	if key == "" {
		return nil, &sender.RequestError{Message: "emailOrUPN is empty"}
	}
	user, reqErr := u.getUserByKey(ctx, key)
	if reqErr == nil {
		return user, nil
	}
	if util.IsLikelyEmail(key) {
		user, reqErr := u.findUserByEmail(ctx, key)
		if reqErr == nil {
			return user, nil
		}
		return nil, reqErr
	}

	return nil, reqErr
}

func (u *userAPI) getUserByKey(ctx context.Context, key string) (msmodels.Userable, *sender.RequestError) {
	cfg := &graphusers.UserItemRequestBuilderGetRequestConfiguration{
		QueryParameters: &graphusers.UserItemRequestBuilderGetQueryParameters{
			Select: []string{"id", "displayName", "mail", "userPrincipalName"},
		},
	}
	call := func(ctx context.Context) (sender.Response, error) {
		return u.client.Users().ByUserId(key).Get(ctx, cfg)
	}

	resp, err := sender.SendRequest(ctx, call, u.snederCfg)
	if err != nil {
		return nil, err
	}

	userResp, ok := resp.(msmodels.Userable)
	if !ok {
		return nil, newTypeError("Userable")
	}

	if userResp.GetId() == nil || strings.TrimSpace(*userResp.GetId()) == "" {
		return nil, &sender.RequestError{Message: fmt.Sprintf("user id is empty for key=%q", key)}
	}
	if userResp.GetDisplayName() == nil || strings.TrimSpace(*userResp.GetDisplayName()) == "" {
		return nil, &sender.RequestError{Message: fmt.Sprintf("user displayName is empty for key=%q", key)}
	}

	return userResp, nil
}

func (u *userAPI) findUserByEmail(ctx context.Context, email string) (msmodels.Userable, *sender.RequestError) {
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
			Select: []string{"id", "displayName", "mail", "userPrincipalName"},
			Top:    &top,
		},
	}

	call := func(ctx context.Context) (sender.Response, error) {
		return u.client.Users().Get(ctx, cfg)
	}

	resp, err := sender.SendRequest(ctx, call, u.snederCfg)
	if err != nil {
		return nil, err
	}

	col, ok := resp.(msmodels.UserCollectionResponseable)
	if !ok {
		return nil, newTypeError("UserCollectionResponseable")
	}

	values := col.GetValue()
	if len(values) == 0 {
		return nil, &sender.RequestError{Message: fmt.Sprintf("user not found by email=%q", email)}
	}
	if len(values) > 1 {
		return nil, &sender.RequestError{Message: fmt.Sprintf("email=%q is ambiguous (multiple users matched)", email)}
	}

	u0 := values[0]
	if u0.GetId() == nil || strings.TrimSpace(*u0.GetId()) == "" {
		return nil, &sender.RequestError{Message: fmt.Sprintf("matched user has empty id for email=%q", email)}
	}
	if u0.GetDisplayName() == nil || strings.TrimSpace(*u0.GetDisplayName()) == "" {
		return nil, &sender.RequestError{Message: fmt.Sprintf("matched user has empty displayName for email=%q", email)}
	}

	return u0, nil
}

func GetMe(ctx context.Context, client *graph.GraphServiceClient, senderCfg *config.SenderConfig) (msmodels.Userable, *sender.RequestError) {
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
