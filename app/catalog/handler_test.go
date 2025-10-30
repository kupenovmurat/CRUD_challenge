package catalog

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) GetAll(offset, limit int, categoryCode string, priceLessThan *decimal.Decimal) ([]models.Product, int64, error) {
	args := m.Called(offset, limit, categoryCode, priceLessThan)
	return args.Get(0).([]models.Product), args.Get(1).(int64), args.Error(2)
}

func (m *MockProductRepository) GetByCode(code string) (*models.Product, error) {
	args := m.Called(code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func TestCatalogHandler_HandleGet(t *testing.T) {
	t.Run("returns products with default pagination", func(t *testing.T) {
		mockRepo := new(MockProductRepository)
		handler := NewCatalogHandler(mockRepo)

		category := &models.Category{ID: 1, Code: "clothing", Name: "Clothing"}
		products := []models.Product{
			{
				ID:         1,
				Code:       "PROD001",
				Price:      decimal.NewFromFloat(10.99),
				CategoryID: &category.ID,
				Category:   category,
			},
		}

		mockRepo.On("GetAll", 0, 10, "", (*decimal.Decimal)(nil)).Return(products, int64(1), nil)

		req := httptest.NewRequest("GET", "/catalog", nil)
		recorder := httptest.NewRecorder()

		handler.HandleGet(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

		var response CatalogResponse
		err := json.NewDecoder(recorder.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), response.Total)
		assert.Len(t, response.Products, 1)
		assert.Equal(t, "PROD001", response.Products[0].Code)
		assert.Equal(t, 10.99, response.Products[0].Price)
		assert.NotNil(t, response.Products[0].Category)
		assert.Equal(t, "clothing", response.Products[0].Category.Code)
		assert.Equal(t, "Clothing", response.Products[0].Category.Name)

		mockRepo.AssertExpectations(t)
	})

	t.Run("returns products with custom pagination", func(t *testing.T) {
		mockRepo := new(MockProductRepository)
		handler := NewCatalogHandler(mockRepo)

		products := []models.Product{}
		mockRepo.On("GetAll", 5, 20, "", (*decimal.Decimal)(nil)).Return(products, int64(100), nil)

		req := httptest.NewRequest("GET", "/catalog?offset=5&limit=20", nil)
		recorder := httptest.NewRecorder()

		handler.HandleGet(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("filters by category", func(t *testing.T) {
		mockRepo := new(MockProductRepository)
		handler := NewCatalogHandler(mockRepo)

		products := []models.Product{}
		mockRepo.On("GetAll", 0, 10, "shoes", (*decimal.Decimal)(nil)).Return(products, int64(0), nil)

		req := httptest.NewRequest("GET", "/catalog?category=shoes", nil)
		recorder := httptest.NewRecorder()

		handler.HandleGet(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("filters by price less than", func(t *testing.T) {
		mockRepo := new(MockProductRepository)
		handler := NewCatalogHandler(mockRepo)

		products := []models.Product{}
		mockRepo.On("GetAll", 0, 10, "", mock.MatchedBy(func(price *decimal.Decimal) bool {
			return price != nil && price.Equal(decimal.NewFromFloat(15.00))
		})).Return(products, int64(0), nil)

		req := httptest.NewRequest("GET", "/catalog?price_less_than=15.00", nil)
		recorder := httptest.NewRecorder()

		handler.HandleGet(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("enforces limit constraints", func(t *testing.T) {
		mockRepo := new(MockProductRepository)
		handler := NewCatalogHandler(mockRepo)

		products := []models.Product{}
		mockRepo.On("GetAll", 0, 100, "", (*decimal.Decimal)(nil)).Return(products, int64(0), nil)

		req := httptest.NewRequest("GET", "/catalog?limit=200", nil)
		recorder := httptest.NewRecorder()

		handler.HandleGet(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("handles repository errors", func(t *testing.T) {
		mockRepo := new(MockProductRepository)
		handler := NewCatalogHandler(mockRepo)

		mockRepo.On("GetAll", 0, 10, "", (*decimal.Decimal)(nil)).
			Return([]models.Product{}, int64(0), errors.New("database error"))

		req := httptest.NewRequest("GET", "/catalog", nil)
		recorder := httptest.NewRecorder()

		handler.HandleGet(recorder, req)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		mockRepo.AssertExpectations(t)
	})
}

func TestCatalogHandler_HandleGetByCode(t *testing.T) {
	t.Run("returns product with variants", func(t *testing.T) {
		mockRepo := new(MockProductRepository)
		handler := NewCatalogHandler(mockRepo)

		category := &models.Category{ID: 1, Code: "clothing", Name: "Clothing"}
		productID := uint(1)
		product := &models.Product{
			ID:         productID,
			Code:       "PROD001",
			Price:      decimal.NewFromFloat(10.99),
			CategoryID: &category.ID,
			Category:   category,
			Variants: []models.Variant{
				{ID: 1, ProductID: productID, Name: "Variant A", SKU: "SKU001A", Price: decimal.NewFromFloat(11.99)},
				{ID: 2, ProductID: productID, Name: "Variant B", SKU: "SKU001B", Price: decimal.Zero},
			},
		}

		mockRepo.On("GetByCode", "PROD001").Return(product, nil)

		req := httptest.NewRequest("GET", "/catalog/PROD001", nil)
		req.SetPathValue("code", "PROD001")
		recorder := httptest.NewRecorder()

		handler.HandleGetByCode(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

		var response ProductDetailsResponse
		err := json.NewDecoder(recorder.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, "PROD001", response.Code)
		assert.Equal(t, 10.99, response.Price)
		assert.NotNil(t, response.Category)
		assert.Equal(t, "clothing", response.Category.Code)
		assert.Len(t, response.Variants, 2)
		assert.Equal(t, "Variant A", response.Variants[0].Name)
		assert.Equal(t, 11.99, response.Variants[0].Price)
		assert.Equal(t, "Variant B", response.Variants[1].Name)
		assert.Equal(t, 10.99, response.Variants[1].Price)

		mockRepo.AssertExpectations(t)
	})

	t.Run("returns 404 when product not found", func(t *testing.T) {
		mockRepo := new(MockProductRepository)
		handler := NewCatalogHandler(mockRepo)

		mockRepo.On("GetByCode", "INVALID").Return((*models.Product)(nil), errors.New("not found"))

		req := httptest.NewRequest("GET", "/catalog/INVALID", nil)
		req.SetPathValue("code", "INVALID")
		recorder := httptest.NewRecorder()

		handler.HandleGetByCode(recorder, req)

		assert.Equal(t, http.StatusNotFound, recorder.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("returns 400 when code is empty", func(t *testing.T) {
		mockRepo := new(MockProductRepository)
		handler := NewCatalogHandler(mockRepo)

		req := httptest.NewRequest("GET", "/catalog/", nil)
		recorder := httptest.NewRecorder()

		handler.HandleGetByCode(recorder, req)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})
}

