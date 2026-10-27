package service

import (
	"context"
	"errors"
	"strings"

	"order-management/services/orders/client"
	"order-management/services/orders/model"
	"order-management/services/orders/repository"
	"order-management/shared/dto"
)

type OrderService struct {
	Repo          *repository.OrderRepository
	ProductClient *client.ProductClient
}

func NewOrderService(repo *repository.OrderRepository, productClient *client.ProductClient) *OrderService {
	return &OrderService{Repo: repo, ProductClient: productClient}
}

func (s *OrderService) Create(ctx context.Context, req dto.OrderRequest) (*model.Order, error) {
	if req.CustomerID == 0 || req.ProductID == 0 || req.Quantity <= 0 {
		return nil, errors.New("customerId, productId and quantity must be positive")
	}

	// Fetch product details (price/name) with timeout, bounded retry and circuit breaker.
	product, err := s.ProductClient.GetByID(ctx, req.ProductID)
	if err != nil {
		return nil, errors.New("product service unavailable or product not found")
	}
	if product.Stock < req.Quantity {
		return nil, errors.New("insufficient product stock")
	}

	// Reserve stock in the Product service before saving the order.
	// The Product service performs an atomic stock >= quantity check.
	if err := s.ProductClient.ChangeStock(ctx, product.ID, req.Quantity, "reserve"); err != nil {
		return nil, err
	}

	order := model.Order{
		CustomerID:  req.CustomerID,
		ProductID:   product.ID,
		ProductName: product.Name,
		Quantity:    req.Quantity,
		UnitPrice:   product.Price,
		TotalAmount: product.Price * float64(req.Quantity),
		Status:      "CREATED",
	}
	if err := s.Repo.Create(&order); err != nil {
		// Compensation: if saving the order fails, return the reserved stock.
		_ = s.ProductClient.ChangeStock(ctx, product.ID, req.Quantity, "release")
		return nil, errors.New("could not save order; stock reservation was rolled back where possible")
	}
	return &order, nil
}

func (s *OrderService) GetAll() ([]model.Order, error) {
	return s.Repo.GetAll()
}

func (s *OrderService) GetByID(id uint) (*model.Order, error) {
	return s.Repo.GetByID(id)
}

func (s *OrderService) UpdateStatus(ctx context.Context, id uint, requestedStatus string) (*model.Order, error) {
	status := strings.ToUpper(strings.TrimSpace(requestedStatus))
	switch status {
	case "CREATED", "CONFIRMED", "CANCELLED":
	default:
		return nil, errors.New("invalid status; use CREATED, CONFIRMED, or CANCELLED")
	}

	order, err := s.Repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if strings.EqualFold(order.Status, status) {
		return order, nil
	}
	if strings.EqualFold(order.Status, "CANCELLED") {
		return nil, errors.New("a cancelled order cannot be changed")
	}

	// Mark the order cancelled only if its previous status has not changed.
	if status == "CANCELLED" {
		if err := s.Repo.UpdateStatusIfCurrent(id, order.Status, status); err != nil {
			return nil, err
		}
		// Compensation is needed because product and order data are in separate databases.
		if err := s.ProductClient.ChangeStock(ctx, order.ProductID, order.Quantity, "release"); err != nil {
			_ = s.Repo.UpdateStatusIfCurrent(id, "CANCELLED", order.Status)
			return nil, errors.New("could not restore product stock; order cancellation was rolled back where possible")
		}
		return s.Repo.GetByID(id)
	}

	if err := s.Repo.UpdateStatusIfCurrent(id, order.Status, status); err != nil {
		return nil, err
	}
	return s.Repo.GetByID(id)
}
