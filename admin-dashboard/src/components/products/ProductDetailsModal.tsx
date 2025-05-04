"use client";
import React from 'react';
import { Modal } from '@/components/ui/modal';
import { Product } from '@/services/product.service';
import { formatPrice } from '@/lib/utils';
import Image from 'next/image';
import Badge from '@/components/ui/badge/Badge';

interface ProductDetailsModalProps {
  isOpen: boolean;
  onClose: () => void;
  product: Product | null;
}

const ProductDetailsModal: React.FC<ProductDetailsModalProps> = ({
  isOpen,
  onClose,
  product
}) => {
  if (!product) return null;

  const renderProductImage = (): React.ReactNode => {
    if (product.images && product.images.length > 0) {
      return (
        <Image
          src={product.images[0].url}
          alt={product.images[0].alt_text || product.title}
          width={300}
          height={300}
          className="w-full h-auto object-cover"
          unoptimized={true}
        />
      );
    }
    return (
      <div className="w-full h-64 bg-gray-100 dark:bg-gray-800 flex items-center justify-center">
        <span className="text-gray-400">No image</span>
      </div>
    );
  };

  const renderAdditionalImages = (): React.ReactNode => {
    if (!product.images || product.images.length <= 1) return null;

    return (
      <div className="mt-4 grid grid-cols-4 gap-2">
        {product.images.slice(1, 5).map((image, index) => (
          <div key={image.id || index} className="rounded-lg overflow-hidden border border-gray-200 dark:border-gray-700">
            <Image
              src={image.url}
              alt={image.alt_text || `${product.title} - ${index + 2}`}
              width={80}
              height={80}
              className="w-full h-auto object-cover"
              unoptimized={true}
            />
          </div>
        ))}
      </div>
    );
  };

  const renderPrice = (): string => {
    if (product.price?.current?.USD) {
      return formatPrice(product.price.current.USD);
    }
    if (product.price?.value) {
      return formatPrice(product.price.value);
    }
    return "N/A";
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      className="max-w-4xl p-6 overflow-y-auto max-h-[90vh]"
    >
      <div className="space-y-6">
        <div className="flex flex-col md:flex-row gap-6">
          {/* Product Image */}
          <div className="w-full md:w-1/3">
            <div className="rounded-lg overflow-hidden border border-gray-200 dark:border-gray-700">
              {renderProductImage()}
            </div>
            {renderAdditionalImages()}
          </div>

          {/* Product Details */}
          <div className="w-full md:w-2/3">
            <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-2">
              {product.title}
            </h2>

            <div className="flex items-center gap-3 mb-4">
              <Badge
                variant={product.is_published ? "success" : "warning"}
              >
                {product.is_published ? "Published" : "Draft"}
              </Badge>

              <Badge
                variant="success"
                className="bg-green-500 text-white px-3 py-1 rounded-full text-xs"
              >
                in_stock
              </Badge>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">
              <div>
                <span className="text-sm text-gray-500 dark:text-gray-400">SKU:</span>
                <span className="ml-2 font-medium text-gray-900 dark:text-white">
                  {product.sku || "N/A"}
                </span>
              </div>

              <div>
                <span className="text-sm text-gray-500 dark:text-gray-400">Price:</span>
                <span className="ml-2 font-semibold text-gray-900 dark:text-white">
                  {renderPrice()}
                </span>
              </div>

              <div>
                <span className="text-sm text-gray-500 dark:text-gray-400">Inventory:</span>
                <span className="ml-2 font-medium text-gray-900 dark:text-white">
                  {product.inventory?.quantity || 0} units
                </span>
              </div>

              <div>
                <span className="text-sm text-gray-500 dark:text-gray-400">Created:</span>
                <span className="ml-2 font-medium text-gray-900 dark:text-white">
                  {new Date(product.created_at).toLocaleDateString()}
                </span>
              </div>
            </div>

            {product.short_description && (
              <div className="mb-4">
                <h3 className="text-md font-semibold text-gray-900 dark:text-white mb-2">
                  Short Description
                </h3>
                <p className="text-gray-600 dark:text-gray-300">
                  {product.short_description}
                </p>
              </div>
            )}
          </div>
        </div>

        {/* Additional Details Sections */}
        <div className="border-t border-gray-200 dark:border-gray-700 pt-6 space-y-6">
          {/* Description */}
          {product.description && (
            <div>
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-3">
                Description
              </h3>
              <div className="text-gray-600 dark:text-gray-300 prose max-w-none dark:prose-invert">
                {product.description}
              </div>
            </div>
          )}

          {/* Categories */}
          {product.categories && product.categories.length > 0 && (
            <div>
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-3">
                Categories
              </h3>
              <div className="flex flex-wrap gap-2">
                {product.categories.map((category) => (
                  <Badge key={category.id} variant="default">
                    {category.name}
                  </Badge>
                ))}
              </div>
            </div>
          )}

          {/* Brand */}
          {product.brand && (
            <div>
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-3">
                Brand
              </h3>
              <p className="text-gray-600 dark:text-gray-300">
                {product.brand.name}
              </p>
            </div>
          )}

          {/* Variants */}
          {product.variants && product.variants.length > 0 && (
            <div>
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-3">
                Variants
              </h3>
              <div className="overflow-x-auto">
                <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
                  <thead className="bg-gray-50 dark:bg-gray-800">
                    <tr>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                        SKU
                      </th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                        Price
                      </th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                        Inventory
                      </th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                        Attributes
                      </th>
                    </tr>
                  </thead>
                  <tbody className="bg-white dark:bg-gray-900 divide-y divide-gray-200 dark:divide-gray-700">
                    {product.variants.map((variant) => (
                      <tr key={variant.id}>
                        <td className="px-4 py-3 whitespace-nowrap text-sm text-gray-900 dark:text-white">
                          {variant.sku}
                        </td>
                        <td className="px-4 py-3 whitespace-nowrap text-sm text-gray-900 dark:text-white">
                          {formatPrice(variant.price)}
                        </td>
                        <td className="px-4 py-3 whitespace-nowrap text-sm text-gray-900 dark:text-white">
                          {variant.inventory_qty}
                        </td>
                        <td className="px-4 py-3 text-sm text-gray-900 dark:text-white">
                          {variant.attributes && variant.attributes.map((attr, index) => (
                            <span key={index} className="inline-block mr-2">
                              {attr.name}: {attr.value}
                            </span>
                          ))}
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
    </Modal>
  );
};

export default ProductDetailsModal;
