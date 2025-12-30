// Package chats provides various chat-related operations. It abstracts the underlying Microsoft Graph API calls.
// Package provides two services implementations - one with cache and one without cache. Both can be instantiated and used interchangeably.
// If cache is enabled, the service will use a caching layer to store and retrieve chat/member references, improving performance and reducing API calls.
// Concepts:
//   - Chats are either one-on-one or group chats.
//   - Users are identified by userID or email.
//   - Some operations require messageID - these can be obtained via ListMessages.
//   - ChatRef and GroupChatRef are references to chats used in method parameters.
//   - The authenticated user (derieved from MSAL) is the one making the API calls (appropriate scopes must be granted).
package chats

import (
	"context"
	"time"

	"github.com/pzsp-teams/lib/models"
)

// Service defines the interface for chat-relaated operations.
// It includes methods for creating chats, managing members, sending messages, and more.
type Service interface {
	// CreateOneOneOne creates a one-on-one chat with the given recipient.
	// The authenticated user is automatically added to the chat.
	CreateOneOneOne(ctx context.Context, recipientRef string) (*models.Chat, error)

	// CreateGroup creates a group chat with the given recipients and topic.
	// The authenticated user may be included by setting includeMe to true.
	CreateGroup(ctx context.Context, recipientRefs []string, topic string, includeMe bool) (*models.Chat, error)

	// AddMemberToGroupChat adds a user to a group chat.
	AddMemberToGroupChat(ctx context.Context, chatRef GroupChatRef, userRef string) (*models.Member, error)

	// RemoveMemberFromGroupChat removes a user from a group chat.
	RemoveMemberFromGroupChat(ctx context.Context, chatRef GroupChatRef, userRef string) error

	// ListGroupChatMembers returns all members of a group chat.
	ListGroupChatMembers(ctx context.Context, chatRef GroupChatRef) ([]*models.Member, error)

	// UpdateGroupChatTopic updates the topic of a group chat.
	UpdateGroupChatTopic(ctx context.Context, chatRef GroupChatRef, topic string) (*models.Chat, error)

	// ListMessages returns all messages in a chat.
	ListMessages(ctx context.Context, chatRef ChatRef) ([]*models.Message, error)

	// SendMessage sends a message to a chat. ContentType specifies the format of the message content.
	SendMessage(ctx context.Context, chatRef ChatRef, content string, contentType models.MessageContentType) (*models.Message, error)

	// DeleteMessage deletes a message from a chat. Action is reversible - soft delete is performed.
	DeleteMessage(ctx context.Context, chatRef ChatRef, messageID string) error

	// GetMessage retrieves a specific message from a chat by its ID.
	GetMessage(ctx context.Context, chatRef ChatRef, messageID string) (*models.Message, error)

	// ListChats returns all chats, optionally filtered by chat type.
	ListChats(ctx context.Context, chatType *models.ChatType) ([]*models.Chat, error)

	// ListAllMessages returns all messages in all chats within the specified time range. Top limits the number of messages returned.
	ListAllMessages(ctx context.Context, startTime, endTime *time.Time, top *int32) ([]*models.Message, error)

	// ListPinnedMessages returns all pinned messages in a chat.
	ListPinnedMessages(ctx context.Context, chatRef ChatRef) ([]*models.Message, error)

	// PinMessage pins a message in a chat.
	PinMessage(ctx context.Context, chatRef ChatRef, messageID string) error

	// UnpinMessage unpins a message in a chat.
	UnpinMessage(ctx context.Context, chatRef ChatRef, pinnedMessageID string) error
}
