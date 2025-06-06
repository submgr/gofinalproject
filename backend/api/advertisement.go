package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"classifieds/database"
	"classifieds/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CreateAdvertisementRequest struct {
	Title       string  `form:"title" binding:"required"`
	Description string  `form:"description"`
	Price       float64 `form:"price" binding:"required"`
	Location    string  `form:"location"`
	CategoryID  uint    `form:"category_id" binding:"required"`
}

type UpdateAdvertisementRequest struct {
	Title       string  `form:"title" binding:"required"`
	Description string  `form:"description"`
	Price       float64 `form:"price" binding:"required"`
	Location    string  `form:"location"`
	CategoryID  uint    `form:"category_id" binding:"required"`
	DeletedImages string `form:"deleted_images"`
}

func CreateAdvertisement(c *gin.Context) {
	var req CreateAdvertisementRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("user_id")
	ad := models.Advertisement{
		Title:       req.Title,
		Description: req.Description,
		Price:       req.Price,
		Location:    req.Location,
		CategoryID:  req.CategoryID,
		UserID:      userID,
		Status:      "active",
	}

	// обрабатываем загрузку изображения
	form, err := c.MultipartForm()
	if err != nil {
		log.Printf("Error parsing multipart form: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form data: " + err.Error()})
		return
	}

	log.Printf("Form files: %+v", form.File)
	if form.File != nil && len(form.File["images"]) > 0 {
		log.Printf("Number of images received: %d", len(form.File["images"]))
		uploadDir := "../storage/uploads"
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			log.Printf("Error creating upload directory: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
			return
		}

		for _, file := range form.File["images"] {
			log.Printf("Processing file: %s, size: %d", file.Filename, file.Size)
			// генерируем уникальное имя файла
			filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(file.Filename))
			filepath := filepath.Join(uploadDir, filename)

			// сохраняем файл
			if err := c.SaveUploadedFile(file, filepath); err != nil {
				log.Printf("Error saving file: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image: " + err.Error()})
				return
			}
			log.Printf("File saved successfully at: %s", filepath)

			// создаем запись изображения
			image := models.Image{
				URL:             "/storage/uploads/" + filename,
				AdvertisementID: ad.ID,
			}
			ad.Images = append(ad.Images, image)
		}
	} else {
		log.Printf("No images found in form")
	}

	if err := database.DB.Create(&ad).Error; err != nil {
		log.Printf("Error creating advertisement: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create advertisement: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, ad)
}

func GetAdvertisements(c *gin.Context) {
	var ads []models.Advertisement
	query := database.DB.Preload("Category").Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, name, email")
	}).Preload("Images")

	// применяем фильтры
	if categoryID := c.Query("category_id"); categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}
	if minPrice := c.Query("min_price"); minPrice != "" {
		query = query.Where("price >= ?", minPrice)
	}
	if maxPrice := c.Query("max_price"); maxPrice != "" {
		query = query.Where("price <= ?", maxPrice)
	}
	if location := c.Query("location"); location != "" {
		query = query.Where("location LIKE ?", "%"+location+"%")
	}
	if search := c.Query("search"); search != "" {
		query = query.Where("title LIKE ? OR description LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if err := query.Find(&ads).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch advertisements"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": ads,
	})
}

func GetAdvertisement(c *gin.Context) {
	id := c.Param("id")
	var ad models.Advertisement
	if err := database.DB.Preload("Category").Preload("User").Preload("Images").Where("id = ?", id).First(&ad).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Advertisement not found"})
		return
	}
	c.JSON(http.StatusOK, ad)
}

func UpdateAdvertisement(c *gin.Context) {
	id := c.Param("id")
	var ad models.Advertisement

	if err := database.DB.First(&ad, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Advertisement not found"})
		return
	}

	// проверяем является ли пользователь владельцем
	userID := c.GetUint("user_id")
	if ad.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to update this advertisement"})
		return
	}

	var req UpdateAdvertisementRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// обрабатываем удаленные изображения
	if req.DeletedImages != "" {
		var deletedImageIds []uint
		if err := json.Unmarshal([]byte(req.DeletedImages), &deletedImageIds); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deleted images format"})
			return
		}

		// удаляем изображения из хранилища и базы данных
		for _, imageID := range deletedImageIds {
			var image models.Image
			if err := database.DB.First(&image, imageID).Error; err != nil {
				continue
			}
			// удаляем файл из хранилища
			if err := os.Remove("." + image.URL); err != nil {
				log.Printf("Error deleting image file: %v", err)
			}
			// удаляем из базы данных
			database.DB.Delete(&image)
		}
	}

	// обновляем объявление
	ad.Title = req.Title
	ad.Description = req.Description
	ad.Price = req.Price
	ad.Location = req.Location
	ad.CategoryID = req.CategoryID

	// обрабатываем новые изображения
	form, err := c.MultipartForm()
	if err == nil {
		if files := form.File["images"]; len(files) > 0 {
			uploadDir := "../storage/uploads"
			if err := os.MkdirAll(uploadDir, 0755); err != nil {
				log.Printf("Error creating upload directory: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
				return
			}

			for _, file := range files {
				log.Printf("Processing file: %s, size: %d", file.Filename, file.Size)
				// генерируем уникальное имя файла
				filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(file.Filename))
				filepath := filepath.Join(uploadDir, filename)

				// сохраняем файл
				if err := c.SaveUploadedFile(file, filepath); err != nil {
					log.Printf("Error saving file: %v", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image: " + err.Error()})
					return
				}
				log.Printf("File saved successfully at: %s", filepath)

				// создаем запись изображения
				image := models.Image{
					URL:             "/storage/uploads/" + filename,
					AdvertisementID: ad.ID,
				}
				if err := database.DB.Create(&image).Error; err != nil {
					log.Printf("Error creating image record: %v", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image record"})
					return
				}
			}
		}
	}

	if err := database.DB.Save(&ad).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update advertisement"})
		return
	}

	c.JSON(http.StatusOK, ad)
}

func DeleteAdvertisement(c *gin.Context) {
	id := c.Param("id")
	var ad models.Advertisement

	if err := database.DB.First(&ad, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Advertisement not found"})
		return
	}

	// проверяем владельца
	userID := c.GetUint("user_id")
	if ad.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to delete this advertisement"})
		return
	}

	if err := database.DB.Delete(&ad).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete advertisement"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Advertisement deleted successfully"})
}

func GetAdvertisementContact(c *gin.Context) {
	id := c.Param("id")
	var ad models.Advertisement

	if err := database.DB.Preload("User").First(&ad, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Advertisement not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"name": ad.User.Name,
		"email": ad.User.Email,
		"phone": ad.User.Phone,
	})
} 