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
	games, err := g.gameService.GetAllGames(r.Context())
	if err != nil {
		errs.NewInternalServerError("Could not fetch the games").ToJSON(w)
		logger.Logger.Errorw("Error decoding request body", "method", r.Method, "error", err, "time", time.Now())
		return
	}

	// Return the token as a JSON response
	w.Header().Set("Content-Type", "application/json")
	jsonResponse := map[string]any{
		"code":    http.StatusOK,
		"message": "Success",
		"data":    games,
	}

	json.NewEncoder(w).Encode(jsonResponse)
	logger.Logger.Infow("Sent all games successfully", "method", r.Method, "time", time.Now())

}

func (g *GameHandler) CreateGameHandler(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		Name       string `json:"name" validate:"required"`
		MaxPlayers int    `json:"max_players" validate:"required"`
		MinPlayers int    `json:"min_players" validate:"required"`
		Instances  int    `json:"instances" validate:"required"`
	}

	// decode the request body
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		logger.Logger.Errorw("Error decoding request body", "method", r.Method, "error", err, "time", time.Now())
		return
	}

	// Validate the request body
	err = validate.Struct(requestBody)
	if err != nil {
		errs.NewInvalidBodyError("Invalid request body").ToJSON(w)
		logger.Logger.Errorw("Validation error", "method", r.Method, "error", err, "request", requestBody, "time", time.Now())
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
		errs.NewInternalServerError("Could not create game").ToJSON(w)
		logger.Logger.Errorw("Validation error", "method", r.Method, "error", err, "request", requestBody, "time", time.Now())
		return
	}

	game.GameID = gameId

	w.Header().Set("Content-Type", "application/json")
	jsonResponse := map[string]any{
		"code":    http.StatusOK,
		"message": "Success",
		"data":    game,
	}
	json.NewEncoder(w).Encode(jsonResponse)
	logger.Logger.Infow("Signup Successful", "method", r.Method, "request", requestBody, "time", time.Now())

}

func (g *GameHandler) GetGameByIdHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameIdStr := vars["id"]
	gameId, err := uuid.Parse(gameIdStr)
	if err != nil {
		errs.NewInternalServerError("Could not parse gameID to convert to uuid").ToJSON(w)
		logger.Logger.Errorw("Validation error", "method", r.Method, "error", err, "time", time.Now())
		return
	}

	game, err := g.gameService.GetGameByID(r.Context(), gameId)
	if err != nil {
		errs.NewInternalServerError("failed to fetch the game by Id").ToJSON(w)
		logger.Logger.Errorw("Validation error", "method", r.Method, "error", err, "time", time.Now())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jsonResponse := map[string]any{
		"code":    http.StatusOK,
		"message": "Success",
		"data":    game,
	}
	json.NewEncoder(w).Encode(jsonResponse)
	logger.Logger.Infow("Signup Successful", "method", r.Method, "time", time.Now())
}

func (g *GameHandler) UpdateGameStatusHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameIdStr := vars["id"]
	gameId, err := uuid.Parse(gameIdStr)
	if err != nil {
		errs.NewInternalServerError("Could not parse gameID to convert to uuid").ToJSON(w)
		logger.Logger.Errorw("Validation error", "method", r.Method, "error", err, "time", time.Now())
		return
	}
	err = g.gameService.UpdateGameStatus(r.Context(), gameId)
	if err != nil {
		if errors.Is(err, errs.ErrGameNotFound) {
			errs.NewInvalidParameterError("").ToJSON(w)
			logger.Logger.Errorw("Validation error", "method", r.Method, "error", err, "time", time.Now())
			return
		}
		errs.NewInternalServerError("Could not update the Game status").ToJSON(w)
		logger.Logger.Errorw("Validation error", "method", r.Method, "error", err, "time", time.Now())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jsonResponse := map[string]any{
		"code":    http.StatusOK,
		"message": "Success",
	}
	json.NewEncoder(w).Encode(jsonResponse)
	logger.Logger.Infow("Signup Successful", "method", r.Method, "time", time.Now())

}

func (g *GameHandler) DeleteGameHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameIdStr := vars["id"]
	gameId, err := uuid.Parse(gameIdStr)
	if err != nil {
		errs.NewInternalServerError("Could not parse gameID to convert to uuid").ToJSON(w)
		logger.Logger.Errorw("Validation error", "method", r.Method, "error", err, "time", time.Now())
		return
	}
	err = g.gameService.DeleteGame(r.Context(), gameId)
	if err != nil {
		errs.NewInternalServerError("Failed to delete game").ToJSON(w)
		logger.Logger.Errorw("Validation error", "method", r.Method, "error", err, "time", time.Now())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jsonResponse := map[string]any{
		"code":    http.StatusOK,
		"message": "Success",
	}
	json.NewEncoder(w).Encode(jsonResponse)
	logger.Logger.Infow("Game deleted successfully", "method", r.Method, "time", time.Now())
}
