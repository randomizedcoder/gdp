package gdp

import (
	"context"
	"log"
	"time"

	"github.com/randomizedcoder/gdp/pkg/gdp_config"
	gdpp "github.com/randomizedcoder/gdp/pkg/prometheus"
	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// destNull sends the protobuf to nowhere!
func (g *GDP) destNull(_ context.Context, binaryProto *[]byte, _ *gdp_config.MarshalConfig) (n int, err error) {

	g.pC.WithLabelValues("destNull", "start", "count").Inc()

	return len(*binaryProto), nil
	//return 0, nil
}

// destKafka sends the protobuf to kafka
// https://pkg.go.dev/github.com/twmb/franz-go/pkg/kgo
func (g *GDP) destKafka(ctx context.Context, binaryProto *[]byte, mc *gdp_config.MarshalConfig) (n int, err error) {

	g.pC.WithLabelValues("destKafka", "start", "count").Inc()

	// if g.debugLevel > 10 {
	// 	log.Printf("destKafka, mc.t: %s h: %X", mc.MarshalType, (*binaryProto)[:KafkaHeaderSizeCst])
	// }

	kgoRecord := g.kgoRecordPool.Get().(*kgo.Record)
	// g.pC.WithLabelValues("kgoRecordPool", "Get", "count").Inc()

	kgoRecord.Context = ctx
	kgoRecord.Topic = mc.Topic
	kgoRecord.Value = *binaryProto

	var json string
	if g.debugLevel > 10 {
		json = g.printPayload("binary", binaryProto, mc)
	}

	n = len(kgoRecord.Value)

	if g.debugLevel > 10 {
		json = g.printPayload("value", &kgoRecord.Value, mc)
	}

	if g.debugLevel > 10 {
		log.Printf("destKafka, mc.t:%s, n:%d, j:%s", mc.MarshalType, n, json)
	}

	var (
		ctxP    context.Context
		cancelP context.CancelFunc
	)
	if g.Config.KafkaProduceTimeout.AsDuration() != 0 {
		// I don't understand why setting a context with a timeout doesn't work,
		// but it definitely doesn't.  It always says the context is canceled. ?!
		ctxP, cancelP = context.WithTimeout(ctx, g.Config.KafkaProduceTimeout.AsDuration())
		defer cancelP()
	}
	// https://pkg.go.dev/google.golang.org/protobuf/types/known/durationpb

	kafkaStartTime := time.Now()

	// https://pkg.go.dev/github.com/twmb/franz-go/pkg/kgo#Client.Produce
	g.kClient.Produce(
		ctxP,
		kgoRecord,
		func(kgoRecord *kgo.Record, err error) {

			dur := time.Since(kafkaStartTime)

			g.kgoRecordPool.Put(kgoRecord)
			// g.pC.WithLabelValues("kgoRecordPool", "Put", "count").Inc()

			*binaryProto = (*binaryProto)[:0]
			g.destBytesPool.Put(binaryProto)
			// g.pC.WithLabelValues("destBytesPool", "Put", "count").Inc()

			//cancelP()
			if err != nil {
				g.pH.WithLabelValues("destKafka", mc.Topic, "error").Observe(dur.Seconds())
				g.pC.WithLabelValues("destKafka", mc.Topic, "error").Inc()
				if g.debugLevel > 10 {
					log.Printf("destKafka %0.6fs Produce err:%v", dur.Seconds(), err)
				}
				return
			}

			g.pH.WithLabelValues("destKafka", mc.Topic, "count").Observe(dur.Seconds())
			g.pC.WithLabelValues("destKafka", mc.Topic, "count").Inc()
			g.pC.WithLabelValues("destKafka", mc.Topic, "n").Add(float64(n))

			if g.debugLevel > 10 {
				//formattedTopic := fmt.Sprintf("%31s", mc.Topic)
				log.Printf("destKafka mc.t:%s sc:%d len:%d %0.6fs %dms j:%s", mc.MarshalType, mc.SchemaID, n, dur.Seconds(), dur.Milliseconds(), json)
			}
		},
	)

	return n, err
}

func (g *GDP) printPayload(c string, binaryProto *[]byte, mc *gdp_config.MarshalConfig) string {

	if len(*binaryProto) < KafkaHeaderSizeCst {
		g.pC.WithLabelValues("printPayload", "short", "count").Inc()
		return ""
	}

	if mc.MarshalType == "ProtobufList" {
		return g.printEnvelopePayload(c, binaryProto)
	} else {
		return g.printRowPayload(c, binaryProto)
	}

}

func (g *GDP) printEnvelopePayload(c string, b *[]byte) string {

	var envelope gdpp.Envelope
	err := proto.Unmarshal((*b)[KafkaHeaderSizeCst:], &envelope)
	if err != nil {
		log.Printf("printEnvelopePayload c:%s unmarshal err: %v", c, err)
		return ""
	}

	jsonBytes, err := protojson.Marshal(&envelope)
	if err != nil {
		log.Printf("printEnvelopePayload c:%s marshal err: %v", c, err)
		return ""
	}

	if g.debugLevel > 10 {
		log.Printf("printPayload c:%s envelope JSON: %s", c, jsonBytes)
	}

	// for _, row := range envelope.Rows {
	// 	jsonBytes, err := protojson.Marshal(row)
	// 	if err != nil {
	// 		log.Printf("Failed to marshal row to JSON: %v", err)
	// 		return
	// 	}

	// 	if g.debugLevel > 10 {
	// 		log.Printf("printPayload envelope JSON: %s", jsonBytes)
	// 	}
	// }

	return string(jsonBytes)
}

func (g *GDP) printRowPayload(c string, b *[]byte) string {

	var row gdpp.PromRecordCounter
	err := proto.Unmarshal((*b)[KafkaHeaderSizeCst:], &row)
	if err != nil {
		log.Printf("printRowPayload c:%s unmarshal err: %v", c, err)
		return ""
	}
	jsonBytes, err := protojson.Marshal(&row)
	if err != nil {
		log.Printf("printRowPayload c:%s marshal err: %v", c, err)
		return ""
	}

	if g.debugLevel > 10 {
		log.Printf("printPayload c:%s row JSON: %s", c, jsonBytes)
	}

	return string(jsonBytes)
}

// // destKafkaProto sends the protobuf in protobuf format to kafka
// // this is to test the serdes method of franz-go
// func (g *GDP) destKafkaProto(ctx context.Context, pc *prometheus.Envelope_PromRecordCounter) (n int, err error) {

// 	kgoRecord := g.kgoRecordPool.Get().(*kgo.Record)
// 	// defer g.kgoRecordPool.Put(kgoRecord)

// 	kgoRecord.Topic = "serdeTest"
// 	kgoRecord.Value = g.kSerde.MustEncode(pc)
// 	//kgoRecord.Value = g.kSerde.MustEncode(*xtcpRecordBinary)
// 	len := len(kgoRecord.Value)

// 	var (
// 		ctxP    context.Context
// 		cancelP context.CancelFunc
// 	)
// 	if g.Config.KafkaProduceTimeout.AsDuration() != 0 {
// 		// I don't understand why setting a context with a timeout doesn't work,
// 		// but it definitely doesn't.  It always says the context is canceled. ?!
// 		ctxP, cancelP = context.WithTimeout(ctx, g.Config.KafkaProduceTimeout.AsDuration())
// 		defer cancelP()
// 	}
// 	// https://pkg.go.dev/google.golang.org/protobuf/types/known/durationpb

// 	kafkaStartTime := time.Now()

// 	g.kClient.Produce(
// 		ctxP,
// 		kgoRecord,
// 		func(kgoRecord *kgo.Record, err error) {
// 			dur := time.Since(kafkaStartTime)

// 			g.kgoRecordPool.Put(kgoRecord)

// 			//cancelP()
// 			if err != nil {
// 				g.pH.WithLabelValues("destKafkaProto", "Produce", "error").Observe(dur.Seconds())
// 				g.pC.WithLabelValues("destKafkaProto", "Produce", "error").Inc()
// 				if g.debugLevel > 10 {
// 					log.Printf("destKafkaProto %0.6fs Produce err:%v", dur.Seconds(), err)
// 				}
// 				return
// 			}

// 			g.pH.WithLabelValues("destKafkaProto", "Produce", "count").Observe(dur.Seconds())
// 			g.pC.WithLabelValues("destKafkaProto", "Produce", "count").Inc()

// 			if g.debugLevel > 10 {
// 				log.Printf("destKafkaProto len:%d %0.6fs %dms", len, dur.Seconds(), dur.Milliseconds())
// 			}
// 		},
// 	)

// 	return 1, err
// }
