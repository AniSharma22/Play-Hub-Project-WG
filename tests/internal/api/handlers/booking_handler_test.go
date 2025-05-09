package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"project2/internal/api/middleware"
	"project2/internal/models"
	mocks "project2/tests/mocks/service"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"project2/internal/api/handlers"
	"project2/pkg/errs"
)

type jsonResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func TestCreateBookingHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBookingService := mocks.NewMockBookingService(ctrl)
	handler := handlers.NewBookingHandler(mockBookingService)

	t.Run("success with valid booking", func(t *testing.T) {
		userId := uuid.NewString()
		slotId := uuid.NewString()
		gameId := uuid.NewString()
		requestBody := []byte(`{"slot_id":"` + slotId + `", "game_id":"` + gameId + `"}`)

		ctx := context.WithValue(context.Background(), middleware.UserIdKey, userId)
		mockBookingService.EXPECT().MakeBooking(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

		req := httptest.NewRequest(http.MethodPost, "/book", bytes.NewBuffer(requestBody)).WithContext(ctx)
		rr := httptest.NewRecorder()
		handler.CreateBookingHandler(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var responseBody jsonResponse
		err := json.NewDecoder(rr.Body).Decode(&responseBody)
		assert.NoError(t, err)
		assert.Equal(t, "Booking created successfully", responseBody.Message)
	})

	t.Run("failure due to missing userId", func(t *testing.T) {
		requestBody := []byte(`{"slot_id":"some-slot-id", "game_id":"some-game-id"}`)
		req := httptest.NewRequest(http.MethodPost, "/book", bytes.NewBuffer(requestBody))
		rr := httptest.NewRecorder()
		handler.CreateBookingHandler(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		var responseBody jsonResponse
		err := json.NewDecoder(rr.Body).Decode(&responseBody)
		assert.NoError(t, err)
		assert.Equal(t, "Could not find the userId", responseBody.Message)
	})

	t.Run("failure due to invalid request body", func(t *testing.T) {
		userId := uuid.NewString()
		req := httptest.NewRequest(http.MethodPost, "/book", bytes.NewBuffer([]byte(`{"slot_id":"123", "game_id":"456"}`))).WithContext(context.WithValue(context.Background(), middleware.UserIdKey, userId))
		rr := httptest.NewRecorder()
		handler.CreateBookingHandler(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		var responseBody jsonResponse
		err := json.NewDecoder(rr.Body).Decode(&responseBody)
		assert.NoError(t, err)
		assert.Equal(t, "Couldn't parse slot ID", responseBody.Message)
	})

	t.Run("failure due to slot timing passed", func(t *testing.T) {
		userId := uuid.NewString()
		slotId := uuid.NewString()
		gameId := uuid.NewString()
		requestBody := []byte(`{"slot_id":"` + slotId + `", "game_id":"` + gameId + `"}`)

		ctx := context.WithValue(context.Background(), middleware.UserIdKey, userId)
		mockBookingService.EXPECT().MakeBooking(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errs.ErrSlotPassed)

		req := httptest.NewRequest(http.MethodPost, "/book", bytes.NewBuffer(requestBody)).WithContext(ctx)
		rr := httptest.NewRecorder()
		handler.CreateBookingHandler(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		var responseBody jsonResponse
		err := json.NewDecoder(rr.Body).Decode(&responseBody)
		assert.NoError(t, err)
		assert.Equal(t, "Slot timing has already passed", responseBody.Message)
	})
}

func TestGetUserBookingsHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBookingService := mocks.NewMockBookingService(ctrl)
	handler := handlers.NewBookingHandler(mockBookingService)

	t.Run("success fetching upcoming bookings", func(t *testing.T) {
		userId := uuid.NewString()
		ctx := context.WithValue(context.Background(), middleware.UserIdKey, userId)

		mockBookingService.EXPECT().GetUpcomingBookings(gomock.Any(), gomock.Any()).Return([]models.Bookings{}, nil)

		req := httptest.NewRequest(http.MethodGet, "/bookings?type=upcoming", nil).WithContext(ctx)
		rr := httptest.NewRecorder()
		handler.GetUserBookingsHandler(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var responseBody jsonResponse
		err := json.NewDecoder(rr.Body).Decode(&responseBody)
		assert.NoError(t, err)
		assert.Equal(t, "Success", responseBody.Message)
	})

	t.Run("success fetching pending-results bookings", func(t *testing.T) {
		userId := uuid.NewString()
		ctx := context.WithValue(context.Background(), middleware.UserIdKey, userId)

		mockBookingService.EXPECT().GetBookingsToUpdateResult(gomock.Any(), gomock.Any()).Return([]models.Bookings{}, nil)

		req := httptest.NewRequest(http.MethodGet, "/bookings?type=pending-results", nil).WithContext(ctx)
		rr := httptest.NewRecorder()
		handler.GetUserBookingsHandler(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var responseBody jsonResponse
		err := json.NewDecoder(rr.Body).Decode(&responseBody)
		assert.NoError(t, err)
		assert.Equal(t, "Success", responseBody.Message)
	})

	t.Run("failure due to invalid type parameter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/bookings?type=invalid", nil)
		rr := httptest.NewRecorder()
		handler.GetUserBookingsHandler(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		var responseBody jsonResponse
		err := json.NewDecoder(rr.Body).Decode(&responseBody)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid type parameter in url", responseBody.Message)
	})

	t.Run("failure when user not found", func(t *testing.T) {
		userId := uuid.NewString()
		ctx := context.WithValue(context.Background(), middleware.UserIdKey, userId)

		mockBookingService.EXPECT().GetUpcomingBookings(gomock.Any(), gomock.Any()).Return(nil, errors.New("user not found"))

		req := httptest.NewRequest(http.MethodGet, "/bookings?type=upcoming", nil).WithContext(ctx)
		rr := httptest.NewRecorder()
		handler.GetUserBookingsHandler(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		var responseBody jsonResponse
		err := json.NewDecoder(rr.Body).Decode(&responseBody)
		assert.NoError(t, err)
		assert.Equal(t, "Couldn't get upcoming bookings", responseBody.Message)
	})
}
