package mcpserver

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (s *Server) serveHTTP() error {
	s.log().Debug("configuring streamable HTTP server")
	mcpHandler := mcp.NewStreamableHTTPHandler(
		func(r *http.Request) *mcp.Server { return s.mcp },
		&mcp.StreamableHTTPOptions{
			Stateless:      false,
			Logger:         s.logger,
			SessionTimeout: 5 * time.Minute,
		},
	)

	endpoint := "/" + strings.Trim(s.cfg.HTTPEndpoint, "/")
	mux := http.NewServeMux()
	mux.Handle(endpoint, mcpHandler)
	if endpoint != "/" {
		mux.Handle(endpoint+"/", mcpHandler)
	}
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "ok",
			"transport": "streamable-http",
			"endpoint":  endpoint,
			"time":      time.Now().UTC().Format(time.RFC3339),
		})
	})

	httpServer := &http.Server{
		Addr:              s.cfg.HTTPAddr,
		Handler:           mux,
		ReadHeaderTimeout: s.cfg.ReadTimeout,
		ReadTimeout:       s.cfg.ReadTimeout,
		WriteTimeout:      s.cfg.WriteTimeout,
		IdleTimeout:       s.cfg.IdleTimeout,
	}
	s.log().Info("serving MCP over streamable HTTP", "addr", s.cfg.HTTPAddr, "endpoint", endpoint)
	return httpServer.ListenAndServe()
}
