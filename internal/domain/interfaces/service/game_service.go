package service_interfaces

import (
	"context"
	"github.com/google/uuid"
	"project2/internal/domain/entities"
)

type GameService interface {
	GetGameByID(ctx context.Context, id uuid.UUID) (*entities.Game, error)
	GetAllGames(ctx context.Context) ([]entities.Game, error)
	GetAllActiveGames(ctx context.Context) ([]entities.Game, error)
	CreateGame(ctx context.Context, game *entities.Game) (uuid.UUID, error)
	DeleteGame(ctx context.Context, id uuid.UUID) error
	UpdateGame(ctx context.Context, game *entities.Game) error
}
