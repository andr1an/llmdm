package session

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/andr1an/llmdm/internal/types"
)

const briefTemplate = `## Campaign Brief - Session {{.Session}}
*{{.CampaignName}}*

### Active Party
{{- if .Characters}}
{{- range .Characters}}
- **{{.Name}}** ({{if .Class}}{{.Class}}{{else}}Unknown{{end}} Lv.{{.Level}}) - HP {{.HP.Current}}/{{.HP.Max}}{{if .Conditions}} - *{{join .Conditions ", "}}*{{end}}
{{- end}}
{{- else}}
- No active player characters recorded.
{{- end}}

### Story So Far
{{.RecentSummaries}}

### Open Plot Threads
{{- if .OpenHooks}}
{{- range .OpenHooks}}
- {{.Hook}} *(opened session {{.SessionOpened}})*
{{- end}}
{{- else}}
- No unresolved hooks.
{{- end}}

### World State
{{- if .WorldFlags}}
{{- range .WorldFlagLines}}
- {{.}}
{{- end}}
{{- else}}
- No world flags recorded.
{{- end}}

### Last Session Ended
{{.LastCheckpoint}}
`

// BriefData contains fields used by the brief template.
type BriefData struct {
	Session         int
	CampaignName    string
	Characters      []types.CharacterSummary
	RecentSummaries string
	OpenHooks       []types.Hook
	WorldFlags      map[string]string
	WorldFlagLines  []string
	LastCheckpoint  string
}

// RenderBrief renders the markdown session brief from template data.
func RenderBrief(data BriefData) (string, error) {
	data.WorldFlagLines = flattenWorldFlags(data.WorldFlags)

	tmpl, err := template.New("session-brief").Funcs(template.FuncMap{
		"join": strings.Join,
	}).Parse(briefTemplate)
	if err != nil {
		return "", fmt.Errorf("parse brief template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("render brief template: %w", err)
	}

	return strings.TrimSpace(buf.String()), nil
}

// FormatRecentSummaries creates markdown for prior session summaries.
func FormatRecentSummaries(sessions []types.SessionMeta) string {
	if len(sessions) == 0 {
		return "No prior session summaries yet."
	}

	lines := make([]string, 0, len(sessions))
	for _, s := range sessions {
		preview := strings.TrimSpace(s.SummaryPreview)
		if preview == "" {
			preview = "No summary recorded."
		}
		lines = append(lines, fmt.Sprintf("- Session %d: %s", s.Session, preview))
	}
	return strings.Join(lines, "\n")
}

// RenderQuickBrief returns a compact markdown brief for fast context restore.
func RenderQuickBrief(campaignName string, latestSession *types.Session, characters []types.CharacterSummary, hooks []types.Hook, flags map[string]string) string {
	var b strings.Builder

	b.WriteString("## Quick Session Brief\n")
	if campaignName != "" {
		b.WriteString("Campaign: ")
		b.WriteString(campaignName)
		b.WriteString("\n")
	}

	if latestSession != nil {
		b.WriteString(fmt.Sprintf("Last session: %d\n", latestSession.Session))
		if latestSession.Summary != "" {
			b.WriteString(strings.TrimSpace(latestSession.Summary))
			b.WriteString("\n")
		}
	}

	b.WriteString("\nActive PCs: ")
	if len(characters) == 0 {
		b.WriteString("none")
	} else {
		names := make([]string, 0, len(characters))
		for _, c := range characters {
			names = append(names, c.Name)
		}
		b.WriteString(strings.Join(names, ", "))
	}

	b.WriteString("\nOpen hooks: ")
	if len(hooks) == 0 {
		b.WriteString("none")
	} else {
		max := len(hooks)
		if max > 5 {
			max = 5
		}
		hookTexts := make([]string, 0, max)
		for i := 0; i < max; i++ {
			hookTexts = append(hookTexts, hooks[i].Hook)
		}
		b.WriteString(strings.Join(hookTexts, "; "))
	}

	if len(flags) > 0 {
		lines := flattenWorldFlags(flags)
		max := len(lines)
		if max > 8 {
			max = 8
		}
		b.WriteString("\nWorld flags: ")
		b.WriteString(strings.Join(lines[:max], "; "))
	}

	return strings.TrimSpace(b.String())
}

func flattenWorldFlags(flags map[string]string) []string {
	if len(flags) == 0 {
		return nil
	}
	lines := make([]string, 0, len(flags))
	for k, v := range flags {
		lines = append(lines, fmt.Sprintf("%s: %s", k, v))
	}
	return lines
}
