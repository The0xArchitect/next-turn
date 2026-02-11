// internal/games/tictactoe/engine.go
package tictactoe

import (
	"strings"

	"nextturn/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var winPatterns = [][3]int{
	{0, 1, 2}, {3, 4, 5}, {6, 7, 8}, // rows
	{0, 3, 6}, {1, 4, 7}, {2, 5, 8}, // cols
	{0, 4, 8}, {2, 4, 6}, // diagonals
}

type Engine struct{}

func (e *Engine) GameType() string     { return "tictactoe" }
func (e *Engine) DisplayName() string  { return "🎮 Tic Tac Toe" }
func (e *Engine) Description() string  { return "Start a new game!" }
func (e *Engine) InitialBoard() string { return "---------" }
func (e *Engine) ThumbnailURL() string { return "tic-tac-toe.jpg" }

func (e *Engine) CheckResult(board string) *string {
	for _, p := range winPatterns {
		a, b, c := string(board[p[0]]), string(board[p[1]]), string(board[p[2]])
		if a != "-" && a == b && b == c {
			return &a
		}
	}
	if !strings.Contains(board, "-") {
		d := "draw"
		return &d
	}
	return nil
}

func (e *Engine) IsValidMove(game *models.Game, position int) bool {
	return position >= 0 && position < 9 && game.Board[position] == '-'
}

func (e *Engine) ApplyMove(board string, position int, symbol string) string {
	return board[:position] + symbol + board[position+1:]
}

func (e *Engine) NextTurn(currentTurn string) string {
	if currentTurn == "X" {
		return "O"
	}
	return "X"
}

func (e *Engine) BuildMessage(game *models.Game) string {
	return Message(game)
}

func (e *Engine) BuildKeyboard(game *models.Game) tgbotapi.InlineKeyboardMarkup {
	return Keyboard(game)
}

func (e *Engine) BuildLobbyWaiting(gameID string) tgbotapi.InlineKeyboardMarkup {
	return LobbyWaiting(gameID)
}

func (e *Engine) BuildLobbyReady(gameID string) tgbotapi.InlineKeyboardMarkup {
	return LobbyReady(gameID)
}

func (e *Engine) BuildFinishedKeyboard(game *models.Game) tgbotapi.InlineKeyboardMarkup {
	return Finished(game)
}
