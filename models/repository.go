package models

// Repository Interface (Inferred from usage)
type Repository interface {
	GetAllProducts(limit, offset int, filters ProductFilters) ([]ProductModel, int, error)
	GetProductByCode(code string) (*ProductModel, error)
	GetCategories() ([]Category, error)
	CreateCategory(newCategory Category) error
}
