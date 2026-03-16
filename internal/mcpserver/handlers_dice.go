package mcpserver

import (
	"context"
	"fmt"
	"strconv"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/andr1an/llmdm/internal/dice"
	"github.com/andr1an/llmdm/internal/types"
)

func (s *Server) handleRoll(ctx context.Context, req *mcp.CallToolRequest, input RollInput) (*mcp.CallToolResult, RollOutput, error) {
	s.logToolEntry("roll", input.CampaignID, "notation", input.Notation, "character", input.Character)

	var result types.RollResult
	var err error

	if input.Advantage && input.Disadvantage {
		// Advantage and disadvantage cancel out
		result, err = dice.Roll(input.Notation)
	} else if input.Advantage {
		// Parse notation to get modifier
		parsed, parseErr := dice.Parse(input.Notation)
		if parseErr != nil {
			s.logToolError("roll", input.CampaignID, parseErr)
			return nil, RollOutput{}, wrapError("parse notation", parseErr)
		}
		result, err = dice.RollWithAdvantage(parsed.Modifier)
	} else if input.Disadvantage {
		parsed, parseErr := dice.Parse(input.Notation)
		if parseErr != nil {
			s.logToolError("roll", input.CampaignID, parseErr)
			return nil, RollOutput{}, wrapError("parse notation", parseErr)
		}
		result, err = dice.RollWithDisadvantage(parsed.Modifier)
	} else {
		result, err = dice.Roll(input.Notation)
	}

	if err != nil {
		s.logToolError("roll", input.CampaignID, err, "notation", input.Notation)
		return nil, RollOutput{}, wrapError("roll dice", err)
	}

	// Log the roll
	err = s.withDB(ctx, input.CampaignID, func(dbCtx *DBContext) error {
		logger := dice.NewLogger(dbCtx.DB)
		return logger.Log(input.CampaignID, input.Session, input.Character, input.Reason, result, input.Advantage, input.Disadvantage)
	})
	if err != nil {
		s.logToolError("roll", input.CampaignID, err)
		return nil, RollOutput{}, wrapError("log roll", err)
	}
	s.logToolExit("roll", input.CampaignID, "total", result.Total, "roll_id", result.RollID)

	return nil, RollOutput{RollResult: result}, nil
}

func (s *Server) handleRollContested(ctx context.Context, req *mcp.CallToolRequest, input RollContestedInput) (*mcp.CallToolResult, RollContestedOutput, error) {
	s.logToolEntry("roll_contested", input.CampaignID, "attacker", input.Attacker, "defender", input.Defender, "contest_type", input.ContestType)

	// Roll both
	attackerResult, err := dice.Roll(input.AttackerNotation)
	if err != nil {
		s.logToolError("roll_contested", input.CampaignID, err, "context", "attacker roll")
		return nil, RollContestedOutput{}, wrapError("attacker roll", err)
	}

	defenderResult, err := dice.Roll(input.DefenderNotation)
	if err != nil {
		s.logToolError("roll_contested", input.CampaignID, err, "context", "defender roll")
		return nil, RollContestedOutput{}, wrapError("defender roll", err)
	}

	// Determine winner
	var winner string
	margin := attackerResult.Total - defenderResult.Total
	if margin > 0 {
		winner = input.Attacker
	} else if margin < 0 {
		winner = input.Defender
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
	err = s.withDB(ctx, input.CampaignID, func(dbCtx *DBContext) error {
		logger := dice.NewLogger(dbCtx.DB)
		if err := logger.Log(input.CampaignID, input.Session, input.Attacker, input.ContestType+" (attacker)", attackerResult, false, false); err != nil {
			return fmt.Errorf("log attacker roll: %w", err)
		}
		if err := logger.Log(input.CampaignID, input.Session, input.Defender, input.ContestType+" (defender)", defenderResult, false, false); err != nil {
			return fmt.Errorf("log defender roll: %w", err)
		}
		return nil
	})
	if err != nil {
		s.logToolError("roll_contested", input.CampaignID, err)
		return nil, RollContestedOutput{}, wrapError("log contested rolls", err)
	}

	s.logToolExit("roll_contested", input.CampaignID, "winner", winner, "margin", margin, "attacker_total", attackerResult.Total, "defender_total", defenderResult.Total)

	return nil, RollContestedOutput{Result: result}, nil
}

func (s *Server) handleRollSavingThrow(ctx context.Context, req *mcp.CallToolRequest, input RollSavingThrowInput) (*mcp.CallToolResult, RollSavingThrowOutput, error) {
	reason := input.Reason
	if reason == "" {
		reason = input.Stat + " saving throw"
	}

	s.logToolEntry("roll_saving_throw", input.CampaignID, "character", input.Character, "stat", input.Stat, "dc", input.DC)

	// Roll 1d20
	notation := "1d20"
	if input.Modifier >= 0 {
		notation += "+" + strconv.Itoa(int(input.Modifier))
	} else {
		notation += strconv.Itoa(int(input.Modifier))
	}

	rollResult, err := dice.Roll(notation)
	if err != nil {
		s.logToolError("roll_saving_throw", input.CampaignID, err)
		return nil, RollSavingThrowOutput{}, wrapError("roll saving throw", err)
	}

	result := types.SavingThrowResult{
		Total:    rollResult.Total,
		Rolled:   rollResult.Rolls[0],
		Modifier: int(input.Modifier),
		DC:       int(input.DC),
		Success:  rollResult.Total >= int(input.DC),
		RollID:   rollResult.RollID,
	}

	// Log the roll
	err = s.withDB(ctx, input.CampaignID, func(dbCtx *DBContext) error {
		logger := dice.NewLogger(dbCtx.DB)
		return logger.Log(input.CampaignID, input.Session, input.Character, reason, rollResult, false, false)
	})
	if err != nil {
		s.logToolError("roll_saving_throw", input.CampaignID, err)
		return nil, RollSavingThrowOutput{}, wrapError("log saving throw", err)
	}

	s.logToolExit("roll_saving_throw", input.CampaignID, "total", result.Total, "dc", result.DC, "success", result.Success)

	return nil, RollSavingThrowOutput{Result: result}, nil
}

func (s *Server) handleGetRollHistory(ctx context.Context, req *mcp.CallToolRequest, input GetRollHistoryInput) (*mcp.CallToolResult, GetRollHistoryOutput, error) {
	limit := input.Limit
	if limit == 0 {
		limit = 50
	}

	s.logToolEntry("get_roll_history", input.CampaignID, "character", input.Character, "limit", limit)

	var records []types.RollRecord
	err := s.withDB(ctx, input.CampaignID, func(dbCtx *DBContext) error {
		logger := dice.NewLogger(dbCtx.DB)
		var err error
		records, err = logger.GetHistory(input.CampaignID, input.Character, input.Session, limit)
		return err
	})
	if err != nil {
		s.logToolError("get_roll_history", input.CampaignID, err)
		return nil, GetRollHistoryOutput{}, wrapError("get roll history", err)
	}

	s.logToolExit("get_roll_history", input.CampaignID, "count", len(records))
	return nil, GetRollHistoryOutput{Records: records}, nil
}

// === Campaign Memory Handlers ===
