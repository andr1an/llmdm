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
