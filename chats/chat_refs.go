package chats

// ChatRef is an interface representing a reference to a chat, which can be either a group chat or a one-on-one chat.
type ChatRef interface {
	chatRef()
}

// GroupChatRef identifies a group chat.
// It may reference a chat by:
//
//   - unique chatID
//
//   - chat topic
//
// Note: Using chat topic may lead to ambiguities (which must be resolved manually) if multiple group chats share the same topic.
type GroupChatRef struct {
	Ref string
}

// OneOnOneChatRef identifies a one-on-one chat.
// It may reference a chat by:
//
//   - unique chatID
//
//   - recipient's reference (userID, email)
//
// Note: Chat must be established between the logged-in user and the recipient for resolution to succeed.,
type OneOnOneChatRef struct {
	Ref string
}

func (GroupChatRef) chatRef()    {}
func (OneOnOneChatRef) chatRef() {}
