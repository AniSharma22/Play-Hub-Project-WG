package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"net/http"
	"project2/internal/api/middleware"
	"project2/internal/domain/entities"
	service_interfaces "project2/internal/domain/interfaces/service"
	"project2/pkg/errs"
	"project2/pkg/logger"
	"project2/pkg/utils"
	"strings"
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
			errs.InvalidRequestError("Could not find the userId").ToJson2(w)
			return
		}

		userId, err = uuid.Parse(userIdStr)
		if err != nil {
			logger.Logger.Errorw("Error parsing user ID", "userID", userIdStr, "error", err, "method", r.Method, "time", time.Now())
			errs.ValidationError("Couldn't parse user id").ToJson2(w)
			return
		}
	} else {
		// Parsing the provided userID from the URL
		userId, err = uuid.Parse(userIdStr)
		if err != nil {
			logger.Logger.Errorw("Error parsing user ID", "userID", userIdStr, "error", err, "method", r.Method, "time", time.Now())
			errs.ValidationError("Couldn't parse user id").ToJson2(w)
			return
		}
	}

	user, err := u.userService.GetUserByID(r.Context(), userId)
	if err != nil {
		logger.Logger.Errorw("Error fetching user", "userID", userId.String(), "error", err, "method", r.Method, "time", time.Now())
		errs.DBError("Couldn't get user").ToJson2(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jsonResponse := map[string]any{
		"code":    http.StatusOK,
		"message": "Success",
		"user":    user,
	}

	if err = utils.JsonEncoder(w, jsonResponse); err != nil {
		return
	}
	logger.Logger.Infow("User returned", "userID", userId.String(), "method", r.Method, "time", time.Now())
}

//func (u *UserHandler) GetAllUsersHandler(w http.ResponseWriter, r *http.Request) {
//	users, err := u.userService.GetAllUsers(r.Context())
//	if err != nil {
//		logger.Logger.Errorw("Error encoding response", "error", err, "method", r.Method, "time", time.Now())
//		errs.DBError("Couldn't get users").ToJson2(w)
//		return
//	}
//
//	w.Header().Set("Content-Type", "application/json")
//	jsonResponse := map[string]any{
//		"code":    http.StatusOK,
//		"message": "Success",
//		"users":   users,
//	}
//	if err = utils.JsonEncoder(w, jsonResponse); err != nil {
//		return
//	}
//	logger.Logger.Infow("List of All Users returned", "method", r.Method, "time", time.Now())
//}

func (u *UserHandler) GetAllUsersHandlerPaginated(w http.ResponseWriter, r *http.Request) {
	limit := r.URL.Query().Get("limit")
	offset := r.URL.Query().Get("offset")
	substring := r.URL.Query().Get("substring")

	users, total, err := u.userService.GetAllUsers(r.Context(), limit, offset, strings.ToLower(substring))
	if err != nil {
		logger.Logger.Errorw("Error encoding response", "error", err, "method", r.Method, "time", time.Now())
		errs.DBError("Couldn't get users").ToJson2(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jsonResponse := map[string]any{
		"code":    http.StatusOK,
		"message": "Success",
		"users":   users,
		"total":   total,
	}
	if err = utils.JsonEncoder(w, jsonResponse); err != nil {
		return
	}
	logger.Logger.Infow("List of All Users returned", "method", r.Method, "time", time.Now())
}

func (u *UserHandler) GetAllUsersPublicHandler(w http.ResponseWriter, r *http.Request) {

	slotIdStr := r.URL.Query().Get("slotId")
	slotId, err := uuid.Parse(slotIdStr)
	if err != nil {
		logger.Logger.Errorw("Error parsing slot ID", "userID", slotIdStr, "error", err, "method", r.Method, "time", time.Now())
		errs.ValidationError("Couldn't parse slot id").ToJson2(w)
		return
	}

	userIdStr, _ := r.Context().Value(middleware.UserIdKey).(string)
	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		logger.Logger.Errorw("Error parsing user ID", "userID", userIdStr, "error", err, "method", r.Method, "time", time.Now())
		errs.ValidationError("Couldn't parse user id").ToJson2(w)
		return
	}

	fmt.Println(slotId, userId)
	users, err := u.userService.GetAllUsersPublic(r.Context(), userId, slotId)
	if err != nil {
		logger.Logger.Errorw("Error encoding response", "error", err, "method", r.Method, "time", time.Now())
		errs.DBError("Couldn't get users").ToJson2(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jsonResponse := map[string]any{
		"code":    http.StatusOK,
		"message": "Success",
		"users":   users,
	}
	if err = utils.JsonEncoder(w, jsonResponse); err != nil {
		return
	}
	logger.Logger.Infow("List of All Users returned", "method", r.Method, "time", time.Now())
}

func (u *UserHandler) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIdStr := vars["id"]
	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		logger.Logger.Errorw("Error parsing user ID", "userID", userIdStr, "error", err, "method", r.Method, "time", time.Now())
		errs.ValidationError("Couldn't parse user id").ToJson2(w)
		return
	}

	err = u.userService.DeleteUser(r.Context(), userId)
	if err != nil {
		logger.Logger.Errorw("Error deleting user", "userID", userId.String(), "error", err, "method", r.Method, "time", time.Now())
		errs.InvalidRequestError("Operation cannot be performed").ToJson2(w)
		return
	}

	jsonResponse := map[string]any{
		"code":    http.StatusOK,
		"message": "Success",
	}
	if err = utils.JsonEncoder(w, jsonResponse); err != nil {
		return
	}
	logger.Logger.Infow("User deleted", "method", r.Method, "userId", userId, "time", time.Now())
}

func (u *UserHandler) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	userIdStr, _ := r.Context().Value(middleware.UserIdKey).(string)
	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		logger.Logger.Errorw("Error parsing user ID", "userID", userIdStr, "error", err, "method", r.Method, "time", time.Now())
		errs.ValidationError("Couldn't parse user id").ToJson2(w)
		return
	}

	var requestBody struct {
		UserName     string `json:"username" validate:"required"`
		Password     string `json:"password"`
		MobileNumber string `json:"mobile_number" validate:"required"`
		ImageUrl     string `json:"image_url" validate:"required"`
	}

	// Decode the request body
	err = json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		errs.InvalidRequestError("Invalid or malformed request body").ToJson2(w)
		logger.Logger.Errorw("Error decoding request body", "method", r.Method, "error", err, "body", r.Body, "time", time.Now())
		return
	}

	// Validate the request body
	err = validate.Struct(requestBody)
	if err != nil {
		errs.ValidationError("Invalid request body").ToJson2(w)
		logger.Logger.Errorw("Validation error", "method", r.Method, "error", err, "request", requestBody, "time", time.Now())
		return
	}
	// create a user object
	user := &entities.User{
		UserID:       userId,
		Username:     strings.ToLower(requestBody.UserName),
		Password:     requestBody.Password,
		MobileNumber: requestBody.MobileNumber,
		ImageUrl:     requestBody.ImageUrl,
	}

	err = u.userService.UpdateUserDetails(r.Context(), user)
	if err != nil {
		errs.DBError("Couldn't update user details").ToJson2(w)
		logger.Logger.Errorw("Couldn't update user details", "error", err, "method", r.Method, "time", time.Now())
		return
	}

	jsonResponse := map[string]any{
		"code":    http.StatusOK,
		"message": "Success",
	}
	if err = utils.JsonEncoder(w, jsonResponse); err != nil {
		return
	}

	logger.Logger.Infow("User updated successfully", "method", r.Method, "user", user, "time", time.Now())
}
