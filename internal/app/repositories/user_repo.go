package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"project2/internal/db"
	"project2/internal/domain/entities"
	interfaces "project2/internal/domain/interfaces/repository"
	"strconv"
)

type userRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) interfaces.UserRepository {
	return &userRepo{
		db: db,
	}
}

// CreateUser creates a new user in the DB
func (r *userRepo) CreateUser(ctx context.Context, user *entities.User) (uuid.UUID, error) {
	// Insert into PostgresSQL and return the user_id
	//query := `
	//	INSERT INTO users (username, email, password, mobile_number, gender, imageUrl)
	//	VALUES ($1, $2, $3, $4, $5)
	//	RETURNING user_id
	//`

	query := (&db.InsertQueryBuilder{
		Table:       "users",
		Columns:     "username, email, password, mobile_number, gender, image_url",
		ReturnValue: "user_id",
	}).Build()

	row := r.db.QueryRowContext(ctx, query, user.Username, user.Email, user.Password, user.MobileNumber, user.Gender, user.ImageUrl)

	// Variable to hold the returned user_id
	var userID uuid.UUID
	err := row.Scan(&userID)
	if err != nil {
		return uuid.Nil, errors.New("failed to insert user into PostgresSQL and retrieve user_id")
	}
	return userID, nil
}

// FetchUserByEmail retrieves a user by their email address.
func (r *userRepo) FetchUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	//query := `SELECT user_id, username, email, password, mobile_number, gender,role FROM users WHERE email = $1`

	query := (&db.SelectQueryBuilder{
		Columns: "user_id, username, email, password, mobile_number, gender,role",
		From:    "users",
		Where:   "email = $1",
	}).Build()

	row := r.db.QueryRowContext(ctx, query, email)

	var user entities.User
	err := row.Scan(&user.UserID, &user.Username, &user.Email, &user.Password, &user.MobileNumber, &user.Gender, &user.Role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("no such user found")
		}
		return nil, fmt.Errorf("failed to fetch user by email: %w", err)
	}

	return &user, nil
}

// FetchUserById retrieves a user by their unique user_id.
func (r *userRepo) FetchUserById(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	//query := `SELECT user_id, username, email, password, mobile_number, gender,role FROM users WHERE user_id = $1`

	query := (&db.SelectQueryBuilder{
		Columns: "user_id, username, email, password, mobile_number, gender, role, image_url",
		From:    "users",
		Where:   "user_id = $1",
	}).Build()

	row := r.db.QueryRowContext(ctx, query, id)

	var user entities.User
	err := row.Scan(&user.UserID, &user.Username, &user.Email, &user.Password, &user.MobileNumber, &user.Gender, &user.Role, &user.ImageUrl)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("no such user found") // No user found
		}
		return nil, fmt.Errorf("failed to fetch user by ID: %w", err)
	}

	return &user, nil
}

// FetchAllUsers retrieves all users from the database.
func (r *userRepo) FetchAllUsers(ctx context.Context, limit, offset, substring string) ([]entities.User, int, error) {
	// Convert limit and offset to integers
	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid limit: %w", err)
	}
	offsetInt, err := strconv.Atoi(offset)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid offset: %w", err)
	}

	var query *db.SelectQueryBuilder
	if substring != "" {
		query = &db.SelectQueryBuilder{
			Columns: "user_id, username, email, password, mobile_number, gender, role, image_url",
			From:    "users",
			Where:   "username LIKE $1",
			Params:  []any{fmt.Sprintf("%%%s%%", substring)}, // Search by substring
			Limit:   fmt.Sprintf("%d", limitInt),             // Make sure it's a string
			Offset:  fmt.Sprintf("%d", offsetInt),            // Make sure it's a string
		}
	} else {
		query = &db.SelectQueryBuilder{
			Columns: "user_id, username, email, password, mobile_number, gender, role, image_url",
			From:    "users",
			Limit:   fmt.Sprintf("%d", limitInt),
			Offset:  fmt.Sprintf("%d", offsetInt),
		}
	}

	// Run the main query to fetch users
	rows, err := r.db.QueryContext(ctx, query.Build(), query.Params...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch all users: %w", err)
	}
	defer rows.Close()

	var users []entities.User
	for rows.Next() {
		var user entities.User
		if err := rows.Scan(&user.UserID, &user.Username, &user.Email, &user.Password, &user.MobileNumber, &user.Gender, &user.Role, &user.ImageUrl); err != nil {
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	// Check if any error occurred during the rows iteration
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("errors encountered during rows iteration: %w", err)
	}

	// Retrieve total count
	var total int
	countQuery := "SELECT COUNT(*) FROM users"
	if substring != "" {
		countQuery += " WHERE username LIKE $1"
	}

	// Use the same substring parameter for counting the total if it's not empty
	if substring != "" {
		if err := r.db.QueryRowContext(ctx, countQuery, fmt.Sprintf("%%%s%%", substring)).Scan(&total); err != nil {
			return nil, 0, fmt.Errorf("failed to get total count: %w", err)
		}
	} else {
		if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
			return nil, 0, fmt.Errorf("failed to get total count: %w", err)
		}
	}

	return users, total, nil
}

// FetchAllUsers retrieves all users from the database who are
func (r *userRepo) FetchAllUsersPublic(ctx context.Context, userID uuid.UUID, slotID uuid.UUID) ([]entities.User, error) {
	//query = SELECT
	//u.user_id,
	//	u.username,
	//	u.image_url,
	//	u.email
	//FROM
	//users u
	//WHERE
	//u.user_id != $1  -- Exclude current user
	//AND u.user_id NOT IN (
	//	SELECT invited_user_id
	//FROM invitations
	//WHERE
	//inviting_user_id = $1
	//AND slot_id = $2
	//AND status IN ('pending', 'accepted')
	//)
	//AND u.user_id NOT IN (
	//	SELECT booked_user_id
	//FROM slot_bookings
	//WHERE slot_id = $2
	//)
	//ORDER BY
	//u.username;

	query := (&db.SelectQueryBuilder{
		Columns: "u.user_id, u.username, u.email, u.image_url",
		From:    "users u",
		Where: "u.user_id != $1" +
			" AND u.user_id NOT IN " +
			"(SELECT invited_user_id FROM invitations WHERE inviting_user_id = $1 AND slot_id = $2 AND status IN ('pending', 'accepted'))",
		OrderBy: "u.username",
	}).Build()

	rows, err := r.db.QueryContext(ctx, query, userID, slotID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch all users: %w", err)
	}
	defer rows.Close()

	var users []entities.User
	for rows.Next() {
		var user entities.User
		if err := rows.Scan(&user.UserID, &user.Username, &user.Email, &user.ImageUrl); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("errs encountered during rows iteration: %w", err)
	}

	return users, nil
}

// EmailAlreadyExists checks if the given email already exists in the database.
func (r *userRepo) EmailAlreadyExists(ctx context.Context, email string) bool {
	//query := `SELECT 1 FROM users WHERE email = $1`

	query := (&db.SelectQueryBuilder{
		Columns: "1",
		From:    "users",
		Where:   "email = $1",
	}).Build()

	row := r.db.QueryRowContext(ctx, query, email)

	var exists bool
	err := row.Scan(&exists)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false
	}

	return exists
}

func (r *userRepo) FetchUserByUsername(ctx context.Context, username string) (*entities.User, error) {
	//query := `SELECT user_id, username, email, password, mobile_number, gender, role, created_at, updated_at FROM users WHERE username = $1`

	query := (&db.SelectQueryBuilder{
		Columns: "user_id, username, email, password, mobile_number, gender, role, created_at, updated_at",
		From:    "users",
		Where:   "username = $1",
	}).Build()

	row := r.db.QueryRowContext(ctx, query, username)

	var user entities.User
	err := row.Scan(
		&user.UserID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.MobileNumber,
		&user.Gender,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("no such user found") // No user found
		}
		return nil, fmt.Errorf("failed to fetch user by username: %w", err)
	}

	return &user, nil
}

func (r *userRepo) RemoveUser(ctx context.Context, userId uuid.UUID) error {
	query := (&db.DeleteQueryBuilder{
		Table: "users",
		Where: "user_id = $1",
	}).Build()

	_, err := r.db.ExecContext(ctx, query, userId)
	if err != nil {
		return err
	}

	return nil
}

func (r *userRepo) UpdateUser(ctx context.Context, user *entities.User) error {
	// Define query components
	var query string
	var args []interface{}

	// Determine the query based on whether password is set or not
	if user.Password != "" {
		query = "UPDATE users SET username = $1, mobile_number = $2, password = $3, image_url = $4 WHERE user_id = $5"
		args = []interface{}{user.Username, user.MobileNumber, user.Password, user.ImageUrl, user.UserID}
	} else {
		query = "UPDATE users SET username = $1, mobile_number = $2, image_url = $3 WHERE user_id = $4"
		args = []interface{}{user.Username, user.MobileNumber, user.ImageUrl, user.UserID}
	}

	// Execute the query
	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (r *userRepo) UpdatePassword(ctx context.Context, email string, password string) error {
	query := (&db.UpdateQueryBuilder{
		Table: "users",
		Where: "email = $1",
		Set:   "password = $2",
	}).Build()

	// Execute the query
	_, err := r.db.ExecContext(ctx, query, email, password)
	if err != nil {
		return err
	}

	return nil
}
