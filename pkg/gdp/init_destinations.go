package gdp

import (
	"context"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/randomizedcoder/gdp/pkg/gdp_config"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/plugin/kprom"
)

const (
	kafkaPingTimeoutCst    = 5 * time.Second
	kafkaPingRetriesCst    = 5
	kafkaPingRetrySleepCst = 1 * time.Second
	//kafkaClientProduceTimeoutCst = 100 * time.Millisecond

	defaultTopicCst = "gdp_default_topic"
)

var (
	validDestinationsMap = map[string]bool{
		"null":  true,
		"kafka": true,
	}
)

func validDestinations() (dests string) {
	for key := range validDestinationsMap {
		dests = dests + key + ","
	}
	return strings.TrimSuffix(dests, ",")
}

// InitDestinations parses the destination config, to determine
// which of the destination modules to use
// This function is using sync.Maps as the way to map to the actual
// function to be used
// This will in future allow the destination to be changed dyanmically
// at runtime (TODO implement this)
func (g *GDP) InitDests(ctx context.Context, wg *sync.WaitGroup) {

	defer wg.Done()

	dest, _, _ := strings.Cut(g.config.Dest, ":")
	if _, ok := validDestinationsMap[dest]; !ok {
		log.Fatalf("InitDests GDP Dest invalid:%s, must be one of:%s", dest, validDestinations())
	}

	g.Destinations.Store("null", func(ctx context.Context, binaryProto *[]byte, i int, mc *gdp_config.MarshalConfig) (n int, err error) {
		return g.destNull(ctx, binaryProto, i, mc)
	})
	g.Destinations.Store("kafka", func(ctx context.Context, binaryProto *[]byte, i int, mc *gdp_config.MarshalConfig) (n int, err error) {
		return g.destKafka(ctx, binaryProto, i, mc)
	})

	f, ok := g.Destinations.Load(dest)
	if !ok {
		log.Fatalf("InitDests GDP Dest load invalid:%s, must be one of:%s", dest, validDestinations())
	}
	g.Destination = f.(func(ctx context.Context, GDPRecordBinary *[]byte, i int, mc *gdp_config.MarshalConfig) (n int, err error))

	// no null initilizer
	g.InitDestinations.Store("kafka", func(ctx context.Context) {
		g.InitDestKafka(ctx)
	})

	if f, ok := g.InitDestinations.Load(dest); ok {
		f.(func(ctx context.Context))(ctx)
	}

	// please note that at this point we _could_ remove any destinations from the map
	// that we aren't using

	g.DestinationReady <- struct{}{}

}

// InitDestKafka creates the franz-go kafka client
func (g *GDP) InitDestKafka(ctx context.Context) {

	// https://github.com/twmb/franz-go/tree/master/plugin/kprom
	kgoMetrics := kprom.NewMetrics("kgo")

	// initialize the kafka client
	// https://pkg.go.dev/github.com/twmb/franz-go/pkg/kgo#ProducerOpt

	broker := strings.Replace(g.config.Dest, "kafka:", "", 1)

	if g.debugLevel > 10 {
		log.Println("config.Dest:", g.config.Dest)
		log.Println("config.KafkaSchemaUrl:", g.config.KafkaSchemaUrl)
		log.Println("broker:", broker)
		log.Println("g.schemaID:", g.schemaID)
	}

	opts := []kgo.Opt{
		// https://pkg.go.dev/github.com/twmb/franz-go/pkg/kgo#DefaultProduceTopic
		kgo.DefaultProduceTopic(defaultTopicCst),
		kgo.ClientID("gdp"),

		// https://pkg.go.dev/github.com/twmb/franz-go/pkg/kgo#SeedBrokers
		kgo.SeedBrokers(broker),

		// https://pkg.go.dev/github.com/twmb/franz-go/pkg/kgo#ProducerBatchCompression
		// https://pkg.go.dev/github.com/twmb/franz-go/pkg/kgo#Lz4Compression
		kgo.ProducerBatchCompression(
			kgo.ZstdCompression(),
			kgo.Lz4Compression(),
			kgo.SnappyCompression(),
			kgo.NoCompression(),
		),

		// https://pkg.go.dev/github.com/twmb/franz-go/pkg/kgo#AllowAutoTopicCreation
		kgo.AllowAutoTopicCreation(),

		// https://pkg.go.dev/github.com/twmb/franz-go/pkg/kgo#WithHooks
		kgo.WithHooks(kgoMetrics),

		// https://pkg.go.dev/github.com/twmb/franz-go/pkg/kgo#MaxBufferedRecords
		// kgo.MaxBufferedRecords(250<<20 / *recordBytes + 1), // default is 10k records
		// https://pkg.go.dev/github.com/twmb/franz-go/pkg/kgo#ProducerBatchMaxBytes
		// kgo.ProducerBatchMaxBytes(int32(*batchMaxBytes)),  // default is ~1MB
		// https://pkg.go.dev/github.com/twmb/franz-go/pkg/kgo#RetryBackoffFn
		// default jittery exponential backoff that ranges from 250ms min to 2.5s max

		// https://pkg.go.dev/github.com/twmb/franz-go/pkg/kgo#DisableIdempotentWrite
		kgo.DisableIdempotentWrite(),

		// https://pkg.go.dev/github.com/twmb/franz-go/pkg/kgo#BrokerMaxWriteBytes
		// maxBrokerWriteBytes: 100 << 20,
		// Kafka socket.request.mag.bytes default is 100<<20 = 104857600 = 100 MB
		// https://github.com/twmb/franz-go/blob/v1.17.1/pkg/kgo/config.go#L483C3-L483C87
		// https://www.wolframalpha.com/input?i=1+%3C%3C+10
		//kgo.BrokerMaxWriteBytes(100 << 18), // 26214400 = 26 MB
		kgo.BrokerMaxWriteBytes(1 << 20), // 1048576 = 1 MB

		// https://pkg.go.dev/github.com/twmb/franz-go/pkg/kgo#ProducerBatchMaxBytes
		// Copied from the benchmark
		// https://github.com/twmb/franz-go/blob/master/examples/bench/main.go#L104
		kgo.ProducerBatchMaxBytes(1000000), // 1 MB

	}

	if g.config.KafkaDebugLevel > 10 {
		// Debugging in the kgo client
		opts = append(opts, kgo.WithLogger(kgo.BasicLogger(os.Stderr, kgo.LogLevelDebug, func() string {
			return time.Now().Format("[2006-01-02 15:04:05.999] ")
		})))
	}

	var errK error
	g.kClient, errK = kgo.NewClient(opts...)
	if errK != nil {
		log.Fatalf("unable to create client:%v", errK)
	}

	// https://pkg.go.dev/github.com/twmb/franz-go/pkg/kgo#Client.AllowRebalance
	g.kClient.AllowRebalance()

	errP := g.pingKafkaWithRetries(ctx, kafkaPingRetriesCst, kafkaPingRetrySleepCst)
	if errP != nil {
		log.Fatalf("InitDestKafka pingKafkaWithRetries errP:%v", errP)
	}

}
