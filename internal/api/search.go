package api

import (
	"context"
	"strings"

	graph "github.com/microsoftgraph/msgraph-sdk-go"
	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	graphsearch "github.com/microsoftgraph/msgraph-sdk-go/search"

	"github.com/pzsp-teams/lib/config"
	"github.com/pzsp-teams/lib/internal/sender"
	"github.com/pzsp-teams/lib/search"
)

type SearchEntity struct {
	ChannelID *string
	TeamID    *string
	MessageID *string
	ChatID    *string
}

type SearchMessage struct {
	Message   msmodels.ChatMessageable
	ChannelID *string
	TeamID    *string
	ChatID    *string
}

type SearchAPI interface {
	// SearchMessages runs POST /search/query with entityTypes=["chatMessage"].
	SearchMessages(ctx context.Context, searchRequest *search.SearchMessagesOptions) (graphsearch.QueryPostResponseable, *sender.RequestError)
}

type searchAPI struct {
	client    *graph.GraphServiceClient
	senderCfg *config.SenderConfig
}

func NewSearch(client *graph.GraphServiceClient, senderCfg *config.SenderConfig) SearchAPI {
	return &searchAPI{client: client, senderCfg: senderCfg}
}

func (s *searchAPI) SearchMessages(ctx context.Context, searchRequest *search.SearchMessagesOptions) (graphsearch.QueryPostResponseable, *sender.RequestError) {
	call := func(ctx context.Context) (sender.Response, error) {
		body := graphsearch.NewQueryPostRequestBody()

		req := msmodels.NewSearchRequest()
		req.SetEntityTypes([]msmodels.EntityType{msmodels.CHATMESSAGE_ENTITYTYPE})

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
	if searchRequest.NotFromMe || searchRequest.NotToMe {
		me, err := GetMe(ctx, s.client, s.senderCfg)
		if err != nil {
			return nil, err
		}
		if searchRequest.NotFromMe {
			if me != nil && me.GetDisplayName() != nil {
				searchRequest.NotFrom = append(searchRequest.NotFrom, strings.TrimSpace(*me.GetDisplayName()))
			}
		}
		if searchRequest.NotToMe {
			if me != nil && me.GetDisplayName() != nil {
				searchRequest.NotTo = append(searchRequest.NotTo, strings.TrimSpace(*me.GetDisplayName()))
			}
		}
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
