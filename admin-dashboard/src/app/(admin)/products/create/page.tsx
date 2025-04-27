"use client";
import React, { useState, ChangeEvent } from "react";
import PageBreadcrumb from "@/components/common/PageBreadCrumb";
import { useRouter } from "next/navigation";
import { useBrands, useCategories } from "@/hooks/useProducts";
import Input from "@/components/form/input/InputField";
import Label from "@/components/form/Label";
import Button from "@/components/ui/button/Button";
import { ChevronDownIcon, PlusIcon } from "@/icons";
import { toast } from "react-hot-toast";
import { ImageUpload } from "@/components/ui/image-upload/ImageUpload";
import { api } from '@/lib/api';
import { useProductContext } from "@/contexts/ProductContext";
import { Product } from "@/services/product.service";
import CreateCategoryModal from "@/components/modals/CreateCategoryModal";
import CreateBrandModal from "@/components/modals/CreateBrandModal";

interface ProductFormData {
  title: string;
  slug: string;
  description: string;
  price: number;
  sku: string;
  inventory_quantity: number;
  inventory_status: string;
  brand_id?: string;
  categories?: string[];
  is_published: boolean;
  images: Array<{
    url: string;
    alt_text: string;
    position: number;
  }>;
}

export default function CreateProductPage() {
  const router = useRouter();
  const { brands, isLoading: brandsLoading, mutate: refreshBrands } = useBrands();
  const { categories, isLoading: categoriesLoading, mutate: refreshCategories } = useCategories();
  const { refreshProducts, addOptimisticProduct } = useProductContext();

  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isCategoryModalOpen, setIsCategoryModalOpen] = useState(false);
  const [isBrandModalOpen, setIsBrandModalOpen] = useState(false);
  const [formData, setFormData] = useState<ProductFormData>({
    title: "",
    slug: "",
    description: "",
    price: 0,
    sku: "",
    inventory_quantity: 0,
    inventory_status: "in_stock",
    is_published: true,
    categories: [],
    images: [{ url: "", alt_text: "", position: 1 }],
  });

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
    const { name, value, type } = e.target as HTMLInputElement;

    if (type === 'checkbox') {
      const checked = (e.target as HTMLInputElement).checked;
      setFormData(prev => ({ ...prev, [name]: checked }));
    } else {
      setFormData(prev => ({ ...prev, [name]: value }));
    }
  };

  const handleCategoryChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const selectedOptions = Array.from(e.target.selectedOptions).map(option => option.value);
    setFormData(prev => ({ ...prev, categories: selectedOptions }));
  };

  const handleInputChange = (
    e: ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>
  ) => {
    const { name, value } = e.target;
    setFormData((prev) => ({
      ...prev,
      [name]: value,
    }));
  };

  const handleImageUpload = (index: number, result: { url: string; alt_text: string; position: number }) => {
    console.log('Image upload result:', result);

    // Make sure we have a valid URL
    if (!result.url) {
      toast.error('Image upload failed. Using placeholder image.');
      // Use a placeholder image URL
      result.url = `https://placehold.co/600x400?text=Product+Image+${index + 1}`;
    }

    setFormData(prev => {
      const newImages = [...prev.images];
      newImages[index] = {
        url: result.url,
        alt_text: result.alt_text || `Product image ${index + 1}`,
        position: result.position || index + 1
      };
      return { ...prev, images: newImages };
    });
  };

  const handleSubmit = async (e?: React.FormEvent) => {
    if (e) e.preventDefault();

    console.log('Submit button clicked');

    // Validate form data
    if (!formData.title) {
      toast.error("Product title is required");
      return;
    }

    if (!formData.slug) {
      toast.error("Product slug is required");
      return;
    }

    if (!formData.price || formData.price <= 0) {
      toast.error("Product price must be greater than 0");
      return;
    }

    if (!formData.sku) {
      toast.error("Product SKU is required");
      return;
    }

    if (formData.images.length === 0) {
      toast.error("At least one product image is required");
      return;
    }

    // Check if any image has an empty URL and use a placeholder
    const processedImages = formData.images.map((img, index) => {
      if (!img.url) {
        toast.error(`Image ${index + 1} has no URL. Using placeholder image.`);
        return {
          ...img,
          url: `https://placehold.co/600x400?text=Product+Image+${index + 1}`,
          alt_text: img.alt_text || `Product image ${index + 1}`,
          position: img.position || index + 1
        };
      }
      return img;
    });

    console.log('All validations passed');
    setIsSubmitting(true);
    toast.success("Creating product...");

    try {
      // Prepare the data for submission
      // Ensure price is a number, not a string
      const price = parseFloat(formData.price.toString());
      if (isNaN(price)) {
        toast.error("Price must be a valid number");
        setIsSubmitting(false);
        return;
      }

      // Ensure inventory_qty is a number
      const inventoryQty = parseInt(formData.inventory_quantity.toString());
      if (isNaN(inventoryQty)) {
        toast.error("Inventory quantity must be a valid number");
        setIsSubmitting(false);
        return;
      }

      const productData = {
        title: formData.title,
        slug: formData.slug,
        description: formData.description,
        short_description: formData.description?.substring(0, 150) || '',
        price: price, // Use the parsed float value
        sku: formData.sku,
        inventory_qty: inventoryQty, // Use the parsed integer value
        inventory_status: formData.inventory_status,
        is_published: formData.is_published,
        brand_id: formData.brand_id ? formData.brand_id : undefined,
        categories: formData.categories?.length ? formData.categories : undefined,
        images: processedImages.filter(img => img.url),
      };

      console.log('Submitting product data:', JSON.stringify(productData, null, 2));

      // Use the existing API instance from lib/api.ts
      // This ensures all the authentication middleware is properly applied

      console.log('Sending API request...');
      // Make sure we're sending the correct structure
      const requestData = {
        product: {
          title: productData.title,
          slug: productData.slug,
          description: productData.description,
          short_description: productData.short_description,
          price: Number(productData.price),
          sku: productData.sku,
          inventory_qty: Number(productData.inventory_qty),
          inventory_status: productData.inventory_status,
          is_published: productData.is_published,
          // Send brand_id as a string, not an object with a value property
          brand_id: productData.brand_id || undefined,
          // Use category_ids instead of categories - send as array of strings
          category_ids: productData.categories || undefined,
          images: productData.images
        }
      };
      console.log('Request data:', JSON.stringify(requestData, null, 2));

      // Check if we have a token
      const token = localStorage.getItem('access_token');
      console.log('Auth token available:', !!token);

      const response = await api.post('/products', requestData);
      console.log('API response:', response.data);

      if (response.data) {
        // Create an optimistic product object to add to the list
        const optimisticProduct: Product = {
          id: response.data.id,
          title: formData.title,
          slug: formData.slug,
          description: formData.description,
          short_description: formData.description?.substring(0, 150) || '',
          sku: formData.sku,
          price: {
            current: {
              USD: Number(formData.price)
            },
            currency: 'USD'
          },
          inventory: {
            status: formData.inventory_status,
            available: formData.inventory_status === 'in_stock',
            quantity: Number(formData.inventory_quantity)
          },
          images: processedImages.filter(img => img.url).map(img => ({
            id: `temp-${Math.random().toString(36).substring(2, 9)}`,
            url: img.url,
            alt_text: img.alt_text,
            position: img.position
          })),
          variants: [],
          categories: [],
          is_published: formData.is_published,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString()
        };

        // Add the new product to the list optimistically
        addOptimisticProduct(optimisticProduct);

        toast.success("Product created successfully");

        // Refresh the product list to get the actual data from the server
        // Use a small timeout to ensure the UI has time to update with the optimistic data first
        setTimeout(() => {
          refreshProducts();
          // Navigate to the products page after a short delay
          setTimeout(() => {
            router.push("/products");
          }, 100);
        }, 300);
      }
    } catch (error: any) {
      console.error('Error creating product:', error);
      console.error('Error details:', error.response?.data || error.message);

      // Handle specific error cases
      if (error.response?.status === 401) {
        toast.error("Authentication error. Please log in again.");
        // Redirect to login page
        router.push("/login");
      } else if (error.response?.status === 403) {
        toast.error("You don't have permission to create products. Admin access required.");
      } else if (error.response?.data?.error?.includes('wrapperspb.StringValue')) {
        // This is a type conversion error
        console.error('Type conversion error detected. The server expects a different format for string values.');
        toast.error("There was an issue with the data format. Please try again.");
      } else {
        toast.error(error.response?.data?.error || error.message || "Failed to create product");
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  const generateSlug = () => {
    const slug = formData.title
      .toLowerCase()
      .replace(/[^\w\s-]/g, '')
      .replace(/[\s_-]+/g, '-')
      .replace(/^-+|-+$/g, '');

    setFormData(prev => ({ ...prev, slug }));
  };

  return (
    <div>
      <PageBreadcrumb pageTitle="Create Product" />

      <div className="rounded-xl border border-gray-200 bg-white p-6 dark:border-gray-700 dark:bg-gray-800">
        <form onSubmit={(e) => e.preventDefault()} className="space-y-6">
          {/* Basic Information */}
          <div className="space-y-4">
            <h2 className="text-xl font-semibold text-gray-900 dark:text-white">Basic Information</h2>

            <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
              <div>
                <Label htmlFor="title">Product Title*</Label>
                <Input
                  id="title"
                  name="title"
                  type="text"
                  placeholder="Enter product title"
                  value={formData.title}
                  onChange={handleInputChange}
                />
              </div>

              <div>
                <Label htmlFor="slug">Slug*</Label>
                <div className="flex gap-2">
                  <Input
                    id="slug"
                    name="slug"
                    type="text"
                    placeholder="product-slug"
                    value={formData.slug}
                    onChange={handleInputChange}
                  />
                  <Button
                    variant="outline"
                    onClick={generateSlug}
                    disabled={!formData.title}
                  >
                    Generate
                  </Button>
                </div>
              </div>
            </div>

            <div>
              <Label htmlFor="description">Full Description*</Label>
              <textarea
                id="description"
                name="description"
                rows={5}
                className="h-auto w-full rounded-lg border appearance-none px-4 py-2.5 text-sm shadow-theme-xs placeholder:text-gray-400 focus:outline-hidden focus:ring-3 dark:bg-gray-900 dark:text-white/90 dark:placeholder:text-white/30 dark:focus:border-brand-800 border-gray-200 focus:border-brand-500 focus:ring-brand-500/20 dark:border-gray-700"
                placeholder="Detailed product description"
                value={formData.description}
                onChange={handleInputChange}
                required
              />
            </div>
          </div>

          {/* Pricing & Inventory */}
          <div className="space-y-4 pt-4">
            <h2 className="text-xl font-semibold text-gray-900 dark:text-white">Pricing & Inventory</h2>

            <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
              <div>
                <Label htmlFor="price">Price*</Label>
                <Input
                  id="price"
                  name="price"
                  type="number"
                  step={0.01}
                  min="0"
                  placeholder="0.00"
                  value={formData.price}
                  onChange={(e) => {
                    // Ensure it's a valid number
                    const value = parseFloat(e.target.value);
                    setFormData(prev => ({
                      ...prev,
                      price: isNaN(value) ? 0 : value
                    }));
                  }}
                />
              </div>

              <div>
                <Label htmlFor="sku">SKU*</Label>
                <Input
                  id="sku"
                  name="sku"
                  type="text"
                  placeholder="PROD-001"
                  value={formData.sku}
                  onChange={handleInputChange}
                />
              </div>

              <div>
                <Label htmlFor="inventory_quantity">Inventory Quantity*</Label>
                <Input
                  id="inventory_quantity"
                  name="inventory_quantity"
                  type="number"
                  min="0"
                  placeholder="0"
                  value={formData.inventory_quantity}
                  onChange={(e) => {
                    // Ensure it's a valid integer
                    const value = parseInt(e.target.value);
                    setFormData(prev => ({
                      ...prev,
                      inventory_quantity: isNaN(value) ? 0 : value
                    }));
                  }}
                />
              </div>
            </div>

            <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
              <div>
                <Label htmlFor="inventory_status">Inventory Status*</Label>
                <div className="relative">
                  <select
                    id="inventory_status"
                    name="inventory_status"
                    value={formData.inventory_status}
                    onChange={(e: ChangeEvent<HTMLSelectElement>) => handleInputChange(e)}
                    className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm dark:bg-gray-700 dark:border-gray-600 dark:text-white"
                  >
                    <option value="in_stock">In Stock</option>
                    <option value="out_of_stock">Out of Stock</option>
                    <option value="backorder">Backorder</option>
                    <option value="preorder">Preorder</option>
                  </select>
                  <span className="absolute text-gray-500 -translate-y-1/2 pointer-events-none right-3 top-1/2 dark:text-gray-400">
                    <ChevronDownIcon/>
                  </span>
                </div>
              </div>

              <div>
                <div className="flex items-center justify-between mb-1">
                  <Label htmlFor="brand_id">Brand</Label>
                  <Button
                    variant="outline"
                    size="sm"
                    type="button"
                    className="text-xs flex items-center gap-1 h-6 px-2"
                    onClick={() => setIsBrandModalOpen(true)}
                  >
                    <PlusIcon className="h-3 w-3" />
                    Add Brand
                  </Button>
                </div>
                <div className="relative">
                  <select
                    id="brand_id"
                    name="brand_id"
                    className="h-11 w-full rounded-lg border appearance-none px-4 py-2.5 text-sm shadow-theme-xs placeholder:text-gray-400 focus:outline-hidden focus:ring-3 dark:bg-gray-900 dark:text-white/90 dark:placeholder:text-white/30 dark:focus:border-brand-800 border-gray-200 focus:border-brand-500 focus:ring-brand-500/20 dark:border-gray-700"
                    value={formData.brand_id}
                    onChange={handleChange}
                    disabled={brandsLoading}
                  >
                    <option value="">Select a brand</option>
                    {brands?.map((brand: any) => (
                      <option key={brand.id} value={brand.id}>
                        {brand.name}
                      </option>
                    ))}
                  </select>
                  <span className="absolute text-gray-500 -translate-y-1/2 pointer-events-none right-3 top-1/2 dark:text-gray-400">
                    <ChevronDownIcon/>
                  </span>
                </div>
              </div>
            </div>

            <div>
              <div className="flex items-center justify-between mb-1">
                <Label htmlFor="categories">Categories</Label>
                <Button
                  variant="outline"
                  size="sm"
                  type="button"
                  className="text-xs flex items-center gap-1 h-6 px-2"
                  onClick={() => setIsCategoryModalOpen(true)}
                >
                  <PlusIcon className="h-3 w-3" />
                  Add Category
                </Button>
              </div>
              <div className="relative">
                <select
                  id="categories"
                  name="categories"
                  className="h-11 w-full rounded-lg border appearance-none px-4 py-2.5 text-sm shadow-theme-xs placeholder:text-gray-400 focus:outline-hidden focus:ring-3 dark:bg-gray-900 dark:text-white/90 dark:placeholder:text-white/30 dark:focus:border-brand-800 border-gray-200 focus:border-brand-500 focus:ring-brand-500/20 dark:border-gray-700"
                  value={formData.categories}
                  onChange={handleCategoryChange}
                  disabled={categoriesLoading}
                  multiple
                  size={4}
                >
                  {categories?.map((category: any) => (
                    <option key={category.id} value={category.id}>
                      {category.name} {category.parent_name ? `(${category.parent_name})` : ''}
                    </option>
                  ))}
                </select>
                <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">Hold Ctrl (or Cmd) to select multiple categories</p>
              </div>
            </div>

            <div className="flex items-center">
              <input
                id="is_published"
                name="is_published"
                type="checkbox"
                className="h-4 w-4 rounded border-gray-300 text-brand-500 focus:ring-brand-500"
                checked={formData.is_published}
                onChange={(e) => setFormData(prev => ({ ...prev, is_published: e.target.checked }))}
              />
              <label htmlFor="is_published" className="ml-2 text-sm text-gray-700 dark:text-gray-300">
                Publish product (visible to customers)
              </label>
            </div>
          </div>

          {/* Images */}
          <div className="space-y-4 pt-4">
            <div className="flex items-center justify-between">
              <h2 className="text-xl font-semibold text-gray-900 dark:text-white">Product Images</h2>
              <Button
                variant="outline"
                size="sm"
                type="button"
                onClick={() => setFormData(prev => ({ ...prev, images: [...prev.images, { url: "", alt_text: "", position: prev.images.length + 1 }] }))}
              >
                Add Image
              </Button>
            </div>

            {formData.images.map((_, index) => (
              <div key={index} className="rounded-lg border border-gray-200 p-4 dark:border-gray-700">
                <ImageUpload
                  onUploadSuccess={(result) => handleImageUpload(index, result)}
                  onUploadError={(error) => {
                    toast.error(error);
                  }}
                  folder="products"
                  defaultAltText={`Product image ${index + 1}`}
                  defaultPosition={index + 1}
                />
                <div className="mt-4 flex justify-end">
                  <Button
                    variant="outline"
                    size="sm"
                    type="button"
                    className="text-danger-500 hover:border-danger-500 hover:bg-danger-500/10"
                    onClick={() => setFormData(prev => ({
                      ...prev,
                      images: prev.images.filter((_, i) => i !== index)
                    }))}
                    disabled={formData.images.length === 1}
                  >
                    Remove Image
                  </Button>
                </div>
              </div>
            ))}
          </div>

          {/* Submit Button */}
          <div className="flex justify-end pt-6">
            <div className="flex gap-3">
              <Button
                variant="outline"
                type="button"
                onClick={() => router.push("/products")}
              >
                Cancel
              </Button>
              <Button
                variant="primary"
                type="button"
                disabled={isSubmitting}
                onClick={handleSubmit}
              >
                {isSubmitting ? "Creating..." : "Create Product"}
              </Button>
            </div>
          </div>
        </form>
      </div>

      {/* Category Modal */}
      <CreateCategoryModal
        isOpen={isCategoryModalOpen}
        onClose={() => setIsCategoryModalOpen(false)}
        onSuccess={() => refreshCategories()}
      />

      {/* Brand Modal */}
      <CreateBrandModal
        isOpen={isBrandModalOpen}
        onClose={() => setIsBrandModalOpen(false)}
        onSuccess={() => refreshBrands()}
      />
    </div>
  );
}
