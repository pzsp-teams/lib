package chats

type ChatRef interface {
	chatRef()
}

type GroupChatRef struct {
	Ref string
}

type OneOnOneChatRef struct {
	Ref string
}

func (GroupChatRef) chatRef()    {}
func (OneOnOneChatRef) chatRef() {}
