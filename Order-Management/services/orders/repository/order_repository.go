package repository

import (
	"errors"

	"gorm.io/gorm"
	"order-management/services/orders/model"
)

type OrderRepository struct{ DB *gorm.DB }

func NewOrderRepository(db *gorm.DB) *OrderRepository { return &OrderRepository{DB: db} }

func (r *OrderRepository) Create(order *model.Order) error {
	return r.DB.Create(order).Error
}

func (r *OrderRepository) GetAll() ([]model.Order, error) {
	var orders []model.Order
	err := r.DB.Order("id ASC").Find(&orders).Error
	return orders, err
}

func (r *OrderRepository) GetByID(id uint) (*model.Order, error) {
	var order model.Order
	if err := r.DB.First(&order, id).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepository) UpdateStatus(id uint, status string) error {
	result := r.DB.Model(&model.Order{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// UpdateStatusIfCurrent prevents two requests from cancelling the same order
// and returning its stock twice.
func (r *OrderRepository) UpdateStatusIfCurrent(id uint, current, next string) error {
	result := r.DB.Model(&model.Order{}).
		Where("id = ? AND status = ?", id, current).
		Update("status", next)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("order status changed; refresh the order and try again")
	}
	return nil
}
