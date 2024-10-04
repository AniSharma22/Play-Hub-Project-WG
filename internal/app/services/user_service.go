package services

import (
	"context"
	"github.com/google/uuid"
	"project2/internal/domain/entities"
	repository_interfaces "project2/internal/domain/interfaces/repository"
	service_interfaces "project2/internal/domain/interfaces/service"
	"sync"
)

type UserService struct {
	userRepo repository_interfaces.UserRepository
	userWG   *sync.WaitGroup
}

func NewUserService(userRepo repository_interfaces.UserRepository) service_interfaces.UserService {
	return &UserService{
		userRepo: userRepo,
		userWG:   &sync.WaitGroup{},
	}
}

func (s *UserService) GetAllUsers(ctx context.Context) ([]entities.User, error) {
	return s.userRepo.FetchAllUsers(ctx)
}

// EmailAlreadyRegistered checks if an email is already registered in the system.
func (s *UserService) EmailAlreadyRegistered(ctx context.Context, email string) bool {
	return s.userRepo.EmailAlreadyExists(ctx, email)
}

// GetUserByID retrieves a user by their ID.
func (s *UserService) GetUserByID(ctx context.Context, userID uuid.UUID) (*entities.User, error) {
	return s.userRepo.FetchUserById(ctx, userID)
}

// GetUserByEmail retrieves a user by their email address.
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	return s.userRepo.FetchUserByEmail(ctx, email)
}

// GetUserByUsername retrieves a user by their username.
func (s *UserService) GetUserByUsername(ctx context.Context, username string) (*entities.User, error) {
	return s.userRepo.FetchUserByUsername(ctx, username)
}
