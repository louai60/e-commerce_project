# Order Service Integration Plan

## Overview

This document outlines the comprehensive plan for implementing the order service and integrating it with all necessary services in our e-commerce microservices architecture. The order service is a critical component that will manage the entire order lifecycle, from creation to fulfillment, and will interact with multiple other services.

## Architecture

The order service will follow our standard microservice architecture pattern:

```
Client Request → API Gateway → Order Service → Database
```

Internal service-to-service communication will use gRPC, while external communication via the API Gateway will use REST.

## Service Dependencies

The order service will integrate with the following services:

1. **Inventory Service**: For checking product availability and reserving inventory
2. **Product Service**: For retrieving product details and validation
3. **User Service**: For user authentication and address validation
4. **Payment Service**: For processing payments and tracking payment status
5. **Notification Service** (future): For sending order confirmations and updates

## Implementation Plan

### Phase 1: Core Order Service Setup

- [ ] Create project structure following standard layout
  - [ ] Create main directories (handlers, models, repository, service, proto)
  - [ ] Set up configuration files and environment variables
  - [ ] Create Dockerfile and docker-compose configuration

- [ ] Design and implement database schema
  - [ ] Create migration files for orders table
  - [ ] Create migration files for order_items table
  - [ ] Create migration files for order_addresses table
  - [ ] Create migration files for order_status_history table
  - [ ] Create migration files for order_payment_transactions table

- [ ] Define proto files
  - [ ] Define order message types
  - [ ] Define order service interface
  - [ ] Generate Go code from proto files

### Phase 2: Core Service Implementation

- [ ] Implement models
  - [ ] Order model
  - [ ] OrderItem model
  - [ ] OrderAddress model
  - [ ] OrderStatusHistory model
  - [ ] OrderPaymentTransaction model

- [ ] Implement repositories
  - [ ] OrderRepository
  - [ ] OrderItemRepository
  - [ ] OrderAddressRepository
  - [ ] OrderStatusHistoryRepository
  - [ ] OrderPaymentTransactionRepository

- [ ] Implement service layer
  - [ ] Basic CRUD operations
  - [ ] Order status management
  - [ ] Order validation logic

- [ ] Implement gRPC handlers
  - [ ] CreateOrder handler
  - [ ] GetOrder handler
  - [ ] UpdateOrderStatus handler
  - [ ] CancelOrder handler
  - [ ] ListOrders handler

### Phase 3: Service Integrations

- [ ] Inventory Service Integration
  - [ ] Implement inventory client
  - [ ] Implement inventory availability checking
  - [ ] Implement inventory reservation
  - [ ] Implement inventory release on cancellation

- [ ] Product Service Integration
  - [ ] Implement product client
  - [ ] Implement product information retrieval
  - [ ] Implement product validation

- [ ] User Service Integration
  - [ ] Implement user client
  - [ ] Implement user authentication and authorization
  - [ ] Implement address validation

- [ ] Payment Service Integration
  - [ ] Implement payment client
  - [ ] Implement payment processing
  - [ ] Implement payment status tracking

### Phase 4: Order Workflow Implementation

- [ ] Order Creation Flow
  - [ ] Validate user and addresses
  - [ ] Check product availability
  - [ ] Reserve inventory
  - [ ] Create order in pending state
  - [ ] Process payment
  - [ ] Confirm order or handle failures

- [ ] Order Fulfillment Flow
  - [ ] Update order status
  - [ ] Trigger inventory updates
  - [ ] Handle shipping information

- [ ] Order Cancellation Flow
  - [ ] Cancel payment if needed
  - [ ] Release inventory reservations
  - [ ] Update order status

- [ ] Implement Saga Pattern
  - [ ] Implement distributed transaction management
  - [ ] Handle compensating transactions for failures
  - [ ] Ensure data consistency across services

### Phase 5: Testing and Optimization

- [ ] Unit Testing
  - [ ] Test individual components
  - [ ] Test service layer logic
  - [ ] Test repository operations

- [ ] Integration Testing
  - [ ] Test service integrations
  - [ ] Test order workflows
  - [ ] Test failure scenarios and recovery

- [ ] Performance Optimization
  - [ ] Implement caching for frequently accessed data
  - [ ] Optimize database queries
  - [ ] Implement connection pooling

## Technical Details

### Database Schema

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

### Proto Definition

```protobuf
syntax = "proto3";

package order;

import "google/protobuf/timestamp.proto";
import "google/protobuf/wrappers.proto";

option go_package = "github.com/louai60/e-commerce_project/backend/order-service/proto";

service OrderService {
  // Order management
  rpc CreateOrder(CreateOrderRequest) returns (Order);
  rpc GetOrder(GetOrderRequest) returns (Order);
  rpc UpdateOrderStatus(UpdateOrderStatusRequest) returns (Order);
  rpc CancelOrder(CancelOrderRequest) returns (CancelOrderResponse);
  rpc ListOrders(ListOrdersRequest) returns (ListOrdersResponse);
  
  // Order item management
  rpc AddOrderItem(AddOrderItemRequest) returns (Order);
  rpc UpdateOrderItem(UpdateOrderItemRequest) returns (Order);
  rpc RemoveOrderItem(RemoveOrderItemRequest) returns (Order);
}

// Order message
message Order {
  string id = 1;
  string user_id = 2;
  string order_number = 3;
  string status = 4;
  double total_amount = 5;
  double subtotal = 6;
  double tax_amount = 7;
  double shipping_amount = 8;
  double discount_amount = 9;
  string currency = 10;
  string payment_method = 11;
  string payment_status = 12;
  string shipping_method = 13;
  string notes = 14;
  google.protobuf.Timestamp created_at = 15;
  google.protobuf.Timestamp updated_at = 16;
  google.protobuf.Timestamp completed_at = 17;
  google.protobuf.Timestamp cancelled_at = 18;
  repeated OrderItem items = 19;
  repeated OrderAddress addresses = 20;
}

// Order item message
message OrderItem {
  string id = 1;
  string order_id = 2;
  string product_id = 3;
  google.protobuf.StringValue variant_id = 4;
  string sku = 5;
  string name = 6;
  int32 quantity = 7;
  double unit_price = 8;
  double subtotal = 9;
  double discount_amount = 10;
  google.protobuf.Timestamp created_at = 11;
  google.protobuf.Timestamp updated_at = 12;
}

// Order address message
message OrderAddress {
  string id = 1;
  string order_id = 2;
  string address_type = 3; // 'SHIPPING' or 'BILLING'
  string first_name = 4;
  string last_name = 5;
  string address_line1 = 6;
  google.protobuf.StringValue address_line2 = 7;
  string city = 8;
  string state = 9;
  string postal_code = 10;
  string country = 11;
  google.protobuf.StringValue phone = 12;
  google.protobuf.StringValue email = 13;
}

// Request and response messages
message CreateOrderRequest {
  string user_id = 1;
  repeated OrderItemInput items = 2;
  OrderAddressInput shipping_address = 3;
  OrderAddressInput billing_address = 4;
  string payment_method = 5;
  string shipping_method = 6;
  google.protobuf.StringValue notes = 7;
  google.protobuf.StringValue coupon_code = 8;
}

message OrderItemInput {
  string product_id = 1;
  google.protobuf.StringValue variant_id = 2;
  int32 quantity = 3;
}

message OrderAddressInput {
  string address_type = 1;
  string first_name = 2;
  string last_name = 3;
  string address_line1 = 4;
  google.protobuf.StringValue address_line2 = 5;
  string city = 6;
  string state = 7;
  string postal_code = 8;
  string country = 9;
  google.protobuf.StringValue phone = 10;
  google.protobuf.StringValue email = 11;
}

message GetOrderRequest {
  oneof identifier {
    string id = 1;
    string order_number = 2;
  }
}

message UpdateOrderStatusRequest {
  string order_id = 1;
  string status = 2;
  google.protobuf.StringValue notes = 3;
}

message CancelOrderRequest {
  string order_id = 1;
  string reason = 2;
}

message CancelOrderResponse {
  bool success = 1;
  string message = 2;
}

message ListOrdersRequest {
  string user_id = 1;
  google.protobuf.StringValue status = 2;
  int32 page = 3;
  int32 limit = 4;
  string sort_by = 5;
  bool sort_desc = 6;
}

message ListOrdersResponse {
  repeated Order orders = 1;
  int32 total = 2;
  int32 page = 3;
  int32 limit = 4;
}

message AddOrderItemRequest {
  string order_id = 1;
  OrderItemInput item = 2;
}

message UpdateOrderItemRequest {
  string order_id = 1;
  string item_id = 2;
  int32 quantity = 3;
}

message RemoveOrderItemRequest {
  string order_id = 1;
  string item_id = 2;
}
```

## Order Service Integration with Inventory Service

The order service will integrate with the inventory service for:

1. **Checking inventory availability** before creating an order
2. **Reserving inventory** when an order is created
3. **Releasing inventory** if an order is cancelled
4. **Confirming inventory deduction** when an order is fulfilled

Example integration code:

```go
// OrderService struct with inventory client
type OrderService struct {
    orderRepo       repository.OrderRepository
    inventoryClient pb.InventoryServiceClient
    productClient   pb.ProductServiceClient
    paymentClient   pb.PaymentServiceClient
    userClient      pb.UserServiceClient
    logger          *zap.Logger
}

// CreateOrder method with inventory integration
func (s *OrderService) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*models.Order, error) {
    // 1. Validate user and addresses
    userResp, err := s.userClient.GetUser(ctx, &userpb.GetUserRequest{Id: req.UserId})
    if err != nil {
        return nil, fmt.Errorf("failed to validate user: %w", err)
    }
    
    // 2. Check inventory availability for all items
    availabilityItems := make([]*inventorypb.AvailabilityCheckItem, 0, len(req.Items))
    for _, item := range req.Items {
        // Get product details
        productResp, err := s.productClient.GetProduct(ctx, &productpb.GetProductRequest{Id: item.ProductId})
        if err != nil {
            return nil, fmt.Errorf("failed to get product details: %w", err)
        }
        
        // Add to availability check
        availabilityItems = append(availabilityItems, &inventorypb.AvailabilityCheckItem{
            ProductId: item.ProductId,
            VariantId: item.VariantId,
            Sku:       productResp.Product.Sku,
            Quantity:  item.Quantity,
        })
    }
    
    // Check inventory availability
    availabilityResp, err := s.inventoryClient.CheckInventoryAvailability(ctx, &inventorypb.CheckInventoryAvailabilityRequest{
        Items: availabilityItems,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to check inventory availability: %w", err)
    }
    
    // If not all items are available, return error with details
    if !availabilityResp.AllAvailable {
        // Create detailed error message
        var unavailableItems []string
        for _, item := range availabilityResp.Items {
            if !item.IsAvailable {
                unavailableItems = append(unavailableItems, fmt.Sprintf("Product %s: requested %d, available %d", 
                    item.ProductId, item.RequestedQuantity, item.AvailableQuantity))
            }
        }
        return nil, fmt.Errorf("some items are not available: %s", strings.Join(unavailableItems, "; "))
    }
    
    // 3. Create order in database
    order, err := s.createOrderInDB(ctx, req, userResp)
    if err != nil {
        return nil, fmt.Errorf("failed to create order: %w", err)
    }
    
    // 4. Reserve inventory
    reservationItems := make([]*inventorypb.ReservationItem, 0, len(req.Items))
    for _, item := range req.Items {
        // Get product details again (or store from earlier)
        productResp, err := s.productClient.GetProduct(ctx, &productpb.GetProductRequest{Id: item.ProductId})
        if err != nil {
            // Rollback order creation
            s.orderRepo.DeleteOrder(ctx, order.ID)
            return nil, fmt.Errorf("failed to get product details for reservation: %w", err)
        }
        
        reservationItems = append(reservationItems, &inventorypb.ReservationItem{
            InventoryItemId: productResp.Product.Sku, // Using SKU to identify inventory item
            Quantity:        item.Quantity,
        })
    }
    
    // Make reservation
    reservationResp, err := s.inventoryClient.ReserveInventory(ctx, &inventorypb.ReserveInventoryRequest{
        Items:           reservationItems,
        ReferenceId:     order.ID,
        ReferenceType:   "ORDER",
        ExpirationMinutes: 30, // Configure as needed
    })
    if err != nil {
        // Rollback order creation
        s.orderRepo.DeleteOrder(ctx, order.ID)
        return nil, fmt.Errorf("failed to reserve inventory: %w", err)
    }
    
    // Store reservation ID with order
    order.InventoryReservationID = reservationResp.ReservationId
    if err := s.orderRepo.UpdateOrder(ctx, order); err != nil {
        // Try to cancel reservation
        s.inventoryClient.CancelReservation(ctx, &inventorypb.CancelReservationRequest{
            ReservationId: reservationResp.ReservationId,
        })
        return nil, fmt.Errorf("failed to update order with reservation ID: %w", err)
    }
    
    // 5. Process payment (simplified)
    paymentResp, err := s.paymentClient.ProcessPayment(ctx, &paymentpb.ProcessPaymentRequest{
        OrderId:     order.ID,
        Amount:      order.TotalAmount,
        Currency:    order.Currency,
        Method:      req.PaymentMethod,
    })
    if err != nil {
        // Cancel reservation and mark order as failed
        s.inventoryClient.CancelReservation(ctx, &inventorypb.CancelReservationRequest{
            ReservationId: reservationResp.ReservationId,
        })
        s.orderRepo.UpdateOrderStatus(ctx, order.ID, "PAYMENT_FAILED", fmt.Sprintf("Payment failed: %v", err))
        return nil, fmt.Errorf("payment processing failed: %w", err)
    }
    
    // 6. Confirm order
    order.PaymentStatus = paymentResp.Status
    order.Status = "CONFIRMED"
    if err := s.orderRepo.UpdateOrder(ctx, order); err != nil {
        s.logger.Error("Failed to update order status after payment",
            zap.Error(err),
            zap.String("order_id", order.ID))
        // Continue despite error, as payment was successful
    }
    
    // 7. Commit inventory (convert reservation to actual deduction)
    _, err = s.inventoryClient.CommitReservation(ctx, &inventorypb.CommitReservationRequest{
        ReservationId: reservationResp.ReservationId,
    })
    if err != nil {
        s.logger.Error("Failed to commit inventory reservation, manual intervention required",
            zap.Error(err),
            zap.String("order_id", order.ID),
            zap.String("reservation_id", reservationResp.ReservationId))
        // Continue despite error, as order and payment were successful
        // This will require manual inventory adjustment
    }
    
    return order, nil
}
```

## Implementation Timeline

1. **Week 1: Setup and Database**
   - Set up project structure
   - Implement database schema and migrations
   - Define proto files

2. **Week 2: Core Implementation**
   - Implement models and repositories
   - Implement basic service layer
   - Implement gRPC handlers

3. **Week 3: Service Integrations**
   - Implement inventory service integration
   - Implement product service integration
   - Implement user service integration

4. **Week 4: Order Workflows**
   - Implement order creation workflow
   - Implement order fulfillment workflow
   - Implement order cancellation workflow

5. **Week 5: Testing and Refinement**
   - Write unit and integration tests
   - Optimize performance
   - Fix bugs and refine implementation

## Implementation Tracker

| Phase | Task | Status | Completion Date | Notes |
|-------|------|--------|----------------|-------|
| **Phase 1** | Create project structure | Not Started | | |
| | Design database schema | Not Started | | |
| | Define proto files | Not Started | | |
| **Phase 2** | Implement models | Not Started | | |
| | Implement repositories | Not Started | | |
| | Implement service layer | Not Started | | |
| | Implement gRPC handlers | Not Started | | |
| **Phase 3** | Inventory Service Integration | Not Started | | |
| | Product Service Integration | Not Started | | |
| | User Service Integration | Not Started | | |
| | Payment Service Integration | Not Started | | |
| **Phase 4** | Order Creation Flow | Not Started | | |
| | Order Fulfillment Flow | Not Started | | |
| | Order Cancellation Flow | Not Started | | |
| | Implement Saga Pattern | Not Started | | |
| **Phase 5** | Unit Testing | Not Started | | |
| | Integration Testing | Not Started | | |
| | Performance Optimization | Not Started | | |

## Conclusion

The order service is a critical component of our e-commerce platform, requiring careful integration with multiple services. This implementation plan provides a structured approach to building a robust, scalable order management system that maintains data consistency across services while handling the complex workflows involved in order processing.
