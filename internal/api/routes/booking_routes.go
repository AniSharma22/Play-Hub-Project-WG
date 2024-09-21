package routes

import (
	"github.com/gorilla/mux"
	"net/http"
	"project2/internal/api/handlers"
	"project2/internal/api/middleware"
)

func InitialiseBookingRouter(r *mux.Router, bookingHandler *handlers.BookingHandler) {
	bookingRouter := r.PathPrefix("/bookings").Subrouter()
	bookingRouter.Use(middleware.JwtAuthMiddleware)

	bookingRouter.HandleFunc("", bookingHandler.CreateBookingHandler).Methods(http.MethodPost)
	bookingRouter.HandleFunc("", bookingHandler.GetUserBookingsHandler).Methods(http.MethodGet)
	//bookingRouter.HandleFunc("/pending-results", bookingHandler.GetPendingResultsHandler).Methods(http.MethodGet)
}
