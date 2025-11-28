package models

import (
	"github.com/shopspring/decimal"
)

// Product represents a product in the catalog.
// It includes a unique code and a price.
type ProductModel struct {
	ID         uint            `gorm:"primaryKey"`
	Code       string          `gorm:"uniqueIndex;not null"`
	Price      decimal.Decimal `gorm:"type:decimal(10,2);not null"`
	CategoryID uint
	Variants   []Variant `gorm:"foreignKey:ProductID"`
	Category   Category
}

type Product struct {
	Code     string          `json:"code"`
	Price    decimal.Decimal `json:"price"`
	Category Category        `json:"category"`
	Variants []Variant       `json:"variants"`
}

type ProductDetail struct {
	Code     string          `json:"code"`
	Price    decimal.Decimal `json:"base_price"`
	Category string          `json:"category"`
	Variants []VariantDetail `json:"variants"`
}

// toProduct method to map the internal model to the public API type.
func (p ProductModel) ToProduct() Product {
	return Product{
		Code:     p.Code,
		Price:    p.Price,
		Variants: p.Variants,
		Category: p.Category,
	}
}

func (p *ProductModel) TableName() string {
	return "products"
}
