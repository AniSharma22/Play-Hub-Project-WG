package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"project2/internal/domain/entities"
	repository_interfaces "project2/internal/domain/interfaces/repository"
	service_interfaces "project2/internal/domain/interfaces/service"
	"project2/internal/models"
	"project2/pkg/errs"
)

type BookingService struct {
	bookRepo    repository_interfaces.BookingRepository
	SlotService service_interfaces.SlotService
	GameService service_interfaces.GameService
	mu          sync.Mutex // Mutex to handle concurrent bookings
}

func NewBookingService(bookRepo repository_interfaces.BookingRepository, slotService service_interfaces.SlotService, gameService service_interfaces.GameService) service_interfaces.BookingService {
	return &BookingService{
		bookRepo:    bookRepo,
		SlotService: slotService,
		GameService: gameService,
	}
}

func (b *BookingService) MakeBooking(ctx context.Context, userID, slotID, gameID uuid.UUID) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Fetch the game and validate
	game, err := b.GameService.GetGameByID(ctx, gameID)
	if err != nil {
		return fmt.Errorf("failed to get game details: %w", errs.ErrDbError)
	}

	if !game.IsActive {
		return fmt.Errorf("cannot create booking as game is disabled: %w", errs.ErrGameDisabled)
	}

	// Fetch the slot and validate
	slot, err := b.SlotService.GetSlotByID(ctx, slotID)
	if err != nil {
		return fmt.Errorf("failed to get slot details: %w", errs.ErrDbError)
	}
	if slot.IsBooked {
		return fmt.Errorf("slot is already booked: %w", errs.ErrAlreadyExists)
	}
	if slot.StartTime.Before(time.Now()) {
		return fmt.Errorf("slot has already passed: %w", errs.ErrSlotPassed)
	}

	// Check if the user is already booked
	if booking, _ := b.bookRepo.FetchBookingBySlotAndUserId(ctx, slotID, userID); booking.BookingId != uuid.Nil {
		return fmt.Errorf("user is already booked in this slot: %w", errs.ErrUserAlreadyBooked)
	}

	// Create new booking
	newBooking := &entities.Booking{SlotID: slotID, UserID: userID, GameID: gameID}
	fmt.Println(newBooking)
	if _, err := b.bookRepo.CreateBooking(ctx, newBooking); err != nil {
		return fmt.Errorf("failed to create booking: %w", errs.ErrDbError)
	}

	bookings, err := b.bookRepo.FetchBookingsBySlotID(ctx, slotID)
	if err != nil {
		return fmt.Errorf("failed to fetch bookings: %w", errs.ErrDbError)
	}

	// Mark slot as booked if the max players are reached
	if len(bookings) == game.MaxPlayers {
		if err := b.SlotService.MarkSlotAsBooked(ctx, slotID); err != nil {
			return fmt.Errorf("failed to update slot status: %w", errs.ErrServiceError)
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

func (b *BookingService) GetBookingById(ctx context.Context, bookingID uuid.UUID) (*entities.Booking, error) {
	return b.bookRepo.FetchBookingByID(ctx, bookingID)
}

func (b *BookingService) DeleteBookingById(ctx context.Context, bookingID uuid.UUID) error {
	return b.bookRepo.RemoveBookingById(ctx, bookingID)
}
