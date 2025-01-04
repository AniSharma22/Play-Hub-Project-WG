package services

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"project2/internal/domain/entities"
	repository_interfaces "project2/internal/domain/interfaces/repository"
	service_interfaces "project2/internal/domain/interfaces/service"
	"project2/pkg/errs"
	"project2/pkg/utils"
	"sync"
)

type GameService struct {
	gameRepo repository_interfaces.GameRepository
	slotRepo repository_interfaces.SlotRepository
	gameWG   *sync.WaitGroup
}

func NewGameService(gameRepo repository_interfaces.GameRepository, slotRepo repository_interfaces.SlotRepository) service_interfaces.GameService {
	return &GameService{
		gameRepo: gameRepo,
		slotRepo: slotRepo,
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

// GetAllActiveGames retrieves all games
func (s *GameService) GetAllActiveGames(ctx context.Context) ([]entities.Game, error) {
	games, err := s.gameRepo.FetchAllActiveGames(ctx)
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

	// insert slots for the newly created game
	go utils.InsertAllSlots(context.TODO(), s.slotRepo, s.gameRepo)
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

// UpdateGame updates the details of a game
func (s *GameService) UpdateGame(ctx context.Context, game *entities.Game) error {
	gameTemp, err := s.gameRepo.FetchGameByID(ctx, game.GameID)
	if err != nil {
		return fmt.Errorf("failed to fetch game by ID: %w", err)
	}
	if gameTemp == nil {
		return fmt.Errorf("game not found: %w", errs.ErrGameNotFound)
	}

	err = s.gameRepo.UpdateGame(ctx, game)
	if err != nil {
		return fmt.Errorf("failed to update game status: %w", err)
	}
	return nil
}
