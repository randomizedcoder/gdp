package gdp

import (
	"encoding/binary"
	"log"

	"github.com/randomizedcoder/gdp/pkg/gdp_config"
	gdpp "github.com/randomizedcoder/gdp/pkg/prometheus"
	"google.golang.org/protobuf/encoding/protodelim"
	"google.golang.org/protobuf/proto"
)

// var (
// 	validMarshallersMap = map[string]bool{
// 		// "protobuf":       true, // https://clickhouse.com/docs/en/interfaces/formats/Protobuf
// 		// "protobufSingle": true, // https://clickhouse.com/docs/en/interfaces/formats/ProtobufSingle
// 		"protobufList": true, // https://clickhouse.com/docs/en/interfaces/formats/ProtobufList
// 	}
// )

// func validMarshallers() (marshallers string) {
// 	for key := range validMarshallersMap {
// 		marshallers = marshallers + key + ","
// 	}
// 	return strings.TrimSuffix(marshallers, ",")
// }

type ByteSliceWriter struct {
	Buf *[]byte
}

func (w *ByteSliceWriter) Write(b []byte) (n int, err error) {
	*w.Buf = append(*w.Buf, b...)
	return len(b), nil
}

//pc *gdpp.Envelope_PromRecordCounter
//*gdpp.Envelope

func (g *GDP) marshal(proto any, mc *gdp_config.MarshalConfig) (buf *[]byte) {

	buf = g.destBytesPool.Get().(*[]byte)
	// g.pC.WithLabelValues("destBytesPool", "Get", "count").Inc()

	// if g.debugLevel > 10 {
	// 	log.Printf("marshal , mc.t: %v, h:%X", mc.MarshalType, (*buf)[:KafkaHeaderSizeCst])
	// }

	if mc.KafkaHeader {
		// Add the Confluent header for protobuf, which is length 6
		// https://docs.confluent.io/platform/current/schema-registry/fundamentals/serdes-develop/indeg.html#wire-format
		*buf = (*buf)[:KafkaHeaderSizeCst]
		(*buf)[0] = 0x00                                    // Magic byte
		binary.BigEndian.PutUint32((*buf)[1:], mc.SchemaID) // Sc
		(*buf)[5] = 0x00                                    // the first message

		// most of the time the actual message type will be just the first message
		// type (which is the array [0]), which would normally be encoded as 1,0 (1
		// for length), this special case is optimized to just 0. So in the most common
		//  case of the first message type being used, a single 0 is encoded as
		// the message-indexes.

		// if g.debugLevel > 1000 {
		// 	log.Printf("marshal protodelim.MarshalTo() mc.SchemaID:%d mc.SchemaID hex:%x", mc.SchemaID, mc.SchemaID)
		// 	log.Printf("marshal header bytes: % X", (*buf)[:KafkaHeaderSizeCst])
		// }
	}

	if mc.MarshalType == "ProtobufList" {

		//(*buf)[5] = 0x00 // the first message
		// if g.debugLevel > 10 {
		// 	log.Printf("marshal set, mc.t: %v, h:%X", mc.MarshalType, (*buf)[:KafkaHeaderSizeCst])
		// }

		e, ok := proto.(*gdpp.Envelope)
		if !ok {
			g.pC.WithLabelValues("marshal", "ProtobufList_type_assertion", "error").Inc()
			if g.debugLevel > 10 {
				log.Printf("marshal: type assertion failed: expected *gdpp.Envelope, got %T", proto)
			}
			g.destBytesPool.Put(buf)
			// g.pC.WithLabelValues("destBytesPool", "Put", "count").Inc()
			return nil
		}
		g.protobufListMarshal(e, mc, buf)

		// if g.debugLevel > 1000 {
		// 	log.Printf("marshal ProtobufList topic:%s len(buf):%d", mc.Topic, len(*buf))
		// }

		// if g.debugLevel > 10 {
		// 	log.Printf("marshal end, mc.t: %v, h:%X", mc.MarshalType, (*buf)[:KafkaHeaderSizeCst])
		// }

		return buf
	}

	//(*buf)[5] = 0x01 // the second message
	// if g.debugLevel > 10 {
	// 	log.Printf("marshal set, mc.t: %v, h:%X", mc.MarshalType, (*buf)[:KafkaHeaderSizeCst])
	// }

	pc, ok := proto.(*gdpp.Envelope_PromRecordCounter)
	if !ok {
		g.pC.WithLabelValues("marshal", "type_assertion", "error").Inc()
		if g.debugLevel > 10 {
			log.Printf("marshal: type assertion failed: expected *gdpp.Envelope_PromRecordCounter, got %T", proto)
		}
		g.destBytesPool.Put(buf)
		// g.pC.WithLabelValues("destBytesPool", "Put", "count").Inc()
		return nil
	}
	g.protobufMarshal(pc, mc, buf)

	// if g.debugLevel > 1000 {
	// 	log.Printf("marshal protobufMarshal topic:%s len(buf):%d", mc.Topic, len(*buf))
	// }

	// if g.debugLevel > 10 {
	// 	log.Printf("marshal end, mc.t: %v, h:%X", mc.MarshalType, (*buf)[:KafkaHeaderSizeCst])
	// }

	return buf
}

// protobufSingleMarshal marshals to protobuf and does error handling
// https://clickhouse.com/docs/en/interfaces/formats#protobufsingle
// https://pkg.go.dev/google.golang.org/protobuf/proto?tab=doc#Marshal
func (g *GDP) protobufMarshal(pc *gdpp.Envelope_PromRecordCounter, mc *gdp_config.MarshalConfig, buf *[]byte) {

	if g.debugLevel > 10 {
		log.Printf("protobufMarshal mc: %v", mc)
	}

	if mc.Protodelim {
		// writer will append from end of buf
		writer := &ByteSliceWriter{Buf: buf}

		// https://pkg.go.dev/google.golang.org/protobuf@v1.36.3/encoding/protodelim#MarshalTo
		n, err := protodelim.MarshalTo(writer, pc)
		if err != nil {
			g.pC.WithLabelValues("protobufMarshal", "MarshalTo", "error").Inc()
			if g.debugLevel > 10 {
				log.Println("protodelim.MarshalTo() err: ", err)
			}
		}

		if g.debugLevel > 10 {
			log.Printf("protobufMarshal: After MarshalTo, n: %d, len(*buf): %d, *buf: % X", n, len(*buf), (*buf)[:KafkaHeaderSizeCst])
			//log.Printf("protobufListMarshal: After MarshalTo, len(writer.Buf): %d, writer.Buf: % X", len(*writer.Buf), *writer.Buf)
			log.Printf("protobufMarshal protodelim.MarshalTo() n:%d", n)
		}
	} else {

		// https://pkg.go.dev/google.golang.org/protobuf/proto?tab=doc#Marshal
		b, err := proto.Marshal(pc)
		if err != nil {
			g.pC.WithLabelValues("protobufMarshal", "Marshal", "error").Inc()
			if g.debugLevel > 10 {
				log.Println("proto.Marshal(pc) err: ", err)
			}
		}
		*buf = append(*buf, b...)
	}
}

// protobufListMarshal marshals the protobuf to binary, and includes
// the confluent header
func (g *GDP) protobufListMarshal(e *gdpp.Envelope, mc *gdp_config.MarshalConfig, buf *[]byte) {

	if g.debugLevel > 10 {
		log.Printf("protobufListMarshal mc: %v", mc)
	}

	if mc.Protodelim {
		// writer will append from end of buf
		writer := &ByteSliceWriter{Buf: buf}

		// https://pkg.go.dev/google.golang.org/protobuf@v1.36.3/encoding/protodelim#MarshalTo
		n, err := protodelim.MarshalTo(writer, e)
		if err != nil {
			g.pC.WithLabelValues("protobufListMarshal", "MarshalTo", "error").Inc()
			if g.debugLevel > 10 {
				log.Println("protodelim.MarshalTo() err: ", err)
			}
		}

		if g.debugLevel > 10 {
			log.Printf("protobufListMarshal: After MarshalTo, n: %d, len(*buf): %d, *buf: % X", n, len(*buf), (*buf)[:KafkaHeaderSizeCst])
			//log.Printf("protobufListMarshal: After MarshalTo, len(writer.Buf): %d, writer.Buf: % X", len(*writer.Buf), *writer.Buf)
			log.Printf("protobufListMarshal protodelim.MarshalTo() n:%d", n)
		}
	} else {
		// https://pkg.go.dev/google.golang.org/protobuf/proto?tab=doc#Marshal
		b, err := proto.Marshal(e)
		if err != nil {
			g.pC.WithLabelValues("protobufListMarshal", "Marshal", "error").Inc()
			if g.debugLevel > 10 {
				log.Println("proto.Marshal(e) err: ", err)
			}
		}
		*buf = append(*buf, b...)
	}
}
