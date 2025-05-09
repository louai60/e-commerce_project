// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v6.30.1
// source: proto/inventory.proto

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	InventoryService_CreateInventoryItem_FullMethodName         = "/inventory.InventoryService/CreateInventoryItem"
	InventoryService_GetInventoryItem_FullMethodName            = "/inventory.InventoryService/GetInventoryItem"
	InventoryService_UpdateInventoryItem_FullMethodName         = "/inventory.InventoryService/UpdateInventoryItem"
	InventoryService_ListInventoryItems_FullMethodName          = "/inventory.InventoryService/ListInventoryItems"
	InventoryService_CreateWarehouse_FullMethodName             = "/inventory.InventoryService/CreateWarehouse"
	InventoryService_GetWarehouse_FullMethodName                = "/inventory.InventoryService/GetWarehouse"
	InventoryService_UpdateWarehouse_FullMethodName             = "/inventory.InventoryService/UpdateWarehouse"
	InventoryService_ListWarehouses_FullMethodName              = "/inventory.InventoryService/ListWarehouses"
	InventoryService_AddInventoryToLocation_FullMethodName      = "/inventory.InventoryService/AddInventoryToLocation"
	InventoryService_RemoveInventoryFromLocation_FullMethodName = "/inventory.InventoryService/RemoveInventoryFromLocation"
	InventoryService_GetInventoryByLocation_FullMethodName      = "/inventory.InventoryService/GetInventoryByLocation"
	InventoryService_ReserveInventory_FullMethodName            = "/inventory.InventoryService/ReserveInventory"
	InventoryService_ConfirmReservation_FullMethodName          = "/inventory.InventoryService/ConfirmReservation"
	InventoryService_CancelReservation_FullMethodName           = "/inventory.InventoryService/CancelReservation"
	InventoryService_CheckInventoryAvailability_FullMethodName  = "/inventory.InventoryService/CheckInventoryAvailability"
	InventoryService_BulkUpdateInventory_FullMethodName         = "/inventory.InventoryService/BulkUpdateInventory"
)

// InventoryServiceClient is the client API for InventoryService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type InventoryServiceClient interface {
	// Product inventory management
	CreateInventoryItem(ctx context.Context, in *CreateInventoryItemRequest, opts ...grpc.CallOption) (*InventoryItemResponse, error)
	GetInventoryItem(ctx context.Context, in *GetInventoryItemRequest, opts ...grpc.CallOption) (*InventoryItemResponse, error)
	UpdateInventoryItem(ctx context.Context, in *UpdateInventoryItemRequest, opts ...grpc.CallOption) (*InventoryItemResponse, error)
	ListInventoryItems(ctx context.Context, in *ListInventoryItemsRequest, opts ...grpc.CallOption) (*ListInventoryItemsResponse, error)
	// Warehouse operations
	CreateWarehouse(ctx context.Context, in *CreateWarehouseRequest, opts ...grpc.CallOption) (*WarehouseResponse, error)
	GetWarehouse(ctx context.Context, in *GetWarehouseRequest, opts ...grpc.CallOption) (*WarehouseResponse, error)
	UpdateWarehouse(ctx context.Context, in *UpdateWarehouseRequest, opts ...grpc.CallOption) (*WarehouseResponse, error)
	ListWarehouses(ctx context.Context, in *ListWarehousesRequest, opts ...grpc.CallOption) (*ListWarehousesResponse, error)
	// Inventory location operations
	AddInventoryToLocation(ctx context.Context, in *AddInventoryToLocationRequest, opts ...grpc.CallOption) (*InventoryLocationResponse, error)
	RemoveInventoryFromLocation(ctx context.Context, in *RemoveInventoryFromLocationRequest, opts ...grpc.CallOption) (*InventoryLocationResponse, error)
	GetInventoryByLocation(ctx context.Context, in *GetInventoryByLocationRequest, opts ...grpc.CallOption) (*ListInventoryLocationsResponse, error)
	// Reservation operations
	ReserveInventory(ctx context.Context, in *ReserveInventoryRequest, opts ...grpc.CallOption) (*ReservationResponse, error)
	ConfirmReservation(ctx context.Context, in *ConfirmReservationRequest, opts ...grpc.CallOption) (*ReservationResponse, error)
	CancelReservation(ctx context.Context, in *CancelReservationRequest, opts ...grpc.CallOption) (*ReservationResponse, error)
	// Inventory check operations
	CheckInventoryAvailability(ctx context.Context, in *CheckInventoryAvailabilityRequest, opts ...grpc.CallOption) (*InventoryAvailabilityResponse, error)
	// Bulk operations
	BulkUpdateInventory(ctx context.Context, in *BulkUpdateInventoryRequest, opts ...grpc.CallOption) (*BulkUpdateInventoryResponse, error)
}

type inventoryServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewInventoryServiceClient(cc grpc.ClientConnInterface) InventoryServiceClient {
	return &inventoryServiceClient{cc}
}

func (c *inventoryServiceClient) CreateInventoryItem(ctx context.Context, in *CreateInventoryItemRequest, opts ...grpc.CallOption) (*InventoryItemResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(InventoryItemResponse)
	err := c.cc.Invoke(ctx, InventoryService_CreateInventoryItem_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *inventoryServiceClient) GetInventoryItem(ctx context.Context, in *GetInventoryItemRequest, opts ...grpc.CallOption) (*InventoryItemResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(InventoryItemResponse)
	err := c.cc.Invoke(ctx, InventoryService_GetInventoryItem_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *inventoryServiceClient) UpdateInventoryItem(ctx context.Context, in *UpdateInventoryItemRequest, opts ...grpc.CallOption) (*InventoryItemResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(InventoryItemResponse)
	err := c.cc.Invoke(ctx, InventoryService_UpdateInventoryItem_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *inventoryServiceClient) ListInventoryItems(ctx context.Context, in *ListInventoryItemsRequest, opts ...grpc.CallOption) (*ListInventoryItemsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ListInventoryItemsResponse)
	err := c.cc.Invoke(ctx, InventoryService_ListInventoryItems_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *inventoryServiceClient) CreateWarehouse(ctx context.Context, in *CreateWarehouseRequest, opts ...grpc.CallOption) (*WarehouseResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(WarehouseResponse)
	err := c.cc.Invoke(ctx, InventoryService_CreateWarehouse_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *inventoryServiceClient) GetWarehouse(ctx context.Context, in *GetWarehouseRequest, opts ...grpc.CallOption) (*WarehouseResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(WarehouseResponse)
	err := c.cc.Invoke(ctx, InventoryService_GetWarehouse_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *inventoryServiceClient) UpdateWarehouse(ctx context.Context, in *UpdateWarehouseRequest, opts ...grpc.CallOption) (*WarehouseResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(WarehouseResponse)
	err := c.cc.Invoke(ctx, InventoryService_UpdateWarehouse_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *inventoryServiceClient) ListWarehouses(ctx context.Context, in *ListWarehousesRequest, opts ...grpc.CallOption) (*ListWarehousesResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ListWarehousesResponse)
	err := c.cc.Invoke(ctx, InventoryService_ListWarehouses_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *inventoryServiceClient) AddInventoryToLocation(ctx context.Context, in *AddInventoryToLocationRequest, opts ...grpc.CallOption) (*InventoryLocationResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(InventoryLocationResponse)
	err := c.cc.Invoke(ctx, InventoryService_AddInventoryToLocation_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *inventoryServiceClient) RemoveInventoryFromLocation(ctx context.Context, in *RemoveInventoryFromLocationRequest, opts ...grpc.CallOption) (*InventoryLocationResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(InventoryLocationResponse)
	err := c.cc.Invoke(ctx, InventoryService_RemoveInventoryFromLocation_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *inventoryServiceClient) GetInventoryByLocation(ctx context.Context, in *GetInventoryByLocationRequest, opts ...grpc.CallOption) (*ListInventoryLocationsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ListInventoryLocationsResponse)
	err := c.cc.Invoke(ctx, InventoryService_GetInventoryByLocation_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *inventoryServiceClient) ReserveInventory(ctx context.Context, in *ReserveInventoryRequest, opts ...grpc.CallOption) (*ReservationResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ReservationResponse)
	err := c.cc.Invoke(ctx, InventoryService_ReserveInventory_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *inventoryServiceClient) ConfirmReservation(ctx context.Context, in *ConfirmReservationRequest, opts ...grpc.CallOption) (*ReservationResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ReservationResponse)
	err := c.cc.Invoke(ctx, InventoryService_ConfirmReservation_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *inventoryServiceClient) CancelReservation(ctx context.Context, in *CancelReservationRequest, opts ...grpc.CallOption) (*ReservationResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ReservationResponse)
	err := c.cc.Invoke(ctx, InventoryService_CancelReservation_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *inventoryServiceClient) CheckInventoryAvailability(ctx context.Context, in *CheckInventoryAvailabilityRequest, opts ...grpc.CallOption) (*InventoryAvailabilityResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(InventoryAvailabilityResponse)
	err := c.cc.Invoke(ctx, InventoryService_CheckInventoryAvailability_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *inventoryServiceClient) BulkUpdateInventory(ctx context.Context, in *BulkUpdateInventoryRequest, opts ...grpc.CallOption) (*BulkUpdateInventoryResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(BulkUpdateInventoryResponse)
	err := c.cc.Invoke(ctx, InventoryService_BulkUpdateInventory_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// InventoryServiceServer is the server API for InventoryService service.
// All implementations must embed UnimplementedInventoryServiceServer
// for forward compatibility.
type InventoryServiceServer interface {
	// Product inventory management
	CreateInventoryItem(context.Context, *CreateInventoryItemRequest) (*InventoryItemResponse, error)
	GetInventoryItem(context.Context, *GetInventoryItemRequest) (*InventoryItemResponse, error)
	UpdateInventoryItem(context.Context, *UpdateInventoryItemRequest) (*InventoryItemResponse, error)
	ListInventoryItems(context.Context, *ListInventoryItemsRequest) (*ListInventoryItemsResponse, error)
	// Warehouse operations
	CreateWarehouse(context.Context, *CreateWarehouseRequest) (*WarehouseResponse, error)
	GetWarehouse(context.Context, *GetWarehouseRequest) (*WarehouseResponse, error)
	UpdateWarehouse(context.Context, *UpdateWarehouseRequest) (*WarehouseResponse, error)
	ListWarehouses(context.Context, *ListWarehousesRequest) (*ListWarehousesResponse, error)
	// Inventory location operations
	AddInventoryToLocation(context.Context, *AddInventoryToLocationRequest) (*InventoryLocationResponse, error)
	RemoveInventoryFromLocation(context.Context, *RemoveInventoryFromLocationRequest) (*InventoryLocationResponse, error)
	GetInventoryByLocation(context.Context, *GetInventoryByLocationRequest) (*ListInventoryLocationsResponse, error)
	// Reservation operations
	ReserveInventory(context.Context, *ReserveInventoryRequest) (*ReservationResponse, error)
	ConfirmReservation(context.Context, *ConfirmReservationRequest) (*ReservationResponse, error)
	CancelReservation(context.Context, *CancelReservationRequest) (*ReservationResponse, error)
	// Inventory check operations
	CheckInventoryAvailability(context.Context, *CheckInventoryAvailabilityRequest) (*InventoryAvailabilityResponse, error)
	// Bulk operations
	BulkUpdateInventory(context.Context, *BulkUpdateInventoryRequest) (*BulkUpdateInventoryResponse, error)
	mustEmbedUnimplementedInventoryServiceServer()
}

// UnimplementedInventoryServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedInventoryServiceServer struct{}

func (UnimplementedInventoryServiceServer) CreateInventoryItem(context.Context, *CreateInventoryItemRequest) (*InventoryItemResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateInventoryItem not implemented")
}
func (UnimplementedInventoryServiceServer) GetInventoryItem(context.Context, *GetInventoryItemRequest) (*InventoryItemResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetInventoryItem not implemented")
}
func (UnimplementedInventoryServiceServer) UpdateInventoryItem(context.Context, *UpdateInventoryItemRequest) (*InventoryItemResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateInventoryItem not implemented")
}
func (UnimplementedInventoryServiceServer) ListInventoryItems(context.Context, *ListInventoryItemsRequest) (*ListInventoryItemsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListInventoryItems not implemented")
}
func (UnimplementedInventoryServiceServer) CreateWarehouse(context.Context, *CreateWarehouseRequest) (*WarehouseResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateWarehouse not implemented")
}
func (UnimplementedInventoryServiceServer) GetWarehouse(context.Context, *GetWarehouseRequest) (*WarehouseResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetWarehouse not implemented")
}
func (UnimplementedInventoryServiceServer) UpdateWarehouse(context.Context, *UpdateWarehouseRequest) (*WarehouseResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateWarehouse not implemented")
}
func (UnimplementedInventoryServiceServer) ListWarehouses(context.Context, *ListWarehousesRequest) (*ListWarehousesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListWarehouses not implemented")
}
func (UnimplementedInventoryServiceServer) AddInventoryToLocation(context.Context, *AddInventoryToLocationRequest) (*InventoryLocationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddInventoryToLocation not implemented")
}
func (UnimplementedInventoryServiceServer) RemoveInventoryFromLocation(context.Context, *RemoveInventoryFromLocationRequest) (*InventoryLocationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoveInventoryFromLocation not implemented")
}
func (UnimplementedInventoryServiceServer) GetInventoryByLocation(context.Context, *GetInventoryByLocationRequest) (*ListInventoryLocationsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetInventoryByLocation not implemented")
}
func (UnimplementedInventoryServiceServer) ReserveInventory(context.Context, *ReserveInventoryRequest) (*ReservationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReserveInventory not implemented")
}
func (UnimplementedInventoryServiceServer) ConfirmReservation(context.Context, *ConfirmReservationRequest) (*ReservationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ConfirmReservation not implemented")
}
func (UnimplementedInventoryServiceServer) CancelReservation(context.Context, *CancelReservationRequest) (*ReservationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CancelReservation not implemented")
}
func (UnimplementedInventoryServiceServer) CheckInventoryAvailability(context.Context, *CheckInventoryAvailabilityRequest) (*InventoryAvailabilityResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CheckInventoryAvailability not implemented")
}
func (UnimplementedInventoryServiceServer) BulkUpdateInventory(context.Context, *BulkUpdateInventoryRequest) (*BulkUpdateInventoryResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BulkUpdateInventory not implemented")
}
func (UnimplementedInventoryServiceServer) mustEmbedUnimplementedInventoryServiceServer() {}
func (UnimplementedInventoryServiceServer) testEmbeddedByValue()                          {}

// UnsafeInventoryServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to InventoryServiceServer will
// result in compilation errors.
type UnsafeInventoryServiceServer interface {
	mustEmbedUnimplementedInventoryServiceServer()
}

func RegisterInventoryServiceServer(s grpc.ServiceRegistrar, srv InventoryServiceServer) {
	// If the following call pancis, it indicates UnimplementedInventoryServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&InventoryService_ServiceDesc, srv)
}

func _InventoryService_CreateInventoryItem_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateInventoryItemRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(InventoryServiceServer).CreateInventoryItem(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: InventoryService_CreateInventoryItem_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(InventoryServiceServer).CreateInventoryItem(ctx, req.(*CreateInventoryItemRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _InventoryService_GetInventoryItem_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetInventoryItemRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(InventoryServiceServer).GetInventoryItem(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: InventoryService_GetInventoryItem_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(InventoryServiceServer).GetInventoryItem(ctx, req.(*GetInventoryItemRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _InventoryService_UpdateInventoryItem_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateInventoryItemRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(InventoryServiceServer).UpdateInventoryItem(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: InventoryService_UpdateInventoryItem_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(InventoryServiceServer).UpdateInventoryItem(ctx, req.(*UpdateInventoryItemRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _InventoryService_ListInventoryItems_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListInventoryItemsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(InventoryServiceServer).ListInventoryItems(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: InventoryService_ListInventoryItems_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(InventoryServiceServer).ListInventoryItems(ctx, req.(*ListInventoryItemsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _InventoryService_CreateWarehouse_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateWarehouseRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(InventoryServiceServer).CreateWarehouse(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: InventoryService_CreateWarehouse_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(InventoryServiceServer).CreateWarehouse(ctx, req.(*CreateWarehouseRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _InventoryService_GetWarehouse_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetWarehouseRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(InventoryServiceServer).GetWarehouse(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: InventoryService_GetWarehouse_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(InventoryServiceServer).GetWarehouse(ctx, req.(*GetWarehouseRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _InventoryService_UpdateWarehouse_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateWarehouseRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(InventoryServiceServer).UpdateWarehouse(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: InventoryService_UpdateWarehouse_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(InventoryServiceServer).UpdateWarehouse(ctx, req.(*UpdateWarehouseRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _InventoryService_ListWarehouses_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListWarehousesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(InventoryServiceServer).ListWarehouses(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: InventoryService_ListWarehouses_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(InventoryServiceServer).ListWarehouses(ctx, req.(*ListWarehousesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _InventoryService_AddInventoryToLocation_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddInventoryToLocationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(InventoryServiceServer).AddInventoryToLocation(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: InventoryService_AddInventoryToLocation_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(InventoryServiceServer).AddInventoryToLocation(ctx, req.(*AddInventoryToLocationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _InventoryService_RemoveInventoryFromLocation_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RemoveInventoryFromLocationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(InventoryServiceServer).RemoveInventoryFromLocation(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: InventoryService_RemoveInventoryFromLocation_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(InventoryServiceServer).RemoveInventoryFromLocation(ctx, req.(*RemoveInventoryFromLocationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _InventoryService_GetInventoryByLocation_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetInventoryByLocationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(InventoryServiceServer).GetInventoryByLocation(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: InventoryService_GetInventoryByLocation_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(InventoryServiceServer).GetInventoryByLocation(ctx, req.(*GetInventoryByLocationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _InventoryService_ReserveInventory_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReserveInventoryRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(InventoryServiceServer).ReserveInventory(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: InventoryService_ReserveInventory_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(InventoryServiceServer).ReserveInventory(ctx, req.(*ReserveInventoryRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _InventoryService_ConfirmReservation_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ConfirmReservationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(InventoryServiceServer).ConfirmReservation(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: InventoryService_ConfirmReservation_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(InventoryServiceServer).ConfirmReservation(ctx, req.(*ConfirmReservationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _InventoryService_CancelReservation_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CancelReservationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(InventoryServiceServer).CancelReservation(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: InventoryService_CancelReservation_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(InventoryServiceServer).CancelReservation(ctx, req.(*CancelReservationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _InventoryService_CheckInventoryAvailability_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CheckInventoryAvailabilityRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(InventoryServiceServer).CheckInventoryAvailability(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: InventoryService_CheckInventoryAvailability_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(InventoryServiceServer).CheckInventoryAvailability(ctx, req.(*CheckInventoryAvailabilityRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _InventoryService_BulkUpdateInventory_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BulkUpdateInventoryRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(InventoryServiceServer).BulkUpdateInventory(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: InventoryService_BulkUpdateInventory_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(InventoryServiceServer).BulkUpdateInventory(ctx, req.(*BulkUpdateInventoryRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// InventoryService_ServiceDesc is the grpc.ServiceDesc for InventoryService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var InventoryService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "inventory.InventoryService",
	HandlerType: (*InventoryServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateInventoryItem",
			Handler:    _InventoryService_CreateInventoryItem_Handler,
		},
		{
			MethodName: "GetInventoryItem",
			Handler:    _InventoryService_GetInventoryItem_Handler,
		},
		{
			MethodName: "UpdateInventoryItem",
			Handler:    _InventoryService_UpdateInventoryItem_Handler,
		},
		{
			MethodName: "ListInventoryItems",
			Handler:    _InventoryService_ListInventoryItems_Handler,
		},
		{
			MethodName: "CreateWarehouse",
			Handler:    _InventoryService_CreateWarehouse_Handler,
		},
		{
			MethodName: "GetWarehouse",
			Handler:    _InventoryService_GetWarehouse_Handler,
		},
		{
			MethodName: "UpdateWarehouse",
			Handler:    _InventoryService_UpdateWarehouse_Handler,
		},
		{
			MethodName: "ListWarehouses",
			Handler:    _InventoryService_ListWarehouses_Handler,
		},
		{
			MethodName: "AddInventoryToLocation",
			Handler:    _InventoryService_AddInventoryToLocation_Handler,
		},
		{
			MethodName: "RemoveInventoryFromLocation",
			Handler:    _InventoryService_RemoveInventoryFromLocation_Handler,
		},
		{
			MethodName: "GetInventoryByLocation",
			Handler:    _InventoryService_GetInventoryByLocation_Handler,
		},
		{
			MethodName: "ReserveInventory",
			Handler:    _InventoryService_ReserveInventory_Handler,
		},
		{
			MethodName: "ConfirmReservation",
			Handler:    _InventoryService_ConfirmReservation_Handler,
		},
		{
			MethodName: "CancelReservation",
			Handler:    _InventoryService_CancelReservation_Handler,
		},
		{
			MethodName: "CheckInventoryAvailability",
			Handler:    _InventoryService_CheckInventoryAvailability_Handler,
		},
		{
			MethodName: "BulkUpdateInventory",
			Handler:    _InventoryService_BulkUpdateInventory_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/inventory.proto",
}
