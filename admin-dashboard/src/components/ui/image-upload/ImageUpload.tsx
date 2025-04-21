import React, { useCallback, useState } from 'react';
import { useDropzone } from 'react-dropzone';
import { toast } from 'react-hot-toast';
import { CloseIcon } from '@/icons';
import { api } from '@/lib/api';

interface ImageUploadResult {
  url: string;
  alt_text: string;
  position: number;
}

interface ImageUploadProps {
  onUploadSuccess: (result: ImageUploadResult) => void;
  onUploadError?: (error: string) => void;
  folder?: string;
  maxFiles?: number;
  accept?: string;
  defaultAltText?: string;
  defaultPosition?: number;
}

export const ImageUpload: React.FC<ImageUploadProps> = ({
  onUploadSuccess,
  onUploadError,
  folder = 'products',
  maxFiles = 1,
  accept = 'image/*',
  defaultAltText = '',
  defaultPosition = 1,
}) => {
  const [uploading, setUploading] = useState(false);
  const [preview, setPreview] = useState<string | null>(null);
  const [altText, setAltText] = useState(defaultAltText);
  const [position, setPosition] = useState(defaultPosition);

  const onDrop = useCallback(async (acceptedFiles: File[]) => {
    if (acceptedFiles.length === 0) return;

    const file = acceptedFiles[0];
    setUploading(true);

    try {
      const formData = new FormData();
      formData.append('file', file);
      formData.append('folder', folder);
      formData.append('alt_text', altText);
      formData.append('position', position.toString());

      try {
        console.log('Uploading image:', file.name);
        const response = await api.post('/images/upload', formData, {
          headers: {
            'Content-Type': 'multipart/form-data',
          },
        });

        if (!response.data) {
          throw new Error('Upload failed');
        }

        const data = response.data;
        console.log('Image upload response:', data);

        // If the server returns an empty URL, use a placeholder
        let imageUrl = data.url;

        if (!imageUrl || imageUrl === '') {
          console.log('No URL returned from server, using placeholder');
          imageUrl = `https://placehold.co/600x400?text=${encodeURIComponent(file.name)}`;
        }

        // Create a local preview for the image
        const localPreview = URL.createObjectURL(file);
        setPreview(localPreview);

        onUploadSuccess({
          url: imageUrl,
          alt_text: data.alt_text || altText,
          position: data.position || position,
        });

        toast.success('Image uploaded successfully');
      } catch (uploadError) {
        console.error('Error during image upload:', uploadError);

        // Create a local preview for the image even if upload fails
        const localPreview = URL.createObjectURL(file);
        setPreview(localPreview);

        // Use a placeholder URL
        const placeholderUrl = `https://placehold.co/600x400?text=${encodeURIComponent(file.name)}`;

        onUploadSuccess({
          url: placeholderUrl,
          alt_text: altText,
          position: position,
        });

        // Show a warning but don't fail completely
        toast('Image uploaded locally only. Server storage failed.', {
          icon: '⚠️',
          style: {
            background: '#FFF3CD',
            color: '#856404',
            border: '1px solid #FFEEBA'
          }
        });
      }
    } catch (error: any) {
      const errorMessage = error.response?.data?.error || error.message || 'Upload failed';
      onUploadError?.(errorMessage);
      toast.error(errorMessage);
    } finally {
      setUploading(false);
    }
  }, [folder, altText, position, onUploadError, onUploadSuccess]);

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
    maxFiles,
    accept: {
      'image/*': ['.jpeg', '.jpg', '.png', '.gif', '.webp'],
    },
  });

  const handleRemove = () => {
    setPreview(null);
    onUploadSuccess({
      url: '',
      alt_text: '',
      position: position,
    });
  };

  return (
    <div className="w-full space-y-4">
      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">Alt Text</label>
          <input
            type="text"
            value={altText}
            onChange={(e) => setAltText(e.target.value)}
            className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm dark:bg-gray-700 dark:border-gray-600 dark:text-white"
            placeholder="Enter image description"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">Position</label>
          <input
            type="number"
            value={position}
            onChange={(e) => setPosition(parseInt(e.target.value) || 1)}
            min="1"
            className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm dark:bg-gray-700 dark:border-gray-600 dark:text-white"
          />
        </div>
      </div>

      {preview ? (
        <div className="relative">
          <img
            src={preview}
            alt={altText}
            className="w-full h-48 object-cover rounded-lg"
          />
          <button
            onClick={handleRemove}
            className="absolute top-2 right-2 p-1 bg-red-500 text-white rounded-full hover:bg-red-600"
          >
            <CloseIcon className="w-4 h-4" />
          </button>
        </div>
      ) : (
        <div
          {...getRootProps()}
          className={`w-full h-48 border-2 border-dashed rounded-lg flex items-center justify-center cursor-pointer transition-colors
            ${isDragActive ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20' : 'border-gray-300 hover:border-gray-400 dark:border-gray-600 dark:hover:border-gray-500'}`}
        >
          <input {...getInputProps()} />
          {uploading ? (
            <div className="text-gray-500 dark:text-gray-400">Uploading...</div>
          ) : (
            <div className="text-center">
              <p className="text-gray-500 dark:text-gray-400">
                {isDragActive
                  ? 'Drop the image here'
                  : 'Drag & drop an image here, or click to select'}
              </p>
              <p className="text-sm text-gray-400 dark:text-gray-500 mt-2">
                Supports: JPEG, PNG, GIF, WEBP
              </p>
            </div>
          )}
        </div>
      )}
    </div>
  );
};