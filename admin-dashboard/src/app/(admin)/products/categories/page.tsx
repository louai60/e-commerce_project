"use client";
import React, { useState, ChangeEvent } from "react";
import PageBreadcrumb from "@/components/common/PageBreadCrumb";
import { useCategories } from "@/hooks/useProducts";
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
import { ProductService } from "@/services/product.service";
import { toast } from "react-hot-toast";

export default function CategoriesPage() {
  const [page, setPage] = useState(1);
  const [limit] = useState(10);
  const { categories, pagination, isLoading, isError, mutate } = useCategories(page, limit);

  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [modalMode, setModalMode] = useState<'create' | 'edit'>('create');
  const [selectedCategory, setSelectedCategory] = useState<any>(null);
  const [formData, setFormData] = useState({
    name: "",
    slug: "",
    description: "",
    parent_id: ""
  });

  const handlePageChange = (newPage: number) => {
    setPage(newPage);
  };

  const openCreateModal = () => {
    setModalMode('create');
    setFormData({
      name: "",
      slug: "",
      description: "",
      parent_id: ""
    });
    setIsModalOpen(true);
  };

  const openEditModal = (category: any) => {
    setModalMode('edit');
    setSelectedCategory(category);
    setFormData({
      name: category.name || "",
      slug: category.slug || "",
      description: category.description || "",
      parent_id: category.parent_id || ""
    });
    setIsModalOpen(true);
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
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
        // Create new category
        await ProductService.createCategory(formData);
        toast.success("Category created successfully");
      } else {
        // Update existing category
        await ProductService.updateCategory(selectedCategory.id, formData);
        toast.success("Category updated successfully");
      }

      setIsModalOpen(false);
      mutate(); // Refresh categories data
    } catch (error: any) {
      console.error("Error saving category:", error);
      toast.error(error.error || `Failed to ${modalMode} category`);
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
          Error loading categories
        </div>
      </div>
    );
  }

  return (
    <div>
      <div className="mb-6 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <PageBreadcrumb pageTitle="Product Categories" />
        <Button
          variant="primary"
          startIcon={<PlusIcon />}
          onClick={openCreateModal}
        >
          Add Category
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
                  <TableCell isHeader>Parent Category</TableCell>
                  <TableCell isHeader>Description</TableCell>
                  <TableCell isHeader className="text-right">Actions</TableCell>
                </TableRow>
              </TableHeader>
              <TableBody>
                {categories.map((category) => (
                  <TableRow key={category.id}>
                    <TableCell>
                      <span className="font-medium text-gray-900 dark:text-white">
                        {category.name}
                      </span>
                    </TableCell>
                    <TableCell>{category.slug}</TableCell>
                    <TableCell>{category.parent_name || "—"}</TableCell>
                    <TableCell>
                      {category.description?.substring(0, 50) || "—"}
                      {category.description?.length > 50 ? "..." : ""}
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center justify-end gap-2">
                        <Button
                          variant="outline"
                          size="sm"
                          className="h-9 w-9 p-0"
                          onClick={() => openEditModal(category)}
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
                {categories.length === 0 && (
                  <TableRow>
                    <TableCell colSpan={5} className="text-center py-8">
                      <p className="text-gray-500 dark:text-gray-400">No categories found</p>
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

      {/* Category Modal */}
      <Modal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        title={modalMode === 'create' ? "Add Category" : "Edit Category"}
      >
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <Label htmlFor="name">Category Name*</Label>
            <Input
              id="name"
              name="name"
              type="text"
              placeholder="Enter category name"
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
                placeholder="category-slug"
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
              placeholder="Category description"
              value={formData.description}
              onChange={handleChange}
            />
          </div>

          <div>
            <Label htmlFor="parent_id">Parent Category</Label>
            <div className="relative">
              <select
                id="parent_id"
                name="parent_id"
                className="h-11 w-full rounded-lg border appearance-none px-4 py-2.5 text-sm shadow-theme-xs placeholder:text-gray-400 focus:outline-hidden focus:ring-3 dark:bg-gray-900 dark:text-white/90 dark:placeholder:text-white/30 dark:focus:border-brand-800 border-gray-200 focus:border-brand-500 focus:ring-brand-500/20 dark:border-gray-700"
                value={formData.parent_id}
                onChange={handleChange}
              >
                <option value="">None (Top Level)</option>
                {categories
                  .filter(cat => modalMode === 'create' || cat.id !== selectedCategory?.id)
                  .map(category => (
                    <option key={category.id} value={category.id}>
                      {category.name}
                    </option>
                  ))
                }
              </select>
            </div>
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
