package model

type Product struct {
	ID    uint    `json:"id" gorm:"primaryKey"`
	Name  string  `json:"name" gorm:"not null"`
	Price float64 `json:"price" gorm:"not null"`
	Stock int     `json:"stock" gorm:"not null"`
}
