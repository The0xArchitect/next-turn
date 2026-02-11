// internal/games/checkers/ui.go
package checkers

import (
	"fmt"

	"nextturn/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const prefix = "ck"

func cellDisplay(cell byte, row, col int, isSelected, isValidDest bool) string {
	if isValidDest {
		return "\u3000" // ideographic space — highlighted destination
	}
	if isSelected {
		if cell == 'w' || cell == 'W' {
			return "🔵"
		}
		return "🟣"
	}
	switch cell {
	case 'w':
		return "⚪" // white man
	case 'W':
		return "⬜️" // white king
	case 'b':
		return "⚫" // black man
	case 'B':
		return "⬛️" // black king
	default:
		return "\u3000" // empty or light square
	}
}

func LobbyWaiting(gameID string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🎮 Join Game", fmt.Sprintf("%s:join:%s", prefix, gameID)),
			tgbotapi.NewInlineKeyboardButtonData("❌ Cancel", fmt.Sprintf("%s:quit:%s", prefix, gameID)),
		),
	)
}

func LobbyReady(gameID string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("▶️ Start Game", fmt.Sprintf("%s:start:%s", prefix, gameID)),
			tgbotapi.NewInlineKeyboardButtonData("❌ Leave", fmt.Sprintf("%s:quit:%s", prefix, gameID)),
		),
	)
}

func Board(gameID, boardState string, engine *Engine, frozen bool, selectedPos int, validMoves map[int]bool) tgbotapi.InlineKeyboardMarkup {
	var kbRows [][]tgbotapi.InlineKeyboardButton
	for row := 0; row < size; row++ {
		var buttons []tgbotapi.InlineKeyboardButton
		for col := 0; col < size; col++ {
			pos := row*size + col
			cell := boardState[pos]
			isSelected := selectedPos == pos
			isValid := validMoves[pos]
			display := cellDisplay(cell, row, col, isSelected, isValid)

			cbData := "noop"
			if !frozen {
				cbData = fmt.Sprintf("%s:tap:%s:%d", prefix, gameID, pos)
			}
			buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(display, cbData))
		}
		kbRows = append(kbRows, buttons)
	}
	return tgbotapi.InlineKeyboardMarkup{InlineKeyboard: kbRows}
}

func Finished(game *models.Game) tgbotapi.InlineKeyboardMarkup {
	e := &Engine{}
	kb := Board(game.ID, game.Board, e, true, -1, nil)
	kb.InlineKeyboard = append(kb.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🔄 Rematch?", fmt.Sprintf("%s:rematch:%s", prefix, game.ID)),
	))
	return kb
}

func Keyboard(game *models.Game) tgbotapi.InlineKeyboardMarkup {
	if game.IsExpired() && game.Status != "finished" {
		return tgbotapi.NewInlineKeyboardMarkup()
	}
	switch game.Status {
	case "waiting":
		return LobbyWaiting(game.ID)
	case "ready":
		return LobbyReady(game.ID)
	case "playing":
		return playingKeyboard(game)
	case "finished":
		return Finished(game)
	default:
		return tgbotapi.NewInlineKeyboardMarkup()
	}
}

func playingKeyboard(game *models.Game) tgbotapi.InlineKeyboardMarkup {
	e := &Engine{}
	isWhiteTurn := game.CurrentTurn == "X"

	selectedPos := -1
	if game.SelectedPos.Valid {
		selectedPos = int(game.SelectedPos.Int32)
	}

	var validMoves map[int]bool
	if selectedPos >= 0 {
		dests := e.GetValidDestinations(game.Board, selectedPos, isWhiteTurn)
		validMoves = make(map[int]bool, len(dests))
		for _, d := range dests {
			validMoves[d] = true
		}
	}

	kb := Board(game.ID, game.Board, e, false, selectedPos, validMoves)

	// Add deselect button if piece is selected
	if selectedPos >= 0 {
		kb.InlineKeyboard = append(kb.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("↩️ Deselect", fmt.Sprintf("%s:deselect:%s", prefix, game.ID)),
		))
	}

	return kb
}

func Message(game *models.Game) string {
	switch game.Status {
	case "waiting":
		return messageWaiting(game)
	case "ready":
		return messageReady(game)
	case "playing":
		return messagePlaying(game)
	case "finished":
		return messageFinished(game)
	default:
		return "Unknown game state"
	}
}

func messageWaiting(game *models.Game) string {
	if game.IsExpired() {
		return "⚫⚪ Checkers\n\n⏰ This game has expired"
	}
	return fmt.Sprintf("⚫⚪ Checkers\n\n%s wants to play!\nTap Join to accept the challenge.", game.CreatorName)
}

func messageReady(game *models.Game) string {
	if game.IsExpired() {
		return "⚫⚪ Checkers\n\n⏰ This game has expired"
	}
	return fmt.Sprintf("⚫⚪ Checkers\n\n⚪ %s vs ⚫ %s\n\nPlayers ready! Host can start the game.",
		game.XPlayerName(), game.OPlayerName())
}

func messagePlaying(game *models.Game) string {
	if game.IsExpired() {
		return "⚫⚪ Checkers\n\n⏰ This game has expired"
	}
	turnSymbol := "⚪"
	if game.CurrentTurn == "O" {
		turnSymbol = "⚫"
	}
	instruction := "Tap a piece to select it"
	if game.SelectedPos.Valid {
		instruction = "Tap a square to move there"
	}
	return fmt.Sprintf("⚫⚪ Checkers\n\n⚪ %s vs ⚫ %s\n\n%s %s's turn\n%s",
		game.XPlayerName(), game.OPlayerName(), turnSymbol, game.CurrentPlayerName(), instruction)
}

func messageFinished(game *models.Game) string {
	var resultText string
	switch game.GetResult() {
	case "creator":
		resultText = fmt.Sprintf("🏆 %s wins!", game.CreatorName)
	case "opponent":
		resultText = fmt.Sprintf("🏆 %s wins!", game.GetOpponentName())
	case "draw":
		resultText = "🤝 It's a draw!"
	case "forfeit_creator":
		resultText = fmt.Sprintf("🚪 %s left. %s wins!", game.CreatorName, game.GetOpponentName())
	case "forfeit_opponent":
		resultText = fmt.Sprintf("🚪 %s left. %s wins!", game.GetOpponentName(), game.CreatorName)
	case "cancelled":
		resultText = "❌ Game was cancelled"
	default:
		resultText = "Game over"
	}
	return fmt.Sprintf("⚫⚪ Checkers\n\n⚪ %s vs ⚫ %s\n\n%s",
		game.XPlayerName(), game.OPlayerName(), resultText)
}
