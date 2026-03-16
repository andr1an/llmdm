package mcpserver

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

func NewJSONLogger(levelText string) (*slog.Logger, error) {
	var level slog.Level
	switch strings.ToLower(strings.TrimSpace(levelText)) {
	case "debug":
		level = slog.LevelDebug
	case "", "info":
		level = slog.LevelInfo
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		return nil, fmt.Errorf("supported levels are debug, info, warn, error")
	}
	return slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: level})), nil
}

func (s *Server) log() *slog.Logger {
	if s != nil && s.logger != nil {
		return s.logger
	}
	return slog.Default()
}

// logToolEntry logs the start of a tool invocation
func (s *Server) logToolEntry(toolName string, campaignID string, extraAttrs ...interface{}) {
	attrs := []interface{}{"tool", toolName}
	if campaignID != "" {
		attrs = append(attrs, "campaign_id", campaignID)
	}
	attrs = append(attrs, extraAttrs...)
	s.log().Debug("tool called", attrs...)
}

// logToolExit logs successful completion of a tool invocation
func (s *Server) logToolExit(toolName string, campaignID string, extraAttrs ...interface{}) {
	attrs := []interface{}{"tool", toolName}
	if campaignID != "" {
		attrs = append(attrs, "campaign_id", campaignID)
	}
	attrs = append(attrs, extraAttrs...)
	s.log().Debug("tool completed", attrs...)
}

// logToolError logs a tool invocation error
func (s *Server) logToolError(toolName string, campaignID string, err error, extraAttrs ...interface{}) {
	attrs := []interface{}{"tool", toolName}
	if campaignID != "" {
		attrs = append(attrs, "campaign_id", campaignID)
	}
	if err != nil {
		attrs = append(attrs, "error", err.Error())
	}
	attrs = append(attrs, extraAttrs...)
	s.log().Error("tool failed", attrs...)
}
