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
	// загружаем переменные окружения
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}

	// инициализируем базу данных
	database.InitDB()

	// инициализируем роутер
	r := gin.Default()

	// обслуживаем статические файлы
	r.Static("/static", "../frontend/static")
	r.Static("/storage", "../storage")
	r.LoadHTMLGlob("../frontend/templates/*")

	// маршруты фронтенда
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
	r.GET("/recover-password", func(c *gin.Context) {
		c.HTML(http.StatusOK, "recover-password.html", nil)
	})

	// маршруты api
	apiGroup := r.Group("/api")
	{
		// проверка здоровья
		apiGroup.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// маршруты аутентификации
		auth := apiGroup.Group("/auth")
		auth.POST("/register", api.Register)
		auth.POST("/login", api.Login)
		auth.POST("/recover-password", api.RecoverPassword)
		auth.POST("/reset-password", api.ResetPassword)

		// защищенные маршруты
		protected := apiGroup.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			protected.POST("/advertisements", api.CreateAdvertisement)
			protected.PUT("/advertisements/:id", api.UpdateAdvertisement)
			protected.DELETE("/advertisements/:id", api.DeleteAdvertisement)
			protected.PUT("/users/:id", api.UpdateUserProfile)
		}

		// публичные маршруты
		public := apiGroup.Group("")
		public.GET("/advertisements", api.GetAdvertisements)
		public.GET("/advertisements/:id", api.GetAdvertisement)
		public.GET("/advertisements/:id/contact", api.GetAdvertisementContact)
		public.GET("/users/:id", api.GetUserProfile)
		public.GET("/users/:id/advertisements", api.GetUserAdvertisements)
		public.GET("/categories", func(c *gin.Context) {
			var categories []models.Category
			if err := database.DB.Find(&categories).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories"})
				return
			}
			c.JSON(http.StatusOK, categories)
		})
	}

	// запускаем сервер
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
} 