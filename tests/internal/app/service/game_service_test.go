package service_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"project2/internal/domain/entities"
	"testing"
)

func TestGameService_GetGameByID(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	gameID := uuid.New()
	expectedGame := &entities.Game{GameID: gameID}

	t.Run("Successful retrieval of game", func(t *testing.T) {
		mockGameRepo.EXPECT().
			FetchGameByID(gomock.Any(), gameID).
			Return(expectedGame, nil).
			Times(1)

		game, err := gameService.GetGameByID(context.Background(), gameID)

		assert.NoError(t, err)
		assert.Equal(t, expectedGame, game)
	})

	t.Run("Error fetching game", func(t *testing.T) {
		mockGameRepo.EXPECT().
			FetchGameByID(gomock.Any(), gameID).
			Return(nil, errors.New("error")).
			Times(1)

		game, err := gameService.GetGameByID(context.Background(), gameID)
		assert.Nil(t, game)
		assert.Error(t, err)
	})

}

func TestGameService_GetAllGames(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	expectedGames := []entities.Game{
		{GameID: uuid.New()},
		{GameID: uuid.New()},
	}

	t.Run("Successful retrieval of games", func(t *testing.T) {
		mockGameRepo.EXPECT().
			FetchAllGames(gomock.Any()).
			Return(expectedGames, nil).
			Times(1)

		games, err := gameService.GetAllGames(context.Background())

		assert.NoError(t, err)
		assert.ElementsMatch(t, expectedGames, games)
	})

	t.Run("Error fetching games", func(t *testing.T) {
		mockGameRepo.EXPECT().
			FetchAllGames(gomock.Any()).
			Return(nil, errors.New("error")).
			Times(1)

		games, err := gameService.GetAllGames(context.Background())
		assert.Nil(t, games)
		assert.Error(t, err)
	})
}

func TestGameService_DeleteGame(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	gameID := uuid.New()

	t.Run("Successful deletion of game", func(t *testing.T) {
		mockGameRepo.EXPECT().
			DeleteGame(gomock.Any(), gameID).
			Return(nil).
			Times(1)

		err := gameService.DeleteGame(context.Background(), gameID)

		assert.NoError(t, err)
	})

	t.Run("Error deleting game", func(t *testing.T) {
		mockGameRepo.EXPECT().
			DeleteGame(gomock.Any(), gameID).
			Return(errors.New("error")).
			Times(1)

		err := gameService.DeleteGame(context.Background(), gameID)
		assert.Error(t, err)
	})
}

func TestGameService_CreateGame(t *testing.T) {
	mockID := uuid.New()
	tests := []struct {
		name          string
		game          *entities.Game
		mockSetup     func()
		expectedID    uuid.UUID
		expectedError error
	}{
		{
			name: "Success",
			game: &entities.Game{},
			mockSetup: func() {
				mockGameRepo.EXPECT().
					CreateGame(gomock.Any(), gomock.Any()).
					Return(mockID, nil).
					Times(1)
			},
			expectedID:    mockID,
			expectedError: nil,
		},
		{
			name: "Error",
			game: &entities.Game{},
			mockSetup: func() {
				mockGameRepo.EXPECT().
					CreateGame(gomock.Any(), gomock.Any()).
					Return(uuid.Nil, errors.New("creation error")).
					Times(1)
			},
			expectedID:    uuid.Nil,
			expectedError: fmt.Errorf("failed to create game: creation error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teardown := setup(t)
			defer teardown()

			tt.mockSetup()

			id, _ := gameService.CreateGame(context.Background(), tt.game)

			assert.Equal(t, tt.expectedID, id)
		})
	}
}

func TestUpdateGameStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name          string
		gameID        uuid.UUID
		status        bool
		mockSetup     func()
		expectedError string
	}{
		{
			name:   "Successful Update",
			gameID: uuid.New(),
			status: true,
			mockSetup: func() {
				mockGameRepo.EXPECT().
					FetchGameByID(gomock.Any(), gomock.Any()).
					Return(&entities.Game{GameID: uuid.New()}, nil).
					Times(1)

				mockGameRepo.EXPECT().
					UpdateGameStatus(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil).
					Times(1)
			},
			expectedError: "",
		},
		{
			name:   "Failed to Fetch Game by ID",
			gameID: uuid.New(),
			status: true,
			mockSetup: func() {
				mockGameRepo.EXPECT().
					FetchGameByID(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("db error")).
					Times(1)
			},
			expectedError: "failed to fetch game by ID: db error",
		},
		{
			name:   "Game Not Found",
			gameID: uuid.New(),
			status: true,
			mockSetup: func() {
				mockGameRepo.EXPECT().
					FetchGameByID(gomock.Any(), gomock.Any()).
					Return(nil, nil).
					Times(1)
			},
			expectedError: "game not found",
		},
		{
			name:   "Failed to Update Game Status",
			gameID: uuid.New(),
			status: true,
			mockSetup: func() {
				mockGameRepo.EXPECT().
					FetchGameByID(gomock.Any(), gomock.Any()).
					Return(&entities.Game{GameID: uuid.New()}, nil).
					Times(1)

				mockGameRepo.EXPECT().
					UpdateGameStatus(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("update error")).
					Times(1)
			},
			expectedError: "failed to update game status: update error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			err := gameService.UpdateGameStatus(context.Background(), tt.gameID, tt.status)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
