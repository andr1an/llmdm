// Package main is the entry point for the D&D Campaign MCP server.
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/andr1an/llmdm/config"
	"github.com/andr1an/llmdm/internal/db"
	"github.com/andr1an/llmdm/internal/dice"
	"github.com/andr1an/llmdm/internal/dm"
	"github.com/andr1an/llmdm/internal/memory"
	"github.com/andr1an/llmdm/internal/session"
	"github.com/andr1an/llmdm/internal/types"
)

// Server holds the MCP server and its dependencies.
type Server struct {
	mcp    *server.MCPServer
	cfg    *config.Config
	dbPath string
	logger *slog.Logger
}

func main() {
	if len(os.Args) < 2 || os.Args[1] != "serve" {
		fmt.Println("Usage: dnd-mcp serve")
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	logger, err := newJSONLogger(cfg.LogLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid LOG_LEVEL %q: %v. Falling back to info.\n", cfg.LogLevel, err)
		logger, _ = newJSONLogger("info")
	}
	slog.SetDefault(logger)

	srv := &Server{
		cfg:    cfg,
		dbPath: cfg.DBPath,
		logger: logger,
	}

	srv.mcp = server.NewMCPServer(
		"D&D Campaign Memory",
		"1.0.0",
		server.WithToolCapabilities(false),
		server.WithRecovery(),
	)

	srv.registerTools()
	srv.logger.Info(
		"server startup",
		"transport", strings.ToLower(strings.TrimSpace(cfg.Transport)),
		"http_addr", cfg.HTTPAddr,
		"http_endpoint", cfg.HTTPEndpoint,
		"log_level", strings.ToLower(strings.TrimSpace(cfg.LogLevel)),
	)

	switch strings.ToLower(strings.TrimSpace(cfg.Transport)) {
	case "http", "streamable-http", "streamable_http":
		if err := srv.serveHTTP(); err != nil {
			srv.logger.Error("http server error", "error", err)
			os.Exit(1)
		}
	default:
		if err := server.ServeStdio(srv.mcp); err != nil {
			srv.logger.Error("stdio server error", "error", err)
			os.Exit(1)
		}
	}
}

func newJSONLogger(levelText string) (*slog.Logger, error) {
	var level slog.Level
	switch strings.ToLower(strings.TrimSpace(levelText)) {
	case "debug":
		level = slog.LevelDebug
	case "", "info":
		level = slog.LevelInfo
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		return nil, fmt.Errorf("supported levels are debug, info, warn, error")
	}
	return slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: level})), nil
}

func (s *Server) log() *slog.Logger {
	if s != nil && s.logger != nil {
		return s.logger
	}
	return slog.Default()
}

func (s *Server) registerTools() {
	// roll tool
	rollTool := mcp.NewTool("roll",
		mcp.WithDescription("Roll dice using standard D&D notation. Supports NdM, NdM+K, NdM-K, NdMkhX (keep highest), NdMklX (keep lowest). All rolls are logged."),
		mcp.WithString("campaign_id",
			mcp.Required(),
			mcp.Description("Campaign ID for logging the roll"),
		),
		mcp.WithString("notation",
			mcp.Required(),
			mcp.Description("Dice notation (e.g., '1d20', '2d6+3', '4d6kh3')"),
		),
		mcp.WithString("reason",
			mcp.Description("Reason for the roll (e.g., 'Stealth check')"),
		),
		mcp.WithString("character",
			mcp.Description("Character making the roll"),
		),
		mcp.WithNumber("session",
			mcp.Description("Current session number"),
		),
		mcp.WithBoolean("advantage",
			mcp.Description("Roll with advantage (2d20 keep highest)"),
		),
		mcp.WithBoolean("disadvantage",
			mcp.Description("Roll with disadvantage (2d20 keep lowest)"),
		),
	)
	s.mcp.AddTool(rollTool, s.handleRoll)

	// roll_contested tool
	contestedTool := mcp.NewTool("roll_contested",
		mcp.WithDescription("Roll a contested check between attacker and defender. Both rolls are logged."),
		mcp.WithString("campaign_id",
			mcp.Required(),
			mcp.Description("Campaign ID"),
		),
		mcp.WithString("attacker",
			mcp.Required(),
			mcp.Description("Attacker's name"),
		),
		mcp.WithString("defender",
			mcp.Required(),
			mcp.Description("Defender's name"),
		),
		mcp.WithString("attacker_notation",
			mcp.Required(),
			mcp.Description("Dice notation for attacker (e.g., '1d20+5')"),
		),
		mcp.WithString("defender_notation",
			mcp.Required(),
			mcp.Description("Dice notation for defender (e.g., '1d20+3')"),
		),
		mcp.WithString("contest_type",
			mcp.Required(),
			mcp.Description("Type of contest (e.g., 'Attack vs AC', 'Deception vs Insight')"),
		),
		mcp.WithNumber("session",
			mcp.Description("Current session number"),
		),
	)
	s.mcp.AddTool(contestedTool, s.handleRollContested)

	// roll_saving_throw tool
	savingThrowTool := mcp.NewTool("roll_saving_throw",
		mcp.WithDescription("Roll a saving throw and check against a DC."),
		mcp.WithString("campaign_id",
			mcp.Required(),
			mcp.Description("Campaign ID"),
		),
		mcp.WithString("character",
			mcp.Required(),
			mcp.Description("Character making the saving throw"),
		),
		mcp.WithString("stat",
			mcp.Required(),
			mcp.Description("Ability score to use"),
			mcp.Enum("STR", "DEX", "CON", "INT", "WIS", "CHA"),
		),
		mcp.WithNumber("modifier",
			mcp.Required(),
			mcp.Description("Total modifier for the saving throw"),
		),
		mcp.WithNumber("dc",
			mcp.Required(),
			mcp.Description("Difficulty class to beat"),
		),
		mcp.WithString("reason",
			mcp.Description("Reason for the saving throw"),
		),
		mcp.WithNumber("session",
			mcp.Description("Current session number"),
		),
	)
	s.mcp.AddTool(savingThrowTool, s.handleRollSavingThrow)

	// get_roll_history tool
	historyTool := mcp.NewTool("get_roll_history",
		mcp.WithDescription("Retrieve dice roll history for a campaign."),
		mcp.WithString("campaign_id",
			mcp.Required(),
			mcp.Description("Campaign ID"),
		),
		mcp.WithString("character",
			mcp.Description("Filter by character name"),
		),
		mcp.WithNumber("session",
			mcp.Description("Filter by session number"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of rolls to return (default: 50)"),
		),
	)
	s.mcp.AddTool(historyTool, s.handleGetRollHistory)

	// === Campaign Memory Tools ===

	// create_campaign tool
	createCampaignTool := mcp.NewTool("create_campaign",
		mcp.WithDescription("Create a new D&D campaign. Returns the campaign ID and database path."),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Campaign name (e.g., 'Lost Mines of Phandelver')"),
		),
		mcp.WithString("description",
			mcp.Description("Brief setting description"),
		),
	)
	s.mcp.AddTool(createCampaignTool, s.handleCreateCampaign)

	// save_character tool
	saveCharacterTool := mcp.NewTool("save_character",
		mcp.WithDescription("Create or fully replace a character (PC or NPC). Use update_character for partial updates."),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign ID")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Character name")),
		mcp.WithString("type", mcp.Required(), mcp.Description("Character type"), mcp.Enum("pc", "npc")),
		mcp.WithString("class", mcp.Description("Character class (e.g., 'Paladin')")),
		mcp.WithString("race", mcp.Description("Character race (e.g., 'Human')")),
		mcp.WithNumber("level", mcp.Description("Character level (default: 1)")),
		mcp.WithNumber("hp_current", mcp.Required(), mcp.Description("Current hit points")),
		mcp.WithNumber("hp_max", mcp.Required(), mcp.Description("Maximum hit points")),
		mcp.WithNumber("str", mcp.Description("Strength score")),
		mcp.WithNumber("dex", mcp.Description("Dexterity score")),
		mcp.WithNumber("con", mcp.Description("Constitution score")),
		mcp.WithNumber("int_stat", mcp.Description("Intelligence score")),
		mcp.WithNumber("wis", mcp.Description("Wisdom score")),
		mcp.WithNumber("cha", mcp.Description("Charisma score")),
		mcp.WithString("backstory", mcp.Description("Character backstory")),
		mcp.WithString("status", mcp.Description("Character status"), mcp.Enum("active", "dead", "missing", "retired")),
		mcp.WithString("notes", mcp.Description("DM private notes (for NPCs)")),
	)
	s.mcp.AddTool(saveCharacterTool, s.handleSaveCharacter)

	// update_character tool
	updateCharacterTool := mcp.NewTool("update_character",
		mcp.WithDescription("Patch specific fields on a character. Only provided fields are updated."),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign ID")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Character name")),
		mcp.WithNumber("hp_current", mcp.Description("New current hit points")),
		mcp.WithNumber("level", mcp.Description("New level")),
		mcp.WithString("status", mcp.Description("New status"), mcp.Enum("active", "dead", "missing", "retired")),
		mcp.WithString("notes", mcp.Description("New DM notes")),
	)
	s.mcp.AddTool(updateCharacterTool, s.handleUpdateCharacter)

	// get_character tool
	getCharacterTool := mcp.NewTool("get_character",
		mcp.WithDescription("Get full character sheet by name."),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign ID")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Character name")),
	)
	s.mcp.AddTool(getCharacterTool, s.handleGetCharacter)

	// list_characters tool
	listCharactersTool := mcp.NewTool("list_characters",
		mcp.WithDescription("List characters in a campaign, optionally filtered by type and status."),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign ID")),
		mcp.WithString("type", mcp.Description("Filter by type"), mcp.Enum("pc", "npc")),
		mcp.WithString("status", mcp.Description("Filter by status"), mcp.Enum("active", "dead", "missing", "retired")),
	)
	s.mcp.AddTool(listCharactersTool, s.handleListCharacters)

	// save_plot_event tool
	savePlotEventTool := mcp.NewTool("save_plot_event",
		mcp.WithDescription("Record a narrative event in the campaign. Creates plot hooks if provided."),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign ID")),
		mcp.WithNumber("session", mcp.Required(), mcp.Description("Session number")),
		mcp.WithString("summary", mcp.Required(), mcp.Description("2-4 sentence narrative description")),
		mcp.WithString("consequences", mcp.Description("What changed in the world")),
	)
	s.mcp.AddTool(savePlotEventTool, s.handleSavePlotEvent)

	// list_open_hooks tool
	listOpenHooksTool := mcp.NewTool("list_open_hooks",
		mcp.WithDescription("List all unresolved plot threads for a campaign."),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign ID")),
	)
	s.mcp.AddTool(listOpenHooksTool, s.handleListOpenHooks)

	// resolve_hook tool
	resolveHookTool := mcp.NewTool("resolve_hook",
		mcp.WithDescription("Mark a plot hook as resolved."),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign ID")),
		mcp.WithString("hook_id", mcp.Required(), mcp.Description("Hook ID to resolve")),
		mcp.WithString("resolution", mcp.Required(), mcp.Description("How the hook was resolved")),
	)
	s.mcp.AddTool(resolveHookTool, s.handleResolveHook)

	// set_world_flag tool
	setWorldFlagTool := mcp.NewTool("set_world_flag",
		mcp.WithDescription("Set a key/value world state flag. Use 'true'/'false' for booleans."),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign ID")),
		mcp.WithString("key", mcp.Required(), mcp.Description("Flag key")),
		mcp.WithString("value", mcp.Required(), mcp.Description("Flag value")),
	)
	s.mcp.AddTool(setWorldFlagTool, s.handleSetWorldFlag)

	// get_world_flags tool
	getWorldFlagsTool := mcp.NewTool("get_world_flags",
		mcp.WithDescription("Get all world state flags for a campaign."),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign ID")),
	)
	s.mcp.AddTool(getWorldFlagsTool, s.handleGetWorldFlags)

	// === Session Management Tools ===

	// start_session tool
	startSessionTool := mcp.NewTool("start_session",
		mcp.WithDescription("Load campaign state and render a session brief for DM context."),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign ID")),
		mcp.WithNumber("session", mcp.Required(), mcp.Description("Session number to start")),
		mcp.WithNumber("recent_sessions", mcp.Description("How many prior sessions to include (default: 3)")),
	)
	s.mcp.AddTool(startSessionTool, s.handleStartSession)

	// end_session tool
	endSessionTool := mcp.NewTool("end_session",
		mcp.WithDescription("Compress and store end-of-session narrative summary."),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign ID")),
		mcp.WithNumber("session", mcp.Required(), mcp.Description("Session number to end")),
		mcp.WithString("raw_events", mcp.Required(), mcp.Description("Full narrative event log for the session")),
		mcp.WithString("dm_notes", mcp.Description("Optional DM notes")),
	)
	s.mcp.AddTool(endSessionTool, s.handleEndSession)

	// checkpoint tool
	checkpointTool := mcp.NewTool("checkpoint",
		mcp.WithDescription("Save a mid-session checkpoint note."),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign ID")),
		mcp.WithNumber("session", mcp.Required(), mcp.Description("Current session number")),
		mcp.WithString("note", mcp.Required(), mcp.Description("Checkpoint note")),
	)
	s.mcp.AddTool(checkpointTool, s.handleCheckpoint)

	// get_session_brief tool
	getSessionBriefTool := mcp.NewTool("get_session_brief",
		mcp.WithDescription("Get a compact markdown briefing for quick context restoration."),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign ID")),
	)
	s.mcp.AddTool(getSessionBriefTool, s.handleGetSessionBrief)

	// list_sessions tool
	listSessionsTool := mcp.NewTool("list_sessions",
		mcp.WithDescription("List historical sessions with summary previews."),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign ID")),
	)
	s.mcp.AddTool(listSessionsTool, s.handleListSessions)

	// get_npc_relationships tool
	relationshipTool := mcp.NewTool("get_npc_relationships",
		mcp.WithDescription("Query relationship edges involving NPCs. Optionally filter to one NPC name."),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign ID")),
		mcp.WithString("npc_name", mcp.Description("Optional NPC name filter")),
	)
	s.mcp.AddTool(relationshipTool, s.handleGetNPCRelationships)

	// export_session_recap tool
	exportRecapTool := mcp.NewTool("export_session_recap",
		mcp.WithDescription("Export a markdown recap for a campaign across all or selected sessions."),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign ID")),
		mcp.WithNumber("from_session", mcp.Description("Optional lower inclusive session bound")),
		mcp.WithNumber("to_session", mcp.Description("Optional upper inclusive session bound")),
	)
	s.mcp.AddTool(exportRecapTool, s.handleExportSessionRecap)
}

func (s *Server) serveHTTP() error {
	s.log().Debug("configuring streamable HTTP server")
	mcpHandler := server.NewStreamableHTTPServer(
		s.mcp,
		server.WithEndpointPath(s.cfg.HTTPEndpoint),
		server.WithHeartbeatInterval(30*time.Second),
		server.WithStateLess(false),
	)

	endpoint := "/" + strings.Trim(s.cfg.HTTPEndpoint, "/")
	mux := http.NewServeMux()
	mux.Handle(endpoint, mcpHandler)
	if endpoint != "/" {
		mux.Handle(endpoint+"/", mcpHandler)
	}
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "ok",
			"transport": "streamable-http",
			"endpoint":  endpoint,
			"time":      time.Now().UTC().Format(time.RFC3339),
		})
	})

	httpServer := &http.Server{
		Addr:    s.cfg.HTTPAddr,
		Handler: mux,
	}
	s.log().Info("serving MCP over streamable HTTP", "addr", s.cfg.HTTPAddr, "endpoint", endpoint)
	return httpServer.ListenAndServe()
}

func (s *Server) openCampaignDB(campaignID string) (*db.DB, error) {
	dbPath := db.CampaignDBPath(s.dbPath, campaignID)
	s.log().Debug("opening campaign database", "campaign_id", campaignID, "db_path", dbPath)
	database, err := db.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err := database.Migrate(); err != nil {
		database.Close()
		return nil, fmt.Errorf("migrate database: %w", err)
	}
	if err := s.reconcileCampaignID(database.DB, campaignID); err != nil {
		database.Close()
		return nil, fmt.Errorf("reconcile campaign id: %w", err)
	}
	s.log().Debug("campaign database ready", "campaign_id", campaignID)
	return database, nil
}

func (s *Server) reconcileCampaignID(rawDB *sql.DB, campaignID string) error {
	if strings.TrimSpace(campaignID) == "" {
		return nil
	}

	var exactMatch int
	if err := rawDB.QueryRow(`SELECT COUNT(1) FROM campaigns WHERE id = ?`, campaignID).Scan(&exactMatch); err != nil {
		return fmt.Errorf("check campaign id existence: %w", err)
	}
	if exactMatch > 0 {
		return nil
	}

	var legacyID string
	var name string
	var description sql.NullString
	var createdAt string
	var currentSession int
	if err := rawDB.QueryRow(`
		SELECT id, name, description, created_at, current_session
		FROM campaigns
		LIMIT 1
	`).Scan(&legacyID, &name, &description, &createdAt, &currentSession); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("load legacy campaign: %w", err)
	}

	var campaignCount int
	if err := rawDB.QueryRow(`SELECT COUNT(1) FROM campaigns`).Scan(&campaignCount); err != nil {
		return fmt.Errorf("count campaigns: %w", err)
	}
	if campaignCount != 1 || legacyID == campaignID {
		return nil
	}

	desc := interface{}(nil)
	if description.Valid {
		desc = description.String
	}

	tx, err := rawDB.Begin()
	if err != nil {
		return fmt.Errorf("begin campaign id reconciliation tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`
		INSERT INTO campaigns (id, name, description, created_at, current_session)
		VALUES (?, ?, ?, ?, ?)
	`, campaignID, name, desc, createdAt, currentSession); err != nil {
		return fmt.Errorf("insert reconciled campaign: %w", err)
	}

	tables := []string{"characters", "plot_events", "plot_hooks", "world_flags", "roll_log", "sessions", "checkpoints"}
	for _, table := range tables {
		query := fmt.Sprintf("UPDATE %s SET campaign_id = ? WHERE campaign_id = ?", table)
		if _, err := tx.Exec(query, campaignID, legacyID); err != nil {
			return fmt.Errorf("update campaign references in %s: %w", table, err)
		}
	}

	if _, err := tx.Exec(`DELETE FROM campaigns WHERE id = ?`, legacyID); err != nil {
		return fmt.Errorf("delete legacy campaign row: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit campaign id reconciliation tx: %w", err)
	}

	s.log().Info("reconciled legacy campaign id", "legacy_id", legacyID, "campaign_id", campaignID)
	return nil
}

func (s *Server) handleRoll(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	campaignID, err := req.RequireString("campaign_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	notation, err := req.RequireString("notation")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	reason := req.GetString("reason", "")
	character := req.GetString("character", "")
	session := int(req.GetFloat("session", 0))
	advantage := req.GetBool("advantage", false)
	disadvantage := req.GetBool("disadvantage", false)
	s.log().Debug(
		"roll request",
		"campaign_id", campaignID,
		"notation", notation,
		"character", character,
		"session", session,
		"advantage", advantage,
		"disadvantage", disadvantage,
	)

	var result types.RollResult

	if advantage && disadvantage {
		// Advantage and disadvantage cancel out
		result, err = dice.Roll(notation)
	} else if advantage {
		// Parse notation to get modifier
		parsed, parseErr := dice.Parse(notation)
		if parseErr != nil {
			return mcp.NewToolResultError(parseErr.Error()), nil
		}
		result = dice.RollWithAdvantage(parsed.Modifier)
	} else if disadvantage {
		parsed, parseErr := dice.Parse(notation)
		if parseErr != nil {
			return mcp.NewToolResultError(parseErr.Error()), nil
		}
		result = dice.RollWithDisadvantage(parsed.Modifier)
	} else {
		result, err = dice.Roll(notation)
	}

	if err != nil {
		s.log().Debug("roll failed", "campaign_id", campaignID, "notation", notation, "error", err)
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Log the roll
	database, err := s.openCampaignDB(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer database.Close()

	logger := dice.NewLogger(database.DB)
	if err := logger.Log(campaignID, session, character, reason, result, advantage, disadvantage); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("roll succeeded but logging failed: %v", err)), nil
	}
	s.log().Debug("roll logged", "campaign_id", campaignID, "roll_id", result.RollID, "total", result.Total)

	jsonResult, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(jsonResult)), nil
}

func (s *Server) handleRollContested(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	campaignID, err := req.RequireString("campaign_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	attacker, err := req.RequireString("attacker")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	defender, err := req.RequireString("defender")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	attackerNotation, err := req.RequireString("attacker_notation")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	defenderNotation, err := req.RequireString("defender_notation")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	contestType, err := req.RequireString("contest_type")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	session := int(req.GetFloat("session", 0))
	s.log().Debug(
		"contested roll request",
		"campaign_id", campaignID,
		"attacker", attacker,
		"defender", defender,
		"contest_type", contestType,
		"session", session,
	)

	// Roll both
	attackerResult, err := dice.Roll(attackerNotation)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("attacker roll failed: %v", err)), nil
	}

	defenderResult, err := dice.Roll(defenderNotation)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("defender roll failed: %v", err)), nil
	}

	// Determine winner
	var winner string
	margin := attackerResult.Total - defenderResult.Total
	if margin > 0 {
		winner = attacker
	} else if margin < 0 {
		winner = defender
		margin = -margin
	} else {
		winner = "tie"
		margin = 0
	}

	result := types.ContestedRollResult{
		Winner:         winner,
		AttackerResult: attackerResult,
		DefenderResult: defenderResult,
		Margin:         margin,
	}

	// Log both rolls
	database, err := s.openCampaignDB(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer database.Close()

	logger := dice.NewLogger(database.DB)
	logger.Log(campaignID, session, attacker, contestType+" (attacker)", attackerResult, false, false)
	logger.Log(campaignID, session, defender, contestType+" (defender)", defenderResult, false, false)
	s.log().Debug(
		"contested roll completed",
		"campaign_id", campaignID,
		"winner", winner,
		"margin", margin,
		"attacker_total", attackerResult.Total,
		"defender_total", defenderResult.Total,
	)

	jsonResult, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(jsonResult)), nil
}

func (s *Server) handleRollSavingThrow(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	campaignID, err := req.RequireString("campaign_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	character, err := req.RequireString("character")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	stat, err := req.RequireString("stat")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	modifier, err := req.RequireFloat("modifier")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	dc, err := req.RequireFloat("dc")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	reason := req.GetString("reason", stat+" saving throw")
	session := int(req.GetFloat("session", 0))
	s.log().Debug(
		"saving throw request",
		"campaign_id", campaignID,
		"character", character,
		"stat", stat,
		"modifier", modifier,
		"dc", dc,
		"session", session,
	)

	// Roll 1d20
	notation := "1d20"
	if modifier >= 0 {
		notation += "+" + strconv.Itoa(int(modifier))
	} else {
		notation += strconv.Itoa(int(modifier))
	}

	rollResult, err := dice.Roll(notation)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result := types.SavingThrowResult{
		Total:    rollResult.Total,
		Rolled:   rollResult.Rolls[0],
		Modifier: int(modifier),
		DC:       int(dc),
		Success:  rollResult.Total >= int(dc),
		RollID:   rollResult.RollID,
	}

	// Log the roll
	database, err := s.openCampaignDB(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer database.Close()

	logger := dice.NewLogger(database.DB)
	logger.Log(campaignID, session, character, reason, rollResult, false, false)
	s.log().Debug(
		"saving throw completed",
		"campaign_id", campaignID,
		"character", character,
		"total", result.Total,
		"dc", result.DC,
		"success", result.Success,
		"roll_id", result.RollID,
	)

	jsonResult, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(jsonResult)), nil
}

func (s *Server) handleGetRollHistory(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	campaignID, err := req.RequireString("campaign_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	character := req.GetString("character", "")
	session := int(req.GetFloat("session", 0))
	limit := int(req.GetFloat("limit", 50))
	s.log().Debug(
		"roll history request",
		"campaign_id", campaignID,
		"character", character,
		"session", session,
		"limit", limit,
	)

	database, err := s.openCampaignDB(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer database.Close()

	logger := dice.NewLogger(database.DB)
	records, err := logger.GetHistory(campaignID, character, session, limit)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	s.log().Debug("roll history loaded", "campaign_id", campaignID, "count", len(records))

	jsonResult, _ := json.Marshal(records)
	return mcp.NewToolResultText(string(jsonResult)), nil
}

// === Campaign Memory Handlers ===

func (s *Server) handleCreateCampaign(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, err := req.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	description := req.GetString("description", "")

	// Create a new campaign DB
	campaignID := generateCampaignID(name)
	database, err := s.openCampaignDB(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer database.Close()

	store := memory.NewStore(database.DB)
	campaign, err := store.CreateCampaignWithID(campaignID, name, description)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result := map[string]interface{}{
		"campaign_id": campaignID,
		"db_path":     database.Path(),
		"campaign":    campaign,
	}
	jsonResult, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(jsonResult)), nil
}

func (s *Server) handleSaveCharacter(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	campaignID, err := req.RequireString("campaign_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	name, err := req.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	charType, err := req.RequireString("type")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	hpCurrent, err := req.RequireFloat("hp_current")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	hpMax, err := req.RequireFloat("hp_max")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	char := &types.Character{
		CampaignID: campaignID,
		Name:       name,
		Type:       charType,
		Class:      req.GetString("class", ""),
		Race:       req.GetString("race", ""),
		Level:      int(req.GetFloat("level", 1)),
		HP: types.HP{
			Current: int(hpCurrent),
			Max:     int(hpMax),
		},
		Stats: types.Stats{
			STR: int(req.GetFloat("str", 10)),
			DEX: int(req.GetFloat("dex", 10)),
			CON: int(req.GetFloat("con", 10)),
			INT: int(req.GetFloat("int_stat", 10)),
			WIS: int(req.GetFloat("wis", 10)),
			CHA: int(req.GetFloat("cha", 10)),
		},
		Backstory:     req.GetString("backstory", ""),
		Status:        req.GetString("status", "active"),
		Notes:         req.GetString("notes", ""),
		Inventory:     []string{},
		Conditions:    []string{},
		PlotFlags:     []string{},
		Relationships: map[string]string{},
	}

	database, err := s.openCampaignDB(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer database.Close()

	store := memory.NewStore(database.DB)
	if err := store.SaveCharacter(char); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result := map[string]interface{}{
		"success":      true,
		"character_id": char.ID,
	}
	jsonResult, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(jsonResult)), nil
}

func (s *Server) handleUpdateCharacter(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	campaignID, err := req.RequireString("campaign_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	name, err := req.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	update := memory.CharacterUpdate{}
	args := req.GetArguments()

	// Check for optional fields
	if hp, ok := args["hp_current"]; ok {
		if hpFloat, ok := hp.(float64); ok {
			hpInt := int(hpFloat)
			update.HPCurrent = &hpInt
		}
	}
	if level, ok := args["level"]; ok {
		if levelFloat, ok := level.(float64); ok {
			levelInt := int(levelFloat)
			update.Level = &levelInt
		}
	}
	if status, ok := args["status"]; ok {
		if statusStr, ok := status.(string); ok {
			update.Status = &statusStr
		}
	}
	if notes, ok := args["notes"]; ok {
		if notesStr, ok := notes.(string); ok {
			update.Notes = &notesStr
		}
	}

	database, err := s.openCampaignDB(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer database.Close()

	store := memory.NewStore(database.DB)
	updatedFields, err := store.UpdateCharacter(campaignID, name, update)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result := map[string]interface{}{
		"success":        true,
		"updated_fields": updatedFields,
	}
	jsonResult, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(jsonResult)), nil
}

func (s *Server) handleGetCharacter(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	campaignID, err := req.RequireString("campaign_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	name, err := req.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	database, err := s.openCampaignDB(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer database.Close()

	store := memory.NewStore(database.DB)
	char, err := store.GetCharacter(campaignID, name)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if char == nil {
		return mcp.NewToolResultError(fmt.Sprintf("character not found: %s", name)), nil
	}

	jsonResult, _ := json.Marshal(char)
	return mcp.NewToolResultText(string(jsonResult)), nil
}

func (s *Server) handleListCharacters(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	campaignID, err := req.RequireString("campaign_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	charType := req.GetString("type", "")
	status := req.GetString("status", "")

	database, err := s.openCampaignDB(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer database.Close()

	store := memory.NewStore(database.DB)
	characters, err := store.ListCharacters(campaignID, charType, status)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	jsonResult, _ := json.Marshal(characters)
	return mcp.NewToolResultText(string(jsonResult)), nil
}

func (s *Server) handleSavePlotEvent(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	campaignID, err := req.RequireString("campaign_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	session, err := req.RequireFloat("session")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	summary, err := req.RequireString("summary")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	event := &types.PlotEvent{
		CampaignID:   campaignID,
		Session:      int(session),
		Summary:      summary,
		Consequences: req.GetString("consequences", ""),
		NPCs:         []string{},
		PCs:          []string{},
		Tags:         []string{},
	}

	database, err := s.openCampaignDB(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer database.Close()

	store := memory.NewStore(database.DB)
	if err := store.SavePlotEvent(event, nil); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result := map[string]interface{}{
		"success":  true,
		"event_id": event.ID,
	}
	jsonResult, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(jsonResult)), nil
}

func (s *Server) handleListOpenHooks(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	campaignID, err := req.RequireString("campaign_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	database, err := s.openCampaignDB(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer database.Close()

	store := memory.NewStore(database.DB)
	hooks, err := store.ListOpenHooks(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	jsonResult, _ := json.Marshal(hooks)
	return mcp.NewToolResultText(string(jsonResult)), nil
}

func (s *Server) handleResolveHook(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	campaignID, err := req.RequireString("campaign_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	hookID, err := req.RequireString("hook_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resolution, err := req.RequireString("resolution")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	database, err := s.openCampaignDB(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer database.Close()

	store := memory.NewStore(database.DB)
	if err := store.ResolveHook(campaignID, hookID, resolution); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result := map[string]interface{}{
		"success": true,
	}
	jsonResult, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(jsonResult)), nil
}

func (s *Server) handleSetWorldFlag(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	campaignID, err := req.RequireString("campaign_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	key, err := req.RequireString("key")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	value, err := req.RequireString("value")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	database, err := s.openCampaignDB(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer database.Close()

	store := memory.NewStore(database.DB)
	if err := store.SetWorldFlag(campaignID, key, value); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result := map[string]interface{}{
		"success": true,
	}
	jsonResult, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(jsonResult)), nil
}

func (s *Server) handleGetWorldFlags(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	campaignID, err := req.RequireString("campaign_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	database, err := s.openCampaignDB(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer database.Close()

	store := memory.NewStore(database.DB)
	flags, err := store.GetWorldFlags(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	jsonResult, _ := json.Marshal(flags)
	return mcp.NewToolResultText(string(jsonResult)), nil
}

func (s *Server) handleStartSession(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	campaignID, err := req.RequireString("campaign_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	sessionNumberRaw, err := req.RequireFloat("session")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	sessionNumber := int(sessionNumberRaw)
	recentSessions := int(req.GetFloat("recent_sessions", 3))
	if recentSessions <= 0 {
		recentSessions = 3
	}

	database, err := s.openCampaignDB(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer database.Close()

	store := memory.NewStore(database.DB)
	campaign, err := store.GetCampaign(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if campaign == nil {
		return mcp.NewToolResultError(fmt.Sprintf("campaign not found: %s", campaignID)), nil
	}

	activeCharacters, err := store.ListCharacters(campaignID, "pc", "active")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	openHooks, err := store.ListOpenHooks(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	worldFlags, err := store.GetWorldFlags(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	lastSession, err := store.GetLastSessionBefore(campaignID, sessionNumber)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	recent, err := store.ListRecentSessionsBefore(campaignID, sessionNumber, recentSessions)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	lastSessionNumber := 0
	lastSessionSummary := ""
	lastCheckpointText := "No checkpoint recorded."
	if lastSession != nil {
		lastSessionNumber = lastSession.Session
		lastSessionSummary = lastSession.Summary
		latestCheckpoint, cpErr := store.GetLatestCheckpoint(campaignID, lastSession.Session)
		if cpErr != nil {
			return mcp.NewToolResultError(cpErr.Error()), nil
		}
		if latestCheckpoint != nil {
			lastCheckpointText = latestCheckpoint.Note
		}
	}

	briefText, err := session.RenderBrief(session.BriefData{
		Session:         sessionNumber,
		CampaignName:    campaign.Name,
		Characters:      activeCharacters,
		RecentSummaries: session.FormatRecentSummaries(recent),
		OpenHooks:       openHooks,
		WorldFlags:      worldFlags,
		LastCheckpoint:  lastCheckpointText,
	})
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result := types.SessionBrief{
		SessionBrief:       briefText,
		ActiveCharacters:   activeCharacters,
		OpenHooks:          openHooks,
		WorldFlags:         worldFlags,
		LastSessionNumber:  lastSessionNumber,
		LastSessionSummary: lastSessionSummary,
		DMSystemPrompt:     dm.BuildSystemPrompt(briefText, openHooks, worldFlags),
	}

	jsonResult, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(jsonResult)), nil
}

func (s *Server) handleEndSession(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	campaignID, err := req.RequireString("campaign_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	sessionNumberRaw, err := req.RequireFloat("session")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	rawEvents, err := req.RequireString("raw_events")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	dmNotes := req.GetString("dm_notes", "")
	sessionNumber := int(sessionNumberRaw)

	database, err := s.openCampaignDB(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer database.Close()

	store := memory.NewStore(database.DB)
	compressor := session.NewCompressor(s.cfg.AnthropicAPIKey)
	compressedSummary, err := compressor.Compress(ctx, rawEvents)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	hooksOpened, err := store.CountHooksOpenedInSession(campaignID, sessionNumber)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	hooksResolved, err := store.CountHooksResolvedSincePreviousSession(campaignID, sessionNumber)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if err := store.UpsertSession(&types.Session{
		CampaignID:    campaignID,
		Session:       sessionNumber,
		Summary:       compressedSummary,
		DMNotes:       dmNotes,
		HooksOpened:   hooksOpened,
		HooksResolved: hooksResolved,
	}); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if err := store.AdvanceCampaignSession(campaignID, sessionNumber+1); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result := map[string]interface{}{
		"success":            true,
		"compressed_summary": compressedSummary,
		"hooks_opened":       hooksOpened,
		"hooks_resolved":     hooksResolved,
		"characters_updated": []string{},
	}
	jsonResult, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(jsonResult)), nil
}

func (s *Server) handleCheckpoint(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	campaignID, err := req.RequireString("campaign_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	sessionNumberRaw, err := req.RequireFloat("session")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	note, err := req.RequireString("note")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	database, err := s.openCampaignDB(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer database.Close()

	store := memory.NewStore(database.DB)
	checkpoint, err := store.CreateCheckpoint(campaignID, int(sessionNumberRaw), note)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result := map[string]interface{}{
		"success":       true,
		"checkpoint_id": checkpoint.ID,
	}
	jsonResult, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(jsonResult)), nil
}

func (s *Server) handleGetSessionBrief(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	campaignID, err := req.RequireString("campaign_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	database, err := s.openCampaignDB(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer database.Close()

	store := memory.NewStore(database.DB)
	campaign, err := store.GetCampaign(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if campaign == nil {
		return mcp.NewToolResultError(fmt.Sprintf("campaign not found: %s", campaignID)), nil
	}

	latestSession, err := store.GetLatestSession(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	activeCharacters, err := store.ListCharacters(campaignID, "pc", "active")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	openHooks, err := store.ListOpenHooks(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	worldFlags, err := store.GetWorldFlags(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result := map[string]string{
		"brief": session.RenderQuickBrief(campaign.Name, latestSession, activeCharacters, openHooks, worldFlags),
	}
	jsonResult, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(jsonResult)), nil
}

func (s *Server) handleListSessions(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	campaignID, err := req.RequireString("campaign_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	database, err := s.openCampaignDB(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer database.Close()

	store := memory.NewStore(database.DB)
	sessions, err := store.ListSessions(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	jsonResult, _ := json.Marshal(sessions)
	return mcp.NewToolResultText(string(jsonResult)), nil
}

func (s *Server) handleGetNPCRelationships(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	campaignID, err := req.RequireString("campaign_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	npcName := req.GetString("npc_name", "")

	database, err := s.openCampaignDB(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer database.Close()

	store := memory.NewStore(database.DB)
	edges, err := store.QueryNPCRelationships(campaignID, npcName)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	jsonResult, _ := json.Marshal(map[string]interface{}{
		"campaign_id": campaignID,
		"npc_name":    npcName,
		"count":       len(edges),
		"edges":       edges,
	})
	return mcp.NewToolResultText(string(jsonResult)), nil
}

func (s *Server) handleExportSessionRecap(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	campaignID, err := req.RequireString("campaign_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	fromSession := int(req.GetFloat("from_session", 0))
	toSession := int(req.GetFloat("to_session", 0))

	database, err := s.openCampaignDB(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer database.Close()

	store := memory.NewStore(database.DB)
	campaign, err := store.GetCampaign(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if campaign == nil {
		return mcp.NewToolResultError(fmt.Sprintf("campaign not found: %s", campaignID)), nil
	}

	sessions, err := store.ListSessions(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	filtered := make([]types.SessionMeta, 0, len(sessions))
	for _, s := range sessions {
		if fromSession > 0 && s.Session < fromSession {
			continue
		}
		if toSession > 0 && s.Session > toSession {
			continue
		}
		filtered = append(filtered, s)
	}
	sort.Slice(filtered, func(i, j int) bool { return filtered[i].Session < filtered[j].Session })

	hooks, err := store.ListOpenHooks(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	flags, err := store.GetWorldFlags(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	markdown := session.RenderRecap(campaign.Name, filtered, hooks, flags)
	result := map[string]interface{}{
		"campaign_id":   campaignID,
		"from_session":  fromSession,
		"to_session":    toSession,
		"session_count": len(filtered),
		"markdown":      markdown,
	}
	jsonResult, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(jsonResult)), nil
}

// generateCampaignID creates a URL-safe campaign ID from a name.
func generateCampaignID(name string) string {
	id := ""
	for _, c := range name {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
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
