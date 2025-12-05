package chats

type Chat struct {
	ID       string
	ChatType string
	Members  []string
	IsHidden bool
}

type DirectChat struct {
	Chat
}

type GroupChat struct {
	Chat
	Topic string
}
