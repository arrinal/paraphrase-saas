package api

import (
	"github.com/arrinal/paraphrase-saas/internal/config"
	"github.com/arrinal/paraphrase-saas/internal/middleware"
	"github.com/arrinal/paraphrase-saas/internal/services"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, cfg *config.Config) {
	// Initialize services
	openAIService := services.NewOpenAIService(cfg)

	// Auth routes (public)
	auth := r.Group("/api/auth")
	{
		auth.POST("/login", HandleLogin(cfg))
		auth.POST("/register", HandleRegister(cfg))
		auth.GET("/verify", middleware.AuthRequired(cfg), HandleVerifySession(cfg))
		auth.POST("/refresh", middleware.AuthRequired(cfg), HandleRefreshToken(cfg))
	}

	// Protected routes
	api := r.Group("/api")
	api.Use(middleware.AuthRequired(cfg))
	{
		api.POST("/paraphrase", HandleParaphrase(openAIService))
		api.GET("/history", HandleGetHistory())
		api.GET("/languages", HandleGetUsedLanguages())
		api.GET("/stats", HandleGetUserStats())
		api.PUT("/settings", HandleUpdateSettings())
	}
}
