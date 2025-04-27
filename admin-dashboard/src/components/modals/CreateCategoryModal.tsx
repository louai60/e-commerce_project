import React, { useState } from "react";
import { Modal } from "@/components/ui/modal";
import Input from "@/components/form/input/InputField";
import Label from "@/components/form/Label";
import Button from "@/components/ui/button/Button";
import { toast } from "react-hot-toast";
import { ProductService } from "@/services/product.service";

interface CreateCategoryModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: () => void;
}

const CreateCategoryModal: React.FC<CreateCategoryModalProps> = ({
  isOpen,
  onClose,
  onSuccess,
}) => {
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [formData, setFormData] = useState({
    name: "",
    slug: "",
    description: "",
    parent_id: "",
  });

  const handleInputChange = (
    e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>
  ) => {
    const { name, value } = e.target;
    setFormData((prev) => ({
      ...prev,
      [name]: value,
    }));
  };

  const generateSlug = () => {
    const slug = formData.name
      .toLowerCase()
      .replace(/[^\w\s-]/g, "")
      .replace(/[\s_-]+/g, "-")
      .replace(/^-+|-+$/g, "");

    setFormData((prev) => ({ ...prev, slug }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    // Validate form data
    if (!formData.name) {
      toast.error("Category name is required");
      return;
    }

    if (!formData.slug) {
      toast.error("Category slug is required");
      return;
    }

    setIsSubmitting(true);

    try {
      const categoryData = {
        name: formData.name,
        slug: formData.slug,
        description: formData.description,
        parent_id: formData.parent_id || undefined,
      };

      await ProductService.createCategory(categoryData);
      toast.success("Category created successfully");

      // Reset form
      setFormData({
        name: "",
        slug: "",
        description: "",
        parent_id: "",
      });

      // Close modal and refresh categories
      onSuccess();
      onClose();
    } catch (error: any) {
      console.error("Error creating category:", error);
      toast.error(error.message || "Failed to create category");
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} className="max-w-md p-5">
      <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-4 border-b border-gray-200 pb-3 dark:border-gray-700">Add Category</h3>
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <Label htmlFor="name">Category Name*</Label>
          <Input
            id="name"
            name="name"
            type="text"
            placeholder="Enter category name"
            value={formData.name}
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
              value={formData.slug}
              onChange={handleInputChange}
            />
            <Button
              variant="outline"
              type="button"
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
            onChange={handleInputChange}
          />
        </div>

        <div className="flex justify-end pt-4">
          <div className="flex gap-3">
            <Button
              variant="outline"
              type="button"
              onClick={onClose}
            >
              Cancel
            </Button>
            <Button
              variant="primary"
              type="submit"
              disabled={isSubmitting}
            >
              {isSubmitting ? "Creating..." : "Create Category"}
            </Button>
          </div>
        </div>
      </form>
    </Modal>
  );
};

export default CreateCategoryModal;
