package mentions

import (
	"fmt"
	"strings"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/internal/util"
	"github.com/pzsp-teams/lib/models"
)

func MapMentions(in []models.Mention) (mentions []msmodels.ChatMessageMentionable, err error) {
	if len(in) == 0 {
		return nil, nil
	}
	out := make([]msmodels.ChatMessageMentionable, 0, len(in))
	for i, mention := range in {
		mapped, mapErr := mapToGraphMention(mention)
		if mapErr != nil {
			return nil, fmt.Errorf("mention[%d]: %w", i, mapErr)
		}
		out = append(out, mapped)
	}
	return out, nil
}

func ValidateAtTags(body *models.MessageBody) error {
	if body == nil {
		return nil
	}
	if body.ContentType != models.MessageContentTypeHTML {
		return nil
	}
	if strings.Contains(body.Content, "<at") && len(body.Mentions) == 0 {
		return fmt.Errorf("content contains <at> tags but mentions list is empty")
	}
	seen := map[int32]struct{}{}
	for _, mention := range body.Mentions {
		if _, exists := seen[mention.AtID]; exists {
			return fmt.Errorf("duplicate at-id %d in mentions", mention.AtID)
		}
		seen[mention.AtID] = struct{}{}
	}
	for _, mention := range body.Mentions {
		atTag := fmt.Sprintf(`<at id="%d"`, mention.AtID)
		if !strings.Contains(body.Content, atTag) {
			return fmt.Errorf("missing %s in content", atTag)
		}
	}
	return nil
}

func mapToGraphMention(mention models.Mention) (msmodels.ChatMessageMentionable, error) {
	if mention.Text == "" {
		return nil, fmt.Errorf("missing Text")
	}
	if mention.TargetID == "" {
		return nil, fmt.Errorf("missing TargetID")
	}
	if mention.AtID < 0 {
		return nil, fmt.Errorf("invalid AtID: %d", mention.AtID)
	}
	txt := mention.Text
	atID := mention.AtID
	targetID := mention.TargetID
	graphMention := msmodels.NewChatMessageMention()
	graphMention.SetId(&atID)
	graphMention.SetMentionText(&txt)

	mentioned := msmodels.NewChatMessageMentionedIdentitySet()
	switch mention.Kind {
	case models.MentionUser:
		user := msmodels.NewIdentity()
		if !util.IsLikelyGUID(targetID) {
			return nil, fmt.Errorf("invalid TargetID for user mention: %s", targetID)
		}
		user.SetId(&targetID)
		user.SetDisplayName(&txt)
		user.SetAdditionalData(map[string]any{
			"userIdentityType": "aadUser",
		})
		mentioned.SetUser(user)
	case models.MentionChannel:
		conv := msmodels.NewTeamworkConversationIdentity()
		if !util.IsLikelyThreadConversationID(targetID) {
			return nil, fmt.Errorf("invalid TargetID for channel mention: %s", targetID)
		}
		conv.SetId(&targetID)
		conv.SetDisplayName(&txt)
		t := msmodels.CHANNEL_TEAMWORKCONVERSATIONIDENTITYTYPE
		conv.SetConversationIdentityType(&t)
		mentioned.SetConversation(conv)
	case models.MentionTeam:
		conv := msmodels.NewTeamworkConversationIdentity()
		if !util.IsLikelyGUID(targetID) {
			return nil, fmt.Errorf("invalid TargetID for team mention: %s", targetID)
		}
		conv.SetId(&targetID)
		conv.SetDisplayName(&txt)
		t := msmodels.TEAM_TEAMWORKCONVERSATIONIDENTITYTYPE
		conv.SetConversationIdentityType(&t)
		mentioned.SetConversation(conv)
	case models.MentionEveryone:
		conv := msmodels.NewTeamworkConversationIdentity()
		if !util.IsLikelyThreadConversationID(targetID) {
			return nil, fmt.Errorf("invalid TargetID for everyone mention (chatID): %s", targetID)
		}
		conv.SetId(&targetID)
		conv.SetDisplayName(&txt)
		t := msmodels.CHAT_TEAMWORKCONVERSATIONIDENTITYTYPE
		conv.SetConversationIdentityType(&t)
		mentioned.SetConversation(conv)
	default:
		return nil, fmt.Errorf("unknown Mention.Kind: %s", mention.Kind)
	}

	graphMention.SetMentioned(mentioned)

	return graphMention, nil
}