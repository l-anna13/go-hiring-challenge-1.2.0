package catalog

import (
	"fmt"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/shopspring/decimal"
)

// getProductCodeFromURL extracts the dynamic segment from a URL path.
func getProductCodeFromURL(r *http.Request) (string, error) {
	// Assuming the URL structure is exactly /catalog/{code}
	p := path.Clean(r.URL.Path)
	parts := strings.Split(p, "/")

	if len(parts) < 3 {
		return "", fmt.Errorf("invalid URL structure")
	}

	code := parts[len(parts)-1]
	if code == "catalog" { // Handles trailing slash if path.Clean didn't remove it ideally
		return "", fmt.Errorf("product code is missing")
	}

	return code, nil
}

// applyPriceInheritanceAndMap transforms the ProductModel into the public ProductDetail DTO,
// applying the price inheritance rule for variants.
func applyPriceInheritanceAndMap(p *models.ProductModel) *models.ProductDetail {
	variants := make([]models.VariantDetail, len(p.Variants))

	var d decimal.Decimal //zero value

	for i, v := range p.Variants {
		finalPrice := v.Price

		// If variant price is 0.0, inherit the product's base price.
		if finalPrice.LessThanOrEqual(d) {
			finalPrice = p.Price
		}

		variants[i] = models.VariantDetail{
			ProductID: v.ProductID,
			Name:      v.Name,
			SKU:       v.SKU,
			Price:     finalPrice,
		}
	}

	return &models.ProductDetail{
		Code:     p.Code,
		Price:    p.Price,
		Category: p.Category.Name,
		Variants: variants,
	}
}

// moved from handler for easy reuse
func getLimitOffset(r *http.Request) (limit, offset, page int) {
	// Set safe defaults
	const defaultPage = 1
	const defaultLimit = 10

	// Extract 'page' from query string
	pageStr := r.URL.Query().Get("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = defaultPage
	}

	// Extract 'limit' (or 'pageSize') from query string
	limitStr := r.URL.Query().Get("limit")
	limit, err = strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 { // Max limit to prevent database overload
		limit = defaultLimit
	}

	// Calculate the offset required by the database
	offset = (page - 1) * limit
	return limit, offset, page
}
