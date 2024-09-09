package repository_test

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"project2/internal/app/repositories"
	"project2/internal/domain/entities"
	"testing"
	"time"
)

func TestGameRepo_FetchGameByID(t *testing.T) {
	db, mock := setup()
	defer db.Close()

	// Test Data
	gameID := uuid.New()
	expectedGame := &entities.Game{
		GameID:     gameID,
		GameName:   "Test Game",
		MinPlayers: 2,
		MaxPlayers: 4,
		Instances:  1,
		IsActive:   true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Success case
	mock.ExpectQuery(`SELECT game_id, game_name, min_players, max_players, instances, is_active, created_at, updated_at FROM games WHERE game_id = \$1`).
		WithArgs(gameID).
		WillReturnRows(sqlmock.NewRows([]string{"game_id", "game_name", "min_players", "max_players", "instances", "is_active", "created_at", "updated_at"}).
			AddRow(expectedGame.GameID, expectedGame.GameName, expectedGame.MinPlayers, expectedGame.MaxPlayers, expectedGame.Instances, expectedGame.IsActive, expectedGame.CreatedAt, expectedGame.UpdatedAt))

	repo := repositories.NewGameRepo(db)
	ctx := context.Background()
	game, err := repo.FetchGameByID(ctx, gameID)

	assert.NoError(t, err)
	assert.Equal(t, expectedGame, game)
	assert.NoError(t, mock.ExpectationsWereMet())

	// No rows case
	mock.ExpectQuery(`SELECT game_id, game_name, min_players, max_players, instances, is_active, created_at, updated_at FROM games WHERE game_id = \$1`).
		WithArgs(gameID).
		WillReturnError(sql.ErrNoRows)

	game, err = repo.FetchGameByID(ctx, gameID)
	assert.NoError(t, err)
	assert.Nil(t, game)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGameRepo_FetchAllGames(t *testing.T) {
	db, mock := setup()
	defer db.Close()

	expectedGames := []entities.Game{
		{
			GameID:     uuid.New(),
			GameName:   "Game 1",
			MinPlayers: 2,
			MaxPlayers: 4,
			Instances:  1,
			IsActive:   true,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
		{
			GameID:     uuid.New(),
			GameName:   "Game 2",
			MinPlayers: 2,
			MaxPlayers: 6,
			Instances:  2,
			IsActive:   false,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
	}

	mock.ExpectQuery(`SELECT game_id, game_name, min_players, max_players, instances, is_active, created_at, updated_at FROM games`).
		WillReturnRows(sqlmock.NewRows([]string{"game_id", "game_name", "min_players", "max_players", "instances", "is_active", "created_at", "updated_at"}).
			AddRow(expectedGames[0].GameID, expectedGames[0].GameName, expectedGames[0].MinPlayers, expectedGames[0].MaxPlayers, expectedGames[0].Instances, expectedGames[0].IsActive, expectedGames[0].CreatedAt, expectedGames[0].UpdatedAt).
			AddRow(expectedGames[1].GameID, expectedGames[1].GameName, expectedGames[1].MinPlayers, expectedGames[1].MaxPlayers, expectedGames[1].Instances, expectedGames[1].IsActive, expectedGames[1].CreatedAt, expectedGames[1].UpdatedAt))

	repo := repositories.NewGameRepo(db)
	ctx := context.Background()
	games, err := repo.FetchAllGames(ctx)

	assert.NoError(t, err)
	assert.Equal(t, expectedGames, games)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGameRepo_CreateGame(t *testing.T) {
	db, mock := setup()
	defer db.Close()

	game := &entities.Game{
		GameName:   "New Game",
		MinPlayers: 2,
		MaxPlayers: 4,
		Instances:  1,
		IsActive:   true,
	}

	gameID := uuid.New()

	mock.ExpectQuery(`INSERT INTO games \(game_name, min_players, max_players, instances, is_active\) VALUES \(\$1, \$2, \$3, \$4, \$5\) RETURNING game_id`).
		WithArgs(game.GameName, game.MinPlayers, game.MaxPlayers, game.Instances, game.IsActive).
		WillReturnRows(sqlmock.NewRows([]string{"game_id"}).AddRow(gameID))

	repo := repositories.NewGameRepo(db)
	ctx := context.Background()
	id, err := repo.CreateGame(ctx, game)

	assert.NoError(t, err)
	assert.Equal(t, gameID, id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGameRepo_DeleteGame(t *testing.T) {
	db, mock := setup()
	defer db.Close()

	gameID := uuid.New()

	// Success case
	mock.ExpectExec(`DELETE FROM games WHERE game_id = \$1`).
		WithArgs(gameID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	repo := repositories.NewGameRepo(db)
	ctx := context.Background()
	err := repo.DeleteGame(ctx, gameID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	// No rows affected case
	mock.ExpectExec(`DELETE FROM games WHERE game_id = \$1`).
		WithArgs(gameID).
		WillReturnResult(sqlmock.NewResult(1, 0))

	mock.
		err = repo.DeleteGame(ctx, gameID)
	assert.Error(t, err)
	assert.Equal(t, fmt.Errorf("no game found with ID %s", gameID), err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGameRepo_UpdateGameStatus(t *testing.T) {
	db, mock := setup()
	defer db.Close()

	gameID := uuid.New()

	mock.ExpectExec(`UPDATE games SET is_active = \$1, updated_at = CURRENT_TIMESTAMP WHERE game_id = \$2`).
		WithArgs(true, gameID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	repo := repositories.NewGameRepo(db)
	ctx := context.Background()
	err := repo.UpdateGameStatus(ctx, gameID, true)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
