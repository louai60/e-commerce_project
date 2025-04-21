"use client";
import {
  Table,
  TableBody,
  TableCell,
  TableHeader,
  TableRow,
} from "../ui/table";
import Badge from "../ui/badge/Badge";
import Image from "next/image";
import { useAdminProducts } from "@/hooks/useAdminProducts";
import { formatPrice } from "@/lib/utils";
import LoadingSpinner from "../ui/loading/LoadingSpinner";
import { useState } from "react";

export default function RecentOrders() {
  const { products, isLoading, isError } = useAdminProducts(1, 5);
  const [selectedView, setSelectedView] = useState<"list" | "grid">("list");

  if (isError) {
    return (
      <div className="rounded-xl border border-gray-200 bg-white p-6 dark:border-gray-700 dark:bg-gray-800">
        <div className="text-center text-red-500">
          Error loading products
        </div>
      </div>
    );
  }

  return (
    <div className="rounded-xl border border-gray-200 bg-white shadow-sm dark:border-gray-700 dark:bg-gray-800">
      <div className="flex flex-wrap items-center justify-between gap-3 p-6">
        <div>
          <h4 className="font-medium text-gray-800 text-xl dark:text-white">
            Recent Products
          </h4>
          <p className="text-sm text-gray-500 dark:text-gray-400">
            Latest products in your inventory
          </p>
        </div>
        <div className="flex items-center gap-3">
          <div className="flex items-center gap-2 rounded-lg bg-gray-100 p-0.5 dark:bg-gray-900">
            <button
              onClick={() => setSelectedView("list")}
              className={`rounded-md px-3 py-2 text-sm font-medium transition-colors ${
                selectedView === "list"
                  ? "bg-white text-gray-900 shadow-sm dark:bg-gray-800 dark:text-white"
                  : "text-gray-500 hover:text-gray-900 dark:text-gray-400 dark:hover:text-white"
              }`}
            >
              List View
            </button>
            <button
              onClick={() => setSelectedView("grid")}
              className={`rounded-md px-3 py-2 text-sm font-medium transition-colors ${
                selectedView === "grid"
                  ? "bg-white text-gray-900 shadow-sm dark:bg-gray-800 dark:text-white"
                  : "text-gray-500 hover:text-gray-900 dark:text-gray-400 dark:hover:text-white"
              }`}
            >
              Grid View
            </button>
          </div>
          <button className="inline-flex items-center gap-2 rounded-lg border border-gray-300 bg-white px-4 py-2.5 text-sm font-medium text-gray-700 shadow-sm hover:bg-gray-50 hover:text-gray-800 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-400 dark:hover:bg-white/[0.03] dark:hover:text-gray-200">
            See all
          </button>
        </div>
      </div>

      <div className="max-w-full overflow-x-auto">
        <Table variant="striped" size="md">
          <TableHeader>
            <TableRow>
              <TableCell isHeader align="left">
                Products
              </TableCell>
              <TableCell isHeader align="left">
                Category
              </TableCell>
              <TableCell isHeader align="right">
                Price
              </TableCell>
              <TableCell isHeader align="center">
                Status
              </TableCell>
            </TableRow>
          </TableHeader>

          <TableBody>
            {isLoading ? (
              <TableRow>
                <TableCell colSpan={4} align="center" className="py-8">
                  <LoadingSpinner />
                </TableCell>
              </TableRow>
            ) : (
              products.map((product) => (
                <TableRow key={product.id} hover>
                  <TableCell>
                    <div className="flex items-center gap-3">
                      <div className="h-[50px] w-[50px] overflow-hidden rounded-lg bg-gray-100 dark:bg-gray-900">
                        <Image
                          width={50}
                          height={50}
                          src={product.image || "/images/product/product-01.jpg"}
                          className="h-[50px] w-[50px] object-cover"
                          alt={product.name || product.title || "Product Image"}
                        />
                      </div>
                      <div>
                        <p className="font-medium text-gray-800 dark:text-white/90">
                          {product.name || product.title || "Unnamed Product"}
                        </p>
                        <span className="text-sm text-gray-500 dark:text-gray-400">
                          {(() => {
                            if (typeof product.variants === 'number') {
                              return product.variants;
                            } else if (Array.isArray(product.variants)) {
                              return product.variants.length;
                            } else if (product.variants && typeof product.variants === 'object') {
                              return Object.keys(product.variants).length;
                            } else {
                              return 0;
                            }
                          })()} Variants
                        </span>
                      </div>
                    </div>
                  </TableCell>
                  <TableCell>
                    <span className="inline-flex items-center rounded-full bg-gray-100 px-2.5 py-0.5 text-sm font-medium text-gray-800 dark:bg-gray-900 dark:text-gray-300">
                      {product.category_id}
                    </span>
                  </TableCell>
                  <TableCell align="right">
                    <span className="font-medium text-gray-800 dark:text-white">
                      {formatPrice(product.price)}
                    </span>
                  </TableCell>
                  <TableCell align="center">
                    <Badge
                      variant={product.status === "Active" ? "success" : "warning"}
                      color="primary"
                      size="sm"
                    >
                      {product.status}
                    </Badge>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}


