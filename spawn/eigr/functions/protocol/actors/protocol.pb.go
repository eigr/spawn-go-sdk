// The Spawn Protocol
//
// Spawn is divided into two main parts namely:
//
//   1. A sidecar proxy that exposes the server part of the Spawn Protocol in
//   the form of an HTTP API.
//   2. A user function, written in any language that supports HTTP, that
//   exposes the client part of the Spawn Protocol.
//
// Both are client and server of their counterparts.
//
// In turn, the proxy exposes an HTTP endpoint for registering a user function
// a.k.a ActorSystem.
//
// A user function that wants to register actors in Proxy Spawn must proceed by
// making a POST request to the following endpoint:
//
// `
// POST /api/v1/system HTTP 1.1
// HOST: localhost
// User-Agent: user-function-client/0.1.0 (this is just example)
// Accept: application/octet-stream
// Content-Type: application/octet-stream
//
// registration request type bytes encoded here :-)
// `
//
// The general flow of a registration action is as follows:
//
// ╔═══════════════════╗                  ╔═══════════════════╗
// ╔═══════════════════╗ ║   User Function   ║                  ║Local Spawn
// Sidecar║                  ║       Actor       ║ ╚═══════════════════╝
// ╚═══════════════════╝                  ╚═══════════════════╝
//          ║                                      ║ ║ ║ ║ ║ ║              HTTP
//          POST               ║                                      ║ ║
//          Registration              ║                                      ║
//          ║               Request                ║ ║
//          ╠─────────────────────────────────────▶║ ║ ║ ║       Upfront start
//          Actors with      ║ ║ ╠───────BEAM Distributed Protocol─────▶║ ║ ║ ║
//          ║                                      ║ ╠───┐Initialize ║ ║ ║   │
//          State ║                                      ║ ║   │  Store ║ ║
//          ║◀──┘ ║           HTTP Registration          ║ ║ ║ Response ║ ║
//          ║◀─────────────────────────────────────╣ ║ ║ ║ ║ ║ ║ ║ ║ ║ ║ ║ ║ ║
//          ║                                      ║ ║ ║ ║ ║ ║ ║ ║ ║ ║ ║
//
//     ███████████                            ███████████ ███████████
//
//
// ## Spawning Actors
//
// Actors are usually created at the beginning of the SDK's communication flow
// with the Proxy by the registration step described above. However, some use
// cases require that Actors can be created ***on the fly***. In other words,
// Spawn is used to bring to life Actors previously registered as Unnameds,
// giving them a name and thus creating a concrete instance at runtime for that
// Actor. Actors created with the Spawn feature are generally used when you want
// to share a behavior while maintaining the isolation characteristics of the
// actors. For these situations we have the Spawning flow described below.
//
// A user function that wants to Spawning new Actors in Proxy Spawn must proceed
// by making a POST request to the following endpoint:
//
// ```
// POST /system/:system_name/actors/spawn HTTP 1.1
// HOST: localhost
// User-Agent: user-function-client/0.1.0 (this is just example)
// Accept: application/octet-stream
// Content-Type: application/octet-stream
//
// SpawnRequest type bytes encoded here :-)
// ```
//
// The general flow of a Spawning Actors is as follows:
//
// ```
// +----------------+ +---------------------+ +-------+ | User Function  | |
// Local Spawn Sidecar |                                     | Actor |
// +----------------+ +---------------------+ +-------+
//         |                                                       | | | HTTP
//         POST SpawnRequest                                | |
//         |------------------------------------------------------>| | | | | |
//         | Upfront start Actors with BEAM Distributed Protocol | |
//         |---------------------------------------------------->| | | | | |
//         |Initialize Statestore | | |---------------------- | | | | | |
//         |<--------------------- | | | |          HTTP SpawnResponse | |
//         |<------------------------------------------------------| | | | |
// ```
//
// Once the system has been initialized, that is, the registration step has been
// successfully completed, then the user function will be able to make requests
// to the System Actors. This is done through a post request to the Proxy at the
// `/system/:name/actors/:actor_name/invoke` endpoint.
//
// A user function that wants to call actors in Proxy Spawn must proceed by
// making a POST request as the follow:
//
// `
// POST /system/:name/actors/:actor_name/invoke HTTP 1.1
// HOST: localhost
// User-Agent: user-function-client/0.1.0 (this is just example)
// Accept: application/octet-stream
// Content-Type: application/octet-stream
//
// invocation request type bytes encoded here :-)
// `
//
// Assuming that two user functions were registered in different separate
// Proxies, the above request would go the following way:
//
// ╔═══════════════════╗                  ╔═══════════════════╗
// ╔═════════════════════════╗        ╔═════════════════════════════╗ ║   User
// Function   ║                  ║Local Spawn Sidecar║              ║ Remote
// User Function B  ║        ║Remote Spawn Sidecar/Actor B ║
// ╚═══════════════════╝                  ╚═══════════════════╝
// ╚═════════════════════════╝        ╚═════════════════════════════╝
//          ║              HTTP POST               ║ ║ ║ ║ Registration ║ ║ ║ ║
//          Request                ║                                     ║ ║
//          ╠─────────────────────────────────────▶║ ║ ║ ║ ╠───┐ ║ ║ ║ ║ │Lookup
//          for                       ║                                    ║ ║
//          ║   │  Actor                          ║ ║ ║ ║◀──┘ ║ ║ ║ ║ ║ BEAM
//          Distributed         ║ ║
//          ╠─────────────────────────────────────╬────────────protocol
//          call──────────▶║ ║                                      ║ ║ ║ ║ ║ ║
//          HTTP POST:             ║ ║                                      ║
//          ║◀──────/api/v1/actors/actions───────╣ ║ ║ ║ ║ ║ ║ ╠───┐ ║ ║ ║ ║
//          │Handle request,                 ║ ║ ║ ║   │execute action ║ ║ ║
//          ║◀──┘                                ║ ║ ║ ║            Reply with
//          the          ║ ║                                      ║
//          ╠────────────result and the ────────▶║ ║ ║ ║             new state
//          of           ║────┐ ║                                      ║ ║ ║ │
//          ║                                      ║ ║ ║    │Store new State ║
//          ║       Send response to the          ║ ║ ◀──┘ ║         Respond to
//          user with         ║◀─────────Spawn Sidecar
//          A────────────╬────────────────────────────────────╣ ║ result value
//          ║                                     ║ ║
//          ║◀─────────────────────────────────────╣ ║ ║ ║ ║ ║ ║ ║ ║ ║ ║
//
//     ███████████                           ████████████ ███████████
//     ███████████
//
//

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.2
// 	protoc        v3.12.4
// source: eigr/functions/protocol/actors/protocol.proto

package actors

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	anypb "google.golang.org/protobuf/types/known/anypb"
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

type Status int32

const (
	Status_UNKNOWN         Status = 0
	Status_OK              Status = 1
	Status_ACTOR_NOT_FOUND Status = 2
	Status_ERROR           Status = 3
)

// Enum value maps for Status.
var (
	Status_name = map[int32]string{
		0: "UNKNOWN",
		1: "OK",
		2: "ACTOR_NOT_FOUND",
		3: "ERROR",
	}
	Status_value = map[string]int32{
		"UNKNOWN":         0,
		"OK":              1,
		"ACTOR_NOT_FOUND": 2,
		"ERROR":           3,
	}
)

func (x Status) Enum() *Status {
	p := new(Status)
	*p = x
	return p
}

func (x Status) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Status) Descriptor() protoreflect.EnumDescriptor {
	return file_eigr_functions_protocol_actors_protocol_proto_enumTypes[0].Descriptor()
}

func (Status) Type() protoreflect.EnumType {
	return &file_eigr_functions_protocol_actors_protocol_proto_enumTypes[0]
}

func (x Status) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Status.Descriptor instead.
func (Status) EnumDescriptor() ([]byte, []int) {
	return file_eigr_functions_protocol_actors_protocol_proto_rawDescGZIP(), []int{0}
}

// Context is where current and/or updated state is stored
// to be transmitted to/from proxy and user function
//
// Params:
//   - state: Actor state passed back and forth between proxy and user function.
//   - metadata: Meta information that comes in invocations
//   - tags: Meta information stored in the actor
//   - caller: ActorId of who is calling target actor
//   - self: ActorId of itself
type Context struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	State    *anypb.Any        `protobuf:"bytes,1,opt,name=state,proto3" json:"state,omitempty"`
	Metadata map[string]string `protobuf:"bytes,4,rep,name=metadata,proto3" json:"metadata,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Tags     map[string]string `protobuf:"bytes,5,rep,name=tags,proto3" json:"tags,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// Who is calling target actor
	Caller *ActorId `protobuf:"bytes,2,opt,name=caller,proto3" json:"caller,omitempty"`
	// The target actor itself
	Self *ActorId `protobuf:"bytes,3,opt,name=self,proto3" json:"self,omitempty"`
}

func (x *Context) Reset() {
	*x = Context{}
	mi := &file_eigr_functions_protocol_actors_protocol_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Context) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Context) ProtoMessage() {}

func (x *Context) ProtoReflect() protoreflect.Message {
	mi := &file_eigr_functions_protocol_actors_protocol_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Context.ProtoReflect.Descriptor instead.
func (*Context) Descriptor() ([]byte, []int) {
	return file_eigr_functions_protocol_actors_protocol_proto_rawDescGZIP(), []int{0}
}

func (x *Context) GetState() *anypb.Any {
	if x != nil {
		return x.State
	}
	return nil
}

func (x *Context) GetMetadata() map[string]string {
	if x != nil {
		return x.Metadata
	}
	return nil
}

func (x *Context) GetTags() map[string]string {
	if x != nil {
		return x.Tags
	}
	return nil
}

func (x *Context) GetCaller() *ActorId {
	if x != nil {
		return x.Caller
	}
	return nil
}

func (x *Context) GetSelf() *ActorId {
	if x != nil {
		return x.Self
	}
	return nil
}

// Noop is used when the input or output value of a function or method
// does not matter to the caller of a Workflow or when the user just wants to
// receive the Context in the request, that is, he does not care about the input
// value only with the state.
type Noop struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *Noop) Reset() {
	*x = Noop{}
	mi := &file_eigr_functions_protocol_actors_protocol_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Noop) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Noop) ProtoMessage() {}

func (x *Noop) ProtoReflect() protoreflect.Message {
	mi := &file_eigr_functions_protocol_actors_protocol_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Noop.ProtoReflect.Descriptor instead.
func (*Noop) Descriptor() ([]byte, []int) {
	return file_eigr_functions_protocol_actors_protocol_proto_rawDescGZIP(), []int{1}
}

// JSON is an alternative that some SDKs can opt in
// it will bypass any type validation in spawn actors state / payloads
type JSONType struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Content string `protobuf:"bytes,1,opt,name=content,proto3" json:"content,omitempty"`
}

func (x *JSONType) Reset() {
	*x = JSONType{}
	mi := &file_eigr_functions_protocol_actors_protocol_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *JSONType) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*JSONType) ProtoMessage() {}

func (x *JSONType) ProtoReflect() protoreflect.Message {
	mi := &file_eigr_functions_protocol_actors_protocol_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use JSONType.ProtoReflect.Descriptor instead.
func (*JSONType) Descriptor() ([]byte, []int) {
	return file_eigr_functions_protocol_actors_protocol_proto_rawDescGZIP(), []int{2}
}

func (x *JSONType) GetContent() string {
	if x != nil {
		return x.Content
	}
	return ""
}

type RegistrationRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ServiceInfo *ServiceInfo `protobuf:"bytes,1,opt,name=service_info,json=serviceInfo,proto3" json:"service_info,omitempty"`
	ActorSystem *ActorSystem `protobuf:"bytes,2,opt,name=actor_system,json=actorSystem,proto3" json:"actor_system,omitempty"`
}

func (x *RegistrationRequest) Reset() {
	*x = RegistrationRequest{}
	mi := &file_eigr_functions_protocol_actors_protocol_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RegistrationRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RegistrationRequest) ProtoMessage() {}

func (x *RegistrationRequest) ProtoReflect() protoreflect.Message {
	mi := &file_eigr_functions_protocol_actors_protocol_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RegistrationRequest.ProtoReflect.Descriptor instead.
func (*RegistrationRequest) Descriptor() ([]byte, []int) {
	return file_eigr_functions_protocol_actors_protocol_proto_rawDescGZIP(), []int{3}
}

func (x *RegistrationRequest) GetServiceInfo() *ServiceInfo {
	if x != nil {
		return x.ServiceInfo
	}
	return nil
}

func (x *RegistrationRequest) GetActorSystem() *ActorSystem {
	if x != nil {
		return x.ActorSystem
	}
	return nil
}

type RegistrationResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Status    *RequestStatus `protobuf:"bytes,1,opt,name=status,proto3" json:"status,omitempty"`
	ProxyInfo *ProxyInfo     `protobuf:"bytes,2,opt,name=proxy_info,json=proxyInfo,proto3" json:"proxy_info,omitempty"`
}

func (x *RegistrationResponse) Reset() {
	*x = RegistrationResponse{}
	mi := &file_eigr_functions_protocol_actors_protocol_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RegistrationResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RegistrationResponse) ProtoMessage() {}

func (x *RegistrationResponse) ProtoReflect() protoreflect.Message {
	mi := &file_eigr_functions_protocol_actors_protocol_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RegistrationResponse.ProtoReflect.Descriptor instead.
func (*RegistrationResponse) Descriptor() ([]byte, []int) {
	return file_eigr_functions_protocol_actors_protocol_proto_rawDescGZIP(), []int{4}
}

func (x *RegistrationResponse) GetStatus() *RequestStatus {
	if x != nil {
		return x.Status
	}
	return nil
}

func (x *RegistrationResponse) GetProxyInfo() *ProxyInfo {
	if x != nil {
		return x.ProxyInfo
	}
	return nil
}

type ServiceInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The name of the actor system, eg, "my-actor-system".
	ServiceName string `protobuf:"bytes,1,opt,name=service_name,json=serviceName,proto3" json:"service_name,omitempty"`
	// The version of the service.
	ServiceVersion string `protobuf:"bytes,2,opt,name=service_version,json=serviceVersion,proto3" json:"service_version,omitempty"`
	// A description of the runtime for the service. Can be anything, but examples
	// might be:
	// - node v10.15.2
	// - OpenJDK Runtime Environment 1.8.0_192-b12
	ServiceRuntime string `protobuf:"bytes,3,opt,name=service_runtime,json=serviceRuntime,proto3" json:"service_runtime,omitempty"`
	// If using a support library, the name of that library, eg "spawn-jvm"
	SupportLibraryName string `protobuf:"bytes,4,opt,name=support_library_name,json=supportLibraryName,proto3" json:"support_library_name,omitempty"`
	// The version of the support library being used.
	SupportLibraryVersion string `protobuf:"bytes,5,opt,name=support_library_version,json=supportLibraryVersion,proto3" json:"support_library_version,omitempty"`
	// Spawn protocol major version accepted by the support library.
	ProtocolMajorVersion int32 `protobuf:"varint,6,opt,name=protocol_major_version,json=protocolMajorVersion,proto3" json:"protocol_major_version,omitempty"`
	// Spawn protocol minor version accepted by the support library.
	ProtocolMinorVersion int32 `protobuf:"varint,7,opt,name=protocol_minor_version,json=protocolMinorVersion,proto3" json:"protocol_minor_version,omitempty"`
}

func (x *ServiceInfo) Reset() {
	*x = ServiceInfo{}
	mi := &file_eigr_functions_protocol_actors_protocol_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ServiceInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServiceInfo) ProtoMessage() {}

func (x *ServiceInfo) ProtoReflect() protoreflect.Message {
	mi := &file_eigr_functions_protocol_actors_protocol_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ServiceInfo.ProtoReflect.Descriptor instead.
func (*ServiceInfo) Descriptor() ([]byte, []int) {
	return file_eigr_functions_protocol_actors_protocol_proto_rawDescGZIP(), []int{5}
}

func (x *ServiceInfo) GetServiceName() string {
	if x != nil {
		return x.ServiceName
	}
	return ""
}

func (x *ServiceInfo) GetServiceVersion() string {
	if x != nil {
		return x.ServiceVersion
	}
	return ""
}

func (x *ServiceInfo) GetServiceRuntime() string {
	if x != nil {
		return x.ServiceRuntime
	}
	return ""
}

func (x *ServiceInfo) GetSupportLibraryName() string {
	if x != nil {
		return x.SupportLibraryName
	}
	return ""
}

func (x *ServiceInfo) GetSupportLibraryVersion() string {
	if x != nil {
		return x.SupportLibraryVersion
	}
	return ""
}

func (x *ServiceInfo) GetProtocolMajorVersion() int32 {
	if x != nil {
		return x.ProtocolMajorVersion
	}
	return 0
}

func (x *ServiceInfo) GetProtocolMinorVersion() int32 {
	if x != nil {
		return x.ProtocolMinorVersion
	}
	return 0
}

type SpawnRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Actors []*ActorId `protobuf:"bytes,1,rep,name=actors,proto3" json:"actors,omitempty"`
}

func (x *SpawnRequest) Reset() {
	*x = SpawnRequest{}
	mi := &file_eigr_functions_protocol_actors_protocol_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SpawnRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SpawnRequest) ProtoMessage() {}

func (x *SpawnRequest) ProtoReflect() protoreflect.Message {
	mi := &file_eigr_functions_protocol_actors_protocol_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SpawnRequest.ProtoReflect.Descriptor instead.
func (*SpawnRequest) Descriptor() ([]byte, []int) {
	return file_eigr_functions_protocol_actors_protocol_proto_rawDescGZIP(), []int{6}
}

func (x *SpawnRequest) GetActors() []*ActorId {
	if x != nil {
		return x.Actors
	}
	return nil
}

type SpawnResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Status *RequestStatus `protobuf:"bytes,1,opt,name=status,proto3" json:"status,omitempty"`
}

func (x *SpawnResponse) Reset() {
	*x = SpawnResponse{}
	mi := &file_eigr_functions_protocol_actors_protocol_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SpawnResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SpawnResponse) ProtoMessage() {}

func (x *SpawnResponse) ProtoReflect() protoreflect.Message {
	mi := &file_eigr_functions_protocol_actors_protocol_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SpawnResponse.ProtoReflect.Descriptor instead.
func (*SpawnResponse) Descriptor() ([]byte, []int) {
	return file_eigr_functions_protocol_actors_protocol_proto_rawDescGZIP(), []int{7}
}

func (x *SpawnResponse) GetStatus() *RequestStatus {
	if x != nil {
		return x.Status
	}
	return nil
}

type ProxyInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ProtocolMajorVersion int32  `protobuf:"varint,1,opt,name=protocol_major_version,json=protocolMajorVersion,proto3" json:"protocol_major_version,omitempty"`
	ProtocolMinorVersion int32  `protobuf:"varint,2,opt,name=protocol_minor_version,json=protocolMinorVersion,proto3" json:"protocol_minor_version,omitempty"`
	ProxyName            string `protobuf:"bytes,3,opt,name=proxy_name,json=proxyName,proto3" json:"proxy_name,omitempty"`
	ProxyVersion         string `protobuf:"bytes,4,opt,name=proxy_version,json=proxyVersion,proto3" json:"proxy_version,omitempty"`
}

func (x *ProxyInfo) Reset() {
	*x = ProxyInfo{}
	mi := &file_eigr_functions_protocol_actors_protocol_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ProxyInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProxyInfo) ProtoMessage() {}

func (x *ProxyInfo) ProtoReflect() protoreflect.Message {
	mi := &file_eigr_functions_protocol_actors_protocol_proto_msgTypes[8]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProxyInfo.ProtoReflect.Descriptor instead.
func (*ProxyInfo) Descriptor() ([]byte, []int) {
	return file_eigr_functions_protocol_actors_protocol_proto_rawDescGZIP(), []int{8}
}

func (x *ProxyInfo) GetProtocolMajorVersion() int32 {
	if x != nil {
		return x.ProtocolMajorVersion
	}
	return 0
}

func (x *ProxyInfo) GetProtocolMinorVersion() int32 {
	if x != nil {
		return x.ProtocolMinorVersion
	}
	return 0
}

func (x *ProxyInfo) GetProxyName() string {
	if x != nil {
		return x.ProxyName
	}
	return ""
}

func (x *ProxyInfo) GetProxyVersion() string {
	if x != nil {
		return x.ProxyVersion
	}
	return ""
}

// When a Host Function is invoked it returns the updated state and return value
// to the call. It can also return a number of side effects to other Actors as a
// result of its computation. These side effects will be forwarded to the
// respective Actors asynchronously and should not affect the Host Function's
// response to its caller. Internally side effects is just a special kind of
// InvocationRequest. Useful for handle handle `recipient list` and `Composed
// Message Processor` patterns:
// https://www.enterpriseintegrationpatterns.com/patterns/messaging/RecipientList.html
// https://www.enterpriseintegrationpatterns.com/patterns/messaging/DistributionAggregate.html
type SideEffect struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Request *InvocationRequest `protobuf:"bytes,1,opt,name=request,proto3" json:"request,omitempty"`
}

func (x *SideEffect) Reset() {
	*x = SideEffect{}
	mi := &file_eigr_functions_protocol_actors_protocol_proto_msgTypes[9]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SideEffect) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SideEffect) ProtoMessage() {}

func (x *SideEffect) ProtoReflect() protoreflect.Message {
	mi := &file_eigr_functions_protocol_actors_protocol_proto_msgTypes[9]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SideEffect.ProtoReflect.Descriptor instead.
func (*SideEffect) Descriptor() ([]byte, []int) {
	return file_eigr_functions_protocol_actors_protocol_proto_rawDescGZIP(), []int{9}
}

func (x *SideEffect) GetRequest() *InvocationRequest {
	if x != nil {
		return x.Request
	}
	return nil
}

// Broadcast a message to many Actors
// Useful for handle `recipient list`, `publish-subscribe channel`, and
// `scatter-gatther` patterns:
// https://www.enterpriseintegrationpatterns.com/patterns/messaging/RecipientList.html
// https://www.enterpriseintegrationpatterns.com/patterns/messaging/PublishSubscribeChannel.html
// https://www.enterpriseintegrationpatterns.com/patterns/messaging/BroadcastAggregate.html
type Broadcast struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Target topic or channel
	// Change this to channel
	ChannelGroup string `protobuf:"bytes,1,opt,name=channel_group,json=channelGroup,proto3" json:"channel_group,omitempty"`
	// Payload
	//
	// Types that are assignable to Payload:
	//
	//	*Broadcast_Value
	//	*Broadcast_Noop
	Payload isBroadcast_Payload `protobuf_oneof:"payload"`
}

func (x *Broadcast) Reset() {
	*x = Broadcast{}
	mi := &file_eigr_functions_protocol_actors_protocol_proto_msgTypes[10]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Broadcast) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Broadcast) ProtoMessage() {}

func (x *Broadcast) ProtoReflect() protoreflect.Message {
	mi := &file_eigr_functions_protocol_actors_protocol_proto_msgTypes[10]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Broadcast.ProtoReflect.Descriptor instead.
func (*Broadcast) Descriptor() ([]byte, []int) {
	return file_eigr_functions_protocol_actors_protocol_proto_rawDescGZIP(), []int{10}
}

func (x *Broadcast) GetChannelGroup() string {
	if x != nil {
		return x.ChannelGroup
	}
	return ""
}

func (m *Broadcast) GetPayload() isBroadcast_Payload {
	if m != nil {
		return m.Payload
	}
	return nil
}

func (x *Broadcast) GetValue() *anypb.Any {
	if x, ok := x.GetPayload().(*Broadcast_Value); ok {
		return x.Value
	}
	return nil
}

func (x *Broadcast) GetNoop() *Noop {
	if x, ok := x.GetPayload().(*Broadcast_Noop); ok {
		return x.Noop
	}
	return nil
}

type isBroadcast_Payload interface {
	isBroadcast_Payload()
}

type Broadcast_Value struct {
	Value *anypb.Any `protobuf:"bytes,3,opt,name=value,proto3,oneof"`
}

type Broadcast_Noop struct {
	Noop *Noop `protobuf:"bytes,4,opt,name=noop,proto3,oneof"`
}

func (*Broadcast_Value) isBroadcast_Payload() {}

func (*Broadcast_Noop) isBroadcast_Payload() {}

// Sends the output of a action of an Actor to the input of another action of an
// Actor Useful for handle `pipes` pattern:
// https://www.enterpriseintegrationpatterns.com/patterns/messaging/PipesAndFilters.html
type Pipe struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Target Actor
	Actor string `protobuf:"bytes,1,opt,name=actor,proto3" json:"actor,omitempty"`
	// Action.
	ActionName string `protobuf:"bytes,2,opt,name=action_name,json=actionName,proto3" json:"action_name,omitempty"`
}

func (x *Pipe) Reset() {
	*x = Pipe{}
	mi := &file_eigr_functions_protocol_actors_protocol_proto_msgTypes[11]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Pipe) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Pipe) ProtoMessage() {}

func (x *Pipe) ProtoReflect() protoreflect.Message {
	mi := &file_eigr_functions_protocol_actors_protocol_proto_msgTypes[11]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Pipe.ProtoReflect.Descriptor instead.
func (*Pipe) Descriptor() ([]byte, []int) {
	return file_eigr_functions_protocol_actors_protocol_proto_rawDescGZIP(), []int{11}
}

func (x *Pipe) GetActor() string {
	if x != nil {
		return x.Actor
	}
	return ""
}

func (x *Pipe) GetActionName() string {
	if x != nil {
		return x.ActionName
	}
	return ""
}

// Sends the input of a action of an Actor to the input of another action of an
// Actor Useful for handle `content-basead router` pattern
// https://www.enterpriseintegrationpatterns.com/patterns/messaging/ContentBasedRouter.html
type Forward struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Target Actor
	Actor string `protobuf:"bytes,1,opt,name=actor,proto3" json:"actor,omitempty"`
	// Action.
	ActionName string `protobuf:"bytes,2,opt,name=action_name,json=actionName,proto3" json:"action_name,omitempty"`
}

func (x *Forward) Reset() {
	*x = Forward{}
	mi := &file_eigr_functions_protocol_actors_protocol_proto_msgTypes[12]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Forward) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Forward) ProtoMessage() {}

func (x *Forward) ProtoReflect() protoreflect.Message {
	mi := &file_eigr_functions_protocol_actors_protocol_proto_msgTypes[12]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Forward.ProtoReflect.Descriptor instead.
func (*Forward) Descriptor() ([]byte, []int) {
	return file_eigr_functions_protocol_actors_protocol_proto_rawDescGZIP(), []int{12}
}

func (x *Forward) GetActor() string {
	if x != nil {
		return x.Actor
	}
	return ""
}

func (x *Forward) GetActionName() string {
	if x != nil {
		return x.ActionName
	}
	return ""
}

// Facts are emitted by actions and represent the internal state of the moment
// at that moment. These are treated by Projections so that visualizations can
// be built around these states.
type Fact struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Uuid      string                 `protobuf:"bytes,1,opt,name=uuid,proto3" json:"uuid,omitempty"`
	State     *anypb.Any             `protobuf:"bytes,2,opt,name=state,proto3" json:"state,omitempty"`
	Metadata  map[string]string      `protobuf:"bytes,3,rep,name=metadata,proto3" json:"metadata,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Timestamp *timestamppb.Timestamp `protobuf:"bytes,4,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
}

func (x *Fact) Reset() {
	*x = Fact{}
	mi := &file_eigr_functions_protocol_actors_protocol_proto_msgTypes[13]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Fact) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Fact) ProtoMessage() {}

func (x *Fact) ProtoReflect() protoreflect.Message {
	mi := &file_eigr_functions_protocol_actors_protocol_proto_msgTypes[13]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Fact.ProtoReflect.Descriptor instead.
func (*Fact) Descriptor() ([]byte, []int) {
	return file_eigr_functions_protocol_actors_protocol_proto_rawDescGZIP(), []int{13}
}

func (x *Fact) GetUuid() string {
	if x != nil {
		return x.Uuid
	}
	return ""
}

func (x *Fact) GetState() *anypb.Any {
	if x != nil {
		return x.State
	}
	return nil
}

func (x *Fact) GetMetadata() map[string]string {
	if x != nil {
		return x.Metadata
	}
	return nil
}

func (x *Fact) GetTimestamp() *timestamppb.Timestamp {
	if x != nil {
		return x.Timestamp
	}
	return nil
}

// Container for archicetural message patterns
type Workflow struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Broadcast *Broadcast    `protobuf:"bytes,2,opt,name=broadcast,proto3" json:"broadcast,omitempty"`
	Effects   []*SideEffect `protobuf:"bytes,1,rep,name=effects,proto3" json:"effects,omitempty"`
	// Types that are assignable to Routing:
	//
	//	*Workflow_Pipe
	//	*Workflow_Forward
	Routing isWorkflow_Routing `protobuf_oneof:"routing"`
}

func (x *Workflow) Reset() {
	*x = Workflow{}
	mi := &file_eigr_functions_protocol_actors_protocol_proto_msgTypes[14]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Workflow) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Workflow) ProtoMessage() {}

func (x *Workflow) ProtoReflect() protoreflect.Message {
	mi := &file_eigr_functions_protocol_actors_protocol_proto_msgTypes[14]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Workflow.ProtoReflect.Descriptor instead.
func (*Workflow) Descriptor() ([]byte, []int) {
	return file_eigr_functions_protocol_actors_protocol_proto_rawDescGZIP(), []int{14}
}

func (x *Workflow) GetBroadcast() *Broadcast {
	if x != nil {
		return x.Broadcast
	}
	return nil
}

func (x *Workflow) GetEffects() []*SideEffect {
	if x != nil {
		return x.Effects
	}
	return nil
}

func (m *Workflow) GetRouting() isWorkflow_Routing {
	if m != nil {
		return m.Routing
	}
	return nil
}

func (x *Workflow) GetPipe() *Pipe {
	if x, ok := x.GetRouting().(*Workflow_Pipe); ok {
		return x.Pipe
	}
	return nil
}

func (x *Workflow) GetForward() *Forward {
	if x, ok := x.GetRouting().(*Workflow_Forward); ok {
		return x.Forward
	}
	return nil
}

type isWorkflow_Routing interface {
	isWorkflow_Routing()
}

type Workflow_Pipe struct {
	Pipe *Pipe `protobuf:"bytes,3,opt,name=pipe,proto3,oneof"`
}

type Workflow_Forward struct {
	Forward *Forward `protobuf:"bytes,4,opt,name=forward,proto3,oneof"`
}

func (*Workflow_Pipe) isWorkflow_Routing() {}

func (*Workflow_Forward) isWorkflow_Routing() {}

// The user function when it wants to send a message to an Actor uses the
// InvocationRequest message type.
//
// Params:
//   - system: See ActorSystem message.
//   - actor: The target Actor, i.e. the one that the user function is calling
//     to perform some computation.
//   - caller: The caller Actor
//   - action_name: The function or method on the target Actor that will receive
//     this request
//     and perform some useful computation with the sent data.
//   - value: This is the value sent by the user function to be computed by the
//     request's target Actor action.
//   - async: Indicates whether the action should be processed synchronously,
//     where a response should be sent back to the user function,
//     or whether the action should be processed asynchronously, i.e. no
//     response sent to the caller and no waiting.
//   - metadata: Meta information or headers
//   - register_ref: If the invocation should register the specific actor with
//     the given name without having to call register before
type InvocationRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	System     *ActorSystem `protobuf:"bytes,1,opt,name=system,proto3" json:"system,omitempty"`
	Actor      *Actor       `protobuf:"bytes,2,opt,name=actor,proto3" json:"actor,omitempty"`
	ActionName string       `protobuf:"bytes,3,opt,name=action_name,json=actionName,proto3" json:"action_name,omitempty"`
	// Types that are assignable to Payload:
	//
	//	*InvocationRequest_Value
	//	*InvocationRequest_Noop
	Payload     isInvocationRequest_Payload `protobuf_oneof:"payload"`
	Async       bool                        `protobuf:"varint,5,opt,name=async,proto3" json:"async,omitempty"`
	Caller      *ActorId                    `protobuf:"bytes,6,opt,name=caller,proto3" json:"caller,omitempty"`
	Metadata    map[string]string           `protobuf:"bytes,8,rep,name=metadata,proto3" json:"metadata,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	ScheduledTo int64                       `protobuf:"varint,9,opt,name=scheduled_to,json=scheduledTo,proto3" json:"scheduled_to,omitempty"`
	Pooled      bool                        `protobuf:"varint,10,opt,name=pooled,proto3" json:"pooled,omitempty"`
	RegisterRef string                      `protobuf:"bytes,11,opt,name=register_ref,json=registerRef,proto3" json:"register_ref,omitempty"`
}

func (x *InvocationRequest) Reset() {
	*x = InvocationRequest{}
	mi := &file_eigr_functions_protocol_actors_protocol_proto_msgTypes[15]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *InvocationRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InvocationRequest) ProtoMessage() {}

func (x *InvocationRequest) ProtoReflect() protoreflect.Message {
	mi := &file_eigr_functions_protocol_actors_protocol_proto_msgTypes[15]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InvocationRequest.ProtoReflect.Descriptor instead.
func (*InvocationRequest) Descriptor() ([]byte, []int) {
	return file_eigr_functions_protocol_actors_protocol_proto_rawDescGZIP(), []int{15}
}

func (x *InvocationRequest) GetSystem() *ActorSystem {
	if x != nil {
		return x.System
	}
	return nil
}

func (x *InvocationRequest) GetActor() *Actor {
	if x != nil {
		return x.Actor
	}
	return nil
}

func (x *InvocationRequest) GetActionName() string {
	if x != nil {
		return x.ActionName
	}
	return ""
}

func (m *InvocationRequest) GetPayload() isInvocationRequest_Payload {
	if m != nil {
		return m.Payload
	}
	return nil
}

func (x *InvocationRequest) GetValue() *anypb.Any {
	if x, ok := x.GetPayload().(*InvocationRequest_Value); ok {
		return x.Value
	}
	return nil
}

func (x *InvocationRequest) GetNoop() *Noop {
	if x, ok := x.GetPayload().(*InvocationRequest_Noop); ok {
		return x.Noop
	}
	return nil
}

func (x *InvocationRequest) GetAsync() bool {
	if x != nil {
		return x.Async
	}
	return false
}

func (x *InvocationRequest) GetCaller() *ActorId {
	if x != nil {
		return x.Caller
	}
	return nil
}

func (x *InvocationRequest) GetMetadata() map[string]string {
	if x != nil {
		return x.Metadata
	}
	return nil
}

func (x *InvocationRequest) GetScheduledTo() int64 {
	if x != nil {
		return x.ScheduledTo
	}
	return 0
}

func (x *InvocationRequest) GetPooled() bool {
	if x != nil {
		return x.Pooled
	}
	return false
}

func (x *InvocationRequest) GetRegisterRef() string {
	if x != nil {
		return x.RegisterRef
	}
	return ""
}

type isInvocationRequest_Payload interface {
	isInvocationRequest_Payload()
}

type InvocationRequest_Value struct {
	Value *anypb.Any `protobuf:"bytes,4,opt,name=value,proto3,oneof"`
}

type InvocationRequest_Noop struct {
	Noop *Noop `protobuf:"bytes,7,opt,name=noop,proto3,oneof"`
}

func (*InvocationRequest_Value) isInvocationRequest_Payload() {}

func (*InvocationRequest_Noop) isInvocationRequest_Payload() {}

// ActorInvocation is a translation message between a local invocation made via
// InvocationRequest and the real Actor that intends to respond to this
// invocation and that can be located anywhere in the cluster.
//
// Params:
//   - actor: The ActorId handling the InvocationRequest request, also called
//     the target Actor.
//   - action_name: The function or method on the target Actor that will receive
//     this request
//     and perform some useful computation with the sent data.
//   - current_context: The current Context with current state value of the
//     target Actor.
//     That is, the same as found via matching in %Actor{name:
//     target_actor, state: %ActorState{state: value} =
//     actor_state}. In this case, the Context type will contain
//     in the value attribute the same `value` as the matching
//     above.
//   - payload: The value to be passed to the function or method corresponding
//     to action_name.
type ActorInvocation struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Actor          *ActorId `protobuf:"bytes,1,opt,name=actor,proto3" json:"actor,omitempty"`
	ActionName     string   `protobuf:"bytes,2,opt,name=action_name,json=actionName,proto3" json:"action_name,omitempty"`
	CurrentContext *Context `protobuf:"bytes,3,opt,name=current_context,json=currentContext,proto3" json:"current_context,omitempty"`
	// Types that are assignable to Payload:
	//
	//	*ActorInvocation_Value
	//	*ActorInvocation_Noop
	Payload isActorInvocation_Payload `protobuf_oneof:"payload"`
	Caller  *ActorId                  `protobuf:"bytes,6,opt,name=caller,proto3" json:"caller,omitempty"`
}

func (x *ActorInvocation) Reset() {
	*x = ActorInvocation{}
	mi := &file_eigr_functions_protocol_actors_protocol_proto_msgTypes[16]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ActorInvocation) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ActorInvocation) ProtoMessage() {}

func (x *ActorInvocation) ProtoReflect() protoreflect.Message {
	mi := &file_eigr_functions_protocol_actors_protocol_proto_msgTypes[16]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ActorInvocation.ProtoReflect.Descriptor instead.
func (*ActorInvocation) Descriptor() ([]byte, []int) {
	return file_eigr_functions_protocol_actors_protocol_proto_rawDescGZIP(), []int{16}
}

func (x *ActorInvocation) GetActor() *ActorId {
	if x != nil {
		return x.Actor
	}
	return nil
}

func (x *ActorInvocation) GetActionName() string {
	if x != nil {
		return x.ActionName
	}
	return ""
}

func (x *ActorInvocation) GetCurrentContext() *Context {
	if x != nil {
		return x.CurrentContext
	}
	return nil
}

func (m *ActorInvocation) GetPayload() isActorInvocation_Payload {
	if m != nil {
		return m.Payload
	}
	return nil
}

func (x *ActorInvocation) GetValue() *anypb.Any {
	if x, ok := x.GetPayload().(*ActorInvocation_Value); ok {
		return x.Value
	}
	return nil
}

func (x *ActorInvocation) GetNoop() *Noop {
	if x, ok := x.GetPayload().(*ActorInvocation_Noop); ok {
		return x.Noop
	}
	return nil
}

func (x *ActorInvocation) GetCaller() *ActorId {
	if x != nil {
		return x.Caller
	}
	return nil
}

type isActorInvocation_Payload interface {
	isActorInvocation_Payload()
}

type ActorInvocation_Value struct {
	Value *anypb.Any `protobuf:"bytes,4,opt,name=value,proto3,oneof"`
}

type ActorInvocation_Noop struct {
	Noop *Noop `protobuf:"bytes,5,opt,name=noop,proto3,oneof"`
}

func (*ActorInvocation_Value) isActorInvocation_Payload() {}

func (*ActorInvocation_Noop) isActorInvocation_Payload() {}

// The user function's response after executing the action originated by the
// local proxy request via ActorInvocation.
//
// Params:
//
//	actor_name: The name of the Actor handling the InvocationRequest request,
//	also called the target Actor. actor_system: The name of ActorSystem
//	registered in Registration step. updated_context: The Context with updated
//	state value of the target Actor after user function has processed a
//	request. value: The value that the original request proxy will forward in
//	response to the InvocationRequest type request.
//	       This is the final response from the point of view of the user who
//	       invoked the Actor call and its subsequent processing.
type ActorInvocationResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ActorName      string   `protobuf:"bytes,1,opt,name=actor_name,json=actorName,proto3" json:"actor_name,omitempty"`
	ActorSystem    string   `protobuf:"bytes,2,opt,name=actor_system,json=actorSystem,proto3" json:"actor_system,omitempty"`
	UpdatedContext *Context `protobuf:"bytes,3,opt,name=updated_context,json=updatedContext,proto3" json:"updated_context,omitempty"`
	// Types that are assignable to Payload:
	//
	//	*ActorInvocationResponse_Value
	//	*ActorInvocationResponse_Noop
	Payload    isActorInvocationResponse_Payload `protobuf_oneof:"payload"`
	Workflow   *Workflow                         `protobuf:"bytes,5,opt,name=workflow,proto3" json:"workflow,omitempty"`
	Checkpoint bool                              `protobuf:"varint,7,opt,name=checkpoint,proto3" json:"checkpoint,omitempty"`
}

func (x *ActorInvocationResponse) Reset() {
	*x = ActorInvocationResponse{}
	mi := &file_eigr_functions_protocol_actors_protocol_proto_msgTypes[17]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ActorInvocationResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ActorInvocationResponse) ProtoMessage() {}

func (x *ActorInvocationResponse) ProtoReflect() protoreflect.Message {
	mi := &file_eigr_functions_protocol_actors_protocol_proto_msgTypes[17]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ActorInvocationResponse.ProtoReflect.Descriptor instead.
func (*ActorInvocationResponse) Descriptor() ([]byte, []int) {
	return file_eigr_functions_protocol_actors_protocol_proto_rawDescGZIP(), []int{17}
}

func (x *ActorInvocationResponse) GetActorName() string {
	if x != nil {
		return x.ActorName
	}
	return ""
}

func (x *ActorInvocationResponse) GetActorSystem() string {
	if x != nil {
		return x.ActorSystem
	}
	return ""
}

func (x *ActorInvocationResponse) GetUpdatedContext() *Context {
	if x != nil {
		return x.UpdatedContext
	}
	return nil
}

func (m *ActorInvocationResponse) GetPayload() isActorInvocationResponse_Payload {
	if m != nil {
		return m.Payload
	}
	return nil
}

func (x *ActorInvocationResponse) GetValue() *anypb.Any {
	if x, ok := x.GetPayload().(*ActorInvocationResponse_Value); ok {
		return x.Value
	}
	return nil
}

func (x *ActorInvocationResponse) GetNoop() *Noop {
	if x, ok := x.GetPayload().(*ActorInvocationResponse_Noop); ok {
		return x.Noop
	}
	return nil
}

func (x *ActorInvocationResponse) GetWorkflow() *Workflow {
	if x != nil {
		return x.Workflow
	}
	return nil
}

func (x *ActorInvocationResponse) GetCheckpoint() bool {
	if x != nil {
		return x.Checkpoint
	}
	return false
}

type isActorInvocationResponse_Payload interface {
	isActorInvocationResponse_Payload()
}

type ActorInvocationResponse_Value struct {
	Value *anypb.Any `protobuf:"bytes,4,opt,name=value,proto3,oneof"`
}

type ActorInvocationResponse_Noop struct {
	Noop *Noop `protobuf:"bytes,6,opt,name=noop,proto3,oneof"`
}

func (*ActorInvocationResponse_Value) isActorInvocationResponse_Payload() {}

func (*ActorInvocationResponse_Noop) isActorInvocationResponse_Payload() {}

// InvocationResponse is the response that the proxy that received the
// InvocationRequest request will forward to the request's original user
// function.
//
// Params:
//
//	status: Status of request. Could be one of [UNKNOWN, OK, ACTOR_NOT_FOUND,
//	ERROR]. system: The original ActorSystem of the InvocationRequest request.
//	actor: The target Actor originally sent in the InvocationRequest message.
//	value: The value resulting from the request processing that the target
//	Actor made.
//	       This value must be passed by the user function to the one who
//	       requested the initial request in InvocationRequest.
type InvocationResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Status *RequestStatus `protobuf:"bytes,1,opt,name=status,proto3" json:"status,omitempty"`
	System *ActorSystem   `protobuf:"bytes,2,opt,name=system,proto3" json:"system,omitempty"`
	Actor  *Actor         `protobuf:"bytes,3,opt,name=actor,proto3" json:"actor,omitempty"`
	// Types that are assignable to Payload:
	//
	//	*InvocationResponse_Value
	//	*InvocationResponse_Noop
	Payload isInvocationResponse_Payload `protobuf_oneof:"payload"`
}

func (x *InvocationResponse) Reset() {
	*x = InvocationResponse{}
	mi := &file_eigr_functions_protocol_actors_protocol_proto_msgTypes[18]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *InvocationResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InvocationResponse) ProtoMessage() {}

func (x *InvocationResponse) ProtoReflect() protoreflect.Message {
	mi := &file_eigr_functions_protocol_actors_protocol_proto_msgTypes[18]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InvocationResponse.ProtoReflect.Descriptor instead.
func (*InvocationResponse) Descriptor() ([]byte, []int) {
	return file_eigr_functions_protocol_actors_protocol_proto_rawDescGZIP(), []int{18}
}

func (x *InvocationResponse) GetStatus() *RequestStatus {
	if x != nil {
		return x.Status
	}
	return nil
}

func (x *InvocationResponse) GetSystem() *ActorSystem {
	if x != nil {
		return x.System
	}
	return nil
}

func (x *InvocationResponse) GetActor() *Actor {
	if x != nil {
		return x.Actor
	}
	return nil
}

func (m *InvocationResponse) GetPayload() isInvocationResponse_Payload {
	if m != nil {
		return m.Payload
	}
	return nil
}

func (x *InvocationResponse) GetValue() *anypb.Any {
	if x, ok := x.GetPayload().(*InvocationResponse_Value); ok {
		return x.Value
	}
	return nil
}

func (x *InvocationResponse) GetNoop() *Noop {
	if x, ok := x.GetPayload().(*InvocationResponse_Noop); ok {
		return x.Noop
	}
	return nil
}

type isInvocationResponse_Payload interface {
	isInvocationResponse_Payload()
}

type InvocationResponse_Value struct {
	Value *anypb.Any `protobuf:"bytes,4,opt,name=value,proto3,oneof"`
}

type InvocationResponse_Noop struct {
	Noop *Noop `protobuf:"bytes,5,opt,name=noop,proto3,oneof"`
}

func (*InvocationResponse_Value) isInvocationResponse_Payload() {}

func (*InvocationResponse_Noop) isInvocationResponse_Payload() {}

type RequestStatus struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Status  Status `protobuf:"varint,1,opt,name=status,proto3,enum=eigr.functions.protocol.Status" json:"status,omitempty"`
	Message string `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *RequestStatus) Reset() {
	*x = RequestStatus{}
	mi := &file_eigr_functions_protocol_actors_protocol_proto_msgTypes[19]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RequestStatus) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RequestStatus) ProtoMessage() {}

func (x *RequestStatus) ProtoReflect() protoreflect.Message {
	mi := &file_eigr_functions_protocol_actors_protocol_proto_msgTypes[19]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RequestStatus.ProtoReflect.Descriptor instead.
func (*RequestStatus) Descriptor() ([]byte, []int) {
	return file_eigr_functions_protocol_actors_protocol_proto_rawDescGZIP(), []int{19}
}

func (x *RequestStatus) GetStatus() Status {
	if x != nil {
		return x.Status
	}
	return Status_UNKNOWN
}

func (x *RequestStatus) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

var File_eigr_functions_protocol_actors_protocol_proto protoreflect.FileDescriptor

var file_eigr_functions_protocol_actors_protocol_proto_rawDesc = []byte{
	0x0a, 0x2d, 0x65, 0x69, 0x67, 0x72, 0x2f, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2f, 0x61, 0x63, 0x74, 0x6f, 0x72, 0x73,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x17, 0x65, 0x69, 0x67, 0x72, 0x2e, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x1a, 0x2a, 0x65, 0x69, 0x67, 0x72, 0x2f, 0x66,
	0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f,
	0x6c, 0x2f, 0x61, 0x63, 0x74, 0x6f, 0x72, 0x73, 0x2f, 0x61, 0x63, 0x74, 0x6f, 0x72, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x19, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x61, 0x6e, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x22, 0xb5, 0x03, 0x0a, 0x07, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x12, 0x2a, 0x0a, 0x05,
	0x73, 0x74, 0x61, 0x74, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x41, 0x6e,
	0x79, 0x52, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x12, 0x4a, 0x0a, 0x08, 0x6d, 0x65, 0x74, 0x61,
	0x64, 0x61, 0x74, 0x61, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x2e, 0x2e, 0x65, 0x69, 0x67,
	0x72, 0x2e, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x63, 0x6f, 0x6c, 0x2e, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x2e, 0x4d, 0x65, 0x74,
	0x61, 0x64, 0x61, 0x74, 0x61, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x08, 0x6d, 0x65, 0x74, 0x61,
	0x64, 0x61, 0x74, 0x61, 0x12, 0x3e, 0x0a, 0x04, 0x74, 0x61, 0x67, 0x73, 0x18, 0x05, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x2a, 0x2e, 0x65, 0x69, 0x67, 0x72, 0x2e, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69,
	0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2e, 0x43, 0x6f, 0x6e,
	0x74, 0x65, 0x78, 0x74, 0x2e, 0x54, 0x61, 0x67, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x04,
	0x74, 0x61, 0x67, 0x73, 0x12, 0x3f, 0x0a, 0x06, 0x63, 0x61, 0x6c, 0x6c, 0x65, 0x72, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x27, 0x2e, 0x65, 0x69, 0x67, 0x72, 0x2e, 0x66, 0x75, 0x6e, 0x63,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2e, 0x61,
	0x63, 0x74, 0x6f, 0x72, 0x73, 0x2e, 0x41, 0x63, 0x74, 0x6f, 0x72, 0x49, 0x64, 0x52, 0x06, 0x63,
	0x61, 0x6c, 0x6c, 0x65, 0x72, 0x12, 0x3b, 0x0a, 0x04, 0x73, 0x65, 0x6c, 0x66, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x27, 0x2e, 0x65, 0x69, 0x67, 0x72, 0x2e, 0x66, 0x75, 0x6e, 0x63, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2e, 0x61, 0x63,
	0x74, 0x6f, 0x72, 0x73, 0x2e, 0x41, 0x63, 0x74, 0x6f, 0x72, 0x49, 0x64, 0x52, 0x04, 0x73, 0x65,
	0x6c, 0x66, 0x1a, 0x3b, 0x0a, 0x0d, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x45, 0x6e,
	0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x1a,
	0x37, 0x0a, 0x09, 0x54, 0x61, 0x67, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03,
	0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14,
	0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x06, 0x0a, 0x04, 0x4e, 0x6f, 0x6f, 0x70,
	0x22, 0x24, 0x0a, 0x08, 0x4a, 0x53, 0x4f, 0x4e, 0x54, 0x79, 0x70, 0x65, 0x12, 0x18, 0x0a, 0x07,
	0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x63,
	0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x22, 0xae, 0x01, 0x0a, 0x13, 0x52, 0x65, 0x67, 0x69, 0x73,
	0x74, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x47,
	0x0a, 0x0c, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x5f, 0x69, 0x6e, 0x66, 0x6f, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x24, 0x2e, 0x65, 0x69, 0x67, 0x72, 0x2e, 0x66, 0x75, 0x6e, 0x63,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2e, 0x53,
	0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x0b, 0x73, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x4e, 0x0a, 0x0c, 0x61, 0x63, 0x74, 0x6f, 0x72,
	0x5f, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2b, 0x2e,
	0x65, 0x69, 0x67, 0x72, 0x2e, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2e, 0x61, 0x63, 0x74, 0x6f, 0x72, 0x73, 0x2e, 0x41,
	0x63, 0x74, 0x6f, 0x72, 0x53, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x52, 0x0b, 0x61, 0x63, 0x74, 0x6f,
	0x72, 0x53, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x22, 0x99, 0x01, 0x0a, 0x14, 0x52, 0x65, 0x67, 0x69,
	0x73, 0x74, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x3e, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x26, 0x2e, 0x65, 0x69, 0x67, 0x72, 0x2e, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2e, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x12, 0x41, 0x0a, 0x0a, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x5f, 0x69, 0x6e, 0x66, 0x6f, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x22, 0x2e, 0x65, 0x69, 0x67, 0x72, 0x2e, 0x66, 0x75, 0x6e, 0x63,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2e, 0x50,
	0x72, 0x6f, 0x78, 0x79, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x09, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x49,
	0x6e, 0x66, 0x6f, 0x22, 0xd8, 0x02, 0x0a, 0x0b, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x49,
	0x6e, 0x66, 0x6f, 0x12, 0x21, 0x0a, 0x0c, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x5f, 0x6e,
	0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x73, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x27, 0x0a, 0x0f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x5f, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12,
	0x27, 0x0a, 0x0f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x5f, 0x72, 0x75, 0x6e, 0x74, 0x69,
	0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x52, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x12, 0x30, 0x0a, 0x14, 0x73, 0x75, 0x70, 0x70,
	0x6f, 0x72, 0x74, 0x5f, 0x6c, 0x69, 0x62, 0x72, 0x61, 0x72, 0x79, 0x5f, 0x6e, 0x61, 0x6d, 0x65,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x12, 0x73, 0x75, 0x70, 0x70, 0x6f, 0x72, 0x74, 0x4c,
	0x69, 0x62, 0x72, 0x61, 0x72, 0x79, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x36, 0x0a, 0x17, 0x73, 0x75,
	0x70, 0x70, 0x6f, 0x72, 0x74, 0x5f, 0x6c, 0x69, 0x62, 0x72, 0x61, 0x72, 0x79, 0x5f, 0x76, 0x65,
	0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x15, 0x73, 0x75, 0x70,
	0x70, 0x6f, 0x72, 0x74, 0x4c, 0x69, 0x62, 0x72, 0x61, 0x72, 0x79, 0x56, 0x65, 0x72, 0x73, 0x69,
	0x6f, 0x6e, 0x12, 0x34, 0x0a, 0x16, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x5f, 0x6d,
	0x61, 0x6a, 0x6f, 0x72, 0x5f, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x06, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x14, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x4d, 0x61, 0x6a, 0x6f,
	0x72, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x34, 0x0a, 0x16, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x63, 0x6f, 0x6c, 0x5f, 0x6d, 0x69, 0x6e, 0x6f, 0x72, 0x5f, 0x76, 0x65, 0x72, 0x73, 0x69,
	0x6f, 0x6e, 0x18, 0x07, 0x20, 0x01, 0x28, 0x05, 0x52, 0x14, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63,
	0x6f, 0x6c, 0x4d, 0x69, 0x6e, 0x6f, 0x72, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x22, 0x4f,
	0x0a, 0x0c, 0x53, 0x70, 0x61, 0x77, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x3f,
	0x0a, 0x06, 0x61, 0x63, 0x74, 0x6f, 0x72, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x27,
	0x2e, 0x65, 0x69, 0x67, 0x72, 0x2e, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2e, 0x61, 0x63, 0x74, 0x6f, 0x72, 0x73, 0x2e,
	0x41, 0x63, 0x74, 0x6f, 0x72, 0x49, 0x64, 0x52, 0x06, 0x61, 0x63, 0x74, 0x6f, 0x72, 0x73, 0x22,
	0x4f, 0x0a, 0x0d, 0x53, 0x70, 0x61, 0x77, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x3e, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x26, 0x2e, 0x65, 0x69, 0x67, 0x72, 0x2e, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2e, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x22, 0xbb, 0x01, 0x0a, 0x09, 0x50, 0x72, 0x6f, 0x78, 0x79, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x34,
	0x0a, 0x16, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x5f, 0x6d, 0x61, 0x6a, 0x6f, 0x72,
	0x5f, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x14,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x4d, 0x61, 0x6a, 0x6f, 0x72, 0x56, 0x65, 0x72,
	0x73, 0x69, 0x6f, 0x6e, 0x12, 0x34, 0x0a, 0x16, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c,
	0x5f, 0x6d, 0x69, 0x6e, 0x6f, 0x72, 0x5f, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x14, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x4d, 0x69,
	0x6e, 0x6f, 0x72, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x1d, 0x0a, 0x0a, 0x70, 0x72,
	0x6f, 0x78, 0x79, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09,
	0x70, 0x72, 0x6f, 0x78, 0x79, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x23, 0x0a, 0x0d, 0x70, 0x72, 0x6f,
	0x78, 0x79, 0x5f, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0c, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x22, 0x52,
	0x0a, 0x0a, 0x53, 0x69, 0x64, 0x65, 0x45, 0x66, 0x66, 0x65, 0x63, 0x74, 0x12, 0x44, 0x0a, 0x07,
	0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2a, 0x2e,
	0x65, 0x69, 0x67, 0x72, 0x2e, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2e, 0x49, 0x6e, 0x76, 0x6f, 0x63, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x52, 0x07, 0x72, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x22, 0x9e, 0x01, 0x0a, 0x09, 0x42, 0x72, 0x6f, 0x61, 0x64, 0x63, 0x61, 0x73, 0x74,
	0x12, 0x23, 0x0a, 0x0d, 0x63, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x5f, 0x67, 0x72, 0x6f, 0x75,
	0x70, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x63, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c,
	0x47, 0x72, 0x6f, 0x75, 0x70, 0x12, 0x2c, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x41, 0x6e, 0x79, 0x48, 0x00, 0x52, 0x05, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x12, 0x33, 0x0a, 0x04, 0x6e, 0x6f, 0x6f, 0x70, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x1d, 0x2e, 0x65, 0x69, 0x67, 0x72, 0x2e, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2e, 0x4e, 0x6f, 0x6f, 0x70,
	0x48, 0x00, 0x52, 0x04, 0x6e, 0x6f, 0x6f, 0x70, 0x42, 0x09, 0x0a, 0x07, 0x70, 0x61, 0x79, 0x6c,
	0x6f, 0x61, 0x64, 0x22, 0x3d, 0x0a, 0x04, 0x50, 0x69, 0x70, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x61,
	0x63, 0x74, 0x6f, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x61, 0x63, 0x74, 0x6f,
	0x72, 0x12, 0x1f, 0x0a, 0x0b, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x6e, 0x61, 0x6d, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x4e, 0x61,
	0x6d, 0x65, 0x22, 0x40, 0x0a, 0x07, 0x46, 0x6f, 0x72, 0x77, 0x61, 0x72, 0x64, 0x12, 0x14, 0x0a,
	0x05, 0x61, 0x63, 0x74, 0x6f, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x61, 0x63,
	0x74, 0x6f, 0x72, 0x12, 0x1f, 0x0a, 0x0b, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x6e, 0x61,
	0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x4e, 0x61, 0x6d, 0x65, 0x22, 0x86, 0x02, 0x0a, 0x04, 0x46, 0x61, 0x63, 0x74, 0x12, 0x12, 0x0a,
	0x04, 0x75, 0x75, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x75, 0x75, 0x69,
	0x64, 0x12, 0x2a, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x14, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x2e, 0x41, 0x6e, 0x79, 0x52, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x12, 0x47, 0x0a,
	0x08, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x2b, 0x2e, 0x65, 0x69, 0x67, 0x72, 0x2e, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2e, 0x46, 0x61, 0x63, 0x74, 0x2e, 0x4d,
	0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x08, 0x6d, 0x65,
	0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x12, 0x38, 0x0a, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74,
	0x61, 0x6d, 0x70, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65,
	0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70,
	0x1a, 0x3b, 0x0a, 0x0d, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x45, 0x6e, 0x74, 0x72,
	0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03,
	0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x89, 0x02,
	0x0a, 0x08, 0x57, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x12, 0x40, 0x0a, 0x09, 0x62, 0x72,
	0x6f, 0x61, 0x64, 0x63, 0x61, 0x73, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x22, 0x2e,
	0x65, 0x69, 0x67, 0x72, 0x2e, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2e, 0x42, 0x72, 0x6f, 0x61, 0x64, 0x63, 0x61, 0x73,
	0x74, 0x52, 0x09, 0x62, 0x72, 0x6f, 0x61, 0x64, 0x63, 0x61, 0x73, 0x74, 0x12, 0x3d, 0x0a, 0x07,
	0x65, 0x66, 0x66, 0x65, 0x63, 0x74, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x23, 0x2e,
	0x65, 0x69, 0x67, 0x72, 0x2e, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2e, 0x53, 0x69, 0x64, 0x65, 0x45, 0x66, 0x66, 0x65,
	0x63, 0x74, 0x52, 0x07, 0x65, 0x66, 0x66, 0x65, 0x63, 0x74, 0x73, 0x12, 0x33, 0x0a, 0x04, 0x70,
	0x69, 0x70, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x65, 0x69, 0x67, 0x72,
	0x2e, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x63, 0x6f, 0x6c, 0x2e, 0x50, 0x69, 0x70, 0x65, 0x48, 0x00, 0x52, 0x04, 0x70, 0x69, 0x70, 0x65,
	0x12, 0x3c, 0x0a, 0x07, 0x66, 0x6f, 0x72, 0x77, 0x61, 0x72, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x20, 0x2e, 0x65, 0x69, 0x67, 0x72, 0x2e, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2e, 0x46, 0x6f, 0x72, 0x77,
	0x61, 0x72, 0x64, 0x48, 0x00, 0x52, 0x07, 0x66, 0x6f, 0x72, 0x77, 0x61, 0x72, 0x64, 0x42, 0x09,
	0x0a, 0x07, 0x72, 0x6f, 0x75, 0x74, 0x69, 0x6e, 0x67, 0x22, 0xec, 0x04, 0x0a, 0x11, 0x49, 0x6e,
	0x76, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x43, 0x0a, 0x06, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x2b, 0x2e, 0x65, 0x69, 0x67, 0x72, 0x2e, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2e, 0x61, 0x63, 0x74, 0x6f, 0x72, 0x73,
	0x2e, 0x41, 0x63, 0x74, 0x6f, 0x72, 0x53, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x52, 0x06, 0x73, 0x79,
	0x73, 0x74, 0x65, 0x6d, 0x12, 0x3b, 0x0a, 0x05, 0x61, 0x63, 0x74, 0x6f, 0x72, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x25, 0x2e, 0x65, 0x69, 0x67, 0x72, 0x2e, 0x66, 0x75, 0x6e, 0x63, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2e, 0x61, 0x63,
	0x74, 0x6f, 0x72, 0x73, 0x2e, 0x41, 0x63, 0x74, 0x6f, 0x72, 0x52, 0x05, 0x61, 0x63, 0x74, 0x6f,
	0x72, 0x12, 0x1f, 0x0a, 0x0b, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x6e, 0x61, 0x6d, 0x65,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x4e, 0x61,
	0x6d, 0x65, 0x12, 0x2c, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x14, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2e, 0x41, 0x6e, 0x79, 0x48, 0x00, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x12, 0x33, 0x0a, 0x04, 0x6e, 0x6f, 0x6f, 0x70, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d,
	0x2e, 0x65, 0x69, 0x67, 0x72, 0x2e, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2e, 0x4e, 0x6f, 0x6f, 0x70, 0x48, 0x00, 0x52,
	0x04, 0x6e, 0x6f, 0x6f, 0x70, 0x12, 0x14, 0x0a, 0x05, 0x61, 0x73, 0x79, 0x6e, 0x63, 0x18, 0x05,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x05, 0x61, 0x73, 0x79, 0x6e, 0x63, 0x12, 0x3f, 0x0a, 0x06, 0x63,
	0x61, 0x6c, 0x6c, 0x65, 0x72, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x27, 0x2e, 0x65, 0x69,
	0x67, 0x72, 0x2e, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2e, 0x61, 0x63, 0x74, 0x6f, 0x72, 0x73, 0x2e, 0x41, 0x63, 0x74,
	0x6f, 0x72, 0x49, 0x64, 0x52, 0x06, 0x63, 0x61, 0x6c, 0x6c, 0x65, 0x72, 0x12, 0x54, 0x0a, 0x08,
	0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x18, 0x08, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x38,
	0x2e, 0x65, 0x69, 0x67, 0x72, 0x2e, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2e, 0x49, 0x6e, 0x76, 0x6f, 0x63, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x2e, 0x4d, 0x65, 0x74, 0x61, 0x64,
	0x61, 0x74, 0x61, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x08, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61,
	0x74, 0x61, 0x12, 0x21, 0x0a, 0x0c, 0x73, 0x63, 0x68, 0x65, 0x64, 0x75, 0x6c, 0x65, 0x64, 0x5f,
	0x74, 0x6f, 0x18, 0x09, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0b, 0x73, 0x63, 0x68, 0x65, 0x64, 0x75,
	0x6c, 0x65, 0x64, 0x54, 0x6f, 0x12, 0x16, 0x0a, 0x06, 0x70, 0x6f, 0x6f, 0x6c, 0x65, 0x64, 0x18,
	0x0a, 0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x70, 0x6f, 0x6f, 0x6c, 0x65, 0x64, 0x12, 0x21, 0x0a,
	0x0c, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x5f, 0x72, 0x65, 0x66, 0x18, 0x0b, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x0b, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x52, 0x65, 0x66,
	0x1a, 0x3b, 0x0a, 0x0d, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x45, 0x6e, 0x74, 0x72,
	0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03,
	0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x42, 0x09, 0x0a,
	0x07, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x22, 0xeb, 0x02, 0x0a, 0x0f, 0x41, 0x63, 0x74,
	0x6f, 0x72, 0x49, 0x6e, 0x76, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x3d, 0x0a, 0x05,
	0x61, 0x63, 0x74, 0x6f, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x27, 0x2e, 0x65, 0x69,
	0x67, 0x72, 0x2e, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2e, 0x61, 0x63, 0x74, 0x6f, 0x72, 0x73, 0x2e, 0x41, 0x63, 0x74,
	0x6f, 0x72, 0x49, 0x64, 0x52, 0x05, 0x61, 0x63, 0x74, 0x6f, 0x72, 0x12, 0x1f, 0x0a, 0x0b, 0x61,
	0x63, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0a, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x49, 0x0a, 0x0f,
	0x63, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x20, 0x2e, 0x65, 0x69, 0x67, 0x72, 0x2e, 0x66, 0x75, 0x6e,
	0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2e,
	0x43, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x52, 0x0e, 0x63, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74,
	0x43, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x12, 0x2c, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x41, 0x6e, 0x79, 0x48, 0x00, 0x52, 0x05,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x33, 0x0a, 0x04, 0x6e, 0x6f, 0x6f, 0x70, 0x18, 0x05, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x65, 0x69, 0x67, 0x72, 0x2e, 0x66, 0x75, 0x6e, 0x63, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2e, 0x4e, 0x6f,
	0x6f, 0x70, 0x48, 0x00, 0x52, 0x04, 0x6e, 0x6f, 0x6f, 0x70, 0x12, 0x3f, 0x0a, 0x06, 0x63, 0x61,
	0x6c, 0x6c, 0x65, 0x72, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x27, 0x2e, 0x65, 0x69, 0x67,
	0x72, 0x2e, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x63, 0x6f, 0x6c, 0x2e, 0x61, 0x63, 0x74, 0x6f, 0x72, 0x73, 0x2e, 0x41, 0x63, 0x74, 0x6f,
	0x72, 0x49, 0x64, 0x52, 0x06, 0x63, 0x61, 0x6c, 0x6c, 0x65, 0x72, 0x42, 0x09, 0x0a, 0x07, 0x70,
	0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x22, 0xf3, 0x02, 0x0a, 0x17, 0x41, 0x63, 0x74, 0x6f, 0x72,
	0x49, 0x6e, 0x76, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x1d, 0x0a, 0x0a, 0x61, 0x63, 0x74, 0x6f, 0x72, 0x5f, 0x6e, 0x61, 0x6d, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x61, 0x63, 0x74, 0x6f, 0x72, 0x4e, 0x61, 0x6d,
	0x65, 0x12, 0x21, 0x0a, 0x0c, 0x61, 0x63, 0x74, 0x6f, 0x72, 0x5f, 0x73, 0x79, 0x73, 0x74, 0x65,
	0x6d, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x61, 0x63, 0x74, 0x6f, 0x72, 0x53, 0x79,
	0x73, 0x74, 0x65, 0x6d, 0x12, 0x49, 0x0a, 0x0f, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x5f,
	0x63, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x20, 0x2e,
	0x65, 0x69, 0x67, 0x72, 0x2e, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2e, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x52,
	0x0e, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x12,
	0x2c, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2e, 0x41, 0x6e, 0x79, 0x48, 0x00, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x33, 0x0a,
	0x04, 0x6e, 0x6f, 0x6f, 0x70, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x65, 0x69,
	0x67, 0x72, 0x2e, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2e, 0x4e, 0x6f, 0x6f, 0x70, 0x48, 0x00, 0x52, 0x04, 0x6e, 0x6f,
	0x6f, 0x70, 0x12, 0x3d, 0x0a, 0x08, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x18, 0x05,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x21, 0x2e, 0x65, 0x69, 0x67, 0x72, 0x2e, 0x66, 0x75, 0x6e, 0x63,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2e, 0x57,
	0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x52, 0x08, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f,
	0x77, 0x12, 0x1e, 0x0a, 0x0a, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x18,
	0x07, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0a, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x70, 0x6f, 0x69, 0x6e,
	0x74, 0x42, 0x09, 0x0a, 0x07, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x22, 0xc4, 0x02, 0x0a,
	0x12, 0x49, 0x6e, 0x76, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x3e, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x26, 0x2e, 0x65, 0x69, 0x67, 0x72, 0x2e, 0x66, 0x75, 0x6e, 0x63, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2e, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x06, 0x73, 0x74, 0x61,
	0x74, 0x75, 0x73, 0x12, 0x43, 0x0a, 0x06, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x2b, 0x2e, 0x65, 0x69, 0x67, 0x72, 0x2e, 0x66, 0x75, 0x6e, 0x63, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2e, 0x61, 0x63,
	0x74, 0x6f, 0x72, 0x73, 0x2e, 0x41, 0x63, 0x74, 0x6f, 0x72, 0x53, 0x79, 0x73, 0x74, 0x65, 0x6d,
	0x52, 0x06, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x12, 0x3b, 0x0a, 0x05, 0x61, 0x63, 0x74, 0x6f,
	0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x25, 0x2e, 0x65, 0x69, 0x67, 0x72, 0x2e, 0x66,
	0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f,
	0x6c, 0x2e, 0x61, 0x63, 0x74, 0x6f, 0x72, 0x73, 0x2e, 0x41, 0x63, 0x74, 0x6f, 0x72, 0x52, 0x05,
	0x61, 0x63, 0x74, 0x6f, 0x72, 0x12, 0x2c, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x41, 0x6e, 0x79, 0x48, 0x00, 0x52, 0x05, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x12, 0x33, 0x0a, 0x04, 0x6e, 0x6f, 0x6f, 0x70, 0x18, 0x05, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x1d, 0x2e, 0x65, 0x69, 0x67, 0x72, 0x2e, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2e, 0x4e, 0x6f, 0x6f, 0x70,
	0x48, 0x00, 0x52, 0x04, 0x6e, 0x6f, 0x6f, 0x70, 0x42, 0x09, 0x0a, 0x07, 0x70, 0x61, 0x79, 0x6c,
	0x6f, 0x61, 0x64, 0x22, 0x62, 0x0a, 0x0d, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x53, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x12, 0x37, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0e, 0x32, 0x1f, 0x2e, 0x65, 0x69, 0x67, 0x72, 0x2e, 0x66, 0x75, 0x6e, 0x63,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2e, 0x53,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x18, 0x0a,
	0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07,
	0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x2a, 0x3d, 0x0a, 0x06, 0x53, 0x74, 0x61, 0x74, 0x75,
	0x73, 0x12, 0x0b, 0x0a, 0x07, 0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57, 0x4e, 0x10, 0x00, 0x12, 0x06,
	0x0a, 0x02, 0x4f, 0x4b, 0x10, 0x01, 0x12, 0x13, 0x0a, 0x0f, 0x41, 0x43, 0x54, 0x4f, 0x52, 0x5f,
	0x4e, 0x4f, 0x54, 0x5f, 0x46, 0x4f, 0x55, 0x4e, 0x44, 0x10, 0x02, 0x12, 0x09, 0x0a, 0x05, 0x45,
	0x52, 0x52, 0x4f, 0x52, 0x10, 0x03, 0x42, 0x42, 0x0a, 0x1a, 0x69, 0x6f, 0x2e, 0x65, 0x69, 0x67,
	0x72, 0x2e, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x63, 0x6f, 0x6c, 0x5a, 0x24, 0x73, 0x70, 0x61, 0x77, 0x6e, 0x2f, 0x65, 0x69, 0x67, 0x72,
	0x2f, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x63, 0x6f, 0x6c, 0x2f, 0x61, 0x63, 0x74, 0x6f, 0x72, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_eigr_functions_protocol_actors_protocol_proto_rawDescOnce sync.Once
	file_eigr_functions_protocol_actors_protocol_proto_rawDescData = file_eigr_functions_protocol_actors_protocol_proto_rawDesc
)

func file_eigr_functions_protocol_actors_protocol_proto_rawDescGZIP() []byte {
	file_eigr_functions_protocol_actors_protocol_proto_rawDescOnce.Do(func() {
		file_eigr_functions_protocol_actors_protocol_proto_rawDescData = protoimpl.X.CompressGZIP(file_eigr_functions_protocol_actors_protocol_proto_rawDescData)
	})
	return file_eigr_functions_protocol_actors_protocol_proto_rawDescData
}

var file_eigr_functions_protocol_actors_protocol_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_eigr_functions_protocol_actors_protocol_proto_msgTypes = make([]protoimpl.MessageInfo, 24)
var file_eigr_functions_protocol_actors_protocol_proto_goTypes = []any{
	(Status)(0),                     // 0: eigr.functions.protocol.Status
	(*Context)(nil),                 // 1: eigr.functions.protocol.Context
	(*Noop)(nil),                    // 2: eigr.functions.protocol.Noop
	(*JSONType)(nil),                // 3: eigr.functions.protocol.JSONType
	(*RegistrationRequest)(nil),     // 4: eigr.functions.protocol.RegistrationRequest
	(*RegistrationResponse)(nil),    // 5: eigr.functions.protocol.RegistrationResponse
	(*ServiceInfo)(nil),             // 6: eigr.functions.protocol.ServiceInfo
	(*SpawnRequest)(nil),            // 7: eigr.functions.protocol.SpawnRequest
	(*SpawnResponse)(nil),           // 8: eigr.functions.protocol.SpawnResponse
	(*ProxyInfo)(nil),               // 9: eigr.functions.protocol.ProxyInfo
	(*SideEffect)(nil),              // 10: eigr.functions.protocol.SideEffect
	(*Broadcast)(nil),               // 11: eigr.functions.protocol.Broadcast
	(*Pipe)(nil),                    // 12: eigr.functions.protocol.Pipe
	(*Forward)(nil),                 // 13: eigr.functions.protocol.Forward
	(*Fact)(nil),                    // 14: eigr.functions.protocol.Fact
	(*Workflow)(nil),                // 15: eigr.functions.protocol.Workflow
	(*InvocationRequest)(nil),       // 16: eigr.functions.protocol.InvocationRequest
	(*ActorInvocation)(nil),         // 17: eigr.functions.protocol.ActorInvocation
	(*ActorInvocationResponse)(nil), // 18: eigr.functions.protocol.ActorInvocationResponse
	(*InvocationResponse)(nil),      // 19: eigr.functions.protocol.InvocationResponse
	(*RequestStatus)(nil),           // 20: eigr.functions.protocol.RequestStatus
	nil,                             // 21: eigr.functions.protocol.Context.MetadataEntry
	nil,                             // 22: eigr.functions.protocol.Context.TagsEntry
	nil,                             // 23: eigr.functions.protocol.Fact.MetadataEntry
	nil,                             // 24: eigr.functions.protocol.InvocationRequest.MetadataEntry
	(*anypb.Any)(nil),               // 25: google.protobuf.Any
	(*ActorId)(nil),                 // 26: eigr.functions.protocol.actors.ActorId
	(*ActorSystem)(nil),             // 27: eigr.functions.protocol.actors.ActorSystem
	(*timestamppb.Timestamp)(nil),   // 28: google.protobuf.Timestamp
	(*Actor)(nil),                   // 29: eigr.functions.protocol.actors.Actor
}
var file_eigr_functions_protocol_actors_protocol_proto_depIdxs = []int32{
	25, // 0: eigr.functions.protocol.Context.state:type_name -> google.protobuf.Any
	21, // 1: eigr.functions.protocol.Context.metadata:type_name -> eigr.functions.protocol.Context.MetadataEntry
	22, // 2: eigr.functions.protocol.Context.tags:type_name -> eigr.functions.protocol.Context.TagsEntry
	26, // 3: eigr.functions.protocol.Context.caller:type_name -> eigr.functions.protocol.actors.ActorId
	26, // 4: eigr.functions.protocol.Context.self:type_name -> eigr.functions.protocol.actors.ActorId
	6,  // 5: eigr.functions.protocol.RegistrationRequest.service_info:type_name -> eigr.functions.protocol.ServiceInfo
	27, // 6: eigr.functions.protocol.RegistrationRequest.actor_system:type_name -> eigr.functions.protocol.actors.ActorSystem
	20, // 7: eigr.functions.protocol.RegistrationResponse.status:type_name -> eigr.functions.protocol.RequestStatus
	9,  // 8: eigr.functions.protocol.RegistrationResponse.proxy_info:type_name -> eigr.functions.protocol.ProxyInfo
	26, // 9: eigr.functions.protocol.SpawnRequest.actors:type_name -> eigr.functions.protocol.actors.ActorId
	20, // 10: eigr.functions.protocol.SpawnResponse.status:type_name -> eigr.functions.protocol.RequestStatus
	16, // 11: eigr.functions.protocol.SideEffect.request:type_name -> eigr.functions.protocol.InvocationRequest
	25, // 12: eigr.functions.protocol.Broadcast.value:type_name -> google.protobuf.Any
	2,  // 13: eigr.functions.protocol.Broadcast.noop:type_name -> eigr.functions.protocol.Noop
	25, // 14: eigr.functions.protocol.Fact.state:type_name -> google.protobuf.Any
	23, // 15: eigr.functions.protocol.Fact.metadata:type_name -> eigr.functions.protocol.Fact.MetadataEntry
	28, // 16: eigr.functions.protocol.Fact.timestamp:type_name -> google.protobuf.Timestamp
	11, // 17: eigr.functions.protocol.Workflow.broadcast:type_name -> eigr.functions.protocol.Broadcast
	10, // 18: eigr.functions.protocol.Workflow.effects:type_name -> eigr.functions.protocol.SideEffect
	12, // 19: eigr.functions.protocol.Workflow.pipe:type_name -> eigr.functions.protocol.Pipe
	13, // 20: eigr.functions.protocol.Workflow.forward:type_name -> eigr.functions.protocol.Forward
	27, // 21: eigr.functions.protocol.InvocationRequest.system:type_name -> eigr.functions.protocol.actors.ActorSystem
	29, // 22: eigr.functions.protocol.InvocationRequest.actor:type_name -> eigr.functions.protocol.actors.Actor
	25, // 23: eigr.functions.protocol.InvocationRequest.value:type_name -> google.protobuf.Any
	2,  // 24: eigr.functions.protocol.InvocationRequest.noop:type_name -> eigr.functions.protocol.Noop
	26, // 25: eigr.functions.protocol.InvocationRequest.caller:type_name -> eigr.functions.protocol.actors.ActorId
	24, // 26: eigr.functions.protocol.InvocationRequest.metadata:type_name -> eigr.functions.protocol.InvocationRequest.MetadataEntry
	26, // 27: eigr.functions.protocol.ActorInvocation.actor:type_name -> eigr.functions.protocol.actors.ActorId
	1,  // 28: eigr.functions.protocol.ActorInvocation.current_context:type_name -> eigr.functions.protocol.Context
	25, // 29: eigr.functions.protocol.ActorInvocation.value:type_name -> google.protobuf.Any
	2,  // 30: eigr.functions.protocol.ActorInvocation.noop:type_name -> eigr.functions.protocol.Noop
	26, // 31: eigr.functions.protocol.ActorInvocation.caller:type_name -> eigr.functions.protocol.actors.ActorId
	1,  // 32: eigr.functions.protocol.ActorInvocationResponse.updated_context:type_name -> eigr.functions.protocol.Context
	25, // 33: eigr.functions.protocol.ActorInvocationResponse.value:type_name -> google.protobuf.Any
	2,  // 34: eigr.functions.protocol.ActorInvocationResponse.noop:type_name -> eigr.functions.protocol.Noop
	15, // 35: eigr.functions.protocol.ActorInvocationResponse.workflow:type_name -> eigr.functions.protocol.Workflow
	20, // 36: eigr.functions.protocol.InvocationResponse.status:type_name -> eigr.functions.protocol.RequestStatus
	27, // 37: eigr.functions.protocol.InvocationResponse.system:type_name -> eigr.functions.protocol.actors.ActorSystem
	29, // 38: eigr.functions.protocol.InvocationResponse.actor:type_name -> eigr.functions.protocol.actors.Actor
	25, // 39: eigr.functions.protocol.InvocationResponse.value:type_name -> google.protobuf.Any
	2,  // 40: eigr.functions.protocol.InvocationResponse.noop:type_name -> eigr.functions.protocol.Noop
	0,  // 41: eigr.functions.protocol.RequestStatus.status:type_name -> eigr.functions.protocol.Status
	42, // [42:42] is the sub-list for method output_type
	42, // [42:42] is the sub-list for method input_type
	42, // [42:42] is the sub-list for extension type_name
	42, // [42:42] is the sub-list for extension extendee
	0,  // [0:42] is the sub-list for field type_name
}

func init() { file_eigr_functions_protocol_actors_protocol_proto_init() }
func file_eigr_functions_protocol_actors_protocol_proto_init() {
	if File_eigr_functions_protocol_actors_protocol_proto != nil {
		return
	}
	file_eigr_functions_protocol_actors_actor_proto_init()
	file_eigr_functions_protocol_actors_protocol_proto_msgTypes[10].OneofWrappers = []any{
		(*Broadcast_Value)(nil),
		(*Broadcast_Noop)(nil),
	}
	file_eigr_functions_protocol_actors_protocol_proto_msgTypes[14].OneofWrappers = []any{
		(*Workflow_Pipe)(nil),
		(*Workflow_Forward)(nil),
	}
	file_eigr_functions_protocol_actors_protocol_proto_msgTypes[15].OneofWrappers = []any{
		(*InvocationRequest_Value)(nil),
		(*InvocationRequest_Noop)(nil),
	}
	file_eigr_functions_protocol_actors_protocol_proto_msgTypes[16].OneofWrappers = []any{
		(*ActorInvocation_Value)(nil),
		(*ActorInvocation_Noop)(nil),
	}
	file_eigr_functions_protocol_actors_protocol_proto_msgTypes[17].OneofWrappers = []any{
		(*ActorInvocationResponse_Value)(nil),
		(*ActorInvocationResponse_Noop)(nil),
	}
	file_eigr_functions_protocol_actors_protocol_proto_msgTypes[18].OneofWrappers = []any{
		(*InvocationResponse_Value)(nil),
		(*InvocationResponse_Noop)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_eigr_functions_protocol_actors_protocol_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   24,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_eigr_functions_protocol_actors_protocol_proto_goTypes,
		DependencyIndexes: file_eigr_functions_protocol_actors_protocol_proto_depIdxs,
		EnumInfos:         file_eigr_functions_protocol_actors_protocol_proto_enumTypes,
		MessageInfos:      file_eigr_functions_protocol_actors_protocol_proto_msgTypes,
	}.Build()
	File_eigr_functions_protocol_actors_protocol_proto = out.File
	file_eigr_functions_protocol_actors_protocol_proto_rawDesc = nil
	file_eigr_functions_protocol_actors_protocol_proto_goTypes = nil
	file_eigr_functions_protocol_actors_protocol_proto_depIdxs = nil
}
