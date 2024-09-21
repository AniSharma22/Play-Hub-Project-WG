package handlers

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"net/http"
	"project2/internal/api/middleware"
	service_interfaces "project2/internal/domain/interfaces/service"
	"project2/pkg/errs"
	"project2/pkg/logger"
	"time"
)

type UserHandler struct {
	userService service_interfaces.UserService
}

func NewUserHandler(userService service_interfaces.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (u *UserHandler) GetUserProfileHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIdStr := vars["id"]

	var userId uuid.UUID
	var err error

	// If the userIdStr is "me", fetch the authenticated user's profile
	if userIdStr == "me" {
		userIdStr, ok := r.Context().Value(middleware.UserIdKey).(string)
		if !ok {
			logger.Logger.Errorw("User ID not found in context", "method", r.Method, "time", time.Now())
			errs.NewUnexpectedError("Could not find the userId").ToJSON(w)
			return
		}

		userId, err = uuid.Parse(userIdStr)
		if err != nil {
			logger.Logger.Errorw("Error parsing user ID", "userID", userIdStr, "error", err, "method", r.Method, "time", time.Now())
			errs.NewInternalServerError("Couldn't parse user id").ToJSON(w)
			return
		}
	} else {
		// Parsing the provided userID from the URL
		userId, err = uuid.Parse(userIdStr)
		if err != nil {
			logger.Logger.Errorw("Error parsing user ID", "userID", userIdStr, "error", err, "method", r.Method, "time", time.Now())
			errs.NewInternalServerError("Couldn't parse user id").ToJSON(w)
			return
		}
	}

	user, err := u.userService.GetUserByID(r.Context(), userId)
	if err != nil {
		logger.Logger.Errorw("Error fetching user", "userID", userId.String(), "error", err, "method", r.Method, "time", time.Now())
		errs.NewInternalServerError("Couldn't get user").ToJSON(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jsonResponse := map[string]any{
		"code":    http.StatusOK,
		"message": "Success",
		"user":    user,
	}
	if err := json.NewEncoder(w).Encode(jsonResponse); err != nil {
		logger.Logger.Errorw("Error encoding response", "userID", userId.String(), "error", err, "method", r.Method, "time", time.Now())
		errs.NewInternalServerError("Error occurred while encoding the response to json").ToJSON(w)
		return
	}

	logger.Logger.Infow("User returned", "userID", userId.String(), "method", r.Method, "time", time.Now())
}

func (u *UserHandler) GetAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	users, err := u.userService.GetAllUsers(r.Context())
	if err != nil {
		logger.Logger.Errorw("Error encoding response", "error", err, "method", r.Method, "time", time.Now())
		errs.NewInternalServerError("Couldn't get users").ToJSON(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jsonResponse := map[string]any{
		"code":    http.StatusOK,
		"message": "Success",
		"users":   users,
	}
	if err := json.NewEncoder(w).Encode(jsonResponse); err != nil {
		logger.Logger.Errorw("Error encoding response", "error", err, "method", r.Method, "time", time.Now())
		errs.NewInternalServerError("Error occurred while encoding the response to json").ToJSON(w)
		return
	}

	logger.Logger.Infow("List of All Users returned", "method", r.Method, "time", time.Now())
}
