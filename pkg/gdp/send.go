package gdp

import (
	"context"
	"log"
	"sync"

	"github.com/randomizedcoder/gdp/pkg/gdp_config"
	"github.com/randomizedcoder/gdp/pkg/prometheus"
)

func (g *GDP) sendEnvelopeWithMarshalConfig(
	ctx context.Context,
	wg *sync.WaitGroup,
	mc *gdp_config.MarshalConfig,
	envelope *prometheus.Envelope,
	i int) error {

	defer wg.Done()

	if g.debugLevel > 10 {
		log.Println("sendEnvelopeWithMarshalConfig start")
	}

	// ProtobufList is essentially the Envelope type
	if mc.MarshalType == "ProtobufList" {
		b := g.marshal(envelope, mc, i)
		if b == nil {
			return nil
		}
		_, err := g.Destination(ctx, b, i, mc)
		if err != nil {
			return err
		}
		g.destBytesPool.Put(b)
		return nil
	}

	// Otherwise, protobuf, so we need to iterate
	for _, pc := range envelope.Rows {

		b := g.marshal(pc, mc, i)
		if b == nil {
			return nil
		}

		_, err := g.Destination(ctx, b, i, mc)
		if err != nil {
			return err
		}
		g.destBytesPool.Put(b)
	}

	return nil
}
