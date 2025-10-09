package crypto

import (
	"golang.org/x/crypto/bcrypt"
)

// PasswordHasher defines the interface for password hashing operations
type PasswordHasher interface {
	HashPassword(password string) (string, error)
	ComparePassword(hashedPassword, password string) error
}

// BcryptHasher implements PasswordHasher using bcrypt algorithm
type BcryptHasher struct {
	cost int
}

// NewBcryptHasher creates a new BcryptHasher with the specified cost
// If cost is 0, it uses bcrypt.DefaultCost
func NewBcryptHasher(cost int) *BcryptHasher {
	if cost == 0 {
		cost = bcrypt.DefaultCost
	}
	return &BcryptHasher{
		cost: cost,
	}
}

// HashPassword hashes a plain text password using bcrypt
func (b *BcryptHasher) HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), b.cost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// ComparePassword compares a hashed password with a plain text password
// Returns nil if they match, error otherwise
func (b *BcryptHasher) ComparePassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
