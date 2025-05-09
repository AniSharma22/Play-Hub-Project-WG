package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"net/http"
	"project2/internal/api/middleware"
	"project2/internal/domain/entities"
	service_interfaces "project2/internal/domain/interfaces/service"
	"project2/pkg/errs"
	"project2/pkg/logger"
	"project2/pkg/utils"
	"strconv"
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
		role, _ := r.Context().Value(middleware.RoleKey).(string)
		var games []entities.Game
		var err error

		if role == "admin" {
			games, err = g.gameService.GetAllGames(r.Context())
		} else {
			games, err = g.gameService.GetAllActiveGames(r.Context())
		}

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

func (g *GameHandler) CreateGameHandlerNew(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		errs.InvalidRequestError("Unable to parse multipart form").ToJson2(w)
		logger.Logger.Errorw("Error parsing new game form", "method", r.Method, "error", err, "time", time.Now())
		return
	}

	// Retrieve and validate the "name" field
	name := r.FormValue("name")
	if name == "" {
		errs.InvalidRequestError("Game name is required").ToJson2(w)
		logger.Logger.Warnw("Game name is missing", "method", r.Method, "time", time.Now())
		return
	}

	// Retrieve and validate the "max_players" field
	maxPlayersStr := r.FormValue("max_players")
	maxPlayers, err := strconv.Atoi(maxPlayersStr)
	if err != nil || maxPlayers <= 0 {
		errs.InvalidRequestError("Invalid max players value").ToJson2(w)
		logger.Logger.Warnw("Invalid max players", "method", r.Method, "value", maxPlayersStr, "time", time.Now())
		return
	}

	// Retrieve and validate the "min_players" field
	minPlayersStr := r.FormValue("min_players")
	minPlayers, err := strconv.Atoi(minPlayersStr)
	if err != nil || minPlayers <= 0 || minPlayers > maxPlayers {
		errs.InvalidRequestError("Invalid min players value").ToJson2(w)
		logger.Logger.Warnw("Invalid min players", "method", r.Method, "value", minPlayersStr, "time", time.Now())
		return
	}

	// Retrieve and validate the "instances" field
	instancesStr := r.FormValue("instances")
	instances, err := strconv.Atoi(instancesStr)
	if err != nil || instances <= 0 {
		errs.InvalidRequestError("Invalid instances value").ToJson2(w)
		logger.Logger.Warnw("Invalid instances", "method", r.Method, "value", instancesStr, "time", time.Now())
		return
	}

	// Retrieve and validate the "is_active" field
	isActiveStr := r.FormValue("isActive")
	isActive := isActiveStr == "true"
	if isActiveStr != "true" && isActiveStr != "false" {
		errs.InvalidRequestError("Invalid active value").ToJson2(w)
		logger.Logger.Warnw("Invalid isActive", "method", r.Method, "value", isActiveStr, "time", time.Now())
		return
	}

	// Retrieve and validate the file
	file, handler, err := r.FormFile("image")
	if err != nil {
		errs.UnexpectedError("Error parsing form file").ToJson2(w)
		logger.Logger.Errorw("Error parsing form file", "method", r.Method, "error", err, "time", time.Now())
		return
	}
	defer file.Close()

	// Check file size/type
	if handler.Size > 10*1024*1024 { // 10MB limit
		errs.InvalidRequestError("Image file size too large").ToJson2(w)
		logger.Logger.Warnw("Image file too large", "method", r.Method, "size", handler.Size, "time", time.Now())
		return
	}

	// Upload the file to S3 bucket
	imageUrl, err := utils.UploadFileToS3(file, handler.Filename)
	if err != nil {
		errs.UnexpectedError("Error uploading file").ToJson2(w)
		logger.Logger.Errorw("Error uploading file", "method", r.Method, "error", err, "time", time.Now())
		return
	}

	// Create a new game entry
	game := &entities.Game{
		GameName:   name,
		ImageUrl:   imageUrl,
		MaxPlayers: maxPlayers,
		MinPlayers: minPlayers,
		Instances:  instances,
		IsActive:   isActive,
	}

	// Call the service to create the game
	gameId, err := g.gameService.CreateGame(r.Context(), game)
	if err != nil {
		errs.DBError("Could not create game").ToJson2(w)
		logger.Logger.Errorw("Failed to create game", "method", r.Method, "error", err, "request_body", game, "time", time.Now())
		return
	}

	// Set the unique game ID generated by the service
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

func (g *GameHandler) UpdateGameHandler(w http.ResponseWriter, r *http.Request) {
	logger.Logger.Infow("Handling UpdateGameStatus request", "method", r.Method, "time", time.Now())

	// Extract and parse the game ID from the URL
	vars := mux.Vars(r)
	gameIdStr := vars["id"]
	gameId, err := uuid.Parse(gameIdStr)
	if err != nil {
		errs.ValidationError("Invalid game ID format").ToJson2(w)
		logger.Logger.Errorw("Error parsing gameID", "method", r.Method, "game_id", gameIdStr, "error", err, "time", time.Now())
		return
	}

	// Parse form data
	err = r.ParseMultipartForm(10 << 20)
	if err != nil {
		errs.InvalidRequestError("Unable to parse form").ToJson2(w)
		logger.Logger.Errorw("Error parsing game form", "method", r.Method, "error", err, "time", time.Now())
		return
	}
	fmt.Println(r.Form)
	// Extract fields and validate
	gameName := r.FormValue("name")
	if gameName == "" {
		errs.ValidationError("Game name is required").ToJson2(w)
		logger.Logger.Warnw("Game name is missing", "method", r.Method, "time", time.Now())
		return
	}

	minPlayers, err := strconv.Atoi(r.FormValue("min_players"))
	if err != nil {
		errs.ValidationError("Invalid minimum players value").ToJson2(w)
		logger.Logger.Warnw("Invalid minimum players", "method", r.Method, "value", r.FormValue("min_players"), "time", time.Now())
		return
	}

	maxPlayers, err := strconv.Atoi(r.FormValue("max_players"))
	if err != nil {
		errs.ValidationError("Invalid maximum players value").ToJson2(w)
		logger.Logger.Warnw("Invalid maximum players", "method", r.Method, "value", r.FormValue("max_players"), "time", time.Now())
		return
	}

	instances, err := strconv.Atoi(r.FormValue("instances"))
	if err != nil {
		errs.ValidationError("Invalid instance value").ToJson2(w)
		logger.Logger.Warnw("Invalid instance", "method", r.Method, "value", r.FormValue("instance"), "time", time.Now())
		return
	}

	// Retrieve and validate the "is_active" field
	isActiveStr := r.FormValue("isActive")
	if isActiveStr != "true" && isActiveStr != "false" {
		errs.InvalidRequestError("Invalid active value").ToJson2(w)
		logger.Logger.Warnw("Invalid isActive", "method", r.Method, "value", isActiveStr, "time", time.Now())
		return
	}
	isActive := isActiveStr == "true"

	// Create game object
	game := &entities.Game{
		GameID:     gameId,
		GameName:   gameName,
		MinPlayers: minPlayers,
		MaxPlayers: maxPlayers,
		Instances:  instances,
		IsActive:   isActive,
	}

	// Attempt to update the game details
	err = g.gameService.UpdateGame(r.Context(), game)
	if err != nil {
		if errors.Is(err, errs.ErrGameNotFound) {
			errs.InvalidRequestError("Game not found").ToJson2(w)
			logger.Logger.Errorw("Game not found", "method", r.Method, "game_id", gameId, "error", err, "time", time.Now())
			return
		}
		errs.UnexpectedError("Failed to update game status").ToJson2(w)
		logger.Logger.Errorw("Failed to update game status", "method", r.Method, "game_id", gameId, "error", err, "time", time.Now())
		return
	}

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	jsonResponse := map[string]any{
		"code":    http.StatusOK,
		"message": "Game status updated successfully",
	}
	if err = utils.JsonEncoder(w, jsonResponse); err != nil {
		logger.Logger.Errorw("Failed to encode JSON response", "error", err)
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
