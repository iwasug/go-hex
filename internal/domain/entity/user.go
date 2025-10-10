package entity

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user entity in the domain
type User struct {
	ID        string
	Email     string
	Password  string // hashed password
	Name      string
	IsDeleted bool
	CreatedBy *string
	UpdatedBy *string
	DeletedBy *string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// NewUser creates a new User instance
func NewUser(email, password, name string) *User {
	now := time.Now()
	return &User{
		ID:        uuid.NewString(),
		Email:     email,
		Password:  password,
		Name:      name,
		IsDeleted: false,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// MarkAsDeleted marks the user as soft deleted
func (u *User) MarkAsDeleted(deletedBy string) {
	now := time.Now()
	u.IsDeleted = true
	u.DeletedBy = &deletedBy
	u.DeletedAt = &now
	u.UpdatedAt = now
}

// Restore restores a soft deleted user
func (u *User) Restore() {
	u.IsDeleted = false
	u.DeletedBy = nil
	u.DeletedAt = nil
	u.UpdatedAt = time.Now()
}

// Update updates the user's information
func (u *User) Update(name, email string, updatedBy string) {
	u.Name = name
	u.Email = email
	u.UpdatedBy = &updatedBy
	u.UpdatedAt = time.Now()
}

// UpdatePassword updates the user's password
func (u *User) UpdatePassword(hashedPassword string, updatedBy string) {
	u.Password = hashedPassword
	u.UpdatedBy = &updatedBy
	u.UpdatedAt = time.Now()
}
