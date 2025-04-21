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
    };
    currency: string;
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

export class ProductService {
  static async getProducts(page = 1, limit = 10, filters = {}): Promise<ProductListResponse> {
    try {
      const params = { page, limit, ...filters };
      console.log('Fetching products with params:', params);
      const response = await api.get('/products', { params });
      console.log('API response:', response.data);

      // Ensure we have a valid response structure
      const data = response.data || {};
      return {
        products: data.products || [],
        total: data.total || 0,
        pagination: data.pagination || {
          current_page: page,
          total_pages: 1,
          per_page: limit,
          total_items: 0
        }
      };
    } catch (error: any) {
      console.error('Error fetching products:', error.response?.data || error);
      throw error.response?.data || { error: 'Failed to fetch products' };
    }
  }

  static async getProduct(id: string): Promise<Product> {
    try {
      const response = await api.get(`/products/${id}`);
      return response.data;
    } catch (error: any) {
      throw error.response?.data || { error: 'Failed to fetch product' };
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
          brand_id: productData.brand_id ? { value: productData.brand_id } : undefined,
          category_ids: productData.categories ? productData.categories.map(id => ({ value: id })) : undefined,
          is_published: productData.is_published
        }
      };

      console.log('Creating product with data:', JSON.stringify(transformedData, null, 2));
      // IMPORTANT: We're not using this service method directly anymore
      // The form component handles the API call directly with proper formatting
      // This is kept for reference and potential future use
      const response = await api.post('/products', transformedData);
      return response.data;
    } catch (error: any) {
      console.error('Error creating product:', error.response?.data || error);
      throw error.response?.data || { message: 'Failed to create product' };
    }
  }

  static async updateProduct(id: string, productData: Partial<ProductCreateRequest>): Promise<Product> {
    try {
      const response = await api.put(`/products/${id}`, { product: productData });
      return response.data;
    } catch (error: any) {
      throw error.response?.data || { error: 'Failed to update product' };
    }
  }

  static async deleteProduct(id: string): Promise<{ success: boolean }> {
    try {
      const response = await api.delete(`/products/${id}`);
      return response.data;
    } catch (error: any) {
      throw error.response?.data || { error: 'Failed to delete product' };
    }
  }

  static async getBrands(page = 1, limit = 10): Promise<any> {
    try {
      const params = { page, limit };
      const response = await api.get('/brands', { params });
      return response.data;
    } catch (error: any) {
      throw error.response?.data || { error: 'Failed to fetch brands' };
    }
  }

  static async createBrand(brandData: any): Promise<any> {
    try {
      const response = await api.post('/brands', { brand: brandData });
      return response.data;
    } catch (error: any) {
      throw error.response?.data || { error: 'Failed to create brand' };
    }
  }

  static async updateBrand(id: string, brandData: any): Promise<any> {
    try {
      const response = await api.put(`/brands/${id}`, { brand: brandData });
      return response.data;
    } catch (error: any) {
      throw error.response?.data || { error: 'Failed to update brand' };
    }
  }

  static async deleteBrand(id: string): Promise<{ success: boolean }> {
    try {
      const response = await api.delete(`/brands/${id}`);
      return response.data;
    } catch (error: any) {
      throw error.response?.data || { error: 'Failed to delete brand' };
    }
  }

  static async getCategories(page = 1, limit = 10): Promise<any> {
    try {
      const params = { page, limit };
      const response = await api.get('/categories', { params });
      return response.data;
    } catch (error: any) {
      throw error.response?.data || { error: 'Failed to fetch categories' };
    }
  }

  static async createCategory(categoryData: any): Promise<any> {
    try {
      const response = await api.post('/categories', { category: categoryData });
      return response.data;
    } catch (error: any) {
      throw error.response?.data || { error: 'Failed to create category' };
    }
  }

  static async updateCategory(id: string, categoryData: any): Promise<any> {
    try {
      const response = await api.put(`/categories/${id}`, { category: categoryData });
      return response.data;
    } catch (error: any) {
      throw error.response?.data || { error: 'Failed to update category' };
    }
  }

  static async deleteCategory(id: string): Promise<{ success: boolean }> {
    try {
      const response = await api.delete(`/categories/${id}`);
      return response.data;
    } catch (error: any) {
      throw error.response?.data || { error: 'Failed to delete category' };
    }
  }
}
