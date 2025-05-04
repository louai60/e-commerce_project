"use client";
import { ProductListResponse, ProductService, BrandListResponse, CategoryListResponse } from '@/services/product.service';
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
    if (endpoint === '/products') {
      result = await ProductService.getProducts(page, limit, filters);
    } else if (endpoint === '/brands') {
      result = await ProductService.getBrands(page, limit);
    } else if (endpoint === '/categories') {
      result = await ProductService.getCategories(page, limit);
    } else {
      throw new Error(`Unsupported endpoint: ${endpoint}`);
    }

    // Only log in development mode
    if (process.env.NODE_ENV === 'development') {
      console.log(`Data fetched from ${endpoint}:`,
        endpoint === '/products' ? `${result?.products?.length || 0} products` : 'data');
    }
    return result;
  } catch (error: unknown) {
    console.error(`Error fetching data from ${endpoint}:`, error);
    throw error;
  }
};

export function useProducts(page = 1, limit = 10, filters: FilterParams = {}) {
  // Use stable key structure for SWR
  const { data, error, isLoading, mutate } = useSWR<ProductListResponse>(
    ['/products', page, limit, filters],
    ([url, page, limit, filters]: [string, number, number, FilterParams]) => fetcher(url, page, limit, filters),
    {
      // Use consistent configuration
      revalidateOnFocus: false,
      dedupingInterval: 2000,
      revalidateOnMount: true,
      shouldRetryOnError: true,
      errorRetryCount: 3,
      revalidateIfStale: true,
      keepPreviousData: true
    }
  );

  // Only log data in development mode
  if (process.env.NODE_ENV === 'development') {
    console.log('useProducts hook data:', data ? `${data.products?.length || 0} products loaded` : 'No data');
  }

  // Ensure we have valid data
  const products = Array.isArray(data?.products) ? data?.products : [];
  const pagination = data?.pagination || {
    current_page: page,
    total_pages: 1,
    per_page: limit,
    total_items: products.length
  };

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
    ['/brands', page, limit, {}],
    ([url, page, limit, filters]) => fetcher(url, page, limit, filters),
    {
      revalidateOnFocus: false,
      dedupingInterval: 2000,
      revalidateOnMount: true,
      keepPreviousData: true
    }
  );

  return {
    brands: data?.brands || [],
    pagination: data?.pagination || { current_page: 1, total_pages: 1, per_page: limit, total_items: 0 },
    isLoading,
    isError: error,
    mutate,
  };
}

export function useCategories(page = 1, limit = 10) {
  const { data, error, isLoading, mutate } = useSWR<CategoryListResponse>(
    ['/categories', page, limit, {}],
    ([url, page, limit, filters]) => fetcher(url, page, limit, filters),
    {
      revalidateOnFocus: false,
      dedupingInterval: 2000,
      revalidateOnMount: true,
      keepPreviousData: true
    }
  );

  return {
    categories: data?.categories || [],
    pagination: data?.pagination || { current_page: 1, total_pages: 1, per_page: limit, total_items: 0 },
    isLoading,
    isError: error,
    mutate,
  };
}
