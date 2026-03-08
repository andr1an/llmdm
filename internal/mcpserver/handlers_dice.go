package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/andr1an/llmdm/internal/dice"
	"github.com/andr1an/llmdm/internal/types"
)

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
		result, err = dice.RollWithAdvantage(parsed.Modifier)
	} else if disadvantage {
		parsed, parseErr := dice.Parse(notation)
		if parseErr != nil {
			return mcp.NewToolResultError(parseErr.Error()), nil
		}
		result, err = dice.RollWithDisadvantage(parsed.Modifier)
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
	if err := logger.Log(campaignID, session, attacker, contestType+" (attacker)", attackerResult, false, false); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("contested roll succeeded but attacker logging failed: %v", err)), nil
	}
	if err := logger.Log(campaignID, session, defender, contestType+" (defender)", defenderResult, false, false); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("contested roll succeeded but defender logging failed: %v", err)), nil
	}
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
	if err := logger.Log(campaignID, session, character, reason, rollResult, false, false); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("saving throw succeeded but logging failed: %v", err)), nil
	}
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
