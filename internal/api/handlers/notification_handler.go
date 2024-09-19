package handlers

import (
	"encoding/json"
	"github.com/google/uuid"
	"net/http"
	"project2/internal/api/middleware"
	"project2/internal/domain/entities"
	service_interfaces "project2/internal/domain/interfaces/service"
	"project2/pkg/errs"
	"project2/pkg/logger"
	"time"
)

type NotificationHandler struct {
	notificationService service_interfaces.NotificationService
}

func NewNotificationHandler(notificationService service_interfaces.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

func (n *NotificationHandler) GetNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	// Extract userId from the context
	userIdStr, ok := r.Context().Value(middleware.UserIdKey).(string)
	if !ok {
		errs.NewUnexpectedError("Could not find the userId").ToJSON(w)
		return
	}

	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		errs.NewInternalServerError("Couldn't parse user id").ToJSON(w)
		return
	}

	notifications, err := n.notificationService.GetUserNotifications(r.Context(), userId)
	if err != nil {
		errs.NewInternalServerError("Error fetching notifications").ToJSON(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jsonResponse := map[string]any{
		"code":    http.StatusOK,
		"message": "Success",
		"bookings": func() []entities.Notification {
			if notifications == nil {
				return []entities.Notification{}
			}
			return notifications
		}(),
	}
	json.NewEncoder(w).Encode(jsonResponse)
	logger.Logger.Infow("User notifications sent successfully", "method", r.Method, "time", time.Now())
}
