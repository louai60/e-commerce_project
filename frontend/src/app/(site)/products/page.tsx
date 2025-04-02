'use client';

import { useProducts } from '@/hooks/useProducts';
import ProductItem from '@/components/Common/ProductItem';
import PreLoader from '@/components/Common/PreLoader';

export default function ProductsPage() {
  const { products, isLoading, isError } = useProducts();

  if (isLoading) return <PreLoader />;
  if (isError) return <div>Error loading products</div>;

  return (
    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
      {products?.map((product) => (
        <ProductItem key={product.id} item={product} />
      ))}
    </div>
  );
}
