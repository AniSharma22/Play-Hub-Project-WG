package handlers

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"net/http"
	"project2/internal/api/middleware"
	service_interfaces "project2/internal/domain/interfaces/service"
	"project2/internal/models"
	"project2/pkg/errs"
	"project2/pkg/logger"
	"time"
)

type BookingHandler struct {
	bookingService service_interfaces.BookingService
}

func NewBookingHandler(bookingService service_interfaces.BookingService) *BookingHandler {
	return &BookingHandler{
		bookingService: bookingService,
	}
}

func (b *BookingHandler) CreateBookingHandler(w http.ResponseWriter, r *http.Request) {
	// Extract userId from the context
	userIdStr, ok := r.Context().Value(middleware.UserIdKey).(string)
	if !ok {
		errs.NewUnexpectedError("Could not find the userId").ToJSON(w)
		logger.Logger.Errorw("UserId not found in request context", "method", r.Method, "time", time.Now())
		return
	}

	var requestBody struct {
		SlotId string `json:"slot_id" validate:"required"`
		GameId string `json:"game_id" validate:"required"`
	}

	// Decode the request body
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		logger.Logger.Errorw("Error decoding request body", "method", r.Method, "error", err, "time", time.Now())
		return
	}

	// Validate the request body
	err = validate.Struct(requestBody)
	if err != nil {
		errs.NewBadRequestError("Invalid request body").ToJSON(w)
		logger.Logger.Errorw("Validation error", "method", r.Method, "error", err, "request", requestBody, "time", time.Now())
		return
	}

	slotId, err := uuid.Parse(requestBody.SlotId)
	if err != nil {
		errs.NewBadRequestError("Couldn't parse slot ID").ToJSON(w)
		logger.Logger.Errorw("Failed to parse slot ID", "method", r.Method, "slotId", requestBody.SlotId, "error", err, "time", time.Now())
		return
	}
	gameId, err := uuid.Parse(requestBody.GameId)
	if err != nil {
		errs.NewBadRequestError("Couldn't parse game ID").ToJSON(w)
		logger.Logger.Errorw("Failed to parse game ID", "method", r.Method, "gameId", requestBody.GameId, "error", err, "time", time.Now())
		return
	}

	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		errs.NewBadRequestError("Couldn't parse user ID").ToJSON(w)
		logger.Logger.Errorw("Failed to parse user ID", "method", r.Method, "userId", userIdStr, "error", err, "time", time.Now())
		return
	}

	err = b.bookingService.MakeBooking(r.Context(), userId, slotId, gameId)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrDbError):
			errs.NewInternalServerError("Couldn't create booking").ToJSON(w)
			logger.Logger.Errorw("Database error during booking creation", "method", r.Method, "userId", userId, "slotId", slotId, "error", err, "time", time.Now())
		case errors.Is(err, errs.ErrServiceError):
			errs.NewInternalServerError("Couldn't create booking").ToJSON(w)
			logger.Logger.Errorw("Service error during booking creation", "method", r.Method, "userId", userId, "slotId", slotId, "error", err, "time", time.Now())
		case errors.Is(err, errs.ErrSlotPassed):
			errs.NewBadRequestError("Slot timing has already passed").ToJSON(w)
			logger.Logger.Warnw("Slot timing passed", "method", r.Method, "userId", userId, "slotId", slotId, "time", time.Now())
		case errors.Is(err, errs.ErrAlreadyExists):
			errs.NewBadRequestError("Slot already booked").ToJSON(w)
			logger.Logger.Warnw("Slot already booked", "method", r.Method, "userId", userId, "slotId", slotId, "time", time.Now())
		default:
			errs.NewBadRequestError("User has already booked in this slot").ToJSON(w)
			logger.Logger.Warnw("User already booked in slot", "method", r.Method, "userId", userId, "slotId", slotId, "time", time.Now())
		}
		return
	}

	// Log success and return response
	w.Header().Set("Content-Type", "application/json")
	jsonResponse := map[string]any{
		"code":    http.StatusOK,
		"message": "Booking created successfully",
	}
	json.NewEncoder(w).Encode(jsonResponse)
	logger.Logger.Infow("Booking created successfully", "method", r.Method, "userId", userId, "slotId", slotId, "time", time.Now())
}

func (b *BookingHandler) GetUserBookingsHandler(w http.ResponseWriter, r *http.Request) {

	var bookings []models.Bookings

	if condition := r.URL.Query().Get("type"); condition != "" {
		switch condition {
		case "upcoming":
			// Extract userId from the context
			userIdStr, ok := r.Context().Value(middleware.UserIdKey).(string)
			if !ok {
				errs.NewUnexpectedError("Could not find the userId").ToJSON(w)
				logger.Logger.Errorw("UserId not found in request context", "method", r.Method, "time", time.Now())
				return
			}

			userId, err := uuid.Parse(userIdStr)
			if err != nil {
				errs.NewInternalServerError("Couldn't parse user ID").ToJSON(w)
				logger.Logger.Errorw("Failed to parse user ID", "method", r.Method, "userId", userIdStr, "error", err, "time", time.Now())
				return
			}

			bookings, err = b.bookingService.GetUpcomingBookings(r.Context(), userId)
			if err != nil {
				errs.NewInternalServerError("Couldn't get upcoming bookings").ToJSON(w)
				logger.Logger.Errorw("Error fetching upcoming bookings", "method", r.Method, "userId", userId, "error", err, "time", time.Now())
				return
			}
			// Log success and return response
			w.Header().Set("Content-Type", "application/json")
			jsonResponse := map[string]any{
				"code":    http.StatusOK,
				"message": "Success",
				"upcoming_bookings": func() []models.Bookings {
					if bookings == nil {
						return []models.Bookings{}
					}
					return bookings
				}(),
			}
			json.NewEncoder(w).Encode(jsonResponse)
			logger.Logger.Infow("Fetched upcoming bookings successfully", "method", r.Method, "time", time.Now())

		case "pending-results":
			// Extract userId from the context
			userIdStr, ok := r.Context().Value(middleware.UserIdKey).(string)
			if !ok {
				errs.NewUnexpectedError("Could not find the userId").ToJSON(w)
				logger.Logger.Errorw("UserId not found in request context", "method", r.Method, "time", time.Now())
				return
			}

			userId, err := uuid.Parse(userIdStr)
			if err != nil {
				errs.NewInternalServerError("Couldn't parse user ID").ToJSON(w)
				logger.Logger.Errorw("Failed to parse user ID", "method", r.Method, "userId", userIdStr, "error", err, "time", time.Now())
				return
			}

			bookings, err = b.bookingService.GetBookingsToUpdateResult(r.Context(), userId)
			if err != nil {
				errs.NewInternalServerError("Couldn't get pending results").ToJSON(w)
				logger.Logger.Errorw("Error fetching pending results", "method", r.Method, "userId", userId, "error", err, "time", time.Now())
				return
			}
			// Log success and return response
			w.Header().Set("Content-Type", "application/json")
			jsonResponse := map[string]any{
				"code":    http.StatusOK,
				"message": "Success",
				"pending_results": func() []models.Bookings {
					if bookings == nil {
						return []models.Bookings{}
					}
					return bookings
				}(),
			}
			json.NewEncoder(w).Encode(jsonResponse)
			logger.Logger.Infow("Fetched upcoming bookings successfully", "method", r.Method, "time", time.Now())
		default:
			errs.NewBadRequestError("Invalid type parameter in url").ToJSON(w)
			logger.Logger.Errorw("Error fetching pending results", "method", r.Method, "type", condition, "time", time.Now())
			return
		}
	}
}

//func (b *BookingHandler) GetPendingResultsHandler(w http.ResponseWriter, r *http.Request) {
//	// Extract userId from the context
//	userIdStr, ok := r.Context().Value(middleware.UserIdKey).(string)
//	if !ok {
//		errs.NewUnexpectedError("Could not find the userId").ToJSON(w)
//		logger.Logger.Errorw("UserId not found in request context", "method", r.Method, "time", time.Now())
//		return
//	}
//
//	userId, err := uuid.Parse(userIdStr)
//	if err != nil {
//		errs.NewInternalServerError("Couldn't parse user ID").ToJSON(w)
//		logger.Logger.Errorw("Failed to parse user ID", "method", r.Method, "userId", userIdStr, "error", err, "time", time.Now())
//		return
//	}
//
//	bookings, err := b.bookingService.GetBookingsToUpdateResult(r.Context(), userId)
//	if err != nil {
//		errs.NewInternalServerError("Couldn't get pending results").ToJSON(w)
//		logger.Logger.Errorw("Error fetching pending results", "method", r.Method, "userId", userId, "error", err, "time", time.Now())
//		return
//	}
//
//	// Log success and return response
//	w.Header().Set("Content-Type", "application/json")
//	jsonResponse := map[string]any{
//		"code":    http.StatusOK,
//		"message": "Success",
//		"pending_results": func() []models.Bookings {
//			if bookings == nil {
//				return []models.Bookings{}
//			}
//			return bookings
//		}(),
//	}
//	json.NewEncoder(w).Encode(jsonResponse)
//	logger.Logger.Infow("Fetched pending results successfully", "method", r.Method, "userId", userId, "time", time.Now())
//}
