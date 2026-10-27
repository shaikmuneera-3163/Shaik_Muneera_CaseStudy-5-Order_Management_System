package controller

import (
	"net/http"
	"strconv"

	"order-management/services/products/model"
	"order-management/services/products/service"
	"order-management/services/products/utility"
	"order-management/shared/dto"
	"order-management/shared/response"

	"github.com/gin-gonic/gin"
)

type ProductController struct {
	Service *service.ProductService
}

func NewProductController(s *service.ProductService) *ProductController {
	return &ProductController{Service: s}
}

func (ctrl *ProductController) Create(c *gin.Context) {
	var req dto.ProductRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid product request")
		return
	}

	if !utility.ValidProduct(req.Name, req.Price, req.Stock) {
		response.Error(c, http.StatusBadRequest, "invalid name, price or stock")
		return
	}

	product := model.Product{
		Name:  req.Name,
		Price: req.Price,
		Stock: req.Stock,
	}

	if err := ctrl.Service.Create(&product); err != nil {
		response.Error(c, http.StatusInternalServerError, "could not create product")
		return
	}

	response.Success(c, http.StatusCreated, "product created", product)
}

func (ctrl *ProductController) GetAll(c *gin.Context) {
	products, err := ctrl.Service.GetAll()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "could not retrieve products")
		return
	}

	response.Success(c, http.StatusOK, "products retrieved", products)
}

func (ctrl *ProductController) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Error(c, http.StatusBadRequest, "invalid product ID")
		return
	}

	product, err := ctrl.Service.GetByID(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, "product not found")
		return
	}

	response.Success(c, http.StatusOK, "product retrieved", product)
}

// ChangeStock is used internally by the Order service. Clients should create
// orders through the Gateway instead of calling this operation directly.
func (ctrl *ProductController) ChangeStock(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Error(c, http.StatusBadRequest, "invalid product ID")
		return
	}

	var req struct {
		Quantity int    `json:"quantity" binding:"required,gt=0"`
		Action   string `json:"action" binding:"required,oneof=reserve release"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "quantity must be positive and action must be reserve or release")
		return
	}

	if err := ctrl.Service.ChangeStock(uint(id), req.Quantity, req.Action); err != nil {
		if err.Error() == "product not found or insufficient stock" || err.Error() == "product not found" {
			response.Error(c, http.StatusConflict, err.Error())
			return
		}
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, http.StatusOK, "product stock updated", gin.H{"productId": uint(id), "quantity": req.Quantity, "action": req.Action})
}
