// internal/games/elephantxo/engine.go
package elephantxo

import (
	"nextturn/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	size      = 7
	winLength = 5
)

var directions = [][2]int{
	{0, 1},  // horizontal →
	{1, 0},  // vertical ↓
	{1, 1},  // diagonal ↘
	{1, -1}, // diagonal ↙
}

type Engine struct{}

func (e *Engine) GameType() string     { return "elephantxo" }
func (e *Engine) DisplayName() string  { return "🐘 Elephant XO" }
func (e *Engine) Description() string  { return "7×7 board, 5 in a row to win!" }
func (e *Engine) InitialBoard() string { return "-------------------------------------------------" } // 49 cells
func (e *Engine) ThumbnailURL() string { return "elephant-xo.jpg" }

func (e *Engine) CheckResult(board string) *string {
	if winner := findWinner(board); winner != "" {
		return &winner
	}
	for i := 0; i < len(board); i++ {
		if board[i] == '-' {
			return nil
		}
	}
	d := "draw"
	return &d
}

func findWinner(board string) string {
	for row := 0; row < size; row++ {
		for col := 0; col < size; col++ {
			cell := board[row*size+col]
			if cell == '-' {
				continue
			}
			for _, dir := range directions {
				if checkLine(board, row, col, dir[0], dir[1], cell) {
					return string(cell)
				}
			}
		}
	}
	return ""
}

func checkLine(board string, row, col, dRow, dCol int, symbol byte) bool {
	for i := 0; i < winLength; i++ {
		r := row + i*dRow
		c := col + i*dCol
		if r < 0 || r >= size || c < 0 || c >= size {
			return false
		}
		if board[r*size+c] != symbol {
			return false
		}
	}
	return true
}

func (e *Engine) IsValidMove(game *models.Game, position int) bool {
	return position >= 0 && position < size*size && game.Board[position] == '-'
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
