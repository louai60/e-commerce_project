"use client";
import React from "react";

type BadgeVariant = "light" | "solid" | "success" | "warning" | "danger";
type BadgeSize = "sm" | "md";
type BadgeColor =
  | "primary"
  | "success"
  | "error"
  | "warning"
  | "info"
  | "light"
  | "dark"
  | "danger";

interface BadgeProps {
  variant?: BadgeVariant;
  size?: BadgeSize;
  color?: BadgeColor;
  startIcon?: React.ReactNode;
  endIcon?: React.ReactNode;
  children: React.ReactNode;
}

const Badge: React.FC<BadgeProps> = ({
  variant = "light",
  color = "primary",
  size = "md",
  startIcon,
  endIcon,
  children,
}) => {
  const baseStyles =
    "inline-flex items-center px-2.5 py-0.5 justify-center gap-1 rounded-full font-medium";

  const sizeStyles = {
    sm: "text-theme-xs",
    md: "text-sm",
  };

  // Define color styles for variants
  const variantStyles = {
    light: {
      primary: "bg-brand-50 text-brand-500 dark:bg-brand-500/15 dark:text-brand-400",
      success: "bg-success-50 text-success-600 dark:bg-success-500/15 dark:text-success-500",
      error: "bg-error-50 text-error-600 dark:bg-error-500/15 dark:text-error-500",
      warning: "bg-warning-50 text-warning-600 dark:bg-warning-500/15 dark:text-orange-400",
      info: "bg-blue-light-50 text-blue-light-500 dark:bg-blue-light-500/15 dark:text-blue-light-500",
      light: "bg-gray-100 text-gray-700 dark:bg-white/5 dark:text-white/80",
      dark: "bg-gray-500 text-white dark:bg-white/5 dark:text-white",
      danger: "bg-error-50 text-error-600 dark:bg-error-500/15 dark:text-error-500",
    },
    solid: {
      primary: "bg-brand-500 text-white dark:text-white",
      success: "bg-success-500 text-white dark:text-white",
      error: "bg-error-500 text-white dark:text-white",
      warning: "bg-warning-500 text-white dark:text-white",
      info: "bg-blue-light-500 text-white dark:text-white",
      light: "bg-gray-400 dark:bg-white/5 text-white dark:text-white/80",
      dark: "bg-gray-700 text-white dark:text-white",
      danger: "bg-error-500 text-white dark:text-white",
    },
    success: {
      primary: "bg-success-500 text-white dark:text-white",
      success: "bg-success-500 text-white dark:text-white",
      error: "bg-error-500 text-white dark:text-white",
      warning: "bg-warning-500 text-white dark:text-white",
      info: "bg-blue-light-500 text-white dark:text-white",
      light: "bg-success-500 text-white dark:text-white",
      dark: "bg-success-700 text-white dark:text-white",
      danger: "bg-error-500 text-white dark:text-white",
    },
    warning: {
      primary: "bg-warning-500 text-white dark:text-white",
      success: "bg-warning-500 text-white dark:text-white",
      error: "bg-warning-500 text-white dark:text-white",
      warning: "bg-warning-500 text-white dark:text-white",
      info: "bg-warning-500 text-white dark:text-white",
      light: "bg-warning-500 text-white dark:text-white",
      dark: "bg-warning-700 text-white dark:text-white",
      danger: "bg-warning-500 text-white dark:text-white",
    },
    danger: {
      primary: "bg-error-500 text-white dark:text-white",
      success: "bg-error-500 text-white dark:text-white",
      error: "bg-error-500 text-white dark:text-white",
      warning: "bg-error-500 text-white dark:text-white",
      info: "bg-error-500 text-white dark:text-white",
      light: "bg-error-500 text-white dark:text-white",
      dark: "bg-error-700 text-white dark:text-white",
      danger: "bg-error-500 text-white dark:text-white",
    },
  };

  // Get styles based on size and variant
  const sizeClass = sizeStyles[size] || sizeStyles.md;

  // Handle invalid variant/color combinations safely
  let colorStyles = "";
  try {
    // Check if variant exists
    if (variantStyles[variant]) {
      // Check if color exists for this variant
      if (variantStyles[variant][color]) {
        colorStyles = variantStyles[variant][color];
      } else {
        // Fallback to primary color if the specified color doesn't exist
        colorStyles = variantStyles[variant].primary;
        console.warn(`Color '${color}' not found for variant '${variant}', using primary instead`);
      }
    } else {
      // Fallback to light variant if the specified variant doesn't exist
      colorStyles = variantStyles.light[color] || variantStyles.light.primary;
      console.warn(`Variant '${variant}' not found, using light variant instead`);
    }
  } catch (error) {
    // Ultimate fallback for any unexpected errors
    colorStyles = "bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-300";
    console.error('Error applying badge styles:', error);
  }

  return (
    <span className={`${baseStyles} ${sizeClass} ${colorStyles}`}>
      {startIcon && <span className="mr-1">{startIcon}</span>}
      {children}
      {endIcon && <span className="ml-1">{endIcon}</span>}
    </span>
  );
};

export default Badge;


