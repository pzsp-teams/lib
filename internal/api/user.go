package api

import (
	"context"
	"strings"

	graph "github.com/microsoftgraph/msgraph-sdk-go"
	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"

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
	call := func(ctx context.Context) (sender.Response, error) {
		return u.client.Users().ByUserId(emailOrUPN).Get(ctx, nil)
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
		return "", newTypeError("ID")
	}

	return *userIDPtr, nil
}
