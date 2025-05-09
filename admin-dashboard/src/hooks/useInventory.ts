"use client";
import {
  InventoryItemsResponse,
  WarehousesResponse,
  TransactionsResponse,
  InventoryService
} from '@/services/inventory.service';
import useSWR from 'swr';

interface FilterParams {
  [key: string]: string | number | boolean | undefined;
}

const fetcher = async (url: string, page: number, limit: number, filters: FilterParams) => {
  // Extract the base endpoint from the URL
  const endpoint = url.split('?')[0];

  // Only log in development mode
  if (process.env.NODE_ENV === 'development') {
    console.log(`Fetching data from ${endpoint} with page=${page}, limit=${limit}`);
  }

  try {
    // Call the appropriate service method based on the endpoint
    let result;
    if (endpoint === '/inventory/items') {
      result = await InventoryService.getInventoryItems(page, limit, filters);
    } else if (endpoint === '/inventory/warehouses') {
      result = await InventoryService.getWarehouses(page, limit, filters);
    } else if (endpoint === '/inventory/transactions') {
      result = await InventoryService.getInventoryTransactions(page, limit, filters);
    } else {
      throw new Error(`Unsupported endpoint: ${endpoint}`);
    }

    // Only log in development mode
    if (process.env.NODE_ENV === 'development') {
      console.log(`Data fetched from ${endpoint}`);
    }
    return result;
  } catch (error) {
    console.error(`Error fetching data from ${endpoint}:`, error);
    throw error;
  }
};

export function useInventoryItems(page = 1, limit = 10, filters: FilterParams = {}) {
  const { data, error, isLoading, mutate } = useSWR<InventoryItemsResponse>(
    ['/inventory/items', page, limit, filters],
    async (args) => {
      const [url, p, l, f] = args;
      return fetcher(url as string, p as number, l as number, f as FilterParams) as Promise<InventoryItemsResponse>;
    },
    {
      revalidateOnFocus: false,
      dedupingInterval: 5000,
      revalidateOnMount: true,
      shouldRetryOnError: true,
      errorRetryCount: 3,
      revalidateIfStale: true,
      keepPreviousData: true,
      revalidateOnReconnect: true
    }
  );

  // Enhanced logging for pagination debugging
  if (process.env.NODE_ENV === 'development') {
    console.log(`useInventoryItems hook - Page: ${page}, Limit: ${limit}`);
    console.log('useInventoryItems hook data:', data ? `${data.items?.length || 0} items loaded` : 'No data');
    if (data?.pagination) {
      console.log('Pagination info:', data.pagination);
    }
  }

  // Ensure we have valid data
  const items = Array.isArray(data?.items) ? data?.items : [];

  // Always use the current page from the request, not from the response
  // This ensures consistency when navigating between pages
  const pagination = {
    ...(data?.pagination || {
      total_pages: 1,
      per_page: limit,
      total_items: items.length
    }),
    current_page: page // Always use the requested page
  };

  return {
    items,
    pagination,
    isLoading,
    isError: error,
    mutate,
  };
}

export function useInventoryItem(id: string | null) {
  const { data, error, isLoading, mutate } = useSWR(
    id ? `/inventory/items/${id}` : null,
    id ? () => InventoryService.getInventoryItem(id) : null,
    {
      revalidateOnFocus: false,
      dedupingInterval: 2000,
      revalidateOnMount: true,
      keepPreviousData: true
    }
  );

  return {
    item: data,
    isLoading,
    isError: error,
    mutate,
  };
}

export function useWarehouses(page = 1, limit = 10, filters: FilterParams = {}) {
  const { data, error, isLoading, mutate } = useSWR<WarehousesResponse>(
    ['/inventory/warehouses', page, limit, filters],
    async (args) => {
      const [url, p, l, f] = args;
      return fetcher(url as string, p as number, l as number, f as FilterParams) as Promise<WarehousesResponse>;
    },
    {
      revalidateOnFocus: false,
      dedupingInterval: 5000,
      revalidateOnMount: true,
      keepPreviousData: true,
      revalidateOnReconnect: true,
      shouldRetryOnError: true,
      errorRetryCount: 3,
      revalidateIfStale: true
    }
  );

  // Ensure we have valid data
  const warehouses = Array.isArray(data?.warehouses) ? data?.warehouses : [];

  // Always use the current page from the request, not from the response
  const pagination = {
    ...(data?.pagination || {
      total_pages: 1,
      per_page: limit,
      total_items: warehouses.length
    }),
    current_page: page // Always use the requested page
  };

  return {
    warehouses,
    pagination,
    isLoading,
    isError: error,
    mutate,
  };
}

export function useWarehouse(id: string | null) {
  const { data, error, isLoading, mutate } = useSWR(
    id ? `/inventory/warehouses/${id}` : null,
    id ? () => InventoryService.getWarehouse(id) : null,
    {
      revalidateOnFocus: false,
      dedupingInterval: 2000,
      revalidateOnMount: true,
      keepPreviousData: true
    }
  );

  return {
    warehouse: data,
    isLoading,
    isError: error,
    mutate,
  };
}

export function useInventoryTransactions(page = 1, limit = 10, filters: FilterParams = {}) {
  const { data, error, isLoading, mutate } = useSWR<TransactionsResponse>(
    ['/inventory/transactions', page, limit, filters],
    async (args) => {
      const [url, p, l, f] = args;
      return fetcher(url as string, p as number, l as number, f as FilterParams) as Promise<TransactionsResponse>;
    },
    {
      revalidateOnFocus: false,
      dedupingInterval: 5000,
      revalidateOnMount: true,
      keepPreviousData: true,
      revalidateOnReconnect: true,
      shouldRetryOnError: true,
      errorRetryCount: 3,
      revalidateIfStale: true
    }
  );

  // Ensure we have valid data
  const transactions = Array.isArray(data?.transactions) ? data?.transactions : [];

  // Always use the current page from the request, not from the response
  const pagination = {
    ...(data?.pagination || {
      total_pages: 1,
      per_page: limit,
      total_items: transactions.length
    }),
    current_page: page // Always use the requested page
  };

  return {
    transactions,
    pagination,
    isLoading,
    isError: error,
    mutate,
  };
}
