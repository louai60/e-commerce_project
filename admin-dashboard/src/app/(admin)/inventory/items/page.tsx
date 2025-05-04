"use client";
import React, { useState, useEffect } from "react";
import Image from "next/image"; // Import next/image
import PageBreadcrumb from "@/components/common/PageBreadCrumb";
// Use GraphQL hooks for inventory items
import { useWarehouses } from "@/hooks/useInventory";
import { useInventoryItemsGraphQL, InventoryItem } from "@/hooks/useInventoryGraphQL";
import { useInventoryContext } from "@/contexts/InventoryContext";
import {
  Table,
  TableBody,
  TableCell,
  TableHeader,
  TableRow
} from "@/components/ui/table";
import Badge from "@/components/ui/badge/Badge";
import Button from "@/components/ui/button/Button";
import { PlusIcon, ArrowRightIcon } from "@/icons";
import Link from "next/link";
import LoadingSpinner from "@/components/ui/loading/LoadingSpinner";
import Pagination from "@/components/ui/pagination";
import { useModal } from "@/hooks/useModal";
import { Card, CardContent } from "@/components/ui/card/Card";

export default function InventoryItemsPage() {
  const [page, setPage] = useState(1);
  const [limit] = useState(10);
  const [filters, setFilters] = useState({
    status: '',
    warehouse_id: '',
    low_stock_only: false
  });

  // Use GraphQL hooks instead of REST hooks
  const { items, pagination, isLoading, isError, refetch } = useInventoryItemsGraphQL(page, limit, {
    lowStockOnly: filters.low_stock_only
  });
  const { warehouses } = useWarehouses(1, 100);
  const { isRefreshing } = useInventoryContext();

  // Debug the items data
  useEffect(() => {
    if (items && items.length > 0) {
      console.log('Inventory items from GraphQL:', items);
      // Check if product data is available
      const hasProductData = items.some((item: InventoryItem) => item.product && item.product.title);
      console.log('Has product data:', hasProductData);
      if (!hasProductData) {
        console.warn('Product data is missing in inventory items');
        // Log each item's product data
        items.forEach((item: InventoryItem, index: number) => {
          console.log(`Item ${index} (${item.sku}):`, item.product);
        });
      }
    }
  }, [items]);

  // State for details modal
  // const detailsModal = useModal();
  // const [selectedItem, setSelectedItem] = useState<InventoryItem | null>(null); // Commented out: Unused variable

  // Add a refresh button to the UI
  const handleRefresh = () => {
    refetch();
  };

  // Handle filter changes
  const handleFilterChange = (key: string, value: string | boolean) => {
    setFilters(prev => ({ ...prev, [key]: value }));
    setPage(1); // Reset to first page when filters change
  };

  // Handle item click for details - Commented out as selectedItem is unused
  // const handleItemClick = (item: InventoryItem) => {
  //   setSelectedItem(item);
  //   detailsModal.openModal();
  // };

  return (
    <div>
      <div className="mb-6 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <PageBreadcrumb pageTitle="Inventory Items" />
        <div className="flex gap-3">
          <Button
            variant="outline"
            onClick={handleRefresh}
            disabled={isLoading || isRefreshing}
          >
            {isRefreshing ? "Refreshing..." : "Refresh"}
          </Button>
          <Link href="/inventory/items/add">
            <Button variant="primary" startIcon={<PlusIcon />}>
              Add Inventory
            </Button>
          </Link>
        </div>
      </div>

      {/* Filters */}
      <Card className="mb-6">
        <CardContent className="p-4">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div>
              <label className="block text-sm font-medium mb-1">Status</label>
              <select
                className="w-full p-2 border rounded-md"
                value={filters.status}
                onChange={(e) => handleFilterChange('status', e.target.value)}
              >
                <option value="">All Statuses</option>
                <option value="IN_STOCK">In Stock</option>
                <option value="LOW_STOCK">Low Stock</option>
                <option value="OUT_OF_STOCK">Out of Stock</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">Warehouse</label>
              <select
                className="w-full p-2 border rounded-md"
                value={filters.warehouse_id}
                onChange={(e) => handleFilterChange('warehouse_id', e.target.value)}
              >
                <option value="">All Warehouses</option>
                {warehouses?.map((warehouse) => (
                  <option key={warehouse.id} value={warehouse.id}>
                    {warehouse.name}
                  </option>
                ))}
              </select>
            </div>
            <div className="flex items-end">
              <label className="flex items-center">
                <input
                  type="checkbox"
                  className="mr-2"
                  checked={filters.low_stock_only}
                  onChange={(e) => handleFilterChange('low_stock_only', e.target.checked)}
                />
                <span className="text-sm font-medium">Low Stock Only</span>
              </label>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Inventory Items Table */}
      <Card>
        <CardContent className="p-0">
          {isLoading ? (
            <div className="flex justify-center items-center h-64">
              <LoadingSpinner />
            </div>
          ) : isError ? (
            <div className="text-center py-8">
              <p className="text-red-500">Error loading inventory items</p>
              <Button variant="outline" className="mt-4" onClick={handleRefresh}>
                Try Again
              </Button>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableCell isHeader className="min-w-[150px]">Product</TableCell>
                    <TableCell isHeader>Product Name</TableCell>
                    <TableCell isHeader>SKU</TableCell>
                    <TableCell isHeader>Total Qty</TableCell>
                    <TableCell isHeader>Available</TableCell>
                    <TableCell isHeader>Reserved</TableCell>
                    <TableCell isHeader>Reorder Point</TableCell>
                    <TableCell isHeader>Status</TableCell>
                    <TableCell isHeader className="text-right">Actions</TableCell>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {Array.isArray(items) && items.length > 0 ? (
                    items.map((item) => (
                      <TableRow
                        key={item?.id || Math.random().toString()}
                        className="hover:bg-gray-50"
                      >
                        <TableCell>
                          <div
                            className="flex items-center gap-3" // Removed cursor-pointer and onClick
                            // onClick={() => handleItemClick(item)} // Commented out: handleItemClick is unused
                          >
                            {item.product?.images && item.product.images.length > 0 ? (
                              <div className="h-10 w-10 rounded-md overflow-hidden bg-gray-100 relative"> {/* Added relative positioning */}
                                <Image // Use next/image
                                  src={item.product.images[0].url}
                                  alt={item.product.title || 'Product Image'} // Added fallback alt text
                                  fill // Use fill to cover the container
                                  style={{ objectFit: 'cover' }} // Maintain object-cover behavior
                                  sizes="(max-width: 768px) 10vw, 5vw" // Provide sizes prop for optimization
                                />
                              </div>
                            ) : (
                              <div className="h-10 w-10 rounded-md bg-gray-100 flex items-center justify-center text-gray-500">
                                No img
                              </div>
                            )}
                          </div>
                        </TableCell>
                        <TableCell>
                          <div /* Removed cursor-pointer and onClick */ /* onClick={() => handleItemClick(item)} */>
                            <span className="font-medium">
                              {item.product?.title || 'Unknown Product'}
                            </span>
                          </div>
                        </TableCell>
                        <TableCell>
                          <div /* Removed cursor-pointer and onClick */ /* onClick={() => handleItemClick(item)} */>
                            {item.sku}
                          </div>
                        </TableCell>
                        <TableCell>
                          <div /* Removed cursor-pointer and onClick */ /* onClick={() => handleItemClick(item)} */>
                            {item.total_quantity}
                          </div>
                        </TableCell>
                        <TableCell>
                          <div /* Removed cursor-pointer and onClick */ /* onClick={() => handleItemClick(item)} */>
                            {item.available_quantity}
                          </div>
                        </TableCell>
                        <TableCell>
                          <div /* Removed cursor-pointer and onClick */ /* onClick={() => handleItemClick(item)} */>
                            {item.reserved_quantity}
                          </div>
                        </TableCell>
                        <TableCell>
                          <div /* Removed cursor-pointer and onClick */ /* onClick={() => handleItemClick(item)} */>
                            {item.reorder_point}
                          </div>
                        </TableCell>
                        <TableCell>
                          <div /* Removed cursor-pointer and onClick */ /* onClick={() => handleItemClick(item)} */>
                            <Badge
                              variant={
                                item.status === 'IN_STOCK'
                                  ? 'success'
                                  : item.status === 'LOW_STOCK'
                                  ? 'warning'
                                  : 'danger'
                              }
                              className={`px-3 py-1 rounded-full text-xs ${
                                item.status === 'IN_STOCK'
                                  ? 'bg-green-100 text-green-800'
                                  : item.status === 'LOW_STOCK'
                                  ? 'bg-yellow-100 text-yellow-800'
                                  : 'bg-red-100 text-red-800'
                              }`}
                            >
                              {item.status}
                            </Badge>
                          </div>
                        </TableCell>
                        <TableCell>
                          <div className="flex items-center justify-end gap-2">
                            <Link href={`/inventory/items/${item?.id}`}>
                              <Button
                                variant="outline"
                                size="sm"
                                className="h-9 w-9 p-0"
                                onClick={(e) => { if (e) e.stopPropagation(); }}
                              >
                                <ArrowRightIcon className="h-4 w-4" />
                              </Button>
                            </Link>
                          </div>
                        </TableCell>
                      </TableRow>
                    ))
                  ) : (
                    <TableRow>
                      <TableCell colSpan={9} className="text-center py-8">
                        No inventory items found
                      </TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
            </div>
          )}

          {/* Pagination */}
          {pagination && pagination.total_pages > 1 && (
            <div className="py-4 px-6">
              <Pagination
                currentPage={page}
                totalPages={pagination.total_pages}
                onPageChange={setPage}
              />
            </div>
          )}
        </CardContent>
      </Card>

      {/* TODO: Add inventory item details modal */}
      {/* This would be implemented as a separate component */}
    </div>
  );
}
