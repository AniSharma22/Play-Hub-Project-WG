package repository_test

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"project2/internal/app/repositories"
	"project2/internal/domain/entities"
	"testing"
	"time"
)

func TestFetchSlotByID(t *testing.T) {
	db, mock := setup()
	defer db.Close()

	ctx := context.Background()
	slotRepo := repositories.NewSlotRepo(db)
	slotID := uuid.New()

	rows := sqlmock.NewRows([]string{"slot_id", "game_id", "slot_date", "start_time", "end_time", "is_booked", "created_at"}).
		AddRow(slotID, uuid.New(), time.Now(), time.Now(), time.Now().Add(20*time.Minute), false, time.Now())

	mock.ExpectQuery(`SELECT slot_id, game_id, slot_date, start_time, end_time, is_booked, created_at FROM slots WHERE slot_id = \$1`).
		WithArgs(slotID).
		WillReturnRows(rows)

	slot, err := slotRepo.FetchSlotByID(ctx, slotID)
	assert.NoError(t, err)
	assert.NotNil(t, slot)
	assert.Equal(t, slotID, slot.SlotID)
}

func TestCreateSlot(t *testing.T) {
	db, mock := setup()
	defer db.Close()

	ctx := context.TODO()
	slotRepo := repositories.NewSlotRepo(db)
	slot := &entities.Slot{
		GameID:    uuid.New(),
		Date:      time.Now(),
		StartTime: time.Now(),
		EndTime:   time.Now().Add(20 * time.Minute),
		IsBooked:  false,
	}

	slotID := uuid.New()
	mock.ExpectQuery(`INSERT INTO slots`).WithArgs(slot.GameID, slot.Date, slot.StartTime, slot.EndTime, slot.IsBooked).
		WillReturnRows(sqlmock.NewRows([]string{"slot_id"}).AddRow(slotID))

	id, err := slotRepo.CreateSlot(ctx, slot)
	assert.NoError(t, err)
	assert.Equal(t, slotID, id)
}

func TestDeleteSlotByID(t *testing.T) {
	db, mock := setup()
	defer db.Close()

	ctx := context.TODO()
	slotRepo := repositories.NewSlotRepo(db)
	slotID := uuid.New()

	mock.ExpectExec(`DELETE FROM slots WHERE slot_id = \$1`).
		WithArgs(slotID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := slotRepo.DeleteSlotByID(ctx, slotID)
	assert.NoError(t, err)
}

func TestFetchSlotsByDate(t *testing.T) {
	db, mock := setup()
	defer db.Close()

	ctx := context.TODO()
	slotRepo := repositories.NewSlotRepo(db)
	date := time.Now()

	rows := sqlmock.NewRows([]string{"slot_id", "game_id", "slot_date", "start_time", "end_time", "is_booked", "created_at"}).
		AddRow(uuid.New(), uuid.New(), date, date, date.Add(20*time.Minute), false, time.Now())

	mock.ExpectQuery(`SELECT slot_id, game_id, slot_date, start_time, end_time, is_booked, created_at FROM slots WHERE slot_date::date = \$1`).
		WithArgs(date.Format("2006-01-02")).
		WillReturnRows(rows)

	slots, err := slotRepo.FetchSlotsByDate(ctx, date)
	assert.NoError(t, err)
	assert.Len(t, slots, 1)
}

func TestFetchSlotByDateAndTime(t *testing.T) {
	db, mock := setup()
	defer db.Close()

	ctx := context.TODO()
	slotRepo := repositories.NewSlotRepo(db)
	date := time.Now()
	startTime := time.Now()

	rows := sqlmock.NewRows([]string{"slot_id", "game_id", "slot_date", "start_time", "end_time", "is_booked", "created_at"}).
		AddRow(uuid.New(), uuid.New(), date, startTime, startTime.Add(20*time.Minute), false, time.Now())

	mock.ExpectQuery(`SELECT slot_id, game_id, slot_date, start_time, end_time, is_booked, created_at FROM slots WHERE slot_date::date = \$1 AND start_time = \$2`).
		WithArgs(date.Format("2006-01-02"), startTime).
		WillReturnRows(rows)

	slot, err := slotRepo.FetchSlotByDateAndTime(ctx, date, startTime)
	assert.NoError(t, err)
	assert.NotNil(t, slot)
}

func TestFetchSlotsByGameID(t *testing.T) {
	db, mock := setup()
	defer db.Close()

	ctx := context.TODO()
	slotRepo := repositories.NewSlotRepo(db)
	gameID := uuid.New()

	rows := sqlmock.NewRows([]string{"slot_id", "game_id", "slot_date", "start_time", "end_time", "is_booked", "created_at"}).
		AddRow(uuid.New(), gameID, time.Now(), time.Now(), time.Now().Add(20*time.Minute), false, time.Now())

	mock.ExpectQuery(`SELECT slot_id, game_id, slot_date, start_time, end_time, is_booked, created_at FROM slots WHERE game_id = \$1`).
		WithArgs(gameID).
		WillReturnRows(rows)

	slots, err := slotRepo.FetchSlotsByGameID(ctx, gameID)
	assert.NoError(t, err)
	assert.Len(t, slots, 1)
}

func TestFetchSlotsByGameIDAndDate(t *testing.T) {
	db, mock := setup()
	defer db.Close()

	ctx := context.TODO()
	slotRepo := repositories.NewSlotRepo(db)
	gameID := uuid.New()
	date := time.Now()

	rows := sqlmock.NewRows([]string{"slot_id", "game_id", "slot_date", "start_time", "end_time", "is_booked", "created_at"}).
		AddRow(uuid.New(), gameID, date, date, date.Add(20*time.Minute), false, time.Now())

	mock.ExpectQuery(`SELECT slot_id, game_id, slot_date, start_time, end_time, is_booked, created_at 
		FROM slots WHERE game_id = \$1 AND slot_date::date = \$2`).
		WithArgs(gameID, date.Format("2006-01-02")).
		WillReturnRows(rows)

	slots, err := slotRepo.FetchSlotsByGameIDAndDate(ctx, gameID, date)
	assert.NoError(t, err)
	assert.Len(t, slots, 1)
}

func TestUpdateSlotStatus(t *testing.T) {
	db, mock := setup()
	defer db.Close()

	ctx := context.TODO()
	slotRepo := repositories.NewSlotRepo(db)
	slotID := uuid.New()

	mock.ExpectExec(`UPDATE slots SET is_booked = \$1 WHERE slot_id = \$2`).
		WithArgs(true, slotID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := slotRepo.UpdateSlotStatus(ctx, slotID, true)
	assert.NoError(t, err)
}
