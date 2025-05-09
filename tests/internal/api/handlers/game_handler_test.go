package handlers_test

import (
	"encoding/json"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"project2/internal/api/handlers"
	"project2/internal/domain/entities"
	"project2/pkg/errs"
	mocks2 "project2/tests/mocks/service"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAllGamesHandler_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockGameService := mocks2.NewMockGameService(ctrl)
	handler := handlers.NewGameHandler(mockGameService)

	mockGames := []entities.Game{
		{GameID: uuid.New(), GameName: "Table Tennis", MaxPlayers: 4, MinPlayers: 2},
		{GameID: uuid.New(), GameName: "Air Hockey", MaxPlayers: 2, MinPlayers: 2},
	}

	mockGameService.EXPECT().GetAllGames(gomock.Any()).Return(mockGames, nil)

	req, err := http.NewRequest(http.MethodGet, "/games", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler.GetAllGamesHandler(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]any
	err = json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Success", response["message"])
	assert.Len(t, response["games"].([]interface{}), 2)
}

func TestGetAllGamesHandler_Failure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockGameService := mocks2.NewMockGameService(ctrl)
	handler := handlers.NewGameHandler(mockGameService)

	mockGameService.EXPECT().GetAllGames(gomock.Any()).Return(nil, errs.NewInternalServerError("Could not fetch the games"))

	req, err := http.NewRequest(http.MethodGet, "/games", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler.GetAllGamesHandler(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestCreateGameHandler_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockGameService := mocks2.NewMockGameService(ctrl)
	handler := handlers.NewGameHandler(mockGameService)

	mockGame := &entities.Game{
		GameID:     uuid.New(),
		GameName:   "Table Tennis",
		MaxPlayers: 4,
		MinPlayers: 2,
		Instances:  1,
	}

	mockGameService.EXPECT().CreateGame(gomock.Any(), gomock.Any()).Return(mockGame.GameID, nil)

	requestBody := `{"name":"Table Tennis","max_players":4,"min_players":2,"instances":1}`
	req, err := http.NewRequest(http.MethodPost, "/games", strings.NewReader(requestBody))
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler.CreateGameHandler(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]any
	err = json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Success", response["message"])
	//assert.Equal(t, mockGame.GameName, response["game"].(map[string]interface{})["GameName"])
}

func TestCreateGameHandler_InvalidInput(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockGameService := mocks2.NewMockGameService(ctrl)
	handler := handlers.NewGameHandler(mockGameService)

	// Sending invalid request body
	requestBody := `{"name":"","max_players":0,"min_players":0,"instances":0}`
	req, err := http.NewRequest(http.MethodPost, "/games", strings.NewReader(requestBody))
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler.CreateGameHandler(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestGetGameByIdHandler_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockGameService := mocks2.NewMockGameService(ctrl)
	handler := handlers.NewGameHandler(mockGameService)

	mockGame := &entities.Game{
		GameID:     uuid.New(),
		GameName:   "Table Tennis",
		MaxPlayers: 4,
		MinPlayers: 2,
	}

	mockGameService.EXPECT().GetGameByID(gomock.Any(), mockGame.GameID).Return(mockGame, nil)

	// Create a new router and register the handler
	router := mux.NewRouter()
	router.HandleFunc("/games/{id}", handler.GetGameByIdHandler).Methods("GET")

	req, err := http.NewRequest(http.MethodGet, "/games/"+mockGame.GameID.String(), nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()

	// Serve the request using the router
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]any
	err = json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Success", response["message"])
	//assert.Equal(t, mockGame.GameName, response["game"].(map[string]interface{})["GameName"])
}

func TestGetGameByIdHandler_NotFound(t *testing.T) {
	// Initialize gomock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a mock service
	mockGameService := mocks2.NewMockGameService(ctrl)
	handler := handlers.NewGameHandler(mockGameService)

	// Create a gameID and set up the mock service expectation
	gameID := uuid.New()
	mockGameService.EXPECT().GetGameByID(gomock.Any(), gameID).Return(nil, errs.NewNotFoundError("Game not found"))

	// Create a new router and register the handler
	router := mux.NewRouter()
	router.HandleFunc("/games/{id}", handler.GetGameByIdHandler).Methods("GET")

	// Create a new request using the gameID in the URL
	req, err := http.NewRequest(http.MethodGet, "/games/"+gameID.String(), nil)
	assert.NoError(t, err)

	// Create a response recorder to capture the output
	rr := httptest.NewRecorder()

	// Serve the request using the router
	router.ServeHTTP(rr, req)

	// Assert that the status code is 404 Not Found
	assert.Equal(t, http.StatusNotFound, rr.Code)

	// Optionally, assert the response body to match the error message
	var responseBody map[string]string
	err = json.NewDecoder(rr.Body).Decode(&responseBody)
	assert.NoError(t, err)
	assert.Equal(t, "Failed to fetch game", responseBody["message"])
}

func TestDeleteGameHandler_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockGameService := mocks2.NewMockGameService(ctrl)
	handler := handlers.NewGameHandler(mockGameService)

	gameID := uuid.New()

	mockGameService.EXPECT().DeleteGame(gomock.Any(), gameID).Return(nil)

	// Create a new router and register the handler
	router := mux.NewRouter()
	router.HandleFunc("/games/{id}", handler.DeleteGameHandler).Methods("DELETE")

	req, err := http.NewRequest(http.MethodDelete, "/games/"+gameID.String(), nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()

	// Serve the request using the router
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestDeleteGameHandler_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockGameService := mocks2.NewMockGameService(ctrl)
	handler := handlers.NewGameHandler(mockGameService)

	gameID := uuid.New()
	mockGameService.EXPECT().DeleteGame(gomock.Any(), gameID).Return(errs.NewNotFoundError("Game not found"))

	// Create a new router and register the handler
	router := mux.NewRouter()
	router.HandleFunc("/games/{id}", handler.DeleteGameHandler).Methods("DELETE")

	req, err := http.NewRequest(http.MethodDelete, "/games/"+gameID.String(), nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	// Serve the request using the router
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestUpdateGameStatusHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGameService := mocks2.NewMockGameService(ctrl)
	handler := handlers.NewGameHandler(mockGameService)

	t.Run("success updating game status", func(t *testing.T) {
		gameID := uuid.New()
		mockGameService.EXPECT().UpdateGameStatus(gomock.Any(), gameID).Return(nil)

		req, err := http.NewRequest(http.MethodPut, "/games/"+gameID.String(), nil)
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		router := mux.NewRouter()
		router.HandleFunc("/games/{id}", handler.UpdateGameStatusHandler).Methods("PUT")
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var responseBody map[string]interface{}
		err = json.NewDecoder(rr.Body).Decode(&responseBody)
		assert.NoError(t, err)
		assert.Equal(t, "Success", responseBody["message"])
	})

	t.Run("failure due to invalid game ID", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPut, "/games/invalid-uuid", nil)
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		router := mux.NewRouter()
		router.HandleFunc("/games/{id}", handler.UpdateGameStatusHandler).Methods("PUT")
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		var responseBody map[string]interface{}
		err = json.NewDecoder(rr.Body).Decode(&responseBody)
		assert.NoError(t, err)
		assert.Equal(t, "Could not parse gameID", responseBody["message"])
	})

	t.Run("failure when game not found", func(t *testing.T) {
		gameID := uuid.New()
		mockGameService.EXPECT().UpdateGameStatus(gomock.Any(), gameID).Return(errs.ErrGameNotFound)

		req, err := http.NewRequest(http.MethodPut, "/games/"+gameID.String(), nil)
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		router := mux.NewRouter()
		router.HandleFunc("/games/{id}", handler.UpdateGameStatusHandler).Methods("PUT")
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		var responseBody map[string]interface{}
		err = json.NewDecoder(rr.Body).Decode(&responseBody)
		assert.NoError(t, err)
		assert.Equal(t, "Game not found", responseBody["message"])
	})

	t.Run("failure when unable to update game status", func(t *testing.T) {
		gameID := uuid.New()
		mockGameService.EXPECT().UpdateGameStatus(gomock.Any(), gameID).Return(errors.New("update failed"))

		req, err := http.NewRequest(http.MethodPut, "/games/"+gameID.String(), nil)
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		router := mux.NewRouter()
		router.HandleFunc("/games/{id}", handler.UpdateGameStatusHandler).Methods("PUT")
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		var responseBody map[string]interface{}
		err = json.NewDecoder(rr.Body).Decode(&responseBody)
		assert.NoError(t, err)
		assert.Equal(t, "Could not update the game status", responseBody["message"])
	})
}
