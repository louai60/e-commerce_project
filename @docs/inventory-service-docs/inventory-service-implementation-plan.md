# Inventory Service Implementation Plan

This document outlines the step-by-step plan for implementing the inventory service in our e-commerce microservices architecture.

## Overview

The inventory service will be responsible for:
- Managing product inventory across multiple warehouses
- Tracking inventory levels and status
- Handling inventory reservations during checkout
- Managing inventory adjustments (restocking, returns, etc.)
- Providing real-time inventory availability information
- Supporting inventory allocation strategies

## Implementation Steps

### Phase 1: Core Service Setup

- [x] **Create Basic Service Structure**
  - [x] Set up directory structure
  - [x] Create main.go file
  - [x] Set up configuration files
  - [x] Create Dockerfile

- [x] **Define Protocol Buffers**
  - [x] Define service interfaces
  - [x] Define message types
  - [ ] Generate Go code from proto files

- [x] **Database Setup**
  - [x] Create database migration files
  - [x] Implement database connection
  - [x] Set up migration mechanism

- [x] **Implement Core Models**
  - [x] Define inventory item model
  - [x] Define warehouse model
  - [x] Define inventory location model
  - [x] Define transaction model
  - [x] Define reservation model

### Phase 2: Repository Layer

- [x] **Implement Repository Interfaces**
  - [x] Define repository interfaces
  - [x] Implement PostgreSQL repository for inventory items
  - [x] Implement PostgreSQL repository for warehouses
  - [x] Implement PostgreSQL repository for inventory locations
  - [x] Implement PostgreSQL repository for transactions
  - [x] Implement PostgreSQL repository for reservations

- [ ] **Implement Caching Layer**
  - [ ] Set up Redis connection
  - [ ] Implement cache manager
  - [ ] Add caching for frequently accessed inventory data

### Phase 3: Service Layer

- [x] **Implement Core Business Logic**
  - [x] Implement inventory service
  - [x] Implement warehouse service
  - [x] Implement inventory transaction logic
  - [x] Implement reservation system

- [x] **Implement Inventory Management Features**
  - [x] Add/remove inventory
  - [x] Transfer inventory between warehouses
  - [x] Adjust inventory levels
  - [x] Track inventory history

- [x] **Implement Reservation System**
  - [x] Create reservations
  - [x] Confirm reservations
  - [x] Cancel reservations
  - [x] Handle reservation expiration

### Phase 4: API Layer

- [x] **Implement gRPC Handlers**
  - [x] Implement inventory item handlers
  - [x] Implement warehouse handlers
  - [x] Implement inventory location handlers
  - [x] Implement reservation handlers

- [x] **Implement Middleware**
  - [x] Add logging middleware
  - [ ] Add authentication middleware
  - [ ] Add validation middleware

### Phase 5: Integration

- [x] **Integrate with Product Service**
  - [x] Update product service to call inventory service when creating/updating products
  - [x] Ensure product service can retrieve inventory information

- [ ] **Integrate with Order Service**
  - [ ] Implement inventory reservation during checkout
  - [ ] Implement inventory confirmation on order placement
  - [ ] Implement inventory release on order cancellation

- [x] **Integrate with API Gateway**
  - [x] Update API gateway to route inventory-related requests
  - [x] Add inventory endpoints to API documentation

#### Integration Details

1. **Product Service Integration**
   - Added inventory client in product service to communicate with inventory service
   - Updated product creation flow to create inventory items automatically
   - Added inventory availability checking during product operations

2. **API Gateway Integration**
   - Added inventory client and handler in API gateway
   - Created REST endpoints for inventory operations:
     - `/api/v1/inventory/check` - Check inventory availability (public)
     - `/api/v1/inventory/items` - List inventory items (admin only)
     - `/api/v1/inventory/items/:product_id` - Get inventory for a product (admin only)
     - `/api/v1/inventory/warehouses` - List warehouses (admin only)

### Phase 6: Testing and Deployment

- [ ] **Write Unit Tests**
  - [ ] Test repository methods
  - [ ] Test service methods
  - [ ] Test gRPC handlers

- [ ] **Write Integration Tests**
  - [ ] Test service-to-service communication
  - [ ] Test database operations
  - [ ] Test caching mechanisms

- [ ] **Deployment**
  - [ ] Update docker-compose.yml
  - [ ] Configure CI/CD pipeline
  - [ ] Deploy to development environment

## Data Model

### Inventory Items
The core entity tracking inventory for products and variants:
```
inventory_items
- id (UUID)
- product_id (UUID)
- variant_id (UUID, nullable)
- sku (string)
- total_quantity (int)
- available_quantity (int)
- reserved_quantity (int)
- reorder_point (int)
- reorder_quantity (int)
- status (string: IN_STOCK, LOW_STOCK, OUT_OF_STOCK, DISCONTINUED)
- last_updated (timestamp)
- created_at (timestamp)
- updated_at (timestamp)
```

### Warehouses
Physical locations where inventory is stored:
```
warehouses
- id (UUID)
- name (string)
- code (string)
- address (text)
- city (string)
- state (string)
- country (string)
- postal_code (string)
- is_active (boolean)
- priority (int)
- created_at (timestamp)
- updated_at (timestamp)
```

### Inventory Locations
Distribution of inventory across warehouses:
```
inventory_locations
- id (UUID)
- inventory_item_id (UUID)
- warehouse_id (UUID)
- quantity (int)
- available_quantity (int)
- reserved_quantity (int)
- created_at (timestamp)
- updated_at (timestamp)
```

### Inventory Transactions
Audit trail of inventory changes:
```
inventory_transactions
- id (UUID)
- inventory_item_id (UUID)
- warehouse_id (UUID, nullable)
- transaction_type (string: STOCK_ADDITION, STOCK_REMOVAL, RESERVATION, RESERVATION_RELEASE, ADJUSTMENT)
- quantity (int)
- reference_id (UUID, nullable)
- reference_type (string, nullable)
- notes (text, nullable)
- created_by (UUID, nullable)
- created_at (timestamp)
```

### Inventory Reservations
Temporary holds on inventory during checkout:
```
inventory_reservations
- id (UUID)
- inventory_item_id (UUID)
- warehouse_id (UUID, nullable)
- quantity (int)
- status (string: PENDING, CONFIRMED, CANCELLED, FULFILLED, EXPIRED)
- expiration_time (timestamp)
- reference_id (UUID, nullable)
- reference_type (string)
- created_at (timestamp)
- updated_at (timestamp)
```

## Service Interactions

### Product Creation Flow
1. Product service creates a new product
2. Product service calls inventory service to create inventory records
3. Inventory service creates inventory items and allocates to warehouses
4. Inventory service returns success/failure to product service

### Checkout Flow
1. Order service calls inventory service to reserve inventory
2. Inventory service creates reservations with expiration time
3. On payment confirmation, order service calls inventory service to confirm reservations
4. Inventory service converts reservations to fulfilled and updates inventory levels

### Inventory Management Flow
1. Admin updates inventory through admin dashboard
2. API gateway routes request to inventory service
3. Inventory service updates inventory levels and creates transaction records
4. Inventory service publishes inventory update events for other services

## Performance Considerations

- Implement caching for frequently accessed inventory data
- Use database indexes for efficient queries
- Implement batch processing for bulk inventory updates
- Consider sharding for high-volume inventory data
- Implement background job for cleaning expired reservations
