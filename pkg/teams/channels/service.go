package channels

import (
	"context"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
)

// Service will be used later
type Service struct {
	api ChannelAPI
}

// NewService will be used later
func NewService(api ChannelAPI) *Service {
	return &Service{api: api}
}

// ListChannels will be used later
func (s *Service) ListChannels(ctx context.Context, teamID string) ([]*Channel, error) {
	resp, err := s.api.ListChannels(ctx, teamID)
	if err != nil {
		return nil, mapError(err)
	}
	var channels []*Channel
	for _, ch := range resp.GetValue() {
		name := deref(ch.GetDisplayName())
		channels = append(channels, &Channel{
			ID:        deref(ch.GetId()),
			Name:      name,
			IsGeneral: name == "General",
		})
	}
	return channels, nil
}

// Get will be used later
func (s *Service) Get(ctx context.Context, teamID, channelID string) (*Channel, error) {
	resp, err := s.api.GetChannel(ctx, teamID, channelID)
	if err != nil {
		return nil, mapError(err)
	}
	name := deref(resp.GetDisplayName())
	return &Channel{
		ID:        deref(resp.GetId()),
		Name:      name,
		IsGeneral: name == "General",
	}, nil
}

// Create will be used later
func (s *Service) Create(ctx context.Context, teamID, name string) (*Channel, error) {
	newChannel := msmodels.NewChannel()
	newChannel.SetDisplayName(&name)

	created, err := s.api.CreateChannel(ctx, teamID, newChannel)
	if err != nil {
		return nil, mapError(err)
	}
	return &Channel{
		ID:        deref(created.GetId()),
		Name:      deref(created.GetDisplayName()),
		IsGeneral: false,
	}, nil
}

// Delete will be used later
func (s *Service) Delete(ctx context.Context, teamID, channelID string) error {
	err := s.api.DeleteChannel(ctx, teamID, channelID)
	if err != nil {
		return mapError(err)
	}
	return nil
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
