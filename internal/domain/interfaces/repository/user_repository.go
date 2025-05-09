package repository_interfaces

import (
	"context"
	"github.com/google/uuid"
	"project2/internal/domain/entities"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *entities.User) (uuid.UUID, error)
	FetchUserByEmail(ctx context.Context, email string) (*entities.User, error)
	FetchUserById(ctx context.Context, id uuid.UUID) (*entities.User, error)
	FetchAllUsers(ctx context.Context, limit, offset, substring string) ([]entities.User, int, error)
	FetchAllUsersPublic(ctx context.Context, userID uuid.UUID, slotID uuid.UUID) ([]entities.User, error)
	EmailAlreadyExists(ctx context.Context, email string) bool
	FetchUserByUsername(ctx context.Context, username string) (*entities.User, error)
	RemoveUser(ctx context.Context, userId uuid.UUID) error
	UpdateUser(ctx context.Context, user *entities.User) error
	UpdatePassword(ctx context.Context, email string, password string) error
}
