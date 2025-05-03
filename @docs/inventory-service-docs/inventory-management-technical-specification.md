# Inventory Management System Technical Specification

## 1. Introduction

This document outlines the technical specifications for implementing an inventory management system in the admin dashboard of the e-commerce platform. The system will allow administrators to manage inventory levels, warehouses, and track inventory movements.

## 2. System Architecture

The inventory management system will integrate with the existing architecture:

```
┌─────────────────┐      ┌─────────────────┐      ┌─────────────────┐
│                 │      │                 │      │                 │
│  Admin Dashboard│◄────►│   API Gateway   │◄────►│ Inventory Service│
│  (Next.js)      │      │   (Go)          │      │  (Go)           │
│                 │      │                 │      │                 │
└─────────────────┘      └─────────────────┘      └─────────────────┘
                                                          │
                                                          ▼
                                                  ┌─────────────────┐
                                                  │                 │
                                                  │  Database       │
                                                  │  (PostgreSQL)   │
                                                  │                 │
                                                  └─────────────────┘
```

### 2.1 Components

1. **Admin Dashboard (Frontend)**
   - Next.js application with React components
   - Communicates with the API Gateway

2. **API Gateway**
   - Go-based service that routes requests to appropriate microservices
   - Handles authentication and authorization

3. **Inventory Service**
   - Go-based microservice for inventory management
   - Manages inventory items, warehouses, and transactions
   - Communicates with the database

4. **Database**
   - PostgreSQL database with tables for inventory items, warehouses, locations, transactions, and reservations

## 3. Data Models

### 3.1 Inventory Item

```typescript
interface InventoryItem {
  id: string;
  product_id: string;
  variant_id?: string;
  sku: string;
  total_quantity: number;
  available_quantity: number;
  reserved_quantity: number;
  reorder_point: number;
  reorder_quantity: number;
  status: string; // "IN_STOCK", "LOW_STOCK", "OUT_OF_STOCK"
  last_updated: string; // ISO date
  created_at: string; // ISO date
  updated_at: string; // ISO date
  locations?: InventoryLocation[];
}
```

### 3.2 Warehouse

```typescript
interface Warehouse {
  id: string;
  name: string;
  code: string;
  address?: string;
  city?: string;
  state?: string;
  country?: string;
  postal_code?: string;
  is_active: boolean;
  priority: number;
  created_at: string; // ISO date
  updated_at: string; // ISO date
}
```

### 3.3 Inventory Location

```typescript
interface InventoryLocation {
  id: string;
  inventory_item_id: string;
  warehouse_id: string;
  quantity: number;
  available_quantity: number;
  reserved_quantity: number;
  created_at: string; // ISO date
  updated_at: string; // ISO date
  warehouse?: Warehouse;
}
```

### 3.4 Inventory Transaction

```typescript
interface InventoryTransaction {
  id: string;
  inventory_item_id: string;
  warehouse_id?: string;
  transaction_type: string; // "STOCK_ADDITION", "STOCK_REMOVAL", "TRANSFER", "RESERVATION", "ADJUSTMENT"
  quantity: number;
  reference_id?: string;
  reference_type?: string;
  notes?: string;
  created_by?: string;
  created_at: string; // ISO date
}
```

## 4. API Endpoints

### 4.1 Inventory Items

| Endpoint | Method | Description | Request Body | Response |
|----------|--------|-------------|-------------|----------|
| `/api/v1/inventory/items` | GET | List inventory items | Query params: page, limit, status, warehouse_id, low_stock_only | List of inventory items with pagination |
| `/api/v1/inventory/items/:id` | GET | Get inventory item details | - | Inventory item with locations |
| `/api/v1/inventory/items` | POST | Create inventory item | Inventory item data | Created inventory item |
| `/api/v1/inventory/items/:id` | PUT | Update inventory item | Updated inventory item data | Updated inventory item |

### 4.2 Warehouses

| Endpoint | Method | Description | Request Body | Response |
|----------|--------|-------------|-------------|----------|
| `/api/v1/inventory/warehouses` | GET | List warehouses | Query params: page, limit, is_active | List of warehouses with pagination |
| `/api/v1/inventory/warehouses/:id` | GET | Get warehouse details | - | Warehouse details |
| `/api/v1/inventory/warehouses` | POST | Create warehouse | Warehouse data | Created warehouse |
| `/api/v1/inventory/warehouses/:id` | PUT | Update warehouse | Updated warehouse data | Updated warehouse |

### 4.3 Inventory Transactions

| Endpoint | Method | Description | Request Body | Response |
|----------|--------|-------------|-------------|----------|
| `/api/v1/inventory/transactions` | GET | List transactions | Query params: page, limit, inventory_item_id, warehouse_id, transaction_type | List of transactions with pagination |
| `/api/v1/inventory/transactions` | POST | Create transaction | Transaction data | Created transaction |

### 4.4 Inventory Operations

| Endpoint | Method | Description | Request Body | Response |
|----------|--------|-------------|-------------|----------|
| `/api/v1/inventory/add` | POST | Add inventory | { inventory_item_id, warehouse_id, quantity, reference_id?, reference_type?, notes? } | Updated inventory location |
| `/api/v1/inventory/remove` | POST | Remove inventory | { inventory_item_id, warehouse_id, quantity, reference_id?, reference_type?, notes? } | Updated inventory location |
| `/api/v1/inventory/transfer` | POST | Transfer inventory | { inventory_item_id, from_warehouse_id, to_warehouse_id, quantity, notes? } | Updated inventory locations |
| `/api/v1/inventory/check` | GET | Check availability | Query params: product_id, quantity | Availability status |

## 5. Frontend Components

### 5.1 Pages

1. **Inventory Dashboard** (`/inventory`)
   - Overview of inventory status
   - Key metrics (total items, low stock items, out of stock items)
   - Recent transactions
   - Low stock alerts

2. **Inventory Items** (`/inventory/items`)
   - Table listing of all inventory items
   - Filtering by status, warehouse, product
   - Pagination
   - Actions for viewing details, updating inventory

3. **Inventory Item Details** (`/inventory/items/[id]`)
   - Detailed view of an inventory item
   - Product information
   - Quantity by warehouse
   - Transaction history
   - Actions for adding/removing inventory

4. **Warehouses** (`/inventory/warehouses`)
   - Table listing of all warehouses
   - Filtering by status, location
   - Pagination
   - Actions for viewing details, editing

5. **Warehouse Details** (`/inventory/warehouses/[id]`)
   - Detailed view of a warehouse
   - Inventory items in the warehouse
   - Actions for managing inventory

6. **Inventory Transactions** (`/inventory/transactions`)
   - Table listing of all transactions
   - Filtering by type, date, item, warehouse
   - Pagination

### 5.2 Components

1. **InventoryMetrics**
   - Display key inventory metrics
   - Low stock alerts
   - Inventory status distribution

2. **InventoryTable**
   - Reusable table for inventory items
   - Sorting, filtering, pagination
   - Actions column

3. **WarehouseTable**
   - Reusable table for warehouses
   - Sorting, filtering, pagination
   - Actions column

4. **InventoryForm**
   - Form for adding/removing inventory
   - Warehouse selection
   - Quantity input
   - Reason/notes input

5. **WarehouseForm**
   - Form for creating/editing warehouses
   - Validation for required fields

6. **InventoryStatusBadge**
   - Visual indicator of inventory status
   - Color-coded for different statuses

7. **InventoryTransactionTable**
   - Table for displaying inventory transactions
   - Filtering by type, date

## 6. State Management

### 6.1 Inventory Context

```typescript
interface InventoryContextType {
  refreshInventory: () => void;
  addInventory: (itemId: string, warehouseId: string, quantity: number, reason?: string) => Promise<boolean>;
  removeInventory: (itemId: string, warehouseId: string, quantity: number, reason?: string) => Promise<boolean>;
  transferInventory: (itemId: string, fromWarehouseId: string, toWarehouseId: string, quantity: number) => Promise<boolean>;
  isRefreshing: boolean;
  isUpdating: boolean;
}
```

### 6.2 Hooks

1. **useInventoryItems**
   - Fetch inventory items with pagination and filtering
   - Cache results with SWR

2. **useInventoryItem**
   - Fetch a single inventory item by ID
   - Include locations and recent transactions

3. **useWarehouses**
   - Fetch warehouses with pagination and filtering
   - Cache results with SWR

4. **useWarehouse**
   - Fetch a single warehouse by ID
   - Include inventory items in the warehouse

5. **useInventoryTransactions**
   - Fetch inventory transactions with pagination and filtering
   - Cache results with SWR

## 7. Integration with Product Management

### 7.1 Product Creation/Editing

- Add inventory fields to product creation/editing forms
- Allow setting initial inventory levels
- Implement SKU generation/management
- Add warehouse allocation during product creation

### 7.2 Product Listing

- Enhance product listing with inventory status
- Add inventory filters to product search
- Implement quick inventory update from product list

## 8. Security Considerations

- All inventory management endpoints require admin authentication
- Implement proper validation for all input data
- Log all inventory transactions with user information
- Implement optimistic locking for inventory updates to prevent race conditions

## 9. Performance Considerations

- Implement caching for inventory data
- Use pagination for all list endpoints
- Optimize database queries with proper indexing
- Consider using WebSockets for real-time inventory updates

## 10. Testing Strategy

- Unit tests for all API endpoints
- Integration tests for inventory operations
- UI tests for inventory management interfaces
- Load testing for concurrent inventory operations

## 11. Deployment Strategy

- Deploy backend changes first
- Deploy frontend changes with feature flags
- Monitor inventory operations after deployment
- Have rollback plan in case of issues
