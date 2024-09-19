package routes

import (
	"github.com/gorilla/mux"
	"project2/internal/api/handlers"
	"project2/internal/api/middleware"
)

// InitialiseUserRouter initializes a sub-router and routes to handle requests coming to "/user" endpoint
func InitialiseUserRouter(r *mux.Router, userHandler *handlers.UserHandler) {

	// Create a user sub-router to handle user requests
	userRouter := r.PathPrefix("/users").Subrouter()
	userRouter.Use(middleware.JwtAuthMiddleware)

	// Routes to handle specific user requests
	userRouter.HandleFunc("/me", userHandler.GetUserProfileHandler)
	userRouter.HandleFunc("", userHandler.GetAllUsersHandler) // no use as of now
	userRouter.HandleFunc("/{userID}", userHandler.GetUserByIdHandler)
}
