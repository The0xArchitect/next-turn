// internal/games/checkers/engine.go
package checkers

import (
	"nextturn/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	size        = 8
	boardLength = 64
)

// Piece symbols: w = white man, W = white king, b = black man, B = black king
// '.' = light square (unplayable), '-' = empty dark square
// White (X) starts at rows 5-7, moves up. Black (O) starts at rows 0-2, moves down.

type JumpMove struct {
	From     int
	To       int
	Captured int
}

type MoveResult struct {
	Board            string
	Captured         int // -1 if none
	Promoted         bool
	MustContinueJump bool
	ContinueFromPos  int // -1 if none
}

type Engine struct{}

func (e *Engine) GameType() string     { return "checkers" }
func (e *Engine) DisplayName() string  { return "⚫⚪ Checkers" }
func (e *Engine) Description() string  { return "Classic checkers - capture all pieces!" }
func (e *Engine) ThumbnailURL() string { return "checkers.jpg" }

func (e *Engine) InitialBoard() string {
	// 8x8 board, pieces on dark squares only
	// Row 0-2: black, Row 5-7: white
	board := make([]byte, boardLength)
	for row := 0; row < size; row++ {
		for col := 0; col < size; col++ {
			pos := row*size + col
			isDark := (row+col)%2 == 1
			if !isDark {
				board[pos] = '.'
			} else if row < 3 {
				board[pos] = 'b'
			} else if row > 4 {
				board[pos] = 'w'
			} else {
				board[pos] = '-'
			}
		}
	}
	return string(board)
}

func (e *Engine) CheckResult(board string) *string {
	whiteCount := 0
	blackCount := 0
	for i := 0; i < len(board); i++ {
		switch board[i] {
		case 'w', 'W':
			whiteCount++
		case 'b', 'B':
			blackCount++
		}
	}
	if whiteCount == 0 {
		s := "O"
		return &s
	}
	if blackCount == 0 {
		s := "X"
		return &s
	}
	return nil
}

// HasValidMoves checks if a player has any legal moves.
func (e *Engine) HasValidMoves(board string, player string) bool {
	isWhite := player == "X"
	var ownPieces []byte
	if isWhite {
		ownPieces = []byte{'w', 'W'}
	} else {
		ownPieces = []byte{'b', 'B'}
	}

	for pos := 0; pos < boardLength; pos++ {
		if !isOwnPiece(board[pos], ownPieces) {
			continue
		}
		if len(getValidMoves(board, pos, isWhite)) > 0 {
			return true
		}
		if len(getValidJumps(board, pos, isWhite)) > 0 {
			return true
		}
	}
	return false
}

// GetMovablePieces returns positions of pieces that can move.
// If any piece can jump, only jumping pieces are returned (mandatory capture).
func (e *Engine) GetMovablePieces(board string, isWhite bool) []int {
	var ownPieces []byte
	if isWhite {
		ownPieces = []byte{'w', 'W'}
	} else {
		ownPieces = []byte{'b', 'B'}
	}

	var jumpingPieces []int
	var movablePieces []int

	for pos := 0; pos < boardLength; pos++ {
		if !isOwnPiece(board[pos], ownPieces) {
			continue
		}
		if len(getValidJumps(board, pos, isWhite)) > 0 {
			jumpingPieces = append(jumpingPieces, pos)
		}
		if len(getValidMoves(board, pos, isWhite)) > 0 {
			movablePieces = append(movablePieces, pos)
		}
	}

	if len(jumpingPieces) > 0 {
		return jumpingPieces
	}
	return movablePieces
}

// GetValidDestinations returns all valid destination squares for a selected piece.
func (e *Engine) GetValidDestinations(board string, pos int, isWhite bool) []int {
	jumps := getValidJumps(board, pos, isWhite)

	// Check if ANY piece has jumps (mandatory jump rule)
	movable := e.GetMovablePieces(board, isWhite)
	hasAnyJumps := false
	for _, p := range movable {
		if len(getValidJumps(board, p, isWhite)) > 0 {
			hasAnyJumps = true
			break
		}
	}

	if hasAnyJumps {
		dests := make([]int, len(jumps))
		for i, j := range jumps {
			dests[i] = j.To
		}
		return dests
	}

	return getValidMoves(board, pos, isWhite)
}

// GetValidJumps returns available jump moves for a piece (exported for handlers).
func (e *Engine) GetValidJumps(board string, pos int, isWhite bool) []JumpMove {
	return getValidJumps(board, pos, isWhite)
}

// GetValidMoves returns non-capture moves (exported for handlers).
func (e *Engine) GetValidMoves(board string, pos int, isWhite bool) []int {
	return getValidMoves(board, pos, isWhite)
}

// ApplyCheckerMove applies a move (regular or jump) and returns the result.
func (e *Engine) ApplyCheckerMove(board string, from, to int, isWhite bool) MoveResult {
	piece := board[from]
	jumps := getValidJumps(board, from, isWhite)

	var jump *JumpMove
	for i := range jumps {
		if jumps[i].To == to {
			jump = &jumps[i]
			break
		}
	}

	b := []byte(board)

	// Clear from
	b[from] = '-'

	// Remove captured piece if jump
	captured := -1
	if jump != nil {
		captured = jump.Captured
		b[captured] = '-'
	}

	// Check promotion
	newPiece := piece
	toRow := to / size
	if piece == 'w' && toRow == 0 {
		newPiece = 'W'
	} else if piece == 'b' && toRow == 7 {
		newPiece = 'B'
	}

	// Place piece
	b[to] = newPiece

	newBoard := string(b)

	// Multi-jump: if it was a jump and piece wasn't promoted, check for more jumps
	mustContinue := false
	continueFrom := -1
	if jump != nil && newPiece == piece {
		moreJumps := getValidJumps(newBoard, to, isWhite)
		if len(moreJumps) > 0 {
			mustContinue = true
			continueFrom = to
		}
	}

	return MoveResult{
		Board:            newBoard,
		Captured:         captured,
		Promoted:         newPiece != piece,
		MustContinueJump: mustContinue,
		ContinueFromPos:  continueFrom,
	}
}

// --- private helpers ---

func isOwnPiece(cell byte, ownPieces []byte) bool {
	for _, p := range ownPieces {
		if cell == p {
			return true
		}
	}
	return false
}

func isEnemyPiece(cell byte, isWhite bool) bool {
	if isWhite {
		return cell == 'b' || cell == 'B'
	}
	return cell == 'w' || cell == 'W'
}

func getValidMoves(board string, pos int, isWhite bool) []int {
	piece := board[pos]
	if piece == '.' || piece == '-' {
		return nil
	}

	isKing := piece == 'W' || piece == 'B'
	row := pos / size
	col := pos % size

	var forwardDirs []int
	if isWhite {
		forwardDirs = []int{-1}
	} else {
		forwardDirs = []int{1}
	}

	var dirs []int
	if isKing {
		dirs = []int{-1, 1}
	} else {
		dirs = forwardDirs
	}

	var moves []int
	for _, dRow := range dirs {
		for _, dCol := range []int{-1, 1} {
			nr := row + dRow
			nc := col + dCol
			if nr >= 0 && nr < size && nc >= 0 && nc < size {
				np := nr*size + nc
				if board[np] == '-' {
					moves = append(moves, np)
				}
			}
		}
	}
	return moves
}

func getValidJumps(board string, pos int, isWhite bool) []JumpMove {
	piece := board[pos]
	if piece == '.' || piece == '-' {
		return nil
	}

	isKing := piece == 'W' || piece == 'B'
	row := pos / size
	col := pos % size

	var forwardDirs []int
	if isWhite {
		forwardDirs = []int{-1}
	} else {
		forwardDirs = []int{1}
	}

	var dirs []int
	if isKing {
		dirs = []int{-1, 1}
	} else {
		dirs = forwardDirs
	}

	var jumps []JumpMove
	for _, dRow := range dirs {
		for _, dCol := range []int{-1, 1} {
			midRow := row + dRow
			midCol := col + dCol
			endRow := row + 2*dRow
			endCol := col + 2*dCol

			if endRow >= 0 && endRow < size && endCol >= 0 && endCol < size {
				midPos := midRow*size + midCol
				endPos := endRow*size + endCol
				if isEnemyPiece(board[midPos], isWhite) && board[endPos] == '-' {
					jumps = append(jumps, JumpMove{From: pos, To: endPos, Captured: midPos})
				}
			}
		}
	}
	return jumps
}

// --- GameEngine interface (unused stubs for checkers) ---

func (e *Engine) IsValidMove(game *models.Game, position int) bool {
	return true // Handled in handlers
}

func (e *Engine) ApplyMove(board string, position int, symbol string) string {
	return board // Not used for checkers
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
