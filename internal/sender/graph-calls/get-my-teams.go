package graphcalls

import (
	"context"

	sender "github.com/pzsp-teams/lib/internal/sender"

	msgraph "github.com/microsoftgraph/msgraph-sdk-go"
	graphmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
)

// GetMyTeams will be used later by other packages
func GetMyTeams(client *msgraph.GraphServiceClient) sender.GraphCall {
	return func(ctx context.Context) (sender.Response, error) {
		received, err := client.Me().JoinedTeams().Get(ctx, nil)
		if err != nil {
			return nil, err
		}

		response, err := parseGetMyTeamsResponse(received)
		if err != nil {
			return nil, err
		}
		return response, nil // TODO return parsed response
	}
}

// nolint
func parseGetMyTeamsResponse(received graphmodels.TeamCollectionResponseable) (sender.Response, error) {
	// TODO implement parsing logic
	return nil, nil
}
