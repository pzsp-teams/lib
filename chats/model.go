package chats

type Chat struct {
	ID       string
	ChatType string
	Members  []string
	IsHidden bool
	Topic    *string
}
