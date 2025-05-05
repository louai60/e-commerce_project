"use client";
import { ProductListResponse, ProductService, BrandListResponse, CategoryListResponse } from '@/services/product.service';
import useSWR from 'swr';

interface FilterParams {
  [key: string]: string | number | boolean | undefined;
}

// Type-specific fetchers for each endpoint
const fetchProducts = async (page: number, limit: number, filters: FilterParams): Promise<ProductListResponse> => {
  if (process.env.NODE_ENV === 'development') {
    console.log(`Fetching products with page=${page}, limit=${limit}`);
  }

  try {
    const result = await ProductService.getProducts(page, limit, filters);

    if (process.env.NODE_ENV === 'development') {
      console.log(`Data fetched: ${result?.products?.length || 0} products`);
    }

    return result;
  } catch (error: unknown) {
    console.error('Error fetching products:', error);
    throw error;
  }
};

const fetchBrands = async (page: number, limit: number): Promise<BrandListResponse> => {
  if (process.env.NODE_ENV === 'development') {
    console.log(`Fetching brands with page=${page}, limit=${limit}`);
  }

  try {
    const result = await ProductService.getBrands(page, limit);

    if (process.env.NODE_ENV === 'development') {
      console.log(`Data fetched: ${result?.brands?.length || 0} brands`);
    }

    return result;
  } catch (error: unknown) {
    console.error('Error fetching brands:', error);
    throw error;
  }
};

const fetchCategories = async (page: number, limit: number): Promise<CategoryListResponse> => {
  if (process.env.NODE_ENV === 'development') {
    console.log(`Fetching categories with page=${page}, limit=${limit}`);
  }

  try {
    const result = await ProductService.getCategories(page, limit);

    if (process.env.NODE_ENV === 'development') {
      console.log(`Data fetched: ${result?.categories?.length || 0} categories`);
    }

    return result;
  } catch (error: unknown) {
    console.error('Error fetching categories:', error);
    throw error;
  }
};

export function useProducts(page = 1, limit = 10, filters: FilterParams = {}) {
  // Use stable key structure for SWR
  const { data, error, isLoading, mutate } = useSWR<ProductListResponse>(
    ['products', page, limit, filters],
    () => fetchProducts(page, limit, filters),
    {
      // Use consistent configuration
      revalidateOnFocus: false,
      revalidateOnMount: true,
      shouldRetryOnError: true,
      errorRetryCount: 3,
      revalidateIfStale: true,
      keepPreviousData: false, // Don't keep previous data to ensure fresh data on page change
      // Force revalidation on page change
      revalidateOnReconnect: true,
      // Reduce the cache time to ensure fresh data on page change
      dedupingInterval: 0
    }
  );

  // Enhanced logging for pagination debugging
  if (process.env.NODE_ENV === 'development') {
    console.log(`useProducts hook - Page: ${page}, Limit: ${limit}`);
    console.log('useProducts hook data:', data ? `${data.products?.length || 0} products loaded` : 'No data');
    if (data?.pagination) {
      console.log('Pagination info from API:', {
        current_page: data.pagination.current_page,
        total_pages: data.pagination.total_pages,
        per_page: data.pagination.per_page,
        total_items: data.pagination.total_items
      });
    }

    // Log the SWR cache key to help debug caching issues
    console.log('SWR cache key:', ['products', page, limit, filters]);
  }

  // Ensure we have valid data
  const products = Array.isArray(data?.products) ? data?.products : [];

  // Calculate pagination information
  const totalItems = data?.total || products.length;
  const totalPages = Math.max(1, Math.ceil(totalItems / limit));

  // Create a consistent pagination object
  const pagination = {
    current_page: page, // Always use the requested page
    total_pages: totalPages,
    per_page: limit,
    total_items: totalItems
  };

  // Log pagination details for debugging
  console.log('useProducts pagination calculation:', {
    page,
    limit,
    totalItems,
    totalPages,
    productsLength: products.length,
    apiPagination: data?.pagination
  });

  return {
    products,
    pagination,
    isLoading,
    isError: error,
    mutate,
  };
}

export function useProduct(id: string | null) {
  const { data, error, isLoading, mutate } = useSWR(
    id ? `/products/${id}` : null,
    id ? () => ProductService.getProduct(id) : null,
    {
      revalidateOnFocus: false,
      dedupingInterval: 2000,
      revalidateOnMount: true,
      keepPreviousData: true
    }
  );

  return {
    product: data,
    isLoading,
    isError: error,
    mutate,
  };
}

export function useBrands(page = 1, limit = 10) {
  const { data, error, isLoading, mutate } = useSWR<BrandListResponse>(
    ['brands', page, limit],
    () => fetchBrands(page, limit),
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
  const brands = Array.isArray(data?.brands) ? data?.brands : [];

  // Calculate pagination information
  const totalItems = data?.total || brands.length;
  const totalPages = Math.max(1, Math.ceil(totalItems / limit));

  // Create a consistent pagination object
  const pagination = {
    current_page: page, // Always use the requested page
    total_pages: totalPages,
    per_page: limit,
    total_items: totalItems
  };

  return {
    brands,
    pagination,
    isLoading,
    isError: error,
    mutate,
  };
}

export function useCategories(page = 1, limit = 10) {
  const { data, error, isLoading, mutate } = useSWR<CategoryListResponse>(
    ['categories', page, limit],
    () => fetchCategories(page, limit),
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
  const categories = Array.isArray(data?.categories) ? data?.categories : [];

  // Calculate pagination information
  const totalItems = data?.total || categories.length;
  const totalPages = Math.max(1, Math.ceil(totalItems / limit));

  // Create a consistent pagination object
  const pagination = {
    current_page: page, // Always use the requested page
    total_pages: totalPages,
    per_page: limit,
    total_items: totalItems
  };

  return {
    categories,
    pagination,
    isLoading,
    isError: error,
    mutate,
  };
}
