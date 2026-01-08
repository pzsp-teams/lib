package api

import (
	"context"
	"strings"

	graph "github.com/microsoftgraph/msgraph-sdk-go"
	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	graphsearch "github.com/microsoftgraph/msgraph-sdk-go/search"

	"github.com/pzsp-teams/lib/config"
	"github.com/pzsp-teams/lib/internal/sender"
	"github.com/pzsp-teams/lib/models"
)

type SearchAPI interface {
	// SearchChatMessages runs POST /search/query with entityTypes=["chatMessage"].
	SearchChatMessages(ctx context.Context, searchRequest *models.SearchMessagesOptions) (graphsearch.QueryPostResponseable, *sender.RequestError)
}

type searchAPI struct {
	client    *graph.GraphServiceClient
	senderCfg *config.SenderConfig
}

func NewSearch(client *graph.GraphServiceClient, senderCfg *config.SenderConfig) SearchAPI {
	return &searchAPI{client: client, senderCfg: senderCfg}
}

func (s *searchAPI) SearchChatMessages(ctx context.Context, searchRequest *models.SearchMessagesOptions) (graphsearch.QueryPostResponseable, *sender.RequestError) {
	call := func(ctx context.Context) (sender.Response, error) {
		body := graphsearch.NewQueryPostRequestBody()

		req := msmodels.NewSearchRequest()
		req.SetEntityTypes([]msmodels.EntityType{msmodels.CHATMESSAGE_ENTITYTYPE})
		req.SetFields([]string{"body", "chatId", "channelIdentity", "id", "messageType", "summary", })

		q := msmodels.NewSearchQuery()
		queryString := strings.TrimSpace(searchRequest.ParseQuery())
		q.SetQueryString(&queryString)
		req.SetQuery(q)

		if searchRequest.SearchPage != nil && searchRequest.SearchPage.From != nil {
			req.SetFrom(searchRequest.SearchPage.From)
		}
		if searchRequest.SearchPage != nil && searchRequest.SearchPage.Size != nil {
			req.SetSize(searchRequest.SearchPage.Size)
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
