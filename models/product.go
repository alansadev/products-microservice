package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type Product struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;"`
	Name        string    `json:"name" gorm:"unique"`
	Description *string   `json:"description,omitempty"`
	ImageURL    *string   `json:"image_url,omitempty"`
	Price       int64     `json:"price"`
	Stock       int64     `json:"stock" gorm:"default:0"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (product *Product) BeforeCreate(tx *gorm.DB) (err error) {
	product.ID = uuid.New()
	return
}
