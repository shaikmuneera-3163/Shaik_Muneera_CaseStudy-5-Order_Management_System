package controller

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"order-management/services/orders/service"
	"order-management/services/orders/utility"
	"order-management/shared/dto"
	"order-management/shared/response"

	"github.com/gin-gonic/gin"
)

type OrderController struct {
	Service *service.OrderService
}

func NewOrderController(s *service.OrderService) *OrderController {
	return &OrderController{Service: s}
}

func (ctrl *OrderController) Create(c *gin.Context) {
	var req dto.OrderRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid order request")
		return
	}

	if req.CustomerID == 0 || !utility.ValidOrder(req.ProductID, req.Quantity) {
		response.Error(c, http.StatusBadRequest, "invalid product ID or quantity")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	order, err := ctrl.Service.Create(ctx, req)
	if err != nil {
		response.Error(c, http.StatusBadGateway, err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "order created", order)
}

func (ctrl *OrderController) GetAll(c *gin.Context) {
	orders, err := ctrl.Service.GetAll()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "could not retrieve orders")
		return
	}

	response.Success(c, http.StatusOK, "orders retrieved", orders)
}

func (ctrl *OrderController) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Error(c, http.StatusBadRequest, "invalid order ID")
		return
	}

	order, err := ctrl.Service.GetByID(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, "order not found")
		return
	}

	response.Success(c, http.StatusOK, "order retrieved", order)
}

// UpdateStatus updates the status of an existing order.
func (ctrl *OrderController) UpdateStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Error(c, http.StatusBadRequest, "invalid order ID")
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "status is required")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	order, err := ctrl.Service.UpdateStatus(ctx, uint(id), req.Status)
	if err != nil {
		if err.Error() == "record not found" {
			response.Error(c, http.StatusNotFound, "order not found")
			return
		}
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, http.StatusOK, "order status updated", order)
}
