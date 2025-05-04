"use client";
import React, { useState, ChangeEvent } from "react";
import PageBreadcrumb from "@/components/common/PageBreadCrumb";
import { useBrands } from "@/hooks/useProducts";
import {
  Table,
  TableBody,
  TableCell,
  TableHeader,
  TableRow
} from "@/components/ui/table";
import Button from "@/components/ui/button/Button";
import { PencilIcon, TrashBinIcon, PlusIcon } from "@/icons";
import LoadingSpinner from "@/components/ui/loading/LoadingSpinner";
import Pagination from "@/components/ui/pagination";
import Modal from "@/components/ui/modal/Modal";
import Input from "@/components/form/input/InputField";
import Label from "@/components/form/Label";
import { ProductService, Brand } from "@/services/product.service"; // Import Brand type
import { toast } from "react-hot-toast";

export default function BrandsPage() {
  const [page, setPage] = useState(1);
  const [limit] = useState(10);
  const { brands, pagination, isLoading, isError, mutate } = useBrands(page, limit);

  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [modalMode, setModalMode] = useState<'create' | 'edit'>('create');
  const [selectedBrand, setSelectedBrand] = useState<Brand | null>(null); // Use Brand type
  const [formData, setFormData] = useState({
    name: "",
    slug: "",
    description: ""
  });

  const handlePageChange = (newPage: number) => {
    setPage(newPage);
  };

  const openCreateModal = () => {
    setModalMode('create');
    setFormData({
      name: "",
      slug: "",
      description: ""
    });
    setIsModalOpen(true);
  };

  const openEditModal = (brand: Brand) => { // Use Brand type
    setModalMode('edit');
    setSelectedBrand(brand);
    setFormData({
      name: brand.name || "",
      slug: brand.slug || "",
      description: brand.description || ""
    });
    setIsModalOpen(true);
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({ ...prev, [name]: value }));
  };

  const handleInputChange = (e: ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({ ...prev, [name]: value }));
  };

  const generateSlug = () => {
    const slug = formData.name
      .toLowerCase()
      .replace(/[^\w\s-]/g, '')
      .replace(/[\s_-]+/g, '-')
      .replace(/^-+|-+$/g, '');

    setFormData(prev => ({ ...prev, slug }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
  };

  const submitForm = async () => {
    setIsSubmitting(true);

    try {
      if (modalMode === 'create') {
        // Create new brand
        await ProductService.createBrand(formData);
        toast.success("Brand created successfully");
      } else if (selectedBrand) { // Add null check for selectedBrand
        // Update existing brand
        await ProductService.updateBrand(selectedBrand.id, formData);
        toast.success("Brand updated successfully");
      } else {
        // Handle the unlikely case where selectedBrand is null in edit mode
        toast.error("Cannot update brand: No brand selected.");
      }

      setIsModalOpen(false);
      mutate(); // Refresh brands data
    } catch (error: unknown) { // Use unknown type
      console.error("Error saving brand:", error);
      // Type check before accessing properties
      let errorMessage = `Failed to ${modalMode} brand`;
      if (typeof error === 'object' && error !== null && 'error' in error && typeof error.error === 'string') {
        errorMessage = error.error;
      } else if (error instanceof Error) {
        errorMessage = error.message;
      }
      toast.error(errorMessage);
    } finally {
      setIsSubmitting(false);
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <LoadingSpinner />
      </div>
    );
  }

  if (isError) {
    return (
      <div className="rounded-xl border border-gray-200 bg-white p-6 dark:border-gray-700 dark:bg-gray-800">
        <div className="text-center text-red-500">
          Error loading brands
        </div>
      </div>
    );
  }

  return (
    <div>
      <div className="mb-6 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <PageBreadcrumb pageTitle="Product Brands" />
        <Button
          variant="primary"
          startIcon={<PlusIcon />}
          onClick={openCreateModal}
        >
          Add Brand
        </Button>
      </div>

      <div className="rounded-xl border border-gray-200 bg-white dark:border-gray-700 dark:bg-gray-800">
        <div className="p-4 md:p-6">
          <div className="overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableCell isHeader>Name</TableCell>
                  <TableCell isHeader>Slug</TableCell>
                  <TableCell isHeader>Description</TableCell>
                  <TableCell isHeader className="text-right">Actions</TableCell>
                </TableRow>
              </TableHeader>
              <TableBody>
                {brands.map((brand) => (
                  <TableRow key={brand.id}>
                    <TableCell>
                      <span className="font-medium text-gray-900 dark:text-white">
                        {brand.name}
                      </span>
                    </TableCell>
                    <TableCell>{brand.slug}</TableCell>
                    <TableCell>
                      {brand.description ? brand.description.substring(0, 50) : "—"}
                      {brand.description && brand.description.length > 50 ? "..." : ""} {/* Check existence before length */}
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center justify-end gap-2">
                        <Button
                          variant="outline"
                          size="sm"
                          className="h-9 w-9 p-0"
                          onClick={() => openEditModal(brand)}
                        >
                          <PencilIcon className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="outline"
                          size="sm"
                          className="h-9 w-9 p-0 text-danger-500 hover:border-danger-500 hover:bg-danger-500/10"
                        >
                          <TrashBinIcon className="h-4 w-4" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
                {brands.length === 0 && (
                  <TableRow>
                    <TableCell colSpan={4} className="text-center py-8">
                      <p className="text-gray-500 dark:text-gray-400">No brands found</p>
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </div>

          {pagination && pagination.total_pages > 1 && (
            <div className="mt-6 flex justify-center">
              <Pagination
                currentPage={pagination.current_page}
                totalPages={pagination.total_pages}
                onPageChange={handlePageChange}
              />
            </div>
          )}
        </div>
      </div>

      {/* Brand Modal */}
      <Modal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        title={modalMode === 'create' ? "Add Brand" : "Edit Brand"}
      >
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <Label htmlFor="name">Brand Name*</Label>
            <Input
              id="name"
              name="name"
              type="text"
              placeholder="Enter brand name"
              defaultValue={formData.name}
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
                placeholder="brand-slug"
                defaultValue={formData.slug}
                onChange={handleInputChange}
              />
              <Button
                variant="outline"
                onClick={generateSlug}
                disabled={!formData.name}
              >
                Generate
              </Button>
            </div>
          </div>

          <div>
            <Label htmlFor="description">Description</Label>
            <textarea
              id="description"
              name="description"
              rows={3}
              className="h-auto w-full rounded-lg border appearance-none px-4 py-2.5 text-sm shadow-theme-xs placeholder:text-gray-400 focus:outline-hidden focus:ring-3 dark:bg-gray-900 dark:text-white/90 dark:placeholder:text-white/30 dark:focus:border-brand-800 border-gray-200 focus:border-brand-500 focus:ring-brand-500/20 dark:border-gray-700"
              placeholder="Brand description"
              value={formData.description}
              onChange={handleChange}
            />
          </div>

          <div className="flex justify-end gap-3 pt-4">
            <Button
              variant="outline"
              onClick={() => setIsModalOpen(false)}
            >
              Cancel
            </Button>
            <Button
              variant="primary"
              disabled={isSubmitting}
              onClick={submitForm}
            >
              {isSubmitting ? "Saving..." : modalMode === 'create' ? "Create" : "Save Changes"}
            </Button>
          </div>
        </form>
      </Modal>
    </div>
  );
}
