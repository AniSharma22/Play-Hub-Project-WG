package services

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"project2/internal/domain/entities"
	repository_interfaces "project2/internal/domain/interfaces/repository"
	service_interfaces "project2/internal/domain/interfaces/service"
	"project2/pkg/errs"
	"sync"
)

type GameService struct {
	gameRepo repository_interfaces.GameRepository
	gameWG   *sync.WaitGroup
}

func NewGameService(gameRepo repository_interfaces.GameRepository) service_interfaces.GameService {
	return &GameService{
		gameRepo: gameRepo,
		gameWG:   &sync.WaitGroup{},
	}
}

// GetGameByID retrieves a game by its ID
func (s *GameService) GetGameByID(ctx context.Context, id uuid.UUID) (*entities.Game, error) {
	game, err := s.gameRepo.FetchGameByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get game by ID: %w", err)
	}
	return game, nil
}

// GetAllGames retrieves all games
func (s *GameService) GetAllGames(ctx context.Context) ([]entities.Game, error) {
	games, err := s.gameRepo.FetchAllGames(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all games: %w", err)
	}
	return games, nil
}

// CreateGame creates a new game
func (s *GameService) CreateGame(ctx context.Context, game *entities.Game) (uuid.UUID, error) {
	id, err := s.gameRepo.CreateGame(ctx, game)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create game: %w", err)
	}
	return id, nil
}

// DeleteGame deletes a game by its ID
func (s *GameService) DeleteGame(ctx context.Context, id uuid.UUID) error {
	err := s.gameRepo.DeleteGame(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete game: %w", err)
	}
	return nil
}

// UpdateGameStatus updates the status of a game (e.g., activate/deactivate)
func (s *GameService) UpdateGameStatus(ctx context.Context, id uuid.UUID) error {
	game, err := s.gameRepo.FetchGameByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to fetch game by ID: %w", err)
	}
	if game == nil {
		return fmt.Errorf("game not found: %w", errs.ErrGameNotFound)
	}

	err = s.gameRepo.UpdateGameStatus(ctx, game.GameID, !game.IsActive)
	if err != nil {
		return fmt.Errorf("failed to update game status: %w", err)
	}
	return nil
}
