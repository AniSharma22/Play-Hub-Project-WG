package utils

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"log"
	"math"
	"project2/internal/config"
	"project2/internal/domain/entities"
	repository_interfaces "project2/internal/domain/interfaces/repository"
	"time"
)

func GetTotalScore(totalWins, totalLosses int) float32 {
	totalGames := totalWins + totalLosses
	return calculateScore(totalWins, totalLosses, totalGames)
}

func calculateScore(totalWins, totalLosses, totalGames int) float32 {
	var winLossRatio float32
	if totalLosses == 0 {
		winLossRatio = float32(totalWins)
	} else {
		winLossRatio = float32(totalWins) / float32(totalLosses)
	}

	var gameFactor float32 = float32(1) + float32(math.Sqrt(float64(totalGames)))
	return (winLossRatio * gameFactor) / 100
}

func GetNameFromEmail(email string) string {
	var name bytes.Buffer
	for i := 0; i < len(email); i++ {
		if email[i] == '.' {
			name.WriteByte(' ')
		} else if email[i] == '@' {
			break
		} else {
			name.WriteByte(email[i])
		}
	}
	return name.String()
}

func CreateJwtToken(userId uuid.UUID, role string) (string, error) {
	// Create JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": userId.String(),
		"role":   role,
		"exp":    time.Now().Add(time.Minute * 5).Unix(), // Token expiry time (5 minute)
	})

	// Sign the token with the secret key
	tokenString, err := token.SignedString(config.MY_SIGNING_KEY)
	if err != nil {
		//logger.Logger.Errorw("Error signing token", "method", r.Method, "error", err, "time", time.Now())
		return "", errors.New("error creating jwt token")
	}

	return tokenString, nil
}

func InsertAllSlots(ctx context.Context, slotRepo repository_interfaces.SlotRepository, gameRepo repository_interfaces.GameRepository) error {
	today := time.Now().Truncate(24 * time.Hour)
	location, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		log.Fatalf("Failed to load location: %v", err)
	}

	// Fetch all games
	games, err := gameRepo.FetchAllGames(ctx)
	if err != nil {
		return fmt.Errorf("errs fetching games: %w", err)
	}

	now := time.Now().In(location)
	startTime := time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, location)
	endTime := time.Date(now.Year(), now.Month(), now.Day(), 18, 0, 0, 0, location)

	for _, game := range games {
		// Check for existing slots for this game on today's date
		existingSlots, err := slotRepo.FetchSlotsByGameIDAndDate(ctx, game.GameID, today)
		if err != nil {
			return fmt.Errorf("errs checking existing slots for game %s: %w", game.GameName, err)
		}

		// If no slots exist, create new slots
		if len(existingSlots) == 0 {
			for current := startTime; current.Before(endTime); current = current.Add(20 * time.Minute) {
				slotEndTime := current.Add(20 * time.Minute)
				if slotEndTime.After(endTime) {
					slotEndTime = endTime
				}

				newSlot := &entities.Slot{
					SlotID:    uuid.New(),
					GameID:    game.GameID,
					Date:      today,
					StartTime: current,
					EndTime:   slotEndTime,
					IsBooked:  false,
				}

				// Insert the new slot
				if _, err := slotRepo.CreateSlot(ctx, newSlot); err != nil {
					return fmt.Errorf("errs inserting slot for game %s: %w", game.GameName, err)
				}
			}
		}
	}
	return nil
}
