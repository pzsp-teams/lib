# Teams API wrapper Lib

[![Go Reference](https://pkg.go.dev/badge/github.com/pzsp-teams/lib.svg)](https://pkg.go.dev/github.com/pzsp-teams/lib)

High-level Go (Golang) library that simplifies interaction with **Microsoft Graph API**.
Provides abstraction over operations related to Teams, Channels, and Chats, adding a layer of automatic caching and name resolution.

## 🚀 Key Features

- **Simplified Authentication**: Built-in MSAL token support.
- **Intelligent Cache**: Automatic mapping of team names to IDs (e.g., "DevOps Team" -> `UUID`), reducing API queries.
- **Facade Architecture**: One main `Client` providing access to all services (`Teams`, `Channels`, `Chats`).
- **Type Safety**: All operations return strongly typed models.

## 📦 Installation

```bash
go get (https://github.com/pzsp-teams/lib)
```

## 🛠️ Architecture & Concepts
The library uses a **Facade Pattern**. The Client struct aggregates domain-specific services:

- **client.Teams**: Manage teams lifecycles and members.
- **client.Channels**: Manage standard and private channels.
- **client.Chats**: Handle messages and chat members.

### The "Reference" concept
Many methods accept a `_Ref` argument. This allows you to pass:
- **UUID**
- **Display Name** (**email** for UserRefs) - this provides convenient usage in interactive applications.
    Library will automatically resolve refs to IDs.

## 💻 Quick Start
Full example usage is showcased [HERE](https://github.com/pzsp-teams/lib/tree/example-cmd-usage/cmd)
Here is a simple example of how to initialize the client and list the current user's teams.

### Client init
```go
import (
    "context"
    "time"
    "[github.com/pzsp-teams/lib](https://github.com/pzsp-teams/lib)"
    "[github.com/pzsp-teams/lib/config](https://github.com/pzsp-teams/lib/config)"
)

func main() {
    ctx := context.Background()

    // Auth config (Azure AD)
    authCfg := &config.AuthConfig{
        ClientID:   "your-client-id",
        Tenant:     "your-tenant-id",
        Email:      "your-email",
        Scopes:     []string{"[https://graph.microsoft.com/.default](https://graph.microsoft.com/.default)"},
        AuthMethod: "DEVICE_CODE", // Or "INTERACTIVE"
    }

    // Cache config
    cacheCfg := &config.CacheConfig{
        Mode:     config.CacheAsync,
        Provider: config.CacheProviderJSONFile, // Local file cache
    }

    // Client init
    client, err := lib.NewClient(ctx, authCfg, nil, cacheCfg)
    if err != nil {
        panic(err)
    }
    defer lib.Close() // Important if using cache
}
```

### 2. Example usage

```go
// List joined teams
teams, _ := client.Teams.ListMyJoined(ctx)
for _, t := range teams {
    fmt.Printf("Team: %s (ID: %s)\n", t.DisplayName, t.ID)
}

// Create a new team
newTeam, _ := client.Teams.CreateViaGroup(ctx, "Project Alpha", "project-alpha", "public")
```

## Authentication
The library uses `config.AuthConfig` to establish the connection.Ensure your Azure App Registration has the necessary **API Permissions** (e.g., `Team.ReadBasic.All`, `Channel.ReadBasic.All`) granted in the Azure Portal.
Complete list of scopes required by all functions is available [HERE](https://github.com/pzsp-teams/lib/blob/example-cmd-usage/.env.template)

There are two available ways to authenticate:
- **INTERACTIVE** - log in window will automatically be opened within your browser.
- **DEVICE CODE** - library will provide you the **URL** and code, which need to be manually opened with browser of your choice.

## Cache
If enabled, stores metadata and non-sensitive mappings, (e.g., `TeamRef` -> `UUID`) to provide efficient reference resolution.

#### ⚠️ Important:
Because the cache might run background goroutines to keep data fresh, you **must** call lib.Close() when your application shuts down. This ensures all background operations complete and prevents memory leaks or race conditions.

```go
defer lib.Close()
```