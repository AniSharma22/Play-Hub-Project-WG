package routes

import (
	"github.com/gorilla/mux"
	"net/http"
	"project2/internal/api/handlers"
	"project2/internal/api/middleware"
)

func InitialiseGameRouter(r *mux.Router, gameHandler *handlers.GameHandler) {
	gameRouter := r.PathPrefix("/games").Subrouter()
	gameRouter.Use(middleware.JwtAuthMiddleware)

	gameRouter.HandleFunc("", gameHandler.GetAllGamesHandler).Methods(http.MethodGet)
	gameRouter.HandleFunc("", middleware.AdminMiddleware(gameHandler.CreateGameHandler)).Methods(http.MethodPost)
	gameRouter.HandleFunc("/{id}", gameHandler.GetGameByIdHandler).Methods(http.MethodGet)
	gameRouter.HandleFunc("/{id}", middleware.AdminMiddleware(gameHandler.UpdateGameStatusHandler)).Methods(http.MethodPut)
	gameRouter.HandleFunc("/{id}", middleware.AdminMiddleware(gameHandler.DeleteGameHandler)).Methods(http.MethodDelete)
}
