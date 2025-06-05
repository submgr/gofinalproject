package models

import (
	"time"

	"gorm.io/gorm"
)

type Image struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	URL             string         `json:"url"`
	AdvertisementID uint           `json:"advertisement_id"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
} 