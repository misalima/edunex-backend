package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/misalima/edunex-backend/internal/core/domain"
	"github.com/misalima/edunex-backend/internal/core/domain_errors"
	"github.com/misalima/edunex-backend/internal/core/interfaces/irepository"
	"github.com/misalima/edunex-backend/internal/core/interfaces/iservice"
	"golang.org/x/crypto/bcrypt"
)

var _ iservice.UserManager = (*UserService)(nil)

type UserService struct {
	userRepo irepository.UserLoader
}

var _ iservice.UserManager = (*UserService)(nil)

func NewUserService(userRepo irepository.UserLoader) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

func (s *UserService) CreateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	if user == nil || user.Name == "" || user.Email == "" || user.Password == "" {
		return nil, domain_errors.NewBadRequestMsg("missing required fields")
	}

	if user.ID != uuid.Nil {
		return nil, domain_errors.NewBadRequestMsg("user ID should not be set")
	}

	existingUser, err := s.userRepo.GetUserByEmail(ctx, user.Email)
	if err != nil {
		if errors.Is(err, domain_errors.ErrUserNotFound) || errors.Is(err, domain_errors.ErrNotFound) {
		} else {
			return nil, err
		}
	} else if existingUser != nil {
		return nil, domain_errors.NewConflictMsg("email already in use")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, domain_errors.WrapUnexpectedMsg(err, "failed to hash password")
	}
	user.Password = string(hashedPassword)

	if user.Role == "" {
		user.Role = "coordinator"
	}

	id, err := s.userRepo.InsertUser(ctx, user)
	if err != nil {
		return nil, err
	}

	createdUser, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	createdUser.Password = ""

	return createdUser, nil
}

func (s *UserService) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	user, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, domain_errors.NewNotFoundMsg("user not found")
	}

	user.Password = ""
	return user, nil
}

func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain_errors.ErrUserNotFound) || errors.Is(err, domain_errors.ErrNotFound) {
			return nil, domain_errors.NewNotFoundMsg("user not found")
		}
		return nil, err
	}

	if user == nil {
		return nil, domain_errors.NewNotFoundMsg("user not found")
	}

	user.Password = ""
	return user, nil
}

func (s *UserService) UpdateUser(ctx context.Context, user *domain.User) error {
	if user == nil || user.ID == uuid.Nil {
		return domain_errors.NewBadRequestMsg("user ID is required")
	}

	if err := s.userRepo.UpdateUser(ctx, user); err != nil {
		return err
	}

	return nil
}

func (s *UserService) ListUsers(ctx context.Context) ([]*domain.User, error) {
	users, err := s.userRepo.ListUsers(ctx)
	if err != nil {
		return nil, err
	}

	for _, u := range users {
		u.Password = ""
	}
	return users, nil
}

func (s *UserService) UpdateUserRole(ctx context.Context, userID uuid.UUID, newRole string) error {
	switch newRole {
	case "admin", "coordinator", "teacher":
	default:
		return domain_errors.NewBadRequestMsg("invalid role")
	}

	if userID == uuid.Nil {
		return domain_errors.NewBadRequestMsg("user ID is required")
	}

	if err := s.userRepo.UpdateRole(ctx, userID, newRole); err != nil {
		return err
	}

	return nil
}
