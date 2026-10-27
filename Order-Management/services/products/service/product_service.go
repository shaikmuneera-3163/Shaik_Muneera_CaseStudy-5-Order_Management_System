package service

import (
	"order-management/services/products/model"
	"order-management/services/products/repository"
)

type ProductService struct {
	Repo *repository.ProductRepository
}

func NewProductService(repo *repository.ProductRepository) *ProductService {
	return &ProductService{Repo: repo}
}

func (s *ProductService) Create(p *model.Product) error {
	return s.Repo.Create(p)
}

func (s *ProductService) GetAll() ([]model.Product, error) {
	return s.Repo.GetAll()
}

func (s *ProductService) GetByID(id uint) (*model.Product, error) {
	return s.Repo.GetByID(id)
}

func (s *ProductService) ChangeStock(id uint, quantity int, action string) error {
	return s.Repo.ChangeStock(id, quantity, action)
}
