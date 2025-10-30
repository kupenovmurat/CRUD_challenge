package catalog

import (
	"net/http"
	"strconv"

	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/shopspring/decimal"
)

type CatalogResponse struct {
	Products []ProductResponse `json:"products"`
	Total    int64             `json:"total"`
}

type ProductResponse struct {
	Code     string           `json:"code"`
	Price    float64          `json:"price"`
	Category *CategorySummary `json:"category,omitempty"`
}

type ProductDetailsResponse struct {
	Code     string            `json:"code"`
	Price    float64           `json:"price"`
	Category *CategorySummary  `json:"category,omitempty"`
	Variants []VariantResponse `json:"variants"`
}

type VariantResponse struct {
	Name  string  `json:"name"`
	SKU   string  `json:"sku"`
	Price float64 `json:"price"`
}

type CategorySummary struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type CatalogHandler struct {
	repo models.ProductRepository
}

func NewCatalogHandler(r models.ProductRepository) *CatalogHandler {
	return &CatalogHandler{
		repo: r,
	}
}

func (h *CatalogHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	offset := 0
	limit := 10

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if val, err := strconv.Atoi(offsetStr); err == nil && val >= 0 {
			offset = val
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil {
			if val < 1 {
				limit = 1
			} else if val > 100 {
				limit = 100
			} else {
				limit = val
			}
		}
	}

	categoryCode := r.URL.Query().Get("category")

	var priceLessThan *decimal.Decimal
	if priceStr := r.URL.Query().Get("price_less_than"); priceStr != "" {
		if price, err := decimal.NewFromString(priceStr); err == nil {
			priceLessThan = &price
		}
	}

	products, total, err := h.repo.GetAll(offset, limit, categoryCode, priceLessThan)
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, "failed to fetch products")
		return
	}

	productResponses := make([]ProductResponse, len(products))
	for i, p := range products {
		productResponses[i] = ProductResponse{
			Code:  p.Code,
			Price: p.Price.InexactFloat64(),
		}
		if p.Category != nil {
			productResponses[i].Category = &CategorySummary{
				Code: p.Category.Code,
				Name: p.Category.Name,
			}
		}
	}

	api.OKResponse(w, CatalogResponse{
		Products: productResponses,
		Total:    total,
	})
}

func (h *CatalogHandler) HandleGetByCode(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if code == "" {
		api.ErrorResponse(w, http.StatusBadRequest, "product code is required")
		return
	}

	product, err := h.repo.GetByCode(code)
	if err != nil {
		api.ErrorResponse(w, http.StatusNotFound, "product not found")
		return
	}

	variants := make([]VariantResponse, len(product.Variants))
	for i, v := range product.Variants {
		price := v.Price
		if price.IsZero() {
			price = product.Price
		}
		variants[i] = VariantResponse{
			Name:  v.Name,
			SKU:   v.SKU,
			Price: price.InexactFloat64(),
		}
	}

	response := ProductDetailsResponse{
		Code:     product.Code,
		Price:    product.Price.InexactFloat64(),
		Variants: variants,
	}

	if product.Category != nil {
		response.Category = &CategorySummary{
			Code: product.Category.Code,
			Name: product.Category.Name,
		}
	}

	api.OKResponse(w, response)
}
