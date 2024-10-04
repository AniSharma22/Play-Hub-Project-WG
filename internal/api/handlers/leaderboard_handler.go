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
	"project2/pkg/utils"
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

	// Start log
	logger.Logger.Infow("Processing request to get game leaderboard", "gameID", gameIdStr, "method", r.Method, "time", time.Now())

	gameId, err := uuid.Parse(gameIdStr)
	if err != nil {
		logger.Logger.Errorw("Error parsing game ID", "gameID", gameIdStr, "error", err, "time", time.Now())
		errs.ValidationError("Couldn't parse game id").ToJson2(w)
		return
	}

	leaderboard, err := l.leaderBoardService.GetGameLeaderboard(r.Context(), gameId)
	if err != nil {
		logger.Logger.Errorw("Error fetching leaderboard", "gameID", gameIdStr, "error", err, "time", time.Now())
		errs.DBError("Couldn't get leaderboard").ToJson2(w)
		return
	}

	// Respond with leaderboard
	w.Header().Set("Content-Type", "application/json")
	jsonResponse := map[string]any{
		"code":    http.StatusOK,
		"message": "Success",
		"leaderboard": func() []models.Leaderboard {
			if leaderboard == nil {
				return []models.Leaderboard{}
			}
			return leaderboard
		}(),
	}
	if err = utils.JsonEncoder(w, jsonResponse); err != nil {
		return
	}
	logger.Logger.Infow("Successfully fetched game leaderboard", "gameID", gameIdStr, "method", r.Method, "time", time.Now())
}

func (l *LeaderboardHandler) RecordUserResultHandler(w http.ResponseWriter, r *http.Request) {
	userIdStr, ok := r.Context().Value(middleware.UserIdKey).(string)
	if !ok {
		logger.Logger.Errorw("Error finding userId in context", "method", r.Method, "time", time.Now())
		errs.InvalidRequestError("Could not find the userId").ToJson2(w)
		return
	}

	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		logger.Logger.Errorw("Error parsing user ID", "userID", userIdStr, "error", err, "time", time.Now())
		errs.ValidationError("Couldn't parse user id").ToJson2(w)
		return
	}

	var requestBody struct {
		GameId    string `json:"game_id" validate:"required"`
		BookingId string `json:"booking_id" validate:"required"`
		Result    string `json:"result" validate:"required"`
	}

	err = json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		logger.Logger.Errorw("Error decoding request body", "method", r.Method, "error", err, "time", time.Now())
		errs.InvalidRequestError("User id is wrong").ToJson2(w)
		return
	}

	err = validate.Struct(requestBody)
	if err != nil {
		logger.Logger.Errorw("Validation error", "method", r.Method, "error", err, "requestBody", requestBody, "time", time.Now())
		errs.ValidationError("Invalid request body").ToJson2(w)
		return
	}

	gameId, err := uuid.Parse(requestBody.GameId)
	if err != nil {
		logger.Logger.Errorw("Error parsing game ID", "gameID", requestBody.GameId, "error", err, "time", time.Now())
		errs.ValidationError("Couldn't parse game id").ToJson2(w)
		return
	}
	bookingId, err := uuid.Parse(requestBody.BookingId)
	if err != nil {
		logger.Logger.Errorw("Error parsing booking ID", "bookingID", requestBody.BookingId, "error", err, "time", time.Now())
		errs.ValidationError("Couldn't parse booking id").ToJson2(w)
		return
	}

	switch requestBody.Result {
	case "win":
		err = l.leaderBoardService.AddWinToUser(r.Context(), userId, gameId, bookingId)
		if err != nil {
			logger.Logger.Errorw("Error adding win to user", "userID", userIdStr, "gameID", requestBody.GameId, "error", err, "time", time.Now())
			errs.DBError("Couldn't add win to user").ToJson2(w)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		jsonResponse := map[string]any{
			"code":    http.StatusOK,
			"message": "Success",
		}
		if err = utils.JsonEncoder(w, jsonResponse); err != nil {
			return
		}
		logger.Logger.Infow("Win added successfully", "userID", userIdStr, "gameID", requestBody.GameId, "method", r.Method, "time", time.Now())

	case "loss":
		err = l.leaderBoardService.AddLossToUser(r.Context(), userId, gameId, bookingId)
		if err != nil {
			logger.Logger.Errorw("Error adding loss to user", "userID", userIdStr, "gameID", requestBody.GameId, "error", err, "time", time.Now())
			errs.DBError("Couldn't add win to user").ToJson2(w)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		jsonResponse := map[string]any{
			"code":    http.StatusOK,
			"message": "Success",
		}
		if err = utils.JsonEncoder(w, jsonResponse); err != nil {
			return
		}
		logger.Logger.Infow("Loss added successfully", "userID", userIdStr, "gameID", requestBody.GameId, "method", r.Method, "time", time.Now())
	default:
		errs.InvalidRequestError("Invalid or missing result parameter in the URL").ToJson2(w)
		logger.Logger.Errorw("Invalid or no parameter sent in the URl", "userID", userIdStr, "gameID", requestBody.GameId, "parameter", requestBody.Result, "time", time.Now())
		return
	}

}
