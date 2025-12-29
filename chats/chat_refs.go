package chats

type ChatRef interface {
	chatRef()
}

type GroupChatRef struct {
	Topic string
}

type OneOnOneChatRef struct {
	UserRef string
}

type ChatIDRef struct {
	ChatID string
}

func (GroupChatRef) chatRef()    {}
func (OneOnOneChatRef) chatRef() {}
func (ChatIDRef) chatRef()       {}
