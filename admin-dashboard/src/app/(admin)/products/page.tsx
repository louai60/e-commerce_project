"use client";
import React, { useState } from "react";
import { useProductContext } from "@/contexts/ProductContext";
import PageBreadcrumb from "@/components/common/PageBreadCrumb";
import { useProducts } from "@/hooks/useProducts";
import {
  Table,
  TableBody,
  TableCell,
  TableHeader,
  TableRow
} from "@/components/ui/table";
import Badge from "@/components/ui/badge/Badge";
import Button from "@/components/ui/button/Button";
import { PencilIcon, TrashBinIcon, PlusIcon } from "@/icons";
import Image from "next/image";
import Link from "next/link";
import { formatPrice } from "@/lib/utils";
import LoadingSpinner from "@/components/ui/loading/LoadingSpinner";
import Pagination from "@/components/ui/pagination";
import { useModal } from "@/hooks/useModal";
import DeleteProductModal from "@/components/products/DeleteProductModal";
import ProductDetailsModal from "@/components/products/ProductDetailsModal";
import { Product } from "@/services/product.service";

export default function ProductsPage() {
  const [page, setPage] = useState(1);
  const [limit] = useState(10);
  const { products, pagination, isLoading, isError, mutate } = useProducts(page, limit);
  const { isRefreshing, deleteProduct } = useProductContext();

  // State for delete modal
  const deleteModal = useModal();
  const [productToDelete, setProductToDelete] = useState<{ id: string; title: string } | null>(null);

  // State for details modal
  const detailsModal = useModal();
  const [selectedProduct, setSelectedProduct] = useState<Product | null>(null);

  // Add a refresh button to the UI
  const handleRefresh = () => {
    mutate();
  };

  // Handle opening the delete modal
  const handleDeleteClick = (product: Product) => {
    setProductToDelete({
      id: product.id,
      title: product.title || 'Untitled Product'
    });
    deleteModal.openModal();
  };

  // Handle opening the details modal
  const handleProductClick = (product: Product) => {
    // Log the product data to help diagnose issues
    console.log('Product clicked:', product);
    console.log('Product images:', product.images);
    console.log('Product price:', product.price);
    console.log('Product inventory:', product.inventory);

    setSelectedProduct(product);
    detailsModal.openModal();
  };

  // Handle the actual deletion
  const handleDeleteConfirm = async (productId: string): Promise<boolean> => {
    try {
      // First check if the product exists in our current list
      const productExists = products?.some(p => p.id === productId);
      if (!productExists) {
        console.warn('Product not found in current list, may have been already deleted');
        // Still refresh the list to ensure UI is up to date
        mutate();
        return true;
      }

      // Call the delete function from context
      const success = await deleteProduct(productId);

      if (success) {
        // No need for toast here as the modal will handle it
        // Refresh the product list after a short delay
        setTimeout(() => {
          mutate();
        }, 500);
        return true;
      } else {
        // No need for toast here as the modal will handle it
        return false;
      }
    } catch (error) {
      console.error('Error deleting product:', error);
      return false;
    }
  };

  // Handle page change
  const handlePageChange = (newPage: number) => {
    setPage(newPage);
  };

  // Only show full-page loading on initial load
  if (isLoading && !products) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <LoadingSpinner />
      </div>
    );
  }

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
    <div>
      <div className="mb-6 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <PageBreadcrumb pageTitle="Products" />
        <div className="flex gap-3">
          <Button
            variant="outline"
            onClick={handleRefresh}
            disabled={isLoading || isRefreshing}
          >
            {isRefreshing ? "Refreshing..." : "Refresh"}
          </Button>
          <Link href="/products/create">
            <Button variant="primary" startIcon={<PlusIcon />}>
              Add Product
            </Button>
          </Link>
        </div>
      </div>

      <div className="rounded-xl border border-gray-200 bg-white dark:border-gray-700 dark:bg-gray-800 relative">
        {/* Overlay loading indicator for refreshes */}
        {(isLoading || isRefreshing) && products && (
          <div className="absolute inset-0 bg-white/70 dark:bg-gray-800/70 flex items-center justify-center z-10">
            <LoadingSpinner />
          </div>
        )}
        <div className="p-4 md:p-6">
          <div className="overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableCell isHeader className="min-w-[250px]">Product</TableCell>
                  <TableCell isHeader>SKU</TableCell>
                  <TableCell isHeader>Price</TableCell>
                  <TableCell isHeader>Inventory</TableCell>
                  <TableCell isHeader>Status</TableCell>
                  <TableCell isHeader className="text-right">Actions</TableCell>
                </TableRow>
              </TableHeader>
              <TableBody>
                {Array.isArray(products) && products.length > 0 ? (
                  products.map((product) => (
                    <TableRow
                      key={product?.id || Math.random().toString()}
                    >
                      <TableCell>
                        <div
                          className="flex items-center gap-3 cursor-pointer"
                          onClick={() => handleProductClick(product)}
                        >
                          <div className="h-10 w-10 overflow-hidden rounded-lg border border-gray-200 dark:border-gray-700">
                            {product?.images && product.images.length > 0 && product.images[0]?.url ? (
                              <Image
                                src={product.images[0].url}
                                alt={product.title || 'Product image'}
                                width={40}
                                height={40}
                                className="h-full w-full object-cover"
                                unoptimized={true}
                              />
                            ) : (
                              <div className="flex h-full w-full items-center justify-center bg-gray-100 dark:bg-gray-800">
                                <span className="text-xs text-gray-400">No image</span>
                              </div>
                            )}
                          </div>
                          <div>
                            <h5 className="font-medium text-gray-900 dark:text-white">
                              {product?.title || 'Untitled Product'}
                            </h5>
                            <p className="text-xs text-gray-500 dark:text-gray-400">
                              {product?.short_description?.substring(0, 50) || "No description"}
                              {product?.short_description?.length > 50 ? "..." : ""}
                            </p>
                          </div>
                        </div>
                      </TableCell>
                      <TableCell>
                        <div className="cursor-pointer" onClick={() => handleProductClick(product)}>
                          {product?.sku || "N/A"}
                        </div>
                      </TableCell>
                      <TableCell>
                        <div className="cursor-pointer" onClick={() => handleProductClick(product)}>
                          {product?.price?.current?.TND ? formatPrice(product.price.current.TND) :
                           product?.price?.current?.USD ? formatPrice(product.price.current.USD) :
                           product?.price?.value ? formatPrice(product.price.value) : "N/A"}
                        </div>
                      </TableCell>
                      <TableCell>
                        <div className="cursor-pointer" onClick={() => handleProductClick(product)}>
                          {product?.inventory?.quantity || 0}
                        </div>
                      </TableCell>
                      <TableCell>
                        <div className="cursor-pointer" onClick={() => handleProductClick(product)}>
                          <Badge
                            variant="success"
                            className="bg-green-500 text-white px-3 py-1 rounded-full text-xs"
                          >
                            in_stock
                          </Badge>
                        </div>
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center justify-end gap-2">
                          <Link href={`/products/edit/${product?.id}`}>
                            <Button
                              variant="outline"
                              size="sm"
                              className="h-9 w-9 p-0"
                            >
                              <PencilIcon className="h-4 w-4" />
                            </Button>
                          </Link>
                          <Button
                            variant="outline"
                            size="sm"
                            className="h-9 w-9 p-0 text-danger-500 hover:border-danger-500 hover:bg-danger-500/10"
                            onClick={() => handleDeleteClick(product)}
                          >
                            <TrashBinIcon className="h-4 w-4" />
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))
                ) : (
                  <TableRow>
                    <TableCell colSpan={6} className="text-center py-8">
                      <p className="text-gray-500 dark:text-gray-400">No products found</p>
                    </TableCell>
                  </TableRow>
                )}
                {/* No products message is now handled in the conditional above */}
              </TableBody>
            </Table>
          </div>

          {pagination && pagination.total_pages > 1 && (
            <div className="mt-6 flex justify-center">
              <Pagination
                currentPage={pagination.current_page}
                totalPages={pagination.total_pages}
                onPageChange={handlePageChange}
              />
            </div>
          )}
        </div>
      </div>

      {/* Delete Product Modal */}
      {productToDelete && (
        <DeleteProductModal
          isOpen={deleteModal.isOpen}
          onClose={deleteModal.closeModal}
          productId={productToDelete.id}
          productTitle={productToDelete.title}
          onDelete={handleDeleteConfirm}
        />
      )}

      {/* Product Details Modal */}
      <ProductDetailsModal
        isOpen={detailsModal.isOpen}
        onClose={detailsModal.closeModal}
        product={selectedProduct}
      />
    </div>
  );
}
