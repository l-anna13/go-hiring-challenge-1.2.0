package models

// Category represents a product category in the catalog.
// It includes a code and unique name.
type Category struct {
	ID       uint           `gorm:"primaryKey"`
	Code     string         `gorm:"not null"`
	Name     string         `gorm:"not null"`
	Products []ProductModel `gorm:"foreignKey:CategoryID"`
}

func (c *Category) TableName() string {
	return "categories"
}
