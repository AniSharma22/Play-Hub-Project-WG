package service_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"project2/internal/domain/entities"
	"testing"
)

func TestGameService_GetGameByID(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	gameID := uuid.New()
	expectedGame := &entities.Game{GameID: gameID}

	mockGameRepo.EXPECT().
		FetchGameByID(gomock.Any(), gameID).
		Return(expectedGame, nil).
		Times(1)

	game, err := gameService.GetGameByID(context.Background(), gameID)

	assert.NoError(t, err)
	assert.Equal(t, expectedGame, game)
}

func TestGameService_GetAllGames(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	expectedGames := []entities.Game{
		{GameID: uuid.New()},
		{GameID: uuid.New()},
	}

	mockGameRepo.EXPECT().
		FetchAllGames(gomock.Any()).
		Return(expectedGames, nil).
		Times(1)

	games, err := gameService.GetAllGames(context.Background())

	assert.NoError(t, err)
	assert.ElementsMatch(t, expectedGames, games)
}

func TestGameService_DeleteGame(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	gameID := uuid.New()

	mockGameRepo.EXPECT().
		DeleteGame(gomock.Any(), gameID).
		Return(nil).
		Times(1)

	err := gameService.DeleteGame(context.Background(), gameID)

	assert.NoError(t, err)
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

func TestGameService_UpdateGameStatus(t *testing.T) {
	tests := []struct {
		name          string
		gameID        uuid.UUID
		status        bool
		mockSetup     func()
		expectedError error
	}{
		{
			name:   "Success",
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
			expectedError: nil,
		},
		{
			name:   "GameNotFound",
			gameID: uuid.New(),
			status: true,
			mockSetup: func() {
				mockGameRepo.EXPECT().
					FetchGameByID(gomock.Any(), gomock.Any()).
					Return(nil, nil).
					Times(1)
			},
			expectedError: errors.New("game not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teardown := setup(t)
			defer teardown()

			tt.mockSetup()

			err := gameService.UpdateGameStatus(context.Background(), tt.gameID, tt.status)

			assert.Equal(t, tt.expectedError, err)
		})
	}
}
