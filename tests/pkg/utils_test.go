package pkg_test

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"project2/internal/domain/entities"
	"project2/pkg/utils"
	"project2/pkg/validation"
	mocks "project2/tests/mocks/repository"
	"testing"
	"time"
)

func TestGetHashedPassword(t *testing.T) {
	password := "Password123!"

	hash, err := utils.GetHashedPassword([]byte(password))
	if err != nil {
		t.Errorf("GetHashedPassword returned an error: %v", err)
	}

	if len(hash) == 0 {
		t.Error("Expected non-empty hash, got empty string")
	}
}

func TestVerifyPassword(t *testing.T) {
	password := "Password123!"
	wrongPassword := "WrongPassword!"

	hash, err := utils.GetHashedPassword([]byte(password))
	if err != nil {
		t.Fatalf("GetHashedPassword returned an error: %v", err)
	}

	if !utils.VerifyPassword([]byte(password), hash) {
		t.Error("VerifyPassword returned false for correct password")
	}

	if utils.VerifyPassword([]byte(wrongPassword), hash) {
		t.Error("VerifyPassword returned true for incorrect password")
	}
}

func TestIsValidPassword(t *testing.T) {
	tests := []struct {
		password string
		expected bool
	}{
		{"Password1!", true},
		{"Pass!", false},      // Less than 8 characters
		{"password1!", false}, // No uppercase letter
		{"PASSWORD1!", false}, // No lowercase letter
		{"Password1", false},  // No special character
		{"Password!@", true},  // Two special characters
	}

	for _, test := range tests {
		if result := validation.IsValidPassword(test.password); result != test.expected {
			t.Errorf("IsValidPassword(%s) = %v; want %v", test.password, result, test.expected)
		}
	}
}

func TestGetNameFromEmail(t *testing.T) {
	tests := []struct {
		email    string
		expected string
	}{
		{"john.doe@example.com", "john doe"},
		{"jane.smith@domain.com", "jane smith"},
		{"user@domain.com", "user"},
	}

	for _, test := range tests {
		result := utils.GetNameFromEmail(test.email)
		if result != test.expected {
			t.Errorf("GetNameFromEmail(%s) = %v; want %v", test.email, result, test.expected)
		}
	}
}

func TestGetTotalScore(t *testing.T) {
	score := utils.GetTotalScore(1, 0)
	expectedScore := float32(0.02)
	assert.Equal(t, expectedScore, score, "The total score should be 0.02")
}

func TestInsertAllSlots(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock repositories
	mockSlotRepo := mocks.NewMockSlotRepository(ctrl)
	mockGameRepo := mocks.NewMockGameRepository(ctrl)

	// Setup test data
	gameID := uuid.New()
	today := time.Now().Truncate(24 * time.Hour)
	games := []entities.Game{
		{GameID: gameID, GameName: "Table Tennis"},
	}

	// Mock FetchAllGames to return test games
	mockGameRepo.EXPECT().
		FetchAllGames(gomock.Any()).
		Return(games, nil).
		Times(1)

	// Mock FetchSlotsByGameIDAndDate to return empty slots (no slots exist)
	mockSlotRepo.EXPECT().
		FetchSlotsByGameIDAndDate(gomock.Any(), gameID, today).
		Return([]entities.Slot{}, nil).
		Times(1)

	// Calculate the number of expected slot creations
	location, err := time.LoadLocation("Asia/Kolkata")
	now := time.Now().In(time.UTC) // Use UTC for consistency
	startTime := time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, location)
	endTime := time.Date(now.Year(), now.Month(), now.Day(), 18, 0, 0, 0, location)
	expectedSlotCount := 0
	for current := startTime; current.Before(endTime); current = current.Add(20 * time.Minute) {
		expectedSlotCount++
	}

	// Mock CreateSlot to return the ID of the created slot
	mockSlotRepo.EXPECT().
		CreateSlot(gomock.Any(), gomock.Any()).
		Return(uuid.New(), nil).
		Times(expectedSlotCount) // Expect the number of slots created

	// Call the function to test
	err = utils.InsertAllSlots(context.Background(), mockSlotRepo, mockGameRepo)
	require.NoError(t, err)
}
