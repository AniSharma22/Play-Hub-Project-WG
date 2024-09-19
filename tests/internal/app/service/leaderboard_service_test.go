package service_test

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"project2/internal/domain/entities"
	"project2/internal/models"
	"project2/pkg/utils"
	"testing"
)

func TestLeaderboardService_GetGameLeaderboard(t *testing.T) {
	gameId := uuid.New()
	leaderboard := []models.Leaderboard{
		{UserName: "username", Score: 0},
	}
	ctx := context.TODO()

	tests := []struct {
		name                string
		mockSetup           func()
		expectedError       bool
		expectedLeaderboard []models.Leaderboard
	}{
		{
			name: "Successful Leaderboard Retrieval",
			mockSetup: func() {
				mockLeaderboardRepo.EXPECT().FetchGameLeaderboard(ctx, gameId).Return(leaderboard, nil)
			},
			expectedError:       false,
			expectedLeaderboard: leaderboard,
		},
		{
			name: "Failed Leaderboard Retrieval",
			mockSetup: func() {
				mockLeaderboardRepo.EXPECT().FetchGameLeaderboard(ctx, gameId).Return(nil, errors.New("errs"))
			},
			expectedError:       true,
			expectedLeaderboard: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teardown := setup(t)
			defer teardown()

			tt.mockSetup()

			result, err := leaderboardService.GetGameLeaderboard(ctx, gameId)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedLeaderboard, result)
		})
	}
}

func TestLeaderboardService_AddWinToUser(t *testing.T) {
	ctx := context.TODO()
	userID := uuid.New()
	gameID := uuid.New()
	bookingID := uuid.New()

	mockUserStats := &entities.Leaderboard{
		UserID: userID,
		GameID: gameID,
		Wins:   5,
		Losses: 2,
		Score:  100.0,
	}

	tests := []struct {
		name          string
		mockSetup     func()
		expectedError bool
	}{
		{
			name: "Successfully add win to user and update booking",
			mockSetup: func() {
				mockLeaderboardRepo.EXPECT().
					FetchUserGameStats(ctx, userID, gameID).
					Return(mockUserStats, nil)

				// Increment the wins and update score
				updatedStats := *mockUserStats
				updatedStats.Wins++
				updatedStats.Score = float64(utils.GetTotalScore(updatedStats.Wins, updatedStats.Losses))

				mockLeaderboardRepo.EXPECT().
					UpdateUserGameStats(ctx, &updatedStats).
					Return(nil)

				mockBookingService.EXPECT().
					UpdateBookingResult(ctx, bookingID, "win").
					Return(nil)
			},
			expectedError: false,
		},
		{
			name: "Fail to fetch user game stats",
			mockSetup: func() {
				mockLeaderboardRepo.EXPECT().
					FetchUserGameStats(ctx, userID, gameID).
					Return(nil, errors.New("database errs"))
			},
			expectedError: true,
		},
		{
			name: "Fail to update user game stats",
			mockSetup: func() {
				mockLeaderboardRepo.EXPECT().
					FetchUserGameStats(ctx, userID, gameID).
					Return(mockUserStats, nil)

				// Simulate failure while updating user stats
				mockLeaderboardRepo.EXPECT().
					UpdateUserGameStats(ctx, mockUserStats).
					Return(errors.New("update errs"))
			},
			expectedError: true,
		},
		{
			name: "Fail to update booking result",
			mockSetup: func() {
				mockLeaderboardRepo.EXPECT().
					FetchUserGameStats(ctx, userID, gameID).
					Return(mockUserStats, nil)

				updatedStats := *mockUserStats
				updatedStats.Wins++
				updatedStats.Score = float64(utils.GetTotalScore(updatedStats.Wins, updatedStats.Losses))

				mockLeaderboardRepo.EXPECT().
					UpdateUserGameStats(ctx, &updatedStats).
					Return(nil)

				// Simulate failure while updating booking result
				mockBookingService.EXPECT().
					UpdateBookingResult(ctx, bookingID, "win").
					Return(errors.New("booking update errs"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teardown := setup(t)
			defer teardown()

			tt.mockSetup()

			err := leaderboardService.AddWinToUser(ctx, userID, gameID, bookingID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLeaderboardService_AddLossToUser(t *testing.T) {
	ctx := context.TODO()
	userID := uuid.New()
	gameID := uuid.New()
	bookingID := uuid.New()

	mockUserStats := &entities.Leaderboard{
		UserID: userID,
		GameID: gameID,
		Wins:   5,
		Losses: 2,
		Score:  100.0,
	}

	tests := []struct {
		name          string
		mockSetup     func()
		expectedError bool
	}{
		{
			name: "Successfully add loss to user and update booking",
			mockSetup: func() {
				mockLeaderboardRepo.EXPECT().
					FetchUserGameStats(ctx, userID, gameID).
					Return(mockUserStats, nil)

				// Increment the losses and update score
				updatedStats := *mockUserStats
				updatedStats.Losses++
				updatedStats.Score = float64(utils.GetTotalScore(updatedStats.Wins, updatedStats.Losses))

				mockLeaderboardRepo.EXPECT().
					UpdateUserGameStats(ctx, &updatedStats).
					Return(nil)

				mockBookingService.EXPECT().
					UpdateBookingResult(ctx, bookingID, "loss").
					Return(nil)
			},
			expectedError: false,
		},
		{
			name: "Fail to fetch user game stats",
			mockSetup: func() {
				mockLeaderboardRepo.EXPECT().
					FetchUserGameStats(ctx, userID, gameID).
					Return(nil, errors.New("database errs"))
			},
			expectedError: true,
		},
		{
			name: "Fail to update user game stats",
			mockSetup: func() {
				mockLeaderboardRepo.EXPECT().
					FetchUserGameStats(ctx, userID, gameID).
					Return(mockUserStats, nil)

				// Simulate failure while updating user stats
				mockLeaderboardRepo.EXPECT().
					UpdateUserGameStats(ctx, mockUserStats).
					Return(errors.New("update errs"))
			},
			expectedError: true,
		},
		{
			name: "Fail to update booking result",
			mockSetup: func() {
				mockLeaderboardRepo.EXPECT().
					FetchUserGameStats(ctx, userID, gameID).
					Return(mockUserStats, nil)

				updatedStats := *mockUserStats
				updatedStats.Losses++
				updatedStats.Score = float64(utils.GetTotalScore(updatedStats.Wins, updatedStats.Losses))

				mockLeaderboardRepo.EXPECT().
					UpdateUserGameStats(ctx, &updatedStats).
					Return(nil)

				// Simulate failure while updating booking result
				mockBookingService.EXPECT().
					UpdateBookingResult(ctx, bookingID, "loss").
					Return(errors.New("booking update errs"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teardown := setup(t)
			defer teardown()

			tt.mockSetup()

			err := leaderboardService.AddLossToUser(ctx, userID, gameID, bookingID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
