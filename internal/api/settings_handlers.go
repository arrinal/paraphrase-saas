package api

import (
	"net/http"

	"github.com/arrinal/paraphrase-saas/internal/db"
	"github.com/arrinal/paraphrase-saas/internal/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type UpdateSettingsRequest struct {
	Name            string `json:"name"`
	Email           string `json:"email"`
	CurrentPassword string `json:"currentPassword,omitempty"`
	NewPassword     string `json:"newPassword,omitempty"`
}

func HandleUpdateSettings() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")
		var req UpdateSettingsRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var user models.User
		if err := db.DB.First(&user, userID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		// If changing password, verify current password
		if req.NewPassword != "" {
			if req.CurrentPassword == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "current password required"})
				return
			}

			if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.CurrentPassword)); err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid current password"})
				return
			}

			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
				return
			}
			user.Password = string(hashedPassword)
		}

		// Update user details
		if req.Name != "" {
			user.Name = req.Name
		}
		if req.Email != "" {
			// Check if email is already taken
			var existingUser models.User
			if err := db.DB.Where("email = ? AND id != ?", req.Email, userID).First(&existingUser).Error; err == nil {
				c.JSON(http.StatusConflict, gin.H{"error": "email already in use"})
				return
			}
			user.Email = req.Email
		}

		if err := db.DB.Save(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update settings"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "settings updated successfully",
			"user": gin.H{
				"id":    user.ID,
				"name":  user.Name,
				"email": user.Email,
			},
		})
	}
}
