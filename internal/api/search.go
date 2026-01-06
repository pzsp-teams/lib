package api

import (
	"context"

	graph "github.com/microsoftgraph/msgraph-sdk-go"
	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	graphsearch "github.com/microsoftgraph/msgraph-sdk-go/search"

	"github.com/pzsp-teams/lib/config"
	"github.com/pzsp-teams/lib/internal/sender"
)

type SearchAPI interface {
	// SearchChatMessages runs POST /search/query with entityTypes=["chatMessage"].
	SearchChatMessages(ctx context.Context, queryString string, from, size *int32) (graphsearch.QueryPostResponseable, *sender.RequestError)
}

type searchAPI struct {
	client    *graph.GraphServiceClient
	senderCfg *config.SenderConfig
}

func NewSearch(client *graph.GraphServiceClient, senderCfg *config.SenderConfig) SearchAPI {
	return &searchAPI{client: client, senderCfg: senderCfg}
}

func (s *searchAPI) SearchChatMessages(ctx context.Context, queryString string, from, size *int32) (graphsearch.QueryPostResponseable, *sender.RequestError) {
	call := func(ctx context.Context) (sender.Response, error) {
		body := graphsearch.NewQueryPostRequestBody()

		req := msmodels.NewSearchRequest()
		req.SetEntityTypes([]msmodels.EntityType{msmodels.CHATMESSAGE_ENTITYTYPE})

		q := msmodels.NewSearchQuery()
		q.SetQueryString(&queryString)
		req.SetQuery(q)

		if from != nil {
			req.SetFrom(from)
		}
		if size != nil {
			req.SetSize(size)
		}

		body.SetRequests([]msmodels.SearchRequestable{req})
		return s.client.Search().Query().PostAsQueryPostResponse(ctx, body, nil)
	}

	resp, err := sender.SendRequest(ctx, call, s.senderCfg)
	if err != nil {
		return nil, err
	}

	out, ok := resp.(graphsearch.QueryPostResponseable)
	if !ok {
		return nil, newTypeError("QueryPostResponseable")
	}
	return out, nil
}
