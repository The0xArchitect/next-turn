// internal/games/checkers/handlers.go
package checkers

import (
	"context"
	"log"
	"strconv"
	"strings"

	"nextturn/internal/core"
	"nextturn/internal/db"
	"nextturn/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Handlers struct {
	DB     *db.DB
	Engine *Engine
}

func NewHandlers(database *db.DB) *Handlers {
	return &Handlers{
		DB:     database,
		Engine: core.Get("checkers").(*Engine),
	}
}

func (h *Handlers) HandleCallback(bot *tgbotapi.BotAPI, cb *tgbotapi.CallbackQuery) {
	data := cb.Data
	parts := strings.Split(data, ":")
	if len(parts) < 3 {
		answerCallback(bot, cb.ID, "Invalid action")
		return
	}

	action := parts[1]
	switch action {
	case "join":
		h.onJoin(bot, cb, parts)
	case "start":
		h.onStart(bot, cb, parts)
	case "quit":
		h.onQuit(bot, cb, parts)
	case "tap":
		h.onTap(bot, cb, parts)
	case "deselect":
		h.onDeselect(bot, cb, parts)
	case "rematch":
		h.onRematch(bot, cb, parts)
	default:
		answerCallback(bot, cb.ID, "Unknown action")
	}
}

func (h *Handlers) onJoin(bot *tgbotapi.BotAPI, cb *tgbotapi.CallbackQuery, parts []string) {
	game, imid := h.getGame(bot, cb, parts)
	if game == nil {
		return
	}

	user := cb.From
	if game.Status != "waiting" {
		answerCallback(bot, cb.ID, "Cannot join this game")
		return
	}
	if user.ID == game.CreatorID {
		answerCallback(bot, cb.ID, "You created this game!")
		return
	}

	ctx := context.Background()
	updated, err := h.DB.JoinGame(ctx, game.ID, user.ID, user.FirstName)
	if err != nil || updated == nil {
		answerCallback(bot, cb.ID, "Could not join")
		return
	}

	answerCallback(bot, cb.ID, "Joined! 🎮")
	h.updateMessage(bot, updated, imid)
}

func (h *Handlers) onStart(bot *tgbotapi.BotAPI, cb *tgbotapi.CallbackQuery, parts []string) {
	game, imid := h.getGame(bot, cb, parts)
	if game == nil {
		return
	}

	user := cb.From
	if game.Status != "ready" {
		answerCallback(bot, cb.ID, "Game not ready")
		return
	}
	if user.ID != game.CreatorID {
		answerCallback(bot, cb.ID, "Only the creator can start")
		return
	}

	ctx := context.Background()
	updated, err := h.DB.StartGame(ctx, game.ID, user.ID)
	if err != nil || updated == nil {
		answerCallback(bot, cb.ID, "Could not start")
		return
	}

	answerCallback(bot, cb.ID, "Game on! 🔥")
	h.updateMessage(bot, updated, imid)
}

func (h *Handlers) onQuit(bot *tgbotapi.BotAPI, cb *tgbotapi.CallbackQuery, parts []string) {
	game, imid := h.getGame(bot, cb, parts)
	if game == nil {
		return
	}

	user := cb.From
	if game.Status == "finished" {
		answerCallback(bot, cb.ID, "Game already ended")
		return
	}
	if !game.IsPlayer(user.ID) {
		answerCallback(bot, cb.ID, "You are not in this game")
		return
	}

	ctx := context.Background()
	updated, err := h.DB.QuitGame(ctx, game.ID, user.ID)
	if err != nil || updated == nil {
		answerCallback(bot, cb.ID, "Could not quit")
		return
	}

	answerCallback(bot, cb.ID, "Game ended")
	h.updateMessage(bot, updated, imid)
}

func (h *Handlers) onTap(bot *tgbotapi.BotAPI, cb *tgbotapi.CallbackQuery, parts []string) {
	game, imid := h.getGame(bot, cb, parts)
	if game == nil {
		return
	}

	if len(parts) < 4 {
		answerCallback(bot, cb.ID, "Invalid position")
		return
	}
	pos, err := strconv.Atoi(parts[3])
	if err != nil || pos < 0 || pos >= boardLength {
		answerCallback(bot, cb.ID, "Invalid position")
		return
	}

	user := cb.From
	if game.Status != "playing" {
		answerCallback(bot, cb.ID, "Game not in progress")
		return
	}
	if !game.IsPlayer(user.ID) {
		answerCallback(bot, cb.ID, "You are not in this game")
		return
	}
	if game.CurrentPlayerID() != user.ID {
		answerCallbackAlert(bot, cb.ID, "Not your turn!")
		return
	}

	isWhite := game.CurrentTurn == "X"
	cell := game.Board[pos]

	// Check if tapped cell is own piece
	isOwn := false
	if isWhite {
		isOwn = cell == 'w' || cell == 'W'
	} else {
		isOwn = cell == 'b' || cell == 'B'
	}

	if isOwn {
		h.handleSelect(bot, cb, game, pos, isWhite, imid)
	} else if game.SelectedPos.Valid {
		h.handleMove(bot, cb, game, pos, isWhite, imid)
	} else {
		answerCallback(bot, cb.ID, "Select one of your pieces first")
	}
}

func (h *Handlers) handleSelect(bot *tgbotapi.BotAPI, cb *tgbotapi.CallbackQuery, game *models.Game, pos int, isWhite bool, imid string) {
	movable := h.Engine.GetMovablePieces(game.Board, isWhite)

	found := false
	for _, p := range movable {
		if p == pos {
			found = true
			break
		}
	}
	if !found {
		// Check if mandatory jump exists
		hasJumps := false
		for _, p := range movable {
			if len(h.Engine.GetValidJumps(game.Board, p, isWhite)) > 0 {
				hasJumps = true
				break
			}
		}
		if hasJumps {
			answerCallbackAlert(bot, cb.ID, "You must jump! Select a piece that can capture.")
		} else {
			answerCallback(bot, cb.ID, "This piece has no valid moves")
		}
		return
	}

	ctx := context.Background()
	updated, err := h.DB.UpdateSelection(ctx, game.ID, &pos)
	if err != nil || updated == nil {
		answerCallback(bot, cb.ID, "Failed to select")
		return
	}

	answerCallback(bot, cb.ID, "Piece selected")
	h.updateMessage(bot, updated, imid)
}

func (h *Handlers) handleMove(bot *tgbotapi.BotAPI, cb *tgbotapi.CallbackQuery, game *models.Game, to int, isWhite bool, imid string) {
	from := int(game.SelectedPos.Int32)
	validDests := h.Engine.GetValidDestinations(game.Board, from, isWhite)

	found := false
	for _, d := range validDests {
		if d == to {
			found = true
			break
		}
	}
	if !found {
		answerCallback(bot, cb.ID, "Invalid move")
		return
	}

	result := h.Engine.ApplyCheckerMove(game.Board, from, to, isWhite)
	ctx := context.Background()

	// Multi-jump continuation
	if result.MustContinueJump {
		contPos := result.ContinueFromPos
		updated, err := h.DB.UpdateBoardWithSelection(ctx, game.ID, result.Board, &contPos)
		if err != nil || updated == nil {
			answerCallback(bot, cb.ID, "Move failed")
			return
		}
		answerCallback(bot, cb.ID, "Jump! Continue capturing...")
		h.updateMessage(bot, updated, imid)
		return
	}

	// Normal end of turn
	nextTurn := h.Engine.NextTurn(game.CurrentTurn)
	updated, err := h.DB.UpdateBoardClearSelection(ctx, game.ID, result.Board, nextTurn)
	if err != nil || updated == nil {
		answerCallback(bot, cb.ID, "Move failed, try again")
		return
	}

	answerCallback(bot, cb.ID, "")

	// Check win by elimination
	boardResult := h.Engine.CheckResult(updated.Board)
	if boardResult != nil {
		var gameResult string
		if *boardResult == "X" {
			gameResult = "creator"
		} else {
			gameResult = "opponent"
		}
		finished, err := h.DB.FinishGame(ctx, updated.ID, gameResult)
		if err == nil && finished != nil {
			updated = finished
		}
	} else {
		// Check if next player has any moves (stalemate = loss)
		nextIsWhite := nextTurn == "X"
		if !h.Engine.HasValidMoves(updated.Board, nextTurn) {
			var gameResult string
			if nextIsWhite {
				gameResult = "opponent"
			} else {
				gameResult = "creator"
			}
			finished, err := h.DB.FinishGame(ctx, updated.ID, gameResult)
			if err == nil && finished != nil {
				updated = finished
			}
		}
		_ = nextIsWhite
	}

	h.updateMessage(bot, updated, imid)
}

func (h *Handlers) onDeselect(bot *tgbotapi.BotAPI, cb *tgbotapi.CallbackQuery, parts []string) {
	game, imid := h.getGame(bot, cb, parts)
	if game == nil {
		return
	}

	user := cb.From
	if game.CurrentPlayerID() != user.ID {
		answerCallback(bot, cb.ID, "Not your turn")
		return
	}
	if !game.SelectedPos.Valid {
		answerCallback(bot, cb.ID, "Nothing selected")
		return
	}

	isWhite := game.CurrentTurn == "X"
	selectedPos := int(game.SelectedPos.Int32)

	// Prevent deselect if forced to complete a jump with this piece
	jumps := h.Engine.GetValidJumps(game.Board, selectedPos, isWhite)
	moves := h.Engine.GetValidMoves(game.Board, selectedPos, isWhite)
	if len(jumps) > 0 && len(moves) == 0 {
		movable := h.Engine.GetMovablePieces(game.Board, isWhite)
		if len(movable) == 1 && movable[0] == selectedPos {
			answerCallbackAlert(bot, cb.ID, "You must complete the jump!")
			return
		}
	}

	ctx := context.Background()
	updated, err := h.DB.UpdateSelection(ctx, game.ID, nil)
	if err != nil || updated == nil {
		answerCallback(bot, cb.ID, "Failed to deselect")
		return
	}

	answerCallback(bot, cb.ID, "Deselected")
	h.updateMessage(bot, updated, imid)
}

func (h *Handlers) onRematch(bot *tgbotapi.BotAPI, cb *tgbotapi.CallbackQuery, parts []string) {
	game, imid := h.getGame(bot, cb, parts)
	if game == nil {
		return
	}

	user := cb.From
	if game.Status != "finished" {
		answerCallback(bot, cb.ID, "Game not finished")
		return
	}
	if !game.IsPlayer(user.ID) {
		answerCallback(bot, cb.ID, "You are not in this game")
		return
	}
	if game.IsExpired() {
		answerCallback(bot, cb.ID, "Game expired, start a new one")
		return
	}

	ctx := context.Background()
	newGame, err := h.DB.CreateRematch(ctx, game, h.Engine.InitialBoard())
	if err != nil || newGame == nil {
		answerCallback(bot, cb.ID, "Could not create rematch")
		return
	}

	answerCallback(bot, cb.ID, "New game started!")
	h.updateMessage(bot, newGame, imid)
}

func (h *Handlers) getGame(bot *tgbotapi.BotAPI, cb *tgbotapi.CallbackQuery, parts []string) (*models.Game, string) {
	imid := ""
	if cb.InlineMessageID != "" {
		imid = cb.InlineMessageID
	}

	if len(parts) < 3 || imid == "" {
		answerCallback(bot, cb.ID, "Invalid action")
		return nil, ""
	}

	gameID := parts[2]
	ctx := context.Background()

	game, err := h.DB.GetGame(ctx, gameID)
	if err != nil {
		log.Printf("getGame error: %v", err)
	}
	if game == nil {
		game, err = h.DB.GetGameByInlineMessageID(ctx, imid)
		if err != nil {
			log.Printf("getGame by imid error: %v", err)
		}
	}

	if game == nil {
		answerCallback(bot, cb.ID, "Game not found")
		return nil, ""
	}

	if game.IsExpired() && game.Status != "finished" {
		h.updateMessage(bot, game, imid)
		answerCallback(bot, cb.ID, "Game expired")
		return nil, ""
	}

	return game, imid
}

func (h *Handlers) updateMessage(bot *tgbotapi.BotAPI, game *models.Game, inlineMessageID string) {
	text := h.Engine.BuildMessage(game)
	kb := h.Engine.BuildKeyboard(game)

	edit := tgbotapi.EditMessageTextConfig{
		BaseEdit: tgbotapi.BaseEdit{
			InlineMessageID: inlineMessageID,
			ReplyMarkup:     &kb,
		},
		Text: text,
	}
	if _, err := bot.Request(edit); err != nil {
		log.Printf("updateMessage error: %v", err)
	}
}

func answerCallback(bot *tgbotapi.BotAPI, callbackID, text string) {
	cb := tgbotapi.NewCallback(callbackID, text)
	if _, err := bot.Request(cb); err != nil {
		log.Printf("answerCallback error: %v", err)
	}
}

func answerCallbackAlert(bot *tgbotapi.BotAPI, callbackID, text string) {
	cb := tgbotapi.NewCallbackWithAlert(callbackID, text)
	if _, err := bot.Request(cb); err != nil {
		log.Printf("answerCallbackAlert error: %v", err)
	}
}
