package database

import (
	"classifieds/models"
	"log"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	var err error
	DB, err = gorm.Open(sqlite.Open("classifieds.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// автоматически мигрируем схему
	err = DB.AutoMigrate(&models.User{}, &models.Category{}, &models.Advertisement{}, &models.Image{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// убеждаемся что колонка advertisement_id существует в таблице images
	if !DB.Migrator().HasColumn(&models.Image{}, "AdvertisementID") {
		if err := DB.Migrator().AddColumn(&models.Image{}, "AdvertisementID"); err != nil {
			log.Fatal("Failed to add AdvertisementID column:", err)
		}
	}

	// инициализируем категории если они не существуют
	var count int64
	DB.Model(&models.Category{}).Count(&count)
	if count == 0 {
		categories := []models.Category{
			{Name: "Недвижимость"},
			{Name: "Транспорт"},
			{Name: "Электроника"},
			{Name: "Работа"},
			{Name: "Услуги"},
			{Name: "Одежда и обувь"},
			{Name: "Мебель и интерьер"},
			{Name: "Спорт и отдых"},
			{Name: "Животные"},
			{Name: "Другое"},
		}
		DB.Create(&categories)
	}

	// создаем директорию для загрузок если она не существует
	uploadDir := "../storage/uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Fatal("Failed to create uploads directory:", err)
	}
} 