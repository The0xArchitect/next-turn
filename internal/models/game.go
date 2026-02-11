// internal/models/game.go
package models

import (
	"database/sql"
	"time"
)

type Game struct {
	ID              string
	InlineMessageID sql.NullString
	GameType        string
	CreatorID       int64
	CreatorName     string
	OpponentID      sql.NullInt64
	OpponentName    sql.NullString
	Status          string
	Board           string
	XPlayer         string
	CurrentTurn     string
	SelectedPos     sql.NullInt32
	Result          sql.NullString
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// 48h expiry
func (g *Game) IsExpired() bool {
	return time.Since(g.CreatedAt) > 48*time.Hour
}

func (g *Game) IsPlayer(userID int64) bool {
	return userID == g.CreatorID || (g.OpponentID.Valid && userID == g.OpponentID.Int64)
}

func (g *Game) CurrentPlayerID() int64 {
	isCreatorX := g.XPlayer == "creator"
	if (isCreatorX && g.CurrentTurn == "X") || (!isCreatorX && g.CurrentTurn == "O") {
		return g.CreatorID
	}
	if g.OpponentID.Valid {
		return g.OpponentID.Int64
	}
	return 0
}

func (g *Game) XPlayerName() string {
	if g.XPlayer == "creator" {
		return g.CreatorName
	}
	if g.OpponentName.Valid {
		return g.OpponentName.String
	}
	return "?"
}

func (g *Game) OPlayerName() string {
	if g.XPlayer == "creator" {
		if g.OpponentName.Valid {
			return g.OpponentName.String
		}
		return "?"
	}
	return g.CreatorName
}

func (g *Game) CurrentPlayerName() string {
	isCreatorX := g.XPlayer == "creator"
	if (isCreatorX && g.CurrentTurn == "X") || (!isCreatorX && g.CurrentTurn == "O") {
		return g.CreatorName
	}
	if g.OpponentName.Valid {
		return g.OpponentName.String
	}
	return "?"
}

func (g *Game) GetOpponentName() string {
	if g.OpponentName.Valid {
		return g.OpponentName.String
	}
	return "?"
}

func (g *Game) GetResult() string {
	if g.Result.Valid {
		return g.Result.String
	}
	return ""
}
