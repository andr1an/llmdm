// Package dice provides dice notation parsing and rolling.
package dice

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/andr1an/llmdm/internal/types"
)

// Supported formats:
//   NdM          — e.g. 1d20, 2d6
//   NdM+K        — e.g. 2d6+3
//   NdM-K        — e.g. 1d20-2
//   NdMkhX       — keep highest X, e.g. 4d6kh3
//   NdMklX       — keep lowest X,  e.g. 4d6kl3

var diceRegex = regexp.MustCompile(`^(\d+)d(\d+)(?:(kh|kl)(\d+))?([+-]\d+)?$`)

// ParsedRoll represents a parsed dice notation.
type ParsedRoll struct {
	Count    int
	Sides    int
	Modifier int
	KeepHigh int // 0 = keep all
	KeepLow  int // 0 = keep all
	Notation string
}

// Parse parses a dice notation string.
func Parse(notation string) (*ParsedRoll, error) {
	notation = strings.ToLower(strings.TrimSpace(notation))
	matches := diceRegex.FindStringSubmatch(notation)
	if matches == nil {
		return nil, fmt.Errorf("invalid dice notation: %s", notation)
	}

	count, _ := strconv.Atoi(matches[1])
	sides, _ := strconv.Atoi(matches[2])

	if count < 1 || count > 100 {
		return nil, fmt.Errorf("dice count must be between 1 and 100")
	}
	if sides < 1 || sides > 1000 {
		return nil, fmt.Errorf("dice sides must be between 1 and 1000")
	}

	roll := &ParsedRoll{
		Count:    count,
		Sides:    sides,
		Notation: notation,
	}

	// Parse keep high/low
	if matches[3] != "" {
		keepCount, _ := strconv.Atoi(matches[4])
		if keepCount < 1 || keepCount > count {
			return nil, fmt.Errorf("keep count must be between 1 and %d", count)
		}
		switch matches[3] {
		case "kh":
			roll.KeepHigh = keepCount
		case "kl":
			roll.KeepLow = keepCount
		}
	}

	// Parse modifier
	if matches[5] != "" {
		roll.Modifier, _ = strconv.Atoi(matches[5])
	}

	return roll, nil
}

// Execute rolls the dice and returns the result.
func (r *ParsedRoll) Execute() types.RollResult {
	rolls := make([]int, r.Count)
	for i := 0; i < r.Count; i++ {
		rolls[i] = rollDie(r.Sides)
	}

	kept := r.applyKeep(rolls)
	total := sum(kept) + r.Modifier

	return types.RollResult{
		Total:     total,
		Rolls:     rolls,
		Kept:      kept,
		Modifier:  r.Modifier,
		Notation:  r.Notation,
		RollID:    uuid.New().String(),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

func (r *ParsedRoll) applyKeep(rolls []int) []int {
	if r.KeepHigh == 0 && r.KeepLow == 0 {
		return rolls
	}

	sorted := make([]int, len(rolls))
	copy(sorted, rolls)
	sort.Ints(sorted)

	if r.KeepHigh > 0 {
		return sorted[len(sorted)-r.KeepHigh:]
	}
	return sorted[:r.KeepLow]
}

// rollDie returns a cryptographically random integer between 1 and sides.
func rollDie(sides int) int {
	max := big.NewInt(int64(sides))
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		// Fallback should never happen with crypto/rand
		panic(fmt.Sprintf("crypto/rand failed: %v", err))
	}
	return int(n.Int64()) + 1
}

func sum(values []int) int {
	total := 0
	for _, v := range values {
		total += v
	}
	return total
}

// Roll parses and executes a dice notation.
func Roll(notation string) (types.RollResult, error) {
	parsed, err := Parse(notation)
	if err != nil {
		return types.RollResult{}, err
	}
	return parsed.Execute(), nil
}

// RollWithAdvantage rolls 2d20 and keeps the higher result.
func RollWithAdvantage(modifier int) types.RollResult {
	roll1 := rollDie(20)
	roll2 := rollDie(20)
	rolls := []int{roll1, roll2}

	kept := []int{max(roll1, roll2)}
	total := kept[0] + modifier

	notation := "2d20kh1"
	if modifier > 0 {
		notation += fmt.Sprintf("+%d", modifier)
	} else if modifier < 0 {
		notation += fmt.Sprintf("%d", modifier)
	}

	return types.RollResult{
		Total:     total,
		Rolls:     rolls,
		Kept:      kept,
		Modifier:  modifier,
		Notation:  notation,
		RollID:    uuid.New().String(),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

// RollWithDisadvantage rolls 2d20 and keeps the lower result.
func RollWithDisadvantage(modifier int) types.RollResult {
	roll1 := rollDie(20)
	roll2 := rollDie(20)
	rolls := []int{roll1, roll2}

	kept := []int{min(roll1, roll2)}
	total := kept[0] + modifier

	notation := "2d20kl1"
	if modifier > 0 {
		notation += fmt.Sprintf("+%d", modifier)
	} else if modifier < 0 {
		notation += fmt.Sprintf("%d", modifier)
	}

	return types.RollResult{
		Total:     total,
		Rolls:     rolls,
		Kept:      kept,
		Modifier:  modifier,
		Notation:  notation,
		RollID:    uuid.New().String(),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

// ModifierFromStat calculates the D&D 5e ability modifier from a stat value.
// Uses floor division as per D&D rules.
func ModifierFromStat(stat int) int {
	diff := stat - 10
	if diff < 0 && diff%2 != 0 {
		return diff/2 - 1
	}
	return diff / 2
}
