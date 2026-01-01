package chats

import (
	"fmt"
	"strings"

	"github.com/pzsp-teams/lib/internal/mentions"
	"github.com/pzsp-teams/lib/models"
)

func isGroupChatRef(chatRef ChatRef) (bool, error) {
	switch chatRef.(type) {
	case GroupChatRef:
		return true, nil
	case OneOnOneChatRef:
		return false, nil
	default:
		return false, fmt.Errorf("unknown chatRef type")
	}
}

func tryAddEveryoneMention(adder *mentions.MentionAdder, chatID string, isGroup bool, raw string) (bool, error) {
	low := strings.ToLower(strings.TrimSpace(raw))
	if low != "everyone" && low != "@everyone" {
		return false, nil
	}
	if !isGroup {
		return false, fmt.Errorf("cannot mention everyone in one-on-one chat")
	}
	adder.Add(models.MentionEveryone, chatID, "Everyone")
	return true, nil
}

func checkThisMentionValidity(isGroup bool, raw string) bool {
	low := strings.ToLower(strings.TrimSpace(raw))
	return !isGroup && (low == "this" || low == "@this")
}
