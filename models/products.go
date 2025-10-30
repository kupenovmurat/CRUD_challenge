package models

import (
	"github.com/shopspring/decimal"
)

type Product struct {
	ID         uint            `gorm:"primaryKey"`
	Code       string          `gorm:"uniqueIndex;not null"`
	Price      decimal.Decimal `gorm:"type:decimal(10,2);not null"`
	CategoryID *uint           `gorm:"null"`
	Category   *Category       `gorm:"foreignKey:CategoryID"`
	Variants   []Variant       `gorm:"foreignKey:ProductID"`
}

func (p *Product) TableName() string {
	return "products"
}
