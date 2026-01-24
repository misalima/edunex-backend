package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/misalima/edunex-backend/internal/core/domain"
	"github.com/misalima/edunex-backend/internal/core/interfaces/irepository"
	"github.com/misalima/edunex-backend/internal/core/interfaces/iservice"
	"golang.org/x/crypto/bcrypt"
)

var _ iservice.UserManager = (*UserService)(nil)

type UserService struct {
	userRepo irepository.UserLoader
}

func NewUserService(userRepo irepository.UserLoader) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

func (s *UserService) CreateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	if user.Name == "" || user.Email == "" || user.Password == "" {
		return nil, errors.New("missing required fields")
	}

	if user.ID != uuid.Nil {
		return nil, errors.New("user ID should not be set")
	}

	existingUser, err := s.userRepo.GetUserByEmail(ctx, user.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("email already in use")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user.Password = string(hashedPassword)

	user.Role = "coordinator"

	id, err := s.userRepo.InsertUser(ctx, user)
	if err != nil {
		return nil, err
	}

	user.ID, err = uuid.Parse(id)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	user, err = s.userRepo.GetUserByID(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return user, nil
}
func (s *UserService) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	user, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}
func (s *UserService) UpdateUser(ctx context.Context, user *domain.User) error {
	if user.ID == uuid.Nil {
		return errors.New("user ID is required")
	}

	err := s.userRepo.UpdateUser(ctx, user)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserService) ListUsers(ctx context.Context) ([]*domain.User, error) {
	users, err := s.userRepo.ListUsers(ctx)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (s *UserService) UpdateUserRole(ctx context.Context, userID string, newRole string) error {
	if newRole != "admin" && newRole != "coordinator" && newRole != "teacher" {
		return errors.New("invalid role")
	}
	return s.userRepo.UpdateRole(ctx, userID, newRole)
}
