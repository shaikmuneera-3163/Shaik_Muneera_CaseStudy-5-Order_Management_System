package service

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"order-management/services/auth/model"
	"order-management/services/auth/repository"
	"order-management/shared/dto"

	"github.com/golang-jwt/jwt/v5"
)

type AuthService struct {
	Repo      *repository.Repository
	JWTSecret string
	Expiry    int
}

func NewAuthService(repo *repository.Repository, secret string, expiry int) *AuthService {
	return &AuthService{
		Repo:      repo,
		JWTSecret: secret,
		Expiry:    expiry,
	}
}

func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

func (s *AuthService) Register(req dto.RegisterRequest) error {
	role := req.Role
	if role == "" {
		role = "user"
	}

	if role != "admin" && role != "user" {
		return errors.New("invalid role")
	}

	user := model.User{
		Username:     req.Username,
		PasswordHash: hashPassword(req.Password),
		Role:         role,
	}

	return s.Repo.Create(&user)
}

func (s *AuthService) Login(req dto.LoginRequest) (string, string, error) {
	user, err := s.Repo.FindByUsername(req.Username)
	if err != nil {
		return "", "", errors.New("invalid username or password")
	}

	if user.PasswordHash != hashPassword(req.Password) {
		return "", "", errors.New("invalid username or password")
	}

	claims := jwt.MapClaims{
		"sub":  user.ID,
		"role": user.Role,
		"exp":  time.Now().Add(time.Duration(s.Expiry) * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(s.JWTSecret))
	if err != nil {
		return "", "", err
	}

	return tokenString, user.Role, nil
}
