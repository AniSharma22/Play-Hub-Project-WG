package routes

import (
	"github.com/gorilla/mux"
	"net/http"
	"project2/internal/api/handlers"
)

func InitialiseAuthRouter(r *mux.Router, authHandler *handlers.AuthHandler) {
	authRouter := r.PathPrefix("/auth").Subrouter()

	authRouter.HandleFunc("/signup", authHandler.SignupHandler).Methods(http.MethodPost)
	authRouter.HandleFunc("/login", authHandler.LoginHandler).Methods(http.MethodPost)
	authRouter.HandleFunc("/logout", authHandler.LogoutHandler).Methods(http.MethodPost)
}
