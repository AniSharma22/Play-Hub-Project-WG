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

func TestCreateBooking(t *testing.T) {
	db, mock := setup()
	defer db.Close()
	repo := repositories.NewBookingRepo(db)

	booking := &entities.Booking{
		BookingID: uuid.New(),
		SlotID:    uuid.New(),
		UserID:    uuid.New(),
	}

	query := "INSERT INTO bookings"
	mock.ExpectQuery(query).
		WithArgs(booking.SlotID, booking.UserID).
		WillReturnRows(sqlmock.NewRows([]string{"booking_id"}).AddRow(booking.BookingID))

	id, err := repo.CreateBooking(context.TODO(), booking)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, id)
}

func TestFetchBookingByID(t *testing.T) {
	db, mock := setup()
	defer db.Close()
	repo := repositories.NewBookingRepo(db)

	bookingID := uuid.New()
	slotID := uuid.New()
	userID := uuid.New()

	rows := sqlmock.NewRows([]string{"booking_id", "slot_id", "user_id", "created_at"}).
		AddRow(bookingID, slotID, userID, time.Now())

	mock.ExpectQuery("SELECT (.+) FROM bookings WHERE booking_id = ?").
		WithArgs(bookingID).
		WillReturnRows(rows)

	booking, err := repo.FetchBookingByID(context.TODO(), bookingID)
	assert.NoError(t, err)
	assert.NotNil(t, booking)
	assert.Equal(t, bookingID, booking.BookingID)
}

func TestDeleteBookingByID(t *testing.T) {
	db, mock := setup()
	defer db.Close()
	repo := repositories.NewBookingRepo(db)

	bookingID := uuid.New()

	mock.ExpectExec("DELETE FROM bookings WHERE booking_id = ?").
		WithArgs(bookingID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.DeleteBookingByID(context.TODO(), bookingID)
	assert.NoError(t, err)
}

func TestFetchBookingsByUserID(t *testing.T) {
	db, mock := setup()
	defer db.Close()
	repo := repositories.NewBookingRepo(db)

	userID := uuid.New()

	rows := sqlmock.NewRows([]string{"booking_id", "slot_id", "user_id", "result", "created_at"}).
		AddRow(uuid.New(), uuid.New(), userID, "win", time.Now())

	mock.ExpectQuery("SELECT (.+) FROM bookings WHERE user_id = ?").
		WithArgs(userID).
		WillReturnRows(rows)

	bookings, err := repo.FetchBookingsByUserID(context.TODO(), userID)
	assert.NoError(t, err)
	assert.Len(t, bookings, 1)
}

func TestFetchBookingsBySlotID(t *testing.T) {
	db, mock := setup()
	defer db.Close()
	repo := repositories.NewBookingRepo(db)

	slotID := uuid.New()

	rows := sqlmock.NewRows([]string{"booking_id", "slot_id", "user_id", "result", "created_at"}).
		AddRow(uuid.New(), slotID, uuid.New(), "win", time.Now())

	mock.ExpectQuery("SELECT (.+) FROM bookings WHERE slot_id = ?").
		WithArgs(slotID).
		WillReturnRows(rows)

	bookings, err := repo.FetchBookingsBySlotID(context.TODO(), slotID)
	assert.NoError(t, err)
	assert.Len(t, bookings, 1)
}

func TestFetchUpcomingBookingsByUserID(t *testing.T) {
	db, mock := setup() // Assuming setup initializes sqlmock and returns db and mock
	defer db.Close()
	repo := repositories.NewBookingRepo(db)

	userID := uuid.New()
	slotID := uuid.New()

	// Mock the query to fetch upcoming bookings
	rows := sqlmock.NewRows([]string{"game_name", "slot_id", "date", "start_time", "end_time"}).
		AddRow("Table Tennis", slotID, time.Now(), time.Now().Add(1*time.Hour), time.Now().Add(2*time.Hour))

	mock.ExpectQuery("SELECT (.+) FROM bookings").
		WithArgs(userID).
		WillReturnRows(rows)

	// Mock the query to fetch booked users for each slot
	userRows := sqlmock.NewRows([]string{"username"}).
		AddRow("john_doe")

	mock.ExpectQuery("SELECT u.username FROM bookings b INNER JOIN users u ON b.user_id = u.user_id WHERE b.slot_id = ?").
		WithArgs(slotID).
		WillReturnRows(userRows)

	// Execute the method
	bookings, err := repo.FetchUpcomingBookingsByUserID(context.TODO(), userID)

	// Assertions
	assert.NoError(t, err)
	assert.Len(t, bookings, 1)
	assert.Equal(t, "Table Tennis", bookings[0].GameName)
	assert.Equal(t, "john_doe", bookings[0].BookedUsers[0]) // Assuming BookedUsers is a field in your result struct
}

func TestUpdateBookingResult(t *testing.T) {
	db, mock := setup()
	defer db.Close()
	repo := repositories.NewBookingRepo(db)

	bookingID := uuid.New()
	result := "win"

	mock.ExpectExec(`UPDATE bookings SET result = \$1 WHERE booking_id = \$2`).
		WithArgs(result, bookingID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.UpdateBookingResult(context.TODO(), bookingID, result)
	assert.NoError(t, err)
}

func TestFetchBookingsToUpdateResult(t *testing.T) {
	db, mock := setup() // setup initializes sqlmock and returns db and mock
	defer db.Close()
	repo := repositories.NewBookingRepo(db) // Using the real BookingRepo

	userID := uuid.New()
	slotID := uuid.New() // Reuse the same slotID for both queries

	// Mock the query to fetch bookings for the user
	rows := sqlmock.NewRows([]string{"booking_id", "game_name", "slot_id", "slot_date", "start_time", "end_time"}).
		AddRow(uuid.New(), "Table Tennis", slotID, time.Now().Add(-1*time.Hour), time.Now().Add(-30*time.Minute), time.Now())

	mock.ExpectQuery("SELECT (.+) FROM bookings").
		WithArgs(userID, time.Now(), "pending").
		WillReturnRows(rows)

	// Mock the query to fetch booked users for the given slot_id
	userRows := sqlmock.NewRows([]string{"username"}).
		AddRow("john_doe")

	mock.ExpectQuery("SELECT u.username FROM bookings b INNER JOIN users u ON b.user_id = u.user_id WHERE b.slot_id = ?").
		WithArgs(slotID).
		WillReturnRows(userRows)

	// Call the method under test
	bookings, err := repo.FetchBookingsToUpdateResult(context.TODO(), userID)

	// Assertions
	assert.NoError(t, err)
	assert.Len(t, bookings, 1)                              // Ensure we get one booking back
	assert.Equal(t, "john_doe", bookings[0].BookedUsers[0]) // Assuming BookedUsers is a field in the result
}

func TestFetchSlotBookedUsers(t *testing.T) {
	db, mock := setup()
	defer db.Close()
	repo := repositories.NewBookingRepo(db)

	slotID := uuid.New()

	rows := sqlmock.NewRows([]string{"username"}).
		AddRow("user1").
		AddRow("user2")

	mock.ExpectQuery("SELECT (.+) FROM bookings").
		WithArgs(slotID).
		WillReturnRows(rows)

	usernames, err := repo.FetchSlotBookedUsers(context.TODO(), slotID)
	assert.NoError(t, err)
	assert.Len(t, usernames, 2)
}

func TestFetchBookingBySlotAndUserId(t *testing.T) {
	// Initialize sqlmock and return db and mock
	db, mock := setup()
	defer db.Close()

	// Create the real BookingRepo
	repo := repositories.NewBookingRepo(db)

	// Test data
	userID := uuid.New()
	slotID := uuid.New()
	bookingID := uuid.New()
	gameName := "Table Tennis"
	date := time.Now().Truncate(24 * time.Hour)
	startTime := date.Add(1 * time.Hour)
	endTime := date.Add(1*time.Hour + 20*time.Minute)

	// Mock the query to fetch booking for the user and slot
	bookingRows := sqlmock.NewRows([]string{"booking_id", "game_name", "slot_date", "start_time", "end_time"}).
		AddRow(bookingID, gameName, date, startTime, endTime)

	mock.ExpectQuery("SELECT (.+) FROM bookings b JOIN slots s ON b.slot_id = s.slot_id JOIN games g ON s.game_id = g.game_id").
		WithArgs(slotID, userID).
		WillReturnRows(bookingRows)

	// Mock the query to fetch booked users for the given slot_id
	userRows := sqlmock.NewRows([]string{"username"}).
		AddRow("john_doe").
		AddRow("jane_smith")

	mock.ExpectQuery(`SELECT u.username FROM bookings b INNER JOIN users u ON b.user_id = u.user_id WHERE b.slot_id = \$1`).
		WithArgs(slotID).
		WillReturnRows(userRows)

	// Call the method under test
	booking, err := repo.FetchBookingBySlotAndUserId(context.TODO(), slotID, userID)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, bookingID, booking.BookingId)
	assert.Equal(t, gameName, booking.GameName)
	assert.Equal(t, startTime, booking.StartTime)
	assert.Equal(t, endTime, booking.EndTime)
	//assert.Equal(t, []string{"john_doe", "jane_smith"}, booking.BookedUsers)
}
