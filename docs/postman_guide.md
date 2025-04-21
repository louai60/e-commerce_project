# NexCart API Postman Guide

This guide provides instructions for setting up and using the NexCart API Postman collection to test the API endpoints.

## Getting Started

### Prerequisites

- [Postman](https://www.postman.com/downloads/) installed on your machine
- Access to the NexCart API (local development or production)

### Importing the Collection

1. Download the NexCart API Postman collection from the `postman/NexCart_API_Collection.json` file in the repository
2. Open Postman
3. Click on "Import" in the top left corner
4. Drag and drop the collection file or browse to select it
5. Click "Import" to add the collection to your workspace

## Setting Up Environment Variables

1. In Postman, click on the "Environments" tab in the left sidebar
2. Click the "+" button to create a new environment
3. Name it "NexCart Local" or "NexCart Production" depending on your setup
4. Add the following variables:

| Variable | Initial Value | Current Value |
|----------|---------------|---------------|
| base_url | http://localhost:8080/api/v1 | http://localhost:8080/api/v1 |
| token | | |
| refresh_token | | |
| product_id | | |
| user_id | | |
| brand_id | | |
| category_id | | |

5. Click "Save" to create the environment
6. Select the environment from the dropdown in the top right corner of Postman

## Authentication

The collection includes requests for authentication. Follow these steps to authenticate:

1. Open the "Users" folder in the collection
2. Find and open the "Login" request
3. In the request body, enter valid credentials:
   ```json
   {
     "email": "your-email@example.com",
     "password": "your-password"
   }
   ```
4. Click "Send" to execute the request
5. If successful, you'll receive a response with tokens
6. The collection includes a script that automatically sets the `token` and `refresh_token` environment variables

## Testing Endpoints

### Products API

#### List Products

1. Open the "Products" folder
2. Find and open the "List Products" request
3. Modify query parameters as needed:
   - `page`: Page number (default: 1)
   - `limit`: Items per page (default: 10)
4. Click "Send" to execute the request
5. Review the response to see the list of products

#### Get Product

1. Open the "Products" folder
2. Find and open the "Get Product" request
3. The request uses the `{{product_id}}` variable in the URL
4. To set this variable:
   - Either manually set it in your environment
   - Or run the "List Products" request first and use the script that automatically sets the ID from the first product
5. Click "Send" to execute the request
6. Review the response to see the product details

#### Create Product (Admin only)

1. Open the "Products" folder
2. Find and open the "Create Product" request
3. Ensure you're authenticated with an admin account
4. Modify the request body as needed:
   ```json
   {
     "product": {
       "title": "New Test Product",
       "slug": "new-test-product",
       "description": "This is a test product created via Postman",
       "short_description": "Test product",
       "price": 99.99,
       "sku": "TEST-001",
       "inventory_qty": 100,
       "inventory_status": "in_stock",
       "is_published": true
     }
   }
   ```
5. Click "Send" to execute the request
6. Review the response to see the created product

### Users API

#### Register

1. Open the "Users" folder
2. Find and open the "Register" request
3. Modify the request body as needed:
   ```json
   {
     "email": "new-user@example.com",
     "password": "securepassword",
     "first_name": "John",
     "last_name": "Doe"
   }
   ```
4. Click "Send" to execute the request
5. Review the response to see the created user and tokens

#### Get Profile

1. Open the "Users" folder
2. Find and open the "Get Profile" request
3. Ensure you're authenticated (run the Login request first)
4. Click "Send" to execute the request
5. Review the response to see your user profile

### Brands API

#### List Brands

1. Open the "Brands" folder
2. Find and open the "List Brands" request
3. Modify query parameters as needed
4. Click "Send" to execute the request
5. Review the response to see the list of brands

#### Create Brand (Admin only)

1. Open the "Brands" folder
2. Find and open the "Create Brand" request
3. Ensure you're authenticated with an admin account
4. Modify the request body as needed:
   ```json
   {
     "brand": {
       "name": "New Test Brand",
       "slug": "new-test-brand",
       "description": "This is a test brand created via Postman"
     }
   }
   ```
5. Click "Send" to execute the request
6. Review the response to see the created brand

### Categories API

#### List Categories

1. Open the "Categories" folder
2. Find and open the "List Categories" request
3. Modify query parameters as needed
4. Click "Send" to execute the request
5. Review the response to see the list of categories

#### Create Category (Admin only)

1. Open the "Categories" folder
2. Find and open the "Create Category" request
3. Ensure you're authenticated with an admin account
4. Modify the request body as needed:
   ```json
   {
     "category": {
       "name": "New Test Category",
       "slug": "new-test-category",
       "description": "This is a test category created via Postman",
       "parent_id": null
     }
   }
   ```
5. Click "Send" to execute the request
6. Review the response to see the created category

## Troubleshooting

### Authentication Issues

If you encounter authentication issues:

1. Ensure you've successfully run the Login request
2. Check that the `token` environment variable is set correctly
3. Verify that the token hasn't expired
4. If expired, run the "Refresh Token" request to get a new token

### Request Failures

If a request fails:

1. Check the response status code and error message
2. Verify that you're using the correct HTTP method
3. Ensure the request body is properly formatted
4. Check that all required parameters are included
5. Verify that you have the necessary permissions for the operation

## Collection Maintenance

### Updating the Collection

If the API changes:

1. Export the updated collection from Postman
2. Replace the `postman/NexCart_API_Collection.json` file in the repository
3. Commit and push the changes
4. Notify the team about the updates

### Adding New Requests

To add new requests to the collection:

1. Right-click on the appropriate folder in the collection
2. Select "Add Request"
3. Configure the request with the appropriate method, URL, headers, and body
4. Add tests and scripts as needed
5. Save the request
6. Export the updated collection and update the repository
