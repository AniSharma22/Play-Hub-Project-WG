package handlers

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"net/http"
	"project2/internal/api/middleware"
	service_interfaces "project2/internal/domain/interfaces/service"
	"project2/internal/models"
	"project2/pkg/errs"
	"project2/pkg/logger"
	"time"
)

type LeaderboardHandler struct {
	leaderBoardService service_interfaces.LeaderboardService
}

func NewLeaderboardHandler(leaderBoardService service_interfaces.LeaderboardService) *LeaderboardHandler {
	return &LeaderboardHandler{
		leaderBoardService: leaderBoardService,
	}
}

func (l *LeaderboardHandler) GetGameLeaderboardHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameIdStr := vars["gameID"]

	gameId, err := uuid.Parse(gameIdStr)
	if err != nil {
		errs.NewInternalServerError("Couldn't parse user id").ToJSON(w)
		return
	}

	leaderboard, err := l.leaderBoardService.GetGameLeaderboard(r.Context(), gameId)
	if err != nil {
		errs.NewInternalServerError("Couldn't get leaderboard").ToJSON(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jsonResponse := map[string]any{
		"code":    http.StatusOK,
		"message": "Success",
		"data": func() []models.Leaderboard {
			if leaderboard == nil {
				return []models.Leaderboard{}
			}
			return leaderboard
		}(),
	}
	json.NewEncoder(w).Encode(jsonResponse)
	logger.Logger.Infow("Leaderboard successfully sent", "method", r.Method, "time", time.Now())
}

func (l *LeaderboardHandler) AddLossToUserHandler(w http.ResponseWriter, r *http.Request) {

	// Extract userId from the context
	userIdStr, ok := r.Context().Value(middleware.UserIdKey).(string)
	if !ok {
		errs.NewUnexpectedError("Could not find the userId").ToJSON(w)
		return
	}

	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		errs.NewInternalServerError("Couldn't parse user id").ToJSON(w)
		return
	}

	var requestBody struct {
		GameId    string `json:"game_id" validate:"required"`
		BookingId string `json:"booking_id" validate:"required"`
	}
	// decode the request body
	err = json.NewDecoder(r.Body).Decode(&requestBody)
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

	gameId, err := uuid.Parse(requestBody.GameId)
	if err != nil {
		errs.NewInternalServerError("Couldn't parse user id").ToJSON(w)
		return
	}
	bookingId, err := uuid.Parse(requestBody.BookingId)
	if err != nil {
		errs.NewInternalServerError("Couldn't parse user id").ToJSON(w)
		return
	}

	err = l.leaderBoardService.AddLossToUser(r.Context(), userId, gameId, bookingId)
	if err != nil {
		errs.NewInternalServerError("Couldn't add loss to user").ToJSON(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jsonResponse := map[string]any{
		"code":    http.StatusOK,
		"message": "Success",
	}
	json.NewEncoder(w).Encode(jsonResponse)
	logger.Logger.Infow("Loss Added successfully to the user", "method", r.Method, "time", time.Now())

}

func (l *LeaderboardHandler) AddWinToUserHandler(w http.ResponseWriter, r *http.Request) {
	// Extract userId from the context
	userIdStr, ok := r.Context().Value(middleware.UserIdKey).(string)
	if !ok {
		errs.NewUnexpectedError("Could not find the userId").ToJSON(w)
		return
	}

	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		errs.NewInternalServerError("Couldn't parse user id").ToJSON(w)
		return
	}

	var requestBody struct {
		GameId    string `json:"game_id" validate:"required"`
		BookingId string `json:"booking_id" validate:"required"`
	}
	// decode the request body
	err = json.NewDecoder(r.Body).Decode(&requestBody)
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

	gameId, err := uuid.Parse(requestBody.GameId)
	if err != nil {
		errs.NewInternalServerError("Couldn't parse user id").ToJSON(w)
		return
	}
	bookingId, err := uuid.Parse(requestBody.BookingId)
	if err != nil {
		errs.NewInternalServerError("Couldn't parse user id").ToJSON(w)
		return
	}

	err = l.leaderBoardService.AddWinToUser(r.Context(), userId, gameId, bookingId)
	if err != nil {
		errs.NewInternalServerError("Couldn't add loss to user").ToJSON(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jsonResponse := map[string]any{
		"code":    http.StatusOK,
		"message": "Success",
	}
	json.NewEncoder(w).Encode(jsonResponse)
	logger.Logger.Infow("Loss Added successfully to the user", "method", r.Method, "time", time.Now())

}
