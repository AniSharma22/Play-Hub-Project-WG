package service_interfaces

import (
	"context"
	"github.com/google/uuid"
	"project2/internal/domain/entities"
	"project2/internal/models"
)

type SlotService interface {
	GetCurrentDayGameSlots(ctx context.Context, gameID uuid.UUID) ([]models.SlotDTO, error)
	GetSlotByID(ctx context.Context, slotID uuid.UUID) (*entities.Slot, error)
	MarkSlotAsBooked(ctx context.Context, slotID uuid.UUID) error
}
