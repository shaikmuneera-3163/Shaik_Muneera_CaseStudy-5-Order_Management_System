package graphql

import (
	"context"
	"fmt"

	"order-management/gateway/client"
	ordermodel "order-management/services/orders/model"
	productmodel "order-management/services/products/model"
	"order-management/shared/dto"

	gql "github.com/graphql-go/graphql"
)

func NewSchema(products *client.ProductClient, orders *client.OrderClient) (gql.Schema, error) {
	productType := gql.NewObject(gql.ObjectConfig{
		Name: "Product",
		Fields: gql.Fields{
			"id":    &gql.Field{Type: gql.Int},
			"name":  &gql.Field{Type: gql.String},
			"price": &gql.Field{Type: gql.Float},
			"stock": &gql.Field{Type: gql.Int},
		},
	})

	orderType := gql.NewObject(gql.ObjectConfig{
		Name: "Order",
		Fields: gql.Fields{
			"id":           &gql.Field{Type: gql.Int},
			"customer_id":  &gql.Field{Type: gql.Int},
			"product_id":   &gql.Field{Type: gql.Int},
			"product_name": &gql.Field{Type: gql.String},
			"quantity":     &gql.Field{Type: gql.Int},
			"unit_price":   &gql.Field{Type: gql.Float},
			"total_amount": &gql.Field{Type: gql.Float},
			"status":       &gql.Field{Type: gql.String},
			"created_at": &gql.Field{
				Type: gql.String,
				Resolve: func(p gql.ResolveParams) (interface{}, error) {
					if o, ok := p.Source.(ordermodel.Order); ok {
						return o.CreatedAt.Format("2006-01-02T15:04:05Z07:00"), nil
					}
					if o, ok := p.Source.(*ordermodel.Order); ok {
						return o.CreatedAt.Format("2006-01-02T15:04:05Z07:00"), nil
					}
					return nil, nil
				},
			},
		},
	})

	query := gql.NewObject(gql.ObjectConfig{
		Name: "Query",
		Fields: gql.Fields{
			// Get all products
			"products": &gql.Field{
				Type: gql.NewList(productType),
				Resolve: func(p gql.ResolveParams) (interface{}, error) {
					ctx := p.Context
					if ctx == nil {
						ctx = context.Background()
					}
					return products.GetAll(ctx)
				},
			},

			// NEW: Get product by ID
			"product": &gql.Field{
				Type: productType,
				Args: gql.FieldConfigArgument{
					"id": &gql.ArgumentConfig{
						Type: gql.NewNonNull(gql.Int),
					},
				},
				Resolve: func(p gql.ResolveParams) (interface{}, error) {
					id, ok := p.Args["id"].(int)
					if !ok || id < 1 {
						return nil, fmt.Errorf("id must be a positive integer")
					}

					ctx := p.Context
					if ctx == nil {
						ctx = context.Background()
					}

					return products.GetByID(ctx, uint(id))
				},
			},

			// Get all orders
			"orders": &gql.Field{
				Type: gql.NewList(orderType),
				Resolve: func(p gql.ResolveParams) (interface{}, error) {
					ctx := p.Context
					if ctx == nil {
						ctx = context.Background()
					}
					return orders.GetAll(ctx)
				},
			},

			// Get order by ID
			"order": &gql.Field{
				Type: orderType,
				Args: gql.FieldConfigArgument{
					"id": &gql.ArgumentConfig{
						Type: gql.NewNonNull(gql.Int),
					},
				},
				Resolve: func(p gql.ResolveParams) (interface{}, error) {
					id, ok := p.Args["id"].(int)
					if !ok || id < 1 {
						return nil, fmt.Errorf("id must be a positive integer")
					}

					ctx := p.Context
					if ctx == nil {
						ctx = context.Background()
					}

					return orders.GetByID(ctx, uint(id))
				},
			},
		},
	})

	mutation := gql.NewObject(gql.ObjectConfig{
		Name: "Mutation",
		Fields: gql.Fields{
			// NEW: Create product
			"createProduct": &gql.Field{
				Type: productType,
				Args: gql.FieldConfigArgument{
					"name": &gql.ArgumentConfig{
						Type: gql.NewNonNull(gql.String),
					},
					"price": &gql.ArgumentConfig{
						Type: gql.NewNonNull(gql.Float),
					},
					"stock": &gql.ArgumentConfig{
						Type: gql.NewNonNull(gql.Int),
					},
				},
				Resolve: func(p gql.ResolveParams) (interface{}, error) {
					name, _ := p.Args["name"].(string)
					price, _ := p.Args["price"].(float64)
					stock, _ := p.Args["stock"].(int)

					if name == "" || price <= 0 || stock < 0 {
						return nil, fmt.Errorf("name must not be empty, price must be positive, and stock cannot be negative")
					}

					ctx := p.Context
					if ctx == nil {
						ctx = context.Background()
					}

					result, err := products.Create(ctx, productmodel.Product{
						Name:  name,
						Price: price,
						Stock: stock,
					})
					if err != nil {
						return nil, err
					}

					// The Product service returns a response envelope.
					// Extract its "data" field for the GraphQL Product type.
					if response, ok := result.(map[string]interface{}); ok {
						if data, exists := response["data"]; exists {
							return data, nil
						}
					}

					return result, nil
				},
			},

			// Create order
			"createOrder": &gql.Field{
				Type: orderType,
				Args: gql.FieldConfigArgument{
					"customerId": &gql.ArgumentConfig{
						Type: gql.NewNonNull(gql.Int),
					},
					"productID": &gql.ArgumentConfig{
						Type: gql.NewNonNull(gql.Int),
					},
					"quantity": &gql.ArgumentConfig{
						Type: gql.NewNonNull(gql.Int),
					},
				},
				Resolve: func(p gql.ResolveParams) (interface{}, error) {
					customerID, _ := p.Args["customerId"].(int)
					productID, _ := p.Args["productID"].(int)
					quantity, _ := p.Args["quantity"].(int)

					if customerID < 1 || productID < 1 || quantity < 1 {
						return nil, fmt.Errorf("customerId, productID and quantity must be positive")
					}

					ctx := p.Context
					if ctx == nil {
						ctx = context.Background()
					}

					return orders.Create(ctx, dto.OrderRequest{
						CustomerID: uint(customerID),
						ProductID:  uint(productID),
						Quantity:   quantity,
					})
				},
			},

			// Update order status
			"updateOrderStatus": &gql.Field{
				Type: orderType,
				Args: gql.FieldConfigArgument{
					"id": &gql.ArgumentConfig{
						Type: gql.NewNonNull(gql.Int),
					},
					"status": &gql.ArgumentConfig{
						Type: gql.NewNonNull(gql.String),
					},
				},
				Resolve: func(p gql.ResolveParams) (interface{}, error) {
					id, _ := p.Args["id"].(int)
					status, _ := p.Args["status"].(string)

					if id < 1 {
						return nil, fmt.Errorf("id must be a positive integer")
					}

					ctx := p.Context
					if ctx == nil {
						ctx = context.Background()
					}

					return orders.UpdateStatus(ctx, uint(id), status)
				},
			},
		},
	})

	return gql.NewSchema(gql.SchemaConfig{
		Query:    query,
		Mutation: mutation,
	})
}
