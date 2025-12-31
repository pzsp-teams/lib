package channels

import (
	"testing"

	"github.com/pzsp-teams/lib/internal/mentions"
	"github.com/pzsp-teams/lib/models"
)

func TestIsTeamRef(t *testing.T) {
	teamRef := "team-A"
	teamID := "tid-1"

	cases := []struct {
		name string
		low  string
		raw  string
		want bool
	}{
		{"keyword team", "team", "team", true},
		{"match by ref", "team-a", teamRef, true},
		{"match by id", "tid-1", teamID, true},
		{"not match", "x", "x", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := isTeamRef(tc.low, tc.raw, teamRef, teamID)
			if got != tc.want {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestIsChannelRef(t *testing.T) {
	channelRef := "General"
	channelID := "cid-1"

	cases := []struct {
		name string
		low  string
		raw  string
		want bool
	}{
		{"keyword channel", "channel", "channel", true},
		{"match by ref", "general", channelRef, true},
		{"match by id", "cid-1", channelID, true},
		{"not match", "x", "x", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := isChannelRef(tc.low, tc.raw, channelRef, channelID)
			if got != tc.want {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestTryAddTeamOrChannelMention_AddsTeamByKeyword(t *testing.T) {
	teamRef := "team-A"
	teamID := "tid-1"
	channelRef := "General"
	channelID := "cid-1"

	out := make([]models.Mention, 0)
	adder := mentions.NewMentionAdder(&out)

	ok := tryAddTeamOrChannelMention(adder, "TEAM", teamRef, teamID, channelRef, channelID)
	if !ok {
		t.Fatalf("expected ok=true")
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 mention, got %d", len(out))
	}
	m := out[0]
	if m.Kind != models.MentionTeam {
		t.Fatalf("expected kind=%q, got %q", models.MentionTeam, m.Kind)
	}
	if m.TargetID != teamID {
		t.Fatalf("expected targetID=%q, got %q", teamID, m.TargetID)
	}
	if m.Text != teamRef {
		t.Fatalf("expected text=%q, got %q", teamRef, m.Text)
	}
	if m.AtID != 0 {
		t.Fatalf("expected AtID=0, got %d", m.AtID)
	}
}

func TestTryAddTeamOrChannelMention_AddsChannelByKeyword(t *testing.T) {
	teamRef := "team-A"
	teamID := "tid-1"
	channelRef := "General"
	channelID := "cid-1"

	out := make([]models.Mention, 0)
	adder := mentions.NewMentionAdder(&out)

	ok := tryAddTeamOrChannelMention(adder, "channel", teamRef, teamID, channelRef, channelID)
	if !ok {
		t.Fatalf("expected ok=true")
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 mention, got %d", len(out))
	}
	m := out[0]
	if m.Kind != models.MentionChannel {
		t.Fatalf("expected kind=%q, got %q", models.MentionChannel, m.Kind)
	}
	if m.TargetID != channelID {
		t.Fatalf("expected targetID=%q, got %q", channelID, m.TargetID)
	}
	if m.Text != channelRef {
		t.Fatalf("expected text=%q, got %q", channelRef, m.Text)
	}
}

func TestTryAddTeamOrChannelMention_TeamTakesPrecedenceOverChannel(t *testing.T) {
	teamRef := "my-team"
	teamID := "tid-1"
	channelRef := "team"
	channelID := "cid-1"

	out := make([]models.Mention, 0)
	adder := mentions.NewMentionAdder(&out)

	ok := tryAddTeamOrChannelMention(adder, "team", teamRef, teamID, channelRef, channelID)
	if !ok {
		t.Fatalf("expected ok=true")
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 mention, got %d", len(out))
	}
	if out[0].Kind != models.MentionTeam {
		t.Fatalf("expected TEAM mention precedence, got %q", out[0].Kind)
	}
}

func TestTryAddTeamOrChannelMention_ReturnsFalseAndDoesNotAdd(t *testing.T) {
	teamRef := "team-A"
	teamID := "tid-1"
	channelRef := "General"
	channelID := "cid-1"

	out := make([]models.Mention, 0)
	adder := mentions.NewMentionAdder(&out)

	ok := tryAddTeamOrChannelMention(adder, "random", teamRef, teamID, channelRef, channelID)
	if ok {
		t.Fatalf("expected ok=false")
	}
	if len(out) != 0 {
		t.Fatalf("expected no mentions added, got %d", len(out))
	}
}

func TestMemberRole(t *testing.T) {
	if got := memberRole(true); got != "owner" {
		t.Fatalf("expected owner, got %q", got)
	}
	if got := memberRole(false); got != "member" {
		t.Fatalf("expected member, got %q", got)
	}
}
