package models

import (
	"github.com/google/uuid"
	"time"
)

type Invitations struct {
	InvitationId uuid.UUID   `json:"invitation_id"`
	SlotId       uuid.UUID   `json:"slot_id"`
	GameId       uuid.UUID   `json:"game_id"`
	GameName     string      `json:"game"`
	ImageUrl     string      `json:"image_url"`
	Date         time.Time   `json:"date"`
	StartTime    time.Time   `json:"start_time"`
	EndTime      time.Time   `json:"end_time"`
	BookedUsers  []BasicUser `json:"booked_users"`
	InvitedBy    string      `json:"invited_by"`
}

type Bookings struct {
	BookingId   uuid.UUID `json:"booking_id"`
	GameId      uuid.UUID `json:"game_id"`
	GameName    string    `json:"game"`
	ImageUrl    string    `json:"image_url"`
	Date        time.Time `json:"date"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	BookedUsers []string  `json:"booked_users"`
}

type SlotDTO struct {
	SlotID      uuid.UUID `json:"slot_id"`
	GameID      uuid.UUID `json:"game_id"`
	Date        time.Time `json:"slot_date"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	IsBooked    bool      `json:"is_booked"`
	BookedUsers []string  `json:"booked_users"`
	CreatedAt   time.Time `json:"created_at"`
}
type LeaderboardDTO struct {
	UserName   string  `json:"user_name"`
	TotalGames int     `json:"total_games"`
	Wins       int     `json:"wins"`
	Losses     int     `json:"losses"`
	Score      float64 `json:"score"`
}

type BasicUser struct {
	UserName  string `json:"user_name"`
	UserImage string `json:"user_image"`
}
