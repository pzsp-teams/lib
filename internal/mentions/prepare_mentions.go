package mentions

import (
	"fmt"
	"strings"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pzsp-teams/lib/internal/util"
	"github.com/pzsp-teams/lib/models"
)

type convCfg struct {
	identityType msmodels.TeamworkConversationIdentityType
	validate     func(string) bool
	label        string
}

var convKinds = map[models.MentionKind]convCfg{
	models.MentionChannel:  {msmodels.CHANNEL_TEAMWORKCONVERSATIONIDENTITYTYPE, util.IsLikelyThreadConversationID, "channel mention"},
	models.MentionTeam:     {msmodels.TEAM_TEAMWORKCONVERSATIONIDENTITYTYPE, util.IsLikelyGUID, "team mention"},
	models.MentionEveryone: {msmodels.CHAT_TEAMWORKCONVERSATIONIDENTITYTYPE, util.IsLikelyThreadConversationID, "everyone mention (chatID)"},
}

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

func mapToGraphMention(m models.Mention) (msmodels.ChatMessageMentionable, error) {
	if m.Text == "" {
		return nil, fmt.Errorf("missing Text")
	}
	if m.AtID < 0 {
		return nil, fmt.Errorf("invalid AtID: %d", m.AtID)
	}
	if m.TargetID == "" {
		return nil, fmt.Errorf("missing TargetID")
	}

	txt, atID, targetID := m.Text, m.AtID, m.TargetID

	graphMention := msmodels.NewChatMessageMention()
	graphMention.SetId(&atID)
	graphMention.SetMentionText(&txt)

	mentioned := msmodels.NewChatMessageMentionedIdentitySet()

	if m.Kind == models.MentionUser {
		userMentioned, err := buildUserMentioned(targetID, txt)
		if err != nil {
			return nil, err
		}
		mentioned.SetUser(userMentioned)
	} else if convCfg, ok := convKinds[m.Kind]; ok {
		convMentioned, err := buildConversationMentioned(
			targetID,
			txt,
			convCfg.identityType,
			convCfg.validate,
			convCfg.label,
		)
		if err != nil {
			return nil, err
		}
		mentioned.SetConversation(convMentioned)
	} else {
		return nil, fmt.Errorf("unsupported MentionKind: %s", m.Kind)
	}

	graphMention.SetMentioned(mentioned)
	return graphMention, nil
}

func buildUserMentioned(targetID, displayName string) (msmodels.Identityable, error) {
	if !util.IsLikelyGUID(targetID) {
		return nil, fmt.Errorf("invalid TargetID for user mention: %s", targetID)
	}
	user := msmodels.NewIdentity()
	user.SetId(&targetID)
	user.SetDisplayName(&displayName)
	user.SetAdditionalData(map[string]any{
		"userIdentityType": "aadUser",
	})
	return user, nil
}

func buildConversationMentioned(
	targetID, displayName string,
	typ msmodels.TeamworkConversationIdentityType,
	validate func(string) bool,
	label string,
) (msmodels.TeamworkConversationIdentityable, error) {
	if !validate(targetID) {
		return nil, fmt.Errorf("invalid TargetID for %s: %s", label, targetID)
	}
	conv := msmodels.NewTeamworkConversationIdentity()
	conv.SetId(&targetID)
	conv.SetDisplayName(&displayName)
	conv.SetConversationIdentityType(&typ)
	return conv, nil
}

func PrepareMentions(body *models.MessageBody) ([]msmodels.ChatMessageMentionable, error) {
	if len(body.Mentions) > 0 && body.ContentType != models.MessageContentTypeHTML {
		return nil, fmt.Errorf("mentions can only be used with HTML content type")
	}
	if err := ValidateAtTags(body); err != nil {
		return nil, err
	}
	return MapMentions(body.Mentions)
}
