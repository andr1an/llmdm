// Package main is the entry point for the D&D Campaign MCP server.
package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/andr1an/llmdm/config"
	"github.com/andr1an/llmdm/internal/mcpserver"
)

func main() {
	if len(os.Args) < 2 || os.Args[1] != "serve" {
		fmt.Println("Usage: dnd-mcp serve")
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	logger, err := mcpserver.NewJSONLogger(cfg.LogLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid LOG_LEVEL %q: %v. Falling back to info.\n", cfg.LogLevel, err)
		logger, _ = mcpserver.NewJSONLogger("info")
	}
	slog.SetDefault(logger)

	srv := mcpserver.New(cfg, logger)
	if err := srv.Serve(); err != nil {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}
}
