package categories

import (
	"encoding/json"
	"net/http"

	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/models"
)

type CategoryResponse struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type CreateCategoryRequest struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type CategoriesHandler struct {
	repo models.CategoryRepository
}

func NewCategoriesHandler(r models.CategoryRepository) *CategoriesHandler {
	return &CategoriesHandler{
		repo: r,
	}
}

func (h *CategoriesHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	categories, err := h.repo.GetAll()
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, "failed to fetch categories")
		return
	}

	responses := make([]CategoryResponse, len(categories))
	for i, c := range categories {
		responses[i] = CategoryResponse{
			Code: c.Code,
			Name: c.Name,
		}
	}

	api.OKResponse(w, responses)
}

func (h *CategoriesHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	var req CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Code == "" || req.Name == "" {
		api.ErrorResponse(w, http.StatusBadRequest, "code and name are required")
		return
	}

	category := &models.Category{
		Code: req.Code,
		Name: req.Name,
	}

	if err := h.repo.Create(category); err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, "failed to create category")
		return
	}

	api.OKResponse(w, CategoryResponse{
		Code: category.Code,
		Name: category.Name,
	})
}

