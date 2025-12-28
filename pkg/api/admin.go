package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pertisk-tech/pertisk-chart/pkg/auth"
)

// AdminMiddleware checks if the user is an admin
func (s *Server) AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// First check authentication
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		tokenString := ""
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		claims, err := auth.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		// Get user to check admin status
		user, err := s.userStore.GetUserByID(claims.UserID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			c.Abort()
			return
		}

		if !user.IsAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("isAdmin", true)

		c.Next()
	}
}

// handleGetConfig retrieves application configuration
func (s *Server) handleGetConfig(c *gin.Context) {
	if s.configStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "config store not available"})
		return
	}

	key := c.Param("key")
	if key == "" {
		// Return all config
		config, err := s.configStore.GetAllConfig()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, config)
		return
	}

	value, err := s.configStore.GetConfig(key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"key": key, "value": value})
}

// handleSetConfig sets application configuration
func (s *Server) handleSetConfig(c *gin.Context) {
	if s.configStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "config store not available"})
		return
	}

	var req struct {
		Key   string `json:"key" binding:"required"`
		Value string `json:"value" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.configStore.SetConfig(req.Key, req.Value); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "configuration updated", "key": req.Key, "value": req.Value})
}

// handleListUsers lists all users (admin only)
func (s *Server) handleListUsers(c *gin.Context) {
	users, err := s.userStore.ListUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

// handleUpdateUser updates a user (admin only)
func (s *Server) handleUpdateUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user ID required"})
		return
	}

	user, err := s.userStore.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var req struct {
		IsAdmin *bool `json:"is_admin"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.IsAdmin != nil {
		user.IsAdmin = *req.IsAdmin
	}

	if err := s.userStore.UpdateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// handleGetDomain returns the configured domain
func (s *Server) handleGetDomain(c *gin.Context) {
	var domain string
	var err error

	if s.configStore != nil {
		domain, err = s.configStore.GetConfig("domain")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// Default to localhost if not set
	if domain == "" {
		domain = "http://localhost:7080"
	}

	c.JSON(http.StatusOK, gin.H{"domain": domain})
}

