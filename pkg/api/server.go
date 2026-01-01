package api

import (
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pertisk-tech/pertisk-chart/pkg/auth"
	"github.com/pertisk-tech/pertisk-chart/pkg/chart"
	"github.com/pertisk-tech/pertisk-chart/pkg/storage"
	"github.com/quic-go/quic-go/http3"
	"gopkg.in/yaml.v3"
)

// Config holds server configuration
type Config struct {
	Port          string
	EnableMetrics bool
	Debug         bool
	EnableHTTP3   bool
	TLSCertFile   string
	TLSKeyFile    string
	EnableZstd    bool
	WebDir        string
}

// Server represents the API server
type Server struct {
	storage    storage.Storage
	userStore  auth.UserStore
	configStore auth.ConfigStore
	config      *Config
	router      *gin.Engine
}

// NewServer creates a new API server
func NewServer(store storage.Storage, userStore auth.UserStore, configStore auth.ConfigStore, config *Config) *Server {
	if !config.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	server := &Server{
		storage:     store,
		userStore:   userStore,
		configStore: configStore,
		config:      config,
		router:      gin.Default(),
	}

	server.setupRoutes()
	return server
}

// Start starts the HTTP server (HTTP/1.1, HTTP/2, and optionally HTTP/3)
func (s *Server) Start(addr string) error {
	handler := http.Handler(s.router)

	// Apply zstd compression if enabled
	if s.config.EnableZstd {
		compressedHandler, err := GetCompressedHandler(handler)
		if err != nil {
			log.Printf("Warning: Failed to enable compression: %v", err)
		} else {
			handler = compressedHandler
		}
	}

	// Start HTTP/3 server if enabled and TLS certs are provided
	if s.config.EnableHTTP3 && s.config.TLSCertFile != "" && s.config.TLSKeyFile != "" {
		return s.startHTTP3(addr, handler)
	}

	// Start regular HTTP/1.1 and HTTP/2 server
	return http.ListenAndServe(addr, handler)
}

// startHTTP3 starts an HTTP/3 server (requires TLS certificates)
func (s *Server) startHTTP3(addr string, handler http.Handler) error {
	// Load TLS certificate and key
	cert, err := tls.LoadX509KeyPair(s.config.TLSCertFile, s.config.TLSKeyFile)
	if err != nil {
		return fmt.Errorf("failed to load TLS certificates: %w", err)
	}

	// Configure TLS for HTTP/3
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{"h3", "h2", "http/1.1"},
	}

	// Create HTTP/3 server
	server := &http3.Server{
		Addr:      addr,
		Handler:   handler,
		TLSConfig: tlsConfig,
	}

	log.Printf("Starting HTTP/3 server on %s", addr)
	return server.ListenAndServe()
}

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	// Determine web directory path
	webDir := s.config.WebDir
	if webDir == "" {
		// Default to ./web for development
		webDir = "./web"
	}
	
	// Serve static files (UI assets)
	s.router.Static("/static", filepath.Join(webDir, "static"))

	// API routes
	api := s.router.Group("/api")
	{
		// Public routes
		api.GET("/health", s.handleHealth)
		api.GET("/charts", s.handleListCharts)
		api.GET("/charts/:name", s.handleGetChart)
		api.GET("/charts/:name/:version", s.handleGetChartVersion)
		api.GET("/charts/:name/:version/values", s.handleGetValues)
		api.GET("/charts/:name/:version/values.yaml", s.handleDownloadValues)
		api.GET("/charts/:name/:version/readme", s.handleGetReadme)
		api.GET("/charts/:name/:version/manifests", s.handleGetManifests)
		api.GET("/charts/:name/:version/resources", s.handleGetResources)
		api.GET("/charts/:name/:version/notes", s.handleGetNotes)
		
		// Authentication routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", s.handleRegister)
			auth.POST("/login", s.handleLogin)
			auth.GET("/me", s.AuthMiddleware(), s.handleGetMe)
		}
		
		// Protected routes (require authentication)
		protected := api.Group("")
		protected.Use(s.AuthMiddleware())
		{
			protected.POST("/charts", s.handleUploadChart)
			protected.DELETE("/charts/:name/:version", s.handleDeleteChart)
		}

		// Admin routes (require admin authentication)
		admin := api.Group("/admin")
		admin.Use(s.AdminMiddleware())
		{
			admin.GET("/config", s.handleGetConfig)
			admin.GET("/config/:key", s.handleGetConfig)
			admin.POST("/config", s.handleSetConfig)
			admin.GET("/users", s.handleListUsers)
			admin.PUT("/users/:id", s.handleUpdateUser)
		}

		// Public config endpoint (for domain)
		api.GET("/config/domain", s.handleGetDomain)
	}

	// Helm repository index
	s.router.GET("/index.yaml", s.handleIndexYAML)
	
	// Chart download
	s.router.GET("/charts/:name/:version/:filename", s.handleDownloadChart)
	
	// Catch-all for SPA routing - serve index.html for all non-API routes
	s.router.NoRoute(func(c *gin.Context) {
		// Don't serve index.html for API routes, static files, index.yaml, or chart downloads
		path := c.Request.URL.Path
		
		// Exclude API routes
		if strings.HasPrefix(path, "/api") {
			c.Status(http.StatusNotFound)
			return
		}
		
		// Exclude static files
		if strings.HasPrefix(path, "/static") {
			c.Status(http.StatusNotFound)
			return
		}
		
		// Exclude index.yaml
		if path == "/index.yaml" {
			c.Status(http.StatusNotFound)
			return
		}
		
		// Exclude chart downloads and values endpoints
		// Pattern: /charts/:name/:version/:filename (3 segments) - chart download
		// Pattern: /charts/:name/:version/values (3 segments, last is "values") - values API
		// Pattern: /charts/:name/:version/values.yaml (3 segments, last is "values.yaml") - values download
		// But allow client-side routes like /charts or /charts/:name (1 or 2 segments)
		if strings.HasPrefix(path, "/charts/") {
			// Count path segments after /charts/
			parts := strings.Split(strings.TrimPrefix(path, "/charts/"), "/")
			// If it has 3 segments, check if it's an API endpoint
			if len(parts) == 3 {
				lastPart := parts[2]
				// Exclude chart downloads, values, and readme endpoints
				if lastPart == "values" || lastPart == "values.yaml" || lastPart == "readme" || strings.HasSuffix(lastPart, ".tgz") {
					c.Status(http.StatusNotFound)
					return
				}
			}
		}
		
		// Serve index.html for all other routes (client-side routing)
		webDir := s.config.WebDir
		if webDir == "" {
			webDir = "./web"
		}
		c.File(filepath.Join(webDir, "index.html"))
	})
}

// handleHealth returns server health status
func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

// handleListCharts returns list of all charts
func (s *Server) handleListCharts(c *gin.Context) {
	charts, err := s.getChartsList()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, charts)
}

// handleGetChart returns information about a specific chart
func (s *Server) handleGetChart(c *gin.Context) {
	name := c.Param("name")
	charts, err := s.getChartsList()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var chartInfo *ChartInfo
	for _, ch := range charts {
		if ch.Name == name {
			chartInfo = &ch
			break
		}
	}

	if chartInfo == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "chart not found"})
		return
	}

	c.JSON(http.StatusOK, chartInfo)
}

// handleGetChartVersion returns information about a specific chart version
func (s *Server) handleGetChartVersion(c *gin.Context) {
	name := c.Param("name")
	version := c.Param("version")

	charts, err := s.getChartsList()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, ch := range charts {
		if ch.Name == name {
			for _, v := range ch.Versions {
				if v.Version == version {
					c.JSON(http.StatusOK, v)
					return
				}
			}
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "chart version not found"})
}

// handleUploadChart handles chart package upload
func (s *Server) handleUploadChart(c *gin.Context) {
	file, err := c.FormFile("chart")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no chart file provided"})
		return
	}

	// Validate filename
	if !strings.HasSuffix(file.Filename, ".tgz") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chart file: must be .tgz"})
		return
	}

	// Open uploaded file
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer src.Close()

	// Read file into memory for parsing and storage
	fileData, err := io.ReadAll(src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Parse chart metadata
	chartData, err := chart.ParseChartFromTarball(bytes.NewReader(fileData))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("failed to parse chart: %v", err)})
		return
	}

	// Generate filename: name-version.tgz
	filename := fmt.Sprintf("%s-%s.tgz", chartData.Name, chartData.Version)

	// Store chart
	if err := s.storage.PutChart(filename, bytes.NewReader(fileData)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"saved":   true,
		"name":    chartData.Name,
		"version": chartData.Version,
	})
}

// handleDeleteChart deletes a chart version
func (s *Server) handleDeleteChart(c *gin.Context) {
	name := c.Param("name")
	version := c.Param("version")

	filename := fmt.Sprintf("%s-%s.tgz", name, version)

	if !s.storage.ChartExists(filename) {
		c.JSON(http.StatusNotFound, gin.H{"error": "chart not found"})
		return
	}

	if err := s.storage.DeleteChart(filename); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

// handleDownloadChart downloads a chart package
func (s *Server) handleDownloadChart(c *gin.Context) {
	name := c.Param("name")
	version := c.Param("version")
	filename := fmt.Sprintf("%s-%s.tgz", name, version)

	if !s.storage.ChartExists(filename) {
		c.JSON(http.StatusNotFound, gin.H{"error": "chart not found"})
		return
	}

	reader, err := s.storage.GetChart(filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer reader.Close()

	c.Header("Content-Type", "application/gzip")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.DataFromReader(http.StatusOK, -1, "application/gzip", reader, nil)
}

// handleGetValues returns the default values.yaml as JSON
func (s *Server) handleGetValues(c *gin.Context) {
	name := c.Param("name")
	version := c.Param("version")
	filename := fmt.Sprintf("%s-%s.tgz", name, version)

	if !s.storage.ChartExists(filename) {
		c.JSON(http.StatusNotFound, gin.H{"error": "chart not found"})
		return
	}

	reader, err := s.storage.GetChart(filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer reader.Close()

	// Extract values.yaml
	valuesData, err := chart.ExtractValuesYAML(reader)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("values.yaml not found: %v", err)})
		return
	}

	// Parse YAML to JSON for easier frontend handling
	var valuesMap map[string]interface{}
	if err := yaml.Unmarshal(valuesData, &valuesMap); err != nil {
		// If parsing fails, return raw YAML as string
		c.JSON(http.StatusOK, gin.H{
			"yaml": string(valuesData),
			"raw":   true,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"yaml": string(valuesData),
		"data": valuesMap,
		"raw":  false,
	})
}

// handleDownloadValues downloads the default values.yaml file
func (s *Server) handleDownloadValues(c *gin.Context) {
	name := c.Param("name")
	version := c.Param("version")
	filename := fmt.Sprintf("%s-%s.tgz", name, version)

	if !s.storage.ChartExists(filename) {
		c.JSON(http.StatusNotFound, gin.H{"error": "chart not found"})
		return
	}

	reader, err := s.storage.GetChart(filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer reader.Close()

	// Extract values.yaml
	valuesData, err := chart.ExtractValuesYAML(reader)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("values.yaml not found: %v", err)})
		return
	}

	valuesFilename := fmt.Sprintf("%s-%s-values.yaml", name, version)
	c.Header("Content-Type", "application/x-yaml")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", valuesFilename))
	c.Data(http.StatusOK, "application/x-yaml", valuesData)
}

// handleGetReadme returns the README.md as JSON
func (s *Server) handleGetReadme(c *gin.Context) {
	name := c.Param("name")
	version := c.Param("version")
	filename := fmt.Sprintf("%s-%s.tgz", name, version)

	if !s.storage.ChartExists(filename) {
		c.JSON(http.StatusNotFound, gin.H{"error": "chart not found"})
		return
	}

	reader, err := s.storage.GetChart(filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer reader.Close()

	// Extract README.md
	readmeData, err := chart.ExtractREADME(reader)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("README.md not found: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"readme": string(readmeData),
		"markdown": string(readmeData),
	})
}

// handleGetManifests returns the rendered manifests as YAML
func (s *Server) handleGetManifests(c *gin.Context) {
	name := c.Param("name")
	version := c.Param("version")
	filename := fmt.Sprintf("%s-%s.tgz", name, version)

	if !s.storage.ChartExists(filename) {
		c.JSON(http.StatusNotFound, gin.H{"error": "chart not found"})
		return
	}

	reader, err := s.storage.GetChart(filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer reader.Close()

	// Extract manifests
	manifestsData, err := chart.ExtractManifests(reader)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("manifests not found: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"manifests": string(manifestsData),
		"yaml":      string(manifestsData),
	})
}

// handleGetResources returns the list of Kubernetes resources
func (s *Server) handleGetResources(c *gin.Context) {
	name := c.Param("name")
	version := c.Param("version")
	filename := fmt.Sprintf("%s-%s.tgz", name, version)

	if !s.storage.ChartExists(filename) {
		c.JSON(http.StatusNotFound, gin.H{"error": "chart not found"})
		return
	}

	reader, err := s.storage.GetChart(filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer reader.Close()

	// Extract resources
	resources, err := chart.ExtractResources(reader)
	if err != nil {
		// If it's a "no templates found" error, return empty list instead of 404
		if strings.Contains(err.Error(), "no template files found") {
			c.JSON(http.StatusOK, gin.H{
				"resources": []chart.ResourceInfo{},
				"message":  "This chart does not have template files in the templates/ directory",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to extract resources: %v", err)})
		return
	}

	// Return empty list if no resources could be parsed (templates may contain only Helm syntax)
	c.JSON(http.StatusOK, gin.H{
		"resources": resources,
		"count":     len(resources),
	})
}

// handleGetNotes returns the NOTES.txt file
func (s *Server) handleGetNotes(c *gin.Context) {
	name := c.Param("name")
	version := c.Param("version")
	filename := fmt.Sprintf("%s-%s.tgz", name, version)

	if !s.storage.ChartExists(filename) {
		c.JSON(http.StatusNotFound, gin.H{"error": "chart not found"})
		return
	}

	reader, err := s.storage.GetChart(filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer reader.Close()

	// Extract NOTES.txt
	notesData, err := chart.ExtractNOTES(reader)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("NOTES.txt not found: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"notes": string(notesData),
		"text":  string(notesData),
	})
}

// handleIndexYAML generates and returns the Helm repository index
func (s *Server) handleIndexYAML(c *gin.Context) {
	index, err := s.generateIndex()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "application/x-yaml")
	c.YAML(http.StatusOK, index)
}

// ChartInfo represents chart information for API responses
type ChartInfo struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Home        string             `json:"home,omitempty"`
	Icon        string             `json:"icon,omitempty"`
	Versions    []ChartVersionInfo `json:"versions"`
}

// ChartVersionInfo represents version information
type ChartVersionInfo struct {
	Version     string    `json:"version"`
	AppVersion  string    `json:"appVersion,omitempty"`
	Description string    `json:"description,omitempty"`
	Created     time.Time `json:"created,omitempty"`
	URLs        []string  `json:"urls"`
	Digest      string    `json:"digest,omitempty"`
}

// getChartsList retrieves and organizes all charts
func (s *Server) getChartsList() ([]ChartInfo, error) {
	files, err := s.storage.ListCharts()
	if err != nil {
		return nil, err
	}

	chartsMap := make(map[string]*ChartInfo)

	for _, file := range files {
		name, version, err := chart.ParseChartName(file)
		if err != nil {
			if s.config.Debug {
				log.Printf("Skipping invalid chart file: %s - %v", file, err)
			}
			continue
		}

		// Read chart metadata
		reader, err := s.storage.GetChart(file)
		if err != nil {
			if s.config.Debug {
				log.Printf("Failed to read chart %s: %v", file, err)
			}
			continue
		}

		chartData, err := chart.ParseChartFromTarball(reader)
		reader.Close()
		if err != nil {
			if s.config.Debug {
				log.Printf("Failed to parse chart %s: %v", file, err)
			}
			continue
		}

		// Calculate digest
		digest, err := s.calculateDigest(file)
		if err != nil {
			if s.config.Debug {
				log.Printf("Failed to calculate digest for %s: %v", file, err)
			}
		}

		// Get base URL
		baseURL := s.getBaseURL()
		url := fmt.Sprintf("%s/charts/%s/%s/%s", baseURL, name, version, filepath.Base(file))

		if chartsMap[name] == nil {
			chartsMap[name] = &ChartInfo{
				Name:        chartData.Name,
				Description: chartData.Description,
				Home:        chartData.Home,
				Icon:        chartData.Icon,
				Versions:    []ChartVersionInfo{},
			}
		}

		chartsMap[name].Versions = append(chartsMap[name].Versions, ChartVersionInfo{
			Version:     chartData.Version,
			AppVersion:  chartData.AppVersion,
			Description: chartData.Description,
			Created:     time.Now(), // Could parse from file mod time
			URLs:        []string{url},
			Digest:      digest,
		})
	}

	// Convert map to slice and sort versions
	var charts []ChartInfo
	for _, ch := range chartsMap {
		// Sort versions (newest first)
		sort.Slice(ch.Versions, func(i, j int) bool {
			return ch.Versions[i].Version > ch.Versions[j].Version
		})
		charts = append(charts, *ch)
	}

	// Sort charts by name
	sort.Slice(charts, func(i, j int) bool {
		return charts[i].Name < charts[j].Name
	})

	return charts, nil
}

// generateIndex generates the Helm repository index.yaml
func (s *Server) generateIndex() (*chart.Index, error) {
	charts, err := s.getChartsList()
	if err != nil {
		return nil, err
	}

	index := &chart.Index{
		APIVersion: "v1",
		Generated:  time.Now().UTC().Format(time.RFC3339),
		Entries:    make(map[string][]chart.ChartVersion),
	}

	for _, ch := range charts {
		var versions []chart.ChartVersion
		for _, v := range ch.Versions {
			versions = append(versions, chart.ChartVersion{
				Name:        ch.Name,
				Version:     v.Version,
				Description: v.Description,
				Home:        ch.Home,
				Icon:        ch.Icon,
				AppVersion:  v.AppVersion,
				URLs:        v.URLs,
				Digest:      v.Digest,
				Created:     v.Created.Format(time.RFC3339),
			})
		}
		index.Entries[ch.Name] = versions
	}

	return index, nil
}

// calculateDigest calculates SHA256 digest of a chart file
func (s *Server) calculateDigest(filename string) (string, error) {
	reader, err := s.storage.GetChart(filename)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, reader); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// getBaseURL returns the base URL for chart downloads
func (s *Server) getBaseURL() string {
	// In a real implementation, this could be configured
	// For now, return a relative path (Helm will resolve it relative to the index.yaml URL)
	return ""
}
