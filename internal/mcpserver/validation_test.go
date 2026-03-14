package mcpserver

import (
	"context"
	"io"
	"log/slog"
	"slices"
	"testing"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/andr1an/llmdm/config"
)

func TestValidateCampaignName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "valid simple", input: "Lost Mines", wantErr: false},
		{name: "valid max length", input: "A123456789012345678901234567890123456789012345678901234567890123", wantErr: false},
		{name: "starts with space", input: " Lost Mines", wantErr: true},
		{name: "starts with digit", input: "1st Campaign", wantErr: true},
		{name: "too long", input: "A1234567890123456789012345678901234567890123456789012345678901234", wantErr: true},
		{name: "empty", input: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCampaignName(tt.input)
			if tt.wantErr && err == nil {
				t.Fatalf("validateCampaignName(%q) expected error, got nil", tt.input)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("validateCampaignName(%q) unexpected error: %v", tt.input, err)
			}
		})
	}
}

func TestGenerateCampaignID(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "Lost Mines of Phandelver", want: "lost-mines-of-phandelver"},
		{input: "  Curse__of--Strahd  ", want: "curse-of-strahd"},
		{input: "!!!", want: "campaign"},
	}

	for _, tt := range tests {
		if got := generateCampaignID(tt.input); got != tt.want {
			t.Fatalf("generateCampaignID(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestValidateMaxLength(t *testing.T) {
	if err := validateMaxLength("notes", "abc", 3); err != nil {
		t.Fatalf("validateMaxLength should allow exact limit: %v", err)
	}
	if err := validateMaxLength("notes", "abcd", 3); err == nil {
		t.Fatal("validateMaxLength expected error for value above limit")
	}
}

func TestToolRegistrationIntegrity(t *testing.T) {
	cfg := &config.Config{DBPath: t.TempDir(), LogLevel: "info", Transport: "stdio", HTTPAddr: "127.0.0.1:0", HTTPEndpoint: "/mcp"}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv := New(cfg, logger)

	c, err := client.NewInProcessClient(srv.mcp)
	if err != nil {
		t.Fatalf("NewInProcessClient() error = %v", err)
	}

	ctx := context.Background()
	if err := c.Start(ctx); err != nil {
		t.Fatalf("client.Start() error = %v", err)
	}

	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{Name: "test-client", Version: "1.0.0"}
	initRequest.Params.Capabilities = mcp.ClientCapabilities{}
	if _, err := c.Initialize(ctx, initRequest); err != nil {
		t.Fatalf("client.Initialize() error = %v", err)
	}

	result, err := c.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		t.Fatalf("client.ListTools() error = %v", err)
	}

	got := make([]string, 0, len(result.Tools))
	for _, tool := range result.Tools {
		got = append(got, tool.Name)
	}
	slices.Sort(got)

	want := []string{
		"checkpoint",
		"create_campaign",
		"end_session",
		"export_session_recap",
		"get_character",
		"get_npc_relationships",
		"get_roll_history",
		"get_session_brief",
		"get_turn_history",
		"get_world_flags",
		"list_campaigns",
		"list_characters",
		"list_open_hooks",
		"list_sessions",
		"resolve_hook",
		"roll",
		"roll_contested",
		"roll_saving_throw",
		"save_character",
		"save_plot_event",
		"set_world_flag",
		"start_session",
		"update_character",
	}
	slices.Sort(want)

	if !slices.Equal(got, want) {
		t.Fatalf("registered tools mismatch\nwant: %v\ngot:  %v", want, got)
	}
}
