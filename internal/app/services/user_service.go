package services

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"project2/internal/api/middleware"
	"project2/internal/domain/entities"
	repository_interfaces "project2/internal/domain/interfaces/repository"
	service_interfaces "project2/internal/domain/interfaces/service"
	"project2/pkg/utils"
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

// GetAllUsers Returns all the details of a user (can be accessed only by admin)
func (s *UserService) GetAllUsers(ctx context.Context, limit, offset, substring string) ([]entities.User, int, error) {
	return s.userRepo.FetchAllUsers(ctx, limit, offset, substring)
}

// GetAllUsersPublic Returns a list of all the users basic details (can be accessed by public)
func (s *UserService) GetAllUsersPublic(ctx context.Context, userID uuid.UUID, slotID uuid.UUID) ([]entities.User, error) {
	return s.userRepo.FetchAllUsersPublic(ctx, userID, slotID)
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

// DeleteUser deletes a user from the database
func (s *UserService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	// id of the user performing the delete operation
	uidStr, _ := ctx.Value(middleware.UserIdKey).(string)
	uid, _ := uuid.Parse(uidStr)
	if uid == userID {
		return errors.New("operation cannot be performed")
	}
	return s.userRepo.RemoveUser(ctx, userID)
}

func (s *UserService) UpdateUserDetails(ctx context.Context, user *entities.User) error {
	if user.Password != "" {
		// if the password field is present
		user.Password, _ = utils.GetHashedPassword([]byte(user.Password))
	}

	err := s.userRepo.UpdateUser(ctx, user)
	if err != nil {
		return err
	}
	return nil
}
