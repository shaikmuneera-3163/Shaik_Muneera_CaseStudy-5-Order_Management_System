package controller

import (
	"net/http"

	"order-management/services/auth/service"
	"order-management/services/auth/utility"
	"order-management/shared/dto"
	"order-management/shared/response"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	Service *service.AuthService
}

func NewAuthController(s *service.AuthService) *AuthController {
	return &AuthController{Service: s}
}

func (ctrl *AuthController) Register(c *gin.Context) {
	var req dto.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request")
		return
	}

	if !utility.ValidCredentials(req.Username, req.Password) {
		response.Error(c, http.StatusBadRequest, "username and password are required")
		return
	}

	if req.Role != "" && !utility.ValidRole(req.Role) {
		response.Error(c, http.StatusBadRequest, "invalid role")
		return
	}

	if err := ctrl.Service.Register(req); err != nil {
		response.Error(c, http.StatusConflict, "could not register user")
		return
	}

	response.Success(c, http.StatusCreated, "user registered", nil)
}

func (ctrl *AuthController) Login(c *gin.Context) {
	var req dto.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request")
		return
	}

	if !utility.ValidCredentials(req.Username, req.Password) {
		response.Error(c, http.StatusBadRequest, "username and password are required")
		return
	}

	token, role, err := ctrl.Service.Login(req)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "invalid username or password")
		return
	}

	response.Success(c, http.StatusOK, "login successful", gin.H{
		"token": token,
		"role":  role,
	})
}
