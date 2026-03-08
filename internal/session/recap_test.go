package session

import (
	"testing"

	"github.com/andr1an/llmdm/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestRenderRecap(t *testing.T) {
	recap := RenderRecap(
		"Lost Mines",
		[]types.SessionMeta{{Session: 1, Date: "2026-03-01", SummaryPreview: "The party met at the inn.", HooksOpened: 2, HooksResolved: 0}},
		[]types.Hook{{Hook: "Find Gundren", SessionOpened: 1}},
		map[string]string{"weather": "storm"},
	)

	assert.Contains(t, recap, "# Session Recap - Lost Mines")
	assert.Contains(t, recap, "## Session 1")
	assert.Contains(t, recap, "The party met at the inn.")
	assert.Contains(t, recap, "Find Gundren")
	assert.Contains(t, recap, "weather: storm")
}

func TestRenderRecap_Empty(t *testing.T) {
	recap := RenderRecap("", nil, nil, nil)

	assert.Contains(t, recap, "No recorded sessions in the selected range.")
	assert.Contains(t, recap, "## Still Open Hooks")
	assert.Contains(t, recap, "## World State Snapshot")
}
