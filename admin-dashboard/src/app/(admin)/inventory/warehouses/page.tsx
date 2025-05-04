"use client";
import React, { useState } from "react";
import PageBreadcrumb from "@/components/common/PageBreadCrumb";
import { useWarehouses } from "@/hooks/useInventory";
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
import { PlusIcon, PencilIcon, ArrowRightIcon } from "@/icons";
import Link from "next/link";
import LoadingSpinner from "@/components/ui/loading/LoadingSpinner";
import Pagination from "@/components/ui/pagination";
// import { useModal } from "@/hooks/useModal"; // Removed unused import
import { Card, CardContent } from "@/components/ui/card/Card";
// import { Warehouse } from "@/services/inventory.service"; // Removed unused import

export default function WarehousesPage() {
  const [page, setPage] = useState(1);
  const [limit] = useState(10);
  const [filters, setFilters] = useState({
    is_active: ''
  });

  const { warehouses, pagination, isLoading, isError, mutate } = useWarehouses(page, limit, filters);
  const { isRefreshing } = useInventoryContext();

  // State for details modal
  // const detailsModal = useModal();
  // const [selectedWarehouse, setSelectedWarehouse] = useState<Warehouse | null>(null); // Commented out: Unused variable

  // Add a refresh button to the UI
  const handleRefresh = () => {
    mutate();
  };

  // Handle filter changes
  const handleFilterChange = (key: string, value: string | boolean) => {
    setFilters(prev => ({ ...prev, [key]: value }));
    setPage(1); // Reset to first page when filters change
  };

  // Handle warehouse click for details - Commented out as selectedWarehouse is unused
  // const handleWarehouseClick = (warehouse: Warehouse) => {
  //   setSelectedWarehouse(warehouse);
  //   detailsModal.openModal();
  // };

  return (
    <div>
      <div className="mb-6 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <PageBreadcrumb pageTitle="Warehouses" />
        <div className="flex gap-3">
          <Button
            variant="outline"
            onClick={handleRefresh}
            disabled={isLoading || isRefreshing}
          >
            {isRefreshing ? "Refreshing..." : "Refresh"}
          </Button>
          <Link href="/inventory/warehouses/add">
            <Button variant="primary" startIcon={<PlusIcon />}>
              Add Warehouse
            </Button>
          </Link>
        </div>
      </div>

      {/* Filters */}
      <Card className="mb-6">
        <CardContent className="p-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium mb-1">Status</label>
              <select
                className="w-full p-2 border rounded-md"
                value={filters.is_active === '' ? '' : filters.is_active === 'true' ? 'true' : 'false'}
                onChange={(e) => {
                  const value = e.target.value;
                  handleFilterChange('is_active', value === '' ? '' : value === 'true');
                }}
              >
                <option value="">All Statuses</option>
                <option value="true">Active</option>
                <option value="false">Inactive</option>
              </select>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Warehouses Table */}
      <Card>
        <CardContent className="p-0">
          {isLoading ? (
            <div className="flex justify-center items-center h-64">
              <LoadingSpinner />
            </div>
          ) : isError ? (
            <div className="text-center py-8">
              <p className="text-red-500">Error loading warehouses</p>
              <Button variant="outline" className="mt-4" onClick={handleRefresh}>
                Try Again
              </Button>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableCell isHeader>Name</TableCell>
                    <TableCell isHeader>Code</TableCell>
                    <TableCell isHeader>Location</TableCell>
                    <TableCell isHeader>Items</TableCell>
                    <TableCell isHeader>Priority</TableCell>
                    <TableCell isHeader>Status</TableCell>
                    <TableCell isHeader className="text-right">Actions</TableCell>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {Array.isArray(warehouses) && warehouses.length > 0 ? (
                    warehouses.map((warehouse) => (
                      <TableRow
                        key={warehouse?.id || Math.random().toString()}
                        className="hover:bg-gray-50"
                      >
                        <TableCell>
                          <div
                            className="font-medium" // Removed cursor-pointer and onClick
                            // onClick={() => handleWarehouseClick(warehouse)}
                          >
                            {warehouse.name}
                          </div>
                        </TableCell>
                        <TableCell>
                          <div /* Removed cursor-pointer and onClick */ /* onClick={() => handleWarehouseClick(warehouse)} */>
                            {warehouse.code}
                          </div>
                        </TableCell>
                        <TableCell>
                          <div /* Removed cursor-pointer and onClick */ /* onClick={() => handleWarehouseClick(warehouse)} */>
                            {[
                              warehouse.city,
                              warehouse.state,
                              warehouse.country
                            ]
                              .filter(Boolean)
                              .join(', ')}
                          </div>
                        </TableCell>
                        <TableCell>
                          <div /* Removed cursor-pointer and onClick */ /* onClick={() => handleWarehouseClick(warehouse)} */>
                            {warehouse.item_count || 0}
                          </div>
                        </TableCell>
                        <TableCell>
                          <div /* Removed cursor-pointer and onClick */ /* onClick={() => handleWarehouseClick(warehouse)} */>
                            {warehouse.priority}
                          </div>
                        </TableCell>
                        <TableCell>
                          <div /* Removed cursor-pointer and onClick */ /* onClick={() => handleWarehouseClick(warehouse)} */>
                            <Badge
                              variant={warehouse.is_active ? 'success' : 'danger'}
                              className={`px-3 py-1 rounded-full text-xs ${
                                warehouse.is_active
                                  ? 'bg-green-100 text-green-800'
                                  : 'bg-red-100 text-red-800'
                              }`}
                            >
                              {warehouse.is_active ? 'Active' : 'Inactive'}
                            </Badge>
                          </div>
                        </TableCell>
                        <TableCell>
                          <div className="flex items-center justify-end gap-2">
                            <Link href={`/inventory/warehouses/edit/${warehouse?.id}`}>
                              <Button
                                variant="outline"
                                size="sm"
                                className="h-9 w-9 p-0"
                                onClick={(e) => { if (e) e.stopPropagation(); }}
                              >
                                <PencilIcon className="h-4 w-4" />
                              </Button>
                            </Link>
                            <Link href={`/inventory/warehouses/${warehouse?.id}`}>
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
                      <TableCell colSpan={7} className="text-center py-8">
                        No warehouses found
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

      {/* TODO: Add warehouse details modal */}
      {/* This would be implemented as a separate component */}
    </div>
  );
}
