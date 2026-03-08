package dm

import (
	"testing"

	"github.com/andr1an/llmdm/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestBuildSystemPrompt(t *testing.T) {
	prompt := BuildSystemPrompt(
		"## Campaign Brief - Session 3\nSomething happened.",
		[]types.Hook{{Hook: "Who betrayed the council?", SessionOpened: 2}},
		map[string]string{"season": "winter"},
	)

	assert.Contains(t, prompt, "You are an expert Dungeon Master")
	assert.Contains(t, prompt, "## Current Campaign State")
	assert.Contains(t, prompt, "Campaign Brief - Session 3")
	assert.Contains(t, prompt, "Who betrayed the council?")
	assert.Contains(t, prompt, "season: winter")
}

func TestBuildSystemPrompt_EmptyState(t *testing.T) {
	prompt := BuildSystemPrompt("brief", nil, map[string]string{})

	assert.Contains(t, prompt, "- No unresolved hooks.")
	assert.Contains(t, prompt, "- No world flags recorded.")
}
