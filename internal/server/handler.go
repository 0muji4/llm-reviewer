package server

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/0muji4/llm-reviewer/internal/agent"
	"github.com/0muji4/llm-reviewer/internal/lsp"
	"github.com/0muji4/llm-reviewer/internal/persona"
	"github.com/0muji4/llm-reviewer/internal/symbol"
	"github.com/0muji4/llm-reviewer/internal/workspace"

	"github.com/mark3labs/mcp-go/mcp"
)

// ReviewHandler は MCP リクエストを Agent のユースケースに変換する Adapter です。
type ReviewHandler struct {
	apiKey     string
	personaDir string
}

// NewReviewHandler は ReviewHandler を生成します。
func NewReviewHandler(apiKey, personaDir string) *ReviewHandler {
	return &ReviewHandler{
		apiKey:     apiKey,
		personaDir: personaDir,
	}
}

// Handle は MCP の review ツール呼び出しを受け取り、Agent の ReAct ループを実行します。
func (h *ReviewHandler) Handle(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	rawPath, err := req.RequireString("project_path")
	if err != nil {
		return mcp.NewToolResultError("project_path is required"), nil
	}
	// 相対パスを絶対パスに解決
	projectPath, err := filepath.Abs(rawPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid project_path: %v", err)), nil
	}
	query, err := req.RequireString("query")
	if err != nil {
		return mcp.NewToolResultError("query is required"), nil
	}
	personaName := req.GetString("persona", "architect")

	// 1. Persona の読み込み
	personaPath := fmt.Sprintf("%s/%s.yaml", h.personaDir, personaName)
	p, err := persona.Load(personaPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to load persona %q: %v", personaName, err)), nil
	}

	// 2. Infrastructure 層の生成
	lspClient, err := lsp.NewClient(projectPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to start LSP: %v", err)), nil
	}
	defer lspClient.Close()

	fsReader := workspace.NewFSReader(projectPath)
	gitDiff := workspace.NewGitDiff(projectPath)
	astResolver := symbol.NewASTResolver(projectPath)

	// 3. UseCase 層（Agent）の生成と実行
	bot, err := agent.NewL5Agent(ctx, h.apiKey, projectPath, p.SystemPrompt, lspClient, fsReader, gitDiff, astResolver)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create agent: %v", err)), nil
	}

	result, err := bot.Run(ctx, query)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("agent error: %v", err)), nil
	}

	return mcp.NewToolResultText(result), nil
}
