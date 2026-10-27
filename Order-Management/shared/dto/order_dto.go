package dto

// OrderRequest contains the information needed to create an order.
type OrderRequest struct {
	CustomerID uint `json:"customerId" binding:"required,gt=0"`
	ProductID  uint `json:"productId" binding:"required,gt=0"`
	Quantity   int  `json:"quantity" binding:"required,gt=0"`
}
