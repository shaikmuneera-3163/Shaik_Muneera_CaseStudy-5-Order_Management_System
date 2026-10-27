# Production-Ready Order Management System

## 1. Project Overview

This project implements a microservices-based Order Management System using Go, Gin, PostgreSQL, GORM, REST APIs, GraphQL, Consul, JWT authentication, and Role-Based Access Control (RBAC).

The application contains four main components:

- **Gateway Service:** Entry point for client requests. Routes requests to the appropriate service and applies authentication and authorization.
- **Auth Service:** Handles user registration, login, password verification, and JWT generation.
- **Product Service:** Manages products, product details, prices, and stock.
- **Order Service:** Creates orders, retrieves orders, lists orders, and updates order statuses.

The project also includes environment configuration, logging, request IDs, recovery middleware, health checks, service discovery, and resilience mechanisms.

## 2. Technologies Used

- Go
- Gin Web Framework
- PostgreSQL
- GORM
- REST API
- GraphQL
- JWT Authentication
- Role-Based Access Control (RBAC)
- Consul Service Discovery
- Environment Variables
- Logging and Middleware
- Resilience: Timeouts, Retries, and Circuit Breaker Components

## 3. Service Ports

| Service | Port | Base URL |
|---|---:|---|
| Gateway | 8080 | `http://localhost:8080` |
| Product Service | 8081 | `http://localhost:8081` |
| Order Service | 8082 | `http://localhost:8082` |
| Auth Service | 8083 | `http://localhost:8083` |
| Consul | 8500 | `http://127.0.0.1:8500` |


## 4. Database Setup

Open PostgreSQL using `psql` or pgAdmin and create the databases:

```sql
CREATE DATABASE productdb;
CREATE DATABASE orderdb;
CREATE DATABASE authdb;
## 6. Environment Configuration

Create or update the `.env` file in the project root.

Configure the variables using the names expected by your application's configuration files:

```env
GATEWAY_PORT=8080

AUTH_SERVICE_PORT=8083
JWT_SECRET=replace_with_a_long_random_secret
JWT_EXPIRY_MINUTES=60

PRODUCT_SERVICE_PORT=8081
PRODUCT_DB_HOST=localhost
PRODUCT_DB_PORT=5432
PRODUCT_DB_USER=postgres
PRODUCT_DB_PASSWORD=your_postgres_password
PRODUCT_DB_NAME=productdb

ORDER_SERVICE_PORT=8082
ORDER_DB_HOST=localhost
ORDER_DB_PORT=5432
ORDER_DB_USER=postgres
ORDER_DB_PASSWORD=your_postgres_password
ORDER_DB_NAME=orderdb

CONSUL_HTTP_ADDR=http://127.0.0.1:8500
```

## 5. How to Run the Entire Project

Open the `Order-Management` project folder in VS Code.

### Step 1: Open the project directory

Open Command Prompt or the VS Code terminal and navigate to your project root:

```cmd
cd C:\Users\MUNEERA\OneDrive\Desktop\final-project\Order-Management
```

Use your actual project location if it is different.

### Step 2: Download dependencies

Run this command from the project root:

```cmd
go mod tidy
```

### Step 3: Start Consul

Open a separate terminal and run:

```cmd
consul agent -dev -client=127.0.0.1
```

Keep this terminal running.

Open the Consul dashboard in your browser:

`http://127.0.0.1:8500/ui`

Check that Consul is running and that the services register successfully.

### Step 4: Start the Auth Service

Open a new terminal:

```cmd
cd C:\Users\MUNEERA\OneDrive\Desktop\final-project\Order-Management
go run ./services/auth
```

Keep this terminal running. The Auth Service should start on port `8083`.

### Step 5: Start the Product Service

Open another terminal:
go run ./services/products
The Product Service should start on port `8081`.

### Step 6: Start the Order Service

Open another terminal:
go run ./services/orders
The Order Service should start on port `8082`.

### Step 7: Start the Gateway
 The Gateway should start on port `8080`.

## 8. Test the Gateway Health

Open SoapUI and create a REST request.

**Request: Gateway Health**

- Method: `GET`
- URL: `http://localhost:8080/health`

Expected response:

```json
{
  "service": "gateway",
  "status": "ok"
}
```

Also check the liveness and readiness endpoints:

**Liveness**

`GET http://localhost:8080/health/live`

**Readiness**

`GET http://localhost:8080/health/ready`

## 9. SoapUI REST API Testing

### Important: Authentication

Most product and order endpoints are protected by JWT authentication.

First, log in through the Gateway and copy the token returned in the successful response. Add the following header to protected requests:

```text
Authorization: Bearer YOUR_JWT_TOKEN
```

For requests containing JSON, also add:

```text
Content-Type: application/json
```

Replace `YOUR_JWT_TOKEN` with your actual token.

### 9.1 Login
`POST http://localhost:8083/auth/register`
```json
{"username":"admin1","password":"Admin@123","role":"admin"}
```
Registration is currently exposed directly on the Auth service, not proxied through the Gateway. Use a test-only admin account for local learning.

### 2. Login through Gateway
`POST http://localhost:8080/auth/login`
```json
{"username":"admin1","password":"Admin@123"}
### 9.2 Create a Product

**Method:** `POST`

**URL:**

`http://localhost:8080/products`

**Headers:**

```text
Content-Type: application/json
Authorization: Bearer YOUR_JWT_TOKEN
```

**Request Body:**

```json
{
  "name": "Laptop",
  "price": 55000,
  "stock": 10
}
```

This operation requires a valid JWT belonging to a user with the `admin` role.

### 9.3 Get All Products

**Method:** `GET`

**URL:**

`http://localhost:8080/products`

**Header:**

```text
Authorization: Bearer YOUR_JWT_TOKEN
```

This request retrieves the available products.

### 9.4 Get a Product by ID

**Method:** `GET`

**URL:**

`http://localhost:8080/products/1`

**Header:**

```text
Authorization: Bearer YOUR_JWT_TOKEN
```

Replace `1` with the actual product ID.

### 9.5 Create an Order

**Method:** `POST`

**URL:**

`http://localhost:8080/api/v1/orders`

**Headers:**

```text
Content-Type: application/json
Authorization: Bearer YOUR_JWT_TOKEN
```

**Request Body:**

```json
{
  "customerId": 101,
  "productId": 1,
  "quantity": 2
}
```

Use an existing product ID with sufficient stock. Save the order ID returned in the response because it will be used in the following requests.

### 9.6 Get All Orders

**Method:** `GET`

**URL:**

`http://localhost:8080/api/v1/orders`

**Header:**

```text
Authorization: Bearer YOUR_JWT_TOKEN
```

This request retrieves the orders available through the Order Service.

### 9.7 Get an Order by ID

**Method:** `GET`

**URL:**

`http://localhost:8080/api/v1/orders/1`

**Header:**

```text
Authorization: Bearer YOUR_JWT_TOKEN
```

Replace `1` with an order ID returned by the Create Order request.

### 9.8 Update Order Status

**Method:** `PUT`

**URL:**

`http://localhost:8080/api/v1/orders/1/status`

**Headers:**

```text
Content-Type: application/json
Authorization: Bearer YOUR_JWT_TOKEN
```

**Request Body:**

```json
{
  "status": "CONFIRMED"
}
```

Replace `1` with your actual order ID.

The status values supported by the project are:

- `CREATED`
- `CONFIRMED`
- `CANCELLED`

Use these values rather than `shipped`, unless you explicitly modify the application's status validation.

## 10. GraphQL Testing

### GraphQL UI

Open the following URL in your browser:

`http://localhost:8080/graphql`

If the GraphiQL editor is enabled, enter an operation in the left panel and click **Run**.

If your GraphQL endpoint requires authentication, configure the HTTP headers in GraphiQL to include your JWT token. If the browser UI does not expose an HTTP headers editor, you can send GraphQL requests through SoapUI using `POST http://localhost:8080/graphql`.

**Important:** The following GraphQL operations use the field names and arguments defined in your project's GraphQL schema. GraphQL is case-sensitive. If an operation reports an unknown field or argument, compare it with `gateway/graphql/schema.go` and use the exact names defined there.

### 10.1 Get All Orders

```graphql
query {
  orders {
    id
    product_id
    product_name
    quantity
    unit_price
    total_amount
    status
    created_at
  }
}
```

This query retrieves the orders and their available fields.

### 10.2 Get One Order

```graphql
query {
  order(id: 1) {
    id
    product_name
    quantity
    total_amount
    status
  }
}
```

Replace `1` with the ID of an existing order.

### 10.3 Create an Order

```graphql
mutation {
  createOrder(productID: 1, quantity: 2) {
    id
    product_id
    product_name
    quantity
    total_amount
    status
  }
}
```

This mutation creates an order for the specified product and quantity, if the arguments match the schema.

### 10.4 Update Order Status

```graphql
mutation {
  updateOrderStatus(id: 1, status: "CONFIRMED") {
    id
    status
    total_amount
  }
}
```

Replace `1` with an existing order ID. The status must be supported by the application.

### 10.5 Get Products

```graphql
query {
  products {
    id
    name
    price
    stock
  }
}
```
A. Get product by ID
query {
  product(id: 1) {
    id
    name
    price
    stock
  }
}


B. Create product
mutation {
  createProduct(name: "Keyboard", price: 2500, stock: 20) {
    id
    name
    price
    stock
  }
}

**GraphQL note:** These examples are intended to make testing straightforward, but the exact available fields and argument names must match the schema implemented in your current code. In particular, check whether `createOrder` requires a `customerId` argument and whether the schema exposes `product_name`, `unit_price`, and `created_at` before using those fields.

## 11. Consul Checks

### Consul UI

Open:

`http://127.0.0.1:8500/ui`

Check the registered services and their health status.

### Service Health Endpoints

**Gateway**

`GET http://localhost:8080/health`

**Product Service**

`GET http://localhost:8081/health`

**Order Service**

`GET http://localhost:8082/health`

**Auth Service**

`GET http://localhost:8083/health`

Each service should return HTTP 200 when it is running and its health endpoint is healthy.

If the Gateway reports that a service is unavailable:

1. Confirm that the relevant service terminal is still running.
2. Open that service's `/health` endpoint.
3. Check the Consul UI to see whether the service is registered and healthy.
4. Verify that the Consul address and registered service ports match your `.env` configuration.
5. Check the service and Gateway terminal logs for connection errors.

## 12. Middleware, Logging, and Resilience

The project includes or is structured to support the following features:

- **JWT Authentication:** Validates tokens before allowing access to protected routes.
- **Role-Based Access Control:** Restricts admin-only operations such as product creation.
- **Request ID Middleware:** Adds a request identifier to help trace a request through logs.
- **Recovery Middleware:** Helps recover from panics and prevents a panic from terminating the entire server process.
- **Logging:** Records request and application information to help troubleshoot failures.
- **Health Checks:** Provide health, liveness, and readiness endpoints.
- **Service Discovery:** Uses Consul to discover downstream services.
- **Resilience:** Includes retry, timeout, and circuit-breaker-related components to improve handling of temporary service failures.

The presence of a component does not necessarily mean every feature is fully integrated into every request path. Verify the actual middleware registration and downstream client implementation before claiming complete production readiness.

## 13. Run Tests

Run the following command from the project root:

```cmd
go test ./...
```

Run tests with the race detector:

```cmd
go test -race ./...
```

Run tests with verbose output:

```cmd
go test -v ./...
```

These commands execute the tests available in the repository. Confirm that database-backed integration tests are configured before interpreting the results as proof that the full application works end to end.

## 14. Common Errors and Solutions

### 404 Not Found

Check the method, port, and URL.

The correct Create Order request through the Gateway is:

`POST http://localhost:8080/api/v1/orders`

Do not use `/orders` for this Gateway route.

### 401 Unauthorized

- Log in first.
- Copy the actual JWT token from the login response.
- Add `Authorization: Bearer YOUR_JWT_TOKEN`.
- Check that the Gateway and Auth Service use the same JWT secret.

### 403 Forbidden

The token may be valid, but the user may not have the required role. Product creation requires the admin role.

### 502 Bad Gateway or Service Unavailable

- Check whether the downstream service is running.
- Test its direct health endpoint.
- Check Consul service registration.
- Review Gateway and service logs.
- Verify the configured service addresses and ports.

### Database Connection Error

- Confirm that PostgreSQL is running.
- Verify the database names, usernames, passwords, and port in `.env`.
- Confirm that `productdb` and `orderdb` exist.
- Check the service startup logs for GORM or PostgreSQL errors.

### GraphQL Unknown Field or Argument

GraphQL names are case-sensitive. Compare the operation with `gateway/graphql/schema.go` and use the exact field and argument names defined there.

## 15. Recommended Testing Order

Follow this sequence when demonstrating the project:

1. Start PostgreSQL.
2. Start Consul.
3. Start Auth Service.
4. Start Product Service.
5. Start Order Service.
6. Start Gateway.
7. Check Gateway health.
8. Log in and obtain a JWT token.
9. Create a product using an admin account, or confirm that a suitable product already exists.
10. List products.
11. Create an order.
12. List orders.
13. Retrieve the created order by ID.
14. Update the order status.
15. Run the GraphQL queries and mutations.
16. Check Consul service health.
17. Run `go test ./...` and `go test -race ./...`.

## 16. Important Notes

- Use Gateway port `8080` for normal REST and GraphQL requests.
- Keep all service terminals running while testing.
- Use a valid JWT token for protected endpoints.
- Use real product and order IDs returned by the application.
- The Product Service and Order Service use separate PostgreSQL databases.
- Keep real secrets out of Git.
- AWS deployment, CI/CD, metrics, distributed tracing, and production security configurations require separate implementation and verification if they are not already present in the repository.

## Conclusion

This project demonstrates a Go-based Order Management System with separate Auth, Product, Order, and Gateway components. It provides REST and GraphQL interfaces, PostgreSQL persistence, JWT-based security, Consul service discovery, middleware, health checks, and resilience-related components.

The README provides the commands and request examples needed to start the project and test the available APIs using SoapUI and GraphQL.
