"use client";

import { useQuery } from '@apollo/client';
import { GET_INVENTORY_ITEMS, GET_WAREHOUSES, GET_INVENTORY_ITEM, GET_INVENTORY_TRANSACTIONS } from '@/graphql/queries/inventory';

// Type definitions
export interface InventoryItemImage {
  url: string;
}

export interface InventoryItemProduct {
  id: string;
  title: string;
  slug: string;
  images?: InventoryItemImage[];
}

export interface InventoryItem {
  id: string;
  sku: string;
  total_quantity: number;
  available_quantity: number;
  reserved_quantity: number;
  reorder_point: number;
  reorder_quantity: number;
  status: string;
  last_updated: string;
  created_at: string;
  updated_at: string;
  product?: InventoryItemProduct;
  locations?: InventoryLocation[];
}

export interface InventoryLocation {
  id: string;
  warehouse_id: string;
  quantity: number;
  available_quantity: number;
  reserved_quantity: number;
  warehouse?: Warehouse;
}

export interface Warehouse {
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
  item_count?: number;
  total_quantity?: number;
  created_at: string;
  updated_at: string;
}

export interface Pagination {
  current_page: number;
  total_pages: number;
  per_page: number;
  total_items: number;
}

export interface InventoryItemsResponse {
  items: InventoryItem[];
  pagination: Pagination;
}

export interface WarehousesResponse {
  warehouses: Warehouse[];
  pagination: Pagination;
}

// Hook for fetching inventory items
export function useInventoryItemsGraphQL(page = 1, limit = 10, filters: { lowStockOnly?: boolean } = {}) {
  const { data, loading, error, refetch } = useQuery(GET_INVENTORY_ITEMS, {
    variables: {
      page,
      limit,
      lowStockOnly: filters.lowStockOnly || false
    },
    fetchPolicy: 'cache-and-network',
  });

  // Ensure we have valid data
  const items = data?.inventoryItems?.items || [];
  const pagination = data?.inventoryItems?.pagination || {
    current_page: page,
    total_pages: 1,
    per_page: limit,
    total_items: items.length
  };

  return {
    items,
    pagination,
    isLoading: loading,
    isError: error,
    refetch,
  };
}

// Hook for fetching a single inventory item
export function useInventoryItemGraphQL(id: string | null) {
  const { data, loading, error, refetch } = useQuery(GET_INVENTORY_ITEM, {
    variables: { id },
    skip: !id,
    fetchPolicy: 'cache-and-network',
  });

  return {
    item: data?.inventoryItem,
    isLoading: loading,
    isError: error,
    refetch,
  };
}

// Hook for fetching warehouses
export function useWarehousesGraphQL(page = 1, limit = 10) {
  const { data, loading, error, refetch } = useQuery(GET_WAREHOUSES, {
    variables: { page, limit },
    fetchPolicy: 'cache-and-network',
  });

  // Ensure we have valid data
  const warehouses = data?.warehouses?.warehouses || [];
  const pagination = data?.warehouses?.pagination || {
    current_page: page,
    total_pages: 1,
    per_page: limit,
    total_items: warehouses.length
  };

  return {
    warehouses,
    pagination,
    isLoading: loading,
    isError: error,
    refetch,
  };
}

// Type definitions for inventory transactions
export interface InventoryTransaction {
  id: string;
  inventory_item_id: string;
  transaction_type: string;
  quantity: number;
  warehouse_id?: string;
  reference_id?: string;
  reference_type?: string;
  notes?: string;
  created_by?: string;
  created_at: string;
  inventory_item?: {
    id: string;
    sku: string;
    product?: {
      id: string;
      title: string;
    };
  };
  warehouse?: {
    id: string;
    name: string;
    code: string;
  };
}

// Hook for fetching inventory transactions
export function useInventoryTransactionsGraphQL(
  page = 1,
  limit = 10,
  filters: {
    transactionType?: string;
    warehouseId?: string;
    dateFrom?: string;
    dateTo?: string;
  } = {}
) {
  const { data, loading, error, refetch } = useQuery(GET_INVENTORY_TRANSACTIONS, {
    variables: {
      page,
      limit,
      transactionType: filters.transactionType || null,
      warehouseId: filters.warehouseId || null,
      dateFrom: filters.dateFrom || null,
      dateTo: filters.dateTo || null
    },
    fetchPolicy: 'cache-and-network',
  });

  // Ensure we have valid data
  const transactions = data?.inventoryTransactions?.transactions || [];
  const pagination = data?.inventoryTransactions?.pagination || {
    current_page: page,
    total_pages: 1,
    per_page: limit,
    total_items: transactions.length
  };

  return {
    transactions,
    pagination,
    isLoading: loading,
    isError: error,
    refetch,
  };
}
