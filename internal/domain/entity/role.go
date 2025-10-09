package entity

import "time"

// Role represents a role entity in the domain
type Role struct {
	ID          string
	Name        string
	Description string
	IsDeleted   bool
	CreatedBy   *string
	UpdatedBy   *string
	DeletedBy   *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

// NewRole creates a new Role instance
func NewRole(name, description string, createdBy string) *Role {
	now := time.Now()
	return &Role{
		Name:        name,
		Description: description,
		IsDeleted:   false,
		CreatedBy:   &createdBy,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// MarkAsDeleted marks the role as soft deleted
func (r *Role) MarkAsDeleted(deletedBy string) {
	now := time.Now()
	r.IsDeleted = true
	r.DeletedBy = &deletedBy
	r.DeletedAt = &now
	r.UpdatedAt = now
}

// Restore restores a soft deleted role
func (r *Role) Restore() {
	r.IsDeleted = false
	r.DeletedBy = nil
	r.DeletedAt = nil
	r.UpdatedAt = time.Now()
}

// Update updates the role's information
func (r *Role) Update(name, description string, updatedBy string) {
	r.Name = name
	r.Description = description
	r.UpdatedBy = &updatedBy
	r.UpdatedAt = time.Now()
}
