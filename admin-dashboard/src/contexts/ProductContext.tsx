"use client";
import React, { createContext, useContext, ReactNode, useState, useCallback } from 'react';
import { useProducts } from '@/hooks/useProducts';
import { Product } from '@/services/product.service';
import { api } from '@/lib/api';

interface ProductContextType {
  refreshProducts: () => void;
  addOptimisticProduct: (product: Product) => void;
  removeOptimisticProduct: (productId: string) => void;
  deleteProduct: (productId: string) => Promise<boolean>;
  isRefreshing: boolean;
  isDeleting: boolean;
}

const ProductContext = createContext<ProductContextType | undefined>(undefined);

export function ProductProvider({ children }: { children: ReactNode }) {
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const { mutate } = useProducts(1, 10, {});

  const refreshProducts = useCallback(() => {
    setIsRefreshing(true);

    if (process.env.NODE_ENV === 'development') {
      console.log('Refreshing products...');
    }

    mutate(undefined, { revalidate: true })
      .then((data) => {
        if (process.env.NODE_ENV === 'development') {
          console.log('Products refreshed successfully:', data?.products?.length || 0, 'products loaded');
        }
      })
      .catch((error: unknown) => {
        console.error('Error refreshing products:', error);
      })
      .finally(() => {
        setIsRefreshing(false);
      });
  }, [mutate]);

  const addOptimisticProduct = useCallback((newProduct: Product) => {
    if (process.env.NODE_ENV === 'development') {
      console.log('Adding optimistic product:', newProduct.title);
    }

    mutate(
      (currentData) => {
        if (!currentData) {
          if (process.env.NODE_ENV === 'development') {
            console.log('No current data available for optimistic update');
          }
          return {
            products: [newProduct],
            total: 1,
            pagination: {
              current_page: 1,
              total_pages: 1,
              per_page: 10,
              total_items: 1
            }
          };
        }

        const productExists = currentData.products.some(p => p.id === newProduct.id);
        if (productExists) {
          if (process.env.NODE_ENV === 'development') {
            console.log('Product already exists in the list, updating it');
          }
          return {
            ...currentData,
            products: currentData.products.map(p =>
              p.id === newProduct.id ? newProduct : p
            )
          };
        }

        if (process.env.NODE_ENV === 'development') {
          console.log('Adding new product to the list');
        }
        return {
          ...currentData,
          products: [newProduct, ...currentData.products],
          total: currentData.total + 1,
          pagination: {
            ...currentData.pagination,
            total_items: currentData.pagination.total_items + 1,
            total_pages: Math.ceil((currentData.pagination.total_items + 1) / currentData.pagination.per_page)
          }
        };
      },
      { revalidate: false }
    );
  }, [mutate]);

  const removeOptimisticProduct = useCallback((productId: string) => {
    if (process.env.NODE_ENV === 'development') {
      console.log('Removing product with ID:', productId);
    }

    mutate(
      (currentData) => {
        if (!currentData) {
          if (process.env.NODE_ENV === 'development') {
            console.log('No current data available for removal');
          }
          return currentData;
        }

        const productExists = currentData.products.some(p => p.id === productId);
        if (!productExists) {
          if (process.env.NODE_ENV === 'development') {
            console.log('Product not found in the list, nothing to remove');
          }
          return currentData;
        }

        if (process.env.NODE_ENV === 'development') {
          console.log('Removing product from the list');
        }
        const filteredProducts = currentData.products.filter(p => p.id !== productId);

        return {
          ...currentData,
          products: filteredProducts,
          total: Math.max(0, currentData.total - 1),
          pagination: {
            ...currentData.pagination,
            total_items: Math.max(0, currentData.pagination.total_items - 1),
            total_pages: Math.max(1, Math.ceil((currentData.pagination.total_items - 1) / currentData.pagination.per_page))
          }
        };
      },
      { revalidate: false }
    );
  }, [mutate]);

  const deleteProduct = useCallback(async (productId: string): Promise<boolean> => {
    if (!productId) return false;

    if (isDeleting) {
      console.warn('Delete operation already in progress');
      return false;
    }

    setIsDeleting(true);

    try {
      if (process.env.NODE_ENV === 'development') {
        console.log('Deleting product with ID:', productId);
      }

      const token = localStorage.getItem('access_token');
      if (!token) {
        console.error('No authentication token found');
        return false;
      }

      try {
        await api.delete(`/products/${productId}`, {
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json'
          }
        });

        removeOptimisticProduct(productId);

        if (process.env.NODE_ENV === 'development') {
          console.log('Product deleted successfully');
        }

        return true;
      } catch (error: unknown) {
        const err = error as { response?: { status?: number; data?: { error?: string } } };
        
        if (err.response) {
          if (err.response.status === 404) {
            console.log('Product not found (404), considering delete successful');
            removeOptimisticProduct(productId);
            return true;
          }

          if (err.response.status === 500 &&
              err.response.data?.error?.includes('product not found')) {
            console.log('Product already deleted, considering operation successful');
            removeOptimisticProduct(productId);
            return true;
          }
        }

        throw error;
      }
    } catch (error: unknown) {
      const err = error as { response?: { data?: unknown }; message?: string };
      console.error('Error deleting product:', err.response?.data || err.message || error);
      return false;
    } finally {
      setIsDeleting(false);
    }
  }, [removeOptimisticProduct, isDeleting]);

  const value = {
    refreshProducts,
    addOptimisticProduct,
    removeOptimisticProduct,
    deleteProduct,
    isRefreshing,
    isDeleting
  };

  return (
    <ProductContext.Provider value={value}>
      {children}
    </ProductContext.Provider>
  );
}

export function useProductContext() {
  const context = useContext(ProductContext);
  if (context === undefined) {
    throw new Error('useProductContext must be used within a ProductProvider');
  }
  return context;
}
