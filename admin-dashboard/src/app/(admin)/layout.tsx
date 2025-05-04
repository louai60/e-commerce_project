"use client";

import { useSidebar } from "@/context/SidebarContext";
import AppHeader from "@/layout/AppHeader";
import AppSidebar from "@/layout/AppSidebar";
import Backdrop from "@/layout/Backdrop";
import React from "react"; // Removed unused useEffect import
import { ProductProvider } from "@/contexts/ProductContext";
import { InventoryProvider } from "@/contexts/InventoryContext";
import { useAuth } from "@/hooks/useAuth";
import { useRouter } from "next/navigation";
import LoadingSpinner from "@/components/common/LoadingSpinner";

export default function AdminLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  // Protect this layout with authentication
  const { isAuthenticated, isAdmin, isLoading } = useAuth(true);
  const { isExpanded, isHovered, isMobileOpen } = useSidebar();
  // const router = useRouter(); // Removed unused router variable

  // Dynamic class for main content margin based on sidebar state
  const mainContentMargin = isMobileOpen
    ? "ml-0"
    : isExpanded || isHovered
    ? "lg:ml-[290px]"
    : "lg:ml-[90px]";

  // If still loading auth state, show loading spinner
  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <LoadingSpinner size="lg" />
      </div>
    );
  }

  // If not authenticated or not admin, this will redirect in the useAuth hook
  if (!isAuthenticated || !isAdmin) {
    return null;
  }

  return (
    <div className="min-h-screen xl:flex">
      {/* Sidebar and Backdrop */}
      <AppSidebar />
      <Backdrop />
      {/* Main Content Area */}
      <div
        className={`flex-1 transition-all duration-300 ease-in-out ${mainContentMargin}`}
      >
        {/* Header */}
        <AppHeader />
        {/* Page Content */}
        <ProductProvider>
          <InventoryProvider>
            <div className="p-4 mx-auto max-w-(--breakpoint-2xl) md:p-6">{children}</div>
          </InventoryProvider>
        </ProductProvider>
      </div>
    </div>
  );
}
