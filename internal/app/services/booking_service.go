package services

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"project2/internal/domain/entities"
	repository_interfaces "project2/internal/domain/interfaces/repository"
	service_interfaces "project2/internal/domain/interfaces/service"
	"project2/internal/models"
	"time"
)

type BookingService struct {
	bookRepo    repository_interfaces.BookingRepository
	SlotService service_interfaces.SlotService
	GameService service_interfaces.GameService
}

func NewBookingService(bookRepo repository_interfaces.BookingRepository, slotService service_interfaces.SlotService, gameService service_interfaces.GameService) service_interfaces.BookingService {
	return &BookingService{
		bookRepo:    bookRepo,
		SlotService: slotService,
		GameService: gameService,
	}
}

func (b *BookingService) MakeBooking(ctx context.Context, userID, slotID uuid.UUID) error {
	// Fetch the slot and validate
	slot, err := b.SlotService.GetSlotByID(ctx, slotID)
	if err != nil {
		return fmt.Errorf("failed to get slot details: %w", err)
	}
	if slot.IsBooked {
		return fmt.Errorf("slot is already booked")
	}
	if slot.StartTime.Before(time.Now()) {
		return fmt.Errorf("slot has already passed")
	}

	// Check if the user is already booked
	if booking, _ := b.bookRepo.FetchBookingBySlotAndUserId(ctx, slotID, userID); booking.BookingId != uuid.Nil {
		return fmt.Errorf("user is already booked in this slot")
	}

	// Create new booking
	newBooking := &entities.Booking{SlotID: slotID, UserID: userID}
	if _, err := b.bookRepo.CreateBooking(ctx, newBooking); err != nil {
		return fmt.Errorf("failed to create booking: %w", err)
	}

	// Fetch game and current bookings
	game, err := b.GameService.GetGameByID(ctx, slot.GameID)
	if err != nil {
		return fmt.Errorf("failed to get game details: %w", err)
	}
	bookings, err := b.bookRepo.FetchBookingsBySlotID(ctx, slotID)
	if err != nil {
		return fmt.Errorf("failed to fetch bookings: %w", err)
	}

	// Mark slot as booked if the max players are reached
	if len(bookings) == game.MaxPlayers {
		if err := b.SlotService.MarkSlotAsBooked(ctx, slotID); err != nil {
			return fmt.Errorf("failed to update slot status: %w", err)
		}
	}

	return nil
}

// GetUpcomingBookings retrieves all upcoming bookings for a given user.
func (b *BookingService) GetUpcomingBookings(ctx context.Context, userID uuid.UUID) ([]models.Bookings, error) {
	return b.bookRepo.FetchUpcomingBookingsByUserID(ctx, userID)
}

func (b *BookingService) GetBookingsToUpdateResult(ctx context.Context, userID uuid.UUID) ([]models.Bookings, error) {
	return b.bookRepo.FetchBookingsToUpdateResult(ctx, userID)
}

func (b *BookingService) UpdateBookingResult(ctx context.Context, bookingId uuid.UUID, result string) error {
	return b.bookRepo.UpdateBookingResult(ctx, bookingId, result)
}

func (b *BookingService) GetSlotBookedUsers(ctx context.Context, slotId uuid.UUID) ([]string, error) {
	return b.bookRepo.FetchSlotBookedUsers(ctx, slotId)
}

func (b *BookingService) GetBookingByUserAndSlotID(ctx context.Context, userID uuid.UUID, slotID uuid.UUID) (models.Bookings, error) {
	return b.bookRepo.FetchBookingBySlotAndUserId(ctx, slotID, userID)
}
