package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"project2/internal/api/handlers"
	"project2/internal/api/middleware"
	"project2/internal/models"
	"project2/pkg/errs"
	mocks "project2/tests/mocks/service"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetGameLeaderboardHandler_Success(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLeaderboardService := mocks.NewMockLeaderboardService(ctrl)
	handler := handlers.NewLeaderboardHandler(mockLeaderboardService)

	gameID := uuid.New()
	mockLeaderboard := []models.Leaderboard{
		{
			UserName: "test_user",
			Score:    50,
		},
	}

	mockLeaderboardService.EXPECT().GetGameLeaderboard(gomock.Any(), gameID).Return(mockLeaderboard, nil)

	// Create a new router and register the handler
	router := mux.NewRouter()
	router.HandleFunc("/leaderboard/{gameID}", handler.GetGameLeaderboardHandler).Methods("GET")

	// Create the request with the gameID in the URL
	req := httptest.NewRequest(http.MethodGet, "/leaderboard/"+gameID.String(), nil)
	rr := httptest.NewRecorder()

	// Act
	router.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]any
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "Success", response["message"])
	assert.NotEmpty(t, response["leaderboard"])
}

func TestGetGameLeaderboardHandler_InvalidGameID(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLeaderboardService := mocks.NewMockLeaderboardService(ctrl)
	handler := handlers.NewLeaderboardHandler(mockLeaderboardService)

	req := httptest.NewRequest("GET", "/leaderboard/invalid-game-id", nil)
	rr := httptest.NewRecorder()

	vars := map[string]string{"gameID": "invalid-game-id"}
	req = mux.SetURLVars(req, vars)

	// Act
	handler.GetGameLeaderboardHandler(rr, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	var response errs.AppError
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Couldn't parse game id", response.Message)
}

func TestGetGameLeaderboardHandler_ErrorFetchingLeaderboard(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLeaderboardService := mocks.NewMockLeaderboardService(ctrl)
	handler := handlers.NewLeaderboardHandler(mockLeaderboardService)
	gameID := uuid.New()

	mockLeaderboardService.EXPECT().GetGameLeaderboard(gomock.Any(), gameID).Return(nil, errors.New("error fetching leaderboard"))

	req := httptest.NewRequest("GET", "/leaderboard/"+gameID.String(), nil)
	rr := httptest.NewRecorder()

	vars := map[string]string{"gameID": gameID.String()}
	req = mux.SetURLVars(req, vars)

	// Act
	handler.GetGameLeaderboardHandler(rr, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	var response errs.AppError
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Couldn't get leaderboard", response.Message)
}

func TestRecordUserResultHandler_AddWinSuccess(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLeaderboardService := mocks.NewMockLeaderboardService(ctrl)
	handler := handlers.NewLeaderboardHandler(mockLeaderboardService)
	userID := uuid.New()
	gameID := uuid.New()
	bookingID := uuid.New()

	mockLeaderboardService.EXPECT().AddWinToUser(gomock.Any(), userID, gameID, bookingID).Return(nil)

	reqBody := map[string]string{
		"game_id":    gameID.String(),
		"booking_id": bookingID.String(),
		"result":     "win",
	}
	reqBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/result", bytes.NewBuffer(reqBytes))
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIdKey, userID.String()))
	rr := httptest.NewRecorder()

	// Act
	handler.RecordUserResultHandler(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]any
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "Success", response["message"])
}

func TestRecordUserResultHandler_AddLossSuccess(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLeaderboardService := mocks.NewMockLeaderboardService(ctrl)
	handler := handlers.NewLeaderboardHandler(mockLeaderboardService)
	userID := uuid.New()
	gameID := uuid.New()
	bookingID := uuid.New()

	mockLeaderboardService.EXPECT().AddLossToUser(gomock.Any(), userID, gameID, bookingID).Return(nil)

	reqBody := map[string]string{
		"game_id":    gameID.String(),
		"booking_id": bookingID.String(),
		"result":     "loss",
	}
	reqBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/result", bytes.NewBuffer(reqBytes))
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIdKey, userID.String()))
	rr := httptest.NewRecorder()

	// Act
	handler.RecordUserResultHandler(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]any
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "Success", response["message"])
}

func TestRecordUserResultHandler_InvalidResult(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLeaderboardService := mocks.NewMockLeaderboardService(ctrl)
	handler := handlers.NewLeaderboardHandler(mockLeaderboardService)
	userID := uuid.New()
	gameID := uuid.New()
	bookingID := uuid.New()

	reqBody := map[string]string{
		"game_id":    gameID.String(),
		"booking_id": bookingID.String(),
		"result":     "invalid",
	}
	reqBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/result", bytes.NewBuffer(reqBytes))
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIdKey, userID.String()))
	rr := httptest.NewRecorder()

	// Act
	handler.RecordUserResultHandler(rr, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var response errs.AppError
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid or missing result parameter in the URL", response.Message)
}

func TestRecordUserResultHandler_ErrorInAddingWin(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLeaderboardService := mocks.NewMockLeaderboardService(ctrl)
	handler := handlers.NewLeaderboardHandler(mockLeaderboardService)
	userID := uuid.New()
	gameID := uuid.New()
	bookingID := uuid.New()

	mockLeaderboardService.EXPECT().AddWinToUser(gomock.Any(), userID, gameID, bookingID).Return(errors.New("error adding win"))

	reqBody := map[string]string{
		"game_id":    gameID.String(),
		"booking_id": bookingID.String(),
		"result":     "win",
	}
	reqBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/result", bytes.NewBuffer(reqBytes))
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIdKey, userID.String()))
	rr := httptest.NewRecorder()

	// Act
	handler.RecordUserResultHandler(rr, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	var response errs.AppError
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Couldn't add win to user", response.Message)
}
