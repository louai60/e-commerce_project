"use client";
import React, { createContext, useContext, ReactNode, useState, useCallback } from 'react';
import { useProducts } from '@/hooks/useProducts';
import { Product, ProductService } from '@/services/product.service';
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
  // Use default parameters to avoid unnecessary re-renders
  const { mutate } = useProducts(1, 10, {});

  // Function to refresh the product list with improved error handling
  const refreshProducts = useCallback(() => {
    setIsRefreshing(true);

    // Only log in development mode
    if (process.env.NODE_ENV === 'development') {
      console.log('Refreshing products...');
    }

    // Use the mutate function from SWR to revalidate the data
    // The revalidate: true parameter forces a revalidation even if the cache is fresh
    mutate(undefined, { revalidate: true })
      .then((data) => {
        // Only log in development mode
        if (process.env.NODE_ENV === 'development') {
          console.log('Products refreshed successfully:', data?.products?.length || 0, 'products loaded');
        }
      })
      .catch((error) => {
        console.error('Error refreshing products:', error);
      })
      .finally(() => {
        setIsRefreshing(false);
      });
  }, [mutate]);

  // Function to optimistically add a product to the list with improved handling
  const addOptimisticProduct = useCallback((newProduct: Product) => {
    // Only log in development mode
    if (process.env.NODE_ENV === 'development') {
      console.log('Adding optimistic product:', newProduct.title);
    }

    mutate(
      (currentData) => {
        if (!currentData) {
          // Only log in development mode
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

        // Check if product already exists to avoid duplicates
        const productExists = currentData.products.some(p => p.id === newProduct.id);
        if (productExists) {
          // Only log in development mode
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

        // Only log in development mode
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
      { revalidate: false } // Don't revalidate immediately
    );
  }, [mutate]);

  // Function to remove a product from the list (for optimistic UI updates)
  const removeOptimisticProduct = useCallback((productId: string) => {
    // Only log in development mode
    if (process.env.NODE_ENV === 'development') {
      console.log('Removing product with ID:', productId);
    }

    mutate(
      (currentData) => {
        if (!currentData) {
          // Only log in development mode
          if (process.env.NODE_ENV === 'development') {
            console.log('No current data available for removal');
          }
          return currentData;
        }

        // Check if product exists before removal
        const productExists = currentData.products.some(p => p.id === productId);
        if (!productExists) {
          // Only log in development mode
          if (process.env.NODE_ENV === 'development') {
            console.log('Product not found in the list, nothing to remove');
          }
          return currentData;
        }

        // Only log in development mode
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
      { revalidate: false } // Don't revalidate immediately
    );
  }, [mutate]);

  // Function to delete a product with proper error handling
  const deleteProduct = useCallback(async (productId: string): Promise<boolean> => {
    if (!productId) return false;

    // Check if we're already deleting - prevent duplicate calls
    if (isDeleting) {
      console.warn('Delete operation already in progress');
      return false;
    }

    setIsDeleting(true);

    try {
      // Only log in development mode
      if (process.env.NODE_ENV === 'development') {
        console.log('Deleting product with ID:', productId);
      }

      // Get the latest token from localStorage
      const token = localStorage.getItem('access_token');
      if (!token) {
        console.error('No authentication token found');
        return false;
      }

      try {
        // Call the API to delete the product with explicit auth header
        await api.delete(`/products/${productId}`, {
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json'
          }
        });

        // Remove the product from the list optimistically
        removeOptimisticProduct(productId);

        // Only log in development mode
        if (process.env.NODE_ENV === 'development') {
          console.log('Product deleted successfully');
        }

        return true;
      } catch (error: any) {
        // Handle specific error cases
        if (error.response) {
          // If the product is not found (404), still consider it a success
          // as the end result is the same - the product is not in the system
          if (error.response.status === 404) {
            console.log('Product not found (404), considering delete successful');
            removeOptimisticProduct(productId);
            return true;
          }

          // If we get a 500 with "product not found" message, it's also a success case
          // This happens when the product was already deleted
          if (error.response.status === 500 &&
              error.response.data?.error?.includes('product not found')) {
            console.log('Product already deleted, considering operation successful');
            removeOptimisticProduct(productId);
            return true;
          }
        }

        throw error; // Re-throw for the outer catch block
      }
    } catch (error: any) {
      console.error('Error deleting product:', error.response?.data || error.message || error);
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
