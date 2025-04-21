# NexCart API Integration Guide for Frontend Developers

This guide provides practical examples and best practices for integrating the NexCart API into your frontend application.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Authentication](#authentication)
3. [API Client Setup](#api-client-setup)
4. [Working with Products](#working-with-products)
5. [User Management](#user-management)
6. [Cart Integration](#cart-integration)
7. [Error Handling](#error-handling)
8. [Performance Optimization](#performance-optimization)

## Getting Started

### Prerequisites

- Node.js (v18 or higher)
- npm or yarn
- A modern frontend framework (React, Vue, Angular, etc.)

### Environment Setup

Create a `.env` file in your project root with the following variables:

```
REACT_APP_API_BASE_URL=http://localhost:8080/api/v1
```

For production:

```
REACT_APP_API_BASE_URL=https://api.nexcart.com/api/v1
```

## Authentication

### Setting Up Authentication

Use the following code to handle authentication in your frontend application:

```javascript
// auth.js
import axios from 'axios';

const API_URL = process.env.REACT_APP_API_BASE_URL;

// Login function
export const login = async (email, password) => {
  try {
    const response = await axios.post(`${API_URL}/users/login`, {
      email,
      password
    });
    
    // Store tokens in localStorage or secure cookie
    localStorage.setItem('token', response.data.token);
    localStorage.setItem('refreshToken', response.data.refresh_token);
    localStorage.setItem('user', JSON.stringify(response.data.user));
    
    return response.data;
  } catch (error) {
    throw error.response?.data || { error: 'Login failed' };
  }
};

// Register function
export const register = async (userData) => {
  try {
    const response = await axios.post(`${API_URL}/users/register`, userData);
    
    // Store tokens in localStorage or secure cookie
    localStorage.setItem('token', response.data.token);
    localStorage.setItem('refreshToken', response.data.refresh_token);
    localStorage.setItem('user', JSON.stringify(response.data.user));
    
    return response.data;
  } catch (error) {
    throw error.response?.data || { error: 'Registration failed' };
  }
};

// Logout function
export const logout = () => {
  localStorage.removeItem('token');
  localStorage.removeItem('refreshToken');
  localStorage.removeItem('user');
  
  // Optionally call the logout endpoint
  axios.post(`${API_URL}/users/logout`);
};

// Get current user
export const getCurrentUser = () => {
  const userStr = localStorage.getItem('user');
  if (!userStr) return null;
  
  return JSON.parse(userStr);
};

// Check if user is authenticated
export const isAuthenticated = () => {
  return !!localStorage.getItem('token');
};

// Refresh token
export const refreshToken = async () => {
  try {
    const refreshToken = localStorage.getItem('refreshToken');
    if (!refreshToken) throw new Error('No refresh token available');
    
    const response = await axios.post(`${API_URL}/users/refresh`, {
      refresh_token: refreshToken
    });
    
    localStorage.setItem('token', response.data.token);
    localStorage.setItem('refreshToken', response.data.refresh_token);
    
    return response.data.token;
  } catch (error) {
    // If refresh fails, log the user out
    logout();
    throw error;
  }
};
```

### Authentication Hook (React)

```javascript
// useAuth.js
import { useState, useEffect, createContext, useContext } from 'react';
import { login, register, logout, getCurrentUser, isAuthenticated } from './auth';

const AuthContext = createContext(null);

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);
  
  useEffect(() => {
    // Check if user is already logged in
    const user = getCurrentUser();
    setUser(user);
    setLoading(false);
  }, []);
  
  const loginUser = async (email, password) => {
    const data = await login(email, password);
    setUser(data.user);
    return data;
  };
  
  const registerUser = async (userData) => {
    const data = await register(userData);
    setUser(data.user);
    return data;
  };
  
  const logoutUser = () => {
    logout();
    setUser(null);
  };
  
  return (
    <AuthContext.Provider value={{ user, loading, login: loginUser, register: registerUser, logout: logoutUser, isAuthenticated: !!user }}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => useContext(AuthContext);
```

## API Client Setup

### Setting Up Axios with Interceptors

```javascript
// apiClient.js
import axios from 'axios';
import { refreshToken, logout } from './auth';

const API_URL = process.env.REACT_APP_API_BASE_URL;

const apiClient = axios.create({
  baseURL: API_URL,
  headers: {
    'Content-Type': 'application/json'
  }
});

// Request interceptor
apiClient.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers['Authorization'] = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor
apiClient.interceptors.response.use(
  (response) => {
    return response;
  },
  async (error) => {
    const originalRequest = error.config;
    
    // If the error is 401 and we haven't already tried to refresh the token
    if (error.response.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;
      
      try {
        // Try to refresh the token
        const newToken = await refreshToken();
        
        // Update the header with the new token
        originalRequest.headers['Authorization'] = `Bearer ${newToken}`;
        
        // Retry the original request
        return apiClient(originalRequest);
      } catch (refreshError) {
        // If refresh fails, log the user out
        logout();
        return Promise.reject(refreshError);
      }
    }
    
    return Promise.reject(error);
  }
);

export default apiClient;
```

## Working with Products

### Fetching Products

```javascript
// productService.js
import apiClient from './apiClient';

export const getProducts = async (page = 1, limit = 10, filters = {}) => {
  try {
    const params = { page, limit, ...filters };
    const response = await apiClient.get('/products', { params });
    return response.data;
  } catch (error) {
    throw error.response?.data || { error: 'Failed to fetch products' };
  }
};

export const getProduct = async (idOrSlug) => {
  try {
    const response = await apiClient.get(`/products/${idOrSlug}`);
    return response.data;
  } catch (error) {
    throw error.response?.data || { error: 'Failed to fetch product' };
  }
};

export const createProduct = async (productData) => {
  try {
    const response = await apiClient.post('/products', { product: productData });
    return response.data;
  } catch (error) {
    throw error.response?.data || { error: 'Failed to create product' };
  }
};

export const updateProduct = async (id, productData) => {
  try {
    const response = await apiClient.put(`/products/${id}`, { product: productData });
    return response.data;
  } catch (error) {
    throw error.response?.data || { error: 'Failed to update product' };
  }
};

export const deleteProduct = async (id) => {
  try {
    const response = await apiClient.delete(`/products/${id}`);
    return response.data;
  } catch (error) {
    throw error.response?.data || { error: 'Failed to delete product' };
  }
};
```

### Product Hooks (React with SWR)

```javascript
// useProducts.js
import useSWR from 'swr';
import apiClient from './apiClient';

const fetcher = (url) => apiClient.get(url).then(res => res.data);

export function useProducts(page = 1, limit = 10, filters = {}) {
  const params = new URLSearchParams({ page, limit, ...filters }).toString();
  const { data, error, isLoading, mutate } = useSWR(
    `/products?${params}`,
    fetcher,
    { revalidateOnMount: true }
  );

  return {
    products: data?.products || [],
    total: data?.total || 0,
    pagination: data?.pagination || {},
    isLoading,
    isError: error,
    mutate,
  };
}

export function useProduct(idOrSlug) {
  const { data, error, isLoading, mutate } = useSWR(
    idOrSlug ? `/products/${idOrSlug}` : null,
    fetcher
  );

  return {
    product: data,
    isLoading,
    isError: error,
    mutate,
  };
}
```

### Product Components (React)

```jsx
// ProductList.jsx
import React, { useState } from 'react';
import { useProducts } from '../hooks/useProducts';
import ProductCard from './ProductCard';
import Pagination from './Pagination';

const ProductList = () => {
  const [page, setPage] = useState(1);
  const [filters, setFilters] = useState({});
  const { products, total, pagination, isLoading, isError } = useProducts(page, 10, filters);

  if (isLoading) return <div>Loading...</div>;
  if (isError) return <div>Error loading products</div>;

  return (
    <div>
      <div className="filters">
        {/* Filter components */}
      </div>
      
      <div className="product-grid">
        {products.map(product => (
          <ProductCard key={product.id} product={product} />
        ))}
      </div>
      
      <Pagination
        currentPage={pagination.current_page}
        totalPages={pagination.total_pages}
        onPageChange={setPage}
      />
    </div>
  );
};

export default ProductList;
```

```jsx
// ProductDetail.jsx
import React from 'react';
import { useParams } from 'react-router-dom';
import { useProduct } from '../hooks/useProducts';
import { useCart } from '../hooks/useCart';

const ProductDetail = () => {
  const { id } = useParams();
  const { product, isLoading, isError } = useProduct(id);
  const { addToCart } = useCart();

  if (isLoading) return <div>Loading...</div>;
  if (isError) return <div>Error loading product</div>;
  if (!product) return <div>Product not found</div>;

  const handleAddToCart = () => {
    addToCart({
      id: product.id,
      title: product.title,
      price: product.price.current.USD,
      discountedPrice: product.price.current.USD, // Apply any discounts if available
      quantity: 1,
      imgs: {
        thumbnails: product.images.map(img => img.sizes.thumbnail),
        previews: product.images.map(img => img.url),
      }
    });
  };

  return (
    <div className="product-detail">
      <div className="product-images">
        {product.images.map(image => (
          <img key={image.id} src={image.url} alt={image.alt} />
        ))}
      </div>
      
      <div className="product-info">
        <h1>{product.title}</h1>
        <p className="price">${product.price.current.USD}</p>
        <div className="description">{product.description}</div>
        
        <div className="inventory">
          <span className={`status ${product.inventory.available ? 'in-stock' : 'out-of-stock'}`}>
            {product.inventory.status}
          </span>
        </div>
        
        <button 
          onClick={handleAddToCart}
          disabled={!product.inventory.available}
        >
          Add to Cart
        </button>
      </div>
      
      <div className="product-details">
        <h2>Specifications</h2>
        <ul>
          {Object.entries(product.specifications).map(([key, value]) => (
            <li key={key}>
              <strong>{key}:</strong> {value}
            </li>
          ))}
        </ul>
      </div>
      
      <div className="product-reviews">
        <h2>Reviews ({product.reviews.summary.total_reviews})</h2>
        <div className="rating">
          Average Rating: {product.reviews.summary.average_rating}
        </div>
        
        {product.reviews.items.map(review => (
          <div key={review.id} className="review">
            <h3>{review.title}</h3>
            <div className="rating">{review.rating} / 5</div>
            <div className="author">By {review.user.name}</div>
            <p>{review.comment}</p>
          </div>
        ))}
      </div>
    </div>
  );
};

export default ProductDetail;
```

## User Management

### User Profile Component (React)

```jsx
// UserProfile.jsx
import React, { useState } from 'react';
import { useAuth } from '../hooks/useAuth';
import apiClient from '../utils/apiClient';

const UserProfile = () => {
  const { user, login } = useAuth();
  const [formData, setFormData] = useState({
    username: user?.username || '',
    first_name: user?.first_name || '',
    last_name: user?.last_name || '',
    phone_number: user?.phone_number || ''
  });
  const [isEditing, setIsEditing] = useState(false);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(null);

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData(prev => ({ ...prev, [name]: value }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError(null);
    setSuccess(null);
    
    try {
      const response = await apiClient.put('/users/profile', formData);
      
      // Update the user in context
      localStorage.setItem('user', JSON.stringify(response.data.user));
      login(response.data.user);
      
      setSuccess('Profile updated successfully');
      setIsEditing(false);
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to update profile');
    }
  };

  if (!user) return <div>Please log in to view your profile</div>;

  return (
    <div className="user-profile">
      <h1>User Profile</h1>
      
      {error && <div className="error">{error}</div>}
      {success && <div className="success">{success}</div>}
      
      {isEditing ? (
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label>Username</label>
            <input
              type="text"
              name="username"
              value={formData.username}
              onChange={handleChange}
              required
            />
          </div>
          
          <div className="form-group">
            <label>First Name</label>
            <input
              type="text"
              name="first_name"
              value={formData.first_name}
              onChange={handleChange}
              required
            />
          </div>
          
          <div className="form-group">
            <label>Last Name</label>
            <input
              type="text"
              name="last_name"
              value={formData.last_name}
              onChange={handleChange}
              required
            />
          </div>
          
          <div className="form-group">
            <label>Phone Number</label>
            <input
              type="text"
              name="phone_number"
              value={formData.phone_number}
              onChange={handleChange}
            />
          </div>
          
          <div className="form-actions">
            <button type="submit">Save Changes</button>
            <button type="button" onClick={() => setIsEditing(false)}>Cancel</button>
          </div>
        </form>
      ) : (
        <div className="profile-info">
          <p><strong>Email:</strong> {user.email}</p>
          <p><strong>Username:</strong> {user.username}</p>
          <p><strong>Name:</strong> {user.first_name} {user.last_name}</p>
          <p><strong>Phone:</strong> {user.phone_number || 'Not provided'}</p>
          <p><strong>Account Type:</strong> {user.user_type}</p>
          <p><strong>Role:</strong> {user.role}</p>
          <p><strong>Account Status:</strong> {user.account_status}</p>
          <p><strong>Member Since:</strong> {new Date(user.created_at).toLocaleDateString()}</p>
          
          <button onClick={() => setIsEditing(true)}>Edit Profile</button>
        </div>
      )}
    </div>
  );
};

export default UserProfile;
```

## Cart Integration

### Cart Hook (React with Redux)

```javascript
// useCart.js
import { useSelector, useDispatch } from 'react-redux';
import { 
  addItemToCart, 
  removeItemFromCart, 
  updateCartItemQuantity, 
  removeAllItemsFromCart,
  selectCartItems,
  selectTotalPrice
} from '../redux/features/cart-slice';

export const useCart = () => {
  const dispatch = useDispatch();
  const cartItems = useSelector(selectCartItems);
  const totalPrice = useSelector(selectTotalPrice);
  
  const addToCart = (item) => {
    dispatch(addItemToCart(item));
  };
  
  const removeFromCart = (itemId) => {
    dispatch(removeItemFromCart(itemId));
  };
  
  const updateQuantity = (itemId, quantity) => {
    dispatch(updateCartItemQuantity({ id: itemId, quantity }));
  };
  
  const clearCart = () => {
    dispatch(removeAllItemsFromCart());
  };
  
  return {
    cartItems,
    totalPrice,
    addToCart,
    removeFromCart,
    updateQuantity,
    clearCart,
    itemCount: cartItems.length
  };
};
```

### Cart Components (React)

```jsx
// CartPage.jsx
import React from 'react';
import { useCart } from '../hooks/useCart';
import { Link } from 'react-router-dom';

const CartPage = () => {
  const { cartItems, totalPrice, removeFromCart, updateQuantity, clearCart } = useCart();

  if (cartItems.length === 0) {
    return (
      <div className="empty-cart">
        <h1>Your Cart is Empty</h1>
        <p>Looks like you haven't added any products to your cart yet.</p>
        <Link to="/products" className="btn">Continue Shopping</Link>
      </div>
    );
  }

  return (
    <div className="cart-page">
      <h1>Your Cart</h1>
      
      <div className="cart-items">
        {cartItems.map(item => (
          <div key={item.id} className="cart-item">
            <div className="item-image">
              {item.imgs && item.imgs.thumbnails && (
                <img src={item.imgs.thumbnails[0]} alt={item.title} />
              )}
            </div>
            
            <div className="item-details">
              <h3>{item.title}</h3>
              <p className="price">${item.discountedPrice}</p>
            </div>
            
            <div className="item-quantity">
              <button 
                onClick={() => updateQuantity(item.id, Math.max(1, item.quantity - 1))}
                disabled={item.quantity <= 1}
              >
                -
              </button>
              <span>{item.quantity}</span>
              <button onClick={() => updateQuantity(item.id, item.quantity + 1)}>
                +
              </button>
            </div>
            
            <div className="item-total">
              ${(item.discountedPrice * item.quantity).toFixed(2)}
            </div>
            
            <button 
              className="remove-item" 
              onClick={() => removeFromCart(item.id)}
            >
              Remove
            </button>
          </div>
        ))}
      </div>
      
      <div className="cart-summary">
        <div className="cart-total">
          <span>Total:</span>
          <span>${totalPrice.toFixed(2)}</span>
        </div>
        
        <div className="cart-actions">
          <button className="clear-cart" onClick={clearCart}>
            Clear Cart
          </button>
          <Link to="/checkout" className="checkout-btn">
            Proceed to Checkout
          </Link>
        </div>
      </div>
    </div>
  );
};

export default CartPage;
```

## Error Handling

### Error Boundary Component (React)

```jsx
// ErrorBoundary.jsx
import React, { Component } from 'react';

class ErrorBoundary extends Component {
  constructor(props) {
    super(props);
    this.state = { hasError: false, error: null, errorInfo: null };
  }

  static getDerivedStateFromError(error) {
    return { hasError: true };
  }

  componentDidCatch(error, errorInfo) {
    this.setState({
      error: error,
      errorInfo: errorInfo
    });
    
    // You can also log the error to an error reporting service
    console.error("Error caught by ErrorBoundary:", error, errorInfo);
  }

  render() {
    if (this.state.hasError) {
      // You can render any custom fallback UI
      return (
        <div className="error-boundary">
          <h1>Something went wrong.</h1>
          <p>We apologize for the inconvenience. Please try refreshing the page or contact support if the problem persists.</p>
          <button onClick={() => window.location.reload()}>
            Refresh Page
          </button>
        </div>
      );
    }

    return this.props.children;
  }
}

export default ErrorBoundary;
```

### API Error Handling

```javascript
// errorHandler.js
export const handleApiError = (error) => {
  if (error.response) {
    // The request was made and the server responded with a status code
    // that falls out of the range of 2xx
    const { status, data } = error.response;
    
    switch (status) {
      case 400:
        return {
          type: 'VALIDATION_ERROR',
          message: data.error || 'Invalid request',
          details: data.details || null
        };
      case 401:
        return {
          type: 'AUTHENTICATION_ERROR',
          message: 'Authentication required'
        };
      case 403:
        return {
          type: 'PERMISSION_ERROR',
          message: 'You do not have permission to perform this action'
        };
      case 404:
        return {
          type: 'NOT_FOUND',
          message: 'The requested resource was not found'
        };
      case 500:
      case 502:
      case 503:
      case 504:
        return {
          type: 'SERVER_ERROR',
          message: 'Server error, please try again later'
        };
      default:
        return {
          type: 'UNKNOWN_ERROR',
          message: data.error || 'An unknown error occurred'
        };
    }
  } else if (error.request) {
    // The request was made but no response was received
    return {
      type: 'NETWORK_ERROR',
      message: 'Network error, please check your connection'
    };
  } else {
    // Something happened in setting up the request that triggered an Error
    return {
      type: 'REQUEST_ERROR',
      message: error.message || 'Error setting up request'
    };
  }
};
```

## Performance Optimization

### Lazy Loading Components (React)

```jsx
// App.jsx
import React, { lazy, Suspense } from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { AuthProvider } from './hooks/useAuth';
import ErrorBoundary from './components/ErrorBoundary';
import Header from './components/Header';
import Footer from './components/Footer';
import Loading from './components/Loading';

// Lazy load components
const Home = lazy(() => import('./pages/Home'));
const ProductList = lazy(() => import('./pages/ProductList'));
const ProductDetail = lazy(() => import('./pages/ProductDetail'));
const Cart = lazy(() => import('./pages/Cart'));
const Checkout = lazy(() => import('./pages/Checkout'));
const Login = lazy(() => import('./pages/Login'));
const Register = lazy(() => import('./pages/Register'));
const UserProfile = lazy(() => import('./pages/UserProfile'));

const App = () => {
  return (
    <ErrorBoundary>
      <AuthProvider>
        <Router>
          <Header />
          <Suspense fallback={<Loading />}>
            <Routes>
              <Route path="/" element={<Home />} />
              <Route path="/products" element={<ProductList />} />
              <Route path="/products/:id" element={<ProductDetail />} />
              <Route path="/cart" element={<Cart />} />
              <Route path="/checkout" element={<Checkout />} />
              <Route path="/login" element={<Login />} />
              <Route path="/register" element={<Register />} />
              <Route path="/profile" element={<UserProfile />} />
            </Routes>
          </Suspense>
          <Footer />
        </Router>
      </AuthProvider>
    </ErrorBoundary>
  );
};

export default App;
```

### Memoizing Components (React)

```jsx
// ProductCard.jsx
import React, { memo } from 'react';
import { Link } from 'react-router-dom';
import { useCart } from '../hooks/useCart';

const ProductCard = ({ product }) => {
  const { addToCart } = useCart();

  const handleAddToCart = (e) => {
    e.preventDefault();
    addToCart({
      id: product.id,
      title: product.title,
      price: product.price.current.USD,
      discountedPrice: product.price.current.USD,
      quantity: 1,
      imgs: {
        thumbnails: product.images.map(img => img.sizes.thumbnail),
        previews: product.images.map(img => img.url),
      }
    });
  };

  return (
    <div className="product-card">
      <Link to={`/products/${product.id}`}>
        <div className="product-image">
          {product.images && product.images.length > 0 && (
            <img src={product.images[0].sizes.medium} alt={product.title} />
          )}
        </div>
        
        <div className="product-info">
          <h3>{product.title}</h3>
          <p className="price">${product.price.current.USD}</p>
          
          <div className="inventory-status">
            <span className={product.inventory.available ? 'in-stock' : 'out-of-stock'}>
              {product.inventory.status}
            </span>
          </div>
        </div>
      </Link>
      
      <button 
        className="add-to-cart"
        onClick={handleAddToCart}
        disabled={!product.inventory.available}
      >
        Add to Cart
      </button>
    </div>
  );
};

// Only re-render if the product data changes
export default memo(ProductCard);
```

### Optimizing API Calls with SWR

```javascript
// swrConfig.js
import { SWRConfig } from 'swr';

const swrConfig = {
  // Revalidate on focus
  revalidateOnFocus: true,
  
  // Revalidate on network reconnection
  revalidateOnReconnect: true,
  
  // Refresh interval in milliseconds (0 means no auto-refresh)
  refreshInterval: 0,
  
  // Retry on error
  shouldRetryOnError: true,
  
  // Maximum number of retries
  errorRetryCount: 3,
  
  // Dedupe requests
  dedupingInterval: 2000,
  
  // Keep previous data when fetching new data
  keepPreviousData: true,
  
  // Custom error handler
  onError: (error, key) => {
    console.error(`SWR Error for ${key}:`, error);
  }
};

export default swrConfig;
```

```jsx
// _app.jsx (Next.js) or index.jsx (React)
import { SWRConfig } from 'swr';
import swrConfig from './utils/swrConfig';

const App = ({ Component, pageProps }) => {
  return (
    <SWRConfig value={swrConfig}>
      <Component {...pageProps} />
    </SWRConfig>
  );
};

export default App;
```
