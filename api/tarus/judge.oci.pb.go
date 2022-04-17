// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.15.8
// source: api/tarus/judge.oci.proto

package tarus

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type OCIJudgeSession struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// StartedAt provides the time session begin.
	CreatedAt *timestamppb.Timestamp `protobuf:"bytes,1,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	// UpdatedAt provides the last time of a successful write.
	UpdatedAt    *timestamppb.Timestamp `protobuf:"bytes,2,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at,omitempty"`
	CommitStatus int32                  `protobuf:"varint,3,opt,name=commit_status,json=commitStatus,proto3" json:"commit_status,omitempty"`
	WorkerId     int32                  `protobuf:"varint,4,opt,name=worker_id,json=workerId,proto3" json:"worker_id,omitempty"`
	ContainerId  string                 `protobuf:"bytes,5,opt,name=container_id,json=containerId,proto3" json:"container_id,omitempty"`
	BinTarget    string                 `protobuf:"bytes,6,opt,name=bin_target,json=binTarget,proto3" json:"bin_target,omitempty"`
	HostWorkdir  string                 `protobuf:"bytes,101,opt,name=host_workdir,json=hostWorkdir,proto3" json:"host_workdir,omitempty"`
}

func (x *OCIJudgeSession) Reset() {
	*x = OCIJudgeSession{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_tarus_judge_oci_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *OCIJudgeSession) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*OCIJudgeSession) ProtoMessage() {}

func (x *OCIJudgeSession) ProtoReflect() protoreflect.Message {
	mi := &file_api_tarus_judge_oci_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use OCIJudgeSession.ProtoReflect.Descriptor instead.
func (*OCIJudgeSession) Descriptor() ([]byte, []int) {
	return file_api_tarus_judge_oci_proto_rawDescGZIP(), []int{0}
}

func (x *OCIJudgeSession) GetCreatedAt() *timestamppb.Timestamp {
	if x != nil {
		return x.CreatedAt
	}
	return nil
}

func (x *OCIJudgeSession) GetUpdatedAt() *timestamppb.Timestamp {
	if x != nil {
		return x.UpdatedAt
	}
	return nil
}

func (x *OCIJudgeSession) GetCommitStatus() int32 {
	if x != nil {
		return x.CommitStatus
	}
	return 0
}

func (x *OCIJudgeSession) GetWorkerId() int32 {
	if x != nil {
		return x.WorkerId
	}
	return 0
}

func (x *OCIJudgeSession) GetContainerId() string {
	if x != nil {
		return x.ContainerId
	}
	return ""
}

func (x *OCIJudgeSession) GetBinTarget() string {
	if x != nil {
		return x.BinTarget
	}
	return ""
}

func (x *OCIJudgeSession) GetHostWorkdir() string {
	if x != nil {
		return x.HostWorkdir
	}
	return ""
}

var File_api_tarus_judge_oci_proto protoreflect.FileDescriptor

var file_api_tarus_judge_oci_proto_rawDesc = []byte{
	0x0a, 0x19, 0x61, 0x70, 0x69, 0x2f, 0x74, 0x61, 0x72, 0x75, 0x73, 0x2f, 0x6a, 0x75, 0x64, 0x67,
	0x65, 0x2e, 0x6f, 0x63, 0x69, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x13, 0x74, 0x61, 0x72,
	0x75, 0x73, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x6a, 0x75, 0x64, 0x67, 0x65, 0x2e, 0x6f, 0x63, 0x69,
	0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x22, 0xae, 0x02, 0x0a, 0x0f, 0x4f, 0x43, 0x49, 0x4a, 0x75, 0x64, 0x67, 0x65, 0x53, 0x65,
	0x73, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x39, 0x0a, 0x0a, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64,
	0x5f, 0x61, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65,
	0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x09, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74,
	0x12, 0x39, 0x0a, 0x0a, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70,
	0x52, 0x09, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x12, 0x23, 0x0a, 0x0d, 0x63,
	0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x5f, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x0c, 0x63, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x12, 0x1b, 0x0a, 0x09, 0x77, 0x6f, 0x72, 0x6b, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x05, 0x52, 0x08, 0x77, 0x6f, 0x72, 0x6b, 0x65, 0x72, 0x49, 0x64, 0x12, 0x21, 0x0a,
	0x0c, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x18, 0x05, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x0b, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x49, 0x64,
	0x12, 0x1d, 0x0a, 0x0a, 0x62, 0x69, 0x6e, 0x5f, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x18, 0x06,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x62, 0x69, 0x6e, 0x54, 0x61, 0x72, 0x67, 0x65, 0x74, 0x12,
	0x21, 0x0a, 0x0c, 0x68, 0x6f, 0x73, 0x74, 0x5f, 0x77, 0x6f, 0x72, 0x6b, 0x64, 0x69, 0x72, 0x18,
	0x65, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x68, 0x6f, 0x73, 0x74, 0x57, 0x6f, 0x72, 0x6b, 0x64,
	0x69, 0x72, 0x42, 0x0e, 0x5a, 0x0c, 0x74, 0x61, 0x72, 0x75, 0x73, 0x2f, 0x3b, 0x74, 0x61, 0x72,
	0x75, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_api_tarus_judge_oci_proto_rawDescOnce sync.Once
	file_api_tarus_judge_oci_proto_rawDescData = file_api_tarus_judge_oci_proto_rawDesc
)

func file_api_tarus_judge_oci_proto_rawDescGZIP() []byte {
	file_api_tarus_judge_oci_proto_rawDescOnce.Do(func() {
		file_api_tarus_judge_oci_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_tarus_judge_oci_proto_rawDescData)
	})
	return file_api_tarus_judge_oci_proto_rawDescData
}

var file_api_tarus_judge_oci_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_api_tarus_judge_oci_proto_goTypes = []interface{}{
	(*OCIJudgeSession)(nil),       // 0: tarus.api.judge.oci.OCIJudgeSession
	(*timestamppb.Timestamp)(nil), // 1: google.protobuf.Timestamp
}
var file_api_tarus_judge_oci_proto_depIdxs = []int32{
	1, // 0: tarus.api.judge.oci.OCIJudgeSession.created_at:type_name -> google.protobuf.Timestamp
	1, // 1: tarus.api.judge.oci.OCIJudgeSession.updated_at:type_name -> google.protobuf.Timestamp
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_api_tarus_judge_oci_proto_init() }
func file_api_tarus_judge_oci_proto_init() {
	if File_api_tarus_judge_oci_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_api_tarus_judge_oci_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*OCIJudgeSession); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_api_tarus_judge_oci_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_api_tarus_judge_oci_proto_goTypes,
		DependencyIndexes: file_api_tarus_judge_oci_proto_depIdxs,
		MessageInfos:      file_api_tarus_judge_oci_proto_msgTypes,
	}.Build()
	File_api_tarus_judge_oci_proto = out.File
	file_api_tarus_judge_oci_proto_rawDesc = nil
	file_api_tarus_judge_oci_proto_goTypes = nil
	file_api_tarus_judge_oci_proto_depIdxs = nil
}