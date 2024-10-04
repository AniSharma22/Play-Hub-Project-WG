package handlers

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"net/http"
	"project2/internal/api/middleware"
	service_interfaces "project2/internal/domain/interfaces/service"
	"project2/internal/models"
	"project2/pkg/errs"
	"project2/pkg/logger"
	"project2/pkg/utils"
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
		errs.InvalidRequestError("Could not find the userId").ToJson2(w)
		logger.Logger.Errorw("User ID not found in context", "method", r.Method, "time", time.Now())
		return
	}

	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		errs.ValidationError("Couldn't parse user id").ToJson2(w)
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
		//http.Error(w, err.Error(), http.StatusBadRequest)
		errs.InvalidRequestError("Invalid or malformed request body").ToJson2(w)
		logger.Logger.Errorw("Error decoding request body", "method", r.Method, "error", err, "time", time.Now())
		return
	}

	// Validate the request body
	err = validate.Struct(requestBody)
	if err != nil {
		errs.ValidationError("Invalid request body").ToJson2(w)
		logger.Logger.Errorw("Validation error", "method", r.Method, "error", err, "request", requestBody, "time", time.Now())
		return
	}
	GameId, err := uuid.Parse(requestBody.GameId)
	if err != nil {
		errs.ValidationError("Couldn't parse game id").ToJson2(w)
		logger.Logger.Errorw("Failed to parse game id", "gameId", requestBody.GameId, "error", err, "time", time.Now())
		return
	}

	InvitedUserId, err := uuid.Parse(requestBody.InvitedUserID)
	if err != nil {
		errs.ValidationError("Couldn't parse invited user id").ToJson2(w)
		logger.Logger.Errorw("Failed to parse invited userId", "InvitedUserID", requestBody.InvitedUserID, "error", err, "time", time.Now())
		return
	}

	SlotId, err := uuid.Parse(requestBody.SlotId)
	if err != nil {
		errs.ValidationError("Couldn't parse slot id").ToJson2(w)
		logger.Logger.Errorw("Failed to parse slotId", "SlotId", requestBody.SlotId, "error", err, "time", time.Now())
		return
	}

	invitationId, err := i.invitationService.MakeInvitation(r.Context(), userId, InvitedUserId, SlotId, GameId)
	if err != nil {
		if errors.Is(err, errs.ErrSelfInviteError) {
			errs.InvalidRequestError("You cannot invite yourself").ToJson2(w)
			logger.Logger.Warnw("Self-invite error", "method", r.Method, "error", err, "time", time.Now())
			return
		} else if errors.Is(err, errs.ErrAlreadyExists) {
			errs.InvalidRequestError("Invitation already exists").ToJson2(w)
			logger.Logger.Warnw("Invitation already exists", "method", r.Method, "error", err, "time", time.Now())
			return
		} else if errors.Is(err, errs.ErrSlotFullyBookedError) {
			errs.InvalidRequestError("The slot is fully booked").ToJson2(w)
			logger.Logger.Warnw("Slot fully booked", "method", r.Method, "error", err, "time", time.Now())
			return
		} else if errors.Is(err, errs.ErrSlotPassed) {
			errs.InvalidRequestError("Cannot invite users to a past slot").ToJson2(w)
			logger.Logger.Warnw("Slot has passed", "method", r.Method, "error", err, "time", time.Now())
			return
		} else {
			errs.UnexpectedError("An unexpected error occurred while creating the invitation").ToJson2(w)
			logger.Logger.Errorw("Unexpected error while creating invitation", "method", r.Method, "error", err, "time", time.Now())
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	jsonResponse := map[string]any{
		"code":          http.StatusOK,
		"message":       "Success",
		"invitation_id": invitationId,
	}
	if err = utils.JsonEncoder(w, jsonResponse); err != nil {
		return
	}
	logger.Logger.Infow("Invite created successfully", "invitationId", invitationId, "method", r.Method, "time", time.Now())
}

func (i *InvitationHandler) UpdateInvitationStatusHandler(w http.ResponseWriter, r *http.Request) {
	logger.Logger.Infow("UpdateInvitationStatusHandler called", "method", r.Method, "time", time.Now())

	vars := mux.Vars(r)
	invitationIdStr := vars["id"]

	invitationId, err := uuid.Parse(invitationIdStr)
	if err != nil {
		errs.ValidationError("Couldn't parse invitation id").ToJson2(w)
		logger.Logger.Errorw("Failed to parse invitationId", "invitationId", invitationIdStr, "error", err, "time", time.Now())
		return
	}

	action := r.URL.Query().Get("action")
	switch action {
	case "accept":
		err = i.invitationService.AcceptInvitation(r.Context(), invitationId)
		if err != nil {
			if errors.Is(err, errs.ErrSlotFullyBookedError) {
				errs.InvalidRequestError("The slot is fully booked").ToJson2(w)
				logger.Logger.Warnw("Slot fully booked", "method", r.Method, "error", err, "time", time.Now())
				return
			} else if errors.Is(err, errs.ErrUserAlreadyBooked) {
				errs.InvalidRequestError("User has already booked in this slot").ToJson2(w)
				logger.Logger.Warnw("Slot fully booked", "method", r.Method, "error", err, "time", time.Now())
				return
			} else {
				errs.UnexpectedError("An unexpected error occurred while accepting the invitation").ToJson2(w)
				logger.Logger.Errorw("Unexpected error while accepting invitation", "method", r.Method, "error", err, "time", time.Now())
				return
			}
		}
		logger.Logger.Infow("Invitation accepted", "invitationId", invitationId, "time", time.Now())

	case "reject":
		err = i.invitationService.RejectInvitation(r.Context(), invitationId)
		if err != nil {
			errs.UnexpectedError("An unexpected error occurred while rejecting the invitation").ToJson2(w)
			logger.Logger.Errorw("Unexpected error while rejecting invitation", "method", r.Method, "error", err, "time", time.Now())
			return
		}
		logger.Logger.Infow("Invitation rejected", "invitationId", invitationId, "time", time.Now())

	default:
		errs.InvalidRequestError("Invalid action (should be accept or reject)").ToJson2(w)
		logger.Logger.Warnw("Invalid action attempted", "action", action, "time", time.Now())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jsonResponse := map[string]any{
		"code":    http.StatusOK,
		"message": "Success",
	}
	if err = utils.JsonEncoder(w, jsonResponse); err != nil {
		return
	}
	logger.Logger.Infow("Invitation status updated successfully", "invitationId", invitationId, "action", action, "time", time.Now())
}

func (i *InvitationHandler) GetPendingInvitationHandler(w http.ResponseWriter, r *http.Request) {
	logger.Logger.Infow("GetPendingInvitationHandler called", "method", r.Method, "time", time.Now())

	userIdStr, ok := r.Context().Value(middleware.UserIdKey).(string)
	if !ok {
		errs.UnexpectedError("Could not find the userId").ToJson2(w)
		logger.Logger.Errorw("User ID not found in context", "method", r.Method, "time", time.Now())
		return
	}

	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		errs.ValidationError("Couldn't parse user id").ToJson2(w)
		logger.Logger.Errorw("Failed to parse userId", "userId", userIdStr, "error", err, "time", time.Now())
		return
	}

	invitations, err := i.invitationService.GetAllPendingInvitations(r.Context(), userId)
	if err != nil {
		errs.DBError("Couldn't get pending invitations").ToJson2(w)
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
	if err = utils.JsonEncoder(w, jsonResponse); err != nil {
		return
	}
	logger.Logger.Infow("Pending Invitations sent successfully", "userId", userId, "method", r.Method, "time", time.Now())
}
