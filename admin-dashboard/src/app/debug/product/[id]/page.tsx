"use client";

import React, { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';

export default function ProductDebugPage() {
  const params = useParams();
  const id = params.id as string;

  // Define a proper type for the product data
  interface ProductDebugData {
    apiSummary: {
      title: string;
      sku: string;
      price: {
        current?: {
          USD?: number;
        };
        value?: number;
      };
      images: number;
      imageUrls: string[];
      tags: number;
      specifications: number;
      inventory: {
        quantity: number;
        status: string;
      };
      variants: number;
    };
    fullApiResponse: Record<string, unknown>;
  }

  const [data, setData] = useState<ProductDebugData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function fetchData() {
      try {
        setLoading(true);
        const response = await fetch(`/api/debug/product-compare/${id}`);

        if (!response.ok) {
          throw new Error(`API responded with status: ${response.status}`);
        }

        const result = await response.json();
        setData(result);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'An unknown error occurred');
      } finally {
        setLoading(false);
      }
    }

    if (id) {
      fetchData();
    }
  }, [id]);

  if (loading) {
    return (
      <div className="p-8">
        <h1 className="text-2xl font-bold mb-4">Product Debug - Loading...</h1>
        <div className="animate-pulse">
          <div className="h-4 bg-gray-200 rounded w-3/4 mb-4"></div>
          <div className="h-4 bg-gray-200 rounded w-1/2 mb-4"></div>
          <div className="h-4 bg-gray-200 rounded w-5/6 mb-4"></div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-8">
        <h1 className="text-2xl font-bold mb-4">Product Debug - Error</h1>
        <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded">
          {error}
        </div>
      </div>
    );
  }

  if (!data) {
    return (
      <div className="p-8">
        <h1 className="text-2xl font-bold mb-4">Product Debug - No Data</h1>
        <p>No data available for product ID: {id}</p>
      </div>
    );
  }

  const { apiSummary, fullApiResponse } = data;

  return (
    <div className="p-8">
      <h1 className="text-2xl font-bold mb-4">Product Debug - ID: {id}</h1>

      <div className="mb-8">
        <h2 className="text-xl font-semibold mb-2">API Response Summary</h2>
        <div className="bg-white shadow overflow-hidden rounded-lg">
          <div className="px-4 py-5 sm:p-6">
            <dl className="grid grid-cols-1 gap-x-4 gap-y-6 sm:grid-cols-2">
              <div>
                <dt className="text-sm font-medium text-gray-500">Title</dt>
                <dd className="mt-1 text-sm text-gray-900">{apiSummary.title}</dd>
              </div>

              <div>
                <dt className="text-sm font-medium text-gray-500">SKU</dt>
                <dd className="mt-1 text-sm text-gray-900">{apiSummary.sku}</dd>
              </div>

              <div>
                <dt className="text-sm font-medium text-gray-500">Price</dt>
                <dd className="mt-1 text-sm text-gray-900">
                  USD: {apiSummary.price?.current?.USD || 'N/A'} |
                  Value: {apiSummary.price?.value || 'N/A'}
                </dd>
              </div>

              <div>
                <dt className="text-sm font-medium text-gray-500">Images</dt>
                <dd className="mt-1 text-sm text-gray-900">
                  Count: {apiSummary.images}
                </dd>
              </div>

              <div>
                <dt className="text-sm font-medium text-gray-500">Image URLs</dt>
                <dd className="mt-1 text-sm text-gray-900">
                  {apiSummary.imageUrls.length > 0 ? (
                    <ul className="list-disc pl-5">
                      {apiSummary.imageUrls.map((url: string, index: number) => (
                        <li key={index} className="truncate max-w-xs">{url}</li>
                      ))}
                    </ul>
                  ) : (
                    'No images'
                  )}
                </dd>
              </div>

              <div>
                <dt className="text-sm font-medium text-gray-500">Tags</dt>
                <dd className="mt-1 text-sm text-gray-900">
                  Count: {apiSummary.tags}
                </dd>
              </div>

              <div>
                <dt className="text-sm font-medium text-gray-500">Specifications</dt>
                <dd className="mt-1 text-sm text-gray-900">
                  Count: {apiSummary.specifications}
                </dd>
              </div>

              <div>
                <dt className="text-sm font-medium text-gray-500">Inventory</dt>
                <dd className="mt-1 text-sm text-gray-900">
                  Quantity: {apiSummary.inventory.quantity} |
                  Status: {apiSummary.inventory.status}
                </dd>
              </div>

              <div>
                <dt className="text-sm font-medium text-gray-500">Variants</dt>
                <dd className="mt-1 text-sm text-gray-900">
                  Count: {apiSummary.variants}
                </dd>
              </div>
            </dl>
          </div>
        </div>
      </div>

      <div className="mb-8">
        <h2 className="text-xl font-semibold mb-2">Full API Response</h2>
        <div className="bg-gray-800 rounded-lg p-4 overflow-auto max-h-96">
          <pre className="text-green-400 text-sm">
            {JSON.stringify(fullApiResponse, null, 2)}
          </pre>
        </div>
      </div>

      <div className="mb-8">
        <h2 className="text-xl font-semibold mb-2">How to Check Database Data</h2>
        <div className="bg-white shadow overflow-hidden rounded-lg">
          <div className="px-4 py-5 sm:p-6">
            <p className="mb-4">To check the product data in the database, run the following commands:</p>

            <div className="bg-gray-100 p-4 rounded-lg mb-4">
              <code className="text-sm">
                cd backend\product-service<br />
                scripts\run_debug_product_data.bat
              </code>
            </div>

            <p className="mb-4">This will generate a file called <code>product_data_debug.txt</code> with all the database data for this product.</p>

            <p>Compare this data with the API response above to identify any discrepancies.</p>
          </div>
        </div>
      </div>
    </div>
  );
}
