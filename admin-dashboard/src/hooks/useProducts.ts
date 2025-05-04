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
    ['brands', page, limit],
    () => fetchBrands(page, limit),
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
    ['categories', page, limit],
    () => fetchCategories(page, limit),
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
