package repository

import (
	"errors"

	"order-management/services/products/model"

	"gorm.io/gorm"
)

type ProductRepository struct {
	DB *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{DB: db}
}

func (r *ProductRepository) Create(p *model.Product) error {
	return r.DB.Create(p).Error
}

func (r *ProductRepository) GetAll() ([]model.Product, error) {
	var products []model.Product
	err := r.DB.Find(&products).Error
	return products, err
}

func (r *ProductRepository) GetByID(id uint) (*model.Product, error) {
	var product model.Product
	if err := r.DB.First(&product, id).Error; err != nil {
		return nil, err
	}
	return &product, nil
}

// ChangeStock updates stock in one SQL statement. A reservation only succeeds
// when enough units are available, even if multiple orders arrive together.
func (r *ProductRepository) ChangeStock(id uint, quantity int, action string) error {
	if quantity <= 0 {
		return errors.New("quantity must be greater than zero")
	}

	var result *gorm.DB
	switch action {
	case "reserve":
		result = r.DB.Model(&model.Product{}).
			Where("id = ? AND stock >= ?", id, quantity).
			UpdateColumn("stock", gorm.Expr("stock - ?", quantity))
	case "release":
		result = r.DB.Model(&model.Product{}).
			Where("id = ?", id).
			UpdateColumn("stock", gorm.Expr("stock + ?", quantity))
	default:
		return errors.New("action must be reserve or release")
	}
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		if action == "reserve" {
			return errors.New("product not found or insufficient stock")
		}
		return errors.New("product not found")
	}
	return nil
}
