package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/misalima/edunex-backend/internal/core/domain"
)

// UserModel maps to the users table and contains GORM tags.
type UserModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name      string    `gorm:"type:varchar(255);not null"`
	Email     string    `gorm:"type:varchar(255);uniqueIndex;not null"`
	Role      string    `gorm:"type:varchar(255);not null"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (UserModel) TableName() string {
	return "users"
}

// ToDomain converts UserModel to domain.User
func (m *UserModel) ToDomain() *domain.User {
	if m == nil {
		return nil
	}
	return &domain.User{
		ID:      m.ID,
		Name:    m.Name,
		Email:   m.Email,
		Role:    m.Role,
		Created: m.CreatedAt,
		Updated: m.UpdatedAt,
	}
}

// FromDomainUser creates a UserModel from domain.User. It does not overwrite ID when it's zero.
func FromDomainUser(u *domain.User) *UserModel {
	if u == nil {
		return nil
	}
	m := &UserModel{
		Name:  u.Name,
		Email: u.Email,
		Role:  u.Role,
	}
	if u.ID != uuid.Nil {
		m.ID = u.ID
	}
	return m
}
