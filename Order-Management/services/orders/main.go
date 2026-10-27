package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"net/http"
	"order-management/shared/discovery"

	"order-management/services/orders/client"
	"order-management/services/orders/config"
	"order-management/services/orders/controller"
	"order-management/services/orders/middleware"
	"order-management/services/orders/model"
	"order-management/services/orders/repository"
	"order-management/services/orders/service"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatal("order database connection failed: ", err)
	}

	// Reuse database connections instead of opening a new connection per request.
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("could not get database connection pool: ", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	if err := db.AutoMigrate(&model.Order{}); err != nil {
		log.Fatal("order database migration failed: ", err)
	}

	repo := repository.NewOrderRepository(db)
	consul, err := discovery.NewConsulClient()
	if err != nil {
		log.Fatal("Consul client initialization failed: ", err)
	}
	productClient := client.NewProductClientWithDiscovery(consul)
	orderService := service.NewOrderService(repo, productClient)
	orderController := controller.NewOrderController(orderService)

	router := gin.New()
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.RecoveryMiddleware())
	router.Use(middleware.LoggingMiddleware())
	router.Use(middleware.SecurityMiddleware())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "orders", "status": "ok"})
	})

	router.GET("/health/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "live"})
	})
	router.GET("/health/ready", func(c *gin.Context) {
		if err := sqlDB.PingContext(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	router.POST("/orders", orderController.Create)
	router.GET("/orders", orderController.GetAll)
	router.GET("/orders/:id", orderController.GetByID)
	router.PUT("/orders/:id/status", orderController.UpdateStatus)

	consulClient, err := discovery.NewConsulClient()
	if err != nil {
		log.Fatal("Consul client initialization failed: ", err)
	}
	if err := consulClient.RegisterService("orders", cfg.Port); err != nil {
		log.Fatal("Consul registration failed: ", err)
	}
	log.Println("Order service registered with Consul")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-stop
		_ = consulClient.DeregisterService("orders", cfg.Port)
		os.Exit(0)
	}()

	log.Println("Order service running on port", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
