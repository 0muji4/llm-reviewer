package server

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// New は MCP サーバーを生成し、ツールを登録して返します。
// ビジネスロジックは handler に委譲し、ここではプロトコル変換のみ行います。
func New(handler *ReviewHandler) *server.MCPServer {
	s := server.NewMCPServer(
		"llm-reviewer",
		"0.1.0",
		server.WithToolCapabilities(false),
	)

	reviewTool := mcp.NewTool("review",
		mcp.WithDescription("指定されたGoプロジェクトに対してLLMベースのコードレビューを実行します。Gemini ReActループにより、LSP・AST・Git差分を活用した深いコード分析を行います。"),
		mcp.WithString("project_path",
			mcp.Required(),
			mcp.Description("レビュー対象のGoプロジェクトの絶対パス"),
		),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("レビューの指示・質問（例: 「このプロジェクトのアーキテクチャをレビューしてください」）"),
		),
		mcp.WithString("persona",
			mcp.Description("使用するペルソナ名（architect, go-expert）。デフォルト: architect"),
		),
	)

	s.AddTool(reviewTool, handler.Handle)

	return s
}
