package models

import (
	"github.com/google/uuid"
	"time"
)

type Invitations struct {
	InvitationId uuid.UUID `json:"invitation_id"`
	SlotId       uuid.UUID `json:"slot_id"`
	GameId       uuid.UUID `json:"game_id"`
	GameName     string    `json:"game"`
	Date         time.Time `json:"date"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	BookedUsers  []string  `json:"booked_users"`
	InvitedBy    string    `json:"invited_by"`
}

type Bookings struct {
	BookingId   uuid.UUID `json:"booking_id"`
	GameId      uuid.UUID `json:"game_id"`
	GameName    string    `json:"game"`
	Date        time.Time `json:"date"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	BookedUsers []string  `json:"booked_users"`
}

type Leaderboard struct {
	UserName string  `json:"user_name"`
	Score    float64 `json:"score"`
}
