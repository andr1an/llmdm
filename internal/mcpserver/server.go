package mcpserver

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/mark3labs/mcp-go/server"

	"github.com/andr1an/llmdm/config"
)

// Server holds the MCP server and its dependencies.
type Server struct {
	mcp    *server.MCPServer
	cfg    *config.Config
	dbPath string
	logger *slog.Logger
}

// New builds a configured MCP server instance and registers all tools.
func New(cfg *config.Config, logger *slog.Logger) *Server {
	srv := &Server{
		cfg:    cfg,
		dbPath: cfg.DBPath,
		logger: logger,
	}

	srv.mcp = server.NewMCPServer(
		"D&D Campaign Memory",
		"1.0.0",
		server.WithToolCapabilities(false),
		server.WithRecovery(),
	)

	srv.registerTools()
	return srv
}

// DBPath returns the database directory path.
func (s *Server) DBPath() string {
	return s.dbPath
}

// Serve starts the MCP server based on configured transport.
func (s *Server) Serve() error {
	s.log().Info(
		"server startup",
		"transport", strings.ToLower(strings.TrimSpace(s.cfg.Transport)),
		"http_addr", s.cfg.HTTPAddr,
		"http_endpoint", s.cfg.HTTPEndpoint,
		"log_level", strings.ToLower(strings.TrimSpace(s.cfg.LogLevel)),
	)

	switch strings.ToLower(strings.TrimSpace(s.cfg.Transport)) {
	case "http", "streamable-http", "streamable_http":
		if err := s.serveHTTP(); err != nil {
			return fmt.Errorf("http server error: %w", err)
		}
		return nil
	default:
		if err := server.ServeStdio(s.mcp); err != nil {
			return fmt.Errorf("stdio server error: %w", err)
		}
		return nil
	}
}
