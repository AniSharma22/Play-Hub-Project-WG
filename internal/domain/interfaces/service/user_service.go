package service_interfaces

import (
	"context"
	"github.com/google/uuid"
	"project2/internal/domain/entities"
)

type UserService interface {
	GetAllUsers(ctx context.Context, limit, offset, substring string) ([]entities.User, int, error)
	GetAllUsersPublic(ctx context.Context, userID uuid.UUID, slotID uuid.UUID) ([]entities.User, error)
	EmailAlreadyRegistered(ctx context.Context, email string) bool
	GetUserByID(ctx context.Context, userID uuid.UUID) (*entities.User, error)
	GetUserByEmail(ctx context.Context, email string) (*entities.User, error)
	GetUserByUsername(ctx context.Context, username string) (*entities.User, error)
	DeleteUser(ctx context.Context, userID uuid.UUID) error
	UpdateUserDetails(ctx context.Context, user *entities.User) error
}
