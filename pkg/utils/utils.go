package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"log"
	"math"
	"net/http"
	"project2/internal/config"
	"project2/internal/domain/entities"
	repository_interfaces "project2/internal/domain/interfaces/repository"
	"project2/pkg/errs"
	"project2/pkg/logger"
	"strings"
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

func GetIP(r *http.Request) string {
	// Check if the request came via a proxy (useful in production)
	// Most proxies add the client IP in the "X-Forwarded-For" header
	// or "X-Real-Ip"
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// "X-Forwarded-For" can contain multiple IP addresses,
		// take the first one which is the client IP
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Fallback to remote address if no proxy is involved
	ip := r.RemoteAddr

	// RemoteAddr can contain the IP and port (e.g., "192.168.1.1:12345")
	// Split it to get only the IP
	ip = strings.Split(ip, ":")[0]

	return ip
}

func InsertAllSlots(ctx context.Context, slotRepo repository_interfaces.SlotRepository, gameRepo repository_interfaces.GameRepository) error {

	location, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		log.Fatalf("Failed to load location: %v", err)
	}

	// Get the current time in the desired location
	nowInLocation := time.Now().In(location)

	// Set time to midnight for the current date
	today := time.Date(
		nowInLocation.Year(),
		nowInLocation.Month(),
		nowInLocation.Day(),
		0, 0, 0, 0,
		location,
	)
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

func JsonEncoder(w http.ResponseWriter, jsonResponse any) error {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(jsonResponse)
	if err != nil {
		logger.Logger.Errorw("Some unexpected error occurred while encoding the response", "error", err, "response", jsonResponse)
		errs.UnexpectedError("Some unexpected error occurred while encoding the response").ToJson2(w)
		return err
	}
	return nil
}
