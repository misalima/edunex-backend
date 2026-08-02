package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/misalima/edunex-backend/internal/core/domain"
	"github.com/misalima/edunex-backend/internal/core/domain_errors"
	"github.com/misalima/edunex-backend/internal/core/interfaces/secondary"
	"github.com/misalima/edunex-backend/internal/infra/postgres/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ensure UserRepository implements the UserLoader interface
var _ secondary.UserLoader = (*UserRepository)(nil)

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
	m := models.FromDomainUser(user)
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
	m := &models.UserModel{}
	if err := r.db.WithContext(ctx).First(m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain_errors.NewNotFoundMsg("user not found")
		}
		return nil, domain_errors.WrapUnexpectedMsg(err, "failed to fetch user")
	}
	return m.ToDomain(), nil
}

// GetUserByEmail fetches a user by email
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	m := &models.UserModel{}
	if err := r.db.WithContext(ctx).First(m, "email = ?", email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain_errors.NewNotFoundMsg("user not found")
		}
		return nil, domain_errors.WrapUnexpectedMsg(err, "failed to fetch user")
	}
	return m.ToDomain(), nil
}

// UpdateUser updates mutable fields of a user
func (r *UserRepository) UpdateUser(ctx context.Context, user *domain.User) error {
	m := models.FromDomainUser(user)
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
	if err := r.db.WithContext(ctx).Model(&models.UserModel{}).Where("id = ?", m.ID).Updates(updates).Error; err != nil {
		return domain_errors.WrapUnexpectedMsg(err, "failed to update user")
	}
	return nil
}

// ListUsers returns all users ordered by created_at
func (r *UserRepository) ListUsers(ctx context.Context) ([]*domain.User, error) {
	var dbModels []models.UserModel
	if err := r.db.WithContext(ctx).Order("created_at desc").Find(&dbModels).Error; err != nil {
		return nil, domain_errors.WrapUnexpectedMsg(err, "failed to list users")
	}
	users := make([]*domain.User, len(dbModels))
	for i := range dbModels {
		users[i] = dbModels[i].ToDomain()
	}
	return users, nil
}

func (r *UserRepository) UpdateRole(ctx context.Context, userID uuid.UUID, role string) error {

	res := r.db.WithContext(ctx).
		Model(&models.UserModel{}).
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

// Exists checks if a user exists by UUID
func (r *UserRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.UserModel{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, domain_errors.WrapUnexpectedMsg(err, "failed to check user existence")
	}
	return count > 0, nil
}

// UpsertUser inserts or updates a user on ID conflict
func (r *UserRepository) UpsertUser(ctx context.Context, user *domain.User) error {
	m := models.FromDomainUser(user)
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	if m.Role == "" {
		m.Role = "coordinator"
	}
	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "email", "updated_at"}),
	}).Create(m).Error
	if err != nil {
		return domain_errors.WrapUnexpectedMsg(err, "failed to upsert user")
	}
	return nil
}


