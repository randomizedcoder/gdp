package gdp

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/randomizedcoder/gdp/pkg/gdp_config"
	"github.com/twmb/franz-go/pkg/kgo"
)

// destNull sends the protobuf to nowhere!
func (g *GDP) destNull(_ context.Context, binaryProto *[]byte, _ int, _ *gdp_config.MarshalConfig) (n int, err error) {

	g.pC.WithLabelValues("destNull", "start", "count").Inc()

	return len(*binaryProto), nil
}

// destKafka sends the protobuf to kafka
func (g *GDP) destKafka(ctx context.Context, binaryProto *[]byte, i int, mc *gdp_config.MarshalConfig) (n int, err error) {

	g.pC.WithLabelValues("destKafka", "start", "count").Inc()

	if g.debugLevel > 10 {
		log.Printf("destKafka i:%d, topic: %s header bytes: % X", i, mc.Topic, (*binaryProto)[:KafkaHeaderSizeCst])
	}

	kgoRecord := g.kgoRecordPool.Get().(*kgo.Record)

	kgoRecord.Topic = mc.Topic // fix me
	kgoRecord.Value = *binaryProto
	len := len(kgoRecord.Value)

	var (
		ctxP    context.Context
		cancelP context.CancelFunc
	)
	if g.config.KafkaProduceTimeout.AsDuration() != 0 {
		// I don't understand why setting a context with a timeout doesn't work,
		// but it definitely doesn't.  It always says the context is canceled. ?!
		ctxP, cancelP = context.WithTimeout(ctx, g.config.KafkaProduceTimeout.AsDuration())
		defer cancelP()
	}
	// https://pkg.go.dev/google.golang.org/protobuf/types/known/durationpb

	kafkaStartTime := time.Now()

	g.kClient.Produce(
		ctxP,
		kgoRecord,
		func(kgoRecord *kgo.Record, err error) {
			dur := time.Since(kafkaStartTime)

			g.kgoRecordPool.Put(kgoRecord)
			*binaryProto = (*binaryProto)[:0]
			g.destBytesPool.Put(binaryProto)

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

			if g.debugLevel > 10 {
				formattedTopic := fmt.Sprintf("%31s", mc.Topic)
				log.Printf("destKafka topic:%s sc:%d len:%d h:%X %0.6fs %dms", formattedTopic, mc.SchemaID, len, (*binaryProto)[:KafkaHeaderSizeCst], dur.Seconds(), dur.Milliseconds())
			}
		},
	)

	// if err := g.kClient.ProduceSync(ctxP, kgoRecord).FirstErr(); err != nil {
	// 	dur := time.Since(kafkaStartTime)
	// 	g.kgoRecordPool.Put(kgoRecord)
	// 	cancelP()
	// 	g.pH.WithLabelValues("destKafka", "ProduceSync", "error").Observe(dur.Seconds())
	// 	g.pC.WithLabelValues("destKafka", "ProduceSync", "error").Inc()
	// 	if g.debugLevel > 10 {
	// 		log.Printf("destKafka %0.6fs ProduceSync err:%v", dur.Seconds(), err)
	// 	}
	// 	return 0, err
	// }
	// g.pH.WithLabelValues("destKafka", "ProduceSync", "count").Observe(time.Since(kafkaStartTime).Seconds())
	// g.pC.WithLabelValues("destKafka", "ProduceSync", "count").Inc()

	// g.kgoRecordPool.Put(kgoRecord)
	// cancelP()

	return 1, err
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
// 	if g.config.KafkaProduceTimeout.AsDuration() != 0 {
// 		// I don't understand why setting a context with a timeout doesn't work,
// 		// but it definitely doesn't.  It always says the context is canceled. ?!
// 		ctxP, cancelP = context.WithTimeout(ctx, g.config.KafkaProduceTimeout.AsDuration())
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
