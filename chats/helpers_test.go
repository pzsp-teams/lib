package chats

import (
	"testing"

	"github.com/pzsp-teams/lib/internal/mentions"
	"github.com/pzsp-teams/lib/models"
	"github.com/stretchr/testify/assert"
)


type unknownChatRef struct{}

func (u unknownChatRef) chatRef() {}

func TestIsGroupChatRef(t *testing.T) {

	tests := []struct {
		name      string
		chatRef   any
		wantGroup bool
		wantErr   bool
	}{
		{
			name:      "Group chat returns true",
			chatRef:   GroupChatRef{Ref: "x"},
			wantGroup: true,
			wantErr:   false,
		},
		{
			name:      "OneOnOne chat returns false",
			chatRef:   OneOnOneChatRef{Ref: "x"},
			wantGroup: false,
			wantErr:   false,
		},
		{
			name:      "Unknown chat ref returns error",
			chatRef:   unknownChatRef{},
			wantGroup: false,
			wantErr:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := isGroupChatRef(tc.chatRef.(ChatRef))

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantGroup, got)
			}
		})
	}
}

func TestTryAddEveryoneMention(t *testing.T) {
	const chatID = "chat-1"

	tests := []struct {
		name         string
		isGroup      bool
		inputs       []string 
		wantOK       bool
		wantErr      bool
		wantMentions []models.Mention
	}{
		{
			name:    "Not everyone mention returns false and no error",
			isGroup: true,
			inputs:  []string{"alice@example.com"},
			wantOK:  false,
			wantErr: false,
			wantMentions: []models.Mention{},
		},
		{
			name:    "Group chat adds everyone mention",
			isGroup: true,
			inputs:  []string{"  @Everyone  "}, 
			wantOK:  true,
			wantErr: false,
			wantMentions: []models.Mention{
				{
					Kind:     models.MentionEveryone,
					TargetID: chatID,
					Text:     "Everyone",
					AtID:     0,
				},
			},
		},
		{
			name:    "OneOnOne chat returns error for everyone mention",
			isGroup: false,
			inputs:  []string{"everyone"},
			wantOK:  false,
			wantErr: true,
			wantMentions: []models.Mention{},
		},
		{
			name:    "Duplicates allowed (adds multiple)",
			isGroup: true,
			inputs:  []string{"everyone", "@everyone"},
			wantOK:  true,
			wantErr: false,
			wantMentions: []models.Mention{
				{Kind: models.MentionEveryone, TargetID: chatID, Text: "Everyone"},
				{Kind: models.MentionEveryone, TargetID: chatID, Text: "Everyone", AtID: 1},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out := make([]models.Mention, 0)
			adder := mentions.NewMentionAdder(&out)

			var lastOK bool
			var lastErr error

			for _, token := range tc.inputs {
				lastOK, lastErr = tryAddEveryoneMention(adder, chatID, tc.isGroup, token)
				
				if lastErr != nil {
					break
				}
			}

			if tc.wantErr {
				assert.Error(t, lastErr)
			} else {
				assert.NoError(t, lastErr)
				assert.Equal(t, tc.wantOK, lastOK, "Unexpected boolean result (ok)")
			}

			assert.Equal(t, len(tc.wantMentions), len(out), "Number of mentions mismatch")
			for i, wantM := range tc.wantMentions {
				gotM := out[i]
				assert.Equal(t, wantM.Kind, gotM.Kind)
				assert.Equal(t, wantM.TargetID, gotM.TargetID)
				assert.Equal(t, wantM.Text, gotM.Text)
				assert.Equal(t, wantM.AtID, gotM.AtID)
			}
		})
	}
}

func TestCheckThisMentionValidity(t *testing.T) {
	tests := []struct {
		name    string
		isGroup bool
		token   string
		want    bool
	}{
		{"OneOnOne with 'this'", false, "this", true},
		{"OneOnOne with '@this'", false, "@this", true},
		{"OneOnOne with other name", false, "alice", false},
		{"OneOnOne with other handle", false, "@alice", false},
		
		{"GroupChat with 'this'", true, "this", false},
		{"GroupChat with '@this'", true, "@this", false},
		{"GroupChat with other name", true, "alice", false},
		{"GroupChat with other handle", true, "@alice", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := checkThisMentionValidity(tc.isGroup, tc.token)
			assert.Equal(t, tc.want, got)
		})
	}
}