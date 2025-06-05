package main

import (
	"log"
	"net/http"
	"os"

	"classifieds/api"
	"classifieds/database"
	"classifieds/middleware"
	"classifieds/models"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Initialize database
	database.InitDB()

	// Initialize router
	r := gin.Default()

	// Serve static files
	r.Static("/static", "../frontend/static")
	r.Static("/storage", "../storage")
	r.LoadHTMLGlob("../frontend/templates/*")

	// Frontend routes
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Доска объявлений",
		})
	})
	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", nil)
	})
	r.GET("/register", func(c *gin.Context) {
		c.HTML(http.StatusOK, "register.html", nil)
	})
	r.GET("/create-ad", func(c *gin.Context) {
		c.HTML(http.StatusOK, "create-ad.html", nil)
	})
	r.GET("/advertisements/:id/edit", func(c *gin.Context) {
		c.HTML(http.StatusOK, "edit-ad.html", nil)
	})
	r.GET("/advertisements/:id", func(c *gin.Context) {
		c.HTML(http.StatusOK, "advertisement.html", nil)
	})
	r.GET("/users/:id", func(c *gin.Context) {
		c.HTML(http.StatusOK, "profile.html", nil)
	})
	r.GET("/users/:id/edit", func(c *gin.Context) {
		c.HTML(http.StatusOK, "edit-profile.html", nil)
	})

	// API routes
	apiGroup := r.Group("/api")
	{
		// Health check
		apiGroup.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// Auth routes
		apiGroup.POST("/auth/register", api.Register)
		apiGroup.POST("/auth/login", api.Login)

		// Protected routes
		protected := apiGroup.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			protected.POST("/advertisements", api.CreateAdvertisement)
			protected.PUT("/advertisements/:id", api.UpdateAdvertisement)
			protected.DELETE("/advertisements/:id", api.DeleteAdvertisement)
			protected.PUT("/users/:id", api.UpdateUserProfile)
		}

		// Public routes
		apiGroup.GET("/advertisements", api.GetAdvertisements)
		apiGroup.GET("/advertisements/:id", api.GetAdvertisement)
		apiGroup.GET("/advertisements/:id/contact", api.GetAdvertisementContact)
		apiGroup.GET("/users/:id", api.GetUserProfile)
		apiGroup.GET("/users/:id/advertisements", api.GetUserAdvertisements)
		apiGroup.GET("/categories", func(c *gin.Context) {
			var categories []models.Category
			if err := database.DB.Find(&categories).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories"})
				return
			}
			c.JSON(http.StatusOK, categories)
		})
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
} 