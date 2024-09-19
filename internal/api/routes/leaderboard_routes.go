package routes

import (
	"github.com/gorilla/mux"
	"net/http"
	"project2/internal/api/handlers"
	"project2/internal/api/middleware"
)

func InitialiseLeaderboardRouter(r *mux.Router, leaderboardHandler *handlers.LeaderboardHandler) {
	leaderboardRouter := r.PathPrefix("/leaderboards").Subrouter()
	leaderboardRouter.Use(middleware.JwtAuthMiddleware)

	leaderboardRouter.HandleFunc("/games/{gameID}", leaderboardHandler.GetGameLeaderboardHandler)
	leaderboardRouter.HandleFunc("/add-win", leaderboardHandler.AddWinToUserHandler).Methods(http.MethodPost)
	leaderboardRouter.HandleFunc("/add-loss", leaderboardHandler.AddLossToUserHandler).Methods(http.MethodPost)
}
