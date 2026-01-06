package chats

import (
	"context"
	"time"

	"github.com/pzsp-teams/lib/models"
)

type chatOps interface {
	CreateOneOnOne(ctx context.Context, userID string) (*models.Chat, error)
	CreateGroup(ctx context.Context, userIDs []string, topic string, includeMe bool) (*models.Chat, error)
	AddMemberToGroupChat(ctx context.Context, chatID, userID string) (*models.Member, error)
	RemoveMemberFromGroupChat(ctx context.Context, chatID, userID string) error
	ListGroupChatMembers(ctx context.Context, chatID string) ([]*models.Member, error)
	UpdateGroupChatTopic(ctx context.Context, chatID, topic string) (*models.Chat, error)
	ListMessages(ctx context.Context, chatID string, includeSystem bool) (*models.MessageCollection, error)
	SendMessage(ctx context.Context, chatID string, body models.MessageBody) (*models.Message, error)
	DeleteMessage(ctx context.Context, chatID, messageID string) error
	GetMessage(ctx context.Context, chatID, messageID string) (*models.Message, error)
	ListChats(ctx context.Context, chatType *models.ChatType) ([]*models.Chat, error)
	ListAllMessages(ctx context.Context, startTime, endTime *time.Time, top *int32) ([]*models.Message, error)
	ListPinnedMessages(ctx context.Context, chatID string) ([]*models.Message, error)
	PinMessage(ctx context.Context, chatID, messageID string) error
	UnpinMessage(ctx context.Context, chatID, messageID string) error
	GetMentions(ctx context.Context, chatID string, isGroup bool, rawMentions []string) ([]models.Mention, error)
	ListMessagesNext(ctx context.Context, chatID, nextLink string, includeSystem bool) (*models.MessageCollection, error)
}
