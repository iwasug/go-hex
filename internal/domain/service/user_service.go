package service

import (
	"context"
	"errors"
	"fmt"

	"golang-chi-hexagonal-base/internal/domain/entity"
	"golang-chi-hexagonal-base/internal/domain/repository"
)

var (
	// ErrUserNotFound is returned when user is not found
	ErrUserNotFound = errors.New("user not found")
	// ErrRoleNotFound is returned when role is not found
	ErrRoleNotFound = errors.New("role not found")
	// ErrRoleAlreadyAssigned is returned when role is already assigned to user
	ErrRoleAlreadyAssigned = errors.New("role already assigned to user")
)

// UserService handles user management operations
type UserService struct {
	userRepo repository.UserRepository
	roleRepo repository.RoleRepository
}

// NewUserService creates a new UserService instance
func NewUserService(
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
) *UserService {
	return &UserService{
		userRepo: userRepo,
		roleRepo: roleRepo,
	}
}

// ListUsersInput represents the input for listing users
type ListUsersInput struct {
	Search         string
	OrderBy        string
	Order          string
	Page           int
	Limit          int
	IncludeDeleted bool
}

// ListUsersResult represents the result of listing users
type ListUsersResult struct {
	Users      []*entity.User
	Total      int
	Page       int
	Limit      int
	TotalPages int
}

// GetByID retrieves a user by ID
func (s *UserService) GetByID(ctx context.Context, id string) (*entity.User, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// List retrieves a list of users with search, sorting, and pagination
func (s *UserService) List(ctx context.Context, input ListUsersInput) (*ListUsersResult, error) {
	// Set defaults
	if input.Page < 1 {
		input.Page = 1
	}
	if input.Limit < 1 {
		input.Limit = 10
	}
	if input.Limit > 100 {
		input.Limit = 100
	}
	if input.OrderBy == "" {
		input.OrderBy = "created_at"
	}
	if input.Order == "" {
		input.Order = "desc"
	}

	// Calculate offset
	offset := (input.Page - 1) * input.Limit

	// Build repository params
	params := repository.ListUsersParams{
		Search:         input.Search,
		OrderBy:        input.OrderBy,
		Order:          input.Order,
		Limit:          input.Limit,
		Offset:         offset,
		IncludeDeleted: input.IncludeDeleted,
	}

	// Get users from repository
	users, total, err := s.userRepo.List(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	// Calculate total pages
	totalPages := (total + input.Limit - 1) / input.Limit

	return &ListUsersResult{
		Users:      users,
		Total:      total,
		Page:       input.Page,
		Limit:      input.Limit,
		TotalPages: totalPages,
	}, nil
}

// UpdateUserInput represents the input for updating a user
type UpdateUserInput struct {
	Name  string
	Email string
}

// Update updates a user's information
func (s *UserService) Update(ctx context.Context, id string, input UpdateUserInput, updatedBy string) (*entity.User, error) {
	// Get existing user
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Check if email is being changed and if it's already in use
	if input.Email != user.Email {
		existingUser, err := s.userRepo.FindByEmail(ctx, input.Email)
		if err == nil && existingUser != nil && existingUser.ID != id {
			return nil, ErrEmailAlreadyExists
		}
	}

	// Update user entity
	user.Update(input.Name, input.Email, updatedBy)

	// Save to repository
	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// Delete soft deletes a user
func (s *UserService) Delete(ctx context.Context, id string, deletedBy string) error {
	// Check if user exists
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return ErrUserNotFound
	}

	// Check if already deleted
	if user.IsDeleted {
		return ErrUserNotFound
	}

	// Soft delete user
	err = s.userRepo.Delete(ctx, id, deletedBy)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// Restore restores a soft deleted user
func (s *UserService) Restore(ctx context.Context, id string) error {
	// Restore user
	err := s.userRepo.Restore(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to restore user: %w", err)
	}

	return nil
}

// GetUserRoles retrieves all roles assigned to a user
func (s *UserService) GetUserRoles(ctx context.Context, userID string) ([]*entity.Role, error) {
	// Check if user exists
	_, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Get user roles
	roles, err := s.userRepo.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	return roles, nil
}

// AssignRole assigns a role to a user
func (s *UserService) AssignRole(ctx context.Context, userID, roleID string) error {
	// Check if user exists
	_, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}

	// Check if role exists
	_, err = s.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		return ErrRoleNotFound
	}

	// Check if role is already assigned
	roles, err := s.userRepo.GetUserRoles(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user roles: %w", err)
	}

	for _, role := range roles {
		if role.ID == roleID {
			return ErrRoleAlreadyAssigned
		}
	}

	// Assign role
	err = s.userRepo.AssignRole(ctx, userID, roleID)
	if err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	return nil
}

// RemoveRole removes a role from a user
func (s *UserService) RemoveRole(ctx context.Context, userID, roleID string) error {
	// Check if user exists
	_, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}

	// Check if role exists
	_, err = s.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		return ErrRoleNotFound
	}

	// Remove role
	err = s.userRepo.RemoveRole(ctx, userID, roleID)
	if err != nil {
		return fmt.Errorf("failed to remove role: %w", err)
	}

	return nil
}
