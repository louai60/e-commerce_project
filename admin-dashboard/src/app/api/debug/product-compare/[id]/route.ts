import { NextRequest, NextResponse } from 'next/server';

export async function GET(
  request: NextRequest,
  { params }: { params: { id: string } }
) {
  const id = params.id;
  
  try {
    // Get the API URL from environment variables
    const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';
    
    // Fetch the product data from the API
    const apiResponse = await fetch(`${apiUrl}/products/${id}`, {
      headers: {
        'Content-Type': 'application/json',
      },
    });
    
    if (!apiResponse.ok) {
      throw new Error(`API responded with status: ${apiResponse.status}`);
    }
    
    const apiData = await apiResponse.json();
    
    // Create a simplified view of the API response for comparison
    const apiSummary = {
      id: apiData.id,
      title: apiData.title,
      sku: apiData.sku,
      price: apiData.price,
      images: apiData.images?.length || 0,
      imageUrls: apiData.images?.map(img => img.url) || [],
      tags: apiData.tags?.length || 0,
      tagValues: apiData.tags || [],
      specifications: apiData.specifications?.length || 0,
      specificationValues: apiData.specifications || [],
      inventory: {
        quantity: apiData.inventory?.quantity || 0,
        status: apiData.inventory?.status || 'unknown',
      },
      shipping: apiData.shipping || null,
      seo: apiData.seo || null,
      variants: apiData.variants?.length || 0,
      variantValues: apiData.variants || [],
    };
    
    // Return the comparison data
    return NextResponse.json({
      apiUrl: `${apiUrl}/products/${id}`,
      apiSummary,
      fullApiResponse: apiData,
    });
  } catch (error) {
    console.error('Error comparing product data:', error);
    return NextResponse.json(
      { error: 'Failed to compare product data' },
      { status: 500 }
    );
  }
}
