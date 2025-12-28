package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User represents a user account
type User struct {
	ID           string    `json:"id" gorm:"primaryKey;type:varchar(255)"`
	Username     string    `json:"username" gorm:"uniqueIndex;not null;type:varchar(255)"`
	Email        string    `json:"email" gorm:"uniqueIndex;not null;type:varchar(255)"`
	PasswordHash string    `json:"-" gorm:"not null;type:text"` // Never serialize password hash
	IsAdmin      bool      `json:"is_admin" gorm:"default:false"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// UserStore interface for user storage
type UserStore interface {
	CreateUser(user *User) error
	GetUserByUsername(username string) (*User, error)
	GetUserByEmail(email string) (*User, error)
	GetUserByID(id string) (*User, error)
	UpdateUser(user *User) error
	DeleteUser(id string) error
	ListUsers() ([]*User, error)
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPassword compares a password with a hash
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateUserID generates a unique user ID
func GenerateUserID() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// ValidateUser validates user input
func ValidateUser(username, email, password string) error {
	if username == "" {
		return errors.New("username is required")
	}
	if len(username) < 3 {
		return errors.New("username must be at least 3 characters")
	}
	if email == "" {
		return errors.New("email is required")
	}
	if password == "" {
		return errors.New("password is required")
	}
	if len(password) < 6 {
		return errors.New("password must be at least 6 characters")
	}
	return nil
}

