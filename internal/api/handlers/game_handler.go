package handlers

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"net/http"
	"project2/internal/domain/entities"
	service_interfaces "project2/internal/domain/interfaces/service"
	"project2/pkg/errs"
	"project2/pkg/logger"
	"project2/pkg/utils"
	"time"
)

type GameHandler struct {
	gameService service_interfaces.GameService
}

func NewGameHandler(gameService service_interfaces.GameService) *GameHandler {
	return &GameHandler{
		gameService: gameService,
	}
}

func (g *GameHandler) GetAllGamesHandler(w http.ResponseWriter, r *http.Request) {
	logger.Logger.Infow("Handling GetAllGames request", "method", r.Method, "time", time.Now())

	if id := r.URL.Query().Get("id"); id != "" {
		gameId, err := uuid.Parse(id)
		if err != nil {
			errs.ValidationError("invalid game id").ToJson2(w)
			logger.Logger.Errorw("Error parsing gameID", "method", r.Method, "game_id", id, "error", err, "time", time.Now())
			return
		}
		game, err := g.gameService.GetGameByID(r.Context(), gameId)
		if err != nil {
			errs.DBError("Failed to fetch game").ToJson2(w)
			logger.Logger.Errorw("Failed to fetch game by ID", "method", r.Method, "game_id", gameId, "error", err, "time", time.Now())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		jsonResponse := map[string]any{
			"code":    http.StatusOK,
			"message": "Success",
			"game":    game,
		}
		if err = utils.JsonEncoder(w, jsonResponse); err != nil {
			return
		}

		logger.Logger.Infow("Successfully fetched game by ID", "method", r.Method, "game_id", gameId, "time", time.Now())

	} else {
		games, err := g.gameService.GetAllGames(r.Context())
		if err != nil {
			errs.DBError("Could not fetch the games").ToJson2(w)
			logger.Logger.Errorw("Error fetching games", "method", r.Method, "error", err, "time", time.Now())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		jsonResponse := map[string]any{
			"code":    http.StatusOK,
			"message": "Success",
			"games":   games,
		}
		if err = utils.JsonEncoder(w, jsonResponse); err != nil {
			return
		}
		logger.Logger.Infow("Successfully fetched all games", "method", r.Method, "games_count", len(games), "time", time.Now())
	}

}

func (g *GameHandler) CreateGameHandler(w http.ResponseWriter, r *http.Request) {
	logger.Logger.Infow("Handling CreateGame request", "method", r.Method, "time", time.Now())

	var requestBody struct {
		Name       string `json:"name" validate:"required"`
		MaxPlayers int    `json:"max_players" validate:"required"`
		MinPlayers int    `json:"min_players" validate:"required"`
		Instances  int    `json:"instances" validate:"required"`
	}

	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		//http.Error(w, err.Error(), http.StatusBadRequest)
		errs.InvalidRequestError("Invalid request body").ToJson2(w)
		logger.Logger.Errorw("Error decoding request body", "method", r.Method, "error", err, "time", time.Now())
		return
	}

	err = validate.Struct(requestBody)
	if err != nil {
		errs.ValidationError("Invalid request body").ToJson2(w)
		logger.Logger.Errorw("Validation error", "method", r.Method, "error", err, "request_body", requestBody, "time", time.Now())
		return
	}

	game := &entities.Game{
		GameName:   requestBody.Name,
		MaxPlayers: requestBody.MaxPlayers,
		MinPlayers: requestBody.MinPlayers,
		Instances:  requestBody.Instances,
	}

	gameId, err := g.gameService.CreateGame(r.Context(), game)
	if err != nil {
		errs.DBError("Could not create game").ToJson2(w)
		logger.Logger.Errorw("Failed to create game", "method", r.Method, "error", err, "request_body", requestBody, "time", time.Now())
		return
	}

	game.GameID = gameId

	w.Header().Set("Content-Type", "application/json")
	jsonResponse := map[string]any{
		"code":    http.StatusOK,
		"message": "Success",
		"game":    game,
	}
	if err = utils.JsonEncoder(w, jsonResponse); err != nil {
		return
	}
	logger.Logger.Infow("Game created successfully", "method", r.Method, "game_id", gameId, "time", time.Now())
}

func (g *GameHandler) GetGameByIdHandler(w http.ResponseWriter, r *http.Request) {
	logger.Logger.Infow("Handling GetGameById request", "method", r.Method, "time", time.Now())

	vars := mux.Vars(r)
	gameIdStr := vars["id"]
	gameId, err := uuid.Parse(gameIdStr)
	if err != nil {
		errs.ValidationError("Could not parse gameID").ToJson2(w)
		logger.Logger.Errorw("Error parsing gameID", "method", r.Method, "game_id", gameIdStr, "error", err, "time", time.Now())
		return
	}

	game, err := g.gameService.GetGameByID(r.Context(), gameId)
	if err != nil {
		errs.DBError("Failed to fetch game").ToJson2(w)
		logger.Logger.Errorw("Failed to fetch game by ID", "method", r.Method, "game_id", gameId, "error", err, "time", time.Now())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jsonResponse := map[string]any{
		"code":    http.StatusOK,
		"message": "Success",
		"game":    game,
	}
	if err = utils.JsonEncoder(w, jsonResponse); err != nil {
		return
	}
	logger.Logger.Infow("Successfully fetched game by ID", "method", r.Method, "game_id", gameId, "time", time.Now())
}

func (g *GameHandler) UpdateGameStatusHandler(w http.ResponseWriter, r *http.Request) {
	logger.Logger.Infow("Handling UpdateGameStatus request", "method", r.Method, "time", time.Now())

	// Extract and parse the game ID from the URL
	vars := mux.Vars(r)
	gameIdStr := vars["id"]
	gameId, err := uuid.Parse(gameIdStr)
	if err != nil {
		// Use ValidationError if the gameID is invalid
		errs.ValidationError("Invalid game ID format").ToJson2(w)
		logger.Logger.Errorw("Error parsing gameID", "method", r.Method, "game_id", gameIdStr, "error", err, "time", time.Now())
		return
	}

	// Attempt to update the game status
	err = g.gameService.UpdateGameStatus(r.Context(), gameId)
	if err != nil {
		if errors.Is(err, errs.ErrGameNotFound) {
			// Use NotFoundError when the game isn't found
			errs.InvalidRequestError("Game not found").ToJson2(w)
			logger.Logger.Errorw("Game not found", "method", r.Method, "game_id", gameId, "error", err, "time", time.Now())
			return
		}
		// Use UnexpectedError for any other unknown errors
		errs.UnexpectedError("Failed to update game status").ToJson2(w)
		logger.Logger.Errorw("Failed to update game status", "method", r.Method, "game_id", gameId, "error", err, "time", time.Now())
		return
	}

	// Respond with a success message
	w.Header().Set("Content-Type", "application/json")
	jsonResponse := map[string]any{
		"code":    http.StatusOK,
		"message": "Game status updated successfully",
	}
	if err = utils.JsonEncoder(w, jsonResponse); err != nil {
		return
	}
	logger.Logger.Infow("Game status updated successfully", "method", r.Method, "game_id", gameId, "time", time.Now())
}

func (g *GameHandler) DeleteGameHandler(w http.ResponseWriter, r *http.Request) {
	logger.Logger.Infow("Handling DeleteGame request", "method", r.Method, "time", time.Now())

	vars := mux.Vars(r)
	gameIdStr := vars["id"]
	gameId, err := uuid.Parse(gameIdStr)
	if err != nil {
		errs.ValidationError("Could not parse gameID").ToJson2(w)
		logger.Logger.Errorw("Error parsing gameID", "method", r.Method, "game_id", gameIdStr, "error", err, "time", time.Now())
		return
	}

	err = g.gameService.DeleteGame(r.Context(), gameId)
	if err != nil {
		if errors.Is(err, errs.ErrGameNotFound) {
			errs.DBError("Failed to delete game: Game not found").ToJson2(w)
			logger.Logger.Errorw("Game not found", "method", r.Method, "game_id", gameId, "error", err, "time", time.Now())
		} else {
			errs.DBError("Failed to delete game").ToJson2(w)
			logger.Logger.Errorw("Failed to delete game", "method", r.Method, "game_id", gameId, "error", err, "time", time.Now())
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jsonResponse := map[string]any{
		"code":    http.StatusOK,
		"message": "Game deleted successfully",
	}
	if err = utils.JsonEncoder(w, jsonResponse); err != nil {
		return
	}
	logger.Logger.Infow("Game deleted successfully", "method", r.Method, "game_id", gameId, "time", time.Now())
}
