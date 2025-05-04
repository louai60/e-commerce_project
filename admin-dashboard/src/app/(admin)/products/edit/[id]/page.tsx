"use client";
import React, { useState, useEffect, ChangeEvent } from "react";
import PageBreadcrumb from "@/components/common/PageBreadCrumb";
import { useRouter, useParams } from "next/navigation";
import { ProductService } from "@/services/product.service";
import { useBrands, useCategories, useProduct } from "@/hooks/useProducts";
import Input from "@/components/form/input/InputField";
import Label from "@/components/form/Label";
import Button from "@/components/ui/button/Button";
import { ChevronDownIcon } from "@/icons";
import { toast } from "react-hot-toast";
import LoadingSpinner from "@/components/ui/loading/LoadingSpinner";
import { ImageUpload } from "@/components/ui/image-upload/ImageUpload";
import { Brand } from "@/services/product.service";

export default function EditProductPage() {
  const params = useParams();
  const id = params.id as string;
  const router = useRouter();
  const { product, isLoading, isError, mutate } = useProduct(id);
  const { brands, isLoading: brandsLoading } = useBrands();
  const { categories, isLoading: categoriesLoading } = useCategories();

  const [isSubmitting, setIsSubmitting] = useState(false);
  const [formData, setFormData] = useState({
    title: "",
    slug: "",
    description: "",
    short_description: "",
    price: "",
    sku: "",
    inventory_qty: "0",
    inventory_status: "in_stock",
    is_published: true,
    brand_id: "",
    images: [{ url: "", alt_text: "", position: 1 }]
  });

  // Populate form when product data is loaded
  useEffect(() => {
    if (product) {
      setFormData({
        title: product.title || "",
        slug: product.slug || "",
        description: product.description || "",
        short_description: product.short_description || "",
        price: product.price?.current?.USD?.toString() || "",
        sku: product.sku || "",
        inventory_qty: product.inventory?.quantity?.toString() || "0",
        inventory_status: product.inventory?.status?.toLowerCase() || "in_stock",
        is_published: product.is_published !== undefined ? product.is_published : true,
        brand_id: product.brand?.id || "",
        images: product.images?.length > 0
          ? product.images.map(img => ({
              url: img.url || "",
              alt_text: img.alt_text || "",
              position: img.position || 1
            }))
          : [{ url: "", alt_text: "", position: 1 }]
      });
    }
  }, [product]);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
    const { name, value, type } = e.target as HTMLInputElement;

    if (type === 'checkbox') {
      const checked = (e.target as HTMLInputElement).checked;
      setFormData(prev => ({ ...prev, [name]: checked }));
    } else {
      setFormData(prev => ({ ...prev, [name]: value }));
    }
  };

  const handleInputChange = (e: ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({ ...prev, [name]: value }));
  };

  const handleImageChange = (index: number, field: string, value: string) => {
    const updatedImages = [...formData.images];
    updatedImages[index] = { ...updatedImages[index], [field]: value };
    setFormData(prev => ({ ...prev, images: updatedImages }));
  };

  const addImageField = () => {
    setFormData(prev => ({
      ...prev,
      images: [...prev.images, { url: "", alt_text: "", position: prev.images.length + 1 }]
    }));
  };

  const removeImageField = (index: number) => {
    if (formData.images.length > 1) {
      const updatedImages = formData.images.filter((_, i) => i !== index);
      // Update positions
      const reorderedImages = updatedImages.map((img, i) => ({
        ...img,
        position: i + 1
      }));
      setFormData(prev => ({ ...prev, images: reorderedImages }));
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    await submitForm();
  };

  const submitForm = async () => {
    setIsSubmitting(true);

    try {
      // Filter out empty image URLs
      const filteredImages = formData.images.filter(img => img.url.trim() !== "");

      // Convert string values to appropriate types
      const productData = {
        ...formData,
        price: parseFloat(formData.price),
        inventory_qty: parseInt(formData.inventory_qty),
        images: filteredImages,
        brand_id: formData.brand_id || undefined
      };

      await ProductService.updateProduct(id, productData);
      toast.success("Product updated successfully");
      mutate(); // Refresh product data
    } catch (error: unknown) {
      const err = error as { error?: string };
      console.error("Error updating product:", error);
      toast.error(err.error || "Failed to update product");
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
      <PageBreadcrumb pageTitle={`Edit Product: ${product.title}`} />

      <div className="rounded-xl border border-gray-200 bg-white p-6 dark:border-gray-700 dark:bg-gray-800">
        <form onSubmit={handleSubmit} className="space-y-6">
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
                  defaultValue={formData.title}
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
                    defaultValue={formData.slug}
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
              <Label htmlFor="short_description">Short Description*</Label>
              <Input
                id="short_description"
                name="short_description"
                type="text"
                placeholder="Brief product description"
                defaultValue={formData.short_description}
                onChange={handleInputChange}
              />
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
                onChange={handleChange}
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
                  defaultValue={formData.price}
                  onChange={handleInputChange}
                />
              </div>

              <div>
                <Label htmlFor="sku">SKU*</Label>
                <div className="flex gap-2">
                  <Input
                    id="sku"
                    name="sku"
                    type="text"
                    placeholder="PROD-001"
                    defaultValue={formData.sku}
                    onChange={handleInputChange}
                  />
                  <Button
                    variant="outline"
                    onClick={async () => {
                      if (!formData.brand_id) {
                        toast.error("Brand is required to generate SKU");
                        return;
                      }

                      try {
                        // Find the brand name from the selected brand_id
                        const selectedBrand = brands?.find(brand => brand.id === formData.brand_id);
                        if (!selectedBrand) {
                          toast.error("Selected brand not found");
                          return;
                        }

                        // Use the product's category if available
                        const categoryName = product.categories && product.categories.length > 0
                          ? product.categories[0].name
                          : "";

                        if (!categoryName) {
                          toast.error("Product must have a category to generate SKU");
                          return;
                        }

                        // Call the API to generate a SKU preview
                        const result = await ProductService.generateSKUPreview(
                          selectedBrand.name,
                          categoryName
                        );

                        // Update the SKU field with the generated SKU
                        setFormData(prev => ({ ...prev, sku: result.sku }));
                        toast.success("SKU generated successfully");
                      } catch (error) {
                        console.error("Failed to generate SKU:", error);
                        toast.error("Failed to generate SKU");
                      }
                    }}
                    disabled={!formData.brand_id || !product.categories || product.categories.length === 0}
                  >
                    Generate
                  </Button>
                </div>
              </div>

              <div>
                <Label htmlFor="inventory_qty">Inventory Quantity*</Label>
                <Input
                  id="inventory_qty"
                  name="inventory_qty"
                  type="number"
                  min="0"
                  placeholder="0"
                  defaultValue={formData.inventory_qty}
                  onChange={handleInputChange}
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
                    className="h-11 w-full rounded-lg border appearance-none px-4 py-2.5 text-sm shadow-theme-xs placeholder:text-gray-400 focus:outline-hidden focus:ring-3 dark:bg-gray-900 dark:text-white/90 dark:placeholder:text-white/30 dark:focus:border-brand-800 border-gray-200 focus:border-brand-500 focus:ring-brand-500/20 dark:border-gray-700"
                    value={formData.inventory_status}
                    onChange={handleChange}
                    required
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
                <Label htmlFor="brand_id">Brand</Label>
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
                    {brands?.map((brand: Brand) => (
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
                onClick={addImageField}
              >
                Add Image
              </Button>
            </div>

            {formData.images.map((image, index) => (
              <div key={index} className="rounded-lg border border-gray-200 p-4 dark:border-gray-700">
                {image.url && (
                  <div className="mb-4">
                    <img
                      src={image.url}
                      alt={image.alt_text}
                      className="h-48 w-full rounded-lg object-cover"
                    />
                  </div>
                )}
                <ImageUpload
                  onUploadSuccess={(result) => {
                    const updatedImages = [...formData.images];
                    updatedImages[index] = {
                      url: result.url,
                      alt_text: result.alt_text,
                      position: result.position
                    };
                    setFormData(prev => ({ ...prev, images: updatedImages }));
                  }}
                  onUploadError={(error) => {
                    toast.error(error);
                  }}
                  folder="products"
                  defaultAltText={image.alt_text || `Product image ${index + 1}`}
                  defaultPosition={image.position || index + 1}
                />
                <div className="mt-4 flex justify-end">
                  <Button
                    variant="outline"
                    size="sm"
                    className="text-danger-500 hover:border-danger-500 hover:bg-danger-500/10"
                    onClick={() => removeImageField(index)}
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
                onClick={() => router.push("/products")}
              >
                Cancel
              </Button>
              <Button
                variant="primary"
                disabled={isSubmitting}
                onClick={submitForm}
              >
                {isSubmitting ? "Saving..." : "Save Changes"}
              </Button>
            </div>
          </div>
        </form>
      </div>
    </div>
  );
}
