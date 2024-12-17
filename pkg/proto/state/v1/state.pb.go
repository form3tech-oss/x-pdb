// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.0
// 	protoc        (unknown)
// source: state/v1/state.proto

package state

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// LockRequest has the information to request a xpdb to be
// locked on a cluster.
type LockRequest struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// The lease_holder_identity of the lock that is going to be created on the cluster.
	LeaseHolderIdentity string `protobuf:"bytes,1,opt,name=lease_holder_identity,json=leaseHolderIdentity,proto3" json:"lease_holder_identity,omitempty"`
	// Namespace of the xpdb to be locked.
	Namespace string `protobuf:"bytes,2,opt,name=namespace,proto3" json:"namespace,omitempty"`
	// LabelSelector used on the xpdb resource to be locked.
	LabelSelector *LabelSelector `protobuf:"bytes,3,opt,name=label_selector,json=labelSelector,proto3" json:"label_selector,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *LockRequest) Reset() {
	*x = LockRequest{}
	mi := &file_state_v1_state_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *LockRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LockRequest) ProtoMessage() {}

func (x *LockRequest) ProtoReflect() protoreflect.Message {
	mi := &file_state_v1_state_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LockRequest.ProtoReflect.Descriptor instead.
func (*LockRequest) Descriptor() ([]byte, []int) {
	return file_state_v1_state_proto_rawDescGZIP(), []int{0}
}

func (x *LockRequest) GetLeaseHolderIdentity() string {
	if x != nil {
		return x.LeaseHolderIdentity
	}
	return ""
}

func (x *LockRequest) GetNamespace() string {
	if x != nil {
		return x.Namespace
	}
	return ""
}

func (x *LockRequest) GetLabelSelector() *LabelSelector {
	if x != nil {
		return x.LabelSelector
	}
	return nil
}

// LockResponse has the information on wether a lock was succeeded or not.
type LockResponse struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Refers to wether the lock was acquired or not.
	Acquired bool `protobuf:"varint,1,opt,name=acquired,proto3" json:"acquired,omitempty"`
	// Error has the error occured during the lock process.
	Error         string `protobuf:"bytes,2,opt,name=error,proto3" json:"error,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *LockResponse) Reset() {
	*x = LockResponse{}
	mi := &file_state_v1_state_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *LockResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LockResponse) ProtoMessage() {}

func (x *LockResponse) ProtoReflect() protoreflect.Message {
	mi := &file_state_v1_state_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LockResponse.ProtoReflect.Descriptor instead.
func (*LockResponse) Descriptor() ([]byte, []int) {
	return file_state_v1_state_proto_rawDescGZIP(), []int{1}
}

func (x *LockResponse) GetAcquired() bool {
	if x != nil {
		return x.Acquired
	}
	return false
}

func (x *LockResponse) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

// UnlockRequest has the information to request a xpdb to be unlocked.
type UnlockRequest struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// The leaseHolderIdentity of the lock that is going to be deleted on the cluster.
	LeaseHolderIdentity string `protobuf:"bytes,1,opt,name=lease_holder_identity,json=leaseHolderIdentity,proto3" json:"lease_holder_identity,omitempty"`
	// Namespace of the xpdb to be unlocked.
	Namespace string `protobuf:"bytes,2,opt,name=namespace,proto3" json:"namespace,omitempty"`
	// LabelSelector used on the xpdb resource to be locked.
	LabelSelector *LabelSelector `protobuf:"bytes,3,opt,name=label_selector,json=labelSelector,proto3" json:"label_selector,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UnlockRequest) Reset() {
	*x = UnlockRequest{}
	mi := &file_state_v1_state_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UnlockRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UnlockRequest) ProtoMessage() {}

func (x *UnlockRequest) ProtoReflect() protoreflect.Message {
	mi := &file_state_v1_state_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UnlockRequest.ProtoReflect.Descriptor instead.
func (*UnlockRequest) Descriptor() ([]byte, []int) {
	return file_state_v1_state_proto_rawDescGZIP(), []int{2}
}

func (x *UnlockRequest) GetLeaseHolderIdentity() string {
	if x != nil {
		return x.LeaseHolderIdentity
	}
	return ""
}

func (x *UnlockRequest) GetNamespace() string {
	if x != nil {
		return x.Namespace
	}
	return ""
}

func (x *UnlockRequest) GetLabelSelector() *LabelSelector {
	if x != nil {
		return x.LabelSelector
	}
	return nil
}

// LockResponse has the information on wether an unlock was succeeded or not.
type UnlockResponse struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Refers to wether the lock was unlocked or not.
	Unlocked bool `protobuf:"varint,1,opt,name=unlocked,proto3" json:"unlocked,omitempty"`
	// Error has the error occured during the unlock process.
	Error         string `protobuf:"bytes,2,opt,name=error,proto3" json:"error,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UnlockResponse) Reset() {
	*x = UnlockResponse{}
	mi := &file_state_v1_state_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UnlockResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UnlockResponse) ProtoMessage() {}

func (x *UnlockResponse) ProtoReflect() protoreflect.Message {
	mi := &file_state_v1_state_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UnlockResponse.ProtoReflect.Descriptor instead.
func (*UnlockResponse) Descriptor() ([]byte, []int) {
	return file_state_v1_state_proto_rawDescGZIP(), []int{3}
}

func (x *UnlockResponse) GetUnlocked() bool {
	if x != nil {
		return x.Unlocked
	}
	return false
}

func (x *UnlockResponse) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

// Gets the state of the pods captured by a xpdb label selector
type GetStateRequest struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Namespace of the xpdb to get.
	Namespace string `protobuf:"bytes,1,opt,name=namespace,proto3" json:"namespace,omitempty"`
	// LabelSelector of the xpdb to get the state from.
	LabelSelector *LabelSelector `protobuf:"bytes,2,opt,name=label_selector,json=labelSelector,proto3" json:"label_selector,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetStateRequest) Reset() {
	*x = GetStateRequest{}
	mi := &file_state_v1_state_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetStateRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetStateRequest) ProtoMessage() {}

func (x *GetStateRequest) ProtoReflect() protoreflect.Message {
	mi := &file_state_v1_state_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetStateRequest.ProtoReflect.Descriptor instead.
func (*GetStateRequest) Descriptor() ([]byte, []int) {
	return file_state_v1_state_proto_rawDescGZIP(), []int{4}
}

func (x *GetStateRequest) GetNamespace() string {
	if x != nil {
		return x.Namespace
	}
	return ""
}

func (x *GetStateRequest) GetLabelSelector() *LabelSelector {
	if x != nil {
		return x.LabelSelector
	}
	return nil
}

// Response of the GetState
type GetStateResponse struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Number of desired healthy pods captured by a xpdb label selector
	DesiredHealthy int32 `protobuf:"varint,1,opt,name=desired_healthy,json=desiredHealthy,proto3" json:"desired_healthy,omitempty"`
	// Number of healthy pods captured by a xpdb label selector
	Healthy       int32 `protobuf:"varint,2,opt,name=healthy,proto3" json:"healthy,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetStateResponse) Reset() {
	*x = GetStateResponse{}
	mi := &file_state_v1_state_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetStateResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetStateResponse) ProtoMessage() {}

func (x *GetStateResponse) ProtoReflect() protoreflect.Message {
	mi := &file_state_v1_state_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetStateResponse.ProtoReflect.Descriptor instead.
func (*GetStateResponse) Descriptor() ([]byte, []int) {
	return file_state_v1_state_proto_rawDescGZIP(), []int{5}
}

func (x *GetStateResponse) GetDesiredHealthy() int32 {
	if x != nil {
		return x.DesiredHealthy
	}
	return 0
}

func (x *GetStateResponse) GetHealthy() int32 {
	if x != nil {
		return x.Healthy
	}
	return 0
}

// A label selector is a label query over a set of resources. The result of matchLabels and
// matchExpressions are ANDed. An empty label selector matches all objects. A null
// label selector matches no objects.
// +structType=atomic
type LabelSelector struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels
	// map is equivalent to an element of matchExpressions, whose key field is "key", the
	// operator is "In", and the values array contains only "value". The requirements are ANDed.
	// +optional
	MatchLabels map[string]string `protobuf:"bytes,1,rep,name=match_labels,json=matchLabels,proto3" json:"match_labels,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	// matchExpressions is a list of label selector requirements. The requirements are ANDed.
	// +optional
	// +listType=atomic
	MatchExpressions []*LabelSelectorRequirement `protobuf:"bytes,2,rep,name=match_expressions,json=matchExpressions,proto3" json:"match_expressions,omitempty"`
	unknownFields    protoimpl.UnknownFields
	sizeCache        protoimpl.SizeCache
}

func (x *LabelSelector) Reset() {
	*x = LabelSelector{}
	mi := &file_state_v1_state_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *LabelSelector) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LabelSelector) ProtoMessage() {}

func (x *LabelSelector) ProtoReflect() protoreflect.Message {
	mi := &file_state_v1_state_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LabelSelector.ProtoReflect.Descriptor instead.
func (*LabelSelector) Descriptor() ([]byte, []int) {
	return file_state_v1_state_proto_rawDescGZIP(), []int{6}
}

func (x *LabelSelector) GetMatchLabels() map[string]string {
	if x != nil {
		return x.MatchLabels
	}
	return nil
}

func (x *LabelSelector) GetMatchExpressions() []*LabelSelectorRequirement {
	if x != nil {
		return x.MatchExpressions
	}
	return nil
}

// A label selector requirement is a selector that contains values, a key, and an operator that
// relates the key and values.
type LabelSelectorRequirement struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// key is the label key that the selector applies to.
	Key *string `protobuf:"bytes,1,opt,name=key,proto3,oneof" json:"key,omitempty"`
	// operator represents a key's relationship to a set of values.
	// Valid operators are In, NotIn, Exists and DoesNotExist.
	Operator *string `protobuf:"bytes,2,opt,name=operator,proto3,oneof" json:"operator,omitempty"`
	// values is an array of string values. If the operator is In or NotIn,
	// the values array must be non-empty. If the operator is Exists or DoesNotExist,
	// the values array must be empty. This array is replaced during a strategic
	// merge patch.
	// +optional
	// +listType=atomic
	Values        []string `protobuf:"bytes,3,rep,name=values,proto3" json:"values,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *LabelSelectorRequirement) Reset() {
	*x = LabelSelectorRequirement{}
	mi := &file_state_v1_state_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *LabelSelectorRequirement) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LabelSelectorRequirement) ProtoMessage() {}

func (x *LabelSelectorRequirement) ProtoReflect() protoreflect.Message {
	mi := &file_state_v1_state_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LabelSelectorRequirement.ProtoReflect.Descriptor instead.
func (*LabelSelectorRequirement) Descriptor() ([]byte, []int) {
	return file_state_v1_state_proto_rawDescGZIP(), []int{7}
}

func (x *LabelSelectorRequirement) GetKey() string {
	if x != nil && x.Key != nil {
		return *x.Key
	}
	return ""
}

func (x *LabelSelectorRequirement) GetOperator() string {
	if x != nil && x.Operator != nil {
		return *x.Operator
	}
	return ""
}

func (x *LabelSelectorRequirement) GetValues() []string {
	if x != nil {
		return x.Values
	}
	return nil
}

var File_state_v1_state_proto protoreflect.FileDescriptor

var file_state_v1_state_proto_rawDesc = []byte{
	0x0a, 0x14, 0x73, 0x74, 0x61, 0x74, 0x65, 0x2f, 0x76, 0x31, 0x2f, 0x73, 0x74, 0x61, 0x74, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x08, 0x73, 0x74, 0x61, 0x74, 0x65, 0x2e, 0x76, 0x31,
	0x22, 0x9f, 0x01, 0x0a, 0x0b, 0x4c, 0x6f, 0x63, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x32, 0x0a, 0x15, 0x6c, 0x65, 0x61, 0x73, 0x65, 0x5f, 0x68, 0x6f, 0x6c, 0x64, 0x65, 0x72,
	0x5f, 0x69, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x13, 0x6c, 0x65, 0x61, 0x73, 0x65, 0x48, 0x6f, 0x6c, 0x64, 0x65, 0x72, 0x49, 0x64, 0x65, 0x6e,
	0x74, 0x69, 0x74, 0x79, 0x12, 0x1c, 0x0a, 0x09, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61,
	0x63, 0x65, 0x12, 0x3e, 0x0a, 0x0e, 0x6c, 0x61, 0x62, 0x65, 0x6c, 0x5f, 0x73, 0x65, 0x6c, 0x65,
	0x63, 0x74, 0x6f, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x73, 0x74, 0x61,
	0x74, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x61, 0x62, 0x65, 0x6c, 0x53, 0x65, 0x6c, 0x65, 0x63,
	0x74, 0x6f, 0x72, 0x52, 0x0d, 0x6c, 0x61, 0x62, 0x65, 0x6c, 0x53, 0x65, 0x6c, 0x65, 0x63, 0x74,
	0x6f, 0x72, 0x22, 0x40, 0x0a, 0x0c, 0x4c, 0x6f, 0x63, 0x6b, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x61, 0x63, 0x71, 0x75, 0x69, 0x72, 0x65, 0x64, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x08, 0x61, 0x63, 0x71, 0x75, 0x69, 0x72, 0x65, 0x64, 0x12, 0x14,
	0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x65,
	0x72, 0x72, 0x6f, 0x72, 0x22, 0xa1, 0x01, 0x0a, 0x0d, 0x55, 0x6e, 0x6c, 0x6f, 0x63, 0x6b, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x32, 0x0a, 0x15, 0x6c, 0x65, 0x61, 0x73, 0x65, 0x5f,
	0x68, 0x6f, 0x6c, 0x64, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x13, 0x6c, 0x65, 0x61, 0x73, 0x65, 0x48, 0x6f, 0x6c, 0x64,
	0x65, 0x72, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x12, 0x1c, 0x0a, 0x09, 0x6e, 0x61,
	0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x6e,
	0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x12, 0x3e, 0x0a, 0x0e, 0x6c, 0x61, 0x62, 0x65,
	0x6c, 0x5f, 0x73, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x17, 0x2e, 0x73, 0x74, 0x61, 0x74, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x61, 0x62, 0x65,
	0x6c, 0x53, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x52, 0x0d, 0x6c, 0x61, 0x62, 0x65, 0x6c,
	0x53, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x22, 0x42, 0x0a, 0x0e, 0x55, 0x6e, 0x6c, 0x6f,
	0x63, 0x6b, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x75, 0x6e,
	0x6c, 0x6f, 0x63, 0x6b, 0x65, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x08, 0x75, 0x6e,
	0x6c, 0x6f, 0x63, 0x6b, 0x65, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x22, 0x6f, 0x0a, 0x0f,
	0x47, 0x65, 0x74, 0x53, 0x74, 0x61, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x1c, 0x0a, 0x09, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x09, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x12, 0x3e, 0x0a,
	0x0e, 0x6c, 0x61, 0x62, 0x65, 0x6c, 0x5f, 0x73, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x73, 0x74, 0x61, 0x74, 0x65, 0x2e, 0x76, 0x31,
	0x2e, 0x4c, 0x61, 0x62, 0x65, 0x6c, 0x53, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x52, 0x0d,
	0x6c, 0x61, 0x62, 0x65, 0x6c, 0x53, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x22, 0x55, 0x0a,
	0x10, 0x47, 0x65, 0x74, 0x53, 0x74, 0x61, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x27, 0x0a, 0x0f, 0x64, 0x65, 0x73, 0x69, 0x72, 0x65, 0x64, 0x5f, 0x68, 0x65, 0x61,
	0x6c, 0x74, 0x68, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0e, 0x64, 0x65, 0x73, 0x69,
	0x72, 0x65, 0x64, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x79, 0x12, 0x18, 0x0a, 0x07, 0x68, 0x65,
	0x61, 0x6c, 0x74, 0x68, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x07, 0x68, 0x65, 0x61,
	0x6c, 0x74, 0x68, 0x79, 0x22, 0xed, 0x01, 0x0a, 0x0d, 0x4c, 0x61, 0x62, 0x65, 0x6c, 0x53, 0x65,
	0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x12, 0x4b, 0x0a, 0x0c, 0x6d, 0x61, 0x74, 0x63, 0x68, 0x5f,
	0x6c, 0x61, 0x62, 0x65, 0x6c, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x28, 0x2e, 0x73,
	0x74, 0x61, 0x74, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x61, 0x62, 0x65, 0x6c, 0x53, 0x65, 0x6c,
	0x65, 0x63, 0x74, 0x6f, 0x72, 0x2e, 0x4d, 0x61, 0x74, 0x63, 0x68, 0x4c, 0x61, 0x62, 0x65, 0x6c,
	0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x0b, 0x6d, 0x61, 0x74, 0x63, 0x68, 0x4c, 0x61, 0x62,
	0x65, 0x6c, 0x73, 0x12, 0x4f, 0x0a, 0x11, 0x6d, 0x61, 0x74, 0x63, 0x68, 0x5f, 0x65, 0x78, 0x70,
	0x72, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x22,
	0x2e, 0x73, 0x74, 0x61, 0x74, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x61, 0x62, 0x65, 0x6c, 0x53,
	0x65, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x52, 0x65, 0x71, 0x75, 0x69, 0x72, 0x65, 0x6d, 0x65,
	0x6e, 0x74, 0x52, 0x10, 0x6d, 0x61, 0x74, 0x63, 0x68, 0x45, 0x78, 0x70, 0x72, 0x65, 0x73, 0x73,
	0x69, 0x6f, 0x6e, 0x73, 0x1a, 0x3e, 0x0a, 0x10, 0x4d, 0x61, 0x74, 0x63, 0x68, 0x4c, 0x61, 0x62,
	0x65, 0x6c, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x3a, 0x02, 0x38, 0x01, 0x22, 0x7f, 0x0a, 0x18, 0x4c, 0x61, 0x62, 0x65, 0x6c, 0x53, 0x65, 0x6c,
	0x65, 0x63, 0x74, 0x6f, 0x72, 0x52, 0x65, 0x71, 0x75, 0x69, 0x72, 0x65, 0x6d, 0x65, 0x6e, 0x74,
	0x12, 0x15, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52,
	0x03, 0x6b, 0x65, 0x79, 0x88, 0x01, 0x01, 0x12, 0x1f, 0x0a, 0x08, 0x6f, 0x70, 0x65, 0x72, 0x61,
	0x74, 0x6f, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x48, 0x01, 0x52, 0x08, 0x6f, 0x70, 0x65,
	0x72, 0x61, 0x74, 0x6f, 0x72, 0x88, 0x01, 0x01, 0x12, 0x16, 0x0a, 0x06, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x09, 0x52, 0x06, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x73,
	0x42, 0x06, 0x0a, 0x04, 0x5f, 0x6b, 0x65, 0x79, 0x42, 0x0b, 0x0a, 0x09, 0x5f, 0x6f, 0x70, 0x65,
	0x72, 0x61, 0x74, 0x6f, 0x72, 0x32, 0xcb, 0x01, 0x0a, 0x0c, 0x53, 0x74, 0x61, 0x74, 0x65, 0x53,
	0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x37, 0x0a, 0x04, 0x4c, 0x6f, 0x63, 0x6b, 0x12, 0x15,
	0x2e, 0x73, 0x74, 0x61, 0x74, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x6f, 0x63, 0x6b, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x16, 0x2e, 0x73, 0x74, 0x61, 0x74, 0x65, 0x2e, 0x76, 0x31,
	0x2e, 0x4c, 0x6f, 0x63, 0x6b, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12,
	0x3d, 0x0a, 0x06, 0x55, 0x6e, 0x6c, 0x6f, 0x63, 0x6b, 0x12, 0x17, 0x2e, 0x73, 0x74, 0x61, 0x74,
	0x65, 0x2e, 0x76, 0x31, 0x2e, 0x55, 0x6e, 0x6c, 0x6f, 0x63, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x18, 0x2e, 0x73, 0x74, 0x61, 0x74, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x55, 0x6e,
	0x6c, 0x6f, 0x63, 0x6b, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x43,
	0x0a, 0x08, 0x47, 0x65, 0x74, 0x53, 0x74, 0x61, 0x74, 0x65, 0x12, 0x19, 0x2e, 0x73, 0x74, 0x61,
	0x74, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x53, 0x74, 0x61, 0x74, 0x65, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1a, 0x2e, 0x73, 0x74, 0x61, 0x74, 0x65, 0x2e, 0x76, 0x31,
	0x2e, 0x47, 0x65, 0x74, 0x53, 0x74, 0x61, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x22, 0x00, 0x42, 0x8b, 0x01, 0x0a, 0x0c, 0x63, 0x6f, 0x6d, 0x2e, 0x73, 0x74, 0x61, 0x74,
	0x65, 0x2e, 0x76, 0x31, 0x42, 0x0a, 0x53, 0x74, 0x61, 0x74, 0x65, 0x50, 0x72, 0x6f, 0x74, 0x6f,
	0x50, 0x01, 0x5a, 0x2e, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x66,
	0x6f, 0x72, 0x6d, 0x33, 0x74, 0x65, 0x63, 0x68, 0x2d, 0x6f, 0x73, 0x73, 0x2f, 0x78, 0x2d, 0x70,
	0x64, 0x62, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x73, 0x74, 0x61,
	0x74, 0x65, 0xa2, 0x02, 0x03, 0x53, 0x58, 0x58, 0xaa, 0x02, 0x08, 0x53, 0x74, 0x61, 0x74, 0x65,
	0x2e, 0x56, 0x31, 0xca, 0x02, 0x08, 0x53, 0x74, 0x61, 0x74, 0x65, 0x5c, 0x56, 0x31, 0xe2, 0x02,
	0x14, 0x53, 0x74, 0x61, 0x74, 0x65, 0x5c, 0x56, 0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74,
	0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x09, 0x53, 0x74, 0x61, 0x74, 0x65, 0x3a, 0x3a, 0x56,
	0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_state_v1_state_proto_rawDescOnce sync.Once
	file_state_v1_state_proto_rawDescData = file_state_v1_state_proto_rawDesc
)

func file_state_v1_state_proto_rawDescGZIP() []byte {
	file_state_v1_state_proto_rawDescOnce.Do(func() {
		file_state_v1_state_proto_rawDescData = protoimpl.X.CompressGZIP(file_state_v1_state_proto_rawDescData)
	})
	return file_state_v1_state_proto_rawDescData
}

var file_state_v1_state_proto_msgTypes = make([]protoimpl.MessageInfo, 9)
var file_state_v1_state_proto_goTypes = []any{
	(*LockRequest)(nil),              // 0: state.v1.LockRequest
	(*LockResponse)(nil),             // 1: state.v1.LockResponse
	(*UnlockRequest)(nil),            // 2: state.v1.UnlockRequest
	(*UnlockResponse)(nil),           // 3: state.v1.UnlockResponse
	(*GetStateRequest)(nil),          // 4: state.v1.GetStateRequest
	(*GetStateResponse)(nil),         // 5: state.v1.GetStateResponse
	(*LabelSelector)(nil),            // 6: state.v1.LabelSelector
	(*LabelSelectorRequirement)(nil), // 7: state.v1.LabelSelectorRequirement
	nil,                              // 8: state.v1.LabelSelector.MatchLabelsEntry
}
var file_state_v1_state_proto_depIdxs = []int32{
	6, // 0: state.v1.LockRequest.label_selector:type_name -> state.v1.LabelSelector
	6, // 1: state.v1.UnlockRequest.label_selector:type_name -> state.v1.LabelSelector
	6, // 2: state.v1.GetStateRequest.label_selector:type_name -> state.v1.LabelSelector
	8, // 3: state.v1.LabelSelector.match_labels:type_name -> state.v1.LabelSelector.MatchLabelsEntry
	7, // 4: state.v1.LabelSelector.match_expressions:type_name -> state.v1.LabelSelectorRequirement
	0, // 5: state.v1.StateService.Lock:input_type -> state.v1.LockRequest
	2, // 6: state.v1.StateService.Unlock:input_type -> state.v1.UnlockRequest
	4, // 7: state.v1.StateService.GetState:input_type -> state.v1.GetStateRequest
	1, // 8: state.v1.StateService.Lock:output_type -> state.v1.LockResponse
	3, // 9: state.v1.StateService.Unlock:output_type -> state.v1.UnlockResponse
	5, // 10: state.v1.StateService.GetState:output_type -> state.v1.GetStateResponse
	8, // [8:11] is the sub-list for method output_type
	5, // [5:8] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_state_v1_state_proto_init() }
func file_state_v1_state_proto_init() {
	if File_state_v1_state_proto != nil {
		return
	}
	file_state_v1_state_proto_msgTypes[7].OneofWrappers = []any{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_state_v1_state_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   9,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_state_v1_state_proto_goTypes,
		DependencyIndexes: file_state_v1_state_proto_depIdxs,
		MessageInfos:      file_state_v1_state_proto_msgTypes,
	}.Build()
	File_state_v1_state_proto = out.File
	file_state_v1_state_proto_rawDesc = nil
	file_state_v1_state_proto_goTypes = nil
	file_state_v1_state_proto_depIdxs = nil
}
