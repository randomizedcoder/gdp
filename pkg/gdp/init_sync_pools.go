package gdp

import (
	"sync"

	gdpp "github.com/randomizedcoder/gdp/pkg/prometheus"
	"github.com/twmb/franz-go/pkg/kgo"
)

const (
	destBytesMaxSizeCst = 10000
)

func (g *GDP) InitSyncPools(wg *sync.WaitGroup) {

	defer wg.Done()

	g.GDPEnvelopePool = sync.Pool{
		New: func() any {
			return new(gdpp.Envelope)
		},
	}

	g.GDPRecordPool = sync.Pool{
		New: func() any {
			return new(gdpp.Envelope_PromRecordCounter)
		},
	}

	g.kgoRecordPool = sync.Pool{
		New: func() any {
			//return new(kgo.Record)
			k := new(kgo.Record)
			k.Value = make([]byte, 0, destBytesMaxSizeCst)
			return k
		},
	}

	g.destBytesPool = sync.Pool{
		New: func() any {
			b := make([]byte, 0, destBytesMaxSizeCst)
			return &b
		},
	}
}
