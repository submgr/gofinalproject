package api

import (
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

	// Handle image upload
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
			// Generate unique filename
			filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(file.Filename))
			filepath := filepath.Join(uploadDir, filename)

			// Save file
			if err := c.SaveUploadedFile(file, filepath); err != nil {
				log.Printf("Error saving file: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image: " + err.Error()})
				return
			}
			log.Printf("File saved successfully at: %s", filepath)

			// Create image record
			image := models.Image{
				URL: "/storage/uploads/" + filename,
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

	// Apply filters
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

	if err := database.DB.Preload("Category").Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, name, email")
	}).Preload("Images").First(&ad, id).Error; err != nil {
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

	// Check ownership
	userID := c.GetUint("user_id")
	if ad.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to update this advertisement"})
		return
	}

	var req CreateAdvertisementRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ad.Title = req.Title
	ad.Description = req.Description
	ad.Price = req.Price
	ad.Location = req.Location
	ad.CategoryID = req.CategoryID

	// Handle image upload
	form, err := c.MultipartForm()
	if err != nil {
		log.Printf("Error parsing multipart form: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form data: " + err.Error()})
		return
	}

	if form.File != nil && len(form.File["images"]) > 0 {
		uploadDir := "../storage/uploads"
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			log.Printf("Error creating upload directory: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
			return
		}

		for _, file := range form.File["images"] {
			// Generate unique filename
			filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(file.Filename))
			filepath := filepath.Join(uploadDir, filename)

			// Save file
			if err := c.SaveUploadedFile(file, filepath); err != nil {
				log.Printf("Error saving file: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image: " + err.Error()})
				return
			}

			// Create image record
			image := models.Image{
				URL: "/storage/uploads/" + filename,
			}
			ad.Images = append(ad.Images, image)
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

	// Check ownership
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