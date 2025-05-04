import { gql } from '@apollo/client';

// Fragment for inventory item fields
export const INVENTORY_ITEM_FIELDS = gql`
  fragment InventoryItemFields on InventoryItem {
    id
    sku
    total_quantity
    available_quantity
    reserved_quantity
    reorder_point
    reorder_quantity
    status
    last_updated
    created_at
    updated_at
    product {
      id
      title
      slug
      images {
        url
      }
    }
  }
`;

// Query to get inventory items with pagination and filters
export const GET_INVENTORY_ITEMS = gql`
  query GetInventoryItems($page: Int!, $limit: Int!, $lowStockOnly: Boolean) {
    inventoryItems(page: $page, limit: $limit, lowStockOnly: $lowStockOnly) {
      items {
        ...InventoryItemFields
      }
      pagination {
        current_page
        total_pages
        per_page
        total_items
      }
    }
  }
  ${INVENTORY_ITEM_FIELDS}
`;

// Query to get a single inventory item by ID
export const GET_INVENTORY_ITEM = gql`
  query GetInventoryItem($id: ID!) {
    inventoryItem(id: $id) {
      ...InventoryItemFields
      locations {
        id
        warehouse_id
        quantity
        available_quantity
        reserved_quantity
        warehouse {
          id
          name
          code
        }
      }
    }
  }
  ${INVENTORY_ITEM_FIELDS}
`;

// Query to get warehouses with pagination
export const GET_WAREHOUSES = gql`
  query GetWarehouses($page: Int!, $limit: Int!) {
    warehouses(page: $page, limit: $limit) {
      warehouses {
        id
        name
        code
        address
        city
        state
        country
        postal_code
        is_active
        priority
        item_count
        total_quantity
        created_at
        updated_at
      }
      pagination {
        current_page
        total_pages
        per_page
        total_items
      }
    }
  }
`;

// Query to get inventory transactions with pagination and filters
export const GET_INVENTORY_TRANSACTIONS = gql`
  query GetInventoryTransactions(
    $page: Int!,
    $limit: Int!,
    $transactionType: String,
    $warehouseId: String,
    $dateFrom: String,
    $dateTo: String
  ) {
    inventoryTransactions(
      page: $page,
      limit: $limit,
      transactionType: $transactionType,
      warehouseId: $warehouseId,
      dateFrom: $dateFrom,
      dateTo: $dateTo
    ) {
      transactions {
        id
        inventory_item_id
        transaction_type
        quantity
        warehouse_id
        reference_id
        reference_type
        notes
        created_by
        created_at
        inventory_item {
          id
          sku
          product {
            id
            title
          }
        }
        warehouse {
          id
          name
          code
        }
      }
      pagination {
        current_page
        total_pages
        per_page
        total_items
      }
    }
  }
`;
