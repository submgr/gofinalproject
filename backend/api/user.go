package api

import (
	"net/http"

	"classifieds/database"
	"classifieds/models"

	"github.com/gin-gonic/gin"
)

func GetUserProfile(c *gin.Context) {
	id := c.Param("id")
	var user models.User

	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// не отправляем чувствительную информацию
	user.Password = ""

	c.JSON(http.StatusOK, user)
}

func GetUserAdvertisements(c *gin.Context) {
	id := c.Param("id")
	var ads []models.Advertisement

	if err := database.DB.Preload("Category").Preload("Images").Where("user_id = ?", id).Find(&ads).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch advertisements"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": ads,
	})
}

type UpdateUserRequest struct {
	Name           string `json:"name" binding:"required"`
	Email          string `json:"email" binding:"required,email"`
	Phone          string `json:"phone"`
	CurrentPassword string `json:"currentPassword"`
	NewPassword    string `json:"newPassword"`
}

func UpdateUserProfile(c *gin.Context) {
	id := c.Param("id")
	var user models.User

	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// проверяем обновляет ли пользователь свой профиль
	userID := c.GetUint("user_id")
	if user.ID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to update this profile"})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// обновляем основную информацию
	user.Name = req.Name
	user.Email = req.Email
	user.Phone = req.Phone

	// обрабатываем обновление пароля если он предоставлен
	if req.CurrentPassword != "" && req.NewPassword != "" {
		// проверяем текущий пароль
		if err := user.CheckPassword(req.CurrentPassword); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Current password is incorrect"})
			return
		}

		// устанавливаем новый пароль
		user.Password = req.NewPassword
		if err := user.HashPassword(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}
	}

	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	// не отправляем пароль в ответе
	user.Password = ""
	c.JSON(http.StatusOK, user)
} 