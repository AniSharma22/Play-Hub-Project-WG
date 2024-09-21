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
	startTime := time.Now()
	logger.Logger.Infow("Signup request received", "method", r.Method, "time", startTime)

	// Define the request body
	var requestBody struct {
		Email    string `json:"email" validate:"required,isValidEmail"`
		Password string `json:"password" validate:"required,isValidPassword"`
		PhoneNo  string `json:"phone_no" validate:"required,isValidPhoneNo"`
		Gender   string `json:"gender" validate:"required,isValidGender"`
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
		errs.NewBadRequestError("Invalid request body").ToJSON(w)
		logger.Logger.Errorw("Validation error", "method", r.Method, "error", err, "request", requestBody, "time", time.Now())
		return
	}

	userBody := &entities.User{
		Email:        requestBody.Email,
		Password:     requestBody.Password,
		Gender:       requestBody.Gender,
		MobileNumber: requestBody.PhoneNo,
	}

	// Call the signup service
	user, err := a.authService.Signup(r.Context(), userBody)
	if err != nil {
		if errors.Is(err, errs.ErrEmailExists) {
			errs.NewConflictError("Email already exists").ToJSON(w)
			logger.Logger.Warnw("Email conflict during signup", "method", r.Method, "request", requestBody, "time", time.Now())
			return
		}
		errs.NewInternalServerError("Internal Server Error").ToJSON(w)
		logger.Logger.Errorw("Internal server error during signup", "method", r.Method, "error", err, "request", requestBody, "time", time.Now())
		return
	}

	// Create a JWT token
	token, err := utils.CreateJwtToken(user.UserID, user.Role)
	if err != nil {
		errs.NewInternalServerError("Failed to generate token").ToJSON(w)
		logger.Logger.Errorw("Failed to generate token", "method", r.Method, "error", err, "time", time.Now())
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

	logger.Logger.Infow("Signup successful", "method", r.Method, "email", requestBody.Email, "role", user.Role, "time", time.Now(), "duration", time.Since(startTime))
}

func (a *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	logger.Logger.Infow("Login request received", "method", r.Method, "time", startTime)

	// Define the request body structure
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
		errs.NewBadRequestError("Invalid request body").ToJSON(w)
		logger.Logger.Errorw("Validation error", "method", r.Method, "error", err, "request", requestBody, "time", time.Now())
		return
	}

	// Call the login service
	user, err := a.authService.Login(r.Context(), requestBody.Email, []byte(requestBody.Password))
	if err != nil {
		if errors.Is(err, errs.ErrInvalidPassword) {
			errs.NewUnauthorizedError("Invalid username or password").ToJSON(w)
			logger.Logger.Warnw("Invalid password", "method", r.Method, "email", requestBody.Email, "time", time.Now())
			return
		}

		if errors.Is(err, errs.ErrUserNotFound) {
			errs.NewNotFoundError("No such user exists").ToJSON(w)
			logger.Logger.Warnw("User not found", "method", r.Method, "email", requestBody.Email, "time", time.Now())
			return
		}

		errs.NewInternalServerError("Internal Server Error").ToJSON(w)
		logger.Logger.Errorw("Internal server error during login", "method", r.Method, "error", err, "request", requestBody, "time", time.Now())
		return
	}

	// Create a JWT token
	token, err := utils.CreateJwtToken(user.UserID, user.Role)
	if err != nil {
		errs.NewInternalServerError("Failed to generate token").ToJSON(w)
		logger.Logger.Errorw("Failed to generate token", "method", r.Method, "error", err, "time", time.Now())
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

	logger.Logger.Infow("Login successful", "method", r.Method, "email", requestBody.Email, "role", user.Role, "time", time.Now(), "duration", time.Since(startTime))
}

func (a *AuthHandler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	logger.Logger.Infow("Logout request received", "method", r.Method, "time", startTime)

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	jsonResponse := map[string]any{
		"code":    http.StatusOK,
		"message": "Logout Successful",
	}
	json.NewEncoder(w).Encode(jsonResponse)

	logger.Logger.Infow("Logout successful", "method", r.Method, "time", time.Now(), "duration", time.Since(startTime))
}
