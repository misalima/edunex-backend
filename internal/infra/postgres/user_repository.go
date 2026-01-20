package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/misalima/edunex-backend/internal/core/domain"
	"github.com/misalima/edunex-backend/internal/core/interfaces/irepository"
	"gorm.io/gorm"
)

// ensure GormUserRepository implements the UserLoader interface
var _ irepository.UserLoader = (*GormUserRepository)(nil)

// userModel maps to the users table and contains GORM tags.
// We keep domain.User free of GORM tags per hexagonal architecture.
type userModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name      string    `gorm:"type:varchar(255);not null"`
	Email     string    `gorm:"type:varchar(255);uniqueIndex;not null"`
	Password  string    `gorm:"type:varchar(255);not null"`
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
	}
	if u.ID != uuid.Nil {
		m.ID = u.ID
	}
	return m
}

// GormUserRepository is the GORM implementation of UserLoader
type GormUserRepository struct {
	db *gorm.DB
}

// NewGormUserRepository creates a new repository using GORM DB
func NewGormUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

// InsertUser inserts a user and returns the created id as string
func (r *GormUserRepository) InsertUser(ctx context.Context, user *domain.User) (string, error) {
	m := fromDomain(user)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return "", err
	}
	return m.ID.String(), nil
}

// GetUserByID fetches a user by UUID
func (r *GormUserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	m := &userModel{}
	if err := r.db.WithContext(ctx).First(m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return m.toDomain(), nil
}

// GetUserByEmail fetches a user by email
func (r *GormUserRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	m := &userModel{}
	if err := r.db.WithContext(ctx).First(m, "email = ?", email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return m.toDomain(), nil
}

// UpdateUser updates mutable fields of a user
func (r *GormUserRepository) UpdateUser(ctx context.Context, user *domain.User) error {
	m := fromDomain(user)
	if m.ID == uuid.Nil {
		return errors.New("user ID is required")
	}
	updates := map[string]interface{}{
		"name":       m.Name,
		"email":      m.Email,
		"password":   m.Password,
		"updated_at": time.Now(),
	}
	if err := r.db.WithContext(ctx).Model(&userModel{}).Where("id = ?", m.ID).Updates(updates).Error; err != nil {
		return err
	}
	return nil
}

// ListUsers returns all users ordered by created_at
func (r *GormUserRepository) ListUsers(ctx context.Context) ([]*domain.User, error) {
	var models []userModel
	if err := r.db.WithContext(ctx).Order("created_at desc").Find(&models).Error; err != nil {
		return nil, err
	}
	users := make([]*domain.User, len(models))
	for i := range models {
		users[i] = models[i].toDomain()
	}
	return users, nil
}
