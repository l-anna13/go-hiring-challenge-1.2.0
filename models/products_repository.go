package models

import (
	"fmt"

	"gorm.io/gorm"
)

type ProductsRepository struct {
	db *gorm.DB
}

// ProductFilters holds all possible filtering criteria extracted from the request.
type ProductFilters struct {
	CategoryCode string  // Filter by category code
	MaxPrice     float64 // Filter by price less than or equal to this value
}

func NewProductsRepository(db *gorm.DB) *ProductsRepository {
	return &ProductsRepository{
		db: db,
	}
}

func (r *ProductsRepository) GetAllProducts(limit, offset int, filters ProductFilters) ([]ProductModel, int, error) {
	var productModels []ProductModel

	query := r.db.Model(&ProductModel{})

	// Apply Category Filter
	if filters.CategoryCode != "" {
		// This requires a JOIN with the Category table to query by code
		query = query.Joins("Category").Where("categories.code = ?", filters.CategoryCode)
	}

	// Apply Max Price Filter
	if filters.MaxPrice > 0 {
		query = query.Where("products.price <= ?", filters.MaxPrice)
	}

	var count64 int64
	if err := query.Count(&count64).Error; err != nil {
		return nil, 0, fmt.Errorf("database query failed")
	}
	totalCount := int(count64)

	if err := query.Preload("Variants").Limit(limit).Offset(offset).Find(&productModels).Error; err != nil {
		return nil, 0, fmt.Errorf("database query failed")
	}

	return productModels, totalCount, nil
}

func (r *ProductsRepository) GetProductByCode(code string) (*ProductModel, error) {
	var productModel ProductModel

	if err := r.db.Model(&ProductModel{}).Preload("Variants").Where("code = ?", code).Find(&productModel).Error; err != nil {
		return &productModel, fmt.Errorf("database query failed")
	}

	return &productModel, nil
}

func (r *ProductsRepository) GetCategories() ([]Category, error) {
	var categories []Category

	r.db.Model(&Category{}).Find(&categories)
	if err := r.db.Model(&Category{}).Find(&categories).Error; err != nil {
		return categories, fmt.Errorf("database query failed")
	}

	return categories, nil
}

func (r *ProductsRepository) CreateCategory(newCategory Category) error {

	result := r.db.Model(&Category{}).Create(&newCategory)
	return result.Error
}
