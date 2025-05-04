"use client";
import React, { useState, useEffect } from "react";
import PageBreadcrumb from "@/components/common/PageBreadCrumb";
// Use GraphQL hooks instead of REST hooks
import { useInventoryTransactionsGraphQL } from "@/hooks/useInventoryGraphQL";
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
import Link from "next/link";
import LoadingSpinner from "@/components/ui/loading/LoadingSpinner";
import Pagination from "@/components/ui/pagination";
import { Card, CardContent } from "@/components/ui/card/Card";
import { formatDate } from "@/lib/utils";

export default function TransactionsPage() {
  const [page, setPage] = useState(1);
  const [limit] = useState(20);
  const [filters, setFilters] = useState({
    transactionType: '',
    warehouseId: '',
    dateFrom: '',
    dateTo: ''
  });

  // Use GraphQL hook instead of REST hook
  const { transactions, pagination, isLoading, isError, refetch } = useInventoryTransactionsGraphQL(page, limit, filters);
  const { warehouses } = useWarehouses(1, 100);
  const { isRefreshing } = useInventoryContext();

  // Debug the transactions data
  useEffect(() => {
    if (transactions && transactions.length > 0) {
      console.log('Inventory transactions from GraphQL:', transactions);
    }
  }, [transactions]);

  // Add a refresh button to the UI
  const handleRefresh = () => {
    refetch();
  };

  // Handle filter changes
  const handleFilterChange = (key: string, value: string) => {
    // Map old filter keys to new filter keys
    const keyMap: Record<string, string> = {
      'transaction_type': 'transactionType',
      'warehouse_id': 'warehouseId',
      'date_from': 'dateFrom',
      'date_to': 'dateTo'
    };

    // Use the mapped key or the original key if not in the map
    const mappedKey = keyMap[key] || key;

    setFilters(prev => ({ ...prev, [mappedKey]: value }));
    setPage(1); // Reset to first page when filters change
  };

  // Get transaction type badge color
  const getTransactionTypeColor = (type: string) => {
    switch (type) {
      case 'STOCK_ADDITION':
        return 'bg-green-100 text-green-800';
      case 'STOCK_REMOVAL':
        return 'bg-red-100 text-red-800';
      case 'TRANSFER':
        return 'bg-blue-100 text-blue-800';
      case 'RESERVATION':
        return 'bg-purple-100 text-purple-800';
      case 'ADJUSTMENT':
        return 'bg-yellow-100 text-yellow-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  // Format transaction type for display
  const formatTransactionType = (type: string) => {
    return type.replace(/_/g, ' ');
  };

  return (
    <div>
      <div className="mb-6 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <PageBreadcrumb pageTitle="Inventory Transactions" />
        <div className="flex gap-3">
          <Button
            variant="outline"
            onClick={handleRefresh}
            disabled={isLoading || isRefreshing}
          >
            {isRefreshing ? "Refreshing..." : "Refresh"}
          </Button>
        </div>
      </div>

      {/* Filters */}
      <Card className="mb-6">
        <CardContent className="p-4">
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
            <div>
              <label className="block text-sm font-medium mb-1">Transaction Type</label>
              <select
                className="w-full p-2 border rounded-md"
                value={filters.transactionType}
                onChange={(e) => handleFilterChange('transaction_type', e.target.value)}
              >
                <option value="">All Types</option>
                <option value="STOCK_ADDITION">Stock Addition</option>
                <option value="STOCK_REMOVAL">Stock Removal</option>
                <option value="TRANSFER">Transfer</option>
                <option value="RESERVATION">Reservation</option>
                <option value="ADJUSTMENT">Adjustment</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">Warehouse</label>
              <select
                className="w-full p-2 border rounded-md"
                value={filters.warehouseId}
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
            <div>
              <label className="block text-sm font-medium mb-1">From Date</label>
              <input
                type="date"
                className="w-full p-2 border rounded-md"
                value={filters.dateFrom}
                onChange={(e) => handleFilterChange('date_from', e.target.value)}
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">To Date</label>
              <input
                type="date"
                className="w-full p-2 border rounded-md"
                value={filters.dateTo}
                onChange={(e) => handleFilterChange('date_to', e.target.value)}
              />
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Transactions Table */}
      <Card>
        <CardContent className="p-0">
          {isLoading ? (
            <div className="flex justify-center items-center h-64">
              <LoadingSpinner />
            </div>
          ) : isError ? (
            <div className="text-center py-8">
              <p className="text-red-500">Error loading transactions</p>
              <Button variant="outline" className="mt-4" onClick={handleRefresh}>
                Try Again
              </Button>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableCell isHeader>Date</TableCell>
                    <TableCell isHeader>Product</TableCell>
                    <TableCell isHeader>SKU</TableCell>
                    <TableCell isHeader>Type</TableCell>
                    <TableCell isHeader>Quantity</TableCell>
                    <TableCell isHeader>Warehouse</TableCell>
                    <TableCell isHeader>Reference</TableCell>
                    <TableCell isHeader>Notes</TableCell>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {Array.isArray(transactions) && transactions.length > 0 ? (
                    transactions.map((transaction) => (
                      <TableRow
                        key={transaction?.id || Math.random().toString()}
                        className="hover:bg-gray-50"
                      >
                        <TableCell>
                          {formatDate(transaction.created_at)}
                        </TableCell>
                        <TableCell>
                          <Link
                            href={`/products/${transaction.inventory_item?.product?.id}`}
                            className="text-blue-600 hover:underline"
                          >
                            {transaction.inventory_item?.product?.title || 'Unknown Product'}
                          </Link>
                        </TableCell>
                        <TableCell>
                          <Link
                            href={`/inventory/items/${transaction.inventory_item?.id}`}
                            className="text-blue-600 hover:underline"
                          >
                            {transaction.inventory_item?.sku || 'Unknown SKU'}
                          </Link>
                        </TableCell>
                        <TableCell>
                          <Badge
                            variant="light"
                            className={`px-3 py-1 rounded-full text-xs ${getTransactionTypeColor(transaction.transaction_type)}`}
                          >
                            {formatTransactionType(transaction.transaction_type)}
                          </Badge>
                        </TableCell>
                        <TableCell>
                          <span className={transaction.quantity > 0 ? 'text-green-600' : 'text-red-600'}>
                            {transaction.quantity > 0 ? '+' : ''}{transaction.quantity}
                          </span>
                        </TableCell>
                        <TableCell>
                          {transaction.warehouse?.name || 'N/A'}
                        </TableCell>
                        <TableCell>
                          {transaction.reference_id ? (
                            <span className="text-sm">
                              {transaction.reference_type}: {transaction.reference_id.substring(0, 8)}...
                            </span>
                          ) : (
                            'N/A'
                          )}
                        </TableCell>
                        <TableCell>
                          <span className="text-sm text-gray-600 truncate max-w-[200px] block">
                            {transaction.notes || 'No notes'}
                          </span>
                        </TableCell>
                      </TableRow>
                    ))
                  ) : (
                    <TableRow>
                      <TableCell colSpan={8} className="text-center py-8">
                        No transactions found
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
    </div>
  );
}
