# Enterprise Order Processing System

A production-inspired **Order Processing System** built with **Golang**, following **Clean Architecture**, **SOLID principles**, and enterprise backend development practices.

The project simulates a real-world e-commerce order lifecycle, including inventory reservation, payment processing, webhook handling, and transactional consistency.

---

## Features

### User Management
- User registration
- Authentication-ready architecture
- Profile management
- Input validation

### Product Management
- Product CRUD operations
- Category management
- Product pricing
- Product availability

### Inventory Management
- Stock reservation
- Stock confirmation
- Stock release
- Inventory tracking
- Transaction-safe inventory operations

### Order Management
- Order creation
- Multi-item orders
- Order status management
- Order lifecycle validation
- Inventory reservation during checkout

### Payment Module
- Payment creation
- Payment status tracking
- Razorpay integration
- Secure webhook verification
- Idempotent webhook processing
- Payment failure handling
- Refund-ready architecture

### Reliability
- ACID-compliant transactions
- Repository pattern
- Service layer architecture
- Idempotent webhook processing
- Optimistic business workflow
- Proper error handling

---

# Tech Stack

| Category | Technology |
|-----------|------------|
| Language | Go |
| Web Framework | Gin |
| ORM | GORM |
| Database | PostgreSQL |
| Payment Gateway | Razorpay |
| Configuration | Viper |
| Logging | zap |
| Authentication | JWT (Planned) |
| API Style | REST |
| Architecture | Clean Architecture |

---

# Project Structure

```
.
├── cmd/
│   └── server/
│       └── main.go
│
├── config/
│
├── internal/
│   ├── models/
│   ├── repository/
│   ├── middleware/
│   ├── routes/
│   ├── handlers/
│   ├── payment/
│   ├── order/
│   ├── inventory/
│   ├── product/
│   ├── category/
│   └── user/
│
├── pkg/
│
├── docs/
│
├── scripts/
│
└── migrations/
```

---

# Architecture

```
               Client
                  │
                  ▼
           REST API (Gin)
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

```
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

Cancellation can occur before shipment.

---

# Payment Flow

```
Create Order
      │
      ▼
Reserve Inventory
      │
      ▼
Create Payment
      │
      ▼
Razorpay Order Created
      │
      ▼
Customer Pays
      │
      ▼
Webhook Received
      │
      ▼
Verify Signature
      │
      ▼
Persist Webhook
      │
      ▼
Update Payment
      │
      ▼
Update Order
      │
      ▼
Commit Transaction
```

---

# Inventory Lifecycle

```
Initial

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

Paid

Available = 98
Reserved  = 2

↓

Packed

Available = 98
Reserved  = 2

↓

Shipped

Available = 98
Reserved  = 0

↓

Delivered
```

If payment fails or the order is cancelled before shipment:

```
Available += Reserved Quantity
Reserved = 0
```

---

# Business Rules

- Inventory is reserved immediately after order creation.
- Available inventory decreases during reservation.
- Reserved inventory is confirmed when the order is shipped.
- Cancelled orders release reserved inventory.
- Duplicate webhook deliveries are ignored.
- Payment processing is idempotent.
- Order state transitions are validated.
- Database operations are transaction-safe.

---

# API Modules

- User
- Category
- Product
- Inventory
- Order
- Payment

---

# Running the Project

## Clone

```bash
git clone https://github.com/<your-username>/Enterprise-Order-Processing.git
```

```bash
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

RAZORPAY_KEY_ID=
RAZORPAY_KEY_SECRET=
RAZORPAY_WEBHOOK_SECRET=
```

---

## Start PostgreSQL

Ensure PostgreSQL is running.

---

## Run

```bash
go run cmd/server/main.go
```

---

# Future Enhancements

- JWT Authentication
- Role-Based Access Control (RBAC)
- Redis Caching
- Background Job Processing
- Event-Driven Architecture
- Kafka Integration
- Outbox Pattern
- CDC with Debezium
- Prometheus Metrics
- Grafana Dashboards
- OpenTelemetry Tracing
- Docker Support
- Kubernetes Deployment
- CI/CD Pipeline
- Notification Service
- Email Service
- Inventory Alerts
- Multi-Warehouse Support
- Distributed Transactions
- Saga Pattern
- Payment Retry Mechanism

---

# Learning Objectives

This project demonstrates practical implementation of:

- Clean Architecture
- SOLID Principles
- Repository Pattern
- Dependency Injection
- Transaction Management
- Inventory Reservation Strategy
- Payment Gateway Integration
- Webhook Processing
- Idempotent APIs
- Enterprise Backend Design
- PostgreSQL with GORM
- REST API Development
- Service-Oriented Design

---

# Contributing

Contributions, suggestions, and improvements are welcome.

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Open a Pull Request

---

## Author

**Khushi Desai**

Associate Software Engineer | Cloud & Backend Development

GitHub: https://github.com/khushidesai23

LinkedIn: [https://www.linkedin.com/in/khushidesai23/](https://www.linkedin.com/in/khushi-desai-ab5154225/)
