package repository_test

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"project2/internal/app/repositories"
	"project2/internal/domain/entities"
	"regexp"
	"testing"
	"time"
)

func TestCreateInvitation(t *testing.T) {
	db, mock := setup()
	defer db.Close()

	repo := repositories.NewInvitationRepo(db)
	invitation := &entities.Invitation{
		InvitingUserID: uuid.New(),
		InvitedUserID:  uuid.New(),
		SlotID:         uuid.New(),
	}

	mock.ExpectQuery(`INSERT INTO invitations`).
		WithArgs(invitation.InvitingUserID, invitation.InvitedUserID, invitation.SlotID).
		WillReturnRows(sqlmock.NewRows([]string{"invitation_id"}).AddRow(uuid.New()))

	id, err := repo.CreateInvitation(context.Background(), invitation)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, id)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteInvitationByID(t *testing.T) {
	db, mock := setup()
	defer db.Close()

	repo := repositories.NewInvitationRepo(db)
	id := uuid.New()

	mock.ExpectExec(`DELETE FROM invitations WHERE invitation_id = ?`).
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.DeleteInvitationByID(context.Background(), id)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateInvitationStatus(t *testing.T) {
	db, mock := setup()
	defer db.Close()

	repo := repositories.NewInvitationRepo(db)
	id := uuid.New()
	status := "accepted"

	mock.ExpectExec(`UPDATE invitations SET status = \$1 WHERE invitation_id = \$2`).
		WithArgs(status, id).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.UpdateInvitationStatus(context.Background(), id, status)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestFetchInvitationByID(t *testing.T) {
	db, mock := setup()
	defer db.Close()

	repo := repositories.NewInvitationRepo(db)
	id := uuid.New()

	rows := sqlmock.NewRows([]string{"invitation_id", "inviting_user_id", "invited_user_id", "slot_id", "status", "created_at"}).
		AddRow(id, uuid.New(), uuid.New(), uuid.New(), "pending", time.Now())

	mock.ExpectQuery(`SELECT invitation_id, inviting_user_id, invited_user_id,slot_id, status, created_at FROM invitations WHERE invitation_id = \$1`).
		WithArgs(id).
		WillReturnRows(rows)

	invitation, err := repo.FetchInvitationByID(context.Background(), id)
	require.NoError(t, err)
	require.NotNil(t, invitation)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestFetchUserInvitations(t *testing.T) {
	db, mock := setup()
	defer db.Close()

	repo := repositories.NewInvitationRepo(db)
	userID := uuid.New()

	rows := sqlmock.NewRows([]string{"invitation_id", "inviting_user_id", "invited_user_id", "status", "created_at"}).
		AddRow(uuid.New(), userID, uuid.New(), "pending", time.Now()).
		AddRow(uuid.New(), uuid.New(), userID, "accepted", time.Now())

	mock.ExpectQuery(`SELECT invitation_id, inviting_user_id, invited_user_id, status, created_at FROM invitations WHERE inviting_user_id = \$1 OR invited_user_id = \$2`).
		WithArgs(userID, userID).
		WillReturnRows(rows)

	invitations, err := repo.FetchUserInvitations(context.Background(), userID)
	require.NoError(t, err)
	require.Len(t, invitations, 2)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestFetchUserPendingInvitations(t *testing.T) {
	db, mock := setup()
	defer db.Close()

	repo := repositories.NewInvitationRepo(db)
	userID := uuid.New()

	// Define the expected query and result set
	query := `
        SELECT
            i.invitation_id,
            s.slot_id,
            g.game_name,
            s.slot_date,
            s.start_time,
            s.end_time,
            ARRAY_AGG(u.username) AS booked_users,
            inviter.username AS invited_by_username
        FROM
            invitations i
            JOIN slots s ON i.slot_id = s.slot_id
            JOIN games g ON s.game_id = g.game_id
            LEFT JOIN bookings b ON s.slot_id = b.slot_id
            LEFT JOIN users u ON b.user_id = u.user_id
            JOIN users inviter ON i.inviting_user_id = inviter.user_id
        WHERE
            i.invited_user_id = $1
            AND i.status = 'pending'
            AND s.start_time > NOW()  
        GROUP BY
            i.invitation_id, s.slot_id, g.game_name, s.slot_date, s.start_time, s.end_time, inviter.username
        ORDER BY
            s.start_time;
    `

	// Create time values for the test data
	now := time.Now()
	slotDate := time.Now().Truncate(24 * time.Hour)
	startTime := now
	endTime := now.Add(20 * time.Minute)

	rows := sqlmock.NewRows([]string{
		"invitation_id", "slot_id", "game_name", "slot_date", "start_time", "end_time", "booked_users", "invited_by_username",
	}).AddRow(
		uuid.New(),
		uuid.New(),
		"Table Tennis",
		slotDate,
		startTime,
		endTime,
		pq.Array([]string{"user1"}),
		"inviter1",
	)

	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(userID).
		WillReturnRows(rows)

	invitations, err := repo.FetchUserPendingInvitations(context.Background(), userID)
	require.NoError(t, err)
	require.Len(t, invitations, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestFetchInvitationByUserAndSlot(t *testing.T) {
	db, mock := setup()
	defer db.Close()

	repo := repositories.NewInvitationRepo(db)
	invitingUserID := uuid.New()
	invitedUserID := uuid.New()
	slotID := uuid.New()

	rows := sqlmock.NewRows([]string{"invitation_id", "inviting_user_id", "invited_user_id", "slot_id"}).
		AddRow(uuid.New(), invitingUserID, invitedUserID, slotID)

	mock.ExpectQuery(`SELECT invitation_id, inviting_user_id, invited_user_id, slot_id FROM invitations WHERE inviting_user_id = \$1 AND invited_user_id = \$2 AND slot_id = \$3`).
		WithArgs(invitingUserID, invitedUserID, slotID).
		WillReturnRows(rows)

	invitation, err := repo.FetchInvitationByUserAndSlot(context.TODO(), invitingUserID, invitedUserID, slotID)
	require.NoError(t, err)
	require.NotNil(t, invitation)
	require.NoError(t, mock.ExpectationsWereMet())
}
