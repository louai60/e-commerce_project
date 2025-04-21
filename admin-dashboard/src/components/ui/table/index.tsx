import React, { ReactNode } from "react";

// Props for Table
interface TableProps {
  children: ReactNode;
  className?: string;
  variant?: "default" | "bordered" | "striped";
  size?: "sm" | "md" | "lg";
}

// Props for TableHeader
interface TableHeaderProps {
  children: ReactNode;
  className?: string;
}

// Props for TableBody
interface TableBodyProps {
  children: ReactNode;
  className?: string;
}

// Props for TableRow
interface TableRowProps {
  children: ReactNode;
  className?: string;
  hover?: boolean;
}

// Props for TableCell
interface TableCellProps {
  children: ReactNode;
  isHeader?: boolean;
  className?: string;
  colSpan?: number;
  align?: "left" | "center" | "right";
}

// Table Component
const Table: React.FC<TableProps> = ({ 
  children, 
  className = "",
  variant = "default",
  size = "md"
}) => {
  const baseStyles = "min-w-full divide-y divide-gray-200 dark:divide-gray-700";
  const variantStyles = {
    default: "bg-white dark:bg-gray-800",
    bordered: "border border-gray-200 dark:border-gray-700",
    striped: "bg-white dark:bg-gray-800 [&>tbody>tr:nth-child(odd)]:bg-gray-50 dark:[&>tbody>tr:nth-child(odd)]:bg-gray-900/50"
  };
  const sizeStyles = {
    sm: "text-sm",
    md: "text-base",
    lg: "text-lg"
  };

  return (
    <table className={`${baseStyles} ${variantStyles[variant]} ${sizeStyles[size]} ${className}`}>
      {children}
    </table>
  );
};

// TableHeader Component
const TableHeader: React.FC<TableHeaderProps> = ({ children, className = "" }) => {
  return (
    <thead className={`bg-gray-50 dark:bg-gray-900/50 ${className}`}>
      {children}
    </thead>
  );
};

// TableBody Component
const TableBody: React.FC<TableBodyProps> = ({ children, className = "" }) => {
  return <tbody className={`divide-y divide-gray-200 dark:divide-gray-700 ${className}`}>{children}</tbody>;
};

// TableRow Component
const TableRow: React.FC<TableRowProps> = ({ 
  children, 
  className = "",
  hover = true 
}) => {
  const hoverStyles = hover ? "hover:bg-gray-50 dark:hover:bg-gray-900/50 transition-colors duration-150" : "";
  return <tr className={`${hoverStyles} ${className}`}>{children}</tr>;
};

// TableCell Component
const TableCell: React.FC<TableCellProps> = ({
  children,
  isHeader = false,
  className = "",
  colSpan,
  align = "left"
}) => {
  const CellTag = isHeader ? "th" : "td";
  const alignStyles = {
    left: "text-left",
    center: "text-center",
    right: "text-right"
  };
  
  return (
    <CellTag 
      colSpan={colSpan} 
      className={`
        ${isHeader ? "px-6 py-3 font-medium text-gray-500 dark:text-gray-400" : "px-6 py-4 text-gray-700 dark:text-gray-300"}
        ${alignStyles[align]}
        ${className}
      `}
    >
      {children}
    </CellTag>
  );
};

export { Table, TableHeader, TableBody, TableRow, TableCell };

