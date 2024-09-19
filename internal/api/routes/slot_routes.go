package routes

import (
	"github.com/gorilla/mux"
	"net/http"
	"project2/internal/api/handlers"
	"project2/internal/api/middleware"
)

func InitialiseSlotRouter(r *mux.Router, slotHandler *handlers.SlotHandler) {

	slotRouter := r.PathPrefix("/slots").Subrouter()
	slotRouter.Use(middleware.JwtAuthMiddleware)

	slotRouter.HandleFunc("/games/{gameID}", slotHandler.GetTodaySlotsHandler).Methods(http.MethodGet)
	slotRouter.HandleFunc("/{id}", slotHandler.GetSlotByIdHandler).Methods(http.MethodGet)
}
