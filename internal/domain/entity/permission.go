package entity

import (
	"time"

	"github.com/google/uuid"
)

// Permission represents a permission entity in the domain
type Permission struct {
	ID          string
	Name        string
	Resource    string
	Action      string
	Description string
	CreatedAt   time.Time
}

// NewPermission creates a new Permission instance
func NewPermission(name, resource, action, description string) *Permission {
	return &Permission{
		ID:          uuid.NewString(),
		Name:        name,
		Resource:    resource,
		Action:      action,
		Description: description,
		CreatedAt:   time.Now(),
	}
}
