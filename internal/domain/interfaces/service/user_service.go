package service_interfaces

import (
	"context"
	"github.com/google/uuid"
	"project2/internal/domain/entities"
)

type UserService interface {
	EmailAlreadyRegistered(ctx context.Context, email string) bool
	GetUserByID(ctx context.Context, userID uuid.UUID) (*entities.User, error)
	GetUserByEmail(ctx context.Context, email string) (*entities.User, error)
	GetUserByUsername(ctx context.Context, username string) (*entities.User, error)
}
