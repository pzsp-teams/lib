package channels

import (
	"context"

	"github.com/pzsp-teams/lib/models"
)

type Service interface {
	ListChannels(ctx context.Context, teamRef string) ([]*models.Channel, error)
	Get(ctx context.Context, teamRef, channelRef string) (*models.Channel, error)
	CreateStandardChannel(ctx context.Context, teamRef, name string) (*models.Channel, error)
	CreatePrivateChannel(ctx context.Context, teamRef, name string, memberRefs, ownerRefs []string) (*models.Channel, error)
	Delete(ctx context.Context, teamRef, channelRef string) error
	SendMessage(ctx context.Context, teamRef, channelRef string, body models.MessageBody) (*models.Message, error)
	ListMessages(ctx context.Context, teamRef, channelRef string, opts *models.ListMessagesOptions) ([]*models.Message, error)
	GetMessage(ctx context.Context, teamRef, channelRef, messageID string) (*models.Message, error)
	ListReplies(ctx context.Context, teamRef, channelRef, messageID string, top *int32) ([]*models.Message, error)
	GetReply(ctx context.Context, teamRef, channelRef, messageID, replyID string) (*models.Message, error)
	ListMembers(ctx context.Context, teamRef, channelRef string) ([]*models.Member, error)
	AddMember(ctx context.Context, teamRef, channelRef, userRef string, isOwner bool) (*models.Member, error)
	UpdateMemberRole(ctx context.Context, teamRef, channelRef, userRef string, isOwner bool) (*models.Member, error)
	RemoveMember(ctx context.Context, teamRef, channelRef, userRef string) error
}
