// Package types contains shared structs for the D&D campaign MCP server.
package types

// Character represents a player character or NPC.
type Character struct {
	ID            string            `json:"id"`
	CampaignID    string            `json:"campaign_id"`
	Name          string            `json:"name"`
	Type          string            `json:"type"` // "pc" | "npc"
	Class         string            `json:"class,omitempty"`
	Race          string            `json:"race,omitempty"`
	Level         int               `json:"level"`
	HP            HP                `json:"hp"`
	Stats         Stats             `json:"stats"`
	Gold          int               `json:"gold"`
	Backstory     string            `json:"backstory,omitempty"`
	Inventory     []string          `json:"inventory"`
	Conditions    []string          `json:"conditions"`
	Relationships map[string]string `json:"relationships"`
	PlotFlags     []string          `json:"plot_flags"`
	Notes         string            `json:"notes,omitempty"` // DM-only
	Status        string            `json:"status"`
	UpdatedAt     string            `json:"updated_at"`
}

// HP represents hit points.
type HP struct {
	Current int `json:"current"`
	Max     int `json:"max"`
}

// Stats represents character ability scores.
type Stats struct {
	STR int `json:"STR"`
	DEX int `json:"DEX"`
	CON int `json:"CON"`
	INT int `json:"INT"`
	WIS int `json:"WIS"`
	CHA int `json:"CHA"`
}

// PlotEvent represents a narrative event in the campaign.
type PlotEvent struct {
	ID           string   `json:"id"`
	CampaignID   string   `json:"campaign_id"`
	Session      int      `json:"session"`
	Summary      string   `json:"summary"`
	NPCs         []string `json:"npcs_involved"`
	PCs          []string `json:"pcs_involved"`
	Consequences string   `json:"consequences,omitempty"`
	Tags         []string `json:"tags"`
	CreatedAt    string   `json:"created_at"`
}

// Hook represents an unresolved plot thread.
type Hook struct {
	ID            string `json:"id"`
	CampaignID    string `json:"campaign_id"`
	Hook          string `json:"hook"`
	SessionOpened int    `json:"session_opened"`
	EventID       string `json:"event_id,omitempty"`
	Resolved      bool   `json:"resolved"`
	Resolution    string `json:"resolution,omitempty"`
}

// RollResult contains the full breakdown of a dice roll.
type RollResult struct {
	Total     int    `json:"total"`
	Rolls     []int  `json:"rolls"`
	Kept      []int  `json:"kept,omitempty"`
	Modifier  int    `json:"modifier"`
	Notation  string `json:"notation"`
	RollID    string `json:"roll_id"`
	Timestamp string `json:"timestamp"`
}

// CharacterSummary is a condensed view of a character.
type CharacterSummary struct {
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	Class      string   `json:"class"`
	Level      int      `json:"level"`
	HP         HP       `json:"hp"`
	Status     string   `json:"status"`
	Conditions []string `json:"conditions"`
}

// SessionBrief is the context package for starting a session.
type SessionBrief struct {
	SessionBrief       string             `json:"session_brief"`
	ActiveCharacters   []CharacterSummary `json:"active_characters"`
	OpenHooks          []Hook             `json:"open_hooks"`
	WorldFlags         map[string]string  `json:"world_flags"`
	LastSessionNumber  int                `json:"last_session_number"`
	LastSessionSummary string             `json:"last_session_summary"`
	DMSystemPrompt     string             `json:"dm_system_prompt"`
}

// Campaign represents a D&D campaign.
type Campaign struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description,omitempty"`
	CreatedAt      string `json:"created_at"`
	CurrentSession int    `json:"current_session"`
}

// Session represents a play session record.
type Session struct {
	CampaignID    string `json:"campaign_id"`
	Session       int    `json:"session"`
	Summary       string `json:"summary,omitempty"`
	DMNotes       string `json:"dm_notes,omitempty"`
	HooksOpened   int    `json:"hooks_opened"`
	HooksResolved int    `json:"hooks_resolved"`
	CreatedAt     string `json:"created_at"`
}

// SessionMeta is a summary view of a session.
type SessionMeta struct {
	Session        int    `json:"session"`
	Date           string `json:"date"`
	SummaryPreview string `json:"summary_preview"`
	HooksOpened    int    `json:"hooks_opened"`
	HooksResolved  int    `json:"hooks_resolved"`
}

// Checkpoint represents a mid-session snapshot.
type Checkpoint struct {
	ID         string `json:"id"`
	CampaignID string `json:"campaign_id"`
	Session    int    `json:"session"`
	Note       string `json:"note"`
	CreatedAt  string `json:"created_at"`
}

// RollRecord is a logged dice roll.
type RollRecord struct {
	ID           string `json:"id"`
	CampaignID   string `json:"campaign_id"`
	Session      int    `json:"session,omitempty"`
	Character    string `json:"character,omitempty"`
	Notation     string `json:"notation"`
	Total        int    `json:"total"`
	Rolls        []int  `json:"rolls"`
	Kept         []int  `json:"kept,omitempty"`
	Modifier     int    `json:"modifier"`
	Reason       string `json:"reason,omitempty"`
	Advantage    bool   `json:"advantage"`
	Disadvantage bool   `json:"disadvantage"`
	CreatedAt    string `json:"created_at"`
}

// ContestedRollResult contains the outcome of a contested roll.
type ContestedRollResult struct {
	Winner         string     `json:"winner"`
	AttackerResult RollResult `json:"attacker_result"`
	DefenderResult RollResult `json:"defender_result"`
	Margin         int        `json:"margin"`
}

// SavingThrowResult contains the outcome of a saving throw.
type SavingThrowResult struct {
	Total    int    `json:"total"`
	Rolled   int    `json:"rolled"`
	Modifier int    `json:"modifier"`
	DC       int    `json:"dc"`
	Success  bool   `json:"success"`
	RollID   string `json:"roll_id"`
}

// RelationshipEdge describes a directional relationship between characters.
type RelationshipEdge struct {
	Source     string `json:"source"`
	SourceType string `json:"source_type"`
	Target     string `json:"target"`
	TargetType string `json:"target_type"`
	Relation   string `json:"relation"`
}
