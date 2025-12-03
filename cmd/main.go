package main

import (
	"context"
	"fmt"
	"os"

	"github.com/pzsp-teams/lib/pkg/teams"
	channelsPkg "github.com/pzsp-teams/lib/pkg/teams/channels"
)

func printUsage() {
	fmt.Println("Usage: teams <command> [arguments]")
	fmt.Println("\nAvailable commands:")
	fmt.Println("  create-channel <team-id> <channel-name>")
	fmt.Println("  list-channels <team-id>")
	fmt.Println("  get-channel <team-id> <channel-id>")
	fmt.Println("  delete-channel <team-id> <channel-id>")
	fmt.Println("  send-message <team-id> <channel-name> <message>")
	fmt.Println("  list-messages <team-id> <channel-name> [top]")
	fmt.Println("  list-replies <team-id> <channel-name> <message-id> [top]")
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
	default:
		fmt.Println("Unknown command:", cmd)
		fmt.Println()
		printUsage()
		os.Exit(1)
	}
}

func handleCreateChannel(client *teams.Client, args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: teams create-channel <team-id> <channel-name>")
		os.Exit(1)
	}
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
	if len(args) < 1 {
		fmt.Println("Usage: teams list-channels <team-id>")
		os.Exit(1)
	}
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
	if len(args) < 2 {
		fmt.Println("Usage: teams get-channel <team-id> <channel-id>")
		os.Exit(1)
	}
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
	if len(args) < 2 {
		fmt.Println("Usage: teams delete-channel <team-id> <channel-id>")
		os.Exit(1)
	}
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
	if len(args) < 3 {
		fmt.Println("Usage: teams send-message <team-id> <channel-name> <message>")
		os.Exit(1)
	}
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
	if len(args) < 2 {
		fmt.Println("Usage: teams list-messages <team-id> <channel-name> [top]")
		os.Exit(1)
	}
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
	if len(args) < 3 {
		fmt.Println("Usage: teams list-replies <team-id> <channel-name> <message-id> [top]")
		os.Exit(1)
	}
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
