package catalog

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/models"
)

// Metadata holds the pagination context for the API response.
type Metadata struct {
	CurrentPage int `json:"current_page"`
	PageSize    int `json:"page_size"`
	FirstItem   int `json:"first_item"`
	LastItem    int `json:"last_item"`
	TotalItems  int `json:"total_items"`
	TotalPages  int `json:"total_pages"`
}

type Response struct {
	Metadata Metadata         `json:"metadata"`
	Products []models.Product `json:"products"`
}

type CatalogHandler struct {
	repo models.Repository
}

func NewCatalogHandler(r *models.ProductsRepository) *CatalogHandler {
	return &CatalogHandler{
		repo: r,
	}
}

func (h *CatalogHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	// get limit and offset valus
	limit, offset, page := getLimitOffset(r)

	// Category Filter
	var filters models.ProductFilters
	categoryCode := r.URL.Query().Get("category")
	filters.CategoryCode = strings.ToLower(categoryCode)

	// Price Filter
	maxPriceStr := r.URL.Query().Get("max_price")
	if mp, err := strconv.ParseFloat(maxPriceStr, 64); err == nil && mp > 0 {
		filters.MaxPrice = mp
	}
	// Fetch data from the repository (returns internal ProductModel)
	productModels, totalItems, err := h.repo.GetAllProducts(limit, offset, filters)
	if err != nil {
		// Log the internal error details
		log.Printf("ERROR: Failed to retrieve products from repo: %v", err)
		// Return a safe, generic error to the client
		http.Error(w, "Error fetching catalog data", http.StatusInternalServerError)
		return
	}

	// Map data from internal models to public API struct using the helper method
	products := make([]models.Product, len(productModels))
	for i, model := range productModels {
		products[i] = model.ToProduct() // Encapsulated mapping logic
	}

	totalPages := (totalItems + limit - 1) / limit // Calculates total pages correctly

	// Calculate first and last item index for the current page
	firstItem := offset + 1
	lastItem := offset + len(products) // Use actual length, which may be less than limit on the last page

	metadata := Metadata{
		CurrentPage: page,
		PageSize:    limit,
		FirstItem:   firstItem,
		LastItem:    lastItem,
		TotalItems:  totalItems,
		TotalPages:  totalPages,
	}

	// Build the final response wrapper
	response := Response{
		Metadata: metadata,
		Products: products,
	}

	//Use the reusable JSON helper to write the response
	api.RespondWithJSON(w, http.StatusOK, response)
}

func (h *CatalogHandler) HandleProductGet(w http.ResponseWriter, r *http.Request) {
	productCode := r.PathValue("code")

	if productCode == "" {
		http.Error(w, "Product code not found in path", http.StatusBadRequest)
		return
	}

	// Fetch data from repository
	productModel, err := h.repo.GetProductByCode(productCode)
	if err != nil {
		// Check for Not Found error specifically (assuming the repo returns a standard error)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, fmt.Sprintf("Product %s not found.", productCode), http.StatusNotFound)
			return
		}
		log.Printf("ERROR: Failed to fetch product details for %s: %v", productCode, err)
		http.Error(w, "Internal server error retrieving product details.", http.StatusInternalServerError)
		return
	}

	// Apply business logic and map to DTO
	productDetail := applyPriceInheritanceAndMap(productModel)

	//Return the result as JSON
	api.RespondWithJSON(w, http.StatusOK, productDetail)
}

func (h *CatalogHandler) HandleCategoryList(w http.ResponseWriter, r *http.Request) {
	categories, err := h.repo.GetCategories()
	if err != nil {
		// Log the error and return 500
		fmt.Printf("ERROR: Failed to fetch categories: %v\n", err)
		http.Error(w, "Internal server error retrieving categories.", http.StatusInternalServerError)
		return
	}

	// Return the result as JSON
	api.RespondWithJSON(w, http.StatusOK, categories)
}

func (h *CatalogHandler) HandleCategoryCreate(w http.ResponseWriter, r *http.Request) {
	var newCategory models.Category

	// Decode the request body into the newCategory struct.
	// We use a NewDecoder to stream the data, which is more robust.
	err := json.NewDecoder(r.Body).Decode(&newCategory)
	if err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Basic Validation (Check for required fields like Code and Name)
	if newCategory.Code == "" || newCategory.Name == "" {
		http.Error(w, "Code and Name fields are required.", http.StatusBadRequest)
		return
	}
	err = h.repo.CreateCategory(newCategory)
	if err != nil {
		// Log the error and return 500
		fmt.Printf("ERROR: Failed to create category: %v\n", err)
		http.Error(w, "Internal server error while creating category.", http.StatusInternalServerError)
		return
	}
	// Return the result as JSON
	api.RespondWithJSON(w, http.StatusOK, newCategory)
}
