package gdp

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/randomizedcoder/gdp/pkg/gdp_config"
	gdpp "github.com/randomizedcoder/gdp/pkg/prometheus"

	"github.com/twmb/franz-go/pkg/kgo"
)

const (
	quantileError    = 0.05
	summaryVecMaxAge = 5 * time.Minute

	// For protobuf the size is at least 6, not 5
	// https://docs.confluent.io/platform/current/schema-registry/fundamentals/serdes-develop/indeg.html#wire-format
	KafkaHeaderSizeCst = 6
)

type GDP struct {
	ctx    context.Context
	cancel context.CancelFunc

	Config *gdp_config.GdpConfig

	MarshalConfigs *sync.Map // gdp_config.MarshalConfig

	GDPEnvelopePool sync.Pool
	GDPRecordPool   sync.Pool
	kgoRecordPool   sync.Pool
	destBytesPool   sync.Pool

	hostname string

	Marshallers  sync.Map
	Marshaller   func(pc *gdpp.Envelope_PromRecordCounter) (buf *[]byte)
	Destinations sync.Map
	Destination  func(ctx context.Context, binaryProto *[]byte, mc *gdp_config.MarshalConfig) (n int, err error)

	InitDestinations sync.Map

	DestinationReady chan struct{} // Signals poller can start

	kClient *kgo.Client
	// kRegClient *sr.Client
	// kSerde     sr.Serde
	schemaID int

	pC *prometheus.CounterVec
	pH *prometheus.SummaryVec
	pG prometheus.Gauge

	debugLevel uint32
}

func NewGDP(ctx context.Context, cancel context.CancelFunc, config *gdp_config.GdpConfig) *GDP {

	g := new(GDP)

	g.ctx = ctx
	g.cancel = cancel

	g.Config = config
	g.debugLevel = g.Config.DebugLevel

	g.Init(ctx)

	return g
}

// newGDPforTests inits just enough for the parse tests to run
func newGDPforTests() *GDP {

	g := new(GDP)

	g.debugLevel = 11

	g.GDPRecordPool = sync.Pool{
		New: func() any {
			return new(gdpp.Envelope_PromRecordCounter)
		},
	}

	return g
}

// RunWithPoller is the main run function for GDP
// it starts everything required, including the netlink socket poller
func (g *GDP) RunWithPoller(ctx context.Context, wg *sync.WaitGroup) {
	g.Run(ctx, wg, true)
}

// RunNoPoller is only for testing, do not run this for real
// This will only monitor the name spaces, and was used for
// testing the kernel/user-land correctly stay in sync, there's
// no leaks, etc.  See also /cmd/nsTest/nsTest.go
// Basically, it just doesn't start the poller
func (g *GDP) RunNoPoller(ctx context.Context, wg *sync.WaitGroup) {
	g.Run(ctx, wg, false)
}

func (g *GDP) Run(ctx context.Context, wg *sync.WaitGroup, runPoller bool) {

	defer wg.Done()

	g.pC.WithLabelValues("Run", "start", "count").Inc()

	if runPoller {
		wg.Add(1)
		go g.Poller(ctx, wg)

		if g.debugLevel > 10 {
			log.Println("GDP.Run() wg.Wait()")
		}
	}

	if g.debugLevel > 10 {
		log.Println("GDP.Run() wg.Wait()")
	}
	wg.Wait()

	g.cancel()

	if g.debugLevel > 10 {
		log.Println("GDP.Run() g.cancel()")
	}

	if g.debugLevel > 10 {
		log.Println("GDP.Run() complete")
	}
}

func (g *GDP) closeDestination() {
	switch g.Config.Dest {
	case "kafka":
		g.kClient.Close()
		// case "udp":
		// 	g.udpConn.Close()
	}
}
