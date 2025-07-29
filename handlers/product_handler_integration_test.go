package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/glebarez/sqlite"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"products/database"
	"products/models"
	"testing"
)

func setupTestDB(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	err = db.AutoMigrate(&models.Product{})
	if err != nil {
		t.Fatalf("failed to auto migrate products: %v", err)
	}

	database.DB = db
}

func setupTestApp() *fiber.App {
	app := fiber.New()
	api := app.Group("/api")
	productGroup := api.Group("/products")
	productGroup.Get("/", GetProducts)
	productGroup.Get("/:id", GetProductByID)
	productGroup.Post("/", CreateProduct)
	productGroup.Patch("/:id", PatchProduct)
	productGroup.Delete("/:id", DeleteProduct)

	return app
}

func TestGetProducts(t *testing.T) {
	setupTestDB(t)
	app := setupTestApp()
	database.DB.Exec("DELETE FROM products")

	testCases := []struct {
		name           string
		setup          func(t *testing.T)
		expectedStatus int
		verifyBody     func(t *testing.T, body []byte)
	}{
		{
			name: "Success - Find Products",
			setup: func(t *testing.T) {
				setupTestDB(t)
				database.DB.Exec("DELETE FROM products")
				database.DB.Create(&models.Product{ID: uuid.New(), Name: "Produto 1", Price: 100})
			},
			expectedStatus: fiber.StatusOK,
			verifyBody: func(t *testing.T, body []byte) {
				var returnedProducts []models.Product
				err := json.Unmarshal(body, &returnedProducts)
				assert.NoError(t, err)
				assert.Len(t, returnedProducts, 1)
			},
		},
		{
			name: "Success - No Products",
			setup: func(t *testing.T) {
				setupTestDB(t)
				database.DB.Exec("DELETE FROM products")
			},
			expectedStatus: fiber.StatusOK,
			verifyBody: func(t *testing.T, body []byte) {
				var returnedProducts []models.Product
				err := json.Unmarshal(body, &returnedProducts)
				assert.NoError(t, err)
				assert.Len(t, returnedProducts, 0)
			},
		},
		{
			name: "Failure - Database error",
			setup: func(t *testing.T) {
				setupTestDB(t)
				sqlDB, err := database.DB.DB()
				assert.NoError(t, err)
				err = sqlDB.Close()
				assert.NoError(t, err)
			},
			expectedStatus: fiber.StatusInternalServerError,
			verifyBody: func(t *testing.T, body []byte) {
				assert.JSONEq(t, `{"error":"Could not fetch products"}`, string(body))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(t)
			req := httptest.NewRequest("GET", "/api/products", nil)
			resp, err := app.Test(req)
			assert.NoError(t, err)
			defer func() {
				if err := resp.Body.Close(); err != nil {
					log.Printf("Falha ao fechar o corpo da resposta: %v", err)
				}
			}()

			assert.Equal(t, tc.expectedStatus, resp.StatusCode)

			body, _ := io.ReadAll(resp.Body)
			tc.verifyBody(t, body)
		})
	}
}

func TestGetProductByID(t *testing.T) {
	setupTestDB(t)
	app := setupTestApp()

	database.DB.Exec("DELETE FROM products")

	mockProduct := models.Product{
		ID:    uuid.New(),
		Name:  "Test Product",
		Price: 1999,
		Stock: 10,
	}
	database.DB.Create(&mockProduct)

	testCases := []struct {
		name           string
		productID      string
		expectedStatus int
		expectedName   string
	}{
		{
			name:           "Success - Find Product by ID",
			productID:      mockProduct.ID.String(),
			expectedStatus: fiber.StatusOK,
			expectedName:   mockProduct.Name,
		},
		{
			name:           "Failure - Product not found",
			productID:      uuid.New().String(),
			expectedStatus: fiber.StatusNotFound,
			expectedName:   "",
		},
		{
			name:           "Failure - Invalid ID",
			productID:      "invalid-id",
			expectedStatus: fiber.StatusBadRequest,
			expectedName:   "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			urlString := fmt.Sprintf("/api/products/%s", tc.productID)
			req := httptest.NewRequest("GET", urlString, nil)

			resp, _ := app.Test(req)
			defer func() {
				if err := resp.Body.Close(); err != nil {
					log.Printf("failed to close response body: %v", err)
				}
			}()

			assert.Equal(t, tc.expectedStatus, resp.StatusCode)

			if tc.expectedStatus == fiber.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				var returnedProduct models.Product
				err := json.Unmarshal(body, &returnedProduct)
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedName, returnedProduct.Name)
			}
		})
	}
}

func TestCreateProduct(t *testing.T) {
	testCases := []struct {
		name           string
		payload        string
		setup          func(t *testing.T)
		expectedStatus int
		verify         func(t *testing.T, resp *http.Response)
	}{
		{
			name:           "Success - Create Product",
			payload:        `{"name":"New Product", "price": 1234, "stock": 50}`,
			setup:          func(t *testing.T) { setupTestDB(t); database.DB.Exec("DELETE FROM products") },
			expectedStatus: fiber.StatusCreated,
			verify: func(t *testing.T, resp *http.Response) {
				var createdProduct models.Product
				body, _ := io.ReadAll(resp.Body)
				err := json.Unmarshal(body, &createdProduct)
				assert.NoError(t, err)
				assert.Equal(t, "New Product", createdProduct.Name)

			},
		},
		{
			name:    "Failure - Invalid Payload",
			payload: `{"name":"New Product", "price": "text"}`,
			setup: func(t *testing.T) {
				setupTestDB(t)
				database.DB.Exec("DELETE FROM products")
			},
			expectedStatus: fiber.StatusBadRequest,
			verify:         func(t *testing.T, resp *http.Response) {},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			app := setupTestApp()
			tc.setup(t)
			req := httptest.NewRequest("POST", "/api/products", bytes.NewBufferString(tc.payload))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			assert.NoError(t, err)
			defer func() {
				if err := resp.Body.Close(); err != nil {
					log.Printf("failed to close response body: %v", err)
				}
			}()

			assert.Equal(t, tc.expectedStatus, resp.StatusCode)
			tc.verify(t, resp)
		})
	}
}

func TestPatchProduct(t *testing.T) {
	testCases := []struct {
		name           string
		productID      func() string
		payload        string
		setup          func(t *testing.T) *models.Product
		expectedStatus int
		verify         func(t *testing.T, resp *http.Response, originalProduct *models.Product)
	}{
		{
			name: "Success - Patch Product",
			productID: func() string {
				var p models.Product
				database.DB.First(&p)
				return p.ID.String()
			},
			payload: `{"name":"Produto Atualizado"}`,
			setup: func(t *testing.T) *models.Product {
				setupTestDB(t)
				database.DB.Exec("DELETE FROM products")
				mockProduct := &models.Product{ID: uuid.New(), Name: "Produto Original", Price: 100}
				database.DB.Create(mockProduct)
				return mockProduct
			},
			expectedStatus: fiber.StatusOK,
			verify: func(t *testing.T, resp *http.Response, originalProduct *models.Product) {
				var updatedProduct models.Product
				body, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)
				err = json.Unmarshal(body, &updatedProduct)
				assert.NoError(t, err)
				assert.Equal(t, "Produto Atualizado", updatedProduct.Name)
				assert.Equal(t, originalProduct.Price, updatedProduct.Price)
			},
		},
		{
			name:      "Failure - Product Not Found",
			productID: func() string { return uuid.New().String() },
			payload:   `{"name":"Produto Fantasma"}`,
			setup: func(t *testing.T) *models.Product {
				setupTestDB(t)
				database.DB.Exec("DELETE FROM products")
				return nil
			},
			expectedStatus: fiber.StatusNotFound,
			verify:         func(t *testing.T, resp *http.Response, originalProduct *models.Product) {},
		},
		{
			name: "Failure - Invalid ID",
			productID: func() string {
				return "invalid-id"
			},
			payload:        `{"name":"invalid"}`,
			setup:          func(t *testing.T) *models.Product { setupTestDB(t); return nil },
			expectedStatus: fiber.StatusBadRequest,
			verify: func(t *testing.T, resp *http.Response, originalProduct *models.Product) {
				var errorResponse map[string]string
				body, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)
				err = json.Unmarshal(body, &errorResponse)
				assert.NoError(t, err)
				assert.Equal(t, "Invalid UUID format", errorResponse["error"])

			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			app := setupTestApp()
			originalProduct := tc.setup(t)
			productID := tc.productID()

			urlString := fmt.Sprintf("/api/products/%s", productID)
			req := httptest.NewRequest(http.MethodPatch, urlString, bytes.NewBufferString(tc.payload))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			assert.NoError(t, err)
			defer func() {
				if err := resp.Body.Close(); err != nil {
					log.Printf("failed to close response body: %v", err)
				}
			}()

			assert.Equal(t, tc.expectedStatus, resp.StatusCode)
			tc.verify(t, resp, originalProduct)
		})
	}
}

func TestDeleteProduct(t *testing.T) {
	testCases := []struct {
		name           string
		productID      func() string
		setup          func(t *testing.T) *models.Product
		expectedStatus int
		verifyDB       func(t *testing.T, originalProduct *models.Product)
	}{
		{
			name: "Sucesso - Deleção de Produto",
			productID: func() string {
				var p models.Product
				database.DB.First(&p)
				return p.ID.String()
			},
			setup: func(t *testing.T) *models.Product {
				setupTestDB(t)
				database.DB.Exec("DELETE FROM products")
				mockProduct := &models.Product{ID: uuid.New(), Name: "Produto Original", Price: 100}
				database.DB.Create(mockProduct)
				return mockProduct
			},
			expectedStatus: fiber.StatusNoContent,
			verifyDB: func(t *testing.T, originalProduct *models.Product) {
				var product models.Product
				err := database.DB.First(&product, originalProduct.ID).Error
				assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
			},
		},
		{
			name:      "Falha - Produto Não Encontrado",
			productID: func() string { return uuid.New().String() },
			setup: func(t *testing.T) *models.Product {
				setupTestDB(t)
				database.DB.Exec("DELETE FROM products")
				return nil
			},
			expectedStatus: fiber.StatusNotFound,
			verifyDB:       func(t *testing.T, originalProduct *models.Product) {},
		},
		{
			name: "Failure - Invalid ID",
			productID: func() string {
				return "invalid-id"
			},
			setup:          func(t *testing.T) *models.Product { setupTestDB(t); return nil },
			expectedStatus: fiber.StatusBadRequest,
			verifyDB:       func(t *testing.T, originalProduct *models.Product) {},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			app := setupTestApp()
			originalProduct := tc.setup(t)
			productID := tc.productID()

			urlString := fmt.Sprintf("/api/products/%s", productID)
			req := httptest.NewRequest(http.MethodDelete, urlString, nil)

			resp, _ := app.Test(req)
			defer func() {
				if err := resp.Body.Close(); err != nil {
					log.Printf("failed to close response body: %v", err)
				}
			}()

			assert.Equal(t, tc.expectedStatus, resp.StatusCode)
			tc.verifyDB(t, originalProduct)
		})
	}
}
