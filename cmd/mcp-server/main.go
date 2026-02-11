package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	mcpserver "github.com/mark3labs/mcp-go/server"

	"github.com/0muji4/llm-reviewer/internal/server"
)

func main() {
	// --- 環境変数の読み込み ---
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY is required")
	}

	personaDir := os.Getenv("PERSONA_DIR")
	if personaDir == "" {
		// デフォルト: カレントディレクトリからの相対パス
		wd, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		personaDir = filepath.Join(wd, "configs", "personas")
	}
	// 相対パスを絶対パスに解決（子プロセスの CWD が異なる場合に備える）
	personaDir, err := filepath.Abs(personaDir)
	if err != nil {
		log.Fatal(err)
	}

	// --- DI: Adapter 層の組み立て ---
	handler := server.NewReviewHandler(apiKey, personaDir)
	s := server.New(handler)

	// --- Framework: MCP stdio サーバーの起動 ---
	fmt.Fprintln(os.Stderr, "llm-reviewer MCP server starting...")
	if err := mcpserver.ServeStdio(s); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
