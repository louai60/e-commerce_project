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
import { Product, ProductService, Brand, Category } from "@/services/product.service"; // Import Brand and Category
import CreateCategoryModal from "@/components/modals/CreateCategoryModal";
import CreateBrandModal from "@/components/modals/CreateBrandModal";

interface ProductFormData {
  title: string;
  slug: string;
  description: string;
  short_description: string;
  product_type: string;
  price: number;
  discount_price?: number;
  sku: string;
  inventory_quantity: number;
  weight?: number;
  brand_id?: string;
  categories?: string[];
  is_published: boolean;
  images: Array<{
    url: string;
    alt_text: string;
    position: number;
  }>;
  specifications: Array<{
    name: string;
    value: string;
    unit: string;
  }>;
  tags: string[];
  seo: {
    meta_title: string;
    meta_description: string;
    keywords: string[];
  };
  shipping: {
    free_shipping: boolean;
    estimated_days: number;
    express_available: boolean;
  };
  variants: Array<{
    title: string;
    sku: string;
    price: number;
    discount_price?: number;
    inventory_quantity: number;
    attributes: Array<{
      name: string;
      value: string;
    }>;
    images: Array<{
      url: string;
      alt_text: string;
      position: number;
    }>;
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
    short_description: "",
    product_type: "physical", // Default to physical product
    price: 0,
    discount_price: 0,
    sku: "",
    inventory_quantity: 0,
    weight: 0,
    is_published: true,
    brand_id: "",
    categories: [],
    images: [{ url: "", alt_text: "", position: 1 }],
    specifications: [{ name: "", value: "", unit: "" }],
    tags: [],
    seo: {
      meta_title: "",
      meta_description: "",
      keywords: []
    },
    shipping: {
      free_shipping: false,
      estimated_days: 3,
      express_available: false
    },
    variants: []
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

    // Handle nested properties (for seo and shipping fields)
    if (name.includes('.')) {
      const [parent, child] = name.split('.');
      setFormData((prev) => ({
        ...prev,
        [parent]: {
          ...(prev[parent as keyof typeof prev] as object),
          [child]: value
        }
      }));
    } else {
      setFormData((prev) => ({
        ...prev,
        [name]: value,
      }));
    }
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
        short_description: formData.short_description || formData.description?.substring(0, 150) || '',
        product_type: formData.product_type,
        price: price,
        discount_price: formData.discount_price,
        sku: formData.sku,
        inventory_quantity: inventoryQty,
        weight: formData.weight,
        is_published: formData.is_published,
        brand_id: formData.brand_id ? formData.brand_id : undefined,
        categories: formData.categories?.length ? formData.categories : undefined,
        images: processedImages.filter(img => img.url),
        specifications: formData.specifications,
        tags: formData.tags,
        seo: formData.seo,
        shipping: formData.shipping,
        variants: formData.variants
      };

      console.log('Submitting product data:', JSON.stringify(productData, null, 2));

      // Create a product request without inventory_qty field
      const requestData = {
        product: {
          title: productData.title,
          slug: productData.slug,
          description: productData.description,
          short_description: productData.short_description,
          product_type: productData.product_type,
          price: Number(productData.price),
          discount_price: productData.discount_price && productData.discount_price > 0 ? Number(productData.discount_price) : undefined,
          sku: productData.sku,
          weight: productData.weight && productData.weight > 0 ? Number(productData.weight) : undefined,
          is_published: productData.is_published,
          brand_id: productData.brand_id || undefined,
          category_ids: productData.categories || undefined,
          images: productData.images,
          specifications: productData.specifications.filter(spec => spec.name && spec.value),
          tags: productData.tags,
          seo: productData.seo.meta_title ? {
            meta_title: productData.seo.meta_title,
            meta_description: productData.seo.meta_description,
            keywords: productData.seo.keywords
          } : undefined,
          shipping: {
            free_shipping: productData.shipping.free_shipping,
            estimated_days: productData.shipping.estimated_days,
            express_available: productData.shipping.express_available
          },
          variants: productData.variants.map(variant => ({
            title: variant.title,
            sku: variant.sku,
            price: Number(variant.price),
            discount_price: variant.discount_price && variant.discount_price > 0 ? Number(variant.discount_price) : undefined,
            inventory_qty: Number(variant.inventory_quantity),
            attributes: variant.attributes,
            images: variant.images
          })),
          inventory: {
            initial_quantity: Number(productData.inventory_quantity)
          }
        }
      };
      console.log('Request data:', JSON.stringify(requestData, null, 2));

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
              TND: Number(formData.price)
            },
            currency: 'TND'
          },
          inventory: {
            status: "in_stock",
            available: true,
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

        // Add the optimistic product to the list
        addOptimisticProduct(optimisticProduct);

        // Refresh the product list
        await refreshProducts();

        // Show success message
        toast.success("Product created successfully");

        // Redirect to the product list page
        router.push('/products');
      }
    } catch (error: unknown) {
      const err = error as { error?: string; message?: string };
      console.error("Error creating product:", error);
      toast.error(err.error || err.message || "Failed to create product");
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

            <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
              <div>
                <Label htmlFor="product_type">Product Type*</Label>
                <select
                  id="product_type"
                  name="product_type"
                  className="h-11 w-full rounded-lg border appearance-none px-4 py-2.5 text-sm shadow-theme-xs placeholder:text-gray-400 focus:outline-hidden focus:ring-3 dark:bg-gray-900 dark:text-white/90 dark:placeholder:text-white/30 dark:focus:border-brand-800 border-gray-200 focus:border-brand-500 focus:ring-brand-500/20 dark:border-gray-700"
                  value={formData.product_type}
                  onChange={handleInputChange}
                >
                  <option value="physical">Physical Product</option>
                  <option value="digital">Digital Product</option>
                  <option value="service">Service</option>
                  <option value="subscription">Subscription</option>
                </select>
              </div>

              <div>
                <Label htmlFor="weight">Weight (kg)</Label>
                <Input
                  id="weight"
                  name="weight"
                  type="number"
                  step={0.01}
                  min="0"
                  placeholder="0.00"
                  value={formData.weight === 0 && document.activeElement?.id === 'weight' ? '' : formData.weight}
                  onChange={(e) => {
                    // Allow empty field (will show as placeholder) but store as null
                    if (e.target.value === '') {
                      setFormData(prev => ({
                        ...prev,
                        weight: 0
                      }));
                    } else {
                      // Store valid numbers
                      const value = parseFloat(e.target.value);
                      if (!isNaN(value)) {
                        setFormData(prev => ({
                          ...prev,
                          weight: value
                        }));
                      }
                    }
                  }}
                />
              </div>
            </div>

            <div>
              <Label htmlFor="short_description">Short Description</Label>
              <textarea
                id="short_description"
                name="short_description"
                rows={2}
                className="h-auto w-full rounded-lg border appearance-none px-4 py-2.5 text-sm shadow-theme-xs placeholder:text-gray-400 focus:outline-hidden focus:ring-3 dark:bg-gray-900 dark:text-white/90 dark:placeholder:text-white/30 dark:focus:border-brand-800 border-gray-200 focus:border-brand-500 focus:ring-brand-500/20 dark:border-gray-700"
                placeholder="Brief product summary (shown in listings)"
                value={formData.short_description}
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
                <Label htmlFor="price">Regular Price*</Label>
                <Input
                  id="price"
                  name="price"
                  type="number"
                  step={0.01}
                  min="0"
                  placeholder="0.00"
                  value={formData.price === 0 && document.activeElement?.id === 'price' ? '' : formData.price}
                  onChange={(e) => {
                    // Allow empty field (will show as placeholder) but store as null
                    if (e.target.value === '') {
                      setFormData(prev => ({
                        ...prev,
                        price: 0
                      }));
                    } else {
                      // Store valid numbers
                      const value = parseFloat(e.target.value);
                      if (!isNaN(value)) {
                        setFormData(prev => ({
                          ...prev,
                          price: value
                        }));
                      }
                    }
                  }}
                />
              </div>

              <div>
                <Label htmlFor="discount_price">Discount Price</Label>
                <Input
                  id="discount_price"
                  name="discount_price"
                  type="number"
                  step={0.01}
                  min="0"
                  placeholder="0.00"
                  value={formData.discount_price === 0 && document.activeElement?.id === 'discount_price' ? '' : formData.discount_price}
                  onChange={(e) => {
                    // Allow empty field (will show as placeholder) but store as null
                    if (e.target.value === '') {
                      setFormData(prev => ({
                        ...prev,
                        discount_price: 0
                      }));
                    } else {
                      // Store valid numbers
                      const value = parseFloat(e.target.value);
                      if (!isNaN(value)) {
                        setFormData(prev => ({
                          ...prev,
                          discount_price: value
                        }));
                      }
                    }
                  }}
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
                    value={formData.sku}
                    onChange={handleInputChange}
                  />
                  <Button
                    variant="outline"
                    onClick={async () => {
                      if (!formData.brand_id || !formData.categories || formData.categories.length === 0) {
                        toast.error("Brand and category are required to generate SKU");
                        return;
                      }

                      try {
                        // Find the brand name from the selected brand_id
                        const selectedBrand = brands?.find(brand => brand.id === formData.brand_id);
                        if (!selectedBrand) {
                          toast.error("Selected brand not found");
                          return;
                        }

                        // Find the category name from the selected category_id
                        const selectedCategory = categories?.find(category => category.id === formData.categories?.[0]);
                        if (!selectedCategory) {
                          toast.error("Selected category not found");
                          return;
                        }

                        // Call the API to generate a SKU preview
                        const result = await ProductService.generateSKUPreview(
                          selectedBrand.name,
                          selectedCategory.name
                        );

                        // Update the SKU field with the generated SKU
                        setFormData(prev => ({ ...prev, sku: result.sku }));
                        toast.success("SKU generated successfully");
                      } catch (error) {
                        console.error("Failed to generate SKU:", error);
                        toast.error("Failed to generate SKU");
                      }
                    }}
                    disabled={!formData.brand_id || !formData.categories || formData.categories.length === 0}
                  >
                    Generate
                  </Button>
                </div>
              </div>
            </div>

            <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
              <div>
                <Label htmlFor="inventory_quantity">Initial Inventory Quantity*</Label>
                <Input
                  id="inventory_quantity"
                  name="inventory_quantity"
                  type="number"
                  min="0"
                  placeholder="0"
                  value={formData.inventory_quantity === 0 && document.activeElement?.id === 'inventory_quantity' ? '' : formData.inventory_quantity}
                  onChange={(e) => {
                    // Allow empty field (will show as placeholder) but store as null
                    if (e.target.value === '') {
                      setFormData(prev => ({
                        ...prev,
                        inventory_quantity: 0
                      }));
                    } else {
                      // Store valid numbers
                      const value = parseInt(e.target.value);
                      if (!isNaN(value)) {
                        setFormData(prev => ({
                          ...prev,
                          inventory_quantity: value
                        }));
                      }
                    }
                  }}
                />
              </div>

              <div className="bg-blue-50 dark:bg-blue-900/20 p-4 rounded-lg border border-blue-200 dark:border-blue-800">
                <div className="flex items-center mb-2">
                  <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 text-blue-500 mr-2" viewBox="0 0 20 20" fill="currentColor">
                    <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clipRule="evenodd" />
                  </svg>
                  <span className="font-medium text-blue-800 dark:text-blue-200">Inventory Management</span>
                </div>
                <p className="text-sm text-blue-700 dark:text-blue-300">
                  Set the initial inventory quantity for this product. After creation, you can manage detailed inventory settings in the inventory management section.
                </p>
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
                    {brands?.map((brand: Brand) => ( // Use Brand type
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
                  {categories?.map((category: Category) => {
                    const parentCategory = category.parent_id
                      ? categories.find(cat => cat.id === category.parent_id)
                      : null;
                    return (
                      <option key={category.id} value={category.id}>
                        {category.name} {parentCategory ? `(${parentCategory.name})` : ''}
                      </option>
                    );
                  })}
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

          {/* Specifications */}
          <div className="space-y-4 pt-4">
            <div className="flex items-center justify-between">
              <h2 className="text-xl font-semibold text-gray-900 dark:text-white">Product Specifications</h2>
              <Button
                variant="outline"
                size="sm"
                type="button"
                onClick={() => setFormData(prev => ({
                  ...prev,
                  specifications: [...prev.specifications, { name: "", value: "", unit: "" }]
                }))}
              >
                Add Specification
              </Button>
            </div>

            {formData.specifications.map((spec, index) => (
              <div key={index} className="rounded-lg border border-gray-200 p-4 dark:border-gray-700">
                <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
                  <div>
                    <Label htmlFor={`spec-name-${index}`}>Name</Label>
                    <Input
                      id={`spec-name-${index}`}
                      type="text"
                      placeholder="e.g., Processor, Material"
                      value={spec.name}
                      onChange={(e) => {
                        const newSpecs = [...formData.specifications];
                        newSpecs[index].name = e.target.value;
                        setFormData(prev => ({ ...prev, specifications: newSpecs }));
                      }}
                    />
                  </div>
                  <div>
                    <Label htmlFor={`spec-value-${index}`}>Value</Label>
                    <Input
                      id={`spec-value-${index}`}
                      type="text"
                      placeholder="e.g., Intel i7, Cotton"
                      value={spec.value}
                      onChange={(e) => {
                        const newSpecs = [...formData.specifications];
                        newSpecs[index].value = e.target.value;
                        setFormData(prev => ({ ...prev, specifications: newSpecs }));
                      }}
                    />
                  </div>
                  <div>
                    <Label htmlFor={`spec-unit-${index}`}>Unit (optional)</Label>
                    <Input
                      id={`spec-unit-${index}`}
                      type="text"
                      placeholder="e.g., GHz, inches"
                      value={spec.unit}
                      onChange={(e) => {
                        const newSpecs = [...formData.specifications];
                        newSpecs[index].unit = e.target.value;
                        setFormData(prev => ({ ...prev, specifications: newSpecs }));
                      }}
                    />
                  </div>
                </div>
                <div className="mt-4 flex justify-end">
                  <Button
                    variant="outline"
                    size="sm"
                    type="button"
                    className="text-danger-500 hover:border-danger-500 hover:bg-danger-500/10"
                    onClick={() => {
                      const newSpecs = formData.specifications.filter((_, i) => i !== index);
                      setFormData(prev => ({ ...prev, specifications: newSpecs }));
                    }}
                    disabled={formData.specifications.length === 1}
                  >
                    Remove Specification
                  </Button>
                </div>
              </div>
            ))}
          </div>

          {/* Tags */}
          <div className="space-y-4 pt-4">
            <h2 className="text-xl font-semibold text-gray-900 dark:text-white">Product Tags</h2>
            <div className="rounded-lg border border-gray-200 p-4 dark:border-gray-700">
              <Label htmlFor="tags">Tags (comma separated)</Label>
              <Input
                id="tags"
                type="text"
                placeholder="e.g., electronics, smartphone, premium"
                value={formData.tags.join(', ')}
                onChange={(e) => {
                  const tagsString = e.target.value;
                  const tagsArray = tagsString.split(',').map(tag => tag.trim()).filter(tag => tag);
                  setFormData(prev => ({ ...prev, tags: tagsArray }));
                }}
              />
              <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
                Enter tags separated by commas. Tags help customers find your products.
              </p>
            </div>
          </div>

          {/* SEO */}
          <div className="space-y-4 pt-4">
            <h2 className="text-xl font-semibold text-gray-900 dark:text-white">SEO Information</h2>
            <div className="rounded-lg border border-gray-200 p-4 dark:border-gray-700 space-y-4">
              <div>
                <Label htmlFor="seo.meta_title">Meta Title</Label>
                <Input
                  id="seo.meta_title"
                  name="seo.meta_title"
                  type="text"
                  placeholder="SEO title (shown in search results)"
                  value={formData.seo.meta_title}
                  onChange={handleInputChange}
                />
              </div>
              <div>
                <Label htmlFor="seo.meta_description">Meta Description</Label>
                <textarea
                  id="seo.meta_description"
                  name="seo.meta_description"
                  rows={2}
                  className="h-auto w-full rounded-lg border appearance-none px-4 py-2.5 text-sm shadow-theme-xs placeholder:text-gray-400 focus:outline-hidden focus:ring-3 dark:bg-gray-900 dark:text-white/90 dark:placeholder:text-white/30 dark:focus:border-brand-800 border-gray-200 focus:border-brand-500 focus:ring-brand-500/20 dark:border-gray-700"
                  placeholder="SEO description (shown in search results)"
                  value={formData.seo.meta_description}
                  onChange={handleInputChange}
                />
              </div>
              <div>
                <Label htmlFor="keywords">Keywords (comma separated)</Label>
                <Input
                  id="keywords"
                  type="text"
                  placeholder="e.g., premium smartphone, high-resolution camera"
                  value={formData.seo.keywords.join(', ')}
                  onChange={(e) => {
                    const keywordsString = e.target.value;
                    const keywordsArray = keywordsString.split(',').map(keyword => keyword.trim()).filter(keyword => keyword);
                    setFormData(prev => ({
                      ...prev,
                      seo: { ...prev.seo, keywords: keywordsArray }
                    }));
                  }}
                />
              </div>
            </div>
          </div>

          {/* Variants */}
          <div className="space-y-4 pt-4">
            <div className="flex items-center justify-between">
              <h2 className="text-xl font-semibold text-gray-900 dark:text-white">Product Variants</h2>
              <Button
                variant="outline"
                size="sm"
                type="button"
                onClick={() => setFormData(prev => ({
                  ...prev,
                  variants: [...prev.variants, {
                    title: "",
                    sku: "",
                    price: prev.price,
                    discount_price: prev.discount_price,
                    inventory_quantity: 0,
                    attributes: [{ name: "", value: "" }],
                    images: []
                  }]
                }))}
              >
                Add Variant
              </Button>
            </div>

            {formData.variants.length > 0 ? (
              formData.variants.map((variant, variantIndex) => (
                <div key={variantIndex} className="rounded-lg border border-gray-200 p-4 dark:border-gray-700 space-y-4">
                  <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                    <div>
                      <Label htmlFor={`variant-title-${variantIndex}`}>Variant Title</Label>
                      <Input
                        id={`variant-title-${variantIndex}`}
                        type="text"
                        placeholder="e.g., Red - Large"
                        value={variant.title}
                        onChange={(e) => {
                          const newVariants = [...formData.variants];
                          newVariants[variantIndex].title = e.target.value;
                          setFormData(prev => ({ ...prev, variants: newVariants }));
                        }}
                      />
                    </div>
                    <div>
                      <Label htmlFor={`variant-sku-${variantIndex}`}>Variant SKU</Label>
                      <Input
                        id={`variant-sku-${variantIndex}`}
                        type="text"
                        placeholder="e.g., PROD-001-RED-L"
                        value={variant.sku}
                        onChange={(e) => {
                          const newVariants = [...formData.variants];
                          newVariants[variantIndex].sku = e.target.value;
                          setFormData(prev => ({ ...prev, variants: newVariants }));
                        }}
                      />
                    </div>
                  </div>

                  <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
                    <div>
                      <Label htmlFor={`variant-price-${variantIndex}`}>Price</Label>
                      <Input
                        id={`variant-price-${variantIndex}`}
                        type="number"
                        step={0.01}
                        min="0"
                        placeholder="0.00"
                        value={variant.price === 0 && document.activeElement?.id === `variant-price-${variantIndex}` ? '' : variant.price}
                        onChange={(e) => {
                          const newVariants = [...formData.variants];
                          if (e.target.value === '') {
                            newVariants[variantIndex].price = 0;
                          } else {
                            const value = parseFloat(e.target.value);
                            if (!isNaN(value)) {
                              newVariants[variantIndex].price = value;
                            }
                          }
                          setFormData(prev => ({ ...prev, variants: newVariants }));
                        }}
                      />
                    </div>
                    <div>
                      <Label htmlFor={`variant-discount-price-${variantIndex}`}>Discount Price</Label>
                      <Input
                        id={`variant-discount-price-${variantIndex}`}
                        type="number"
                        step={0.01}
                        min="0"
                        placeholder="0.00"
                        value={variant.discount_price === 0 && document.activeElement?.id === `variant-discount-price-${variantIndex}` ? '' : variant.discount_price}
                        onChange={(e) => {
                          const newVariants = [...formData.variants];
                          if (e.target.value === '') {
                            newVariants[variantIndex].discount_price = 0;
                          } else {
                            const value = parseFloat(e.target.value);
                            if (!isNaN(value)) {
                              newVariants[variantIndex].discount_price = value;
                            }
                          }
                          setFormData(prev => ({ ...prev, variants: newVariants }));
                        }}
                      />
                    </div>
                    <div>
                      <Label htmlFor={`variant-inventory-${variantIndex}`}>Inventory Quantity</Label>
                      <Input
                        id={`variant-inventory-${variantIndex}`}
                        type="number"
                        min="0"
                        placeholder="0"
                        value={variant.inventory_quantity === 0 && document.activeElement?.id === `variant-inventory-${variantIndex}` ? '' : variant.inventory_quantity}
                        onChange={(e) => {
                          const newVariants = [...formData.variants];
                          if (e.target.value === '') {
                            newVariants[variantIndex].inventory_quantity = 0;
                          } else {
                            const value = parseInt(e.target.value);
                            if (!isNaN(value)) {
                              newVariants[variantIndex].inventory_quantity = value;
                            }
                          }
                          setFormData(prev => ({ ...prev, variants: newVariants }));
                        }}
                      />
                    </div>
                  </div>

                  {/* Variant Attributes */}
                  <div className="space-y-2">
                    <div className="flex items-center justify-between">
                      <Label>Attributes</Label>
                      <Button
                        variant="outline"
                        size="sm"
                        type="button"
                        onClick={() => {
                          const newVariants = [...formData.variants];
                          newVariants[variantIndex].attributes.push({ name: "", value: "" });
                          setFormData(prev => ({ ...prev, variants: newVariants }));
                        }}
                      >
                        Add Attribute
                      </Button>
                    </div>

                    {variant.attributes.map((attr, attrIndex) => (
                      <div key={attrIndex} className="grid grid-cols-1 gap-2 md:grid-cols-2 border border-gray-100 p-2 rounded dark:border-gray-700">
                        <div>
                          <Label htmlFor={`attr-name-${variantIndex}-${attrIndex}`}>Name</Label>
                          <Input
                            id={`attr-name-${variantIndex}-${attrIndex}`}
                            type="text"
                            placeholder="e.g., Color, Size"
                            value={attr.name}
                            onChange={(e) => {
                              const newVariants = [...formData.variants];
                              newVariants[variantIndex].attributes[attrIndex].name = e.target.value;
                              setFormData(prev => ({ ...prev, variants: newVariants }));
                            }}
                          />
                        </div>
                        <div className="flex gap-2">
                          <div className="flex-grow">
                            <Label htmlFor={`attr-value-${variantIndex}-${attrIndex}`}>Value</Label>
                            <Input
                              id={`attr-value-${variantIndex}-${attrIndex}`}
                              type="text"
                              placeholder="e.g., Red, Large"
                              value={attr.value}
                              onChange={(e) => {
                                const newVariants = [...formData.variants];
                                newVariants[variantIndex].attributes[attrIndex].value = e.target.value;
                                setFormData(prev => ({ ...prev, variants: newVariants }));
                              }}
                            />
                          </div>
                          <div className="flex items-end">
                            <Button
                              variant="outline"
                              size="sm"
                              type="button"
                              className="text-danger-500 hover:border-danger-500 hover:bg-danger-500/10 mb-0.5"
                              onClick={() => {
                                const newVariants = [...formData.variants];
                                newVariants[variantIndex].attributes = newVariants[variantIndex].attributes.filter((_, i) => i !== attrIndex);
                                setFormData(prev => ({ ...prev, variants: newVariants }));
                              }}
                              disabled={variant.attributes.length === 1}
                            >
                              <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M3 6h18"></path><path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6"></path><path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2"></path><line x1="10" y1="11" x2="10" y2="17"></line><line x1="14" y1="11" x2="14" y2="17"></line></svg>
                            </Button>
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>

                  {/* Variant Images */}
                  <div className="space-y-2">
                    <div className="flex items-center justify-between">
                      <Label>Variant Images</Label>
                      <Button
                        variant="outline"
                        size="sm"
                        type="button"
                        onClick={() => {
                          const newVariants = [...formData.variants];
                          newVariants[variantIndex].images.push({
                            url: "",
                            alt_text: `${variant.title || `Variant ${variantIndex + 1}`} image ${newVariants[variantIndex].images.length + 1}`,
                            position: newVariants[variantIndex].images.length + 1
                          });
                          setFormData(prev => ({ ...prev, variants: newVariants }));
                        }}
                      >
                        Add Image
                      </Button>
                    </div>

                    {variant.images.length > 0 ? (
                      variant.images.map((_, imgIndex) => (
                        <div key={imgIndex} className="rounded-lg border border-gray-200 p-4 dark:border-gray-700">
                          <ImageUpload
                            onUploadSuccess={(result) => {
                              const newVariants = [...formData.variants];
                              newVariants[variantIndex].images[imgIndex] = {
                                url: result.url,
                                alt_text: result.alt_text || `${variant.title || `Variant ${variantIndex + 1}`} image ${imgIndex + 1}`,
                                position: result.position || imgIndex + 1
                              };
                              setFormData(prev => ({ ...prev, variants: newVariants }));
                            }}
                            onUploadError={(error) => {
                              toast.error(error);
                            }}
                            folder="variants"
                            defaultAltText={`${variant.title || `Variant ${variantIndex + 1}`} image ${imgIndex + 1}`}
                            defaultPosition={imgIndex + 1}
                          />
                          <div className="mt-4 flex justify-end">
                            <Button
                              variant="outline"
                              size="sm"
                              type="button"
                              className="text-danger-500 hover:border-danger-500 hover:bg-danger-500/10"
                              onClick={() => {
                                const newVariants = [...formData.variants];
                                newVariants[variantIndex].images = newVariants[variantIndex].images.filter((_, i) => i !== imgIndex);
                                setFormData(prev => ({ ...prev, variants: newVariants }));
                              }}
                            >
                              Remove Image
                            </Button>
                          </div>
                        </div>
                      ))
                    ) : (
                      <div className="rounded-lg border border-gray-200 p-4 dark:border-gray-700 text-center text-gray-500 dark:text-gray-400">
                        <p>No variant images added. Add images specific to this variant.</p>
                      </div>
                    )}
                  </div>

                  <div className="flex justify-end">
                    <Button
                      variant="outline"
                      size="sm"
                      type="button"
                      className="text-danger-500 hover:border-danger-500 hover:bg-danger-500/10"
                      onClick={() => {
                        setFormData(prev => ({
                          ...prev,
                          variants: prev.variants.filter((_, i) => i !== variantIndex)
                        }));
                      }}
                    >
                      Remove Variant
                    </Button>
                  </div>
                </div>
              ))
            ) : (
              <div className="rounded-lg border border-gray-200 p-4 dark:border-gray-700 text-center text-gray-500 dark:text-gray-400">
                <p>No variants added. Add variants if this product comes in different options like sizes or colors.</p>
              </div>
            )}
          </div>

          {/* Shipping */}
          <div className="space-y-4 pt-4">
            <h2 className="text-xl font-semibold text-gray-900 dark:text-white">Shipping Information</h2>
            <div className="rounded-lg border border-gray-200 p-4 dark:border-gray-700 space-y-4">
              <div className="flex items-center">
                <input
                  id="free_shipping"
                  type="checkbox"
                  className="h-4 w-4 rounded border-gray-300 text-brand-500 focus:ring-brand-500"
                  checked={formData.shipping.free_shipping}
                  onChange={(e) => {
                    setFormData(prev => ({
                      ...prev,
                      shipping: { ...prev.shipping, free_shipping: e.target.checked }
                    }));
                  }}
                />
                <label htmlFor="free_shipping" className="ml-2 text-sm text-gray-700 dark:text-gray-300">
                  Free Shipping
                </label>
              </div>
              <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                <div>
                  <Label htmlFor="shipping.estimated_days">Estimated Delivery (days)</Label>
                  <Input
                    id="shipping.estimated_days"
                    name="shipping.estimated_days"
                    type="number"
                    min="1"
                    placeholder="3"
                    value={formData.shipping.estimated_days}
                    onChange={(e) => {
                      const value = parseInt(e.target.value);
                      if (!isNaN(value)) {
                        setFormData(prev => ({
                          ...prev,
                          shipping: { ...prev.shipping, estimated_days: value }
                        }));
                      }
                    }}
                  />
                </div>
                <div className="flex items-center">
                  <input
                    id="express_available"
                    type="checkbox"
                    className="h-4 w-4 rounded border-gray-300 text-brand-500 focus:ring-brand-500"
                    checked={formData.shipping.express_available}
                    onChange={(e) => {
                      setFormData(prev => ({
                        ...prev,
                        shipping: { ...prev.shipping, express_available: e.target.checked }
                      }));
                    }}
                  />
                  <label htmlFor="express_available" className="ml-2 text-sm text-gray-700 dark:text-gray-300">
                    Express Shipping Available
                  </label>
                </div>
              </div>
            </div>
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
