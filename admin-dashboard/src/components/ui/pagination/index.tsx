import React from "react";
import { ChevronLeftIcon, ChevronRightIcon } from "@/icons";

interface PaginationProps {
  currentPage: number;
  totalPages: number;
  onPageChange: (page: number) => void;
  showPageNumbers?: boolean;
  maxPageButtons?: number;
}

const Pagination: React.FC<PaginationProps> = ({
  currentPage,
  totalPages,
  onPageChange,
  showPageNumbers = true,
  maxPageButtons = 5,
}) => {
  const getPageNumbers = (): number[] => {
    const pageNumbers: number[] = [];
    
    // Calculate the range of page numbers to show
    let startPage = Math.max(1, currentPage - Math.floor(maxPageButtons / 2));
    const endPage = Math.min(totalPages, startPage + maxPageButtons - 1);
    
    // Adjust if we're near the end
    if (endPage - startPage + 1 < maxPageButtons) {
      startPage = Math.max(1, endPage - maxPageButtons + 1);
    }
    
    // Add page numbers
    for (let i = startPage; i <= endPage; i++) {
      pageNumbers.push(i);
    }
    
    return pageNumbers;
  };

  const handlePreviousPage = (): void => {
    if (currentPage > 1) {
      onPageChange(currentPage - 1);
    }
  };

  const handleNextPage = (): void => {
    if (currentPage < totalPages) {
      onPageChange(currentPage + 1);
    }
  };

  const handlePageClick = (pageNumber: number): void => {
    onPageChange(pageNumber);
  };

  return (
    <div className="flex items-center justify-center gap-1">
      {/* Previous button */}
      <button
        onClick={handlePreviousPage}
        disabled={currentPage === 1}
        className={`flex h-9 w-9 items-center justify-center rounded-lg border border-gray-200 ${
          currentPage === 1
            ? "cursor-not-allowed text-gray-400 dark:border-gray-700 dark:text-gray-600"
            : "text-gray-500 hover:border-gray-300 hover:bg-gray-100 dark:border-gray-700 dark:text-gray-400 dark:hover:border-gray-600 dark:hover:bg-gray-800"
        }`}
        aria-label="Previous page"
      >
        <ChevronLeftIcon className="h-5 w-5" />
      </button>

      {/* Page numbers */}
      {showPageNumbers &&
        getPageNumbers().map((pageNumber) => (
          <button
            key={pageNumber}
            onClick={() => handlePageClick(pageNumber)}
            className={`flex h-9 min-w-[36px] items-center justify-center rounded-lg px-3 text-sm font-medium ${
              currentPage === pageNumber
                ? "bg-brand-500 text-white"
                : "text-gray-500 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-800"
            }`}
            aria-label={`Page ${pageNumber}`}
            aria-current={currentPage === pageNumber ? "page" : undefined}
          >
            {pageNumber}
          </button>
        ))}

      {/* Next button */}
      <button
        onClick={handleNextPage}
        disabled={currentPage === totalPages}
        className={`flex h-9 w-9 items-center justify-center rounded-lg border border-gray-200 ${
          currentPage === totalPages
            ? "cursor-not-allowed text-gray-400 dark:border-gray-700 dark:text-gray-600"
            : "text-gray-500 hover:border-gray-300 hover:bg-gray-100 dark:border-gray-700 dark:text-gray-400 dark:hover:border-gray-600 dark:hover:bg-gray-800"
        }`}
        aria-label="Next page"
      >
        <ChevronRightIcon className="h-5 w-5" />
      </button>
    </div>
  );
};

export default Pagination;
