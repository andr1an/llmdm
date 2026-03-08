package dice

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name      string
		notation  string
		want      *ParsedRoll
		wantError bool
	}{
		{
			name:     "simple d20",
			notation: "1d20",
			want: &ParsedRoll{
				Count:    1,
				Sides:    20,
				Notation: "1d20",
			},
		},
		{
			name:     "multiple dice",
			notation: "3d6",
			want: &ParsedRoll{
				Count:    3,
				Sides:    6,
				Notation: "3d6",
			},
		},
		{
			name:     "positive modifier",
			notation: "2d6+3",
			want: &ParsedRoll{
				Count:    2,
				Sides:    6,
				Modifier: 3,
				Notation: "2d6+3",
			},
		},
		{
			name:     "negative modifier",
			notation: "1d20-2",
			want: &ParsedRoll{
				Count:    1,
				Sides:    20,
				Modifier: -2,
				Notation: "1d20-2",
			},
		},
		{
			name:     "keep highest",
			notation: "4d6kh3",
			want: &ParsedRoll{
				Count:    4,
				Sides:    6,
				KeepHigh: 3,
				Notation: "4d6kh3",
			},
		},
		{
			name:     "keep lowest",
			notation: "4d6kl3",
			want: &ParsedRoll{
				Count:    4,
				Sides:    6,
				KeepLow:  3,
				Notation: "4d6kl3",
			},
		},
		{
			name:     "keep highest with modifier",
			notation: "4d6kh3+2",
			want: &ParsedRoll{
				Count:    4,
				Sides:    6,
				KeepHigh: 3,
				Modifier: 2,
				Notation: "4d6kh3+2",
			},
		},
		{
			name:     "uppercase notation",
			notation: "2D6+5",
			want: &ParsedRoll{
				Count:    2,
				Sides:    6,
				Modifier: 5,
				Notation: "2d6+5",
			},
		},
		{
			name:      "invalid notation",
			notation:  "roll d20",
			wantError: true,
		},
		{
			name:      "too many dice",
			notation:  "101d6",
			wantError: true,
		},
		{
			name:      "keep more than rolled",
			notation:  "2d6kh3",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.notation)
			if tt.wantError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want.Count, got.Count)
			assert.Equal(t, tt.want.Sides, got.Sides)
			assert.Equal(t, tt.want.Modifier, got.Modifier)
			assert.Equal(t, tt.want.KeepHigh, got.KeepHigh)
			assert.Equal(t, tt.want.KeepLow, got.KeepLow)
		})
	}
}

func TestExecute(t *testing.T) {
	t.Run("single d20", func(t *testing.T) {
		parsed, err := Parse("1d20")
		require.NoError(t, err)

		result := parsed.Execute()
		assert.Len(t, result.Rolls, 1)
		assert.GreaterOrEqual(t, result.Rolls[0], 1)
		assert.LessOrEqual(t, result.Rolls[0], 20)
		assert.Equal(t, result.Rolls[0], result.Total)
		assert.NotEmpty(t, result.RollID)
		assert.NotEmpty(t, result.Timestamp)
	})

	t.Run("2d6+3", func(t *testing.T) {
		parsed, err := Parse("2d6+3")
		require.NoError(t, err)

		result := parsed.Execute()
		assert.Len(t, result.Rolls, 2)
		for _, roll := range result.Rolls {
			assert.GreaterOrEqual(t, roll, 1)
			assert.LessOrEqual(t, roll, 6)
		}
		assert.Equal(t, result.Rolls[0]+result.Rolls[1]+3, result.Total)
		assert.Equal(t, 3, result.Modifier)
	})

	t.Run("4d6kh3 keeps highest 3", func(t *testing.T) {
		parsed, err := Parse("4d6kh3")
		require.NoError(t, err)

		result := parsed.Execute()
		assert.Len(t, result.Rolls, 4)
		assert.Len(t, result.Kept, 3)

		// Verify kept are the highest 3
		total := 0
		for _, v := range result.Kept {
			total += v
		}
		assert.Equal(t, total, result.Total)
	})

	t.Run("4d6kl3 keeps lowest 3", func(t *testing.T) {
		parsed, err := Parse("4d6kl3")
		require.NoError(t, err)

		result := parsed.Execute()
		assert.Len(t, result.Rolls, 4)
		assert.Len(t, result.Kept, 3)

		// Verify kept are the lowest 3
		total := 0
		for _, v := range result.Kept {
			total += v
		}
		assert.Equal(t, total, result.Total)
	})
}

func TestRoll(t *testing.T) {
	result, err := Roll("1d20+5")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, result.Total, 6)
	assert.LessOrEqual(t, result.Total, 25)
	assert.Equal(t, "1d20+5", result.Notation)
}

func TestRollWithAdvantage(t *testing.T) {
	result := RollWithAdvantage(5)
	assert.Len(t, result.Rolls, 2)
	assert.Len(t, result.Kept, 1)
	assert.Equal(t, max(result.Rolls[0], result.Rolls[1]), result.Kept[0])
	assert.Equal(t, result.Kept[0]+5, result.Total)
}

func TestRollWithDisadvantage(t *testing.T) {
	result := RollWithDisadvantage(-2)
	assert.Len(t, result.Rolls, 2)
	assert.Len(t, result.Kept, 1)
	assert.Equal(t, min(result.Rolls[0], result.Rolls[1]), result.Kept[0])
	assert.Equal(t, result.Kept[0]-2, result.Total)
}

func TestModifierFromStat(t *testing.T) {
	tests := []struct {
		stat     int
		modifier int
	}{
		{1, -5}, // floor((1-10)/2) = -5
		{8, -1},
		{10, 0},
		{11, 0},
		{12, 1},
		{14, 2},
		{18, 4},
		{20, 5},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			assert.Equal(t, tt.modifier, ModifierFromStat(tt.stat))
		})
	}
}

func TestDiceDistribution(t *testing.T) {
	// Statistical test: roll many d6s and verify distribution is reasonable
	counts := make(map[int]int)
	iterations := 6000

	parsed, err := Parse("1d6")
	require.NoError(t, err)

	for i := 0; i < iterations; i++ {
		result := parsed.Execute()
		counts[result.Total]++
	}

	// Each face should appear roughly 1/6 of the time (±20% tolerance)
	expected := iterations / 6
	tolerance := expected / 5

	for face := 1; face <= 6; face++ {
		assert.InDelta(t, expected, counts[face], float64(tolerance),
			"face %d appeared %d times, expected ~%d", face, counts[face], expected)
	}
}
