package categories

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockCategoryRepository struct {
	mock.Mock
}

func (m *MockCategoryRepository) GetAll() ([]models.Category, error) {
	args := m.Called()
	return args.Get(0).([]models.Category), args.Error(1)
}

func (m *MockCategoryRepository) Create(category *models.Category) error {
	args := m.Called(category)
	return args.Error(0)
}

func TestCategoriesHandler_HandleGet(t *testing.T) {
	t.Run("returns all categories", func(t *testing.T) {
		mockRepo := new(MockCategoryRepository)
		handler := NewCategoriesHandler(mockRepo)

		categories := []models.Category{
			{ID: 1, Code: "clothing", Name: "Clothing"},
			{ID: 2, Code: "shoes", Name: "Shoes"},
			{ID: 3, Code: "accessories", Name: "Accessories"},
		}

		mockRepo.On("GetAll").Return(categories, nil)

		req := httptest.NewRequest("GET", "/categories", nil)
		recorder := httptest.NewRecorder()

		handler.HandleGet(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

		var response []CategoryResponse
		err := json.NewDecoder(recorder.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Len(t, response, 3)
		assert.Equal(t, "clothing", response[0].Code)
		assert.Equal(t, "Clothing", response[0].Name)

		mockRepo.AssertExpectations(t)
	})

	t.Run("handles repository errors", func(t *testing.T) {
		mockRepo := new(MockCategoryRepository)
		handler := NewCategoriesHandler(mockRepo)

		mockRepo.On("GetAll").Return([]models.Category{}, errors.New("database error"))

		req := httptest.NewRequest("GET", "/categories", nil)
		recorder := httptest.NewRecorder()

		handler.HandleGet(recorder, req)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		mockRepo.AssertExpectations(t)
	})
}

func TestCategoriesHandler_HandleCreate(t *testing.T) {
	t.Run("creates a new category", func(t *testing.T) {
		mockRepo := new(MockCategoryRepository)
		handler := NewCategoriesHandler(mockRepo)

		reqBody := CreateCategoryRequest{
			Code: "electronics",
			Name: "Electronics",
		}
		body, _ := json.Marshal(reqBody)

		mockRepo.On("Create", mock.AnythingOfType("*models.Category")).Return(nil)

		req := httptest.NewRequest("POST", "/categories", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()

		handler.HandleCreate(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

		var response CategoryResponse
		err := json.NewDecoder(recorder.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, "electronics", response.Code)
		assert.Equal(t, "Electronics", response.Name)

		mockRepo.AssertExpectations(t)
	})

	t.Run("returns 400 for invalid JSON", func(t *testing.T) {
		mockRepo := new(MockCategoryRepository)
		handler := NewCategoriesHandler(mockRepo)

		req := httptest.NewRequest("POST", "/categories", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()

		handler.HandleCreate(recorder, req)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("returns 400 when code is missing", func(t *testing.T) {
		mockRepo := new(MockCategoryRepository)
		handler := NewCategoriesHandler(mockRepo)

		reqBody := CreateCategoryRequest{
			Name: "Electronics",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/categories", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()

		handler.HandleCreate(recorder, req)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("returns 400 when name is missing", func(t *testing.T) {
		mockRepo := new(MockCategoryRepository)
		handler := NewCategoriesHandler(mockRepo)

		reqBody := CreateCategoryRequest{
			Code: "electronics",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/categories", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()

		handler.HandleCreate(recorder, req)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("handles repository errors", func(t *testing.T) {
		mockRepo := new(MockCategoryRepository)
		handler := NewCategoriesHandler(mockRepo)

		reqBody := CreateCategoryRequest{
			Code: "electronics",
			Name: "Electronics",
		}
		body, _ := json.Marshal(reqBody)

		mockRepo.On("Create", mock.AnythingOfType("*models.Category")).Return(errors.New("database error"))

		req := httptest.NewRequest("POST", "/categories", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()

		handler.HandleCreate(recorder, req)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		mockRepo.AssertExpectations(t)
	})
}

