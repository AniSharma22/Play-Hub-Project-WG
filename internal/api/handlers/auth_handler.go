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
	logger.Logger.Infow("Signup request received", "method", r.Method)

	var requestBody struct {
		Email    string `json:"email" validate:"required,isValidEmail"`
		Password string `json:"password" validate:"required,isValidPassword"`
		PhoneNo  string `json:"phone_no" validate:"required,isValidPhoneNo"`
		Gender   string `json:"gender" validate:"required,isValidGender"`
	}

	// Decode the request body allowing no extra fields
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&requestBody)
	if err != nil {
		logger.Logger.Errorw("Error decoding request body", "method", r.Method, "error", err)
		errs.InvalidRequestError("Invalid or malformed request body").ToJson2(w)
		return
	}

	// Validate the request body
	err = validate.Struct(requestBody)
	if err != nil {
		logger.Logger.Errorw("Validation error", "method", r.Method, "error", err, "request", requestBody, "time", time.Now())
		errs.ValidationError("Invalid email, password, phone number, or gender format").ToJson2(w)
		return
	}

	// Prepare user entity
	userBody := &entities.User{
		Email:        requestBody.Email,
		Password:     requestBody.Password,
		Gender:       requestBody.Gender,
		MobileNumber: requestBody.PhoneNo,
	}

	// Call the signup service
	user, err := a.authService.Signup(r.Context(), userBody)
	if err != nil {
		// email conflict error
		if errors.Is(err, errs.ErrEmailExists) {
			logger.Logger.Warnw("Email conflict during signup", "method", r.Method, "request", requestBody)
			errs.InvalidRequestError("Email already exists").ToJson2(w)
			return
		} else {
			// DB error
			logger.Logger.Warnw("DB error occured while signing up a new user", "method", r.Method, "request", requestBody)
			errs.DBError("db error occured while signing up").ToJson2(w)
			return
		}
	}

	// Create a JWT token
	token, err := utils.CreateJwtToken(user.UserID, user.Role)
	if err != nil {
		logger.Logger.Errorw("Failed to generate token", "method", r.Method, "error", err)
		errs.UnexpectedError("Failed to generate token").ToJson2(w)
		return
	}

	// Return the token as a JSON response
	jsonResponse := map[string]any{
		"code":  http.StatusOK,
		"token": token,
		"role":  user.Role,
	}

	if err = utils.JsonEncoder(w, jsonResponse); err != nil {
		return
	}
	logger.Logger.Infow("Signup successful", "method", r.Method, "email", requestBody.Email, "role", user.Role)
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
		logger.Logger.Errorw("Error decoding request body", "method", r.Method, "error", err, "time", time.Now())
		errs.InvalidRequestError("Invalid or malformed request body").ToJson2(w)
		return
	}

	// Validate the request body
	err = validate.Struct(requestBody)
	if err != nil {
		//errs.NewBadRequestError("Invalid request body").ToJSON(w)
		//logger.Logger.Errorw("Validation error", "method", r.Method, "error", err, "request", requestBody, "time", time.Now())
		//return
		logger.Logger.Errorw("Validation error", "method", r.Method, "error", err, "request", requestBody, "time", time.Now())
		errs.ValidationError("Invalid email or password").ToJson2(w)
		return
	}

	// Call the login service
	user, err := a.authService.Login(r.Context(), requestBody.Email, []byte(requestBody.Password))
	if err != nil {
		if errors.Is(err, errs.ErrInvalidPassword) {
			//errs.NewUnauthorizedError("Invalid username or password").ToJSON(w)
			errs.InvalidRequestError("Invalid username or password").ToJson2(w)
			logger.Logger.Warnw("Invalid password", "method", r.Method, "email", requestBody.Email, "time", time.Now())
			return
		}

		if errors.Is(err, errs.ErrUserNotFound) {
			//errs.NewNotFoundError("No such user exists").ToJSON(w)
			errs.InvalidRequestError("No such user exists").ToJson2(w)
			logger.Logger.Warnw("User not found", "method", r.Method, "email", requestBody.Email, "time", time.Now())
			return
		}

		errs.UnexpectedError("Some internal Server Error").ToJson2(w)
		logger.Logger.Errorw("Internal server error during login", "method", r.Method, "error", err, "request", requestBody, "time", time.Now())
		return
	}

	// Create a JWT token
	token, err := utils.CreateJwtToken(user.UserID, user.Role)
	if err != nil {
		errs.UnexpectedError("Failed to generate token").ToJson2(w)
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
	if err = utils.JsonEncoder(w, jsonResponse); err != nil {
		return
	}
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
	if err := utils.JsonEncoder(w, jsonResponse); err != nil {
		return
	}
	logger.Logger.Infow("Logout successful", "method", r.Method, "time", time.Now(), "duration", time.Since(startTime))
}
