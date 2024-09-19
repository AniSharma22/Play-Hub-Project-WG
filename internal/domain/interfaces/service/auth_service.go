package service_interfaces

import (
	"context"
	"project2/internal/domain/entities"
)

type AuthService interface {
	Signup(ctx context.Context, user *entities.User) (*entities.User, error)
	Login(ctx context.Context, email string, password []byte) (*entities.User, error)
}
