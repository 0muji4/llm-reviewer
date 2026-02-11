package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage: mcp-client <project_path> <query> [persona]")
		os.Exit(1)
	}

	projectPath := os.Args[1]
	query := os.Args[2]
	personaName := "architect"
	if len(os.Args) >= 4 {
		personaName = os.Args[3]
	}

	serverBin := os.Getenv("MCP_SERVER_BIN")
	if serverBin == "" {
		serverBin = "mcp-server"
	}

	// --- MCP クライアントの起動（サーバープロセスを spawn） ---
	c, err := client.NewStdioMCPClient(
		serverBin,
		os.Environ(),
	)
	if err != nil {
		log.Fatalf("failed to create MCP client: %v", err)
	}
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// --- Initialize ハンドシェイク ---
	initReq := mcp.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initReq.Params.ClientInfo = mcp.Implementation{
		Name:    "llm-reviewer-client",
		Version: "0.1.0",
	}

	initResult, err := c.Initialize(ctx, initReq)
	if err != nil {
		log.Fatalf("failed to initialize: %v", err)
	}
	fmt.Fprintf(os.Stderr, "Connected to: %s %s\n", initResult.ServerInfo.Name, initResult.ServerInfo.Version)

	// --- review ツールの呼び出し ---
	toolReq := mcp.CallToolRequest{}
	toolReq.Params.Name = "review"
	toolReq.Params.Arguments = map[string]any{
		"project_path": projectPath,
		"query":        query,
		"persona":      personaName,
	}

	fmt.Fprintf(os.Stderr, "Reviewing %s with persona %q...\n", projectPath, personaName)

	result, err := c.CallTool(ctx, toolReq)
	if err != nil {
		log.Fatalf("tool call failed: %v", err)
	}

	if result.IsError {
		fmt.Fprintln(os.Stderr, "Review failed:")
	}

	for _, content := range result.Content {
		if tc, ok := content.(mcp.TextContent); ok {
			fmt.Println(tc.Text)
		}
	}
}
