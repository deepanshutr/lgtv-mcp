package main

import (
	"fmt"
	"log"
	"os"

	"github.com/deepanshutr/lgtv-mcp/internal/core"
	"github.com/deepanshutr/lgtv-mcp/internal/tools"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// MCP protocol uses stdout — divert all logging to stderr or a file.
	log.SetOutput(os.Stderr)

	if logPath := os.Getenv("LGTV_MCP_LOG"); logPath != "" {
		f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
		if err == nil {
			log.SetOutput(f)
		}
	}

	coreURL := os.Getenv("LGTV_CORE_URL")
	if coreURL == "" {
		coreURL = "http://127.0.0.1:8765"
	}
	client := core.New(coreURL)

	s := server.NewMCPServer("lgtv-mcp", "0.1.0",
		server.WithToolCapabilities(false),
	)
	tools.Register(s, client)

	log.Printf("lgtv-mcp starting; core=%s", coreURL)
	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintln(os.Stderr, "fatal:", err)
		os.Exit(1)
	}
}
