package chats

import (
	"testing"

	"github.com/pzsp-teams/lib/internal/mentions"
	"github.com/pzsp-teams/lib/models"
)

func TestIsGroupChatRef_Group(t *testing.T) {
	got, err := isGroupChatRef(GroupChatRef{Ref: "x"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got {
		t.Fatalf("expected true, got false")
	}
}

func TestIsGroupChatRef_OneOnOne(t *testing.T) {
	got, err := isGroupChatRef(OneOnOneChatRef{Ref: "x"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got {
		t.Fatalf("expected false, got true")
	}
}

type unknownChatRef struct{}

func (u unknownChatRef) chatRef() {}
func TestIsGroupChatRef_Unknown(t *testing.T) {
	_, err := isGroupChatRef(unknownChatRef{})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestTryAddEveryoneMention_NotEveryone_ReturnsFalseNoErrorAndDoesNotAdd(t *testing.T) {
	out := make([]models.Mention, 0)
	adder := mentions.NewMentionAdder(&out)

	ok, err := tryAddEveryoneMention(adder, "chat-1", true, "alice@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatalf("expected ok=false, got true")
	}
	if len(out) != 0 {
		t.Fatalf("expected no mentions added, got %d", len(out))
	}
}

func TestTryAddEveryoneMention_GroupChat_AddsEveryoneMention(t *testing.T) {
	out := make([]models.Mention, 0)
	adder := mentions.NewMentionAdder(&out)

	ok, err := tryAddEveryoneMention(adder, "chat-1", true, "  @Everyone  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatalf("expected ok=true, got false")
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 mention added, got %d", len(out))
	}

	m := out[0]
	if m.Kind != models.MentionEveryone {
		t.Fatalf("expected kind=%q, got %q", models.MentionEveryone, m.Kind)
	}
	if m.TargetID != "chat-1" {
		t.Fatalf("expected targetID=chat-1, got %q", m.TargetID)
	}
	if m.Text != "Everyone" {
		t.Fatalf("expected text=Everyone, got %q", m.Text)
	}
	if m.AtID != 0 {
		t.Fatalf("expected AtID=0, got %d", m.AtID)
	}
}

func TestTryAddEveryoneMention_OneOnOneChat_ReturnsErrorAndDoesNotAdd(t *testing.T) {
	out := make([]models.Mention, 0)
	adder := mentions.NewMentionAdder(&out)

	ok, err := tryAddEveryoneMention(adder, "chat-1", false, "everyone")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if ok {
		t.Fatalf("expected ok=false, got true")
	}
	if len(out) != 0 {
		t.Fatalf("expected no mentions added, got %d", len(out))
	}
}

func TestTryAddEveryoneMention_DuplicatesAllowed(t *testing.T) {
	out := make([]models.Mention, 0)
	adder := mentions.NewMentionAdder(&out)

	ok1, err1 := tryAddEveryoneMention(adder, "chat-1", true, "everyone")
	if err1 != nil || !ok1 {
		t.Fatalf("expected first add ok, got ok=%v err=%v", ok1, err1)
	}

	ok2, err2 := tryAddEveryoneMention(adder, "chat-1", true, "@everyone")
	if err2 != nil {
		t.Fatalf("unexpected error: %v", err2)
	}
	if !ok2 {
		t.Fatalf("expected ok=true (it recognized everyone), got false")
	}

	if len(out) != 2 {
		t.Fatalf("expected dedup to keep 2 mentions, got %d", len(out))
	}

	if out[0].AtID != 0 {
		t.Fatalf("expected AtID=0, got %d", out[0].AtID)
	}
}
