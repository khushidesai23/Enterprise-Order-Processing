# Enterprise Order Processing System

A production-inspired **Enterprise Order Processing System** built with **Go (Golang)** following **Clean Architecture**, **SOLID principles**, and enterprise backend development practices.

The project simulates a complete e-commerce order lifecycle, including inventory reservation, payment processing with **Razorpay**, webhook handling, JWT authentication, and transactional consistency.

---

# Features

## Authentication

* JWT-based authentication
* Secure login
* Protected API routes
* Swagger Bearer authentication support

## User Management

* User registration
* User CRUD operations
* Password hashing using bcrypt
* Input validation

## Category Management

* Category CRUD operations

## Product Management

* Product CRUD operations
* Category mapping
* Product availability management

## Inventory Management

* Inventory CRUD operations
* Stock reservation
* Stock confirmation
* Stock release
* Transaction-safe inventory updates

## Order Management

* Multi-item order creation
* Order lifecycle management
* Automatic inventory reservation
* Order cancellation
* Order status validation

## Payment Module

* Razorpay Order creation
* Razorpay Checkout integration
* Checkout signature verification
* Secure webhook verification
* Idempotent webhook processing
* Payment status synchronization
* Automatic order status update after successful payment
* Refund-ready architecture

## API Documentation

* Swagger UI integration
* Interactive API testing
* JWT Authorization support

---

# Tech Stack

| Category        | Technology         |
| --------------- | ------------------ |
| Language        | Go                 |
| Framework       | Gin                |
| ORM             | GORM               |
| Database        | PostgreSQL         |
| Authentication  | JWT                |
| Payment Gateway | Razorpay           |
| Configuration   | Viper              |
| Logging         | Zap                |
| Documentation   | Swagger (Swaggo)   |
| API             | REST               |

---

# Project Structure

```text
.
├── cmd/
│   └── server/
│       └── main.go
│
├── config/
│
├── docs/
│
├── internal/
│   ├── api/
│   ├── database/
│   ├── models/
│   ├── repository/
│   └── modules/
│       ├── auth/
│       ├── user/
│       ├── category/
│       ├── product/
│       ├── inventory/
│       ├── order/
│       └── payment/
│
├── pkg/
│
└── README.md
```

---

# Architecture

```text
                Client
                   │
                   ▼
            Gin HTTP Router
                   │
                   ▼
        JWT Authentication Middleware
                   │
                   ▼
              Route Handlers
                   │
                   ▼
                Services
                   │
                   ▼
             Repository Layer
                   │
                   ▼
               PostgreSQL
```

---

# Order Lifecycle

```text
     Created
        │
        ▼
 Payment Pending
        │
        ▼
      Paid
        │
        ▼
      Packed
        │
        ▼
     Shipped
        │
        ▼
    Delivered
```

Orders can be cancelled before shipment.

---

# Payment Flow

```text
  Create Order
        │
        ▼
  Reserve Inventory
        │
        ▼
  Create Razorpay Order
        │
        ▼
  Customer Completes Payment
        │
        ▼
  Verify Checkout Signature
        │
        ▼
  Receive Razorpay Webhook
        │
        ▼
  Verify Webhook Signature
        │
        ▼
  Persist Webhook
        │
        ▼
  Update Payment Status
        │
        ▼
  Update Order Status
        │
        ▼
  Commit Transaction
```

---

# Inventory Flow

```text
Available = 100
Reserved  = 0
↓
Order Created

Available = 98
Reserved  = 2
↓
Payment Pending

Available = 98
Reserved  = 2
↓
Payment Successful

Available = 98
Reserved  = 2
↓
Order Shipped

Available = 98
Reserved  = 0
```

If an order is cancelled before shipment:

```text
Available += Reserved Quantity
Reserved = 0
```

---

# Business Rules

* Inventory is reserved immediately after order creation.
* Reserved inventory is confirmed when the order is shipped.
* Cancelled orders release reserved inventory.
* Payment processing is idempotent.
* Duplicate webhook deliveries are ignored.
* Razorpay webhook signatures are verified.
* Checkout signatures are verified before accepting successful payments.
* Order status transitions are validated.
* All critical database operations are transaction-safe.

---

# API Modules

* Authentication
* User
* Category
* Product
* Inventory
* Order
* Payment

---

# Running the Project

## Clone Repository

```bash
git clone https://github.com/<your-username>/Enterprise-Order-Processing.git

cd Enterprise-Order-Processing
```

---

## Install Dependencies

```bash
go mod tidy
```

---

## Configure Environment

Create a `.env` file.

```env
APP_ENV=development

SERVER_PORT=8080

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=order_processing
DB_SSLMODE=disable

JWT_SECRET=your-secret-key
JWT_EXPIRATION=24h

RAZORPAY_KEY_ID=rzp_test_xxxxxxxxx
RAZORPAY_KEY_SECRET=xxxxxxxxxxxxxxxx
RAZORPAY_WEBHOOK_SECRET=xxxxxxxxxxxxxxxx
```

---

## Start PostgreSQL

Ensure PostgreSQL is running.

---

## Run the Application

```bash
go run ./cmd/server
```

---

# Swagger API Documentation

Generate Swagger files:

```bash
swag init -g ./cmd/server/main.go --parseInternal --parseDependency
```

Open Swagger UI:

```text
http://localhost:8080/swagger/index.html
```

Login using `/auth/login`, copy the JWT, click **Authorize**, and enter:

```text
Bearer <your-jwt-token>
```

---

# Razorpay Setup

1. Create a Razorpay Test Account.
2. Generate Test API Keys.
3. Add the keys to the `.env` file.
4. Configure a Webhook in the Razorpay Dashboard.

Webhook URL:

```text
https://<your-ngrok-url>/api/v1/payments/webhook
```

Events to subscribe:

* payment.captured
* payment.failed
* refund.created

Copy the generated **Webhook Secret** into:

```env
RAZORPAY_WEBHOOK_SECRET=xxxxxxxxxxxxxxxx
```

---

# Local Webhook Testing with ngrok

Since Razorpay cannot send webhooks to `localhost`, expose your local server using ngrok.

Start ngrok:

```bash
ngrok http 8080
```

Example:

```text
https://abcd-1234.ngrok-free.app
```

Configure Razorpay Webhook:

```text
https://abcd-1234.ngrok-free.app/api/v1/payments/webhook
```

---

# Payment Testing Flow

1. Register/Login
2. Copy JWT from `/auth/login`
3. Authorize in Swagger
4. Create User (if needed)
5. Create Product
6. Create Inventory
7. Create Order
8. Create Payment
9. Complete Razorpay Checkout
10. Razorpay sends Webhook
11. Payment status updates automatically
12. Order status changes to **PAID**

---

# Author

**Khushi Desai**

Associate Software Engineer | Cloud & Backend Development

**GitHub:** https://github.com/khushidesai23

**LinkedIn:** https://www.linkedin.com/in/khushi-desai-ab5154225/
