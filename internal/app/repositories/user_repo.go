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
	//	INSERT INTO users (username, email, password, mobile_number, gender)
	//	VALUES ($1, $2, $3, $4, $5)
	//	RETURNING user_id
	//`

	query := (&db.InsertQueryBuilder{
		Table:       "users",
		Columns:     "username, email, password, mobile_number, gender",
		ReturnValue: "user_id",
	}).Build()

	row := r.db.QueryRowContext(ctx, query, user.Username, user.Email, user.Password, user.MobileNumber, user.Gender)

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
		Columns: "user_id, username, email, password, mobile_number, gender,role",
		From:    "users",
		Where:   "user_id = $1",
	}).Build()

	row := r.db.QueryRowContext(ctx, query, id)

	var user entities.User
	err := row.Scan(&user.UserID, &user.Username, &user.Email, &user.Password, &user.MobileNumber, &user.Gender, &user.Role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("no such user found") // No user found
		}
		return nil, fmt.Errorf("failed to fetch user by ID: %w", err)
	}

	return &user, nil
}

// FetchAllUsers retrieves all users from the database.
func (r *userRepo) FetchAllUsers(ctx context.Context) ([]entities.User, error) {
	//query := `SELECT user_id, username, email, password, mobile_number, gender,role FROM users`

	query := (&db.SelectQueryBuilder{
		Columns: "user_id, username, email, password, mobile_number, gender,role",
		From:    "users",
	}).Build()

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch all users: %w", err)
	}
	defer rows.Close()

	var users []entities.User
	for rows.Next() {
		var user entities.User
		if err := rows.Scan(&user.UserID, &user.Username, &user.Email, &user.Password, &user.MobileNumber, &user.Gender, &user.Role); err != nil {
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
