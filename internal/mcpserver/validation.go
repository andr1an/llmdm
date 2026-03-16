package mcpserver

import (
	"fmt"
	"regexp"
	"unicode/utf8"
)

var campaignNamePattern = regexp.MustCompile(`^[A-Za-z]`)

const (
	maxCharacterBackstoryLength = 8000
	maxCharacterNotesLength     = 4000
	maxSessionRawEventsLength   = 50000
	maxSessionDMNotesLength     = 4000
	maxCheckpointNoteLength     = 4000
)

// generateCampaignID creates a URL-safe campaign ID from a name.
func generateCampaignID(name string) string {
	id := ""
	for _, c := range name {
		if c >= 'a' && c <= 'z' {
			id += string(c)
		} else if c >= 'A' && c <= 'Z' {
			id += string(c + 32) // lowercase
		} else if c == ' ' || c == '-' || c == '_' {
			if len(id) > 0 && id[len(id)-1] != '-' {
				id += "-"
			}
		}
	}
	// Remove trailing dash
	for len(id) > 0 && id[len(id)-1] == '-' {
		id = id[:len(id)-1]
	}
	if id == "" {
		id = "campaign"
	}
	return id
}

func validateCampaignName(name string) error {
	if utf8.RuneCountInString(name) > 64 {
		return fmt.Errorf("campaign name must be 64 characters or fewer")
	}
	if !campaignNamePattern.MatchString(name) {
		return fmt.Errorf("campaign name must start with a letter")
	}
	return nil
}

func validateMaxLength(fieldName, value string, maxRunes int) error {
	if utf8.RuneCountInString(value) > maxRunes {
		return fmt.Errorf("%s must be %d characters or fewer", fieldName, maxRunes)
	}
	return nil
}

var validAlignments = map[string]bool{
	"Lawful Good":     true,
	"Neutral Good":    true,
	"Chaotic Good":    true,
	"Lawful Neutral":  true,
	"True Neutral":    true,
	"Chaotic Neutral": true,
	"Lawful Evil":     true,
	"Neutral Evil":    true,
	"Chaotic Evil":    true,
}

var validAbilityScores = map[string]bool{
	"STR": true, "DEX": true, "CON": true,
	"INT": true, "WIS": true, "CHA": true,
}

func validateAlignment(alignment string) error {
	if alignment == "" {
		return nil
	}
	if !validAlignments[alignment] {
		return fmt.Errorf("invalid alignment: %s", alignment)
	}
	return nil
}

func validateAC(ac int) error {
	if ac < 0 || ac > 30 {
		return fmt.Errorf("armor class must be between 0 and 30, got: %d", ac)
	}
	return nil
}

func validateSpellcastingAbility(ability string) error {
	if ability == "" {
		return nil
	}
	if !validAbilityScores[ability] {
		return fmt.Errorf("invalid spellcasting ability: %s (must be STR, DEX, CON, INT, WIS, or CHA)", ability)
	}
	return nil
}
