# Order Service and Inventory Service Integration

## Overview

This document details the integration between the Order Service and Inventory Service in our e-commerce microservices architecture. This integration is critical for maintaining accurate inventory levels, preventing overselling, and ensuring a smooth customer experience.

## Integration Points

The Order Service interacts with the Inventory Service at several key points in the order lifecycle:

1. **Inventory Availability Check**: Before creating an order
2. **Inventory Reservation**: When an order is created
3. **Inventory Release**: If an order is cancelled
4. **Inventory Deduction**: When an order is fulfilled

## gRPC Interface

The integration uses the following gRPC methods exposed by the Inventory Service:

### 1. CheckInventoryAvailability

Checks if requested quantities of products are available in inventory.

**Request:**
```protobuf
message CheckInventoryAvailabilityRequest {
  repeated AvailabilityCheckItem items = 1;
}

message AvailabilityCheckItem {
  string product_id = 1;
  google.protobuf.StringValue variant_id = 2;
  string sku = 3;
  int32 quantity = 4;
}
```

**Response:**
```protobuf
message InventoryAvailabilityResponse {
  repeated ItemAvailability items = 1;
  bool all_available = 2;
}

message ItemAvailability {
  string product_id = 1;
  google.protobuf.StringValue variant_id = 2;
  string sku = 3;
  int32 requested_quantity = 4;
  int32 available_quantity = 5;
  bool is_available = 6;
  string status = 7;
}
```

### 2. ReserveInventory

Creates temporary holds on inventory items during the checkout process.

**Request:**
```protobuf
message ReserveInventoryRequest {
  repeated ReservationItem items = 1;
  string reference_id = 2;
  string reference_type = 3;
  int32 expiration_minutes = 4;
}

message ReservationItem {
  string inventory_item_id = 1;
  string warehouse_id = 2;
  int32 quantity = 3;
}
```

**Response:**
```protobuf
message ReservationResponse {
  string reservation_id = 1;
  bool success = 2;
  string message = 3;
}
```

### 3. CommitReservation

Converts a temporary reservation into a permanent inventory deduction.

**Request:**
```protobuf
message CommitReservationRequest {
  string reservation_id = 1;
}
```

**Response:**
```protobuf
message ReservationResponse {
  string reservation_id = 1;
  bool success = 2;
  string message = 3;
}
```

### 4. CancelReservation

Releases a temporary inventory reservation.

**Request:**
```protobuf
message CancelReservationRequest {
  string reservation_id = 1;
}
```

**Response:**
```protobuf
message ReservationResponse {
  string reservation_id = 1;
  bool success = 2;
  string message = 3;
}
```

## Integration Flow

### Order Creation Flow

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   Order Service │     │ Product Service  │     │Inventory Service│
└────────┬────────┘     └────────┬────────┘     └────────┬────────┘
         │                       │                       │
         │  Get Product Details  │                       │
         │───────────────────────>                       │
         │                       │                       │
         │  Product Details      │                       │
         │<───────────────────────                       │
         │                       │                       │
         │                       │  Check Availability   │
         │───────────────────────────────────────────────>
         │                       │                       │
         │                       │  Availability Response│
         │<───────────────────────────────────────────────
         │                       │                       │
         │ Create Order (DB)     │                       │
         │─────┐                 │                       │
         │     │                 │                       │
         │<────┘                 │                       │
         │                       │                       │
         │                       │  Reserve Inventory    │
         │───────────────────────────────────────────────>
         │                       │                       │
         │                       │  Reservation Response │
         │<───────────────────────────────────────────────
         │                       │                       │
         │ Process Payment       │                       │
         │─────┐                 │                       │
         │     │                 │                       │
         │<────┘                 │                       │
         │                       │                       │
         │                       │  Commit Reservation   │
         │───────────────────────────────────────────────>
         │                       │                       │
         │                       │  Commit Response      │
         │<───────────────────────────────────────────────
         │                       │                       │
```

### Order Cancellation Flow

```
┌─────────────────┐     ┌─────────────────┐
│   Order Service │     │Inventory Service│
└────────┬────────┘     └────────┬────────┘
         │                       │
         │ Update Order Status   │
         │─────┐                 │
         │     │                 │
         │<────┘                 │
         │                       │
         │  Cancel Reservation   │
         │───────────────────────>
         │                       │
         │  Cancellation Response│
         │<───────────────────────
         │                       │
```

## Implementation Details

### 1. Inventory Client in Order Service

The Order Service will include an Inventory Service client:

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

// NewOrderService creates a new order service
func NewOrderService(
    orderRepo repository.OrderRepository,
    inventoryClient pb.InventoryServiceClient,
    productClient pb.ProductServiceClient,
    paymentClient pb.PaymentServiceClient,
    userClient pb.UserServiceClient,
    logger *zap.Logger,
) *OrderService {
    return &OrderService{
        orderRepo:       orderRepo,
        inventoryClient: inventoryClient,
        productClient:   productClient,
        paymentClient:   paymentClient,
        userClient:      userClient,
        logger:          logger,
    }
}
```

### 2. Inventory Availability Check

Before creating an order, the Order Service checks if all requested items are available:

```go
// Check inventory availability
func (s *OrderService) checkInventoryAvailability(ctx context.Context, items []*pb.OrderItemInput) error {
    // Convert order items to availability check items
    availabilityItems := make([]*inventorypb.AvailabilityCheckItem, 0, len(items))
    for _, item := range items {
        // Get product details to get SKU
        productResp, err := s.productClient.GetProduct(ctx, &productpb.GetProductRequest{
            Identifier: &productpb.GetProductRequest_Id{Id: item.ProductId},
        })
        if err != nil {
            return fmt.Errorf("failed to get product details: %w", err)
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
        return fmt.Errorf("failed to check inventory availability: %w", err)
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
        return fmt.Errorf("some items are not available: %s", strings.Join(unavailableItems, "; "))
    }
    
    return nil
}
```

### 3. Inventory Reservation

When an order is created, the Order Service reserves the inventory:

```go
// Reserve inventory for order
func (s *OrderService) reserveInventory(ctx context.Context, order *models.Order, items []*models.OrderItem) (string, error) {
    // Convert order items to reservation items
    reservationItems := make([]*inventorypb.ReservationItem, 0, len(items))
    for _, item := range items {
        reservationItems = append(reservationItems, &inventorypb.ReservationItem{
            InventoryItemId: item.SKU, // Using SKU to identify inventory item
            Quantity:        int32(item.Quantity),
        })
    }
    
    // Make reservation
    reservationResp, err := s.inventoryClient.ReserveInventory(ctx, &inventorypb.ReserveInventoryRequest{
        Items:             reservationItems,
        ReferenceId:       order.ID,
        ReferenceType:     "ORDER",
        ExpirationMinutes: 30, // Configure as needed
    })
    if err != nil {
        return "", fmt.Errorf("failed to reserve inventory: %w", err)
    }
    
    if !reservationResp.Success {
        return "", fmt.Errorf("inventory reservation failed: %s", reservationResp.Message)
    }
    
    return reservationResp.ReservationId, nil
}
```

### 4. Commit Reservation

After payment is processed, the Order Service commits the inventory reservation:

```go
// Commit inventory reservation
func (s *OrderService) commitInventoryReservation(ctx context.Context, reservationID string) error {
    resp, err := s.inventoryClient.CommitReservation(ctx, &inventorypb.CommitReservationRequest{
        ReservationId: reservationID,
    })
    if err != nil {
        return fmt.Errorf("failed to commit inventory reservation: %w", err)
    }
    
    if !resp.Success {
        return fmt.Errorf("inventory commit failed: %s", resp.Message)
    }
    
    return nil
}
```

### 5. Cancel Reservation

If an order is cancelled, the Order Service releases the inventory reservation:

```go
// Cancel inventory reservation
func (s *OrderService) cancelInventoryReservation(ctx context.Context, reservationID string) error {
    resp, err := s.inventoryClient.CancelReservation(ctx, &inventorypb.CancelReservationRequest{
        ReservationId: reservationID,
    })
    if err != nil {
        return fmt.Errorf("failed to cancel inventory reservation: %w", err)
    }
    
    if !resp.Success {
        return fmt.Errorf("inventory cancellation failed: %s", resp.Message)
    }
    
    return nil
}
```

## Error Handling

### 1. Inventory Unavailable

If inventory is unavailable during order creation, the Order Service returns a detailed error message:

```go
// Example error response
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

### 2. Reservation Failure

If inventory reservation fails, the Order Service rolls back the order creation:

```go
// Create order with inventory reservation
func (s *OrderService) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*models.Order, error) {
    // ... validation and other steps ...
    
    // Create order in database
    order, err := s.createOrderInDB(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("failed to create order: %w", err)
    }
    
    // Reserve inventory
    reservationID, err := s.reserveInventory(ctx, order, orderItems)
    if err != nil {
        // Rollback order creation
        if deleteErr := s.orderRepo.DeleteOrder(ctx, order.ID); deleteErr != nil {
            s.logger.Error("Failed to delete order after reservation failure",
                zap.Error(deleteErr),
                zap.String("order_id", order.ID))
        }
        return nil, err
    }
    
    // Update order with reservation ID
    order.InventoryReservationID = reservationID
    if err := s.orderRepo.UpdateOrder(ctx, order); err != nil {
        // Try to cancel reservation
        if cancelErr := s.cancelInventoryReservation(ctx, reservationID); cancelErr != nil {
            s.logger.Error("Failed to cancel reservation after order update failure",
                zap.Error(cancelErr),
                zap.String("reservation_id", reservationID))
        }
        return nil, fmt.Errorf("failed to update order with reservation ID: %w", err)
    }
    
    // ... payment processing and other steps ...
    
    return order, nil
}
```

### 3. Commit Failure

If inventory commit fails after payment processing, the Order Service logs the error but continues with the order:

```go
// Process payment and commit inventory
func (s *OrderService) processPaymentAndCommitInventory(ctx context.Context, order *models.Order) error {
    // Process payment
    paymentResp, err := s.paymentClient.ProcessPayment(ctx, &paymentpb.ProcessPaymentRequest{
        OrderId:  order.ID,
        Amount:   order.TotalAmount,
        Currency: order.Currency,
        Method:   order.PaymentMethod,
    })
    if err != nil {
        // Cancel reservation and mark order as failed
        if cancelErr := s.cancelInventoryReservation(ctx, order.InventoryReservationID); cancelErr != nil {
            s.logger.Error("Failed to cancel reservation after payment failure",
                zap.Error(cancelErr),
                zap.String("reservation_id", order.InventoryReservationID))
        }
        s.orderRepo.UpdateOrderStatus(ctx, order.ID, "PAYMENT_FAILED", fmt.Sprintf("Payment failed: %v", err))
        return fmt.Errorf("payment processing failed: %w", err)
    }
    
    // Update order status
    order.PaymentStatus = paymentResp.Status
    order.Status = "CONFIRMED"
    if err := s.orderRepo.UpdateOrder(ctx, order); err != nil {
        s.logger.Error("Failed to update order status after payment",
            zap.Error(err),
            zap.String("order_id", order.ID))
        // Continue despite error, as payment was successful
    }
    
    // Commit inventory reservation
    if err := s.commitInventoryReservation(ctx, order.InventoryReservationID); err != nil {
        s.logger.Error("Failed to commit inventory reservation, manual intervention required",
            zap.Error(err),
            zap.String("order_id", order.ID),
            zap.String("reservation_id", order.InventoryReservationID))
        // Continue despite error, as order and payment were successful
        // This will require manual inventory adjustment
    }
    
    return nil
}
```

## Monitoring and Alerting

### 1. Metrics

The integration includes monitoring of key metrics:

- **Inventory Check Latency**: Time taken to check inventory availability
- **Reservation Success Rate**: Percentage of successful inventory reservations
- **Commit Success Rate**: Percentage of successful inventory commits
- **Cancellation Success Rate**: Percentage of successful inventory cancellations

### 2. Alerts

Alerts are configured for critical issues:

- **High Reservation Failure Rate**: Alert if reservation failures exceed threshold
- **High Commit Failure Rate**: Alert if commit failures exceed threshold
- **Inventory Inconsistency**: Alert if inventory levels become inconsistent

## Conclusion

The integration between the Order Service and Inventory Service is critical for maintaining accurate inventory levels and preventing overselling. By implementing proper inventory checks, reservations, and commits, we ensure a smooth customer experience while maintaining data consistency across services.

The design follows the Saga pattern to manage distributed transactions, with compensating transactions to handle failures at each step. This ensures that inventory levels remain accurate even in the face of failures in the order process.
