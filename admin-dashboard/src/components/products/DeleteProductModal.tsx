"use client";
import React, { useState } from 'react';
import { Modal } from '@/components/ui/modal';
import Button from '@/components/ui/button/Button';
import { TrashBinIcon } from '@/icons';
import { toast } from 'react-hot-toast';

interface DeleteProductModalProps {
  isOpen: boolean;
  onClose: () => void;
  productId: string;
  productTitle: string;
  onDelete: (productId: string) => Promise<boolean>;
}

const DeleteProductModal: React.FC<DeleteProductModalProps> = ({
  isOpen,
  onClose,
  productId,
  productTitle,
  onDelete
}) => {
  const [isDeleting, setIsDeleting] = useState(false);

  const handleDelete = async (): Promise<void> => {
    if (!productId) return;

    setIsDeleting(true);
    try {
      const success = await onDelete(productId);

      if (success) {
        toast.success('Product deleted successfully');
        onClose();
      } else {
        toast.error('Failed to delete product. Please try again.');
      }
    } catch (error: unknown) {
      console.error('Error deleting product:', error);
      toast.error('An error occurred while deleting the product');
    } finally {
      setIsDeleting(false);
    }
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      className="max-w-[500px] p-6"
    >
      <div className="text-center">
        <div className="relative flex items-center justify-center z-1 mb-7">
          <div className="w-20 h-20 rounded-full bg-danger-50 dark:bg-danger-900/20 flex items-center justify-center">
            <TrashBinIcon className="h-10 w-10 text-danger-500" />
          </div>
        </div>

        <h4 className="mb-2 text-xl font-semibold text-gray-800 dark:text-white/90">
          Delete Product
        </h4>
        <p className="text-sm leading-6 text-gray-500 dark:text-gray-400 mb-6">
          Are you sure you want to delete <span className="font-medium">{productTitle}</span>? This action cannot be undone.
        </p>

        <div className="flex items-center justify-center w-full gap-3 mt-7">
          <Button
            variant="outline"
            onClick={onClose}
            disabled={isDeleting}
          >
            Cancel
          </Button>
          <Button
            variant="danger"
            onClick={handleDelete}
            disabled={isDeleting}
            isLoading={isDeleting}
          >
            {isDeleting ? 'Deleting...' : 'Delete Product'}
          </Button>
        </div>
      </div>
    </Modal>
  );
};

export default DeleteProductModal;
