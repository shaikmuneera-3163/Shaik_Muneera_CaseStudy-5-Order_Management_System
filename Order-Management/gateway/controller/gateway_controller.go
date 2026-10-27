package controller

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"order-management/gateway/client"
	"order-management/services/products/model"
	"order-management/shared/dto"
	"order-management/shared/response"

	"github.com/gin-gonic/gin"
)

type GatewayController struct {
	Auth    *client.AuthClient
	Product *client.ProductClient
	Order   *client.OrderClient
}

func NewGatewayController(
	auth *client.AuthClient,
	product *client.ProductClient,
	order *client.OrderClient,
) *GatewayController {
	return &GatewayController{
		Auth:    auth,
		Product: product,
		Order:   order,
	}
}

func (g *GatewayController) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid login request")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	result, err := g.Auth.Login(ctx, req)
	if err != nil {
		response.Error(c, http.StatusBadGateway, "authentication service unavailable or login failed")
		return
	}

	c.JSON(http.StatusOK, result)
}

func (g *GatewayController) GetProducts(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	products, err := g.Product.GetAll(ctx)
	if err != nil {
		response.Error(c, http.StatusBadGateway, "product service unavailable")
		return
	}

	response.Success(c, http.StatusOK, "products retrieved", products)
}

func (g *GatewayController) CreateOrder(c *gin.Context) {
	var req dto.OrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid order request")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	result, err := g.Order.Create(ctx, req)
	if err != nil {
		response.Error(c, http.StatusBadGateway, err.Error())
		return
	}

	c.JSON(http.StatusCreated, result)
}
func (g *GatewayController) CreateProduct(c *gin.Context) {
	var product model.Product

	if err := c.ShouldBindJSON(&product); err != nil {
		response.Error(
			c,
			http.StatusBadRequest,
			"invalid product request",
		)
		return
	}

	ctx, cancel := context.WithTimeout(
		c.Request.Context(),
		5*time.Second,
	)
	defer cancel()

	result, err := g.Product.Create(ctx, product)
	if err != nil {
		response.Error(
			c,
			http.StatusBadGateway,
			"product service unavailable or product creation failed",
		)
		return
	}

	c.JSON(http.StatusCreated, result)
}

func (g *GatewayController) GetOrders(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	orders, err := g.Order.GetAll(ctx)
	if err != nil {
		response.Error(c, http.StatusBadGateway, "order service unavailable")
		return
	}
	response.Success(c, http.StatusOK, "orders retrieved", orders)
}

func (g *GatewayController) GetOrderByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Error(c, http.StatusBadRequest, "invalid order ID")
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	order, err := g.Order.GetByID(ctx, uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, "order not found or order service unavailable")
		return
	}
	response.Success(c, http.StatusOK, "order retrieved", order)
}

func (g *GatewayController) UpdateOrderStatus(c *gin.Context) {
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
	order, err := g.Order.UpdateStatus(ctx, uint(id), req.Status)
	if err != nil {
		response.Error(c, http.StatusBadGateway, "could not update order status: "+err.Error())
		return
	}
	response.Success(c, http.StatusOK, "order status updated", order)
}

func (g *GatewayController) GetProductByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Error(c, http.StatusBadRequest, "invalid product ID")
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	product, err := g.Product.GetByID(ctx, uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, "product not found or product service unavailable")
		return
	}
	response.Success(c, http.StatusOK, "product retrieved", product)
}
