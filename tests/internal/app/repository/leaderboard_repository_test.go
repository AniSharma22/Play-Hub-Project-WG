package repository_test

import (
	"context"
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"project2/internal/app/repositories"
	"project2/internal/domain/entities"
	"testing"
	"time"
)

func TestFetchGameLeaderboard(t *testing.T) {
	db, mock := setup() // Initialize mock DB and setup
	defer db.Close()
	repo := repositories.NewLeaderboardRepo(db)

	gameID := uuid.New()

	// Mock SQL rows for leaderboard entries
	rows := sqlmock.NewRows([]string{"username", "score"}).
		AddRow("john_doe", 100).
		AddRow("jane_smith", 90)

	mock.ExpectQuery("SELECT u.username, l.score FROM leaderboard l INNER JOIN users u ON l.user_id = u.user_id WHERE l.game_id =").
		WithArgs(gameID).
		WillReturnRows(rows)

	// Execute the method
	leaderboard, err := repo.FetchGameLeaderboard(context.TODO(), gameID)

	// Assertions
	assert.NoError(t, err)
	assert.Len(t, leaderboard, 2)
	assert.Equal(t, "john_doe", leaderboard[0].UserName)
	assert.Equal(t, float64(100), leaderboard[0].Score)
}

func TestFetchUserGameStats(t *testing.T) {
	db, mock := setup()
	defer db.Close()
	repo := repositories.NewLeaderboardRepo(db)

	userID := uuid.New()
	gameID := uuid.New()

	// Mock row for user game stats
	row := sqlmock.NewRows([]string{"score_id", "user_id", "game_id", "wins", "losses", "score", "created_at"}).
		AddRow(uuid.New(), userID, gameID, 5, 2, 200, time.Now())

	mock.ExpectQuery("SELECT score_id, user_id, game_id, wins, losses, score, created_at FROM leaderboard WHERE user_id =").
		WithArgs(userID, gameID).
		WillReturnRows(row)

	// Execute the method
	stats, err := repo.FetchUserGameStats(context.TODO(), userID, gameID)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, userID, stats.UserID)
	assert.Equal(t, 5, stats.Wins)
	assert.Equal(t, float64(200), stats.Score)
}

func TestFetchUserGameStats_NoRows(t *testing.T) {
	db, mock := setup()
	defer db.Close()
	repo := repositories.NewLeaderboardRepo(db)

	userID := uuid.New()
	gameID := uuid.New()

	// No rows found
	mock.ExpectQuery("SELECT score_id, user_id, game_id, wins, losses, score, created_at FROM leaderboard WHERE user_id =").
		WithArgs(userID, gameID).
		WillReturnError(sql.ErrNoRows)

	// Execute the method
	stats, err := repo.FetchUserGameStats(context.TODO(), userID, gameID)

	// Assertions
	assert.NoError(t, err)
	assert.Nil(t, stats)
}

func TestFetchUserOverallStats(t *testing.T) {
	db, mock := setup()
	defer db.Close()
	repo := repositories.NewLeaderboardRepo(db)

	userID := uuid.New()

	// Mock SQL rows for user stats across all games
	rows := sqlmock.NewRows([]string{"score_id", "user_id", "game_id", "wins", "losses", "score", "created_at"}).
		AddRow(uuid.New(), userID, uuid.New(), 10, 3, 300, time.Now()).
		AddRow(uuid.New(), userID, uuid.New(), 7, 2, 150, time.Now())

	mock.ExpectQuery("SELECT score_id, user_id, game_id, wins, losses, score, created_at FROM leaderboard WHERE user_id =").
		WithArgs(userID).
		WillReturnRows(rows)

	// Execute the method
	stats, err := repo.FetchUserOverallStats(context.TODO(), userID)

	// Assertions
	assert.NoError(t, err)
	assert.Len(t, stats, 2)
	assert.Equal(t, 10, stats[0].Wins)
	assert.Equal(t, float64(300), stats[0].Score)
}

func TestUpdateUserGameStats_ExistingEntry(t *testing.T) {
	db, mock := setup()
	defer db.Close()
	repo := repositories.NewLeaderboardRepo(db)

	leaderboard := &entities.Leaderboard{
		ScoreID: uuid.New(),
		UserID:  uuid.New(),
		GameID:  uuid.New(),
		Wins:    10,
		Losses:  5,
		Score:   300,
	}

	// Mock check for existing entry
	mock.ExpectQuery("SELECT EXISTS").WithArgs(leaderboard.ScoreID).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Mock update query
	mock.ExpectExec("UPDATE leaderboard SET wins =").WithArgs(leaderboard.Wins, leaderboard.Losses, leaderboard.Score, leaderboard.ScoreID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Execute the method
	err := repo.UpdateUserGameStats(context.TODO(), leaderboard)

	// Assertions
	assert.NoError(t, err)
}

func TestUpdateUserGameStats_NewEntry(t *testing.T) {
	db, mock := setup()
	defer db.Close()
	repo := repositories.NewLeaderboardRepo(db)

	leaderboard := &entities.Leaderboard{
		ScoreID: uuid.New(),
		UserID:  uuid.New(),
		GameID:  uuid.New(),
		Wins:    8,
		Losses:  3,
		Score:   250,
	}

	// Mock check for non-existing entry
	mock.ExpectQuery("SELECT EXISTS").WithArgs(leaderboard.ScoreID).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	// Mock insert query
	mock.ExpectExec("INSERT INTO leaderboard").WithArgs(leaderboard.ScoreID, leaderboard.UserID, leaderboard.GameID, leaderboard.Wins, leaderboard.Losses, leaderboard.Score).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Execute the method
	err := repo.UpdateUserGameStats(context.TODO(), leaderboard)

	// Assertions
	assert.NoError(t, err)
}
