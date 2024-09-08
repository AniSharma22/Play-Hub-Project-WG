package service_test

import (
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"project2/internal/domain/entities"
	"project2/internal/models"
	"testing"
	"time"
)

func TestInvitationService_MakeInvitation(t *testing.T) {
	setup := setup(t) // Assuming you have a setup function
	defer setup()

	invitingUserID := uuid.New()
	invitedUserID := uuid.New()
	slotID := uuid.New()
	ctx := context.TODO()

	tests := []struct {
		name           string
		mockSetup      func()
		expectedResult uuid.UUID
		expectedError  bool
	}{
		{
			name: "success",
			mockSetup: func() {
				mockInvitationRepo.EXPECT().FetchInvitationByUserAndSlot(ctx, invitingUserID, invitedUserID, slotID).Return(nil, nil).Times(1)
				mockSlotService.EXPECT().GetSlotByID(ctx, slotID).Return(&entities.Slot{IsBooked: false, EndTime: time.Now().Add(1 * time.Hour)}, nil).Times(1)
				mockInvitationRepo.EXPECT().CreateInvitation(ctx, gomock.Any()).Return(uuid.New(), nil).Times(1)
			},
			expectedResult: uuid.New(),
			expectedError:  false,
		},
		{
			name: "invitation already exists",
			mockSetup: func() {
				mockInvitationRepo.EXPECT().FetchInvitationByUserAndSlot(ctx, invitingUserID, invitedUserID, slotID).Return(&entities.Invitation{}, nil)
			},
			expectedResult: uuid.Nil,
			expectedError:  true,
		},
		{
			name: "slot already booked",
			mockSetup: func() {
				mockInvitationRepo.EXPECT().FetchInvitationByUserAndSlot(ctx, invitingUserID, invitedUserID, slotID).Return(nil, nil)
				mockSlotService.EXPECT().GetSlotByID(ctx, slotID).Return(&entities.Slot{IsBooked: true}, nil)
			},
			expectedResult: uuid.Nil,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			result, err := invitationService.MakeInvitation(ctx, invitingUserID, invitedUserID, slotID)

			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected an error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect an error, but got: %v", err)
				}
				if result == uuid.Nil {
					t.Errorf("Expected a valid invitation ID, but got none")
				}
			}
		})
	}
}

func TestInvitationService_AcceptInvitation(t *testing.T) {

	invitationID := uuid.New()
	ctx := context.TODO()
	slotID := uuid.New()

	tests := []struct {
		name          string
		mockSetup     func()
		expectedError bool
	}{
		{
			name: "success",
			mockSetup: func() {
				mockInvitationRepo.EXPECT().FetchInvitationByID(ctx, invitationID).Return(&entities.Invitation{SlotID: slotID}, nil)
				mockSlotService.EXPECT().GetSlotByID(ctx, slotID).Return(&entities.Slot{IsBooked: false}, nil)
				mockBookingService.EXPECT().GetBookingByUserAndSlotID(ctx, gomock.Any(), slotID).Return(models.Bookings{BookingId: uuid.Nil}, nil)
				mockBookingService.EXPECT().MakeBooking(ctx, gomock.Any(), slotID).Return(nil)
				mockInvitationRepo.EXPECT().DeleteInvitationByID(ctx, invitationID).Return(nil)
			},
			expectedError: false,
		},
		{
			name: "slot already booked",
			mockSetup: func() {
				mockInvitationRepo.EXPECT().FetchInvitationByID(ctx, invitationID).Return(&entities.Invitation{SlotID: slotID}, nil)
				mockSlotService.EXPECT().GetSlotByID(ctx, slotID).Return(&entities.Slot{IsBooked: true}, nil)
				mockInvitationRepo.EXPECT().DeleteInvitationByID(ctx, invitationID).Return(nil)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teardown := setup(t)
			defer teardown()
			tt.mockSetup()

			err := invitationService.AcceptInvitation(ctx, invitationID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInvitationService_RejectInvitation(t *testing.T) {
	setup := setup(t)
	defer setup()

	invitationID := uuid.New()
	ctx := context.TODO()

	tests := []struct {
		name          string
		mockSetup     func()
		expectedError bool
	}{
		{
			name: "success",
			mockSetup: func() {
				mockInvitationRepo.EXPECT().DeleteInvitationByID(ctx, invitationID).Return(nil)
			},
			expectedError: false,
		},
		{
			name: "error while deleting",
			mockSetup: func() {
				mockInvitationRepo.EXPECT().DeleteInvitationByID(ctx, invitationID).Return(errors.New("delete error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			err := invitationService.RejectInvitation(ctx, invitationID)

			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected an error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect an error, but got: %v", err)
				}
			}
		})
	}
}

func TestInvitationService_GetAllPendingInvitations(t *testing.T) {

	userID := uuid.New()
	ctx := context.TODO()
	pendingInvitations := []models.Invitations{
		{
			InvitationId: uuid.New(),
			SlotId:       uuid.New(),
		},
	}

	tests := []struct {
		name           string
		mockSetup      func()
		expectedOutput []models.Invitations
		expectedError  bool
	}{
		{
			name: "success",
			mockSetup: func() {
				mockInvitationRepo.EXPECT().FetchUserPendingInvitations(ctx, userID).Return(pendingInvitations, nil)
			},
			expectedOutput: pendingInvitations,
			expectedError:  false,
		},
		{
			name: "error fetching invitations",
			mockSetup: func() {
				mockInvitationRepo.EXPECT().FetchUserPendingInvitations(ctx, userID).Return(nil, errors.New("fetch error"))
			},
			expectedOutput: nil,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teardown := setup(t)
			defer teardown()
			tt.mockSetup()

			result, err := invitationService.GetAllPendingInvitations(ctx, userID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOutput, result)
			}
		})
	}
}
