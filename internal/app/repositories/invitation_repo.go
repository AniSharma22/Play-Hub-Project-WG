package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"project2/internal/db"
	"project2/internal/domain/entities"
	interfaces "project2/internal/domain/interfaces/repository"
	"project2/internal/models"
)

type invitationRepo struct {
	db *sql.DB
}

func NewInvitationRepo(db *sql.DB) interfaces.InvitationRepository {
	return &invitationRepo{db: db}
}

// CreateInvitation inserts a new invitation into the database and returns the created invitation ID.
func (r *invitationRepo) CreateInvitation(ctx context.Context, invitation *entities.Invitation) (uuid.UUID, error) {
	//query := `INSERT INTO invitations (inviting_user_id, invited_user_id, slot_id,game_id) VALUES ($1, $2, $3, $4) RETURNING invitation_id`

	query := (&db.InsertQueryBuilder{
		Table:       "invitations",
		Columns:     "inviting_user_id, invited_user_id, slot_id,game_id",
		ReturnValue: "invitation_id",
	}).Build()

	var id uuid.UUID
	err := r.db.QueryRowContext(ctx, query, invitation.InvitingUserID, invitation.InvitedUserID, invitation.SlotID, invitation.GameID).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create invitation: %w", err)
	}
	return id, nil
}

// DeleteInvitationByID removes an invitation from the database by its ID.
func (r *invitationRepo) DeleteInvitationByID(ctx context.Context, id uuid.UUID) error {
	//query := `DELETE FROM invitations WHERE invitation_id = $1`

	query := (&db.DeleteQueryBuilder{
		Table: "invitations",
		Where: "invitation_id = $1",
	}).Build()

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete invitation: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no invitation found with ID %s", id)
	}

	return nil
}

// UpdateInvitationStatus updates the status of an invitation by its ID.
func (r *invitationRepo) UpdateInvitationStatus(ctx context.Context, id uuid.UUID, status string) error {
	//query := `UPDATE invitations SET status = $1 WHERE invitation_id = $2`

	query := (&db.UpdateQueryBuilder{
		Table: "invitations",
		Set:   "status = $1",
		Where: "invitation_id = $2",
	}).Build()

	_, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update invitation status: %w", err)
	}
	return nil
}

// FetchInvitationByID retrieves an invitation by its ID.
func (r *invitationRepo) FetchInvitationByID(ctx context.Context, id uuid.UUID) (*entities.Invitation, error) {
	//query := `SELECT invitation_id, inviting_user_id, invited_user_id,slot_id,game_id, status, created_at FROM invitations WHERE invitation_id = $1`

	query := (&db.SelectQueryBuilder{
		Columns: "invitation_id, inviting_user_id, invited_user_id, slot_id,game_id, status, created_at",
		From:    "invitations",
		Where:   "invitation_id = $1",
	}).Build()
	row := r.db.QueryRowContext(ctx, query, id)

	var invitation entities.Invitation
	err := row.Scan(&invitation.InvitationID, &invitation.InvitingUserID, &invitation.InvitedUserID, &invitation.SlotID, &invitation.GameID, &invitation.Status, &invitation.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // No invitation found
		}
		return nil, fmt.Errorf("failed to fetch invitation by ID: %w", err)
	}

	return &invitation, nil
}

// FetchUserInvitations retrieves all invitations sent to or from a specific user.
func (r *invitationRepo) FetchUserInvitations(ctx context.Context, userID uuid.UUID) ([]entities.Invitation, error) {
	//query := `SELECT invitation_id, inviting_user_id, invited_user_id,game_id, status, created_at FROM invitations WHERE inviting_user_id = $1 OR invited_user_id = $2`

	query := (&db.SelectQueryBuilder{
		Columns: "invitation_id, inviting_user_id, invited_user_id,game_id, status, created_at",
		From:    "invitations",
		Where:   "inviting_user_id = $1 OR invited_user_id = $2",
	}).Build()
	rows, err := r.db.QueryContext(ctx, query, userID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user invitations: %w", err)
	}
	defer rows.Close()

	var invitations []entities.Invitation
	for rows.Next() {
		var invitation entities.Invitation
		if err := rows.Scan(&invitation.InvitationID, &invitation.InvitingUserID, &invitation.InvitedUserID, &invitation.GameID, &invitation.Status, &invitation.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan invitation row: %w", err)
		}
		invitations = append(invitations, invitation)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("errs occurred while iterating over invitations: %w", err)
	}

	return invitations, nil
}

// FetchUserPendingInvitations retrieves all the pending invitations of the user
func (r *invitationRepo) FetchUserPendingInvitations(ctx context.Context, userID uuid.UUID) ([]models.Invitations, error) {
	query := (&db.SelectQueryBuilder{
		Columns: "i.invitation_id, " +
			"s.slot_id, g.game_id, " +
			"g.game_name, g.image_url, s.slot_date, " +
			"s.start_time, " +
			"s.end_time, " +
			// Fixed COALESCE with matching types using json_agg instead
			"COALESCE(json_agg(json_build_object('user_name', u.username, 'user_image', u.image_url)) " +
			"FILTER (WHERE u.username IS NOT NULL), '[]'::json) AS booked_users, " +
			"inviter.username AS invited_by_username",
		From: "invitations i " +
			"JOIN slots s ON i.slot_id = s.slot_id " +
			"JOIN games g ON s.game_id = g.game_id " +
			"LEFT JOIN bookings b ON s.slot_id = b.slot_id " +
			"LEFT JOIN users u ON b.user_id = u.user_id " +
			"JOIN users inviter ON i.inviting_user_id = inviter.user_id",
		Where:   "i.invited_user_id = $1 AND i.status = 'pending' AND s.start_time > CURRENT_TIMESTAMP AND g.is_active = true",
		GroupBy: "i.invitation_id, s.slot_id, g.game_id, g.game_name, g.image_url, s.slot_date, s.start_time, s.end_time, inviter.username",
		OrderBy: "s.start_time, i.invitation_id",
	}).Build()

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending invitations: %w", err)
	}
	defer rows.Close()

	var invitations []models.Invitations
	for rows.Next() {
		var invitation models.Invitations
		var bookedUsersJSON []byte // Use []byte to receive the JSON array

		err := rows.Scan(
			&invitation.InvitationId,
			&invitation.SlotId,
			&invitation.GameId,
			&invitation.GameName,
			&invitation.ImageUrl,
			&invitation.Date,
			&invitation.StartTime,
			&invitation.EndTime,
			&bookedUsersJSON, // Scan into JSON byte array
			&invitation.InvitedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan invitation: %w", err)
		}

		// Parse the JSON array into []BasicUser
		var bookedUsers []models.BasicUser
		if len(bookedUsersJSON) > 0 {
			err = json.Unmarshal(bookedUsersJSON, &bookedUsers)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal booked users: %w", err)
			}
		}

		invitation.BookedUsers = bookedUsers
		invitations = append(invitations, invitation)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return invitations, nil
}

// FetchUserSentInvitations retrieves all the sent invitations of the user
func (r *invitationRepo) FetchUserSentInvitations(ctx context.Context, userID uuid.UUID) ([]models.Invitations, error) {
	query := (&db.SelectQueryBuilder{
		Columns: "i.invitation_id, " +
			"s.slot_id, g.game_id, " +
			"g.game_name, g.image_url, s.slot_date, " +
			"s.start_time, " +
			"s.end_time, " +
			"COALESCE(json_agg(json_build_object('user_name', u.username, 'user_image', u.image_url)) " +
			"FILTER (WHERE u.username IS NOT NULL), '[]'::json) AS booked_users, " +
			"invited_user.username AS invited_username", // Fetch only the invited user's username
		From: "invitations i " +
			"JOIN slots s ON i.slot_id = s.slot_id " +
			"JOIN games g ON s.game_id = g.game_id " +
			"LEFT JOIN bookings b ON s.slot_id = b.slot_id " +
			"LEFT JOIN users u ON b.user_id = u.user_id " +
			"JOIN users invited_user ON i.invited_user_id = invited_user.user_id", // Join for invited user
		Where:   "i.inviting_user_id = $1 AND i.status = 'pending' AND s.start_time > CURRENT_TIMESTAMP AND g.is_active = true",
		GroupBy: "i.invitation_id, s.slot_id, g.game_id, g.game_name, g.image_url, s.slot_date, s.start_time, s.end_time, invited_user.username", // Group by username
		OrderBy: "s.start_time, i.invitation_id",
	}).Build()

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending invitations: %w", err)
	}
	defer rows.Close()

	var invitations []models.Invitations
	for rows.Next() {
		var invitation models.Invitations
		var bookedUsersJSON []byte // Use []byte to receive the JSON array

		err := rows.Scan(
			&invitation.InvitationId,
			&invitation.SlotId,
			&invitation.GameId,
			&invitation.GameName,
			&invitation.ImageUrl,
			&invitation.Date,
			&invitation.StartTime,
			&invitation.EndTime,
			&bookedUsersJSON, // Scan into JSON byte array
			&invitation.InvitedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan invitation: %w", err)
		}

		// Parse the JSON array into []BasicUser
		var bookedUsers []models.BasicUser
		if len(bookedUsersJSON) > 0 {
			err = json.Unmarshal(bookedUsersJSON, &bookedUsers)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal booked users: %w", err)
			}
		}

		invitation.BookedUsers = bookedUsers
		invitations = append(invitations, invitation)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return invitations, nil
}

// FetchInvitationByUserAndSlot retrieves an invitation based on user and slot id
func (r *invitationRepo) FetchInvitationByUserAndSlot(ctx context.Context, invitingUserID uuid.UUID, invitedUserID uuid.UUID, slotID uuid.UUID) (*entities.Invitation, error) {
	//query := `
	//	SELECT invitation_id, inviting_user_id, invited_user_id, slot_id,game_id
	//	FROM invitations
	//	WHERE inviting_user_id = $1 AND invited_user_id = $2 AND slot_id = $3
	//`

	query := (&db.SelectQueryBuilder{
		Columns: "invitation_id, inviting_user_id, invited_user_id, slot_id,game_id",
		From:    "invitations",
		Where:   "inviting_user_id = $1 AND invited_user_id = $2 AND slot_id = $3",
	}).Build()

	row := r.db.QueryRowContext(ctx, query, invitingUserID, invitedUserID, slotID)

	var invitation entities.Invitation
	err := row.Scan(&invitation.InvitationID, &invitation.InvitingUserID, &invitation.InvitedUserID, &invitation.SlotID, &invitation.GameID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// No matching invitation found
			return nil, nil
		}
		return nil, fmt.Errorf("failed to fetch invitation by user and slot: %w", err)
	}

	return &invitation, nil
}

// FetchPendingInvitationStatus checks for the existence of pending invitations for a user.
func (r *invitationRepo) FetchPendingInvitationStatus(ctx context.Context, userID uuid.UUID) (bool, error) {

	// Build the query to check for pending invitations with a slot that starts in the future
	query := (&db.SelectQueryBuilder{
		Columns: "i.invitation_id", // We only need the invitation_id to check for existence
		From:    "invitations i JOIN slots s ON i.slot_id = s.slot_id",
		Where:   "i.invited_user_id = $1 AND i.status = 'pending' AND s.start_time > CURRENT_TIMESTAMP",
	}).Build()

	// Execute the query
	row := r.db.QueryRowContext(ctx, query, userID)

	// Variable to store the result
	var invitationID uuid.UUID

	// Scan the result into the invitationID field (if any row is returned)
	err := row.Scan(&invitationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// If no rows are found, return false
			return false, nil
		}
		// If another error occurs, return false and the error
		return false, err
	}

	// If a row is found (invitationID is populated), return true
	return true, nil
}
