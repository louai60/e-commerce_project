# Inventory Service Architecture

## Overview

The Inventory Service is a microservice responsible for managing product inventory in the e-commerce system. It handles inventory tracking, stock management, and provides inventory data to other services. This document outlines the architecture, communication patterns, and key components of the Inventory Service.

## Architecture Components

### 1. Core Components

The Inventory Service consists of the following core components:

- **gRPC Server**: Exposes inventory management operations to other services
- **Inventory Service Layer**: Contains business logic for inventory operations
- **Repository Layer**: Handles data persistence and retrieval
- **Database**: Stores inventory data in PostgreSQL

### 2. Data Model

The Inventory Service manages the following data entities:

#### Inventory Items
- Represents the inventory for a product or variant
- Contains quantities, status, and reorder information
- Linked to products via product_id

#### Inventory Locations
- Represents inventory at specific warehouses
- Contains quantity information per location
- Enables multi-warehouse inventory management

#### Inventory Transactions
- Records all inventory movements (additions, removals)
- Provides audit trail for inventory changes
- Contains reference information for traceability

#### Inventory Reservations
- Manages temporary holds on inventory
- Used during checkout process
- Contains expiration time for automatic release

## Communication Patterns

### 1. Service Integration

The Inventory Service communicates with other services using:

- **gRPC**: For synchronous service-to-service communication
- **Event-Based Communication**: For asynchronous updates (future implementation)

### 2. API Gateway Integration

The API Gateway integrates with the Inventory Service to:

- Create inventory items when products are created
- Fetch inventory data for product responses
- Update inventory quantities
- Check inventory availability

### 3. Communication Flow

When a product is created:

1. The API Gateway receives the product creation request with inventory data
2. The API Gateway creates the product in the Product Service
3. The API Gateway calls the Inventory Service to create an inventory item
4. The Inventory Service creates the inventory item and returns the result
5. The API Gateway fetches the inventory data and includes it in the product response

## Key Operations

### 1. Inventory Creation

When creating inventory:

```
CreateInventoryItem(
    productID string,
    sku string,
    initialQty int,
    reorderPoint int,
    reorderQuantity int
) -> InventoryItem
```

- Creates a new inventory item for a product
- Sets initial quantity, reorder points
- Determines inventory status based on quantity and reorder point
- Creates transaction record for initial stock

### 2. Inventory Retrieval

To retrieve inventory data:

```
GetInventoryItem(
    productID string
) -> InventoryItem
```

- Fetches inventory data for a product
- Includes available quantity, reserved quantity, and status
- Includes inventory locations if available

### 3. Inventory Updates

For inventory quantity changes:

```
AddInventoryToLocation(
    inventoryItemID string,
    warehouseID string,
    quantity int,
    referenceID string,
    referenceType string,
    notes string
) -> InventoryLocation
```

```
RemoveInventoryFromLocation(
    inventoryItemID string,
    warehouseID string,
    quantity int,
    referenceID string,
    referenceType string,
    notes string
) -> InventoryLocation
```

- Updates inventory quantities at specific locations
- Creates transaction records for audit trail
- Updates inventory status based on new quantities

### 4. Inventory Availability Check

To check if inventory is available:

```
CheckInventoryAvailability(
    items []AvailabilityCheckItem
) -> InventoryAvailabilityResponse
```

- Checks if requested quantities are available
- Returns availability status for each item
- Used during checkout process

## Database Schema

The Inventory Service uses the following database tables:

### inventory_items
```sql
CREATE TABLE inventory_items (
    id UUID PRIMARY KEY,
    product_id UUID NOT NULL,
    variant_id UUID,
    sku VARCHAR(50) NOT NULL,
    total_quantity INTEGER NOT NULL DEFAULT 0,
    available_quantity INTEGER NOT NULL DEFAULT 0,
    reserved_quantity INTEGER NOT NULL DEFAULT 0,
    reorder_point INTEGER NOT NULL DEFAULT 0,
    reorder_quantity INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL,
    last_updated TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);
```

### inventory_locations
```sql
CREATE TABLE inventory_locations (
    id UUID PRIMARY KEY,
    inventory_item_id UUID NOT NULL REFERENCES inventory_items(id),
    warehouse_id UUID NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 0,
    available_quantity INTEGER NOT NULL DEFAULT 0,
    reserved_quantity INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);
```

### inventory_transactions
```sql
CREATE TABLE inventory_transactions (
    id UUID PRIMARY KEY,
    inventory_item_id UUID NOT NULL REFERENCES inventory_items(id),
    warehouse_id UUID,
    transaction_type VARCHAR(20) NOT NULL,
    quantity INTEGER NOT NULL,
    reference_id UUID,
    reference_type VARCHAR(50),
    notes TEXT,
    created_by UUID,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);
```

### inventory_reservations
```sql
CREATE TABLE inventory_reservations (
    id UUID PRIMARY KEY,
    inventory_item_id UUID NOT NULL REFERENCES inventory_items(id),
    warehouse_id UUID,
    quantity INTEGER NOT NULL,
    status VARCHAR(20) NOT NULL,
    expiration_time TIMESTAMP WITH TIME ZONE NOT NULL,
    reference_id UUID,
    reference_type VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);
```

## Configuration

The Inventory Service is configured using environment variables:

- `DB_HOST`: Database host
- `DB_PORT`: Database port
- `DB_USER`: Database username
- `DB_PASSWORD`: Database password
- `DB_NAME`: Database name
- `SERVICE_PORT`: gRPC server port

## Deployment

The Inventory Service can be deployed as:

- Standalone service
- Docker container
- Kubernetes pod

## Future Enhancements

Planned enhancements for the Inventory Service include:

1. Event-driven architecture for real-time inventory updates
2. Inventory forecasting and analytics
3. Batch inventory operations for improved performance
4. Integration with supplier systems for automatic reordering
