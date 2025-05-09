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

func TestBookingService_MakeBooking(t *testing.T) {
	teardown := setup(t)
	defer teardown()
	ctx := context.TODO()

	userID := uuid.New()
	slotID := uuid.New()
	gameID := uuid.New()

	t.Run("should fail to get slot details", func(t *testing.T) {
		mockSlotService.EXPECT().GetSlotByID(ctx, slotID).Return(nil, errors.New("slot not found"))

		err := bookingService.MakeBooking(ctx, userID, slotID)
		assert.EqualError(t, err, "failed to get slot details: slot not found")
	})

	t.Run("should fail if slot is already booked", func(t *testing.T) {
		slot := &entities.Slot{
			IsBooked: true,
		}
		mockSlotService.EXPECT().GetSlotByID(ctx, slotID).Return(slot, nil)

		err := bookingService.MakeBooking(ctx, userID, slotID)
		assert.EqualError(t, err, "slot is already booked")
	})

	t.Run("should fail if slot time has passed", func(t *testing.T) {
		slot := &entities.Slot{
			IsBooked:  false,
			StartTime: time.Now().Add(-time.Hour),
		}
		mockSlotService.EXPECT().GetSlotByID(ctx, slotID).Return(slot, nil)

		err := bookingService.MakeBooking(ctx, userID, slotID)
		assert.EqualError(t, err, "slot has already passed")
	})

	t.Run("should fail if user is already booked", func(t *testing.T) {
		slot := &entities.Slot{
			IsBooked:  false,
			StartTime: time.Now().Add(time.Hour),
			GameID:    gameID,
		}
		mockSlotService.EXPECT().GetSlotByID(ctx, slotID).Return(slot, nil)
		mockBookingRepo.EXPECT().FetchBookingBySlotAndUserId(ctx, slotID, userID).Return(models.Bookings{BookingId: uuid.New()}, nil)

		err := bookingService.MakeBooking(ctx, userID, slotID)
		assert.EqualError(t, err, "user is already booked in this slot")
	})

	t.Run("should fail to create booking", func(t *testing.T) {
		slot := &entities.Slot{
			IsBooked:  false,
			StartTime: time.Now().Add(time.Hour),
			GameID:    gameID,
		}
		mockSlotService.EXPECT().GetSlotByID(ctx, slotID).Return(slot, nil)
		// Return a booking with a nil UUID, meaning the user hasn't booked this slot yet
		mockBookingRepo.EXPECT().FetchBookingBySlotAndUserId(ctx, slotID, userID).Return(models.Bookings{BookingId: uuid.Nil}, nil)
		// Simulate failure in CreateBooking
		mockBookingRepo.EXPECT().CreateBooking(ctx, gomock.Any()).Return(uuid.New(), errors.New("create booking failed"))

		err := bookingService.MakeBooking(ctx, userID, slotID)
		assert.EqualError(t, err, "failed to create booking: create booking failed")
	})

	t.Run("should fail to get game details", func(t *testing.T) {
		slot := &entities.Slot{
			IsBooked:  false,
			StartTime: time.Now().Add(time.Hour),
			GameID:    gameID,
		}
		mockSlotService.EXPECT().GetSlotByID(ctx, slotID).Return(slot, nil)
		mockBookingRepo.EXPECT().FetchBookingBySlotAndUserId(ctx, slotID, userID).Return(models.Bookings{BookingId: uuid.Nil}, nil)
		mockBookingRepo.EXPECT().CreateBooking(ctx, gomock.Any()).Return(uuid.New(), nil)
		mockGameService.EXPECT().GetGameByID(ctx, gameID).Return(nil, errors.New("game not found"))

		err := bookingService.MakeBooking(ctx, userID, slotID)
		assert.EqualError(t, err, "failed to get game details: game not found")
	})

	t.Run("should fail to fetch bookings for slot", func(t *testing.T) {
		slot := &entities.Slot{
			IsBooked:  false,
			StartTime: time.Now().Add(time.Hour),
			GameID:    gameID,
		}
		mockSlotService.EXPECT().GetSlotByID(ctx, slotID).Return(slot, nil)
		mockBookingRepo.EXPECT().FetchBookingBySlotAndUserId(ctx, slotID, userID).Return(models.Bookings{BookingId: uuid.Nil}, nil)
		mockBookingRepo.EXPECT().CreateBooking(ctx, gomock.Any()).Return(uuid.New(), nil)
		mockGameService.EXPECT().GetGameByID(ctx, gameID).Return(&entities.Game{MaxPlayers: 4}, nil)
		mockBookingRepo.EXPECT().FetchBookingsBySlotID(ctx, slotID).Return(nil, errors.New("failed to fetch bookings"))

		err := bookingService.MakeBooking(ctx, userID, slotID)
		assert.EqualError(t, err, "failed to fetch bookings: failed to fetch bookings")
	})

	t.Run("should mark slot as booked when max players reached", func(t *testing.T) {
		slot := &entities.Slot{
			IsBooked:  false,
			StartTime: time.Now().Add(time.Hour),
			GameID:    gameID,
		}
		mockSlotService.EXPECT().GetSlotByID(ctx, slotID).Return(slot, nil)
		mockBookingRepo.EXPECT().FetchBookingBySlotAndUserId(ctx, slotID, userID).Return(models.Bookings{BookingId: uuid.Nil}, nil)
		mockBookingRepo.EXPECT().CreateBooking(ctx, gomock.Any()).Return(uuid.New(), nil)
		mockGameService.EXPECT().GetGameByID(ctx, gameID).Return(&entities.Game{MaxPlayers: 2}, nil)
		mockBookingRepo.EXPECT().FetchBookingsBySlotID(ctx, slotID).Return([]entities.Booking{{}, {}}, nil)
		mockSlotService.EXPECT().MarkSlotAsBooked(ctx, slotID).Return(nil)

		err := bookingService.MakeBooking(ctx, userID, slotID)
		assert.NoError(t, err)
	})
}

func TestBookingService_GetUpcomingBookings(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	// Define inputs
	userID := uuid.New()

	// Mock return data
	expectedBookings := []models.Bookings{
		{BookingId: uuid.New()},
	}

	// Define mocks
	mockBookingRepo.EXPECT().FetchUpcomingBookingsByUserID(gomock.Any(), userID).Return(expectedBookings, nil)

	// Call the service method
	bookings, err := bookingService.GetUpcomingBookings(context.TODO(), userID)

	// Assert no errs and correct return value
	assert.NoError(t, err)
	assert.Equal(t, len(bookings), len(expectedBookings))
}

func TestBookingService_GetBookingsToUpdateResult(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	// Define inputs
	userID := uuid.New()

	// Mock return data
	expectedBookings := []models.Bookings{
		{BookingId: uuid.New()},
	}

	// Define mocks
	mockBookingRepo.EXPECT().FetchBookingsToUpdateResult(gomock.Any(), userID).Return(expectedBookings, nil)

	// Call the service method
	bookings, err := bookingService.GetBookingsToUpdateResult(context.TODO(), userID)

	// Assert no errs and correct return value
	assert.NoError(t, err)
	assert.Equal(t, len(bookings), len(expectedBookings))
}

func TestBookingService_UpdateBookingResult(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	// Define inputs
	bookingID := uuid.New()
	result := "win"

	// Define mocks
	mockBookingRepo.EXPECT().UpdateBookingResult(gomock.Any(), bookingID, result).Return(nil)

	// Call the service method
	err := bookingService.UpdateBookingResult(context.TODO(), bookingID, result)

	// Assert no errs
	assert.NoError(t, err)
}

func TestBookingService_GetSlotBookedUsers(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	// Define inputs
	slotID := uuid.New()

	// Mock return data
	expectedUsers := []string{"user1", "user2"}

	// Define mocks
	mockBookingRepo.EXPECT().FetchSlotBookedUsers(gomock.Any(), slotID).Return(expectedUsers, nil)

	// Call the service method
	users, err := bookingService.GetSlotBookedUsers(context.TODO(), slotID)

	// Assert no errs and correct return value
	assert.NoError(t, err)
	assert.Equal(t, len(users), len(expectedUsers))

}

func TestBookingService_GetBookingByUserAndSlotID(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	// Define inputs
	userID := uuid.New()
	slotID := uuid.New()
	bookingID := uuid.New()

	// Mock return data
	expectedBooking := models.Bookings{
		BookingId: bookingID,
	}

	// Define mocks
	mockBookingRepo.EXPECT().FetchBookingBySlotAndUserId(gomock.Any(), slotID, userID).Return(expectedBooking, nil)

	// Call the service method
	booking, err := bookingService.GetBookingByUserAndSlotID(context.TODO(), userID, slotID)

	// Assert no errs and correct return value
	assert.NoError(t, err)
	assert.Equal(t, booking.BookingId, expectedBooking.BookingId)
}
