//
// prometheus.proto
//

//---------------------------------------------------
//
// This protobuf allows collecting of data from go programs that
// have prometheus counters and histrograms described
//
// Please note the key idea is to constrain the label entropy, so
// we have exactly x3 labels per metric
//
// The go programs have the promethus counters implmented like this:

// startTime := time.Now()
// defer func() {
//   d.pH.WithLabelValues("Deserialize", "complete", "count").Observe(time.Since(startTime).Seconds())
// }()
// d.pC.WithLabelValues("Deserialize", "start", "count").Inc()

// Error counters
// d.pC.WithLabelValues("Deserialize", "netlinkerDoneCh", "error").Inc()

// x.pC = promauto.NewCounterVec(
//   prometheus.CounterOpts{
//     Subsystem: "xtcp",
//     Name:      "counts",
//     Help:      "xtcp counts",
//   },
//   []string{"function", "variable", "type"},
// )

// Not doing the historgrams yet

// x.pH = promauto.NewSummaryVec(
//   prometheus.SummaryOpts{
//     Subsystem: "xtcp",
//     Name:      "histograms",
//     Help:      "xtcp historgrams",
//     Objectives: map[float64]float64{
//       0.1:  quantileError,
//       0.5:  quantileError,
//       0.99: quantileError,
//     },
//     MaxAge: summaryVecMaxAge,
//   },
//   []string{"function", "variable", "type"},
// )
//---------------------------------------------------

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        (unknown)
// source: prometheus/v1/prometheus_protolist.proto

package prometheus

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

// https://clickhouse.com/docs/en/interfaces/formats#protobuflist
type Envelope struct {
	state         protoimpl.MessageState        `protogen:"open.v1"`
	Rows          []*Envelope_PromRecordCounter `protobuf:"bytes,1,rep,name=rows,proto3" json:"rows,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Envelope) Reset() {
	*x = Envelope{}
	mi := &file_prometheus_v1_prometheus_protolist_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Envelope) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Envelope) ProtoMessage() {}

func (x *Envelope) ProtoReflect() protoreflect.Message {
	mi := &file_prometheus_v1_prometheus_protolist_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Envelope.ProtoReflect.Descriptor instead.
func (*Envelope) Descriptor() ([]byte, []int) {
	return file_prometheus_v1_prometheus_protolist_proto_rawDescGZIP(), []int{0}
}

func (x *Envelope) GetRows() []*Envelope_PromRecordCounter {
	if x != nil {
		return x.Rows
	}
	return nil
}

// repeated per the test 02240
// https://github.com/ClickHouse/ClickHouse/blob/master/tests/queries/0_stateless/format_schemas/02240_protobuflist_format_persons_syntax2.proto#L67
type PromRecordCounter struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	TimestampNs   float64                `protobuf:"fixed64,10,opt,name=timestamp_ns,json=timestampNs,proto3" json:"timestamp_ns,omitempty"`
	Hostname      string                 `protobuf:"bytes,20,opt,name=hostname,proto3" json:"hostname,omitempty"`
	Pop           string                 `protobuf:"bytes,30,opt,name=pop,proto3" json:"pop,omitempty"`
	Label         string                 `protobuf:"bytes,50,opt,name=label,proto3" json:"label,omitempty"` // free form string
	Tag           string                 `protobuf:"bytes,60,opt,name=tag,proto3" json:"tag,omitempty"`     // free form string
	PollCounter   uint64                 `protobuf:"varint,70,opt,name=poll_counter,json=pollCounter,proto3" json:"poll_counter,omitempty"`
	RecordCounter uint64                 `protobuf:"varint,80,opt,name=record_counter,json=recordCounter,proto3" json:"record_counter,omitempty"`
	Function      string                 `protobuf:"bytes,90,opt,name=function,proto3" json:"function,omitempty"` // []string{"function", "variable", "type"},
	Variable      string                 `protobuf:"bytes,100,opt,name=variable,proto3" json:"variable,omitempty"`
	Type          string                 `protobuf:"bytes,110,opt,name=type,proto3" json:"type,omitempty"`
	Value         float64                `protobuf:"fixed64,120,opt,name=value,proto3" json:"value,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *PromRecordCounter) Reset() {
	*x = PromRecordCounter{}
	mi := &file_prometheus_v1_prometheus_protolist_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PromRecordCounter) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PromRecordCounter) ProtoMessage() {}

func (x *PromRecordCounter) ProtoReflect() protoreflect.Message {
	mi := &file_prometheus_v1_prometheus_protolist_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PromRecordCounter.ProtoReflect.Descriptor instead.
func (*PromRecordCounter) Descriptor() ([]byte, []int) {
	return file_prometheus_v1_prometheus_protolist_proto_rawDescGZIP(), []int{1}
}

func (x *PromRecordCounter) GetTimestampNs() float64 {
	if x != nil {
		return x.TimestampNs
	}
	return 0
}

func (x *PromRecordCounter) GetHostname() string {
	if x != nil {
		return x.Hostname
	}
	return ""
}

func (x *PromRecordCounter) GetPop() string {
	if x != nil {
		return x.Pop
	}
	return ""
}

func (x *PromRecordCounter) GetLabel() string {
	if x != nil {
		return x.Label
	}
	return ""
}

func (x *PromRecordCounter) GetTag() string {
	if x != nil {
		return x.Tag
	}
	return ""
}

func (x *PromRecordCounter) GetPollCounter() uint64 {
	if x != nil {
		return x.PollCounter
	}
	return 0
}

func (x *PromRecordCounter) GetRecordCounter() uint64 {
	if x != nil {
		return x.RecordCounter
	}
	return 0
}

func (x *PromRecordCounter) GetFunction() string {
	if x != nil {
		return x.Function
	}
	return ""
}

func (x *PromRecordCounter) GetVariable() string {
	if x != nil {
		return x.Variable
	}
	return ""
}

func (x *PromRecordCounter) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *PromRecordCounter) GetValue() float64 {
	if x != nil {
		return x.Value
	}
	return 0
}

type Envelope_PromRecordCounter struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	TimestampNs   float64                `protobuf:"fixed64,10,opt,name=timestamp_ns,json=timestampNs,proto3" json:"timestamp_ns,omitempty"`
	Hostname      string                 `protobuf:"bytes,20,opt,name=hostname,proto3" json:"hostname,omitempty"`
	Pop           string                 `protobuf:"bytes,30,opt,name=pop,proto3" json:"pop,omitempty"`
	Label         string                 `protobuf:"bytes,50,opt,name=label,proto3" json:"label,omitempty"` // free form string
	Tag           string                 `protobuf:"bytes,60,opt,name=tag,proto3" json:"tag,omitempty"`     // free form string
	PollCounter   uint64                 `protobuf:"varint,70,opt,name=poll_counter,json=pollCounter,proto3" json:"poll_counter,omitempty"`
	RecordCounter uint64                 `protobuf:"varint,80,opt,name=record_counter,json=recordCounter,proto3" json:"record_counter,omitempty"`
	Function      string                 `protobuf:"bytes,90,opt,name=function,proto3" json:"function,omitempty"` // []string{"function", "variable", "type"},
	Variable      string                 `protobuf:"bytes,100,opt,name=variable,proto3" json:"variable,omitempty"`
	Type          string                 `protobuf:"bytes,110,opt,name=type,proto3" json:"type,omitempty"`
	Value         float64                `protobuf:"fixed64,120,opt,name=value,proto3" json:"value,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Envelope_PromRecordCounter) Reset() {
	*x = Envelope_PromRecordCounter{}
	mi := &file_prometheus_v1_prometheus_protolist_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Envelope_PromRecordCounter) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Envelope_PromRecordCounter) ProtoMessage() {}

func (x *Envelope_PromRecordCounter) ProtoReflect() protoreflect.Message {
	mi := &file_prometheus_v1_prometheus_protolist_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Envelope_PromRecordCounter.ProtoReflect.Descriptor instead.
func (*Envelope_PromRecordCounter) Descriptor() ([]byte, []int) {
	return file_prometheus_v1_prometheus_protolist_proto_rawDescGZIP(), []int{0, 0}
}

func (x *Envelope_PromRecordCounter) GetTimestampNs() float64 {
	if x != nil {
		return x.TimestampNs
	}
	return 0
}

func (x *Envelope_PromRecordCounter) GetHostname() string {
	if x != nil {
		return x.Hostname
	}
	return ""
}

func (x *Envelope_PromRecordCounter) GetPop() string {
	if x != nil {
		return x.Pop
	}
	return ""
}

func (x *Envelope_PromRecordCounter) GetLabel() string {
	if x != nil {
		return x.Label
	}
	return ""
}

func (x *Envelope_PromRecordCounter) GetTag() string {
	if x != nil {
		return x.Tag
	}
	return ""
}

func (x *Envelope_PromRecordCounter) GetPollCounter() uint64 {
	if x != nil {
		return x.PollCounter
	}
	return 0
}

func (x *Envelope_PromRecordCounter) GetRecordCounter() uint64 {
	if x != nil {
		return x.RecordCounter
	}
	return 0
}

func (x *Envelope_PromRecordCounter) GetFunction() string {
	if x != nil {
		return x.Function
	}
	return ""
}

func (x *Envelope_PromRecordCounter) GetVariable() string {
	if x != nil {
		return x.Variable
	}
	return ""
}

func (x *Envelope_PromRecordCounter) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *Envelope_PromRecordCounter) GetValue() float64 {
	if x != nil {
		return x.Value
	}
	return 0
}

var File_prometheus_v1_prometheus_protolist_proto protoreflect.FileDescriptor

const file_prometheus_v1_prometheus_protolist_proto_rawDesc = "" +
	"\n" +
	"(prometheus/v1/prometheus_protolist.proto\x12\rprometheus.v1\"\x84\x03\n" +
	"\bEnvelope\x12=\n" +
	"\x04rows\x18\x01 \x03(\v2).prometheus.v1.Envelope.PromRecordCounterR\x04rows\x1a\xb8\x02\n" +
	"\x11PromRecordCounter\x12!\n" +
	"\ftimestamp_ns\x18\n" +
	" \x01(\x01R\vtimestampNs\x12\x1a\n" +
	"\bhostname\x18\x14 \x01(\tR\bhostname\x12\x10\n" +
	"\x03pop\x18\x1e \x01(\tR\x03pop\x12\x14\n" +
	"\x05label\x182 \x01(\tR\x05label\x12\x10\n" +
	"\x03tag\x18< \x01(\tR\x03tag\x12!\n" +
	"\fpoll_counter\x18F \x01(\x04R\vpollCounter\x12%\n" +
	"\x0erecord_counter\x18P \x01(\x04R\rrecordCounter\x12\x1a\n" +
	"\bfunction\x18Z \x01(\tR\bfunction\x12\x1a\n" +
	"\bvariable\x18d \x01(\tR\bvariable\x12\x12\n" +
	"\x04type\x18n \x01(\tR\x04type\x12\x14\n" +
	"\x05value\x18x \x01(\x01R\x05value\"\xb8\x02\n" +
	"\x11PromRecordCounter\x12!\n" +
	"\ftimestamp_ns\x18\n" +
	" \x01(\x01R\vtimestampNs\x12\x1a\n" +
	"\bhostname\x18\x14 \x01(\tR\bhostname\x12\x10\n" +
	"\x03pop\x18\x1e \x01(\tR\x03pop\x12\x14\n" +
	"\x05label\x182 \x01(\tR\x05label\x12\x10\n" +
	"\x03tag\x18< \x01(\tR\x03tag\x12!\n" +
	"\fpoll_counter\x18F \x01(\x04R\vpollCounter\x12%\n" +
	"\x0erecord_counter\x18P \x01(\x04R\rrecordCounter\x12\x1a\n" +
	"\bfunction\x18Z \x01(\tR\bfunction\x12\x1a\n" +
	"\bvariable\x18d \x01(\tR\bvariable\x12\x12\n" +
	"\x04type\x18n \x01(\tR\x04type\x12\x14\n" +
	"\x05value\x18x \x01(\x01R\x05valueB\x94\x01\n" +
	"\x11com.prometheus.v1B\x18PrometheusProtolistProtoP\x01Z\x10./pkg/prometheus\xa2\x02\x03PXX\xaa\x02\rPrometheus.V1\xca\x02\rPrometheus\\V1\xe2\x02\x19Prometheus\\V1\\GPBMetadata\xea\x02\x0ePrometheus::V1b\x06proto3"

var (
	file_prometheus_v1_prometheus_protolist_proto_rawDescOnce sync.Once
	file_prometheus_v1_prometheus_protolist_proto_rawDescData []byte
)

func file_prometheus_v1_prometheus_protolist_proto_rawDescGZIP() []byte {
	file_prometheus_v1_prometheus_protolist_proto_rawDescOnce.Do(func() {
		file_prometheus_v1_prometheus_protolist_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_prometheus_v1_prometheus_protolist_proto_rawDesc), len(file_prometheus_v1_prometheus_protolist_proto_rawDesc)))
	})
	return file_prometheus_v1_prometheus_protolist_proto_rawDescData
}

var file_prometheus_v1_prometheus_protolist_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_prometheus_v1_prometheus_protolist_proto_goTypes = []any{
	(*Envelope)(nil),                   // 0: prometheus.v1.Envelope
	(*PromRecordCounter)(nil),          // 1: prometheus.v1.PromRecordCounter
	(*Envelope_PromRecordCounter)(nil), // 2: prometheus.v1.Envelope.PromRecordCounter
}
var file_prometheus_v1_prometheus_protolist_proto_depIdxs = []int32{
	2, // 0: prometheus.v1.Envelope.rows:type_name -> prometheus.v1.Envelope.PromRecordCounter
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_prometheus_v1_prometheus_protolist_proto_init() }
func file_prometheus_v1_prometheus_protolist_proto_init() {
	if File_prometheus_v1_prometheus_protolist_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_prometheus_v1_prometheus_protolist_proto_rawDesc), len(file_prometheus_v1_prometheus_protolist_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_prometheus_v1_prometheus_protolist_proto_goTypes,
		DependencyIndexes: file_prometheus_v1_prometheus_protolist_proto_depIdxs,
		MessageInfos:      file_prometheus_v1_prometheus_protolist_proto_msgTypes,
	}.Build()
	File_prometheus_v1_prometheus_protolist_proto = out.File
	file_prometheus_v1_prometheus_protolist_proto_goTypes = nil
	file_prometheus_v1_prometheus_protolist_proto_depIdxs = nil
}
