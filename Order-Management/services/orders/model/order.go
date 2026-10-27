package model

import "time"

type Order struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	CustomerID  uint      `json:"customer_id" gorm:"not null;index"`
	ProductID   uint      `json:"product_id" gorm:"not null;index"`
	ProductName string    `json:"product_name" gorm:"not null"`
	Quantity    int       `json:"quantity" gorm:"not null"`
	UnitPrice   float64   `json:"unit_price" gorm:"not null"`
	TotalAmount float64   `json:"total_amount" gorm:"not null"`
	Status      string    `json:"status" gorm:"not null;index"`
	CreatedAt   time.Time `json:"created_at"`
}
