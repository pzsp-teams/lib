package teams

import (
	"context"

	"github.com/pzsp-teams/lib/internal/domain/channels"
)

type Channel = channels.Channel

type ChannelService interface {
	List(ctx context.Context, teamID string) ([]*Channel, error)
	Get(ctx context.Context, teamID, channelID string) (*Channel, error)
	Create(ctx context.Context, teamID, name string) (*Channel, error)
	Delete(ctx context.Context, teamID, channelID string) error
}

type channelService struct {
	inner *channels.Service
}

func (s *channelService) List(ctx context.Context, teamID string) ([]*Channel, error) {
	return s.inner.ListChannels(ctx, teamID)
}

func (s channelService) Get(ctx context.Context, teamID, channelID string) (*Channel, error) {
	return s.inner.Get(ctx, teamID, channelID)
}

func (s *channelService) Create(ctx context.Context, teamID, name string) (*Channel, error) {
	return s.inner.Create(ctx, teamID, name)
}

func (s *channelService) Delete(ctx context.Context, teamID, channelID string) error {
	return s.inner.Delete(ctx, teamID, channelID)
}