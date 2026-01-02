package channels

import (
	"testing"

	"github.com/pzsp-teams/lib/internal/mentions"
	"github.com/pzsp-teams/lib/models"
	"github.com/stretchr/testify/assert"
)

func TestIsTeamRef(t *testing.T) {
	teamRef := "team-A"
	teamID := "tid-1"

	tests := []struct {
		name string
		low  string
		raw  string
		want bool
	}{
		{
			name: "keyword team",
			low:  "team",
			raw:  "team",
			want: true,
		},
		{
			name: "match by ref",
			low:  "team-a",
			raw:  teamRef,
			want: true,
		},
		{
			name: "match by id",
			low:  "tid-1",
			raw:  teamID,
			want: true,
		},
		{
			name: "not match",
			low:  "x",
			raw:  "x",
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := isTeamRef(tc.low, tc.raw, teamRef, teamID)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestIsChannelRef(t *testing.T) {
	channelRef := "General"
	channelID := "cid-1"

	tests := []struct {
		name string
		low  string
		raw  string
		want bool
	}{
		{
			name: "keyword channel",
			low:  "channel",
			raw:  "channel",
			want: true,
		},
		{
			name: "match by ref",
			low:  "general",
			raw:  channelRef,
			want: true,
		},
		{
			name: "match by id",
			low:  "cid-1",
			raw:  channelID,
			want: true,
		},
		{
			name: "not match",
			low:  "x",
			raw:  "x",
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := isChannelRef(tc.low, tc.raw, channelRef, channelID)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestTryAddTeamOrChannelMention(t *testing.T) {
	const (
		defaultTeamRef    = "team-A"
		defaultTeamID     = "tid-1"
		defaultChannelRef = "General"
		defaultChannelID  = "cid-1"
	)

	tests := []struct {
		name         string
		token        string
		teamRef      string 
		channelRef   string
		wantOK       bool
		wantMentions []models.Mention
	}{
		{
			name:  "Adds Team by keyword (uppercase input)",
			token: "TEAM",
			wantOK: true,
			wantMentions: []models.Mention{
				{
					Kind:     models.MentionTeam,
					TargetID: defaultTeamID,
					Text:     defaultTeamRef,
					AtID:     0,
				},
			},
		},
		{
			name:  "Adds Channel by keyword (lowercase input)",
			token: "channel",
			wantOK: true,
			wantMentions: []models.Mention{
				{
					Kind:     models.MentionChannel,
					TargetID: defaultChannelID,
					Text:     defaultChannelRef,
				},
			},
		},
		{
			name:       "Team takes precedence over Channel when names collide",
			token:      "team",
			teamRef:    "my-team",
			channelRef: "team",
			wantOK:     true,
			wantMentions: []models.Mention{
				{
					Kind:     models.MentionTeam,
					TargetID: defaultTeamID,
					Text:     "my-team",
				},
			},
		},
		{
			name:         "Returns false and does not add mention when no match",
			token:        "random",
			wantOK:       false,
			wantMentions: []models.Mention{}, 
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tRef := defaultTeamRef
			if tc.teamRef != "" {
				tRef = tc.teamRef
			}
			cRef := defaultChannelRef
			if tc.channelRef != "" {
				cRef = tc.channelRef
			}

			out := make([]models.Mention, 0)
			adder := mentions.NewMentionAdder(&out)

			gotOK := tryAddTeamOrChannelMention(adder, tc.token, tRef, defaultTeamID, cRef, defaultChannelID)

			assert.Equal(t, tc.wantOK, gotOK, "Incorrect boolean return value")
			assert.Equal(t, tc.wantMentions, out, "Mentions slice mismatch")
		})
	}
}