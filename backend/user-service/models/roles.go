package models

// User Types
const (
    UserTypeCustomer = "customer"
    UserTypeSeller  = "seller"
    UserTypeAdmin   = "admin"
)

// Customer Roles
const (
    RoleGuest          = "guest"
    RoleRegistered     = "registered"
    RolePremiumMember  = "premium"
)

// Seller Roles
const (
    RoleBasicSeller    = "basic_seller"
    RoleVerifiedSeller = "verified_seller"
)

// Admin Roles
const (
    RoleSupportAgent   = "support_agent"
    RoleWarehouseStaff = "warehouse_staff"
    RoleSuperAdmin     = "super_admin"
)

// Permission definitions
type Permission string

const (
    // Customer Permissions
    PermBrowseProducts    Permission = "browse_products"
    PermManageCart        Permission = "manage_cart"
    PermPlaceOrders       Permission = "place_orders"
    PermWriteReviews      Permission = "write_reviews"
    PermTrackOrders       Permission = "track_orders"
    
    // Seller Permissions
    PermManageProducts    Permission = "manage_products"
    PermManageInventory   Permission = "manage_inventory"
    PermViewSalesReports  Permission = "view_sales_reports"
    PermProcessOrders     Permission = "process_orders"
    
    // Admin Permissions
    PermManageUsers       Permission = "manage_users"
    PermManageRefunds     Permission = "manage_refunds"
    PermManageWarehouse   Permission = "manage_warehouse"
    PermFullAccess        Permission = "full_access"
)

// RolePermissions maps roles to their permissions
var RolePermissions = map[string][]Permission{
    RoleGuest: {
        PermBrowseProducts,
    },
    RoleRegistered: {
        PermBrowseProducts,
        PermManageCart,
        PermPlaceOrders,
        PermWriteReviews,
        PermTrackOrders,
    },
    RolePremiumMember: {
        PermBrowseProducts,
        PermManageCart,
        PermPlaceOrders,
        PermWriteReviews,
        PermTrackOrders,
    },
    RoleBasicSeller: {
        PermManageProducts,
        PermManageInventory,
        PermViewSalesReports,
        PermProcessOrders,
    },
    RoleVerifiedSeller: {
        PermManageProducts,
        PermManageInventory,
        PermViewSalesReports,
        PermProcessOrders,
    },
    RoleSupportAgent: {
        PermManageRefunds,
        PermManageUsers,
    },
    RoleWarehouseStaff: {
        PermManageWarehouse,
        PermManageInventory,
    },
    RoleSuperAdmin: {
        PermFullAccess,
    },
}