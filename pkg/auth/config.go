package auth

import (
	"time"

	"gorm.io/gorm"
)

// AppConfig represents application configuration
type AppConfig struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Key       string    `json:"key" gorm:"uniqueIndex;not null;type:varchar(255)"`
	Value     string    `json:"value" gorm:"type:text"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// ConfigStore interface for configuration storage
type ConfigStore interface {
	GetConfig(key string) (string, error)
	SetConfig(key, value string) error
	GetAllConfig() (map[string]string, error)
}

// DBConfigStore implements ConfigStore using GORM
type DBConfigStore struct {
	db *gorm.DB
}

// NewDBConfigStore creates a new database-backed config store
func NewDBConfigStore(db *gorm.DB) (*DBConfigStore, error) {
	store := &DBConfigStore{db: db}
	
	// Auto-migrate
	if err := db.AutoMigrate(&AppConfig{}); err != nil {
		return nil, err
	}
	
	return store, nil
}

// GetConfig retrieves a configuration value by key
func (s *DBConfigStore) GetConfig(key string) (string, error) {
	var config AppConfig
	result := s.db.Where("key = ?", key).First(&config)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return "", nil // Return empty string if not found
		}
		return "", result.Error
	}
	return config.Value, nil
}

// SetConfig sets a configuration value
func (s *DBConfigStore) SetConfig(key, value string) error {
	var config AppConfig
	result := s.db.Where("key = ?", key).First(&config)
	
	if result.Error == gorm.ErrRecordNotFound {
		// Create new config
		config = AppConfig{
			Key:   key,
			Value: value,
		}
		return s.db.Create(&config).Error
	} else if result.Error != nil {
		return result.Error
	}
	
	// Update existing config
	config.Value = value
	return s.db.Save(&config).Error
}

// GetAllConfig retrieves all configuration values
func (s *DBConfigStore) GetAllConfig() (map[string]string, error) {
	var configs []AppConfig
	if err := s.db.Find(&configs).Error; err != nil {
		return nil, err
	}
	
	result := make(map[string]string)
	for _, config := range configs {
		result[config.Key] = config.Value
	}
	return result, nil
}

