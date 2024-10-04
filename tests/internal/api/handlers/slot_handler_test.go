package handlers_test

import (
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"project2/internal/domain/entities"
	mocks "project2/tests/mocks/service"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"project2/internal/api/handlers"
)

func TestGetSlotByIdHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSlotService := mocks.NewMockSlotService(ctrl)
	handler := handlers.NewSlotHandler(mockSlotService)

	t.Run("success with valid slot ID", func(t *testing.T) {
		slotId := uuid.New()
		mockSlotService.EXPECT().GetSlotByID(gomock.Any(), slotId).Return(&entities.Slot{SlotID: slotId}, nil)

		req := httptest.NewRequest(http.MethodGet, "/slots/"+slotId.String(), nil)
		rr := httptest.NewRecorder()
		router := mux.NewRouter()
		router.HandleFunc("/slots/{id}", handler.GetSlotByIdHandler)
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var responseBody map[string]interface{}
		err := json.NewDecoder(rr.Body).Decode(&responseBody)
		assert.NoError(t, err)
		assert.Equal(t, "Success", responseBody["message"])
		assert.NotNil(t, responseBody["slot"])
	})

	t.Run("failure due to invalid slot ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/slots/invalid-uuid", nil)
		rr := httptest.NewRecorder()
		router := mux.NewRouter()
		router.HandleFunc("/slots/{id}", handler.GetSlotByIdHandler)
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		var responseBody map[string]interface{}
		err := json.NewDecoder(rr.Body).Decode(&responseBody)
		assert.NoError(t, err)
		assert.Equal(t, "Couldn't parse slot id", responseBody["message"])
	})

	t.Run("failure when slot not found", func(t *testing.T) {
		slotId := uuid.New()
		mockSlotService.EXPECT().GetSlotByID(gomock.Any(), slotId).Return(nil, errors.New("slot not found"))

		req := httptest.NewRequest(http.MethodGet, "/slots/"+slotId.String(), nil)
		rr := httptest.NewRecorder()
		router := mux.NewRouter()
		router.HandleFunc("/slots/{id}", handler.GetSlotByIdHandler)
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		var responseBody map[string]interface{}
		err := json.NewDecoder(rr.Body).Decode(&responseBody)
		assert.NoError(t, err)
		assert.Equal(t, "Couldn't get slot", responseBody["message"])
	})
}

func TestGetTodaySlotsHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSlotService := mocks.NewMockSlotService(ctrl)
	handler := handlers.NewSlotHandler(mockSlotService)

	t.Run("success fetching today's slots", func(t *testing.T) {
		gameId := uuid.New()
		mockSlotService.EXPECT().GetCurrentDayGameSlots(gomock.Any(), gameId).Return([]entities.Slot{{GameID: gameId}}, nil)

		req := httptest.NewRequest(http.MethodGet, "/slots/today/"+gameId.String(), nil)
		rr := httptest.NewRecorder()
		router := mux.NewRouter()
		router.HandleFunc("/slots/today/{gameID}", handler.GetTodaySlotsHandler)
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var responseBody map[string]interface{}
		err := json.NewDecoder(rr.Body).Decode(&responseBody)
		assert.NoError(t, err)
		assert.Equal(t, "Success", responseBody["message"])
		assert.NotNil(t, responseBody["slots"])
	})

	t.Run("failure due to invalid game ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/slots/today/invalid-uuid", nil)
		rr := httptest.NewRecorder()
		router := mux.NewRouter()
		router.HandleFunc("/slots/today/{gameID}", handler.GetTodaySlotsHandler)
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		var responseBody map[string]interface{}
		err := json.NewDecoder(rr.Body).Decode(&responseBody)
		assert.NoError(t, err)
		assert.Equal(t, "Couldn't parse game id", responseBody["message"])
	})

	t.Run("failure when today's slots not found", func(t *testing.T) {
		gameId := uuid.New()
		mockSlotService.EXPECT().GetCurrentDayGameSlots(gomock.Any(), gameId).Return(nil, errors.New("no slots available"))

		req := httptest.NewRequest(http.MethodGet, "/slots/today/"+gameId.String(), nil)
		rr := httptest.NewRecorder()
		router := mux.NewRouter()
		router.HandleFunc("/slots/today/{gameID}", handler.GetTodaySlotsHandler)
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		var responseBody map[string]interface{}
		err := json.NewDecoder(rr.Body).Decode(&responseBody)
		assert.NoError(t, err)
		assert.Equal(t, "Couldn't get slots", responseBody["message"])
	})
}
