package config

import "testing"

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
