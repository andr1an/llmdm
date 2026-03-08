package config

import (
	"testing"
	"time"
)

func TestLoad_DefaultTransportIsStdio(t *testing.T) {
	t.Setenv("MCP_TRANSPORT", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	if cfg.Transport != "stdio" {
		t.Fatalf("cfg.Transport = %q, want %q", cfg.Transport, "stdio")
	}
}

func TestLoad_ParsesTransportAliases(t *testing.T) {
	t.Setenv("MCP_TRANSPORT", " Streamable_HTTP ")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	if cfg.Transport != "streamable-http" {
		t.Fatalf("cfg.Transport = %q, want %q", cfg.Transport, "streamable-http")
	}
}

func TestLoad_InvalidTransportReturnsError(t *testing.T) {
	t.Setenv("MCP_TRANSPORT", "socket")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want non-nil")
	}
}

func TestLoad_DefaultHTTPTimeouts(t *testing.T) {
	t.Setenv("READ_TIMEOUT", "")
	t.Setenv("WRITE_TIMEOUT", "")
	t.Setenv("IDLE_TIMEOUT", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	if cfg.ReadTimeout != 15*time.Second {
		t.Fatalf("cfg.ReadTimeout = %v, want %v", cfg.ReadTimeout, 15*time.Second)
	}
	if cfg.WriteTimeout != 60*time.Second {
		t.Fatalf("cfg.WriteTimeout = %v, want %v", cfg.WriteTimeout, 60*time.Second)
	}
	if cfg.IdleTimeout != 60*time.Second {
		t.Fatalf("cfg.IdleTimeout = %v, want %v", cfg.IdleTimeout, 60*time.Second)
	}
}

func TestLoad_InvalidReadTimeoutReturnsError(t *testing.T) {
	t.Setenv("READ_TIMEOUT", "abc")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want non-nil")
	}
}
