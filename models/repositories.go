package models

import "github.com/shopspring/decimal"

type ProductRepository interface {
	GetAll(offset, limit int, categoryCode string, priceLessThan *decimal.Decimal) ([]Product, int64, error)
	GetByCode(code string) (*Product, error)
}

type CategoryRepository interface {
	GetAll() ([]Category, error)
	Create(category *Category) error
}

