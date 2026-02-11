// internal/core/engine.go
package core

import (
	"nextturn/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// GameEngine defines the interface all games must implement.
type GameEngine interface {
	GameType() string
	DisplayName() string
	Description() string
	InitialBoard() string
	ThumbnailURL() string

	// Game logic
	CheckResult(board string) *string // nil = ongoing, "X"/"O" = winner, "draw"
	IsValidMove(game *models.Game, position int) bool
	ApplyMove(board string, position int, symbol string) string
	NextTurn(currentTurn string) string

	// UI builders
	BuildMessage(game *models.Game) string
	BuildKeyboard(game *models.Game) tgbotapi.InlineKeyboardMarkup
	BuildLobbyWaiting(gameID string) tgbotapi.InlineKeyboardMarkup
	BuildLobbyReady(gameID string) tgbotapi.InlineKeyboardMarkup
	BuildFinishedKeyboard(game *models.Game) tgbotapi.InlineKeyboardMarkup
}
