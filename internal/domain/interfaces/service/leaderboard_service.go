package service_interfaces

import (
	"context"
	"github.com/google/uuid"
	"project2/internal/domain/entities"
	"project2/internal/models"
)

type LeaderboardService interface {
	GetGameLeaderboard(ctx context.Context, gameId uuid.UUID) ([]models.LeaderboardDTO, error)
	GetUserGameStats(ctx context.Context, userId uuid.UUID, gameId uuid.UUID) (*entities.Leaderboard, error)
	AddWinToUser(ctx context.Context, userId uuid.UUID, gameId uuid.UUID, bookingId uuid.UUID) error
	AddLossToUser(ctx context.Context, userId uuid.UUID, gameId uuid.UUID, bookingId uuid.UUID) error
}
