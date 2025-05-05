/**
 * Utility functions for handling product and inventory statuses
 */

/**
 * Normalizes inventory status strings to a consistent format
 * @param status The raw status string from the API
 * @returns A normalized status string in uppercase format
 */
export const normalizeInventoryStatus = (status: string | undefined): string => {
  if (!status) return 'UNKNOWN';
  
  // Convert to uppercase for consistent comparison
  const upperStatus = status.toUpperCase();
  
  // Map common variations to standard formats
  if (upperStatus === 'IN_STOCK' || upperStatus === 'INSTOCK') {
    return 'IN_STOCK';
  } else if (upperStatus === 'LOW_STOCK' || upperStatus === 'LOWSTOCK') {
    return 'LOW_STOCK';
  } else if (upperStatus === 'OUT_OF_STOCK' || upperStatus === 'OUTOFSTOCK') {
    return 'OUT_OF_STOCK';
  }
  
  return upperStatus;
};

/**
 * Gets a user-friendly display text for inventory status
 * @param status The normalized status string
 * @returns A user-friendly display text
 */
export const getInventoryStatusDisplay = (status: string | undefined): string => {
  const normalizedStatus = normalizeInventoryStatus(status);
  
  switch (normalizedStatus) {
    case 'IN_STOCK':
      return 'In Stock';
    case 'LOW_STOCK':
      return 'Low Stock';
    case 'OUT_OF_STOCK':
      return 'Out of Stock';
    case 'DISCONTINUED':
      return 'Discontinued';
    case 'BACKORDERED':
      return 'Backordered';
    default:
      return status || 'Unknown';
  }
};

/**
 * Gets the appropriate variant for a Badge component based on inventory status
 * @param status The inventory status
 * @returns The appropriate variant for the Badge component
 */
export const getInventoryStatusVariant = (status: string | undefined): 'success' | 'warning' | 'danger' => {
  const normalizedStatus = normalizeInventoryStatus(status);
  
  switch (normalizedStatus) {
    case 'IN_STOCK':
      return 'success';
    case 'LOW_STOCK':
      return 'warning';
    case 'OUT_OF_STOCK':
    case 'DISCONTINUED':
      return 'danger';
    case 'BACKORDERED':
      return 'warning';
    default:
      return 'warning';
  }
};

/**
 * Gets the appropriate CSS classes for an inventory status badge
 * @param status The inventory status
 * @returns CSS classes for styling the badge
 */
export const getInventoryStatusClasses = (status: string | undefined): string => {
  const normalizedStatus = normalizeInventoryStatus(status);
  
  switch (normalizedStatus) {
    case 'IN_STOCK':
      return 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400';
    case 'LOW_STOCK':
    case 'BACKORDERED':
      return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400';
    case 'OUT_OF_STOCK':
    case 'DISCONTINUED':
      return 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400';
    default:
      return 'bg-gray-100 text-gray-800 dark:bg-gray-900/30 dark:text-gray-400';
  }
};
