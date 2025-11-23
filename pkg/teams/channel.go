package teams

import (
	"context"

	"github.com/pzsp-teams/lib/pkg/teams/channels"
)

// Channel will be used later
type Channel = channels.Channel

// ChannelService will be used later
type ChannelService interface {
	List(ctx context.Context, teamID string) ([]*Channel, error)
	Get(ctx context.Context, teamID, channelID string) (*Channel, error)
	Create(ctx context.Context, teamID, name string) (*Channel, error)
	Delete(ctx context.Context, teamID, channelID string) error
}

type channelService struct {
	inner *channels.Service
}

// List will be used later
func (s *channelService) List(ctx context.Context, teamID string) ([]*Channel, error) {
	return s.inner.ListChannels(ctx, teamID)
}

// Get will be used later
func (s *channelService) Get(ctx context.Context, teamID, channelID string) (*Channel, error) {
	return s.inner.Get(ctx, teamID, channelID)
}

// Create will be used later
func (s *channelService) Create(ctx context.Context, teamID, name string) (*Channel, error) {
	return s.inner.Create(ctx, teamID, name)
}

// Delete will be used later
func (s *channelService) Delete(ctx context.Context, teamID, channelID string) error {
	return s.inner.Delete(ctx, teamID, channelID)
}
