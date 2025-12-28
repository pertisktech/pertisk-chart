package api

import (
	"bytes"
	"crypto/sha256"
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
	"github.com/pertisk-tech/pertisk-chart/pkg/chart"
	"github.com/pertisk-tech/pertisk-chart/pkg/storage"
	"gopkg.in/yaml.v3"
)

// Config holds server configuration
type Config struct {
	Port          string
	EnableMetrics bool
	Debug         bool
}

// Server represents the API server
type Server struct {
	storage storage.Storage
	config  *Config
	router  *gin.Engine
}

// NewServer creates a new API server
func NewServer(store storage.Storage, config *Config) *Server {
	if !config.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	server := &Server{
		storage: store,
		config:  config,
		router:  gin.Default(),
	}

	server.setupRoutes()
	return server
}

// Start starts the HTTP server
func (s *Server) Start(addr string) error {
	return http.ListenAndServe(addr, s.router)
}

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	// Serve static files (UI)
	s.router.Static("/static", "./web/static")
	s.router.LoadHTMLGlob("web/templates/*")
	
	// Web UI routes
	s.router.GET("/", s.handleIndex)
	s.router.GET("/charts", s.handleChartsPage)
	
	// API routes
	api := s.router.Group("/api")
	{
		api.GET("/health", s.handleHealth)
		api.GET("/charts", s.handleListCharts)
		api.GET("/charts/:name", s.handleGetChart)
		api.GET("/charts/:name/:version", s.handleGetChartVersion)
		api.POST("/charts", s.handleUploadChart)
		api.DELETE("/charts/:name/:version", s.handleDeleteChart)
	}
	
	// Helm repository index
	s.router.GET("/index.yaml", s.handleIndexYAML)
	
	// Chart download
	s.router.GET("/charts/:name/:version/:filename", s.handleDownloadChart)
}

// handleIndex serves the main UI page
func (s *Server) handleIndex(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": "Pertisk Chart Repository",
	})
}

// handleChartsPage serves the charts listing page
func (s *Server) handleChartsPage(c *gin.Context) {
	charts, err := s.getChartsList()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": err.Error(),
		})
		return
	}
	
	c.HTML(http.StatusOK, "charts.html", gin.H{
		"title":  "Charts",
		"charts": charts,
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
		"saved": true,
		"name":  chartData.Name,
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
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Home        string          `json:"home,omitempty"`
	Icon        string          `json:"icon,omitempty"`
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
				Version:      v.Version,
				Description:  v.Description,
				Home:         ch.Home,
				Icon:         ch.Icon,
				AppVersion:   v.AppVersion,
				URLs:         v.URLs,
				Digest:       v.Digest,
				Created:      v.Created.Format(time.RFC3339),
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

