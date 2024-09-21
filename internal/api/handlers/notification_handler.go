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
		logger.Logger.Errorw("Error finding userId in context", "method", r.Method, "time", time.Now())
		errs.NewUnexpectedError("Could not find the userId").ToJSON(w)
		return
	}

	logger.Logger.Infow("Processing request to get notifications", "userID", userIdStr, "method", r.Method, "time", time.Now())
	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		logger.Logger.Errorw("Error parsing user ID", "userID", userIdStr, "error", err, "time", time.Now())
		errs.NewInternalServerError("Couldn't parse user id").ToJSON(w)
		return
	}

	notifications, err := n.notificationService.GetUserNotifications(r.Context(), userId)
	if err != nil {
		logger.Logger.Errorw("Error fetching notifications", "userID", userIdStr, "error", err, "time", time.Now())
		errs.NewInternalServerError("Error fetching notifications").ToJSON(w)
		return
	}

	// Respond with notifications
	w.Header().Set("Content-Type", "application/json")
	jsonResponse := map[string]any{
		"code":    http.StatusOK,
		"message": "Success",
		"notifications": func() []entities.Notification {
			if notifications == nil {
				return []entities.Notification{}
			}
			return notifications
		}(),
	}
	if err := json.NewEncoder(w).Encode(jsonResponse); err != nil {
		logger.Logger.Errorw("Error encoding response", "method", r.Method, "error", err, "time", time.Now())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logger.Logger.Infow("User notifications sent successfully", "userID", userIdStr, "method", r.Method, "time", time.Now())
}
