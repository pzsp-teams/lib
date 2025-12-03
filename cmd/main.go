package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"

	"github.com/pzsp-teams/lib/pkg/teams"
	channelsPkg "github.com/pzsp-teams/lib/pkg/teams/channels"
)

func printUsage() {
	fmt.Println("Usage: teams <command> [arguments]")
	fmt.Println("\nChannel commands:")
	fmt.Println("  create-channel <team-id> <channel-name>")
	fmt.Println("  list-channels <team-id>")
	fmt.Println("  get-channel <team-id> <channel-id>")
	fmt.Println("  delete-channel <team-id> <channel-id>")
	fmt.Println("  send-message <team-id> <channel-name> <message>")
	fmt.Println("  list-messages <team-id> <channel-name> [top]")
	fmt.Println("  list-replies <team-id> <channel-name> <message-id> [top]")
	fmt.Println("\nTeam commands:")
	fmt.Println("  list-my-teams")
	fmt.Println("  get-team <team-id>")
	fmt.Println("  create-team <display-name> <mail-nickname> <visibility>")
	fmt.Println("  create-team-from-template <display-name> <description...>")
	fmt.Println("  update-team <team-id> <new-display-name> [new-description...]")
	fmt.Println("  archive-team <team-id> [spo-readonly=true|false]")
	fmt.Println("  unarchive-team <team-id>")
	fmt.Println("  delete-team <team-id>")
	fmt.Println("  restore-team <deleted-group-id>")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	cmd := os.Args[1]
	authConfig := loadAuthConfig()
	senderConfig := newSenderConfig()

	client, err := teams.NewClient(context.TODO(), authConfig, senderConfig)
	if err != nil {
		fmt.Printf("Error creating Teams client: %v\n", err)
		os.Exit(1)
	}

	switch cmd {
	case "create-channel":
		if len(os.Args) < 4 {
			fmt.Println("Usage: teams create-channel <team-id> <channel-name>")
			os.Exit(1)
		}
		handleCreateChannel(client, os.Args[2:])
	case "list-channels":
		if len(os.Args) < 3 {
			fmt.Println("Usage: teams list-channels <team-id>")
			os.Exit(1)
		}
		handleListChannels(client, os.Args[2:])
	case "get-channel":
		if len(os.Args) < 4 {
			fmt.Println("Usage: teams get-channel <team-id> <channel-id>")
			os.Exit(1)
		}
		handleGetChannel(client, os.Args[2:])
	case "delete-channel":
		if len(os.Args) < 4 {
			fmt.Println("Usage: teams delete-channel <team-id> <channel-id>")
			os.Exit(1)
		}
		handleDeleteChannel(client, os.Args[2:])
	case "send-message":
		if len(os.Args) < 5 {
			fmt.Println("Usage: teams send-message <team-id> <channel-name> <message>")
			os.Exit(1)
		}
		handleSendMessage(client, os.Args[2:])
	case "list-messages":
		if len(os.Args) < 4 {
			fmt.Println("Usage: teams list-messages <team-id> <channel-name> [top]")
			os.Exit(1)
		}
		handleListMessages(client, os.Args[2:])
	case "list-replies":
		if len(os.Args) < 5 {
			fmt.Println("Usage: teams list-replies <team-id> <channel-name> <message-id> [top]")
			os.Exit(1)
		}
		handleListReplies(client, os.Args[2:])
	case "list-my-teams":
		handleListMyTeams(client)
	case "get-team":
		if len(os.Args) < 3 {
			fmt.Println("Usage: teams get-team <team-id>")
			os.Exit(1)
		}
		handleGetTeam(client, os.Args[2:])
	case "create-team":
		if len(os.Args) < 5 {
			fmt.Println("Usage: teams create-team <display-name> <mail-nickname> <visibility>")
			os.Exit(1)
		}
		handleCreateTeam(client, os.Args[2:])
	case "create-team-from-template":
		if len(os.Args) < 4 {
			fmt.Println("Usage: teams create-team-from-template <display-name> <description...>")
			os.Exit(1)
		}
		handleCreateTeamFromTemplate(client, os.Args[2:])
	case "update-team":
		if len(os.Args) < 4 {
			fmt.Println("Usage: teams update-team <team-id> <new-display-name> [new-description...]")
			os.Exit(1)
		}
		handleUpdateTeam(client, os.Args[2:])
	case "archive-team":
		if len(os.Args) < 3 {
			fmt.Println("Usage: teams archive-team <team-id> [spo-readonly=true|false]")
			os.Exit(1)
		}
		handleArchiveTeam(client, os.Args[2:])
	case "unarchive-team":
		if len(os.Args) < 3 {
			fmt.Println("Usage: teams unarchive-team <team-id>")
			os.Exit(1)
		}
		handleUnarchiveTeam(client, os.Args[2:])
	case "delete-team":
		if len(os.Args) < 3 {
			fmt.Println("Usage: teams delete-team <team-id>")
			os.Exit(1)
		}
		handleDeleteTeam(client, os.Args[2:])
	case "restore-team":
		if len(os.Args) < 3 {
			fmt.Println("Usage: teams restore-team <deleted-group-id>")
			os.Exit(1)
		}
		handleRestoreTeam(client, os.Args[2:])

	default:
		fmt.Println("Unknown command:", cmd)
		fmt.Println()
		printUsage()
		os.Exit(1)
	}
}

func handleCreateChannel(client *teams.Client, args []string) {
	teamID := args[0]
	channelName := args[1]
	channel, err := client.Channels.Create(context.TODO(), teamID, channelName)
	if err != nil {
		fmt.Printf("Error creating channel: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Channel created with ID: %s\n", channel.ID)
}

func handleListChannels(client *teams.Client, args []string) {
	teamID := args[0]
	channels, err := client.Channels.ListChannels(context.TODO(), teamID)
	if err != nil {
		fmt.Printf("Error listing channels: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Channels:")
	for _, ch := range channels {
		fmt.Printf("- %s (ID: %s)\n", ch.Name, ch.ID)
	}
}

func handleGetChannel(client *teams.Client, args []string) {
	teamID := args[0]
	channelID := args[1]
	channel, err := client.Channels.Get(context.TODO(), teamID, channelID)
	if err != nil {
		fmt.Printf("Error getting channel: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Channel ID: %s, Name: %s, Is General: %v\n", channel.ID, channel.Name, channel.IsGeneral)
}

func handleDeleteChannel(client *teams.Client, args []string) {
	teamID := args[0]
	channelID := args[1]
	err := client.Channels.Delete(context.TODO(), teamID, channelID)
	if err != nil {
		fmt.Printf("Error deleting channel: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Channel deleted successfully.")
}

func handleSendMessage(client *teams.Client, args []string) {
	teamID := args[0]
	channelName := args[1]
	messageContent := args[2]

	channels, err := client.Channels.ListChannels(context.TODO(), teamID)
	if err != nil {
		fmt.Printf("Error listing channels: %v\n", err)
		os.Exit(1)
	}

	var channelID string
	for _, ch := range channels {
		if ch.Name == channelName {
			channelID = ch.ID
			break
		}
	}

	if channelID == "" {
		fmt.Printf("Error: Channel '%s' not found in team\n", channelName)
		fmt.Println("Available channels:")
		for _, ch := range channels {
			fmt.Printf("- %s\n", ch.Name)
		}
		os.Exit(1)
	}

	message, err := client.Channels.SendMessage(context.TODO(), teamID, channelID, channelsPkg.MessageBody{
		Content: messageContent,
	})
	if err != nil {
		fmt.Printf("Error sending message: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Message sent successfully to channel '%s' (ID: %s)\n", channelName, message.ID)
}

func handleListMessages(client *teams.Client, args []string) {
	teamID := args[0]
	channelName := args[1]

	channels, err := client.Channels.ListChannels(context.TODO(), teamID)
	if err != nil {
		fmt.Printf("Error listing channels: %v\n", err)
		os.Exit(1)
	}

	var channelID string
	for _, ch := range channels {
		if ch.Name == channelName {
			channelID = ch.ID
			break
		}
	}

	if channelID == "" {
		fmt.Printf("Error: Channel '%s' not found in team\n", channelName)
		fmt.Println("Available channels:")
		for _, ch := range channels {
			fmt.Printf("- %s\n", ch.Name)
		}
		os.Exit(1)
	}

	var opts *channelsPkg.ListMessagesOptions
	if len(args) > 2 {
		var top int32
		_, err := fmt.Sscanf(args[2], "%d", &top)
		if err != nil {
			fmt.Printf("Error: Invalid top value: %v\n", err)
			os.Exit(1)
		}
		opts = &channelsPkg.ListMessagesOptions{Top: &top}
	}

	messages, err := client.Channels.ListMessages(context.TODO(), teamID, channelID, opts)
	if err != nil {
		fmt.Printf("Error listing messages: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Messages in channel '%s':\n", channelName)
	for _, msg := range messages {
		fmt.Printf("\nID: %s\n", msg.ID)
		fmt.Printf("From: %s\n", getMessageFrom(msg))
		fmt.Printf("Created: %s\n", msg.CreatedDateTime.Format("2006-01-02 15:04:05"))
		fmt.Printf("Content: %s\n", msg.Content)
		if msg.ReplyCount > 0 {
			fmt.Printf("Replies: %d\n", msg.ReplyCount)
		}
	}
}

func handleListReplies(client *teams.Client, args []string) {
	teamID := args[0]
	channelName := args[1]
	messageID := args[2]

	channels, err := client.Channels.ListChannels(context.TODO(), teamID)
	if err != nil {
		fmt.Printf("Error listing channels: %v\n", err)
		os.Exit(1)
	}

	var channelID string
	for _, ch := range channels {
		if ch.Name == channelName {
			channelID = ch.ID
			break
		}
	}

	if channelID == "" {
		fmt.Printf("Error: Channel '%s' not found in team\n", channelName)
		fmt.Println("Available channels:")
		for _, ch := range channels {
			fmt.Printf("- %s\n", ch.Name)
		}
		os.Exit(1)
	}

	var top *int32
	if len(args) > 3 {
		var topVal int32
		_, err := fmt.Sscanf(args[3], "%d", &topVal)
		if err != nil {
			fmt.Printf("Error: Invalid top value: %v\n", err)
			os.Exit(1)
		}
		top = &topVal
	}

	replies, err := client.Channels.ListReplies(context.TODO(), teamID, channelID, messageID, top)
	if err != nil {
		fmt.Printf("Error listing replies: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Replies to message %s:\n", messageID)
	for _, reply := range replies {
		fmt.Printf("\nID: %s\n", reply.ID)
		fmt.Printf("From: %s\n", getMessageFrom(reply))
		fmt.Printf("Created: %s\n", reply.CreatedDateTime.Format("2006-01-02 15:04:05"))
		fmt.Printf("Content: %s\n", reply.Content)
	}
}

func getMessageFrom(msg *channelsPkg.Message) string {
	if msg.From != nil {
		if msg.From.DisplayName != "" {
			return msg.From.DisplayName
		}
		return msg.From.UserID
	}
	return "Unknown"
}

func handleListMyTeams(client *teams.Client) {
	ctx := context.TODO()
	ts, err := client.Teams.ListMyJoined(ctx)
	if err != nil {
		fmt.Printf("Error listing teams: %v\n", err)
		os.Exit(1)
	}
	if len(ts) == 0 {
		fmt.Println("You are not a member of any teams.")
		return
	}

	fmt.Println("Teams:")
	for _, t := range ts {
		state := ""
		if t.IsArchived {
			state = " (archived)"
		}
		fmt.Printf("- %s (ID: %s)%s\n", t.DisplayName, t.ID, state)
	}
}

func handleGetTeam(client *teams.Client, args []string) {
	teamID := args[0]
	ctx := context.TODO()
	t, err := client.Teams.Get(ctx, teamID)
	if err != nil {
		fmt.Printf("Error getting team: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Team ID: %s\nName: %s\nDescription: %s\nVisibility: %s\nArchived: %v\n",
		t.ID, t.DisplayName, t.Description, t.Visibility, t.IsArchived)
}

func handleCreateTeam(client *teams.Client, args []string) {
	displayName := args[0]
	mailNickname := args[1]
	visibility := args[2] 

	ctx := context.TODO()
	t, err := client.Teams.CreateViaGroup(ctx, displayName, mailNickname, visibility)
	if err != nil {
		fmt.Printf("Error creating team: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Team created: %s (ID: %s)\n", t.DisplayName, t.ID)
}

func handleCreateTeamFromTemplate(client *teams.Client, args []string) {
	displayName := args[0]
	description := strings.Join(args[1:], " ")

	ctx := context.TODO()
	id, err := client.Teams.CreateFromTemplate(ctx, displayName, description, nil)
	if err != nil {
		fmt.Printf("Error creating team from template: %v\n", err)
		os.Exit(1)
	}
	if id == "" {
		fmt.Println("Team creation started (async). It may take some time before the team is available.")
	} else {
		fmt.Printf("Team creation started. Team (group) ID: %s\n", id)
	}
}

func handleUpdateTeam(client *teams.Client, args []string) {
	teamID := args[0]
	newName := args[1]

	var newDesc *string
	if len(args) > 2 {
		desc := strings.Join(args[2:], " ")
		newDesc = &desc
	}

	patch := msmodels.NewTeam()
	patch.SetDisplayName(&newName)
	if newDesc != nil {
		patch.SetDescription(newDesc)
	}

	ctx := context.TODO()
	t, err := client.Teams.Update(ctx, teamID, patch)
	if err != nil {
		fmt.Printf("Error updating team: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Team updated: %s (ID: %s)\n", t.DisplayName, t.ID)
}

func handleArchiveTeam(client *teams.Client, args []string) {
	teamID := args[0]
	var spo *bool
	if len(args) > 1 {
		val, err := strconv.ParseBool(args[1])
		if err != nil {
			fmt.Printf("Error: invalid spo-readonly value (expected true/false): %v\n", err)
			os.Exit(1)
		}
		spo = &val
	}

	ctx := context.TODO()
	if err := client.Teams.Archive(ctx, teamID, spo); err != nil {
		fmt.Printf("Error archiving team: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Team archived.")
}

func handleUnarchiveTeam(client *teams.Client, args []string) {
	teamID := args[0]
	ctx := context.TODO()
	if err := client.Teams.Unarchive(ctx, teamID); err != nil {
		fmt.Printf("Error unarchiving team: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Team unarchived.")
}

func handleDeleteTeam(client *teams.Client, args []string) {
	teamID := args[0]
	ctx := context.TODO()
	if err := client.Teams.Delete(ctx, teamID); err != nil {
		fmt.Printf("Error deleting team: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Team delete request sent (soft-delete in M365 Group).")
}

func handleRestoreTeam(client *teams.Client, args []string) {
	deletedGroupID := args[0]
	ctx := context.TODO()
	id, err := client.Teams.RestoreDeleted(ctx, deletedGroupID)
	if err != nil {
		fmt.Printf("Error restoring team: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Team restored. Directory object ID: %s\n", id)
}
