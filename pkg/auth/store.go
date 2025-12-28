package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FileUserStore implements UserStore using file-based storage
type FileUserStore struct {
	filePath string
	users    map[string]*User
	mu       sync.RWMutex
}

// NewFileUserStore creates a new file-based user store
func NewFileUserStore(dataDir string) (*FileUserStore, error) {
	store := &FileUserStore{
		filePath: filepath.Join(dataDir, "users.json"),
		users:    make(map[string]*User),
	}

	// Create data directory if it doesn't exist
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	// Load existing users
	if err := store.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return store, nil
}

// load reads users from file
func (s *FileUserStore) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return nil
	}

	return json.Unmarshal(data, &s.users)
}

// save writes users to file
func (s *FileUserStore) save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(s.users, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.filePath, data, 0600)
}

// CreateUser creates a new user
func (s *FileUserStore) CreateUser(user *User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if username already exists
	for _, u := range s.users {
		if u.Username == user.Username {
			return fmt.Errorf("username already exists")
		}
		if u.Email == user.Email {
			return fmt.Errorf("email already exists")
		}
	}

	s.users[user.ID] = user
	return s.save()
}

// GetUserByUsername retrieves a user by username
func (s *FileUserStore) GetUserByUsername(username string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, user := range s.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}

// GetUserByEmail retrieves a user by email
func (s *FileUserStore) GetUserByEmail(email string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, user := range s.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}

// GetUserByID retrieves a user by ID
func (s *FileUserStore) GetUserByID(id string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[id]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

// UpdateUser updates a user
func (s *FileUserStore) UpdateUser(user *User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[user.ID]; !exists {
		return fmt.Errorf("user not found")
	}

	user.UpdatedAt = time.Now()
	s.users[user.ID] = user
	return s.save()
}

// DeleteUser deletes a user
func (s *FileUserStore) DeleteUser(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[id]; !exists {
		return fmt.Errorf("user not found")
	}

	delete(s.users, id)
	return s.save()
}

// ListUsers returns all users
func (s *FileUserStore) ListUsers() ([]*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]*User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}
	return users, nil
}

