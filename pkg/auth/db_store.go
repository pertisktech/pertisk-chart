package auth

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// DBUserStore implements UserStore using a database (SQLite or PostgreSQL)
type DBUserStore struct {
	db *gorm.DB
}

// NewDBUserStore creates a new database-backed user store
func NewDBUserStore(db *gorm.DB) (*DBUserStore, error) {
	store := &DBUserStore{db: db}

	// Auto-migrate the schema
	if err := store.migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return store, nil
}

// migrate runs database migrations
func (s *DBUserStore) migrate() error {
	return s.db.AutoMigrate(&User{})
}

// CreateUser creates a new user
func (s *DBUserStore) CreateUser(user *User) error {
	// Check if username already exists
	var existingUser User
	if err := s.db.Where("username = ?", user.Username).First(&existingUser).Error; err == nil {
		return fmt.Errorf("username already exists")
	}

	// Check if email already exists
	if err := s.db.Where("email = ?", user.Email).First(&existingUser).Error; err == nil {
		return fmt.Errorf("email already exists")
	}

	// Set timestamps
	now := time.Now()
	if user.CreatedAt.IsZero() {
		user.CreatedAt = now
	}
	if user.UpdatedAt.IsZero() {
		user.UpdatedAt = now
	}

	return s.db.Create(user).Error
}

// GetUserByUsername retrieves a user by username
func (s *DBUserStore) GetUserByUsername(username string) (*User, error) {
	var user User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail retrieves a user by email
func (s *DBUserStore) GetUserByEmail(email string) (*User, error) {
	var user User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByID retrieves a user by ID
func (s *DBUserStore) GetUserByID(id string) (*User, error) {
	var user User
	if err := s.db.First(&user, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// UpdateUser updates a user
func (s *DBUserStore) UpdateUser(user *User) error {
	user.UpdatedAt = time.Now()
	return s.db.Save(user).Error
}

// DeleteUser deletes a user
func (s *DBUserStore) DeleteUser(id string) error {
	result := s.db.Delete(&User{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

// ListUsers returns all users
func (s *DBUserStore) ListUsers() ([]*User, error) {
	var users []*User
	if err := s.db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

