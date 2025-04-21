"use client";
import React from "react";
import PageBreadcrumb from "@/components/common/PageBreadCrumb";
import { useProduct } from "@/hooks/useProducts";
import { useRouter } from "next/navigation";
import Button from "@/components/ui/button/Button";
import { PencilIcon, TrashBinIcon, ChevronLeftIcon } from "@/icons";
import LoadingSpinner from "@/components/ui/loading/LoadingSpinner";
import Image from "next/image";
import { formatPrice } from "@/lib/utils";
import Badge from "@/components/ui/badge/Badge";
import Link from "next/link";

export default function ProductDetailPage({ params }: { params: { id: string } }) {
  const router = useRouter();
  const { product, isLoading, isError } = useProduct(params.id);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <LoadingSpinner />
      </div>
    );
  }

  if (isError || !product) {
    return (
      <div className="rounded-xl border border-gray-200 bg-white p-6 dark:border-gray-700 dark:bg-gray-800">
        <div className="text-center text-red-500">
          Error loading product
        </div>
      </div>
    );
  }

  return (
    <div>
      <div className="mb-6 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <PageBreadcrumb pageTitle={product.title} />
        <div className="flex gap-2">
          <Link href="/products">
            <Button variant="outline" startIcon={<ChevronLeftIcon />}>
              Back to Products
            </Button>
          </Link>
          <Link href={`/products/edit/${product.id}`}>
            <Button variant="outline" startIcon={<PencilIcon />}>
              Edit
            </Button>
          </Link>
          <Button 
            variant="outline" 
            className="text-danger-500 hover:border-danger-500 hover:bg-danger-500/10"
            startIcon={<TrashBinIcon />}
          >
            Delete
          </Button>
        </div>
      </div>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        {/* Product Images */}
        <div className="lg:col-span-1">
          <div className="rounded-xl border border-gray-200 bg-white p-4 dark:border-gray-700 dark:bg-gray-800">
            <div className="mb-4 overflow-hidden rounded-lg border border-gray-200 dark:border-gray-700">
              {product.images && product.images.length > 0 ? (
                <Image
                  src={product.images[0].url}
                  alt={product.images[0].alt_text || product.title}
                  width={500}
                  height={500}
                  className="h-full w-full object-cover"
                />
              ) : (
                <div className="flex h-64 w-full items-center justify-center bg-gray-100 dark:bg-gray-800">
                  <span className="text-gray-400">No image available</span>
                </div>
              )}
            </div>
            
            {product.images && product.images.length > 1 && (
              <div className="grid grid-cols-4 gap-2">
                {product.images.slice(0, 4).map((image, index) => (
                  <div 
                    key={image.id || index} 
                    className="overflow-hidden rounded-lg border border-gray-200 dark:border-gray-700"
                  >
                    <Image
                      src={image.url}
                      alt={image.alt_text || `${product.title} image ${index + 1}`}
                      width={100}
                      height={100}
                      className="h-full w-full object-cover"
                    />
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
        
        {/* Product Details */}
        <div className="lg:col-span-2">
          <div className="rounded-xl border border-gray-200 bg-white p-6 dark:border-gray-700 dark:bg-gray-800">
            <div className="mb-6 border-b border-gray-200 pb-6 dark:border-gray-700">
              <h1 className="mb-2 text-2xl font-bold text-gray-900 dark:text-white">
                {product.title}
              </h1>
              <p className="mb-4 text-gray-500 dark:text-gray-400">
                {product.short_description}
              </p>
              
              <div className="flex flex-wrap items-center gap-4">
                <div className="flex items-center gap-2">
                  <span className="text-sm text-gray-500 dark:text-gray-400">Price:</span>
                  <span className="text-xl font-semibold text-gray-900 dark:text-white">
                    {product.price?.current?.USD ? formatPrice(product.price.current.USD) : "N/A"}
                  </span>
                </div>
                
                <div className="flex items-center gap-2">
                  <span className="text-sm text-gray-500 dark:text-gray-400">SKU:</span>
                  <span className="font-medium text-gray-900 dark:text-white">
                    {product.sku || "N/A"}
                  </span>
                </div>
                
                <div className="flex items-center gap-2">
                  <span className="text-sm text-gray-500 dark:text-gray-400">Status:</span>
                  <Badge
                    variant={
                      product.inventory?.available
                        ? "success"
                        : "danger"
                    }
                  >
                    {product.inventory?.status || "OUT_OF_STOCK"}
                  </Badge>
                </div>
              </div>
            </div>
            
            <div className="mb-6 grid grid-cols-1 gap-6 md:grid-cols-2">
              <div>
                <h2 className="mb-3 text-lg font-semibold text-gray-900 dark:text-white">
                  Basic Information
                </h2>
                <div className="space-y-2">
                  <div className="flex justify-between">
                    <span className="text-sm text-gray-500 dark:text-gray-400">ID:</span>
                    <span className="text-sm font-medium text-gray-900 dark:text-white">
                      {product.id}
                    </span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-sm text-gray-500 dark:text-gray-400">Slug:</span>
                    <span className="text-sm font-medium text-gray-900 dark:text-white">
                      {product.slug}
                    </span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-sm text-gray-500 dark:text-gray-400">Brand:</span>
                    <span className="text-sm font-medium text-gray-900 dark:text-white">
                      {product.brand?.name || "N/A"}
                    </span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-sm text-gray-500 dark:text-gray-400">Published:</span>
                    <span className="text-sm font-medium text-gray-900 dark:text-white">
                      {product.is_published ? "Yes" : "No"}
                    </span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-sm text-gray-500 dark:text-gray-400">Created:</span>
                    <span className="text-sm font-medium text-gray-900 dark:text-white">
                      {product.created_at ? new Date(product.created_at).toLocaleString() : "N/A"}
                    </span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-sm text-gray-500 dark:text-gray-400">Updated:</span>
                    <span className="text-sm font-medium text-gray-900 dark:text-white">
                      {product.updated_at ? new Date(product.updated_at).toLocaleString() : "N/A"}
                    </span>
                  </div>
                </div>
              </div>
              
              <div>
                <h2 className="mb-3 text-lg font-semibold text-gray-900 dark:text-white">
                  Inventory
                </h2>
                <div className="space-y-2">
                  <div className="flex justify-between">
                    <span className="text-sm text-gray-500 dark:text-gray-400">Quantity:</span>
                    <span className="text-sm font-medium text-gray-900 dark:text-white">
                      {product.inventory?.quantity || 0}
                    </span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-sm text-gray-500 dark:text-gray-400">Status:</span>
                    <span className="text-sm font-medium text-gray-900 dark:text-white">
                      {product.inventory?.status || "OUT_OF_STOCK"}
                    </span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-sm text-gray-500 dark:text-gray-400">Available:</span>
                    <span className="text-sm font-medium text-gray-900 dark:text-white">
                      {product.inventory?.available ? "Yes" : "No"}
                    </span>
                  </div>
                </div>
                
                {product.categories && product.categories.length > 0 && (
                  <>
                    <h2 className="mb-3 mt-6 text-lg font-semibold text-gray-900 dark:text-white">
                      Categories
                    </h2>
                    <div className="flex flex-wrap gap-2">
                      {product.categories.map(category => (
                        <Badge key={category.id} variant="secondary">
                          {category.name}
                        </Badge>
                      ))}
                    </div>
                  </>
                )}
              </div>
            </div>
            
            <div className="mb-6">
              <h2 className="mb-3 text-lg font-semibold text-gray-900 dark:text-white">
                Description
              </h2>
              <div className="rounded-lg border border-gray-200 bg-gray-50 p-4 dark:border-gray-700 dark:bg-gray-900">
                <p className="text-sm text-gray-700 dark:text-gray-300">
                  {product.description || "No description available."}
                </p>
              </div>
            </div>
            
            {product.variants && product.variants.length > 0 && (
              <div>
                <h2 className="mb-3 text-lg font-semibold text-gray-900 dark:text-white">
                  Variants
                </h2>
                <div className="overflow-x-auto">
                  <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
                    <thead className="bg-gray-50 dark:bg-gray-800">
                      <tr>
                        <th scope="col" className="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400">
                          SKU
                        </th>
                        <th scope="col" className="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400">
                          Attributes
                        </th>
                        <th scope="col" className="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400">
                          Price
                        </th>
                        <th scope="col" className="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400">
                          Inventory
                        </th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-200 bg-white dark:divide-gray-700 dark:bg-gray-800">
                      {product.variants.map(variant => (
                        <tr key={variant.id}>
                          <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-900 dark:text-white">
                            {variant.sku}
                          </td>
                          <td className="px-4 py-3 text-sm text-gray-900 dark:text-white">
                            {variant.attributes?.map(attr => (
                              <div key={attr.name}>
                                <span className="font-medium">{attr.name}:</span> {attr.value}
                              </div>
                            ))}
                          </td>
                          <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-900 dark:text-white">
                            {variant.price?.current?.USD ? formatPrice(variant.price.current.USD) : "N/A"}
                          </td>
                          <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-900 dark:text-white">
                            {variant.inventory?.quantity || 0}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
