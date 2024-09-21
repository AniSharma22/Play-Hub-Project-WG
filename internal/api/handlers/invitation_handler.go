package handlers

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"net/http"
	"project2/internal/api/middleware"
	service_interfaces "project2/internal/domain/interfaces/service"
	"project2/internal/models"
	"project2/pkg/errs"
	"project2/pkg/logger"
	"time"
)

type InvitationHandler struct {
	invitationService service_interfaces.InvitationService
}

func NewInvitationHandler(invitationService service_interfaces.InvitationService) *InvitationHandler {
	return &InvitationHandler{
		invitationService: invitationService,
	}
}

func (i *InvitationHandler) CreateInvitationHandler(w http.ResponseWriter, r *http.Request) {
	logger.Logger.Infow("CreateInvitationHandler called", "method", r.Method, "time", time.Now())

	// Extract userId from the context
	userIdStr, ok := r.Context().Value(middleware.UserIdKey).(string)
	if !ok {
		errs.NewUnexpectedError("Could not find the userId").ToJSON(w)
		logger.Logger.Errorw("User ID not found in context", "method", r.Method, "time", time.Now())
		return
	}

	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		errs.NewInternalServerError("Couldn't parse user id").ToJSON(w)
		logger.Logger.Errorw("Failed to parse userId", "userId", userIdStr, "error", err, "time", time.Now())
		return
	}

	var requestBody struct {
		InvitedUserID string `json:"invited_user_id" validate:"required"`
		SlotId        string `json:"slot_id" validate:"required"`
		GameId        string `json:"game_id" validate:"required"`
	}

	// decode the request body
	err = json.NewDecoder(r.Body).Decode(&requestBody)
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
	GameId, err := uuid.Parse(requestBody.GameId)
	if err != nil {
		errs.NewInternalServerError("Couldn't parse game id").ToJSON(w)
		logger.Logger.Errorw("Failed to parse game id", "gameId", requestBody.GameId, "error", err, "time", time.Now())
		return
	}

	InvitedUserId, err := uuid.Parse(requestBody.InvitedUserID)
	if err != nil {
		errs.NewInternalServerError("Couldn't parse invited user id").ToJSON(w)
		logger.Logger.Errorw("Failed to parse invited userId", "InvitedUserID", requestBody.InvitedUserID, "error", err, "time", time.Now())
		return
	}

	SlotId, err := uuid.Parse(requestBody.SlotId)
	if err != nil {
		errs.NewInternalServerError("Couldn't parse slot id").ToJSON(w)
		logger.Logger.Errorw("Failed to parse slotId", "SlotId", requestBody.SlotId, "error", err, "time", time.Now())
		return
	}

	invitationId, err := i.invitationService.MakeInvitation(r.Context(), userId, InvitedUserId, SlotId, GameId)
	if err != nil {
		errs.NewInternalServerError("Couldn't create invitation").ToJSON(w)
		logger.Logger.Errorw("Failed to create invitation", "method", r.Method, "error", err, "time", time.Now())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jsonResponse := map[string]any{
		"code":    http.StatusOK,
		"message": "Success",
		"data":    invitationId,
	}
	json.NewEncoder(w).Encode(jsonResponse)
	logger.Logger.Infow("Invite created successfully", "invitationId", invitationId, "method", r.Method, "time", time.Now())
}

func (i *InvitationHandler) UpdateInvitationStatusHandler(w http.ResponseWriter, r *http.Request) {
	logger.Logger.Infow("UpdateInvitationStatusHandler called", "method", r.Method, "time", time.Now())

	vars := mux.Vars(r)
	invitationIdStr := vars["id"]

	invitationId, err := uuid.Parse(invitationIdStr)
	if err != nil {
		errs.NewInternalServerError("Couldn't parse invitation id").ToJSON(w)
		logger.Logger.Errorw("Failed to parse invitationId", "invitationId", invitationIdStr, "error", err, "time", time.Now())
		return
	}

	action := r.URL.Query().Get("action")
	switch action {
	case "accept":
		err = i.invitationService.AcceptInvitation(r.Context(), invitationId)
		if err != nil {
			errs.NewInternalServerError("Couldn't accept invitation").ToJSON(w)
			logger.Logger.Errorw("Failed to accept invitation", "invitationId", invitationId, "error", err, "time", time.Now())
			return
		}
		logger.Logger.Infow("Invitation accepted", "invitationId", invitationId, "time", time.Now())

	case "reject":
		err = i.invitationService.RejectInvitation(r.Context(), invitationId)
		if err != nil {
			errs.NewInternalServerError("Couldn't reject invitation").ToJSON(w)
			logger.Logger.Errorw("Failed to reject invitation", "invitationId", invitationId, "error", err, "time", time.Now())
			return
		}
		logger.Logger.Infow("Invitation rejected", "invitationId", invitationId, "time", time.Now())

	default:
		errs.NewBadRequestError("Invalid action").ToJSON(w)
		logger.Logger.Warnw("Invalid action attempted", "action", action, "time", time.Now())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jsonResponse := map[string]any{
		"code":    http.StatusOK,
		"message": "Success",
	}
	json.NewEncoder(w).Encode(jsonResponse)
	logger.Logger.Infow("Invitation status updated successfully", "invitationId", invitationId, "action", action, "time", time.Now())
}

func (i *InvitationHandler) GetPendingInvitationHandler(w http.ResponseWriter, r *http.Request) {
	logger.Logger.Infow("GetPendingInvitationHandler called", "method", r.Method, "time", time.Now())

	userIdStr, ok := r.Context().Value(middleware.UserIdKey).(string)
	if !ok {
		errs.NewUnexpectedError("Could not find the userId").ToJSON(w)
		logger.Logger.Errorw("User ID not found in context", "method", r.Method, "time", time.Now())
		return
	}

	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		errs.NewInternalServerError("Couldn't parse user id").ToJSON(w)
		logger.Logger.Errorw("Failed to parse userId", "userId", userIdStr, "error", err, "time", time.Now())
		return
	}

	invitations, err := i.invitationService.GetAllPendingInvitations(r.Context(), userId)
	if err != nil {
		errs.NewInternalServerError("Couldn't get pending invitations").ToJSON(w)
		logger.Logger.Errorw("Failed to fetch pending invitations", "userId", userId, "error", err, "time", time.Now())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jsonResponse := map[string]any{
		"code":    http.StatusOK,
		"message": "Success",
		"pending_invitations": func() []models.Invitations {
			if invitations == nil {
				return []models.Invitations{}
			}
			return invitations
		}(),
	}
	json.NewEncoder(w).Encode(jsonResponse)
	logger.Logger.Infow("Pending Invitations sent successfully", "userId", userId, "method", r.Method, "time", time.Now())
}
