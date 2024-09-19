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

	var requestBody struct {
		InvitedUserID string `json:"invited_user_id" validate:"required"`
		SlotId        string `json:"slot_id" validate:"required"`
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
		errs.NewInvalidBodyError("Invalid request body").ToJSON(w)
		logger.Logger.Errorw("Validation error", "method", r.Method, "error", err, "request", requestBody, "time", time.Now())
		return
	}

	InvitedUserId, err := uuid.Parse(requestBody.InvitedUserID)
	if err != nil {
		errs.NewInternalServerError("Couldn't parse user id").ToJSON(w)
		return
	}

	SlotId, err := uuid.Parse(requestBody.SlotId)
	if err != nil {
		errs.NewInternalServerError("Couldn't parse slot id").ToJSON(w)
		return
	}

	invitationId, err := i.invitationService.MakeInvitation(r.Context(), userId, InvitedUserId, SlotId)
	if err != nil {
		errs.NewInternalServerError("Couldn't create invitation").ToJSON(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jsonResponse := map[string]any{
		"code":    http.StatusOK,
		"message": "Success",
		"data":    invitationId,
	}
	json.NewEncoder(w).Encode(jsonResponse)
	logger.Logger.Infow("Invite created successfully", "method", r.Method, "time", time.Now())
}

func (i *InvitationHandler) UpdateInvitationStatusHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	invitationIdStr := vars["id"]

	invitationId, err := uuid.Parse(invitationIdStr)
	if err != nil {
		errs.NewInternalServerError("Couldn't parse invitation id").ToJSON(w)
		return
	}

	action := r.URL.Query().Get("action")
	switch action {
	case "accept":
		err = i.invitationService.AcceptInvitation(r.Context(), invitationId)
		if err != nil {
			errs.NewInternalServerError("Couldn't accept invitation").ToJSON(w)
			return
		}

	case "reject":
		err = i.invitationService.RejectInvitation(r.Context(), invitationId)
		if err != nil {
			errs.NewInternalServerError("Couldn't reject invitation").ToJSON(w)
			return
		}

	default:
		errs.NewBadRequestError("Invalid action").ToJSON(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jsonResponse := map[string]any{
		"code":    http.StatusOK,
		"message": "Success",
	}
	json.NewEncoder(w).Encode(jsonResponse)
	logger.Logger.Infow("Invitation Update Successful", "method", r.Method, "time", time.Now())
}

func (i *InvitationHandler) GetPendingInvitationHandler(w http.ResponseWriter, r *http.Request) {
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

	invitations, err := i.invitationService.GetAllPendingInvitations(r.Context(), userId)
	if err != nil {
		errs.NewInternalServerError("Couldn't get pending invitations").ToJSON(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jsonResponse := map[string]any{
		"code":    http.StatusOK,
		"message": "Success",
		"data": func() []models.Invitations {
			if invitations == nil {
				return []models.Invitations{}
			}
			return invitations
		}(),
	}
	json.NewEncoder(w).Encode(jsonResponse)
	logger.Logger.Infow("Pending Invitations sent successfully", "method", r.Method, "time", time.Now())
}
