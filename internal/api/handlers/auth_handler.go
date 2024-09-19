package handlers

import (
	"encoding/json"
	"errors"
	"github.com/go-playground/validator"
	"net/http"
	"project2/internal/domain/entities"
	service_interfaces "project2/internal/domain/interfaces/service"
	"project2/pkg/errs"
	"project2/pkg/logger"
	"project2/pkg/utils"
	"project2/pkg/validation"
	"time"
)

var validate *validator.Validate

type AuthHandler struct {
	authService service_interfaces.AuthService
}

func init() {
	// Initialise a new validator
	validate = validator.New()

	// Register custom validation functions
	validate.RegisterValidation("isValidEmail", validation.IsValidEmail)
	validate.RegisterValidation("isValidPassword", validation.IsValidPassword)
	validate.RegisterValidation("isValidPhoneNo", validation.IsValidPhoneNumber)
	validate.RegisterValidation("isValidGender", validation.IsValidGender)
}

func NewAuthHandler(authService service_interfaces.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (a *AuthHandler) SignupHandler(w http.ResponseWriter, r *http.Request) {
	// define the request body
	var requestBody struct {
		Email    string `json:"email" validate:"required,isValidEmail"`
		Password string `json:"password" validate:"required,isValidPassword"`
		PhoneNo  string `json:"phone_no" validate:"required,isValidPhoneNo"`
		Gender   string `json:"gender" validate:"required,isValidGender"`
	}

	// decode the request body
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		logger.Logger.Errorw("Error decoding request body", "method", r.Method, "error", err, "time", time.Now())
		return
	}

	// Validate the request body
	err = validate.Struct(requestBody)
	if err != nil {
		errs.NewInvalidBodyError("Invalid request body").ToJSON(w)
		logger.Logger.Errorw("Validation error", "method", r.Method, "error", err, "request", requestBody, "time", time.Now())
		return
	}

	userBody := &entities.User{
		Email:        requestBody.Email,
		Password:     requestBody.Password,
		Gender:       requestBody.Gender,
		MobileNumber: requestBody.PhoneNo,
	}

	// call the signup service
	user, err := a.authService.Signup(r.Context(), userBody)
	if err != nil {
		if errors.Is(err, errs.ErrEmailExists) {
			errs.NewConflictError("Email already exists").ToJSON(w)
			logger.Logger.Errorw("Email Already Exists : Conflict Error", "method", r.Method, "error", err, "request", requestBody, "time", time.Now())
			return
		}
		errs.NewInternalServerError("Internal Server Error").ToJSON(w)
		return
	}

	// create a jwt token
	token, err := utils.CreateJwtToken(user.UserID, user.Role)
	if err != nil {
		errs.NewInternalServerError("Failed to generate token").ToJSON(w)
		return
	}

	// Return the token as a JSON response
	w.Header().Set("Content-Type", "application/json")
	jsonResponse := map[string]any{
		"code":  http.StatusOK,
		"token": token,
		"role":  user.Role,
	}
	json.NewEncoder(w).Encode(jsonResponse)
	logger.Logger.Infow("Signup Successful", "method", r.Method, "request", requestBody, "time", time.Now())
}

func (a *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {

	// Define the request body structure;;
	var requestBody struct {
		Email    string `json:"email" validate:"required"`
		Password string `json:"password" validate:"required"`
	}

	// Decode the request body
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		logger.Logger.Errorw("Error decoding request body", "method", r.Method, "error", err, "time", time.Now())
		return
	}

	// Validate the request body
	err = validate.Struct(requestBody)
	if err != nil {
		errs.NewInvalidBodyError("Invalid request body").ToJSON(w)
		logger.Logger.Errorw("Validation error", "method", r.Method, "error", err, "request", requestBody, "time", time.Now())
		return
	}

	// call the login service
	user, err := a.authService.Login(r.Context(), requestBody.Email, []byte(requestBody.Password))
	if err != nil {
		// Check if the error is an "invalid password" error
		if errors.Is(err, errs.ErrInvalidPassword) {
			errs.NewUnauthorizedError("Invalid username or password").ToJSON(w)
			return
		}

		// Check if the error is a "user not found" error
		if errors.Is(err, errs.ErrUserNotFound) {
			errs.NewNotFoundError("No such user exists").ToJSON(w)
			return
		}

		// Only option is Internal api error
		errs.NewInternalServerError("Internal Server Error").ToJSON(w)
		return
	}

	// create a jwt token
	token, err := utils.CreateJwtToken(user.UserID, user.Role)
	if err != nil {
		errs.NewInternalServerError("Failed to generate token").ToJSON(w)
		return
	}

	// Return the token as a JSON response
	w.Header().Set("Content-Type", "application/json")
	jsonResponse := map[string]any{
		"code":  http.StatusOK,
		"token": token,
		"role":  user.Role,
	}
	json.NewEncoder(w).Encode(jsonResponse)
	logger.Logger.Infow("Login Successful", "method", r.Method, "request", requestBody, "time", time.Now())
}

func (a *AuthHandler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jsonResponse := map[string]any{
		"code":    http.StatusOK,
		"message": "Logout Successful",
	}
	json.NewEncoder(w).Encode(jsonResponse)
	logger.Logger.Infow("Logout Successful", "method", r.Method, "time", time.Now())
}
