package catalog

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mocks and Stubs (To make the code testable) ---

// MockRepository using testify/mock
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) GetProductByCode(code string) (*models.ProductModel, error) {
	args := m.Called(code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProductModel), args.Error(1)
}

func (m *MockRepository) GetAllProducts(limit, offset int, filters models.ProductFilters) ([]models.ProductModel, int, error) {
	args := m.Called(limit, offset, filters)
	// Return the results cast to the expected types
	return args.Get(0).([]models.ProductModel), args.Int(1), args.Error(2)
}

func (m *MockRepository) GetCategories() ([]models.Category, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Category), args.Error(1)
}

// CreateCategory implements the CatalogRepository interface by calling the mock function.
func (m *MockRepository) CreateCategory(c models.Category) error {
	args := m.Called(c)
	return args.Error(0)
}

func TestHandleGet(t *testing.T) {
	// A standard set of mock products to be returned on success
	mockProductModels := []models.ProductModel{
		{
			Code:     "Laptop",
			Price:    decimal.NewFromFloat(1200.00), // Updated to use decimal.Decimal
			Category: models.Category{},
		},
		{
			Code:     "Smartphone",
			Price:    decimal.NewFromFloat(800.00), // Updated to use decimal.Decimal
			Category: models.Category{},
		},
	}
	totalItems := 53 // Total items in the database

	tests := []struct {
		name                 string
		url                  string // Full request URL with query parameters
		mockProducts         []models.ProductModel
		mockTotal            int
		mockError            error
		expectedStatus       int
		expectedProductCount int
		expectedTotalPages   int
		expectedTotalItems   int
		expectedCategory     string
		expectedMaxPrice     float64
	}{
		{
			name:                 "Success_NoFilters_DefaultPagination",
			url:                  "/products",
			mockProducts:         mockProductModels,
			mockTotal:            totalItems,
			mockError:            nil,
			expectedStatus:       http.StatusOK,
			expectedProductCount: 2,
			expectedTotalPages:   6, // 53 items / 10 limit = 5.3 -> 6 pages
			expectedTotalItems:   totalItems,
			expectedCategory:     "",
			expectedMaxPrice:     0,
		},
		{
			name:                 "Success_WithCategoryFilter",
			url:                  "/products?category=Elec", // Check case insensitivity
			mockProducts:         mockProductModels,
			mockTotal:            15,
			mockError:            nil,
			expectedStatus:       http.StatusOK,
			expectedProductCount: 2,
			expectedTotalPages:   2, // 15 items / 10 limit = 1.5 -> 2 pages
			expectedTotalItems:   15,
			expectedCategory:     "elec",
			expectedMaxPrice:     0,
		},
		{
			name:                 "Success_WithMaxPriceFilter",
			url:                  "/products?max_price=1000.50",
			mockProducts:         mockProductModels,
			mockTotal:            30,
			mockError:            nil,
			expectedStatus:       http.StatusOK,
			expectedProductCount: 2,
			expectedTotalPages:   3, // 30 items / 10 limit = 3 pages
			expectedTotalItems:   30,
			expectedCategory:     "",
			expectedMaxPrice:     1000.50,
		},
		{
			name:                 "Success_WithPagination",
			url:                  "/products?page=3&limit=5",
			mockProducts:         mockProductModels[:1], // Only return one item
			mockTotal:            12,
			mockError:            nil,
			expectedStatus:       http.StatusOK,
			expectedProductCount: 1,
			expectedTotalPages:   3, // 12 items / 5 limit = 2.4 -> 3 pages
			expectedTotalItems:   12,
			expectedCategory:     "",
			expectedMaxPrice:     0,
		},
		{
			name:                 "Success_InvalidMaxPrice_Ignored",
			url:                  "/products?max_price=not_a_number",
			mockProducts:         mockProductModels,
			mockTotal:            totalItems,
			mockError:            nil,
			expectedStatus:       http.StatusOK,
			expectedProductCount: 2,
			expectedTotalPages:   6,
			expectedTotalItems:   totalItems,
			expectedCategory:     "",
			expectedMaxPrice:     0, // Should be ignored
		},
		{
			name:                 "Failure_RepositoryError",
			url:                  "/products",
			mockProducts:         []models.ProductModel{},
			mockTotal:            0,
			mockError:            errors.New("db connection timeout"),
			expectedStatus:       http.StatusInternalServerError,
			expectedProductCount: 0,
			expectedTotalPages:   0,
			expectedTotalItems:   0,
			expectedCategory:     "",
			expectedMaxPrice:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			handler := &CatalogHandler{repo: mockRepo}

			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			rr := httptest.NewRecorder()

			// Prepare the expected ProductFilters object for assertion
			expectedFilters := models.ProductFilters{
				CategoryCode: tt.expectedCategory,
				MaxPrice:     tt.expectedMaxPrice,
			}

			// Set up the mock expectation for GetAllProducts
			mockRepo.On("GetAllProducts",
				mock.AnythingOfType("int"), // limit
				mock.AnythingOfType("int"), // offset
				expectedFilters,            // filters (asserts filter values are correct)
			).Return(tt.mockProducts, tt.mockTotal, tt.mockError).Maybe()

			// Execute the function under test
			handler.HandleGet(rr, req)

			// Assertion 1: Status Code
			if rr.Code != tt.expectedStatus {
				t.Fatalf("Expected status code %d, got %d. Body: %s", tt.expectedStatus, rr.Code, rr.Body.String())
			}

			// If expecting failure, we stop here as the body will be a simple error message.
			if tt.expectedStatus != http.StatusOK {
				mockRepo.AssertExpectations(t)
				return
			}

			// Assertion 2: Response Body Structure and Content
			var response Response
			if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
				t.Fatalf("Could not unmarshal response body: %v", err)
			}

			// Check product count
			if len(response.Products) != tt.expectedProductCount {
				t.Errorf("Expected %d products, got %d", tt.expectedProductCount, len(response.Products))
			}

			// Check metadata
			if response.Metadata.TotalPages != tt.expectedTotalPages {
				t.Errorf("Expected TotalPages %d, got %d", tt.expectedTotalPages, response.Metadata.TotalPages)
			}
			if response.Metadata.TotalItems != tt.expectedTotalItems {
				t.Errorf("Expected TotalItems %d, got %d", tt.expectedTotalItems, response.Metadata.TotalItems)
			}

			// Verify that all expectations were met
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestHandleProductGet(t *testing.T) {
	// Define test cases
	tests := []struct {
		name           string
		productCode    string
		mockSetup      func(m *MockRepository)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "Success - Product Found",
			productCode: "PROD123",
			mockSetup: func(m *MockRepository) {
				m.On("GetProductByCode", "PROD123").Return(&models.ProductModel{
					Code:     "PROD123",
					Price:    decimal.NewFromFloat(100.50),
					Category: models.Category{Name: "Test Widget"},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"code":"PROD123","final_price":100.5,"name":"Test Widget"}`,
		},
		{
			name:        "Error - Product Not Found",
			productCode: "PROD-999",
			mockSetup: func(m *MockRepository) {
				// Simulating a repository error that contains "not found"
				m.On("GetProductByCode", "PROD-999").Return(nil, errors.New("record not found in db"))
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Product PROD-999 not found.",
		},
		{
			name:        "Error - Internal Server Error",
			productCode: "PROD-ERR",
			mockSetup: func(m *MockRepository) {
				// Simulating a DB connection failure or other error
				m.On("GetProductByCode", "PROD-ERR").Return(nil, errors.New("db connection failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal server error retrieving product details.",
		},
		{
			name:        "Error - Bad Request (Empty Code)",
			productCode: "", // This triggers the extraction error in our stub
			mockSetup: func(m *MockRepository) {
				// Repo should not be called if URL extraction fails
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid product code specified in URL.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 1. Setup Mock
			mockRepo := new(MockRepository)
			tt.mockSetup(mockRepo)

			// 2. Setup Handler
			handler := &CatalogHandler{repo: mockRepo}

			// 3. Create Request and Recorder
			req := httptest.NewRequest(http.MethodGet, "/catalog/"+tt.productCode, nil)

			// Create a ResponseRecorder to record the response
			rr := httptest.NewRecorder()
			mockReq := &http.Request{
				Method: req.Method,
				URL:    req.URL,
			}
			mockReq = http.SetPathValue(mockReq, "code", tt.productCode)
			// ACT: Call the handler with the mocked request
			handler.HandleProductGet(rr, mockReq)

			// 5. Assertions
			assert.Equal(t, tt.expectedStatus, rr.Code)

			// Clean up body string (remove newlines usually added by http.Error)
			body := strings.TrimSpace(rr.Body.String())
			assert.Contains(t, body, strings.TrimSpace(tt.expectedBody))

			// Verify that the expected mock methods were called
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestHandleCategoryList(t *testing.T) {
	categoriesFixture := []models.Category{
		{ID: 1, Code: "ELEC", Name: "Electronics"},
		{ID: 2, Code: "BOOK", Name: "Books"},
	}

	tests := []struct {
		name           string
		mockSetup      func(m *MockRepository)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Success - Returns List of Categories",
			mockSetup: func(m *MockRepository) {
				m.On("GetCategories").Return(categoriesFixture, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"ID":1,"Code":"ELEC","Name":"Electronics","Products":null},{"ID":2,"Code":"BOOK","Name":"Books","Products":null}]`,
		},
		{
			name: "Error - Internal Server Error",
			mockSetup: func(m *MockRepository) {
				m.On("GetCategories").Return(nil, errors.New("database connection failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal server error retrieving categories.",
		},
		{
			name: "Success - Returns Empty List",
			mockSetup: func(m *MockRepository) {
				m.On("GetCategories").Return([]models.Category{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `[]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 1. Setup Mock
			mockRepo := new(MockRepository)
			tt.mockSetup(mockRepo)

			// 2. Setup Handler
			handler := &CatalogHandler{repo: mockRepo}

			// 3. Create Request and Recorder
			req, _ := http.NewRequest("GET", "/categories", nil)
			rr := httptest.NewRecorder()

			// 4. Execute
			handler.HandleCategoryList(rr, req)

			// 5. Assertions
			assert.Equal(t, tt.expectedStatus, rr.Code)

			// Clean up body string (remove newlines usually added by http.Error)
			body := strings.TrimSpace(rr.Body.String())

			// For non-error cases, we check the JSON structure.
			if tt.expectedStatus == http.StatusOK {
				// We need to compare the JSON body by unmarshaling, as GORM tags
				// lead to default JSON serialization of the fields (including null Products array).
				var actual []models.Category
				err := json.Unmarshal([]byte(body), &actual)
				assert.NoError(t, err)

				var expected []models.Category
				err = json.Unmarshal([]byte(tt.expectedBody), &expected)
				assert.NoError(t, err)

				assert.Equal(t, expected, actual)

			} else {
				// For error cases, we check the simple error message.
				assert.Contains(t, body, strings.TrimSpace(tt.expectedBody))
			}

			// Verify that the expected mock methods were called
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestHandleCategoryCreate(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		mockCreateErr  error
		expectedStatus int
		expectedBody   string // Substring to check in the response body
	}{
		{
			name:           "Success_ValidInput",
			requestBody:    `{"code": "ELEC", "name": "Electronics"}`,
			mockCreateErr:  nil,
			expectedStatus: http.StatusOK,
			expectedBody:   `"code":"ELEC"`,
		},
		{
			name:           "Failure_InvalidJSON",
			requestBody:    `{"code": "ELEC", "name": "Electronics"`, // Missing closing brace
			mockCreateErr:  nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid request body",
		},
		{
			name:           "Failure_MissingCode",
			requestBody:    `{"name": "Books"}`,
			mockCreateErr:  nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Code and Name fields are required",
		},
		{
			name:           "Failure_MissingName",
			requestBody:    `{"code": "BOOKS"}`,
			mockCreateErr:  nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Code and Name fields are required",
		},
		{
			name:           "Failure_RepositoryError",
			requestBody:    `{"code": "FASH", "name": "Fashion"}`,
			mockCreateErr:  errors.New("database integrity constraint violation"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal server error while creating category",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup the mock repository to return the predefined error
			mockRepo := new(MockRepository)

			// Set up the expectation ONLY for cases where the handler will attempt to call CreateCategory.
			// Cases like InvalidJSON or Missing Fields will fail validation before hitting the repository.
			if tt.expectedStatus != http.StatusBadRequest {
				// We expect the repository method CreateCategory to be called with any Category type
				// and return the predefined error (or nil for success).
				mockRepo.On("CreateCategory", mock.AnythingOfType("Category")).Return(tt.mockCreateErr).Once()
			}

			// Instantiate the handler with the mock repository
			handler := &CatalogHandler{repo: mockRepo}

			// Create the request
			req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBufferString(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")

			// Create a Response Recorder
			rr := httptest.NewRecorder()

			// Execute the function under test
			handler.HandleCategoryCreate(rr, req)

			// Assertions
			if rr.Code != tt.expectedStatus {
				t.Errorf("Test %s: Expected status code %d, got %d", tt.name, tt.expectedStatus, rr.Code)
			}

			// Check if the response body contains the expected content (error message or successful JSON field)
			bodyStr := rr.Body.String()
			if !bytes.Contains([]byte(bodyStr), []byte(tt.expectedBody)) {
				t.Errorf("Test %s: Expected response body to contain '%s', but got: %s", tt.name, tt.expectedBody, bodyStr)
			}
		})
	}
}
