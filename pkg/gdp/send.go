package gdp

import (
	"context"
	"sync"
	"time"

	"github.com/randomizedcoder/gdp/pkg/gdp_config"
	"github.com/randomizedcoder/gdp/pkg/prometheus"
)

func (g *GDP) sendEnvelopeWithMarshalConfig(
	ctx context.Context,
	wg *sync.WaitGroup,
	pollingLoops uint64,
	mc *gdp_config.MarshalConfig,
	envelope *prometheus.Envelope) error {

	defer wg.Done()

	startTime := time.Now()
	defer func() {
		g.pH.WithLabelValues("sendEnvelopeWithMarshalConfig", "complete", "count").Observe(time.Since(startTime).Seconds())
	}()
	g.pC.WithLabelValues("sendEnvelopeWithMarshalConfig", "start", "count").Inc()

	// if g.debugLevel > 10 {
	// 	log.Println("sendEnvelopeWithMarshalConfig start")
	// }

	// ProtobufList is essentially the Envelope type
	if mc.MarshalType == "ProtobufList" {
		b := g.marshal(envelope, mc)
		if b == nil {
			return nil
		}
		// if pollingLoops == 10 {
		// 	filename := "envelope_binary_loop_10.bin"
		// 	err := os.WriteFile(filename, *b, 0644)
		// 	if err != nil {
		// 		log.Printf("sendEnvelopeWithMarshalConfig writeBinaryEnvelope error: %v", err)
		// 	} else {
		// 		log.Printf("sendEnvelopeWithMarshalConfig writeBinaryEnvelope complete: %s", filename)
		// 	}
		// 	jsonBytes, err := protojson.Marshal(envelope)
		// 	if err != nil {
		// 		log.Printf("performPoll marshalEnvelope error: %v", err)
		// 	} else {
		// 		err := os.WriteFile("xenvelope_loop_10.json", jsonBytes, 0644)
		// 		if err != nil {
		// 			log.Printf("performPoll writeEnvelope error: %v", err)
		// 		}
		// 	}
		// }
		_, err := g.Destination(ctx, b, mc)
		if err != nil {
			return err
		}
		return nil
	}

	// Otherwise, protobuf, so we need to iterate
	for _, pc := range envelope.Rows {

		b := g.marshal(pc, mc)
		if b == nil {
			return nil
		}
		// if pollingLoops == 10 {
		// 	filename := fmt.Sprintf("record_binary_loop_10_%d.bin", i)
		// 	err := os.WriteFile(filename, *b, 0644)
		// 	if err != nil {
		// 		log.Printf("sendEnvelopeWithMarshalConfig writeBinaryRecord error: %v", err)
		// 	} else {
		// 		log.Printf("sendEnvelopeWithMarshalConfig writeBinaryRecord complete: %s", filename)
		// 	}
		// 	jsonBytes, err := protojson.Marshal(envelope)
		// 	if err != nil {
		// 		log.Printf("performPoll marshalEnvelope error: %v", err)
		// 	} else {
		// 		err := os.WriteFile("xenvelope_loop_10.json", jsonBytes, 0644)
		// 		if err != nil {
		// 			log.Printf("performPoll writeEnvelope error: %v", err)
		// 		}
		// 	}
		// }
		_, err := g.Destination(ctx, b, mc)
		if err != nil {
			return err
		}
		// if g.debugLevel > 10 {
		// 	log.Printf("sendEnvelopeWithMarshalConfig i:%d", i)
		// }
		// if i == 1 {
		// 	return nil
		// }
	}

	return nil
}
