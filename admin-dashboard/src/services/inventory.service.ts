import { api } from '@/lib/api';

export interface InventoryItem {
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
  last_updated: string;
  created_at: string;
  updated_at: string;
  locations?: InventoryLocation[];
  product?: {
    id: string;
    title: string;
    slug: string;
    images?: { url: string }[];
  };
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
  created_at: string;
  updated_at: string;
  item_count?: number;
  total_quantity?: number;
}

export interface InventoryLocation {
  id: string;
  inventory_item_id: string;
  warehouse_id: string;
  quantity: number;
  available_quantity: number;
  reserved_quantity: number;
  created_at: string;
  updated_at: string;
  warehouse?: Warehouse;
}

export interface InventoryTransaction {
  id: string;
  inventory_item_id: string;
  warehouse_id?: string;
  transaction_type: string;
  quantity: number;
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

export interface InventoryItemsResponse {
  items: InventoryItem[];
  total: number;
  pagination: {
    current_page: number;
    total_pages: number;
    per_page: number;
    total_items: number;
  };
}

export interface WarehousesResponse {
  warehouses: Warehouse[];
  total: number;
  pagination: {
    current_page: number;
    total_pages: number;
    per_page: number;
    total_items: number;
  };
}

export interface TransactionsResponse {
  transactions: InventoryTransaction[];
  total: number;
  pagination: {
    current_page: number;
    total_pages: number;
    per_page: number;
    total_items: number;
  };
}

export class InventoryService {
  static async getInventoryItems(page = 1, limit = 10, filters = {}): Promise<InventoryItemsResponse> {
    try {
      const params = { page, limit, ...filters };
      const response = await api.get('/inventory/items', { params });
      return response.data;
    } catch (error) {
      console.error('Error fetching inventory items:', error);
      throw error;
    }
  }

  static async getInventoryItem(id: string): Promise<InventoryItem> {
    try {
      const response = await api.get(`/inventory/items/${id}`);
      return response.data;
    } catch (error) {
      console.error(`Error fetching inventory item ${id}:`, error);
      throw error;
    }
  }

  static async getWarehouses(page = 1, limit = 10, filters = {}): Promise<WarehousesResponse> {
    try {
      const params = { page, limit, ...filters };
      const response = await api.get('/inventory/warehouses', { params });
      return response.data;
    } catch (error) {
      console.error('Error fetching warehouses:', error);
      throw error;
    }
  }

  static async getWarehouse(id: string): Promise<Warehouse> {
    try {
      const response = await api.get(`/inventory/warehouses/${id}`);
      return response.data;
    } catch (error) {
      console.error(`Error fetching warehouse ${id}:`, error);
      throw error;
    }
  }

  static async createWarehouse(data: Partial<Warehouse>): Promise<Warehouse> {
    try {
      const response = await api.post('/inventory/warehouses', data);
      return response.data;
    } catch (error) {
      console.error('Error creating warehouse:', error);
      throw error;
    }
  }

  static async updateWarehouse(id: string, data: Partial<Warehouse>): Promise<Warehouse> {
    try {
      const response = await api.put(`/inventory/warehouses/${id}`, data);
      return response.data;
    } catch (error) {
      console.error(`Error updating warehouse ${id}:`, error);
      throw error;
    }
  }

  static async getInventoryTransactions(page = 1, limit = 10, filters = {}): Promise<TransactionsResponse> {
    try {
      const params = { page, limit, ...filters };
      const response = await api.get('/inventory/transactions', { params });
      return response.data;
    } catch (error) {
      console.error('Error fetching inventory transactions:', error);
      throw error;
    }
  }

  static async addInventory(
    inventoryItemId: string,
    warehouseId: string,
    quantity: number,
    referenceType?: string,
    notes?: string
  ): Promise<InventoryLocation> {
    try {
      const response = await api.post('/inventory/add', {
        inventory_item_id: inventoryItemId,
        warehouse_id: warehouseId,
        quantity,
        reference_type: referenceType || 'MANUAL_ADDITION',
        notes
      });
      return response.data;
    } catch (error) {
      console.error('Error adding inventory:', error);
      throw error;
    }
  }

  static async removeInventory(
    inventoryItemId: string,
    warehouseId: string,
    quantity: number,
    referenceType?: string,
    notes?: string
  ): Promise<InventoryLocation> {
    try {
      const response = await api.post('/inventory/remove', {
        inventory_item_id: inventoryItemId,
        warehouse_id: warehouseId,
        quantity,
        reference_type: referenceType || 'MANUAL_REMOVAL',
        notes
      });
      return response.data;
    } catch (error) {
      console.error('Error removing inventory:', error);
      throw error;
    }
  }

  static async transferInventory(
    inventoryItemId: string,
    fromWarehouseId: string,
    toWarehouseId: string,
    quantity: number,
    notes?: string
  ): Promise<{ from: InventoryLocation, to: InventoryLocation }> {
    try {
      const response = await api.post('/inventory/transfer', {
        inventory_item_id: inventoryItemId,
        from_warehouse_id: fromWarehouseId,
        to_warehouse_id: toWarehouseId,
        quantity,
        notes
      });
      return response.data;
    } catch (error) {
      console.error('Error transferring inventory:', error);
      throw error;
    }
  }

  static async checkInventoryAvailability(productId: string, quantity: number): Promise<{ available: boolean }> {
    try {
      const response = await api.get('/inventory/check', {
        params: {
          product_id: productId,
          quantity
        }
      });
      return response.data;
    } catch (error) {
      console.error('Error checking inventory availability:', error);
      throw error;
    }
  }
}
