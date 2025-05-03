# Order Service Technical Design

## 1. Introduction

The Order Service is a critical component of our e-commerce microservices architecture, responsible for managing the entire order lifecycle from creation to fulfillment. It serves as the central coordination point for order processing, interacting with multiple other services to ensure a seamless ordering experience.

## 2. Architecture

### 2.1 Service Architecture

The Order Service follows our standard microservice architecture pattern:

```
Client Request → API Gateway → Order Service → Database
```

Internal service-to-service communication uses gRPC, while external communication via the API Gateway uses REST.

### 2.2 Technology Stack

- **Language**: Go 1.24+
- **Framework**: Standard Go libraries with gRPC
- **Database**: PostgreSQL
- **Caching**: Redis
- **Message Broker**: (Future implementation for event-driven updates)
- **Containerization**: Docker
- **Configuration**: YAML and environment variables
- **Logging**: Zap logger

## 3. Data Model

### 3.1 Core Entities

#### Order
Represents the main order record:
- Contains order metadata (status, totals, timestamps)
- Links to user, payment, and shipping information
- Tracks order lifecycle through status changes

#### Order Item
Represents individual items within an order:
- Links to products and variants
- Contains pricing and quantity information
- Maintains a snapshot of product data at time of order

#### Order Address
Represents shipping and billing addresses:
- Contains address details
- Supports different address types (shipping/billing)
- Maintains contact information

#### Order Status History
Tracks the history of order status changes:
- Records status transitions
- Maintains audit trail with timestamps
- Includes optional notes for each status change

#### Order Payment Transaction
Records payment transactions related to an order:
- Tracks payment attempts and results
- Stores transaction IDs for reference
- Maintains payment gateway responses

### 3.2 Entity Relationships

```
Order
 ├── Order Items (1:N)
 ├── Order Addresses (1:N)
 ├── Order Status History (1:N)
 └── Order Payment Transactions (1:N)
```

## 4. Service Interfaces

### 4.1 gRPC Service Definition

The Order Service exposes the following gRPC methods:

#### Order Management
- `CreateOrder`: Creates a new order
- `GetOrder`: Retrieves an order by ID or order number
- `UpdateOrderStatus`: Updates the status of an order
- `CancelOrder`: Cancels an existing order
- `ListOrders`: Retrieves a paginated list of orders

#### Order Item Management
- `AddOrderItem`: Adds an item to an existing order
- `UpdateOrderItem`: Updates an item in an existing order
- `RemoveOrderItem`: Removes an item from an existing order

### 4.2 Service Dependencies

The Order Service depends on the following services:

- **Inventory Service**: For checking product availability and reserving inventory
- **Product Service**: For retrieving product details and validation
- **User Service**: For user authentication and address validation
- **Payment Service**: For processing payments and tracking payment status

## 5. Business Logic

### 5.1 Order Creation Flow

1. **Validation**
   - Validate user and address information
   - Validate product availability
   - Validate payment information

2. **Inventory Check**
   - Check inventory availability for all items
   - Return error if any items are unavailable

3. **Order Creation**
   - Create order record in pending state
   - Create order items
   - Create order addresses

4. **Inventory Reservation**
   - Reserve inventory for all items
   - Store reservation ID with order

5. **Payment Processing**
   - Process payment through payment service
   - Store payment transaction details

6. **Order Confirmation**
   - Update order status to confirmed
   - Commit inventory reservation

### 5.2 Order Fulfillment Flow

1. **Status Update**
   - Update order status to processing
   - Record status change in history

2. **Shipping Preparation**
   - Generate shipping labels
   - Update order with tracking information

3. **Fulfillment**
   - Update order status to shipped
   - Record status change in history

4. **Completion**
   - Update order status to delivered
   - Record status change in history

### 5.3 Order Cancellation Flow

1. **Validation**
   - Check if order can be cancelled
   - Validate cancellation reason

2. **Payment Reversal**
   - Initiate refund if payment was processed
   - Record refund transaction

3. **Inventory Release**
   - Release reserved inventory
   - Record inventory transaction

4. **Status Update**
   - Update order status to cancelled
   - Record status change in history

### 5.4 Saga Pattern Implementation

The Order Service implements the Saga pattern to manage distributed transactions across multiple services:

1. **Local Transaction**: Create order record
2. **Compensating Transaction**: Delete order if subsequent steps fail
3. **Local Transaction**: Reserve inventory
4. **Compensating Transaction**: Release inventory if subsequent steps fail
5. **Local Transaction**: Process payment
6. **Compensating Transaction**: Refund payment if subsequent steps fail
7. **Local Transaction**: Confirm order

## 6. Database Schema

```sql
-- Orders table
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    order_number VARCHAR(50) UNIQUE NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    total_amount DECIMAL(10,2) NOT NULL,
    subtotal DECIMAL(10,2) NOT NULL,
    tax_amount DECIMAL(10,2) NOT NULL,
    shipping_amount DECIMAL(10,2) NOT NULL,
    discount_amount DECIMAL(10,2) DEFAULT 0,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    payment_method VARCHAR(50),
    payment_status VARCHAR(50) DEFAULT 'PENDING',
    shipping_method VARCHAR(50),
    notes TEXT,
    inventory_reservation_id VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ
);

-- Order items table
CREATE TABLE order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL,
    product_id UUID NOT NULL,
    variant_id UUID,
    sku VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    quantity INT NOT NULL,
    unit_price DECIMAL(10,2) NOT NULL,
    subtotal DECIMAL(10,2) NOT NULL,
    discount_amount DECIMAL(10,2) DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT fk_order FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE
);

-- Order addresses table
CREATE TABLE order_addresses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL,
    address_type VARCHAR(20) NOT NULL, -- 'SHIPPING' or 'BILLING'
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    address_line1 VARCHAR(255) NOT NULL,
    address_line2 VARCHAR(255),
    city VARCHAR(100) NOT NULL,
    state VARCHAR(100) NOT NULL,
    postal_code VARCHAR(20) NOT NULL,
    country VARCHAR(100) NOT NULL,
    phone VARCHAR(20),
    email VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT fk_order_address FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
    CONSTRAINT order_address_type_unique UNIQUE (order_id, address_type)
);

-- Order status history table
CREATE TABLE order_status_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL,
    notes TEXT,
    created_by UUID,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT fk_order_status FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE
);

-- Order payment transactions table
CREATE TABLE order_payment_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL,
    transaction_id VARCHAR(255) NOT NULL,
    payment_method VARCHAR(50) NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    status VARCHAR(50) NOT NULL,
    gateway_response TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT fk_order_payment FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE
);
```

## 7. API Contracts

### 7.1 CreateOrder

**Request:**
```json
{
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "items": [
    {
      "product_id": "123e4567-e89b-12d3-a456-426614174001",
      "variant_id": "123e4567-e89b-12d3-a456-426614174002",
      "quantity": 2
    }
  ],
  "shipping_address": {
    "address_type": "SHIPPING",
    "first_name": "John",
    "last_name": "Doe",
    "address_line1": "123 Main St",
    "address_line2": "Apt 4B",
    "city": "New York",
    "state": "NY",
    "postal_code": "10001",
    "country": "USA",
    "phone": "555-123-4567",
    "email": "john.doe@example.com"
  },
  "billing_address": {
    "address_type": "BILLING",
    "first_name": "John",
    "last_name": "Doe",
    "address_line1": "123 Main St",
    "address_line2": "Apt 4B",
    "city": "New York",
    "state": "NY",
    "postal_code": "10001",
    "country": "USA",
    "phone": "555-123-4567",
    "email": "john.doe@example.com"
  },
  "payment_method": "CREDIT_CARD",
  "shipping_method": "STANDARD",
  "notes": "Please leave package at the door",
  "coupon_code": "SUMMER10"
}
```

**Response:**
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174003",
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "order_number": "ORD-12345678",
  "status": "CONFIRMED",
  "total_amount": 99.99,
  "subtotal": 89.99,
  "tax_amount": 5.00,
  "shipping_amount": 5.00,
  "discount_amount": 0.00,
  "currency": "USD",
  "payment_method": "CREDIT_CARD",
  "payment_status": "PAID",
  "shipping_method": "STANDARD",
  "notes": "Please leave package at the door",
  "created_at": "2024-05-03T12:00:00Z",
  "updated_at": "2024-05-03T12:00:00Z",
  "items": [
    {
      "id": "123e4567-e89b-12d3-a456-426614174004",
      "order_id": "123e4567-e89b-12d3-a456-426614174003",
      "product_id": "123e4567-e89b-12d3-a456-426614174001",
      "variant_id": "123e4567-e89b-12d3-a456-426614174002",
      "sku": "PROD-001-VAR-002",
      "name": "Product Name - Variant Name",
      "quantity": 2,
      "unit_price": 44.99,
      "subtotal": 89.98,
      "discount_amount": 0.00
    }
  ],
  "addresses": [
    {
      "id": "123e4567-e89b-12d3-a456-426614174005",
      "order_id": "123e4567-e89b-12d3-a456-426614174003",
      "address_type": "SHIPPING",
      "first_name": "John",
      "last_name": "Doe",
      "address_line1": "123 Main St",
      "address_line2": "Apt 4B",
      "city": "New York",
      "state": "NY",
      "postal_code": "10001",
      "country": "USA",
      "phone": "555-123-4567",
      "email": "john.doe@example.com"
    },
    {
      "id": "123e4567-e89b-12d3-a456-426614174006",
      "order_id": "123e4567-e89b-12d3-a456-426614174003",
      "address_type": "BILLING",
      "first_name": "John",
      "last_name": "Doe",
      "address_line1": "123 Main St",
      "address_line2": "Apt 4B",
      "city": "New York",
      "state": "NY",
      "postal_code": "10001",
      "country": "USA",
      "phone": "555-123-4567",
      "email": "john.doe@example.com"
    }
  ]
}
```

### 7.2 GetOrder

**Request:**
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174003"
}
```

**Response:**
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174003",
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "order_number": "ORD-12345678",
  "status": "CONFIRMED",
  "total_amount": 99.99,
  "subtotal": 89.99,
  "tax_amount": 5.00,
  "shipping_amount": 5.00,
  "discount_amount": 0.00,
  "currency": "USD",
  "payment_method": "CREDIT_CARD",
  "payment_status": "PAID",
  "shipping_method": "STANDARD",
  "notes": "Please leave package at the door",
  "created_at": "2024-05-03T12:00:00Z",
  "updated_at": "2024-05-03T12:00:00Z",
  "items": [
    {
      "id": "123e4567-e89b-12d3-a456-426614174004",
      "order_id": "123e4567-e89b-12d3-a456-426614174003",
      "product_id": "123e4567-e89b-12d3-a456-426614174001",
      "variant_id": "123e4567-e89b-12d3-a456-426614174002",
      "sku": "PROD-001-VAR-002",
      "name": "Product Name - Variant Name",
      "quantity": 2,
      "unit_price": 44.99,
      "subtotal": 89.98,
      "discount_amount": 0.00
    }
  ],
  "addresses": [
    {
      "id": "123e4567-e89b-12d3-a456-426614174005",
      "order_id": "123e4567-e89b-12d3-a456-426614174003",
      "address_type": "SHIPPING",
      "first_name": "John",
      "last_name": "Doe",
      "address_line1": "123 Main St",
      "address_line2": "Apt 4B",
      "city": "New York",
      "state": "NY",
      "postal_code": "10001",
      "country": "USA",
      "phone": "555-123-4567",
      "email": "john.doe@example.com"
    },
    {
      "id": "123e4567-e89b-12d3-a456-426614174006",
      "order_id": "123e4567-e89b-12d3-a456-426614174003",
      "address_type": "BILLING",
      "first_name": "John",
      "last_name": "Doe",
      "address_line1": "123 Main St",
      "address_line2": "Apt 4B",
      "city": "New York",
      "state": "NY",
      "postal_code": "10001",
      "country": "USA",
      "phone": "555-123-4567",
      "email": "john.doe@example.com"
    }
  ]
}
```

## 8. Error Handling

### 8.1 Error Types

The Order Service defines the following error types:

- **ValidationError**: Invalid input data
- **NotFoundError**: Requested resource not found
- **InventoryError**: Inventory-related errors
- **PaymentError**: Payment-related errors
- **DatabaseError**: Database operation errors
- **ServiceError**: Errors from dependent services

### 8.2 Error Responses

Error responses follow a standard format:

```json
{
  "error": {
    "code": "INVENTORY_UNAVAILABLE",
    "message": "Some items are not available",
    "details": [
      {
        "product_id": "123e4567-e89b-12d3-a456-426614174001",
        "requested_quantity": 5,
        "available_quantity": 2
      }
    ]
  }
}
```

## 9. Caching Strategy

The Order Service implements caching for frequently accessed data:

- **Order Details**: Cache order details for quick retrieval
- **Order Lists**: Cache paginated order lists
- **Order Status**: Cache order status for quick checks

Cache invalidation occurs on:
- Order creation
- Order status updates
- Order cancellation

## 10. Monitoring and Logging

### 10.1 Metrics

The Order Service collects the following metrics:

- **Order Creation Rate**: Orders created per minute
- **Order Fulfillment Rate**: Orders fulfilled per minute
- **Order Cancellation Rate**: Orders cancelled per minute
- **Average Order Value**: Average value of orders
- **Service Latency**: Response time for service operations

### 10.2 Logging

The Order Service implements structured logging with the following levels:

- **DEBUG**: Detailed debugging information
- **INFO**: General operational information
- **WARN**: Warning events that might cause issues
- **ERROR**: Error events that might still allow the service to continue
- **FATAL**: Severe error events that will lead to service termination

### 10.3 Tracing

The Order Service implements distributed tracing to track requests across services:

- **Request ID**: Unique identifier for each request
- **Trace ID**: Identifier for tracing a request across services
- **Span ID**: Identifier for specific operations within a trace

## 11. Scaling Considerations

### 11.1 Horizontal Scaling

The Order Service is designed for horizontal scaling:

- Stateless service design
- Database connection pooling
- Caching for read-heavy operations

### 11.2 Database Scaling

Database scaling strategies include:

- Read replicas for read-heavy operations
- Sharding for write-heavy operations
- Indexing for query optimization

### 11.3 Caching Scaling

Caching scaling strategies include:

- Redis cluster for distributed caching
- Cache partitioning for improved performance
- Cache warming for frequently accessed data

## 12. Security Considerations

### 12.1 Authentication and Authorization

The Order Service implements:

- JWT-based authentication
- Role-based access control
- User ownership validation

### 12.2 Data Protection

Data protection measures include:

- Encryption of sensitive data
- PCI compliance for payment information
- Data anonymization for analytics

### 12.3 Input Validation

Input validation includes:

- Request validation at API boundaries
- Parameter sanitization
- Rate limiting for API endpoints

## 13. Future Enhancements

### 13.1 Event-Driven Architecture

Future enhancements include:

- Implementing event sourcing for order events
- Publishing order events to a message broker
- Subscribing to events from other services

### 13.2 Advanced Analytics

Advanced analytics features include:

- Order trend analysis
- Customer purchase patterns
- Inventory demand forecasting

### 13.3 Machine Learning Integration

Machine learning integration includes:

- Fraud detection for orders
- Personalized recommendations
- Dynamic pricing strategies
