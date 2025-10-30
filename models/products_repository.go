package models

import (
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type ProductsRepository struct {
	db *gorm.DB
}

func NewProductsRepository(db *gorm.DB) *ProductsRepository {
	return &ProductsRepository{
		db: db,
	}
}

func (r *ProductsRepository) GetAll(offset, limit int, categoryCode string, priceLessThan *decimal.Decimal) ([]Product, int64, error) {
	var products []Product
	var total int64

	query := r.db.Model(&Product{})

	if categoryCode != "" {
		query = query.Joins("JOIN categories ON categories.id = products.category_id").
			Where("categories.code = ?", categoryCode)
	}

	if priceLessThan != nil {
		query = query.Where("products.price < ?", priceLessThan)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("Category").Preload("Variants").
		Offset(offset).Limit(limit).
		Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (r *ProductsRepository) GetByCode(code string) (*Product, error) {
	var product Product
	if err := r.db.Preload("Category").Preload("Variants").Where("code = ?", code).First(&product).Error; err != nil {
		return nil, err
	}
	return &product, nil
}
