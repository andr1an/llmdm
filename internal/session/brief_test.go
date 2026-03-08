package session

import (
	"testing"

	"github.com/andr1an/llmdm/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderBrief(t *testing.T) {
	text, err := RenderBrief(BriefData{
		Session:      4,
		CampaignName: "Lost Mines",
		Characters: []types.CharacterSummary{{
			Name:       "Aldric",
			Class:      "Paladin",
			Level:      5,
			HP:         types.HP{Current: 45, Max: 45},
			Conditions: []string{"Blessed"},
		}},
		RecentSummaries: "- Session 3: The party escaped Cragmaw Keep.",
		OpenHooks: []types.Hook{{
			Hook:          "Who leads the dragon cult?",
			SessionOpened: 3,
		}},
		WorldFlags:     map[string]string{"weather": "rain"},
		LastCheckpoint: "The party camped at the ridge.",
	})
	require.NoError(t, err)
	assert.Contains(t, text, "Campaign Brief - Session 4")
	assert.Contains(t, text, "Aldric")
	assert.Contains(t, text, "Who leads the dragon cult?")
	assert.Contains(t, text, "The party camped at the ridge.")
}

func TestRenderQuickBrief(t *testing.T) {
	brief := RenderQuickBrief(
		"Lost Mines",
		&types.Session{Session: 7, Summary: "The heroes reclaimed the keep."},
		[]types.CharacterSummary{{Name: "Aldric"}, {Name: "Zara"}},
		[]types.Hook{{Hook: "Find the missing heir"}},
		map[string]string{"season": "winter"},
	)

	assert.Contains(t, brief, "Quick Session Brief")
	assert.Contains(t, brief, "Last session: 7")
	assert.Contains(t, brief, "Aldric, Zara")
	assert.Contains(t, brief, "Find the missing heir")
}

func TestFallbackCompress(t *testing.T) {
	summary := fallbackCompress("The party crossed the river and met a druid.")
	assert.Contains(t, summary, "OPEN HOOKS:")
}
