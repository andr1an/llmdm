package db

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestIsValidCampaignID(t *testing.T) {
	tests := []struct {
		name       string
		campaignID string
		want       bool
	}{
		{name: "simple", campaignID: "campaign", want: true},
		{name: "with dash", campaignID: "lost-mines", want: true},
		{name: "empty", campaignID: "", want: false},
		{name: "leading dash", campaignID: "-campaign", want: false},
		{name: "trailing dash", campaignID: "campaign-", want: false},
		{name: "double dash", campaignID: "lost--mines", want: false},
		{name: "has digit", campaignID: "campaign2", want: false},
		{name: "uppercase", campaignID: "Campaign", want: false},
		{name: "space", campaignID: "lost mines", want: false},
		{name: "underscore", campaignID: "lost_mines", want: false},
		{name: "relative traversal", campaignID: "../../tmp/pwn", want: false},
		{name: "absolute path", campaignID: "/tmp/pwn", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidCampaignID(tt.campaignID)
			if got != tt.want {
				t.Fatalf("IsValidCampaignID(%q) = %v, want %v", tt.campaignID, got, tt.want)
			}
		})
	}
}

func TestCampaignDBPath(t *testing.T) {
	base := filepath.Join("var", "db")

	path, err := CampaignDBPath(base, "lost-mines")
	if err != nil {
		t.Fatalf("CampaignDBPath returned unexpected error: %v", err)
	}
	want := filepath.Join(base, "lost-mines.db")
	if path != want {
		t.Fatalf("CampaignDBPath path = %q, want %q", path, want)
	}

	if _, err := CampaignDBPath(base, "../../tmp/pwn"); err == nil {
		t.Fatal("CampaignDBPath should reject traversal campaign_id")
	}
	if _, err := CampaignDBPath(base, "/tmp/pwn"); err == nil {
		t.Fatal("CampaignDBPath should reject absolute-path campaign_id")
	}
}

func TestOpenCreatesPrivateDBDirectory(t *testing.T) {
	root := t.TempDir()
	dbPath := filepath.Join(root, "campaigns", "lost-mines.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open returned unexpected error: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	dirInfo, err := os.Stat(filepath.Dir(dbPath))
	if err != nil {
		t.Fatalf("stat db directory: %v", err)
	}

	perm := dirInfo.Mode().Perm()
	if runtime.GOOS == "windows" {
		// Windows does not map Unix permission bits reliably.
		return
	}
	if perm&0o077 != 0 {
		t.Fatalf("db directory permissions = %03o, want no group/other permissions", perm)
	}
}
