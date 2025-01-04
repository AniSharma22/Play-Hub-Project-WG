package routes

import (
	"github.com/gorilla/mux"
	"net/http"
	"project2/internal/api/handlers"
	"project2/internal/api/middleware"
)

// InitialiseUserRouter initializes a sub-router and routes to handle requests coming to "/user" endpoint
func InitialiseUserRouter(r *mux.Router, userHandler *handlers.UserHandler) {

	// Create a user sub-router to handle user requests
	userRouter := r.PathPrefix("/users").Subrouter()
	userRouter.Use(middleware.JwtAuthMiddleware)

	// Routes to handle specific user requests
	userRouter.HandleFunc("", middleware.AdminMiddleware(userHandler.GetAllUsersHandlerPaginated)).Methods(http.MethodGet)
	userRouter.HandleFunc("/public", userHandler.GetAllUsersPublicHandler).Methods(http.MethodGet)
	userRouter.HandleFunc("/{id}", userHandler.GetUserProfileHandler).Methods(http.MethodGet)
	userRouter.HandleFunc("/{id}", middleware.AdminMiddleware(userHandler.DeleteUserHandler)).Methods(http.MethodDelete)
	userRouter.HandleFunc("", userHandler.UpdateUserHandler).Methods(http.MethodPut)
}
