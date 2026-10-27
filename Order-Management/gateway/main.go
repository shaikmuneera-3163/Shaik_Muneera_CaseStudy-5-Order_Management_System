package main

import (
	"log"
	"os"

	"order-management/gateway/client"
	"order-management/gateway/config"
	"order-management/gateway/controller"
	gql "order-management/gateway/graphql"
	"order-management/gateway/middleware"
	"order-management/shared/discovery"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()
	if os.Getenv("JWT_SECRET") == "" {
		log.Fatal("JWT_SECRET is missing from .env")
	}

	consul, err := discovery.NewConsulClient()
	if err != nil {
		log.Fatal("Consul client initialization failed: ", err)
	}

	authClient := client.NewAuthClientWithDiscovery(consul)
	productClient := client.NewProductClientWithDiscovery(consul)
	orderClient := client.NewOrderClientWithDiscovery(consul)
	ctrl := controller.NewGatewayController(authClient, productClient, orderClient)

	schema, err := gql.NewSchema(productClient, orderClient)
	if err != nil {
		log.Fatal("could not create GraphQL schema: ", err)
	}

	router := gin.New()
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.RecoveryMiddleware())
	router.Use(middleware.LoggingMiddleware())
	router.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"service": "gateway", "status": "ok"}) })
	router.GET("/health/live", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "live"})
	})
	router.GET("/health/ready", func(c *gin.Context) {
		for _, serviceName := range []string{"auth", "products", "orders"} {
			if _, err := consul.GetServiceURL(serviceName); err != nil {
				c.JSON(503, gin.H{"status": "not ready", "dependency": serviceName})
				return
			}
		}
		c.JSON(200, gin.H{"status": "ready"})
	})

	router.POST("/auth/login", ctrl.Login)

	protected := router.Group("/")
	protected.Use(middleware.AuthMiddleware())
	protected.GET("/products", ctrl.GetProducts)
	protected.GET("/products/:id", ctrl.GetProductByID)
	protected.POST("/products", middleware.RequireRole("admin"), ctrl.CreateProduct)
	protected.POST("/api/v1/orders", ctrl.CreateOrder)
	protected.GET("/api/v1/orders", ctrl.GetOrders)
	protected.GET("/api/v1/orders/:id", ctrl.GetOrderByID)
	protected.PUT("/api/v1/orders/:id/status", ctrl.UpdateOrderStatus)

	// GraphiQL UI is available at GET /graphql. GraphQL operations are sent to POST /graphql.
	router.POST("/graphql", gql.NewHandler(schema))
	router.GET("/graphql", gql.NewHandler(schema))

	log.Println("Gateway running on port", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
