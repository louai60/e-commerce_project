package handlers

import (
	"context"
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/golang/protobuf/ptypes/wrappers"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/louai60/e-commerce_project/backend/inventory-service/models"
	pb "github.com/louai60/e-commerce_project/backend/inventory-service/proto"
	"github.com/louai60/e-commerce_project/backend/inventory-service/service"
)

// InventoryHandler handles gRPC requests for inventory operations
type InventoryHandler struct {
	inventoryService *service.InventoryService
	warehouseService *service.WarehouseService
	logger           *zap.Logger
	pb.UnimplementedInventoryServiceServer
}

// NewInventoryHandler creates a new inventory handler
func NewInventoryHandler(
	inventoryService *service.InventoryService,
	warehouseService *service.WarehouseService,
	logger *zap.Logger,
) *InventoryHandler {
	return &InventoryHandler{
		inventoryService: inventoryService,
		warehouseService: warehouseService,
		logger:           logger,
	}
}

// CreateInventoryItem creates a new inventory item
func (h *InventoryHandler) CreateInventoryItem(ctx context.Context, req *pb.CreateInventoryItemRequest) (*pb.InventoryItemResponse, error) {
	h.logger.Info("CreateInventoryItem request received", zap.String("product_id", req.ProductId), zap.String("sku", req.Sku))

	// Convert warehouse allocations
	var warehouseAllocations []models.WarehouseAllocation
	for _, allocation := range req.WarehouseAllocations {
		warehouseAllocations = append(warehouseAllocations, models.WarehouseAllocation{
			WarehouseID: allocation.WarehouseId,
			Quantity:    int(allocation.Quantity),
		})
	}

	// Convert variant ID
	var variantID *string
	if req.VariantId != nil && req.VariantId.Value != "" {
		variantID = &req.VariantId.Value
	}

	// Create inventory item
	item, err := h.inventoryService.CreateInventoryItem(
		ctx,
		req.ProductId,
		req.Sku,
		variantID,
		int(req.InitialQuantity),
		int(req.ReorderPoint),
		int(req.ReorderQuantity),
		warehouseAllocations,
	)

	if err != nil {
		h.logger.Error("Failed to create inventory item", zap.Error(err))
		return nil, mapErrorToGRPCStatus(err)
	}

	// Convert to protobuf response
	pbItem, err := mapInventoryItemToProto(item)
	if err != nil {
		h.logger.Error("Failed to map inventory item to proto", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to map inventory item to proto")
	}

	return &pb.InventoryItemResponse{
		InventoryItem: pbItem,
	}, nil
}

// GetInventoryItem retrieves an inventory item
func (h *InventoryHandler) GetInventoryItem(ctx context.Context, req *pb.GetInventoryItemRequest) (*pb.InventoryItemResponse, error) {
	var id, productID, sku string

	// Extract the identifier based on which field is set
	switch req.Identifier.(type) {
	case *pb.GetInventoryItemRequest_Id:
		id = req.GetId()
		h.logger.Info("GetInventoryItem request received", zap.String("id", id))
	case *pb.GetInventoryItemRequest_ProductId:
		productID = req.GetProductId()
		h.logger.Info("GetInventoryItem request received", zap.String("product_id", productID))
	case *pb.GetInventoryItemRequest_Sku:
		sku = req.GetSku()
		h.logger.Info("GetInventoryItem request received", zap.String("sku", sku))
	default:
		h.logger.Warn("GetInventoryItem request received with no identifier")
		return nil, status.Error(codes.InvalidArgument, "No identifier provided")
	}

	// Get inventory item
	item, err := h.inventoryService.GetInventoryItem(ctx, id, productID, sku)
	if err != nil {
		h.logger.Error("Failed to get inventory item", zap.Error(err))
		return nil, mapErrorToGRPCStatus(err)
	}

	// Convert to protobuf response
	pbItem, err := mapInventoryItemToProto(item)
	if err != nil {
		h.logger.Error("Failed to map inventory item to proto", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to map inventory item to proto")
	}

	return &pb.InventoryItemResponse{
		InventoryItem: pbItem,
	}, nil
}

// UpdateInventoryItem updates an inventory item
func (h *InventoryHandler) UpdateInventoryItem(ctx context.Context, req *pb.UpdateInventoryItemRequest) (*pb.InventoryItemResponse, error) {
	h.logger.Info("UpdateInventoryItem request received", zap.String("id", req.Id))

	// Convert optional fields
	var reorderPoint, reorderQty *int
	var statusValue *string

	if req.ReorderPoint != nil {
		rp := int(req.ReorderPoint.Value)
		reorderPoint = &rp
	}

	if req.ReorderQuantity != nil {
		rq := int(req.ReorderQuantity.Value)
		reorderQty = &rq
	}

	if req.Status != nil && req.Status.Value != "" {
		s := req.Status.Value
		statusValue = &s
	}

	// Update inventory item
	item, err := h.inventoryService.UpdateInventoryItem(ctx, req.Id, reorderPoint, reorderQty, statusValue)
	if err != nil {
		h.logger.Error("Failed to update inventory item", zap.Error(err))
		return nil, mapErrorToGRPCStatus(err)
	}

	// Convert to protobuf response
	pbItem, err := mapInventoryItemToProto(item)
	if err != nil {
		h.logger.Error("Failed to map inventory item to proto", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to map inventory item to proto")
	}

	return &pb.InventoryItemResponse{
		InventoryItem: pbItem,
	}, nil
}

// ListInventoryItems retrieves a paginated list of inventory items
func (h *InventoryHandler) ListInventoryItems(ctx context.Context, req *pb.ListInventoryItemsRequest) (*pb.ListInventoryItemsResponse, error) {
	h.logger.Info("ListInventoryItems request received",
		zap.Int32("page", req.Page),
		zap.Int32("limit", req.Limit))

	// Extract optional filters
	var statusFilter, warehouseIDFilter string
	if req.Status != nil && req.Status.Value != "" {
		statusFilter = req.Status.Value
	}
	if req.WarehouseId != nil && req.WarehouseId.Value != "" {
		warehouseIDFilter = req.WarehouseId.Value
	}

	// Get inventory items
	items, total, err := h.inventoryService.ListInventoryItems(
		ctx,
		int(req.Page),
		int(req.Limit),
		statusFilter,
		warehouseIDFilter,
		req.LowStockOnly,
	)
	if err != nil {
		h.logger.Error("Failed to list inventory items", zap.Error(err))
		return nil, mapErrorToGRPCStatus(err)
	}

	// Convert to protobuf response
	var pbItems []*pb.InventoryItem
	for _, item := range items {
		pbItem, err := mapInventoryItemToProto(item)
		if err != nil {
			h.logger.Error("Failed to map inventory item to proto", zap.Error(err))
			continue
		}
		pbItems = append(pbItems, pbItem)
	}

	return &pb.ListInventoryItemsResponse{
		InventoryItems: pbItems,
		Total:          int32(total),
	}, nil
}

// Helper functions for mapping between domain models and protobuf messages

// mapInventoryItemToProto converts a domain inventory item to a protobuf message
func mapInventoryItemToProto(item *models.InventoryItem) (*pb.InventoryItem, error) {
	// Convert timestamps
	createdAt := &timestamp.Timestamp{
		Seconds: item.CreatedAt.Unix(),
		Nanos:   int32(item.CreatedAt.Nanosecond()),
	}
	updatedAt := &timestamp.Timestamp{
		Seconds: item.UpdatedAt.Unix(),
		Nanos:   int32(item.UpdatedAt.Nanosecond()),
	}
	lastUpdated := &timestamp.Timestamp{
		Seconds: item.LastUpdated.Unix(),
		Nanos:   int32(item.LastUpdated.Nanosecond()),
	}

	// Convert variant ID
	var variantID *wrappers.StringValue
	if item.VariantID != nil {
		variantID = &wrappers.StringValue{Value: *item.VariantID}
	}

	// Convert inventory locations
	var pbLocations []*pb.InventoryLocation
	for _, location := range item.Locations {
		pbLocation, err := mapInventoryLocationToProto(&location)
		if err != nil {
			return nil, err
		}
		pbLocations = append(pbLocations, pbLocation)
	}

	return &pb.InventoryItem{
		Id:                item.ID,
		ProductId:         item.ProductID,
		VariantId:         variantID,
		Sku:               item.SKU,
		TotalQuantity:     int32(item.TotalQuantity),
		AvailableQuantity: int32(item.AvailableQuantity),
		ReservedQuantity:  int32(item.ReservedQuantity),
		ReorderPoint:      int32(item.ReorderPoint),
		ReorderQuantity:   int32(item.ReorderQuantity),
		Status:            item.Status,
		LastUpdated:       lastUpdated,
		CreatedAt:         createdAt,
		UpdatedAt:         updatedAt,
		Locations:         pbLocations,
	}, nil
}

// mapInventoryLocationToProto converts a domain inventory location to a protobuf message
func mapInventoryLocationToProto(location *models.InventoryLocation) (*pb.InventoryLocation, error) {
	// Convert timestamps
	createdAt := &timestamp.Timestamp{
		Seconds: location.CreatedAt.Unix(),
		Nanos:   int32(location.CreatedAt.Nanosecond()),
	}
	updatedAt := &timestamp.Timestamp{
		Seconds: location.UpdatedAt.Unix(),
		Nanos:   int32(location.UpdatedAt.Nanosecond()),
	}

	// Convert warehouse if available
	var pbWarehouse *pb.Warehouse
	if location.Warehouse != nil {
		var err error
		pbWarehouse, err = mapWarehouseToProto(location.Warehouse)
		if err != nil {
			return nil, err
		}
	}

	return &pb.InventoryLocation{
		Id:                location.ID,
		InventoryItemId:   location.InventoryItemID,
		WarehouseId:       location.WarehouseID,
		Quantity:          int32(location.Quantity),
		AvailableQuantity: int32(location.AvailableQuantity),
		ReservedQuantity:  int32(location.ReservedQuantity),
		CreatedAt:         createdAt,
		UpdatedAt:         updatedAt,
		Warehouse:         pbWarehouse,
	}, nil
}

// mapWarehouseToProto converts a domain warehouse to a protobuf message
func mapWarehouseToProto(warehouse *models.Warehouse) (*pb.Warehouse, error) {
	// Convert timestamps
	createdAt := &timestamp.Timestamp{
		Seconds: warehouse.CreatedAt.Unix(),
		Nanos:   int32(warehouse.CreatedAt.Nanosecond()),
	}
	updatedAt := &timestamp.Timestamp{
		Seconds: warehouse.UpdatedAt.Unix(),
		Nanos:   int32(warehouse.UpdatedAt.Nanosecond()),
	}

	return &pb.Warehouse{
		Id:         warehouse.ID,
		Name:       warehouse.Name,
		Code:       warehouse.Code,
		Address:    warehouse.Address,
		City:       warehouse.City,
		State:      warehouse.State,
		Country:    warehouse.Country,
		PostalCode: warehouse.PostalCode,
		IsActive:   warehouse.IsActive,
		Priority:   int32(warehouse.Priority),
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}, nil
}

// mapProtoToInventoryItem converts a protobuf message to a domain inventory item
func mapProtoToInventoryItem(pbItem *pb.InventoryItem) (*models.InventoryItem, error) {
	// Convert timestamps
	var createdAt, updatedAt, lastUpdated time.Time
	if pbItem.CreatedAt != nil {
		createdAt = time.Unix(pbItem.CreatedAt.Seconds, int64(pbItem.CreatedAt.Nanos))
	}
	if pbItem.UpdatedAt != nil {
		updatedAt = time.Unix(pbItem.UpdatedAt.Seconds, int64(pbItem.UpdatedAt.Nanos))
	}
	if pbItem.LastUpdated != nil {
		lastUpdated = time.Unix(pbItem.LastUpdated.Seconds, int64(pbItem.LastUpdated.Nanos))
	}

	// Convert variant ID
	var variantID *string
	if pbItem.VariantId != nil && pbItem.VariantId.Value != "" {
		variantID = &pbItem.VariantId.Value
	}

	// Convert inventory locations
	var locations []models.InventoryLocation
	for _, pbLocation := range pbItem.Locations {
		location, err := mapProtoToInventoryLocation(pbLocation)
		if err != nil {
			return nil, err
		}
		locations = append(locations, *location)
	}

	return &models.InventoryItem{
		ID:                pbItem.Id,
		ProductID:         pbItem.ProductId,
		VariantID:         variantID,
		SKU:               pbItem.Sku,
		TotalQuantity:     int(pbItem.TotalQuantity),
		AvailableQuantity: int(pbItem.AvailableQuantity),
		ReservedQuantity:  int(pbItem.ReservedQuantity),
		ReorderPoint:      int(pbItem.ReorderPoint),
		ReorderQuantity:   int(pbItem.ReorderQuantity),
		Status:            pbItem.Status,
		LastUpdated:       lastUpdated,
		CreatedAt:         createdAt,
		UpdatedAt:         updatedAt,
		Locations:         locations,
	}, nil
}

// mapProtoToInventoryLocation converts a protobuf message to a domain inventory location
func mapProtoToInventoryLocation(pbLocation *pb.InventoryLocation) (*models.InventoryLocation, error) {
	// Convert timestamps
	var createdAt, updatedAt time.Time
	if pbLocation.CreatedAt != nil {
		createdAt = time.Unix(pbLocation.CreatedAt.Seconds, int64(pbLocation.CreatedAt.Nanos))
	}
	if pbLocation.UpdatedAt != nil {
		updatedAt = time.Unix(pbLocation.UpdatedAt.Seconds, int64(pbLocation.UpdatedAt.Nanos))
	}

	// Convert warehouse if available
	var warehouse *models.Warehouse
	if pbLocation.Warehouse != nil {
		var err error
		warehouse, err = mapProtoToWarehouse(pbLocation.Warehouse)
		if err != nil {
			return nil, err
		}
	}

	return &models.InventoryLocation{
		ID:                pbLocation.Id,
		InventoryItemID:   pbLocation.InventoryItemId,
		WarehouseID:       pbLocation.WarehouseId,
		Quantity:          int(pbLocation.Quantity),
		AvailableQuantity: int(pbLocation.AvailableQuantity),
		ReservedQuantity:  int(pbLocation.ReservedQuantity),
		CreatedAt:         createdAt,
		UpdatedAt:         updatedAt,
		Warehouse:         warehouse,
	}, nil
}

// mapProtoToWarehouse converts a protobuf message to a domain warehouse
func mapProtoToWarehouse(pbWarehouse *pb.Warehouse) (*models.Warehouse, error) {
	// Convert timestamps
	var createdAt, updatedAt time.Time
	if pbWarehouse.CreatedAt != nil {
		createdAt = time.Unix(pbWarehouse.CreatedAt.Seconds, int64(pbWarehouse.CreatedAt.Nanos))
	}
	if pbWarehouse.UpdatedAt != nil {
		updatedAt = time.Unix(pbWarehouse.UpdatedAt.Seconds, int64(pbWarehouse.UpdatedAt.Nanos))
	}

	return &models.Warehouse{
		ID:         pbWarehouse.Id,
		Name:       pbWarehouse.Name,
		Code:       pbWarehouse.Code,
		Address:    pbWarehouse.Address,
		City:       pbWarehouse.City,
		State:      pbWarehouse.State,
		Country:    pbWarehouse.Country,
		PostalCode: pbWarehouse.PostalCode,
		IsActive:   pbWarehouse.IsActive,
		Priority:   int(pbWarehouse.Priority),
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}, nil
}

// mapErrorToGRPCStatus maps domain errors to gRPC status errors
func mapErrorToGRPCStatus(err error) error {
	switch err {
	case models.ErrNotFound:
		return status.Error(codes.NotFound, err.Error())
	case models.ErrAlreadyExists:
		return status.Error(codes.AlreadyExists, err.Error())
	case models.ErrInvalidInput:
		return status.Error(codes.InvalidArgument, err.Error())
	case models.ErrInsufficientInventory:
		return status.Error(codes.FailedPrecondition, err.Error())
	case models.ErrReservationExpired:
		return status.Error(codes.FailedPrecondition, err.Error())
	case models.ErrReservationNotFound:
		return status.Error(codes.NotFound, err.Error())
	case models.ErrReservationInvalidState:
		return status.Error(codes.FailedPrecondition, err.Error())
	case models.ErrWarehouseNotFound:
		return status.Error(codes.NotFound, err.Error())
	case models.ErrWarehouseInactive:
		return status.Error(codes.FailedPrecondition, err.Error())
	case models.ErrInvalidQuantity:
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, "Internal server error")
	}
}

// CreateWarehouse creates a new warehouse
func (h *InventoryHandler) CreateWarehouse(ctx context.Context, req *pb.CreateWarehouseRequest) (*pb.WarehouseResponse, error) {
	h.logger.Info("CreateWarehouse request received", zap.String("name", req.Name), zap.String("code", req.Code))

	// Create warehouse
	warehouse, err := h.warehouseService.CreateWarehouse(
		ctx,
		req.Name,
		req.Code,
		req.Address,
		req.City,
		req.State,
		req.Country,
		req.PostalCode,
		int(req.Priority),
	)

	if err != nil {
		h.logger.Error("Failed to create warehouse", zap.Error(err))
		return nil, mapErrorToGRPCStatus(err)
	}

	// Convert to protobuf response
	pbWarehouse, err := mapWarehouseToProto(warehouse)
	if err != nil {
		h.logger.Error("Failed to map warehouse to proto", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to map warehouse to proto")
	}

	return &pb.WarehouseResponse{
		Warehouse: pbWarehouse,
	}, nil
}

// GetWarehouse retrieves a warehouse
func (h *InventoryHandler) GetWarehouse(ctx context.Context, req *pb.GetWarehouseRequest) (*pb.WarehouseResponse, error) {
	var id, code string

	// Extract the identifier based on which field is set
	switch req.Identifier.(type) {
	case *pb.GetWarehouseRequest_Id:
		id = req.GetId()
		h.logger.Info("GetWarehouse request received", zap.String("id", id))
	case *pb.GetWarehouseRequest_Code:
		code = req.GetCode()
		h.logger.Info("GetWarehouse request received", zap.String("code", code))
	default:
		h.logger.Warn("GetWarehouse request received with no identifier")
		return nil, status.Error(codes.InvalidArgument, "No identifier provided")
	}

	// Get warehouse
	warehouse, err := h.warehouseService.GetWarehouse(ctx, id, code)
	if err != nil {
		h.logger.Error("Failed to get warehouse", zap.Error(err))
		return nil, mapErrorToGRPCStatus(err)
	}

	// Convert to protobuf response
	pbWarehouse, err := mapWarehouseToProto(warehouse)
	if err != nil {
		h.logger.Error("Failed to map warehouse to proto", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to map warehouse to proto")
	}

	return &pb.WarehouseResponse{
		Warehouse: pbWarehouse,
	}, nil
}

// UpdateWarehouse updates a warehouse
func (h *InventoryHandler) UpdateWarehouse(ctx context.Context, req *pb.UpdateWarehouseRequest) (*pb.WarehouseResponse, error) {
	h.logger.Info("UpdateWarehouse request received", zap.String("id", req.Id))

	// Convert optional fields
	var priority *int
	var isActive *bool

	if req.Priority != nil {
		p := int(req.Priority.Value)
		priority = &p
	}

	if req.IsActive != nil {
		a := req.IsActive.Value
		isActive = &a
	}

	// Extract string fields
	name := ""
	if req.Name != nil {
		name = req.Name.Value
	}

	address := ""
	if req.Address != nil {
		address = req.Address.Value
	}

	city := ""
	if req.City != nil {
		city = req.City.Value
	}

	state := ""
	if req.State != nil {
		state = req.State.Value
	}

	country := ""
	if req.Country != nil {
		country = req.Country.Value
	}

	postalCode := ""
	if req.PostalCode != nil {
		postalCode = req.PostalCode.Value
	}

	// Update warehouse
	warehouse, err := h.warehouseService.UpdateWarehouse(
		ctx,
		req.Id,
		name,
		address,
		city,
		state,
		country,
		postalCode,
		priority,
		isActive,
	)

	if err != nil {
		h.logger.Error("Failed to update warehouse", zap.Error(err))
		return nil, mapErrorToGRPCStatus(err)
	}

	// Convert to protobuf response
	pbWarehouse, err := mapWarehouseToProto(warehouse)
	if err != nil {
		h.logger.Error("Failed to map warehouse to proto", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to map warehouse to proto")
	}

	return &pb.WarehouseResponse{
		Warehouse: pbWarehouse,
	}, nil
}

// ListWarehouses retrieves a paginated list of warehouses
func (h *InventoryHandler) ListWarehouses(ctx context.Context, req *pb.ListWarehousesRequest) (*pb.ListWarehousesResponse, error) {
	h.logger.Info("ListWarehouses request received",
		zap.Int32("page", req.Page),
		zap.Int32("limit", req.Limit))

	// Extract optional filters
	var isActive *bool
	if req.IsActive != nil {
		a := req.IsActive.Value
		isActive = &a
	}

	// Get warehouses
	warehouses, total, err := h.warehouseService.ListWarehouses(
		ctx,
		int(req.Page),
		int(req.Limit),
		isActive,
	)
	if err != nil {
		h.logger.Error("Failed to list warehouses", zap.Error(err))
		return nil, mapErrorToGRPCStatus(err)
	}

	// Convert to protobuf response
	var pbWarehouses []*pb.Warehouse
	for _, warehouse := range warehouses {
		pbWarehouse, err := mapWarehouseToProto(warehouse)
		if err != nil {
			h.logger.Error("Failed to map warehouse to proto", zap.Error(err))
			continue
		}
		pbWarehouses = append(pbWarehouses, pbWarehouse)
	}

	return &pb.ListWarehousesResponse{
		Warehouses: pbWarehouses,
		Total:      int32(total),
	}, nil
}

// AddInventoryToLocation adds inventory to a specific warehouse location
func (h *InventoryHandler) AddInventoryToLocation(ctx context.Context, req *pb.AddInventoryToLocationRequest) (*pb.InventoryLocationResponse, error) {
	h.logger.Info("AddInventoryToLocation request received",
		zap.String("inventory_item_id", req.InventoryItemId),
		zap.String("warehouse_id", req.WarehouseId),
		zap.Int32("quantity", req.Quantity))

	// Add inventory to location
	location, err := h.inventoryService.AddInventoryToLocation(
		ctx,
		req.InventoryItemId,
		req.WarehouseId,
		int(req.Quantity),
		req.ReferenceId,
		req.ReferenceType,
		req.Notes,
	)

	if err != nil {
		h.logger.Error("Failed to add inventory to location", zap.Error(err))
		return nil, mapErrorToGRPCStatus(err)
	}

	// Convert to protobuf response
	pbLocation, err := mapInventoryLocationToProto(location)
	if err != nil {
		h.logger.Error("Failed to map inventory location to proto", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to map inventory location to proto")
	}

	return &pb.InventoryLocationResponse{
		InventoryLocation: pbLocation,
	}, nil
}

// RemoveInventoryFromLocation removes inventory from a specific warehouse location
func (h *InventoryHandler) RemoveInventoryFromLocation(ctx context.Context, req *pb.RemoveInventoryFromLocationRequest) (*pb.InventoryLocationResponse, error) {
	h.logger.Info("RemoveInventoryFromLocation request received",
		zap.String("inventory_item_id", req.InventoryItemId),
		zap.String("warehouse_id", req.WarehouseId),
		zap.Int32("quantity", req.Quantity))

	// Remove inventory from location
	location, err := h.inventoryService.RemoveInventoryFromLocation(
		ctx,
		req.InventoryItemId,
		req.WarehouseId,
		int(req.Quantity),
		req.ReferenceId,
		req.ReferenceType,
		req.Notes,
	)

	if err != nil {
		h.logger.Error("Failed to remove inventory from location", zap.Error(err))
		return nil, mapErrorToGRPCStatus(err)
	}

	// Convert to protobuf response
	pbLocation, err := mapInventoryLocationToProto(location)
	if err != nil {
		h.logger.Error("Failed to map inventory location to proto", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to map inventory location to proto")
	}

	return &pb.InventoryLocationResponse{
		InventoryLocation: pbLocation,
	}, nil
}

// GetInventoryByLocation retrieves inventory items at a specific warehouse
func (h *InventoryHandler) GetInventoryByLocation(ctx context.Context, req *pb.GetInventoryByLocationRequest) (*pb.ListInventoryLocationsResponse, error) {
	h.logger.Info("GetInventoryByLocation request received",
		zap.String("warehouse_id", req.WarehouseId),
		zap.Int32("page", req.Page),
		zap.Int32("limit", req.Limit))

	// Get inventory by location
	locations, total, err := h.inventoryService.GetInventoryByLocation(
		ctx,
		req.WarehouseId,
		int(req.Page),
		int(req.Limit),
	)

	if err != nil {
		h.logger.Error("Failed to get inventory by location", zap.Error(err))
		return nil, mapErrorToGRPCStatus(err)
	}

	// Convert to protobuf response
	var pbLocations []*pb.InventoryLocation
	for i := range locations {
		pbLocation, err := mapInventoryLocationToProto(&locations[i])
		if err != nil {
			h.logger.Error("Failed to map inventory location to proto", zap.Error(err))
			continue
		}
		pbLocations = append(pbLocations, pbLocation)
	}

	return &pb.ListInventoryLocationsResponse{
		InventoryLocations: pbLocations,
		Total:              int32(total),
	}, nil
}

// ReserveInventory creates temporary holds on inventory items
func (h *InventoryHandler) ReserveInventory(ctx context.Context, req *pb.ReserveInventoryRequest) (*pb.ReservationResponse, error) {
	h.logger.Info("ReserveInventory request received",
		zap.Int("items_count", len(req.Items)),
		zap.String("reference_type", req.ReferenceType))

	// Convert reservation items
	var items []models.ReservationItem
	for _, item := range req.Items {
		var warehouseID *string
		if item.WarehouseId != nil && item.WarehouseId.Value != "" {
			warehouseID = &item.WarehouseId.Value
		}

		items = append(items, models.ReservationItem{
			InventoryItemID: item.InventoryItemId,
			Quantity:        int(item.Quantity),
			WarehouseID:     warehouseID,
		})
	}

	// Create reservation
	reservation, err := h.inventoryService.ReserveInventory(
		ctx,
		items,
		req.ReferenceId,
		req.ReferenceType,
		int(req.ReservationMinutes),
	)

	if err != nil {
		h.logger.Error("Failed to reserve inventory", zap.Error(err))
		return nil, mapErrorToGRPCStatus(err)
	}

	// Convert to protobuf response
	pbReservation, err := mapReservationToProto(reservation)
	if err != nil {
		h.logger.Error("Failed to map reservation to proto", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to map reservation to proto")
	}

	return &pb.ReservationResponse{
		Reservation: pbReservation,
		Success:     true,
		Message:     "Reservation created successfully",
	}, nil
}

// ConfirmReservation confirms a pending reservation
func (h *InventoryHandler) ConfirmReservation(ctx context.Context, req *pb.ConfirmReservationRequest) (*pb.ReservationResponse, error) {
	h.logger.Info("ConfirmReservation request received", zap.String("reservation_id", req.ReservationId))

	// Confirm reservation
	reservation, err := h.inventoryService.ConfirmReservation(ctx, req.ReservationId)
	if err != nil {
		h.logger.Error("Failed to confirm reservation", zap.Error(err))
		return nil, mapErrorToGRPCStatus(err)
	}

	// Convert to protobuf response
	pbReservation, err := mapReservationToProto(reservation)
	if err != nil {
		h.logger.Error("Failed to map reservation to proto", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to map reservation to proto")
	}

	return &pb.ReservationResponse{
		Reservation: pbReservation,
		Success:     true,
		Message:     "Reservation confirmed successfully",
	}, nil
}

// CancelReservation cancels a pending reservation
func (h *InventoryHandler) CancelReservation(ctx context.Context, req *pb.CancelReservationRequest) (*pb.ReservationResponse, error) {
	h.logger.Info("CancelReservation request received", zap.String("reservation_id", req.ReservationId))

	// Cancel reservation
	reservation, err := h.inventoryService.CancelReservation(ctx, req.ReservationId)
	if err != nil {
		h.logger.Error("Failed to cancel reservation", zap.Error(err))
		return nil, mapErrorToGRPCStatus(err)
	}

	// Convert to protobuf response
	pbReservation, err := mapReservationToProto(reservation)
	if err != nil {
		h.logger.Error("Failed to map reservation to proto", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to map reservation to proto")
	}

	return &pb.ReservationResponse{
		Reservation: pbReservation,
		Success:     true,
		Message:     "Reservation cancelled successfully",
	}, nil
}

// CheckInventoryAvailability checks if requested quantities are available
func (h *InventoryHandler) CheckInventoryAvailability(ctx context.Context, req *pb.CheckInventoryAvailabilityRequest) (*pb.InventoryAvailabilityResponse, error) {
	h.logger.Info("CheckInventoryAvailability request received", zap.Int("items_count", len(req.Items)))

	// Convert availability check items
	var items []models.AvailabilityCheckItem
	for _, item := range req.Items {
		var variantID *string
		if item.VariantId != nil && item.VariantId.Value != "" {
			variantID = &item.VariantId.Value
		}

		items = append(items, models.AvailabilityCheckItem{
			ProductID: item.ProductId,
			VariantID: variantID,
			SKU:       item.Sku,
			Quantity:  int(item.Quantity),
		})
	}

	// Check availability
	availabilityResults, allAvailable, err := h.inventoryService.CheckInventoryAvailability(ctx, items)
	if err != nil {
		h.logger.Error("Failed to check inventory availability", zap.Error(err))
		return nil, mapErrorToGRPCStatus(err)
	}

	// Convert to protobuf response
	var pbItems []*pb.ItemAvailability
	for _, result := range availabilityResults {
		var variantID *wrappers.StringValue
		if result.VariantID != nil {
			variantID = &wrappers.StringValue{Value: *result.VariantID}
		}

		pbItems = append(pbItems, &pb.ItemAvailability{
			ProductId:         result.ProductID,
			VariantId:         variantID,
			Sku:               result.SKU,
			RequestedQuantity: int32(result.RequestedQuantity),
			AvailableQuantity: int32(result.AvailableQuantity),
			IsAvailable:       result.IsAvailable,
			Status:            result.Status,
		})
	}

	return &pb.InventoryAvailabilityResponse{
		Items:        pbItems,
		AllAvailable: allAvailable,
	}, nil
}

// BulkUpdateInventory updates multiple inventory items in a single call
func (h *InventoryHandler) BulkUpdateInventory(ctx context.Context, req *pb.BulkUpdateInventoryRequest) (*pb.BulkUpdateInventoryResponse, error) {
	h.logger.Info("BulkUpdateInventory request received", zap.Int("items_count", len(req.Items)))

	// This is a placeholder implementation
	// In a real implementation, you would process each item and update inventory accordingly
	// For now, we'll just return a mock response

	var results []*pb.BulkUpdateResult
	successCount := 0
	failureCount := 0

	for _, item := range req.Items {
		// Try to get the inventory item
		inventoryItem, err := h.inventoryService.GetInventoryItem(ctx, "", "", item.Sku)
		if err != nil {
			// Item not found or other error
			results = append(results, &pb.BulkUpdateResult{
				Sku:     item.Sku,
				Success: false,
				Message: "Item not found or error: " + err.Error(),
			})
			failureCount++
			continue
		}

		// Determine if we're adding or removing inventory
		var updateErr error

		if item.QuantityDelta > 0 {
			// Adding inventory
			_, updateErr = h.inventoryService.AddInventoryToLocation(
				ctx,
				inventoryItem.ID,
				item.WarehouseId,
				int(item.QuantityDelta),
				item.ReferenceId,
				item.ReferenceType,
				item.Notes,
			)
		} else if item.QuantityDelta < 0 {
			// Removing inventory
			_, updateErr = h.inventoryService.RemoveInventoryFromLocation(
				ctx,
				inventoryItem.ID,
				item.WarehouseId,
				int(-item.QuantityDelta), // Convert negative to positive
				item.ReferenceId,
				item.ReferenceType,
				item.Notes,
			)
		} else {
			// No change
			results = append(results, &pb.BulkUpdateResult{
				Sku:     item.Sku,
				Success: true,
				Message: "No change (quantity delta is 0)",
			})
			successCount++
			continue
		}

		if updateErr != nil {
			// Update failed
			results = append(results, &pb.BulkUpdateResult{
				Sku:     item.Sku,
				Success: false,
				Message: "Update failed: " + updateErr.Error(),
			})
			failureCount++
		} else {
			// Update succeeded, get the updated item
			updatedItem, err := h.inventoryService.GetInventoryItem(ctx, inventoryItem.ID, "", "")
			if err != nil {
				// Failed to get updated item
				results = append(results, &pb.BulkUpdateResult{
					Sku:     item.Sku,
					Success: true,
					Message: "Update succeeded but failed to retrieve updated item",
				})
			} else {
				// Convert to protobuf
				pbItem, err := mapInventoryItemToProto(updatedItem)
				if err != nil {
					results = append(results, &pb.BulkUpdateResult{
						Sku:     item.Sku,
						Success: true,
						Message: "Update succeeded but failed to convert updated item",
					})
				} else {
					results = append(results, &pb.BulkUpdateResult{
						Sku:         item.Sku,
						Success:     true,
						Message:     "Update succeeded",
						UpdatedItem: pbItem,
					})
				}
			}
			successCount++
		}
	}

	return &pb.BulkUpdateInventoryResponse{
		Results:      results,
		SuccessCount: int32(successCount),
		FailureCount: int32(failureCount),
	}, nil
}

// mapReservationToProto converts a domain reservation to a protobuf message
func mapReservationToProto(reservation *models.InventoryReservation) (*pb.InventoryReservation, error) {
	// Convert timestamps
	createdAt := &timestamp.Timestamp{
		Seconds: reservation.CreatedAt.Unix(),
		Nanos:   int32(reservation.CreatedAt.Nanosecond()),
	}
	updatedAt := &timestamp.Timestamp{
		Seconds: reservation.UpdatedAt.Unix(),
		Nanos:   int32(reservation.UpdatedAt.Nanosecond()),
	}
	expirationTime := &timestamp.Timestamp{
		Seconds: reservation.ExpirationTime.Unix(),
		Nanos:   int32(reservation.ExpirationTime.Nanosecond()),
	}

	// Convert optional fields
	var warehouseID, referenceID *wrappers.StringValue
	if reservation.WarehouseID != nil {
		warehouseID = &wrappers.StringValue{Value: *reservation.WarehouseID}
	}
	if reservation.ReferenceID != nil {
		referenceID = &wrappers.StringValue{Value: *reservation.ReferenceID}
	}

	return &pb.InventoryReservation{
		Id:              reservation.ID,
		InventoryItemId: reservation.InventoryItemID,
		WarehouseId:     warehouseID,
		Quantity:        int32(reservation.Quantity),
		Status:          reservation.Status,
		ExpirationTime:  expirationTime,
		ReferenceId:     referenceID,
		ReferenceType:   reservation.ReferenceType,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
	}, nil
}
