package service_test

import (
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"project2/internal/domain/entities"
	"testing"
	"time"
)

func TestSlotService_GetCurrentDayGameSlots(t *testing.T) {
	ctx := context.TODO()
	gameID := uuid.New()
	slot := entities.Slot{
		SlotID:    uuid.New(),
		GameID:    gameID,
		StartTime: time.Now(),
		IsBooked:  false,
	}
	slots := []entities.Slot{slot}

	teardown := setup(t)
	defer teardown()

	// Test case: Successful retrieval of current day slots
	mockSlotRepo.EXPECT().FetchSlotsByGameIDAndDate(ctx, gameID, gomock.Any()).Return(slots, nil).Times(1)

	returnedSlots, err := slotService.GetCurrentDayGameSlots(ctx, gameID)
	assert.NoError(t, err)
	assert.Equal(t, slots, returnedSlots)

	// Test case: Error fetching slots
	mockSlotRepo.EXPECT().FetchSlotsByGameIDAndDate(ctx, gameID, gomock.Any()).Return(nil, errors.New("some error")).Times(1)

	_, err = slotService.GetCurrentDayGameSlots(ctx, gameID)
	assert.Error(t, err)
}

func TestSlotService_GetSlotByID(t *testing.T) {
	ctx := context.TODO()
	slotID := uuid.New()
	slot := &entities.Slot{
		SlotID: slotID,
	}

	teardown := setup(t)
	defer teardown()

	// Test case: Successful retrieval of slot by ID
	mockSlotRepo.EXPECT().FetchSlotByID(ctx, slotID).Return(slot, nil).Times(1)

	returnedSlot, err := slotService.GetSlotByID(ctx, slotID)
	assert.NoError(t, err)
	assert.Equal(t, slot, returnedSlot)

	// Test case: Error fetching slot by ID
	mockSlotRepo.EXPECT().FetchSlotByID(ctx, slotID).Return(nil, errors.New("not found")).Times(1)

	_, err = slotService.GetSlotByID(ctx, slotID)
	assert.Error(t, err)
}

func TestSlotService_MarkSlotAsBooked(t *testing.T) {
	ctx := context.TODO()
	slotID := uuid.New()

	teardown := setup(t)
	defer teardown()

	// Test case: Successfully mark slot as booked
	mockSlotRepo.EXPECT().UpdateSlotStatus(ctx, slotID, true).Return(nil).Times(1)

	err := slotService.MarkSlotAsBooked(ctx, slotID)
	assert.NoError(t, err)

	// Test case: Error marking slot as booked
	mockSlotRepo.EXPECT().UpdateSlotStatus(ctx, slotID, true).Return(errors.New("update failed")).Times(1)

	err = slotService.MarkSlotAsBooked(ctx, slotID)
	assert.Error(t, err)
}
