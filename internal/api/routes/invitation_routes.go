package routes

import (
	"github.com/gorilla/mux"
	"net/http"
	"project2/internal/api/handlers"
	"project2/internal/api/middleware"
)

func InitialiseInvitationRouter(r *mux.Router, invitationHandler *handlers.InvitationHandler) {
	invitationRouter := r.PathPrefix("/invitations").Subrouter()
	invitationRouter.Use(middleware.JwtAuthMiddleware)

	invitationRouter.HandleFunc("", invitationHandler.CreateInvitationHandler)
	invitationRouter.HandleFunc("/{id}", invitationHandler.UpdateInvitationStatusHandler)
	invitationRouter.HandleFunc("", invitationHandler.GetPendingInvitationHandler).Methods(http.MethodGet)
}
