package repository

import (
	"order-management/services/auth/model"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *model.User) error
	FindByUsername(username string) (*model.User, error)
}

type Repository struct {
	DB *gorm.DB
}

func NewUserRepository(db *gorm.DB) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) Create(user *model.User) error {
	return r.DB.Create(user).Error
}

func (r *Repository) FindByUsername(username string) (*model.User, error) {
	var user model.User

	err := r.DB.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}
