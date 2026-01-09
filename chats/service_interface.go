// Package chats provides various chat-related operations. It abstracts the underlying Microsoft Graph API calls.
// Package provides two services implementations - one with cache and one without cache. Both can be instantiated and used interchangeably.
// If cache is enabled, the service will use a caching layer to store and retrieve chat/member references, improving performance and reducing API calls.
// Concepts:
//   - Chats are either one-on-one or group chats.
//   - Users are identified by userID or email.
//   - Some operations require messageID - these can be obtained via ListMessages.
//   - ChatRef and GroupChatRef are references to chats used in method parameters.
//   - If chatRef is a topic and is not unique, an ambiguity error is returned.
//   - The authenticated user (derived from MSAL) is the one making the API calls (appropriate scopes must be granted).
//
// If an async cached service is used, call Wait() to ensure all background cache updates are finished.
package chats

import (
	"context"
	"time"

	"github.com/pzsp-teams/lib/models"
)

// Service defines the interface for chat-related operations.
// It includes methods for creating chats, managing members, sending messages, and more.
type Service interface {
	// CreateOneOneOne creates a one-on-one chat with the given recipient.
	// The authenticated user is automatically added to the chat.
	CreateOneOnOne(ctx context.Context, recipientRef string) (*models.Chat, error)

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
	//
	// NextLink in the returned MessageCollection can be used to retrieve the next page of messages.
	ListMessages(ctx context.Context, chatRef ChatRef, includeSystem bool, nextLink *string) (*models.MessageCollection, error)

	// SendMessage sends a message to a chat.
	// Body parameter is the body of the message. It includes:
	//   - Content: the text or html content of the message.
	//   - ContentType: the type of content (text or html).
	//   - Mentions: optional mentions to include in the message.
	SendMessage(ctx context.Context, chatRef ChatRef, body models.MessageBody) (*models.Message, error)

	// DeleteMessage deletes a message from a chat. Action is reversible - soft delete is performed.
	DeleteMessage(ctx context.Context, chatRef ChatRef, messageID string) error

	// GetMessage retrieves a specific message from a chat by its ID.
	GetMessage(ctx context.Context, chatRef ChatRef, messageID string) (*models.Message, error)

	// ListChats returns all chats, optionally filtered by chat type.
	ListChats(ctx context.Context, chatType *models.ChatType) ([]*models.Chat, error)

	// ListAllMessages returns all messages in all chats within the specified time range. Top limits the number of messages returned.
	//
	// Note: This operation does not work in delegated permission mode.
	ListAllMessages(ctx context.Context, startTime, endTime *time.Time, top *int32) ([]*models.Message, error)

	// ListPinnedMessages returns all pinned messages in a chat.
	ListPinnedMessages(ctx context.Context, chatRef ChatRef) ([]*models.Message, error)

	// PinMessage pins a message in a chat.
	PinMessage(ctx context.Context, chatRef ChatRef, messageID string) error

	// UnpinMessage unpins a message in a chat.
	UnpinMessage(ctx context.Context, chatRef ChatRef, pinnedMessageID string) error

	// GetMentions resolves raw mention strings to Mention objects in the context of a chat. Raw mentions can be:
	//   - Emails
	//   - Everyone (for group chats)
	//   - User IDs
	GetMentions(ctx context.Context, chatRef ChatRef, rawMentions []string) ([]models.Mention, error)

	SearchMessages(ctx context.Context, chatRef ChatRef, opts *models.SearchMessagesOptions) (*models.SearchResults, error)
}
