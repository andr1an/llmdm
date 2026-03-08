package session

import (
	"fmt"
	"sort"
	"strings"

	"github.com/andr1an/llmdm/internal/types"
)

// RenderRecap creates a markdown export for a campaign session range.
func RenderRecap(campaignName string, sessions []types.SessionMeta, openHooks []types.Hook, worldFlags map[string]string) string {
	var b strings.Builder

	b.WriteString("# Session Recap")
	if campaignName != "" {
		b.WriteString(" - ")
		b.WriteString(campaignName)
	}
	b.WriteString("\n\n")

	if len(sessions) == 0 {
		b.WriteString("No recorded sessions in the selected range.\n")
	} else {
		for _, s := range sessions {
			b.WriteString(fmt.Sprintf("## Session %d\n", s.Session))
			if s.Date != "" {
				b.WriteString(fmt.Sprintf("Date: %s\n", s.Date))
			}
			if strings.TrimSpace(s.SummaryPreview) != "" {
				b.WriteString(strings.TrimSpace(s.SummaryPreview))
				b.WriteString("\n")
			}
			b.WriteString(fmt.Sprintf("Hooks opened: %d\n", s.HooksOpened))
			b.WriteString(fmt.Sprintf("Hooks resolved: %d\n\n", s.HooksResolved))
		}
	}

	b.WriteString("## Still Open Hooks\n")
	if len(openHooks) == 0 {
		b.WriteString("- None\n")
	} else {
		for _, h := range openHooks {
			b.WriteString(fmt.Sprintf("- %s (opened session %d)\n", h.Hook, h.SessionOpened))
		}
	}
	b.WriteString("\n")

	b.WriteString("## World State Snapshot\n")
	if len(worldFlags) == 0 {
		b.WriteString("- None\n")
	} else {
		keys := make([]string, 0, len(worldFlags))
		for k := range worldFlags {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			b.WriteString(fmt.Sprintf("- %s: %s\n", k, worldFlags[k]))
		}
	}

	return strings.TrimSpace(b.String())
}
