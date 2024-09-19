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
		return
	}

	var requestBody struct {
		SlotId string `json:"slot_id" validate:"required"`
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
	slotId, err := uuid.Parse(requestBody.SlotId)
	if err != nil {
		errs.NewInternalServerError("Couldn't parse slot id").ToJSON(w)
		return
	}
	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		errs.NewInternalServerError("Couldn't parse user id").ToJSON(w)
		return
	}

	err = b.bookingService.MakeBooking(r.Context(), userId, slotId)
	if err != nil {
		if errors.Is(err, errs.ErrDbError) {
			// error due to some repo call
			errs.NewInternalServerError("Couldn't create booking").ToJSON(w)
			return
		} else if errors.Is(err, errs.ErrServiceError) {
			// error due to some service call
			errs.NewInternalServerError("Couldn't create booking").ToJSON(w)
			return
		} else if errors.Is(err, errs.ErrSlotPassed) {
			// slot timing has passed
			errs.NewBadRequestError("Slot timing has already passed.").ToJSON(w)
			return
		} else if errors.Is(err, errs.ErrAlreadyExists) {
			// slot if already fully booked
			errs.NewBadRequestError("Slot already booked").ToJSON(w)
			return
		} else {
			// user has already booked in this slot
			errs.NewBadRequestError("User has already booked in this slot").ToJSON(w)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	jsonResponse := map[string]any{
		"code":    http.StatusOK,
		"message": "Success",
	}
	json.NewEncoder(w).Encode(jsonResponse)
	logger.Logger.Infow("Signup Successful", "method", r.Method, "request", requestBody, "time", time.Now())
}

func (b *BookingHandler) GetUpcomingBookingsHandler(w http.ResponseWriter, r *http.Request) {
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

	bookings, err := b.bookingService.GetUpcomingBookings(r.Context(), userId)
	if err != nil {
		errs.NewInternalServerError("Couldn't get upcoming bookings").ToJSON(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jsonResponse := map[string]any{
		"code":    http.StatusOK,
		"message": "Success",
		"bookings": func() []models.Bookings {
			if bookings == nil {
				return []models.Bookings{}
			}
			return bookings
		}(),
	}
	json.NewEncoder(w).Encode(jsonResponse)
	logger.Logger.Infow("Upcoming Bookings sent successfully", "method", r.Method, "time", time.Now())
}

func (b *BookingHandler) GetPendingResultsHandler(w http.ResponseWriter, r *http.Request) {
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

	bookings, err := b.bookingService.GetBookingsToUpdateResult(r.Context(), userId)
	if err != nil {
		errs.NewInternalServerError("Couldn't get upcoming bookings").ToJSON(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jsonResponse := map[string]any{
		"code":    http.StatusOK,
		"message": "Success",
		"bookings": func() []models.Bookings {
			if bookings == nil {
				return []models.Bookings{}
			}
			return bookings
		}(),
	}
	json.NewEncoder(w).Encode(jsonResponse)
	logger.Logger.Infow("Upcoming Bookings sent successfully", "method", r.Method, "time", time.Now())
}
