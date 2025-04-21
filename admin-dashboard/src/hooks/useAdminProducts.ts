"use client";
import { api } from '@/lib/api';
import useSWR from 'swr';

interface Product {
  id: string;
  name?: string;
  title?: string;
  description: string;
  price: number;
  category_id: string;
  image: string;
  variants: number | any[] | Record<string, any>;
  status: "Active" | "Inactive";
  created_at: string;
}

interface ProductsResponse {
  products: Product[];
  total_count: number;
}

const fetcher = async (url: string) => {
  const response = await api.get(url);
  return response.data;
};

export function useAdminProducts(page = 1, limit = 10) {
  const { data, error, isLoading, mutate } = useSWR<ProductsResponse>(
    `/products?page=${page}&limit=${limit}`,
    fetcher
  );

  return {
    products: data?.products || [],
    totalCount: data?.total_count || 0,
    isLoading,
    isError: error,
    mutate,
  };
}