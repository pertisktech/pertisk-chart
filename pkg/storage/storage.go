package storage

import (
	"io"
	"os"
	"path/filepath"
)

// Storage interface for chart storage backends
type Storage interface {
	// ListCharts returns all chart packages in storage
	ListCharts() ([]string, error)
	
	// GetChart retrieves a chart package by name
	GetChart(name string) (io.ReadCloser, error)
	
	// PutChart stores a chart package
	PutChart(name string, reader io.Reader) error
	
	// DeleteChart removes a chart package
	DeleteChart(name string) error
	
	// ChartExists checks if a chart exists
	ChartExists(name string) bool
}

// LocalStorage implements Storage using local filesystem
type LocalStorage struct {
	rootDir string
}

// NewLocalStorage creates a new local storage instance
func NewLocalStorage(rootDir string) (*LocalStorage, error) {
	if err := os.MkdirAll(rootDir, 0755); err != nil {
		return nil, err
	}
	return &LocalStorage{rootDir: rootDir}, nil
}

// ListCharts returns all chart packages
func (s *LocalStorage) ListCharts() ([]string, error) {
	var charts []string
	err := filepath.Walk(s.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".tgz" {
			relPath, err := filepath.Rel(s.rootDir, path)
			if err != nil {
				return err
			}
			charts = append(charts, relPath)
		}
		return nil
	})
	return charts, err
}

// GetChart retrieves a chart package
func (s *LocalStorage) GetChart(name string) (io.ReadCloser, error) {
	path := filepath.Join(s.rootDir, name)
	return os.Open(path)
}

// PutChart stores a chart package
func (s *LocalStorage) PutChart(name string, reader io.Reader) error {
	path := filepath.Join(s.rootDir, name)
	
	// Create directory if needed
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	
	_, err = io.Copy(file, reader)
	return err
}

// DeleteChart removes a chart package
func (s *LocalStorage) DeleteChart(name string) error {
	path := filepath.Join(s.rootDir, name)
	return os.Remove(path)
}

// ChartExists checks if a chart exists
func (s *LocalStorage) ChartExists(name string) bool {
	path := filepath.Join(s.rootDir, name)
	_, err := os.Stat(path)
	return err == nil
}

