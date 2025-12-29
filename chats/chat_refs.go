package chats

type ChatRef interface {
	chatRef()
}

type GroupChatRef struct {
	Topic  string
	ChatID *string
}

type OneOnOneChatRef struct {
	UserRef string
	ChatID  *string
}

func (GroupChatRef) chatRef()    {}
func (OneOnOneChatRef) chatRef() {}
