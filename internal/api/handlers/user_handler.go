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

	user, err := u.userService.GetUserByID(r.Context(), userId)
	if err != nil {
		errs.NewInternalServerError("Couldn't get user").ToJSON(w)
	}

	w.Header().Set("Content-Type", "application/json")
	jsonResponse := map[string]any{
		"code":    http.StatusOK,
		"message": "Success",
		"data":    user,
	}
	json.NewEncoder(w).Encode(jsonResponse)
	logger.Logger.Infow("User profile returned", "method", r.Method, "time", time.Now())

}

func (u *UserHandler) GetAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	// No such service
}

func (u *UserHandler) GetUserByIdHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIdStr := vars["userID"]

	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		errs.NewInternalServerError("Couldn't parse user id").ToJSON(w)
		return
	}

	user, err := u.userService.GetUserByID(r.Context(), userId)
	if err != nil {
		errs.NewInternalServerError("Couldn't get user").ToJSON(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jsonResponse := map[string]any{
		"code":    http.StatusOK,
		"message": "Success",
		"data":    user,
	}
	json.NewEncoder(w).Encode(jsonResponse)
	logger.Logger.Infow("User returned", "method", r.Method, "time", time.Now())
}
