"use client";
import React, { createContext, useContext, ReactNode, useState, useCallback } from 'react';
// import { api } from '@/lib/api'; // Removed unused import
// import { toast } from 'react-hot-toast'; // Removed unused import

interface InventoryContextType {
  refreshInventory: () => void;
  isRefreshing: boolean;
  isDeleting: boolean;
}

const InventoryContext = createContext<InventoryContextType | undefined>(undefined);

export function InventoryProvider({ children }: { children: ReactNode }) {
  const [isRefreshing, setIsRefreshing] = useState(false);
  // Using isDeleting state but not the setter
  const [isDeleting] = useState(false);

  const refreshInventory = useCallback(() => {
    setIsRefreshing(true);
    // Implementation here
    setIsRefreshing(false);
  }, []);

  const value = {
    refreshInventory,
    isRefreshing,
    isDeleting
  };

  return (
    <InventoryContext.Provider value={value}>
      {children}
    </InventoryContext.Provider>
  );
}

export function useInventoryContext() {
  const context = useContext(InventoryContext);
  if (context === undefined) {
    throw new Error('useInventoryContext must be used within an InventoryProvider');
  }
  return context;
}
