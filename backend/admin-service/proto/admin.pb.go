// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v6.30.1
// source: proto/admin.proto

package proto

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Request message for GetDashboardStats
type GetDashboardStatsRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetDashboardStatsRequest) Reset() {
	*x = GetDashboardStatsRequest{}
	mi := &file_proto_admin_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetDashboardStatsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetDashboardStatsRequest) ProtoMessage() {}

func (x *GetDashboardStatsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_admin_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetDashboardStatsRequest.ProtoReflect.Descriptor instead.
func (*GetDashboardStatsRequest) Descriptor() ([]byte, []int) {
	return file_proto_admin_proto_rawDescGZIP(), []int{0}
}

// Response message for GetDashboardStats
type GetDashboardStatsResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	TotalUsers    int64                  `protobuf:"varint,1,opt,name=total_users,json=totalUsers,proto3" json:"total_users,omitempty"`
	TotalProducts int64                  `protobuf:"varint,2,opt,name=total_products,json=totalProducts,proto3" json:"total_products,omitempty"`
	TotalRevenue  float64                `protobuf:"fixed64,3,opt,name=total_revenue,json=totalRevenue,proto3" json:"total_revenue,omitempty"`
	TotalOrders   int64                  `protobuf:"varint,4,opt,name=total_orders,json=totalOrders,proto3" json:"total_orders,omitempty"` // Add more stats as needed
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetDashboardStatsResponse) Reset() {
	*x = GetDashboardStatsResponse{}
	mi := &file_proto_admin_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetDashboardStatsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetDashboardStatsResponse) ProtoMessage() {}

func (x *GetDashboardStatsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_admin_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetDashboardStatsResponse.ProtoReflect.Descriptor instead.
func (*GetDashboardStatsResponse) Descriptor() ([]byte, []int) {
	return file_proto_admin_proto_rawDescGZIP(), []int{1}
}

func (x *GetDashboardStatsResponse) GetTotalUsers() int64 {
	if x != nil {
		return x.TotalUsers
	}
	return 0
}

func (x *GetDashboardStatsResponse) GetTotalProducts() int64 {
	if x != nil {
		return x.TotalProducts
	}
	return 0
}

func (x *GetDashboardStatsResponse) GetTotalRevenue() float64 {
	if x != nil {
		return x.TotalRevenue
	}
	return 0
}

func (x *GetDashboardStatsResponse) GetTotalOrders() int64 {
	if x != nil {
		return x.TotalOrders
	}
	return 0
}

var File_proto_admin_proto protoreflect.FileDescriptor

const file_proto_admin_proto_rawDesc = "" +
	"\n" +
	"\x11proto/admin.proto\x12\x05admin\"\x1a\n" +
	"\x18GetDashboardStatsRequest\"\xab\x01\n" +
	"\x19GetDashboardStatsResponse\x12\x1f\n" +
	"\vtotal_users\x18\x01 \x01(\x03R\n" +
	"totalUsers\x12%\n" +
	"\x0etotal_products\x18\x02 \x01(\x03R\rtotalProducts\x12#\n" +
	"\rtotal_revenue\x18\x03 \x01(\x01R\ftotalRevenue\x12!\n" +
	"\ftotal_orders\x18\x04 \x01(\x03R\vtotalOrders2f\n" +
	"\fAdminService\x12V\n" +
	"\x11GetDashboardStats\x12\x1f.admin.GetDashboardStatsRequest\x1a .admin.GetDashboardStatsResponseBCZAgithub.com/louai60/e-commerce_project/backend/admin-service/protob\x06proto3"

var (
	file_proto_admin_proto_rawDescOnce sync.Once
	file_proto_admin_proto_rawDescData []byte
)

func file_proto_admin_proto_rawDescGZIP() []byte {
	file_proto_admin_proto_rawDescOnce.Do(func() {
		file_proto_admin_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_proto_admin_proto_rawDesc), len(file_proto_admin_proto_rawDesc)))
	})
	return file_proto_admin_proto_rawDescData
}

var file_proto_admin_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_proto_admin_proto_goTypes = []any{
	(*GetDashboardStatsRequest)(nil),  // 0: admin.GetDashboardStatsRequest
	(*GetDashboardStatsResponse)(nil), // 1: admin.GetDashboardStatsResponse
}
var file_proto_admin_proto_depIdxs = []int32{
	0, // 0: admin.AdminService.GetDashboardStats:input_type -> admin.GetDashboardStatsRequest
	1, // 1: admin.AdminService.GetDashboardStats:output_type -> admin.GetDashboardStatsResponse
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_proto_admin_proto_init() }
func file_proto_admin_proto_init() {
	if File_proto_admin_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_proto_admin_proto_rawDesc), len(file_proto_admin_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_proto_admin_proto_goTypes,
		DependencyIndexes: file_proto_admin_proto_depIdxs,
		MessageInfos:      file_proto_admin_proto_msgTypes,
	}.Build()
	File_proto_admin_proto = out.File
	file_proto_admin_proto_goTypes = nil
	file_proto_admin_proto_depIdxs = nil
}
