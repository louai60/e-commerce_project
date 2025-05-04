import { api } from '@/lib/api';

export interface ProductImage {
  id: string;
  url: string;
  alt_text: string;
  position: number;
}

export interface ProductVariant {
  id: string;
  sku: string;
  price: number;
  inventory_qty: number;
  attributes: {
    name: string;
    value: string;
  }[];
}

export interface ProductCategory {
  id: string;
  name: string;
  slug: string;
}

export interface ProductBrand {
  id: string;
  name: string;
  slug: string;
}

export interface Product {
  id: string;
  title: string;
  slug: string;
  description: string;
  short_description: string;
  sku: string;
  price: {
    current: {
      USD: number;
      EUR?: number;
      USD_DISCOUNT?: number;
      EUR_DISCOUNT?: number;
    };
    currency: string;
    value?: number;
    savings_percentage?: number;
  };
  inventory: {
    status: string;
    available: boolean;
    quantity: number;
  };
  images: ProductImage[];
  variants: ProductVariant[];
  categories: ProductCategory[];
  brand?: ProductBrand;
  is_published: boolean;
  created_at: string;
  updated_at: string;
}

export interface ProductListResponse {
  products: Product[];
  total: number;
  pagination: {
    current_page: number;
    total_pages: number;
    per_page: number;
    total_items: number;
  };
}

export interface ProductCreateRequest {
  title: string;
  slug: string;
  description: string;
  short_description: string;
  price: number;
  sku: string;
  inventory_qty: number;
  inventory_status: string;
  is_published: boolean;
  brand_id?: string;
  categories?: string[];
  images: {
    url: string;
    alt_text: string;
    position: number;
  }[];
}

export interface Brand {
  id: string;
  name: string;
  slug: string;
  description?: string;
  logo?: string;
  created_at: string;
  updated_at: string;
}

export interface BrandListResponse {
  brands: Brand[];
  total: number;
  pagination: {
    current_page: number;
    total_pages: number;
    per_page: number;
    total_items: number;
  };
}

export interface Category {
  id: string;
  name: string;
  slug: string;
  description?: string;
  parent_id?: string;
  created_at: string;
  updated_at: string;
}

export interface CategoryListResponse {
  categories: Category[];
  total: number;
  pagination: {
    current_page: number;
    total_pages: number;
    per_page: number;
    total_items: number;
  };
}

interface SKUPreviewParams {
  brand: string;
  category: string;
  color?: string;
  size?: string;
}

export class ProductService {
  static async getProducts(page = 1, limit = 10, filters = {}): Promise<ProductListResponse> {
    try {
      const params = { page, limit, ...filters };
      console.log('Fetching products with params:', params);
      const response = await api.get('/products', { params });
      console.log('API response:', response.data);

      // Add more detailed logging for debugging
      if (response.data && response.data.products) {
        console.log('Number of products:', response.data.products.length);
        if (response.data.products.length > 0) {
          const firstProduct = response.data.products[0];
          console.log('First product sample:', {
            id: firstProduct.id,
            title: firstProduct.title,
            price: firstProduct.price,
            images: firstProduct.images,
            inventory: firstProduct.inventory
          });
        }
      }

      // Ensure we have a valid response structure
      const data = response.data || {};

      // Create a properly structured response
      const result = {
        products: data.products || [],
        total: data.total || 0,
        pagination: data.pagination || {
          current_page: page,
          total_pages: 1,
          per_page: limit,
          total_items: 0
        }
      };

      console.log('Returning structured product data:', result);
      return result;
    } catch (error: unknown) {
      const err = error as { response?: { data?: unknown } };
      console.error('Error fetching products:', err.response?.data || error);
      throw err.response?.data || { error: 'Failed to fetch products' };
    }
  }

  static async getProduct(id: string): Promise<Product> {
    try {
      console.log(`Fetching product with ID: ${id}`);
      const response = await api.get(`/products/${id}`);
      console.log('Product API response:', JSON.stringify(response.data, null, 2));

      // Add detailed logging for debugging
      if (response.data) {
        console.log('Product details:', {
          id: response.data.id,
          title: response.data.title,
          price: response.data.price,
          images: response.data.images,
          inventory: response.data.inventory,
          specifications: response.data.specifications,
          tags: response.data.tags
        });

        // Check for missing or empty data
        const missingData = [];
        if (!response.data.images || response.data.images.length === 0) missingData.push('images');
        if (!response.data.price || !response.data.price.current || !response.data.price.current.USD) missingData.push('price');
        if (!response.data.inventory || response.data.inventory.quantity === 0) missingData.push('inventory');
        if (!response.data.specifications || response.data.specifications.length === 0) missingData.push('specifications');
        if (!response.data.tags || response.data.tags.length === 0) missingData.push('tags');

        if (missingData.length > 0) {
          console.warn('Missing or empty data in product response:', missingData.join(', '));
        }
      }

      return response.data;
    } catch (error: unknown) {
      const err = error as { response?: { data?: unknown } };
      console.error('Error fetching product:', err.response?.data || error);
      throw err.response?.data || { error: 'Failed to fetch product' };
    }
  }

  static async createProduct(productData: ProductCreateRequest): Promise<Product> {
    try {
      // Transform the data to match backend expectations
      const transformedData = {
        product: {
          title: productData.title,
          slug: productData.slug,
          description: productData.description,
          short_description: productData.short_description,
          price: {
            current: {
              USD: Number(productData.price) // Ensure it's a number
            },
            currency: 'USD'
          },
          sku: productData.sku,
          inventory: {
            status: productData.inventory_status,
            quantity: Number(productData.inventory_qty), // Ensure it's a number
            available: productData.inventory_status === 'in_stock'
          },
          images: productData.images,
          brand_id: productData.brand_id || undefined,
          category_ids: productData.categories || undefined,
          is_published: productData.is_published
        }
      };

      console.log('Creating product with data:', JSON.stringify(transformedData, null, 2));
      const response = await api.post('/products', transformedData);
      return response.data;
    } catch (error: unknown) {
      const err = error as { response?: { data?: unknown } };
      console.error('Error creating product:', err.response?.data || error);
      throw err.response?.data || { message: 'Failed to create product' };
    }
  }

  static async updateProduct(id: string, productData: Partial<ProductCreateRequest>): Promise<Product> {
    try {
      const response = await api.put(`/products/${id}`, { product: productData });
      return response.data;
    } catch (error: unknown) {
      const err = error as { response?: { data?: unknown } };
      throw err.response?.data || { error: 'Failed to update product' };
    }
  }

  static async deleteProduct(id: string): Promise<{ success: boolean }> {
    try {
      const response = await api.delete(`/products/${id}`);
      return response.data;
    } catch (error: unknown) {
      const err = error as { response?: { data?: unknown } };
      throw err.response?.data || { error: 'Failed to delete product' };
    }
  }

  static async getBrands(page = 1, limit = 10): Promise<BrandListResponse> {
    try {
      const params = { page, limit };
      const response = await api.get('/brands', { params });
      return response.data;
    } catch (error: unknown) {
      const err = error as { response?: { data?: unknown } };
      throw err.response?.data || { error: 'Failed to fetch brands' };
    }
  }

  static async createBrand(brandData: Omit<Brand, 'id' | 'created_at' | 'updated_at'>): Promise<Brand> {
    try {
      const response = await api.post('/brands', { brand: brandData });
      return response.data;
    } catch (error: unknown) {
      const err = error as { response?: { data?: unknown } };
      throw err.response?.data || { error: 'Failed to create brand' };
    }
  }

  static async updateBrand(id: string, brandData: Partial<Omit<Brand, 'id' | 'created_at' | 'updated_at'>>): Promise<Brand> {
    try {
      const response = await api.put(`/brands/${id}`, { brand: brandData });
      return response.data;
    } catch (error: unknown) {
      const err = error as { response?: { data?: unknown } };
      throw err.response?.data || { error: 'Failed to update brand' };
    }
  }

  static async deleteBrand(id: string): Promise<{ success: boolean }> {
    try {
      const response = await api.delete(`/brands/${id}`);
      return response.data;
    } catch (error: unknown) {
      const err = error as { response?: { data?: unknown } };
      throw err.response?.data || { error: 'Failed to delete brand' };
    }
  }

  static async getCategories(page = 1, limit = 10): Promise<CategoryListResponse> {
    try {
      const params = { page, limit };
      const response = await api.get('/categories', { params });
      return response.data;
    } catch (error: unknown) {
      const err = error as { response?: { data?: unknown } };
      throw err.response?.data || { error: 'Failed to fetch categories' };
    }
  }

  static async createCategory(categoryData: Omit<Category, 'id' | 'created_at' | 'updated_at'>): Promise<Category> {
    try {
      const response = await api.post('/categories', { category: categoryData });
      return response.data;
    } catch (error: unknown) {
      const err = error as { response?: { data?: unknown } };
      throw err.response?.data || { error: 'Failed to create category' };
    }
  }

  static async updateCategory(id: string, categoryData: Partial<Omit<Category, 'id' | 'created_at' | 'updated_at'>>): Promise<Category> {
    try {
      const response = await api.put(`/categories/${id}`, { category: categoryData });
      return response.data;
    } catch (error: unknown) {
      const err = error as { response?: { data?: unknown } };
      throw err.response?.data || { error: 'Failed to update category' };
    }
  }

  static async deleteCategory(id: string): Promise<{ success: boolean }> {
    try {
      const response = await api.delete(`/categories/${id}`);
      return response.data;
    } catch (error: unknown) {
      const err = error as { response?: { data?: unknown } };
      throw err.response?.data || { error: 'Failed to delete category' };
    }
  }

  static async generateSKUPreview(brandName: string, categoryName: string, color?: string, size?: string): Promise<{ sku: string }> {
    try {
      const params: SKUPreviewParams = {
        brand: brandName,
        category: categoryName,
        ...(color ? { color } : {}),
        ...(size ? { size } : {})
      };
      const response = await api.get('/products/sku/preview', { params });
      return response.data;
    } catch (error: unknown) {
      const err = error as { response?: { data?: unknown } };
      throw err.response?.data || { error: 'Failed to generate SKU preview' };
    }
  }
}
