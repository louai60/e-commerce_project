"use client";
import React from "react";
import PageBreadcrumb from "@/components/common/PageBreadCrumb";
import { useInventoryItemsGraphQL, useWarehousesGraphQL, InventoryItem, Warehouse } from "@/hooks/useInventoryGraphQL";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card/Card";
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell, Legend } from 'recharts';
import LoadingSpinner from "@/components/ui/loading/LoadingSpinner";
import Link from "next/link";
import Button from "@/components/ui/button/Button";
import { PlusIcon, ArrowRightIcon } from "@/icons";

export default function InventoryDashboard() {
  // Use GraphQL hooks instead of REST hooks
  const { items, isLoading: isLoadingItems } = useInventoryItemsGraphQL(1, 5, { lowStockOnly: true });
  const { warehouses, isLoading: isLoadingWarehouses } = useWarehousesGraphQL(1, 10);

  // Calculate inventory metrics
  const totalItems = items?.length || 0;
  const lowStockItems = items?.filter((item: InventoryItem) => item.status === "LOW_STOCK")?.length || 0;
  const outOfStockItems = items?.filter((item: InventoryItem) => item.status === "OUT_OF_STOCK")?.length || 0;

  // Prepare data for status distribution chart
  const statusData = [
    { name: 'In Stock', value: totalItems - lowStockItems - outOfStockItems },
    { name: 'Low Stock', value: lowStockItems },
    { name: 'Out of Stock', value: outOfStockItems },
  ];

  const COLORS = ['#4ade80', '#facc15', '#f87171'];

  // Prepare data for warehouse distribution chart
  const warehouseData = warehouses?.map((warehouse: Warehouse) => ({
    name: warehouse.name,
    items: warehouse.item_count || 0,
  })) || [];

  return (
    <div>
      <div className="mb-6 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <PageBreadcrumb pageTitle="Inventory Dashboard" />
        <div className="flex gap-3">
          <Link href="/inventory/items">
            <Button variant="outline" endIcon={<ArrowRightIcon />}>
              View All Items
            </Button>
          </Link>
          <Link href="/inventory/warehouses">
            <Button variant="primary" startIcon={<PlusIcon />}>
              Add Warehouse
            </Button>
          </Link>
        </div>
      </div>

      {/* Inventory Metrics */}
      <div className="grid grid-cols-1 gap-4 md:grid-cols-3 mb-6">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Total Inventory Items</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{totalItems}</div>
            <p className="text-xs text-muted-foreground">
              Across all warehouses
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Low Stock Items</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{lowStockItems}</div>
            <p className="text-xs text-muted-foreground">
              Items below reorder point
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Out of Stock Items</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{outOfStockItems}</div>
            <p className="text-xs text-muted-foreground">
              Items with zero quantity
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Charts */}
      <div className="grid grid-cols-1 gap-4 md:grid-cols-2 mb-6">
        <Card>
          <CardHeader>
            <CardTitle>Inventory Status Distribution</CardTitle>
          </CardHeader>
          <CardContent>
            {isLoadingItems ? (
              <div className="flex justify-center items-center h-64">
                <LoadingSpinner />
              </div>
            ) : (
              <ResponsiveContainer width="100%" height={300}>
                <PieChart>
                  <Pie
                    data={statusData}
                    cx="50%"
                    cy="50%"
                    labelLine={false}
                    outerRadius={80}
                    fill="#8884d8"
                    dataKey="value"
                    label={({ name, percent }: { name: string; percent: number }) => `${name}: ${(percent * 100).toFixed(0)}%`}
                  >
                    {statusData.map((_, index) => (
                      <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                    ))}
                  </Pie>
                  <Tooltip />
                  <Legend />
                </PieChart>
              </ResponsiveContainer>
            )}
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Warehouse Distribution</CardTitle>
          </CardHeader>
          <CardContent>
            {isLoadingWarehouses ? (
              <div className="flex justify-center items-center h-64">
                <LoadingSpinner />
              </div>
            ) : (
              <ResponsiveContainer width="100%" height={300}>
                <BarChart
                  data={warehouseData}
                  margin={{
                    top: 5,
                    right: 30,
                    left: 20,
                    bottom: 5,
                  }}
                >
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="name" />
                  <YAxis />
                  <Tooltip />
                  <Bar dataKey="items" fill="#3b82f6" />
                </BarChart>
              </ResponsiveContainer>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Low Stock Items */}
      <Card>
        <CardHeader>
          <CardTitle>Low Stock Items</CardTitle>
        </CardHeader>
        <CardContent>
          {isLoadingItems ? (
            <div className="flex justify-center items-center h-32">
              <LoadingSpinner />
            </div>
          ) : items && items.length > 0 ? (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b">
                    <th className="text-left py-3 px-4">Product</th>
                    <th className="text-left py-3 px-4">SKU</th>
                    <th className="text-left py-3 px-4">Available</th>
                    <th className="text-left py-3 px-4">Reorder Point</th>
                    <th className="text-left py-3 px-4">Status</th>
                    <th className="text-right py-3 px-4">Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {items.map((item: InventoryItem) => (
                    <tr key={item.id} className="border-b hover:bg-gray-50">
                      <td className="py-3 px-4">
                        {item.product?.title || 'Unknown Product'}
                      </td>
                      <td className="py-3 px-4">{item.sku}</td>
                      <td className="py-3 px-4">{item.available_quantity}</td>
                      <td className="py-3 px-4">{item.reorder_point}</td>
                      <td className="py-3 px-4">
                        <span className={`px-2 py-1 rounded-full text-xs ${
                          item.status === 'IN_STOCK'
                            ? 'bg-green-100 text-green-800'
                            : item.status === 'LOW_STOCK'
                            ? 'bg-yellow-100 text-yellow-800'
                            : 'bg-red-100 text-red-800'
                        }`}>
                          {item.status}
                        </span>
                      </td>
                      <td className="py-3 px-4 text-right">
                        <Link href={`/inventory/items/${item.id}`}>
                          <Button variant="outline" size="sm">
                            View
                          </Button>
                        </Link>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <div className="text-center py-8">
              <p className="text-gray-500">No low stock items found</p>
            </div>
          )}
          <div className="mt-4 text-right">
            <Link href="/inventory/items?status=LOW_STOCK">
              <Button variant="outline" endIcon={<ArrowRightIcon />}>
                View All Low Stock Items
              </Button>
            </Link>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
