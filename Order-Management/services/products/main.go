package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"net/http"
	"order-management/shared/discovery"

	"order-management/services/products/config"
	"order-management/services/products/controller"
	"order-management/services/products/middleware"
	"order-management/services/products/model"
	"order-management/services/products/repository"
	"order-management/services/products/service"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatal("product database connection failed: ", err)
	}

	// Reuse database connections instead of opening a new connection per request.
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("could not get database connection pool: ", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	if err := db.AutoMigrate(&model.Product{}); err != nil {
		log.Fatal("product migration failed: ", err)
	}

	repo := repository.NewProductRepository(db)
	productService := service.NewProductService(repo)
	productController := controller.NewProductController(productService)

	router := gin.New()
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.RecoveryMiddleware())
	router.Use(middleware.LoggingMiddleware())
	router.Use(middleware.SecurityMiddleware())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "products", "status": "ok"})
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

	router.POST("/products", productController.Create)
	router.GET("/products", productController.GetAll)
	router.GET("/products/:id", productController.GetByID)
	// Internal service-to-service operation; user-facing order APIs stay in the Gateway.
	router.POST("/internal/products/:id/stock", productController.ChangeStock)

	consulClient, err := discovery.NewConsulClient()
	if err != nil {
		log.Fatal("Consul client initialization failed: ", err)
	}
	if err := consulClient.RegisterService("products", cfg.Port); err != nil {
		log.Fatal("Consul registration failed: ", err)
	}
	log.Println("Product service registered with Consul")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-stop
		_ = consulClient.DeregisterService("products", cfg.Port)
		os.Exit(0)
	}()

	log.Println("Product service running on port", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
