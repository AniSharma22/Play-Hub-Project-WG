package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"project2/internal/api/handlers"
	"project2/internal/api/middleware"
	"project2/internal/models"
	mocks2 "project2/tests/mocks/service"
	"testing"
)

func TestCreateInvitationHandler_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockInvitationService := mocks2.NewMockInvitationService(ctrl)

	// Create a mock invitation ID
	invitationID := uuid.New()

	// Setup handler
	handler := handlers.NewInvitationHandler(mockInvitationService)

	// Setup mock for service behavior
	mockInvitationService.EXPECT().MakeInvitation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(invitationID, nil)

	// Define the input request
	body := map[string]interface{}{
		"invited_user_id": uuid.New().String(),
		"slot_id":         uuid.New().String(),
		"game_id":         uuid.New().String(),
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/invitation", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIdKey, uuid.New().String()))

	// Setup response recorder
	rr := httptest.NewRecorder()

	// Call the handler
	handler.CreateInvitationHandler(rr, req)

	// Verify the status code
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status OK, but got %v", rr.Code)
	}

	type response struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    string `json:"invitation_id"`
	}
	expectedResponse := response{
		Code:    200,
		Message: "Success",
		Data:    invitationID.String(),
	}
	var actualResponse response

	json.Unmarshal(rr.Body.Bytes(), &actualResponse)

	assert.Equal(t, expectedResponse.Code, actualResponse.Code)
	assert.Equal(t, expectedResponse.Data, actualResponse.Data)
	assert.Equal(t, expectedResponse.Message, actualResponse.Message)
}

func TestCreateInvitationHandler_ValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockInvitationService := mocks2.NewMockInvitationService(ctrl)
	handler := handlers.NewInvitationHandler(mockInvitationService)

	// Define invalid input (missing required fields)
	body := map[string]interface{}{
		"slot_id": uuid.New().String(),
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/invitation", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIdKey, uuid.New().String()))

	// Setup response recorder
	rr := httptest.NewRecorder()

	// Call the handler
	handler.CreateInvitationHandler(rr, req)

	// Verify the status code
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status BadRequest, but got %v", rr.Code)
	}

	// Check for validation error message
	//expectedResponse := map[string]interface{}{
	//	"code":    http.StatusBadRequest,
	//	"message": "Invalid request body",
	//}
	//expectedJSON, _ := json.Marshal(expectedResponse)
	//if rr.Body.String() != string(expectedJSON) {
	//	t.Errorf("Expected response %s, but got %s", string(expectedJSON), rr.Body.String())
	//}
}

func TestUpdateInvitationStatusHandler_Accept(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockInvitationService := mocks2.NewMockInvitationService(ctrl)
	handler := handlers.NewInvitationHandler(mockInvitationService)

	invitationID := uuid.New()

	// Set up mock for successful accept
	mockInvitationService.EXPECT().AcceptInvitation(gomock.Any(), invitationID).Return(nil)

	// Create the request with "accept" action
	req := httptest.NewRequest(http.MethodPatch, "/invitation/"+invitationID.String()+"?action=accept", nil)
	req = mux.SetURLVars(req, map[string]string{"id": invitationID.String()})

	rr := httptest.NewRecorder()

	// Call the handler
	handler.UpdateInvitationStatusHandler(rr, req)

	// Verify the status code
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status OK, but got %v", rr.Code)
	}

	// Verify response message
	//expectedResponse := map[string]interface{}{
	//	"code":    http.StatusOK,
	//	"message": "Success",
	//}
	//expectedJSON, _ := json.Marshal(expectedResponse)
	//if rr.Body.String() != string(expectedJSON) {
	//	t.Errorf("Expected response %s, but got %s", string(expectedJSON), rr.Body.String())
	//}
}

func TestUpdateInvitationStatusHandler_Reject(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockInvitationService := mocks2.NewMockInvitationService(ctrl)
	handler := handlers.NewInvitationHandler(mockInvitationService)

	invitationID := uuid.New()

	// Set up mock for successful reject
	mockInvitationService.EXPECT().RejectInvitation(gomock.Any(), invitationID).Return(nil)

	// Create the request with "reject" action
	req := httptest.NewRequest(http.MethodPatch, "/invitation/"+invitationID.String()+"?action=reject", nil)
	req = mux.SetURLVars(req, map[string]string{"id": invitationID.String()})

	rr := httptest.NewRecorder()

	// Call the handler
	handler.UpdateInvitationStatusHandler(rr, req)

	// Verify the status code
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status OK, but got %v", rr.Code)
	}

	// Verify response message
	//expectedResponse := map[string]interface{}{
	//	"code":    http.StatusOK,
	//	"message": "Success",
	//}
	//expectedJSON, _ := json.Marshal(expectedResponse)
	//if rr.Body.String() != string(expectedJSON) {
	//	t.Errorf("Expected response %s, but got %s", string(expectedJSON), rr.Body.String())
	//}
}

func TestGetPendingInvitationHandler_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockInvitationService := mocks2.NewMockInvitationService(ctrl)
	handler := handlers.NewInvitationHandler(mockInvitationService)

	userId := uuid.New()

	// Mock the pending invitations
	invitations := []models.Invitations{
		{InvitationId: uuid.New(), SlotId: uuid.New(), GameId: uuid.New()},
	}

	mockInvitationService.EXPECT().GetAllPendingInvitations(gomock.Any(), userId).Return(invitations, nil)

	req := httptest.NewRequest(http.MethodGet, "/pending_invitations", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIdKey, userId.String()))

	rr := httptest.NewRecorder()

	handler.GetPendingInvitationHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status OK, but got %v", rr.Code)
	}

	// Expected response
	//expectedResponse := map[string]interface{}{
	//	"code":                http.StatusOK,
	//	"message":             "Success",
	//	"pending_invitations": invitations,
	//}
	//expectedJSON, _ := json.Marshal(expectedResponse)
	//if rr.Body.String() != string(expectedJSON) {
	//	t.Errorf("Expected response %s, but got %s", string(expectedJSON), rr.Body.String())
	//}
}
