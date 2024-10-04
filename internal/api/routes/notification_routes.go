package routes

import (
	"github.com/gorilla/mux"
	"project2/internal/api/handlers"
	"project2/internal/api/middleware"
)

func InitialiseNotificationRouter(r *mux.Router, notificationHandler *handlers.NotificationHandler) {

	notificationRouter := r.PathPrefix("/notifications").Subrouter()
	notificationRouter.Use(middleware.JwtAuthMiddleware)

	notificationRouter.HandleFunc("", notificationHandler.GetNotificationsHandler)
}
