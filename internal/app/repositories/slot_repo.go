package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"project2/internal/domain/entities"
	interfaces "project2/internal/domain/interfaces/repository"
	"time"
)

type slotRepo struct {
	db *sql.DB
}

func NewSlotRepo(db *sql.DB) interfaces.SlotRepository {
	return &slotRepo{
		db: db,
	}
}

// FetchSlotByID retrieves a slot by its ID.
func (r *slotRepo) FetchSlotByID(ctx context.Context, id uuid.UUID) (*entities.Slot, error) {
	query := `SELECT slot_id, game_id, slot_date, start_time, end_time, is_booked, created_at FROM slots WHERE slot_id = $1`
	row := r.db.QueryRowContext(ctx, query, id)

	var slot entities.Slot
	err := row.Scan(&slot.SlotID, &slot.GameID, &slot.Date, &slot.StartTime, &slot.EndTime, &slot.IsBooked, &slot.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // No slot found
		}
		return nil, fmt.Errorf("failed to fetch slot by ID: %w", err)
	}

	return &slot, nil
}

// CreateSlot inserts a new slot into the database and returns the created slot ID.
func (r *slotRepo) CreateSlot(ctx context.Context, slot *entities.Slot) (uuid.UUID, error) {
	query := `INSERT INTO slots (game_id, slot_date, start_time, end_time, is_booked) VALUES ($1, $2, $3, $4, $5) RETURNING slot_id`
	var id uuid.UUID
	err := r.db.QueryRowContext(ctx, query, slot.GameID, slot.Date, slot.StartTime, slot.EndTime, slot.IsBooked).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create slot: %w", err)
	}
	return id, nil
}

// DeleteSlotByID removes a slot from the database by its ID.
func (r *slotRepo) DeleteSlotByID(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM slots WHERE slot_id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete slot: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no slot found with ID %s", id)
	}

	return nil
}

// FetchSlotsByDate retrieves all slots for a specific date.
func (r *slotRepo) FetchSlotsByDate(ctx context.Context, date time.Time) ([]entities.Slot, error) {
	dateStr := date.Format("2006-01-02")
	query := `SELECT slot_id, game_id, slot_date, start_time, end_time, is_booked, created_at FROM slots WHERE slot_date::date = $1`
	rows, err := r.db.QueryContext(ctx, query, dateStr)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch slots by date: %w", err)
	}
	defer rows.Close()

	var slots []entities.Slot
	for rows.Next() {
		var slot entities.Slot
		if err := rows.Scan(&slot.SlotID, &slot.GameID, &slot.Date, &slot.StartTime, &slot.EndTime, &slot.IsBooked, &slot.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan slot row: %w", err)
		}
		slots = append(slots, slot)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("errs occurred while iterating over slots: %w", err)
	}

	return slots, nil
}

// FetchSlotByDateAndTime retrieves a slot by its date and start time.
func (r *slotRepo) FetchSlotByDateAndTime(ctx context.Context, date time.Time, startTime time.Time) (*entities.Slot, error) {
	dateStr := date.Format("2006-01-02")
	query := `SELECT slot_id, game_id, slot_date, start_time, end_time, is_booked, created_at FROM slots WHERE slot_date::date = $1 AND start_time = $2`
	row := r.db.QueryRowContext(ctx, query, dateStr, startTime)

	var slot entities.Slot
	err := row.Scan(&slot.SlotID, &slot.GameID, &slot.Date, &slot.StartTime, &slot.EndTime, &slot.IsBooked, &slot.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // No slot found
		}
		return nil, fmt.Errorf("failed to fetch slot by date and time: %w", err)
	}

	return &slot, nil
}

// FetchSlotsByGameID retrieves all slots associated with a specific game ID.
func (r *slotRepo) FetchSlotsByGameID(ctx context.Context, gameID uuid.UUID) ([]entities.Slot, error) {
	query := `SELECT slot_id, game_id, slot_date, start_time, end_time, is_booked, created_at FROM slots WHERE game_id = $1`
	rows, err := r.db.QueryContext(ctx, query, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch slots by game ID: %w", err)
	}
	defer rows.Close()

	var slots []entities.Slot
	for rows.Next() {
		var slot entities.Slot
		if err := rows.Scan(&slot.SlotID, &slot.GameID, &slot.Date, &slot.StartTime, &slot.EndTime, &slot.IsBooked, &slot.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan slot row: %w", err)
		}
		slots = append(slots, slot)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("errs occurred while iterating over slots: %w", err)
	}

	return slots, nil
}

// FetchSlotsByGameIDAndDate retrieves all slots for a specific game on a given date.
func (r *slotRepo) FetchSlotsByGameIDAndDate(ctx context.Context, gameID uuid.UUID, date time.Time) ([]entities.Slot, error) {
	// Convert the Go date to a string in the format YYYY-MM-DD for PostgresSQL comparison
	dateStr := date.Format("2006-01-02")

	query := `SELECT slot_id, game_id, slot_date, start_time, end_time, is_booked, created_at 
	          FROM slots 
	          WHERE game_id = $1 AND slot_date::date = $2`

	rows, err := r.db.QueryContext(ctx, query, gameID, dateStr)
	if err != nil {
		return nil, fmt.Errorf("failed to query slots by game ID and date: %w", err)
	}
	defer rows.Close()

	var slots []entities.Slot
	for rows.Next() {
		var slot entities.Slot
		err := rows.Scan(&slot.SlotID, &slot.GameID, &slot.Date, &slot.StartTime, &slot.EndTime, &slot.IsBooked, &slot.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan slot: %w", err)
		}
		slots = append(slots, slot)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration errs: %w", err)
	}

	return slots, nil
}

// UpdateSlotStatus updates the booking status of a specific slot.
func (r *slotRepo) UpdateSlotStatus(ctx context.Context, slotID uuid.UUID, isBooked bool) error {
	// Define the SQL query to update the is_booked status of the slot
	query := `UPDATE slots 
	          SET is_booked = $1 
	          WHERE slot_id = $2`

	// Execute the query with the provided slotID and isBooked status
	_, err := r.db.ExecContext(ctx, query, isBooked, slotID)
	if err != nil {
		return fmt.Errorf("failed to update slot status: %w", err)
	}

	return nil
}
