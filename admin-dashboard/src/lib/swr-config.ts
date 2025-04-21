import { SWRConfiguration } from 'swr';

/**
 * Global SWR configuration for the admin dashboard
 *
 * - revalidateOnFocus: false - Don't revalidate when window gets focus
 * - revalidateOnReconnect: true - Revalidate when browser regains connection
 * - dedupingInterval: 2000 - Deduplicate requests within 2 seconds
 * - errorRetryCount: 3 - Retry failed requests 3 times
 * - revalidateIfStale: true - Revalidate if data is stale
 * - keepPreviousData: true - Keep previous data while fetching new data
 * - refreshInterval: 0 - Don't auto-refresh (we use manual refresh)
 */
export const swrConfig: SWRConfiguration = {
  revalidateOnFocus: false,
  revalidateOnReconnect: true,
  dedupingInterval: 2000, // Reduced from 5000ms to 2000ms
  errorRetryCount: 3,
  revalidateIfStale: true,
  keepPreviousData: true,
  refreshInterval: 0, // No auto-refresh
};
