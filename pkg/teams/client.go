package teams

import (
	graph "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/pzsp-teams/lib/internal/api"
	"github.com/pzsp-teams/lib/internal/domain/channels"
	"github.com/pzsp-teams/lib/internal/sender"
)

type Client struct {
	channels ChannelService
}

type Options struct {
	MaxRetries     int
	NextRetryDelay int
	Timeout        int
}

func NewClient(graphClient *graph.GraphServiceClient, opts *Options) *Client {
	techParams := sender.RequestTechParams{
		MaxRetries:     opts.MaxRetries,
		NextRetryDelay: opts.NextRetryDelay,
		Timeout:        opts.Timeout,
	}
	channelAPI := api.NewChannelsAPI(graphClient, techParams)
	chSvc := channels.NewService(channelAPI)

	return &Client{
		channels: &channelService{inner: chSvc},
	}
}

