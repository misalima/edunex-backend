package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/misalima/edunex-backend/internal/core/domain"
	"github.com/misalima/edunex-backend/internal/core/domain_errors"
	"github.com/misalima/edunex-backend/internal/core/interfaces/irepository"
	"gorm.io/gorm"
)

// ensure UserRepository implements the UserLoader interface
var _ irepository.UserLoader = (*UserRepository)(nil)

// userModel maps to the users table and contains GORM tags.
// We keep domain.User free of GORM tags per hexagonal architecture.
type userModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name      string    `gorm:"type:varchar(255);not null"`
	Email     string    `gorm:"type:varchar(255);uniqueIndex;not null"`
	Password  string    `gorm:"type:varchar(255);not null"`
	Role      string    `gorm:"type:varchar(255);not null"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (userModel) TableName() string {
	return "users"
}

// toDomain converts userModel to domain.User
func (m *userModel) toDomain() *domain.User {
	if m == nil {
		return nil
	}
	return &domain.User{
		ID:       m.ID,
		Name:     m.Name,
		Email:    m.Email,
		Password: m.Password,
		Role:     m.Role,
		Created:  m.CreatedAt,
		Updated:  m.UpdatedAt,
	}
}

// fromDomain creates a userModel from domain.User. It does not overwrite ID when it's zero.
func fromDomain(u *domain.User) *userModel {
	m := &userModel{
		Name:     u.Name,
		Email:    u.Email,
		Password: u.Password,
		Role:     u.Role,
	}
	if u.ID != uuid.Nil {
		m.ID = u.ID
	}
	return m
}

// UserRepository is the GORM implementation of UserLoader
type UserRepository struct {
	db *gorm.DB
}

// NewGormUserRepository creates a new repository using GORM DB
func NewGormUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// InsertUser inserts a user and returns the created id as string
func (r *UserRepository) InsertUser(ctx context.Context, user *domain.User) (uuid.UUID, error) {
	m := fromDomain(user)
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return uuid.Nil, domain_errors.WrapUnexpectedMsg(err, "failed to insert user")
	}
	return m.ID, nil
}

// GetUserByID fetches a user by UUID
func (r *UserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	m := &userModel{}
	if err := r.db.WithContext(ctx).First(m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain_errors.NewNotFoundMsg("user not found")
		}
		return nil, domain_errors.WrapUnexpectedMsg(err, "failed to fetch user")
	}
	return m.toDomain(), nil
}

// GetUserByEmail fetches a user by email
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	m := &userModel{}
	if err := r.db.WithContext(ctx).First(m, "email = ?", email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain_errors.NewNotFoundMsg("user not found")
		}
		return nil, domain_errors.WrapUnexpectedMsg(err, "failed to fetch user")
	}
	return m.toDomain(), nil
}

// UpdateUser updates mutable fields of a user
func (r *UserRepository) UpdateUser(ctx context.Context, user *domain.User) error {
	m := fromDomain(user)
	if m.ID == uuid.Nil {
		return domain_errors.NewBadRequestMsg("user ID is required for update")
	}

	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}

	if m.Name != "" {
		updates["name"] = m.Name
	}
	if m.Email != "" {
		updates["email"] = m.Email
	}
	if m.Password != "" {
		updates["password"] = m.Password
	}

	if err := r.db.WithContext(ctx).Model(&userModel{}).Where("id = ?", m.ID).Updates(updates).Error; err != nil {
		return domain_errors.WrapUnexpectedMsg(err, "failed to update user")
	}
	return nil
}

// ListUsers returns all users ordered by created_at
func (r *UserRepository) ListUsers(ctx context.Context) ([]*domain.User, error) {
	var models []userModel
	if err := r.db.WithContext(ctx).Order("created_at desc").Find(&models).Error; err != nil {
		return nil, domain_errors.WrapUnexpectedMsg(err, "failed to list users")
	}
	users := make([]*domain.User, len(models))
	for i := range models {
		users[i] = models[i].toDomain()
	}
	return users, nil
}

func (r *UserRepository) UpdateRole(ctx context.Context, userID uuid.UUID, role string) error {

	res := r.db.WithContext(ctx).
		Model(&userModel{}).
		Where("id = ?", userID).
		Update("role", role)

	if res.Error != nil {
		return domain_errors.WrapUnexpectedMsg(res.Error, "failed to update user role")
	}

	if res.RowsAffected == 0 {
		return domain_errors.NewNotFoundMsg("user not found")
	}

	return nil
}
