package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"net/http"
	"order-management/shared/discovery"

	"order-management/services/auth/config"
	"order-management/services/auth/controller"
	"order-management/services/auth/middleware"
	"order-management/services/auth/model"
	"order-management/services/auth/repository"
	"order-management/services/auth/service"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatal("auth database connection failed: ", err)
	}

	if err := db.AutoMigrate(&model.User{}); err != nil {
		log.Fatal("auth database migration failed: ", err)
	}

	repo := repository.NewUserRepository(db)
	authService := service.NewAuthService(repo, cfg.JWTSecret, cfg.JWTExpiry)
	authController := controller.NewAuthController(authService)

	router := gin.New()
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.RecoveryMiddleware())
	router.Use(middleware.LoggingMiddleware())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "auth", "status": "ok"})
	})

	router.POST("/auth/register", authController.Register)
	router.POST("/auth/login", authController.Login)

	consulClient, err := discovery.NewConsulClient()
	if err != nil {
		log.Fatal("Consul client initialization failed: ", err)
	}
	if err := consulClient.RegisterService("auth", cfg.Port); err != nil {
		log.Fatal("Consul registration failed: ", err)
	}
	log.Println("Auth service registered with Consul")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-stop
		_ = consulClient.DeregisterService("auth", cfg.Port)
		os.Exit(0)
	}()

	log.Println("Auth service running on port", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
