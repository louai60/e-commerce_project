import useSWR from 'swr';
import axios from 'axios';

const fetcher = (url: string) => axios.get(url).then(res => res.data);

export function useProducts(page = 1, limit = 10) {
  const { data, error, isLoading, mutate } = useSWR(
    `/api/products?page=${page}&limit=${limit}`,
    fetcher,
    { revalidateOnMount: true }
  );

  return {
    products: data?.products,
    total: data?.total,
    isLoading,
    isError: error,
    mutate,
  };
}

export function useProduct(id: string) {
  const { data, error, isLoading } = useSWR(
    id ? `/api/products/${id}` : null,
    fetcher
  );

  return {
    product: data,
    isLoading,
    isError: error,
  };
}