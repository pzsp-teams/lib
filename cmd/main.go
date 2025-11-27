package main

import (
	"context"
	"fmt"
	"os"

	"github.com/pzsp-teams/lib/pkg/teams"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: teams <command> [arguments]")
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
	default:
		fmt.Println("Unknown command:", cmd)
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
