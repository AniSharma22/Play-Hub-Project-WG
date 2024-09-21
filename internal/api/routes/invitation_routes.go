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

	invitationRouter.HandleFunc("", invitationHandler.CreateInvitationHandler).Methods(http.MethodPost)
	invitationRouter.HandleFunc("/{id}", invitationHandler.UpdateInvitationStatusHandler).Methods(http.MethodPatch)
	invitationRouter.HandleFunc("", invitationHandler.GetPendingInvitationHandler).Methods(http.MethodGet)
}
