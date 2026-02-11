// internal/games/connect4/engine.go
package connect4

import (
	"nextturn/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	cols      = 7
	rows      = 6
	boardSize = cols * rows // 42
	winLength = 4
)

var directions = [][2]int{
	{0, 1},  // horizontal →
	{1, 0},  // vertical ↓
	{1, 1},  // diagonal ↘
	{1, -1}, // diagonal ↙
}

type Engine struct{}

func (e *Engine) GameType() string     { return "connect4" }
func (e *Engine) DisplayName() string  { return "🔴🔵 Connect 4" }
func (e *Engine) Description() string  { return "Drop discs, connect 4 to win!" }
func (e *Engine) InitialBoard() string { return "------------------------------------------" } // 42 cells
func (e *Engine) ThumbnailURL() string { return "connect-4.jpg" }

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
	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			cell := board[row*cols+col]
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
		if r < 0 || r >= rows || c < 0 || c >= cols {
			return false
		}
		if board[r*cols+c] != symbol {
			return false
		}
	}
	return true
}

// FindWinningPositions returns the positions of a winning line, or nil.
func FindWinningPositions(board string) []int {
	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			cell := board[row*cols+col]
			if cell == '-' {
				continue
			}
			for _, dir := range directions {
				if positions := getLinePositions(board, row, col, dir[0], dir[1], cell); positions != nil {
					return positions
				}
			}
		}
	}
	return nil
}

func getLinePositions(board string, row, col, dRow, dCol int, symbol byte) []int {
	positions := make([]int, 0, winLength)
	for i := 0; i < winLength; i++ {
		r := row + i*dRow
		c := col + i*dCol
		if r < 0 || r >= rows || c < 0 || c >= cols {
			return nil
		}
		if board[r*cols+c] != symbol {
			return nil
		}
		positions = append(positions, r*cols+c)
	}
	return positions
}

// IsValidMove checks if a disc can be dropped into the given column.
func (e *Engine) IsValidMove(game *models.Game, column int) bool {
	if column < 0 || column >= cols {
		return false
	}
	// Top cell of column must be empty
	return game.Board[column] == '-'
}

// findDropRow returns the lowest empty row in a column, or -1.
func findDropRow(board string, column int) int {
	for row := rows - 1; row >= 0; row-- {
		if board[row*cols+column] == '-' {
			return row
		}
	}
	return -1
}

// ApplyMove drops the disc into the column. The position param is the column index.
func (e *Engine) ApplyMove(board string, column int, symbol string) string {
	row := findDropRow(board, column)
	if row < 0 {
		return board
	}
	pos := row*cols + column
	return board[:pos] + symbol + board[pos+1:]
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
