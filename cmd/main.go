package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"

	lib "github.com/pzsp-teams/lib"
	channelsPkg "github.com/pzsp-teams/lib/channels"
)

func printUsage() {
	fmt.Println("Usage: teams <command> [arguments]")
	fmt.Println("\nChannel commands:")
	fmt.Println("  create-channel <team-name> <channel-name>")
	fmt.Println("  create-private-channel <team-name> <channel-name> <members> [owners]")
	fmt.Println("      members / owners: comma-separated user ids or emails")
	fmt.Println("  list-channels <team-name>")
	fmt.Println("  get-channel <team-name> <channel-name>")
	fmt.Println("  delete-channel <team-name> <channel-name>")
	fmt.Println("  send-message <team-name> <channel-name> <message>")
	fmt.Println("  list-messages <team-name> <channel-name> [top]")
	fmt.Println("  list-replies <team-name> <channel-name> <message-id> [top]")
	fmt.Println("\nTeam commands:")
	fmt.Println("  list-my-teams")
	fmt.Println("  get-team <team-name>")
	fmt.Println("  create-team <display-name>")
	fmt.Println("  create-team-from-template <display-name> <description...>")
	fmt.Println("  update-team <team-name> <new-display-name> [new-description...]")
	fmt.Println("  archive-team <team-name> [spo-readonly=true|false]")
	fmt.Println("  unarchive-team <team-name>")
	fmt.Println("  delete-team <team-name>")
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

	client, err := lib.NewClient(context.TODO(), authConfig, senderConfig)
	if err != nil {
		fmt.Printf("Error creating Teams client: %v\n", err)
		os.Exit(1)
	}

	switch cmd {
	case "create-channel":
		if len(os.Args) < 4 {
			fmt.Println("Usage: teams create-channel <team-name> <channel-name>")
			os.Exit(1)
		}
		handleCreateChannel(client, os.Args[2:])

	case "create-private-channel":
		if len(os.Args) < 5 {
			fmt.Println("Usage: teams create-private-channel <team-name> <channel-name> <members> [owners]")
			fmt.Println("       members / owners: comma-separated user ids or emails")
			os.Exit(1)
		}
		handleCreatePrivateChannel(client, os.Args[2:])

	case "list-channels":
		if len(os.Args) < 3 {
			fmt.Println("Usage: teams list-channels <team-name>")
			os.Exit(1)
		}
		handleListChannels(client, os.Args[2:])

	case "get-channel":
		if len(os.Args) < 4 {
			fmt.Println("Usage: teams get-channel <team-name> <channel-name>")
			os.Exit(1)
		}
		handleGetChannel(client, os.Args[2:])

	case "delete-channel":
		if len(os.Args) < 4 {
			fmt.Println("Usage: teams delete-channel <team-name> <channel-name>")
			os.Exit(1)
		}
		handleDeleteChannel(client, os.Args[2:])

	case "send-message":
		if len(os.Args) < 5 {
			fmt.Println("Usage: teams send-message <team-name> <channel-name> <message>")
			os.Exit(1)
		}
		handleSendMessage(client, os.Args[2:])

	case "list-messages":
		if len(os.Args) < 4 {
			fmt.Println("Usage: teams list-messages <team-name> <channel-name> [top]")
			os.Exit(1)
		}
		handleListMessages(client, os.Args[2:])

	case "list-replies":
		if len(os.Args) < 5 {
			fmt.Println("Usage: teams list-replies <team-name> <channel-name> <message-id> [top]")
			os.Exit(1)
		}
		handleListReplies(client, os.Args[2:])

	case "list-my-teams":
		handleListMyTeams(client)

	case "get-team":
		if len(os.Args) < 3 {
			fmt.Println("Usage: teams get-team <team-name>")
			os.Exit(1)
		}
		handleGetTeam(client, os.Args[2:])

	case "create-team":
		if len(os.Args) < 3 {
			fmt.Println("Usage: teams create-team <display-name>")
			os.Exit(1)
		}
		handleCreateTeam(client, os.Args[2:])

	case "create-team-from-template":
		if len(os.Args) < 3 {
			fmt.Println("Usage: teams create-team-from-template <display-name> <description...>")
			os.Exit(1)
		}
		handleCreateTeamFromTemplate(client, os.Args[2:])

	case "update-team":
		if len(os.Args) < 4 {
			fmt.Println("Usage: teams update-team <team-name> <new-display-name> [new-description...]")
			os.Exit(1)
		}
		handleUpdateTeam(client, os.Args[2:])

	case "archive-team":
		if len(os.Args) < 3 {
			fmt.Println("Usage: teams archive-team <team-name> [spo-readonly=true|false]")
			os.Exit(1)
		}
		handleArchiveTeam(client, os.Args[2:])

	case "unarchive-team":
		if len(os.Args) < 3 {
			fmt.Println("Usage: teams unarchive-team <team-name>")
			os.Exit(1)
		}
		handleUnarchiveTeam(client, os.Args[2:])

	case "delete-team":
		if len(os.Args) < 3 {
			fmt.Println("Usage: teams delete-team <team-name>")
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

func handleCreateChannel(client *lib.Client, args []string) {
	teamName := args[0]
	channelName := args[1]

	channel, err := client.Channels.CreateStandardChannel(context.TODO(), teamName, channelName)
	if err != nil {
		fmt.Printf("Error creating channel: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Channel created with ID: %s\n", channel.ID)
}

func handleCreatePrivateChannel(client *lib.Client, args []string) {
	teamName := args[0]
	channelName := args[1]
	rawMembers := args[2]

	memberRefs := parseUserList(rawMembers)
	if len(memberRefs) == 0 {
		fmt.Println("Error: at least one member must be specified (comma-separated list of user ids/emails)")
		os.Exit(1)
	}

	var ownerRefs []string
	if len(args) > 3 {
		ownerRefs = parseUserList(args[3])
	}

	channel, err := client.Channels.CreatePrivateChannel(context.TODO(), teamName, channelName, memberRefs, ownerRefs)
	if err != nil {
		fmt.Printf("Error creating private channel: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Private channel created: %s (ID: %s)\n", channel.Name, channel.ID)
	fmt.Printf("Members: %v\n", memberRefs)
	if len(ownerRefs) > 0 {
		fmt.Printf("Owners: %v\n", ownerRefs)
	}
}

func parseUserList(arg string) []string {
	parts := strings.Split(arg, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func handleListChannels(client *lib.Client, args []string) {
	teamName := args[0]

	channels, err := client.Channels.ListChannels(context.TODO(), teamName)
	if err != nil {
		fmt.Printf("Error listing channels: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Channels:")
	for _, ch := range channels {
		fmt.Printf("- %s (ID: %s)\n", ch.Name, ch.ID)
	}
}

func handleGetChannel(client *lib.Client, args []string) {
	teamName := args[0]
	channelName := args[1]

	channel, err := client.Channels.Get(context.TODO(), teamName, channelName)
	if err != nil {
		fmt.Printf("Error getting channel: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Channel ID: %s, Name: %s, Is General: %v\n", channel.ID, channel.Name, channel.IsGeneral)
}

func handleDeleteChannel(client *lib.Client, args []string) {
	teamName := args[0]
	channelName := args[1]

	err := client.Channels.Delete(context.TODO(), teamName, channelName)
	if err != nil {
		fmt.Printf("Error deleting channel: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Channel deleted successfully.")
}

func handleSendMessage(client *lib.Client, args []string) {
	teamName := args[0]
	channelName := args[1]
	messageContent := args[2]

	message, err := client.Channels.SendMessage(context.TODO(), teamName, channelName, channelsPkg.MessageBody{
		Content: messageContent,
	})
	if err != nil {
		fmt.Printf("Error sending message: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Message sent successfully to channel '%s' (ID: %s)\n", channelName, message.ID)
}

func handleListMessages(client *lib.Client, args []string) {
	teamName := args[0]
	channelName := args[1]

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

	messages, err := client.Channels.ListMessages(context.TODO(), teamName, channelName, opts)
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

func handleListReplies(client *lib.Client, args []string) {
	teamName := args[0]
	channelName := args[1]
	messageID := args[2]

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

	replies, err := client.Channels.ListReplies(context.TODO(), teamName, channelName, messageID, top)
	if err != nil {
		fmt.Printf("Error listing replies: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Replies to message %s in channel '%s':\n", messageID, channelName)
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

func handleListMyTeams(client *lib.Client) {
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

func handleGetTeam(client *lib.Client, args []string) {
	teamName := args[0]
	ctx := context.TODO()

	t, err := client.Teams.Get(ctx, teamName)
	if err != nil {
		fmt.Printf("Error getting team: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Team ID: %s\nName: %s\nDescription: %s\nVisibility: %s\nArchived: %v\n",
		t.ID, t.DisplayName, t.Description, t.Visibility, t.IsArchived)
}

func handleCreateTeam(client *lib.Client, args []string) {
	displayName := args[0]
	mailNickname := displayName
	visibility := "public"

	ctx := context.TODO()
	t, err := client.Teams.CreateViaGroup(ctx, displayName, mailNickname, visibility)
	if err != nil {
		fmt.Printf("Error creating team: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Team created: %s (ID: %s)\n", t.DisplayName, t.ID)
}

func handleCreateTeamFromTemplate(client *lib.Client, args []string) {
	displayName := args[0]
	description := ""
	if len(args) > 1 {
		description = strings.Join(args[1:], " ")
	}

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

func handleUpdateTeam(client *lib.Client, args []string) {
	teamName := args[0]
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
	t, err := client.Teams.Update(ctx, teamName, patch)
	if err != nil {
		fmt.Printf("Error updating team: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Team updated: %s (ID: %s)\n", t.DisplayName, t.ID)
}

func handleArchiveTeam(client *lib.Client, args []string) {
	teamName := args[0]
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
	if err := client.Teams.Archive(ctx, teamName, spo); err != nil {
		fmt.Printf("Error archiving team: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Team archived.")
}

func handleUnarchiveTeam(client *lib.Client, args []string) {
	teamName := args[0]
	ctx := context.TODO()
	if err := client.Teams.Unarchive(ctx, teamName); err != nil {
		fmt.Printf("Error unarchiving team: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Team unarchived.")
}

func handleDeleteTeam(client *lib.Client, args []string) {
	teamName := args[0]
	ctx := context.TODO()
	if err := client.Teams.Delete(ctx, teamName); err != nil {
		fmt.Printf("Error deleting team: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Team delete request sent (soft-delete in M365 Group).")
}

func handleRestoreTeam(client *lib.Client, args []string) {
	deletedGroupID := args[0]
	ctx := context.TODO()
	id, err := client.Teams.RestoreDeleted(ctx, deletedGroupID)
	if err != nil {
		fmt.Printf("Error restoring team: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Team restored. Directory object ID: %s\n", id)
}
