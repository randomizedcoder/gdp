package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/bufbuild/protovalidate-go"
	"github.com/pkg/profile"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/randomizedcoder/gdp/pkg/gdp"
	"github.com/randomizedcoder/gdp/pkg/gdp_config"
	"google.golang.org/protobuf/types/known/durationpb"
)

const (
	debugLevelCst = 111

	kafkaDebugLevelCst = 1

	pollFrequencyCst = 60 * time.Second
	pollTimeoutCst   = 5 * time.Second

	signalChannelSizeCst = 10
	cancelSleepTimeCst   = 5 * time.Second

	promListenCst           = ":8888" // [::1]:8888
	promPathCst             = "/metrics"
	promMaxRequestsInFlight = 10
	promEnableOpenMetrics   = true

	// Redpanda
	destCst = "kafka:redpanda-0:9092"

	//topicCst = "gdp"

	// relative to the container
	promProtoFileCst     = "/prometheus.proto"
	promListProtoFileCst = "/prometheus_protolist.proto"
	kafkaSchemaUrlCst    = "http://redpanda-0:18081"

	// setting produce timeout cause the context to be cancelled(). Change this to see the curious behavior with the kgo ctx
	kafkaProduceTimeoutCst = 0 * time.Second

	labelCst = ""
	tagCst   = ""

	DestWriteFilesCst = 0

	// startSleepCst = 10 * time.Second

	base10    = 10
	sixtyFour = 64
)

var (
	// Passed by "go build -ldflags" for the show version
	commit  string
	date    string
	version string

	debugLevel uint
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	complete := make(chan struct{}, signalChannelSizeCst)
	go initSignalHandler(cancel, complete)

	pollFrequency := flag.Duration("frequency", pollFrequencyCst, "Poll frequency")
	pollTimeout := flag.Duration("timeout", pollTimeoutCst, "Poll timeout per name space")
	maxLoops := flag.Uint64("maxLoops", 0, "Maximum number of loops, or zero (0) for forever")

	dest := flag.String("dest", destCst, "kafka:127.0.0.1:9092, udp:127.0.0.1:13000, or nsq:127.0.0.1:4150")
	destWriteFiles := flag.Uint("destWriteFiles", DestWriteFilesCst, "Write out the marshalled data to destWriteFiles number of files ( for debugging only )")

	promProtoFile := flag.String("promProtoFile", promProtoFileCst, "promProtoFile for registering with the schema registry")
	promListProtoFile := flag.String("promListProtoFile", promListProtoFileCst, "promListProtoFileCst for registering with the schema registry")

	kafkaSchemaUrl := flag.String("kafkaSchemaUrl", kafkaSchemaUrlCst, "kafka schema registry URL")
	produceTimeout := flag.Duration("produceTimeout", kafkaProduceTimeoutCst, "Kafka produce timeout (context.WithTimeout)")

	label := flag.String("label", labelCst, "label applied to the protobuf")
	tag := flag.String("tag", tagCst, "label applied to the protobuf")

	promListen := flag.String("promListen", promListenCst, "Prometheus http listening socket")
	promPath := flag.String("promPath", promPathCst, "Prometheus http path")
	// curl -s http://[::1]:9000/metrics 2>&1 | grep -v "#"
	// curl -s http://127.0.0.1:9000/metrics 2>&1 | grep -v "#"

	// Maximum number of CPUs that can be executing simultaneously
	// https://golang.org/pkg/runtime/#GOMAXPROCS -> zero (0) means default
	goMaxProcs := flag.Uint("goMaxProcs", 4, "goMaxProcs = https://golang.org/pkg/runtime/#GOMAXPROCS")

	// ./gdp --profile.mode cpu
	// timeout 1h ./gdp --profile.mode cpu
	profileMode := flag.String("profile.mode", "", "enable profiling mode, one of [cpu, mem, mutex, block]")

	v := flag.Bool("v", false, "show version")

	conf := flag.Bool("conf", false, "show config")

	d := flag.Uint("d", debugLevelCst, "debug level")
	kd := flag.Uint("kd", kafkaDebugLevelCst, "kafka debug level")

	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.LUTC | log.Lshortfile | log.Lmsgprefix)

	// Print version information passed in via ldflags in the Makefile
	if *v {
		log.Printf("gdp commit:%s\tdate(UTC):%s\tversion:%s", commit, date, version)
		os.Exit(0)
	}

	environmentOverrideDebugLevel(d, *d)
	environmentOverrideKafkaDebugLevel(kd, *kd)

	debugLevel = *d

	if debugLevel > 10 {
		fmt.Println("*pollFrequency:", *pollFrequency)
		fmt.Println("*pollTimeout:", *pollTimeout)
		fmt.Println("*maxLoops:", *maxLoops)
		fmt.Println("*dest:", *dest)
		fmt.Println("*destWriteFiles:", *destWriteFiles)
		fmt.Println("*promProtoFile:", *promProtoFile)
		fmt.Println("*promListProtoFile:", *promListProtoFile)
		fmt.Println("*kafkaSchemaUrl:", *kafkaSchemaUrl)
		fmt.Println("*produceTimeout:", *produceTimeout)
		fmt.Println("*tag:", *tag)
		fmt.Println("*label:", *label)
		fmt.Println("*promListen:", *promListen)
		fmt.Println("*promPath:", *promPath)
		fmt.Println("*goMaxProcs:", *goMaxProcs)
		fmt.Println("*d:", *d)
		fmt.Println("*kd:", *kd)
	}

	c := &gdp_config.GdpConfig{
		// https://pkg.go.dev/google.golang.org/protobuf/types/known/durationpb
		PollFrequency:       durationpb.New(*pollFrequency),
		PollTimeout:         durationpb.New(*pollTimeout),
		MaxLoops:            *maxLoops,
		Dest:                *dest,
		DestWriteFiles:      uint32(*destWriteFiles),
		PromProtoFile:       *promProtoFile,
		PromListProtoFile:   *promListProtoFile,
		KafkaSchemaUrl:      *kafkaSchemaUrl,
		KafkaProduceTimeout: durationpb.New(*produceTimeout),
		DebugLevel:          uint32(*d),
		KafkaDebugLevel:     uint32(*kd),
		Label:               *label,
		Tag:                 *tag,
		RequestConfig: &gdp_config.GdpConfig_PromethusRequestConfig{
			MetricsUrl:     "http://localhost" + *promListen + *promPath,
			AllowPrefix:    "gdp_count",
			MaxRetries:     3,
			RequestBackoff: durationpb.New(100 * time.Millisecond),
		},
	}

	if debugLevel > 100 {
		printConfig(c, "Before environmentOverrideConfig")
	}

	environmentOverrideConfig(c, debugLevel)

	if debugLevel > 100 {
		printConfig(c, "After environmentOverrideConfig")
	}

	if *conf {
		printConfig(c, "conf argument")
		os.Exit(0)
	}

	environmentOverrideGoMaxProcs(goMaxProcs, debugLevel)
	if runtime.NumCPU() > int(*goMaxProcs) {
		mp := runtime.GOMAXPROCS(int(*goMaxProcs))
		if debugLevel > 10 {
			log.Printf("Main runtime.GOMAXPROCS now:%d was:%d\n", *goMaxProcs, mp)
		}
	}

	// "github.com/pkg/profile"
	// https://dave.cheney.net/2013/07/07/introducing-profile-super-simple-profiling-for-go-programs
	// e.g. ./gdp -profile.mode trace
	// go tool trace trace.out
	// e.g. ./gdp -profile.mode cpu
	// go tool pprof -http=":8081" gdp cpu.pprof

	if debugLevel > 10 {
		log.Println("*profileMode:", *profileMode)
	}
	switch *profileMode {
	case "cpu":
		defer profile.Start(profile.CPUProfile, profile.ProfilePath(".")).Stop()
	case "mem":
		defer profile.Start(profile.MemProfile, profile.ProfilePath(".")).Stop()
	case "memheap":
		defer profile.Start(profile.MemProfileHeap, profile.ProfilePath(".")).Stop()
	case "mutex":
		defer profile.Start(profile.MutexProfile, profile.ProfilePath(".")).Stop()
	case "block":
		defer profile.Start(profile.BlockProfile, profile.ProfilePath(".")).Stop()
	case "trace":
		defer profile.Start(profile.TraceProfile, profile.ProfilePath(".")).Stop()
	case "goroutine":
		defer profile.Start(profile.GoroutineProfile, profile.ProfilePath(".")).Stop()
	default:
		if debugLevel > 1000 {
			log.Println("No profiling")
		}
	}

	environmentOverrideProm(promPath, promListen, debugLevel)
	go initPromHandler(*promPath, *promListen)
	if debugLevel > 10 {
		log.Println("Prometheus http listener started on:", *promListen, *promPath)
	}

	if err := protovalidate.Validate(c); err != nil {
		log.Fatal("config validation failed:", err)
	}

	if debugLevel > 10 {
		log.Println("config validation succeeded")
	}

	gdp := gdp.NewGDP(ctx, cancel, c)

	if debugLevel > 10 {
		log.Println("gdp.Run(ctx, &wg)")
	}

	var wg sync.WaitGroup
	wg.Add(1)
	gdp.RunWithPoller(ctx, &wg)

	if debugLevel > 10 {
		log.Println("gdp.RunWithPoller(ctx) complete. wg.Wait()")
	}

	wg.Wait()
	complete <- struct{}{}

	if debugLevel > 10 {
		log.Println("gdp.go Main complete - farewell")
	}
}

// initSignalHandler sets up signal handling for the process, and
// will call cancel() when received
func initSignalHandler(cancel context.CancelFunc, complete <-chan struct{}) {

	c := make(chan os.Signal, signalChannelSizeCst)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c
	log.Printf("Signal caught, closing application")
	cancel()

	log.Printf("Signal caught, cancel() called, and sleeping to allow goroutines to close, sleeping:%s",
		cancelSleepTimeCst.String())
	timer := time.NewTimer(cancelSleepTimeCst)

	select {
	case <-complete:
		log.Printf("<-complete exit(0)")
	case <-timer.C:
		// if we exit here, this means all the other go routines didn't shutdown
		// need to investigate why
		log.Printf("Sleep complete, goodbye! exit(0)")
	}

	os.Exit(0)
}

// initPromHandler starts the prom handler with error checking
// https://pkg.go.dev/github.com/prometheus/client_golang/prometheus/promhttp?tab=doc#HandlerOpts
func initPromHandler(promPath string, promListen string) {
	http.Handle(promPath, promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{
			EnableOpenMetrics:   promEnableOpenMetrics,
			MaxRequestsInFlight: promMaxRequestsInFlight,
		},
	))
	go func() {
		err := http.ListenAndServe(promListen, nil)
		if err != nil {
			log.Fatal("prometheus error", err)
		}
	}()
}

// environmentOverrideProm MUTATES promListen, promPath, if the environment
// variables exist.  This allows over riding the cli flags
//
//lint:ignore SA4009 this is nasty, but it's going to be ok
func environmentOverrideProm(promListen, promPath *string, debugLevel uint) {
	key := "PROM_LISTEN"
	if value, exists := os.LookupEnv(key); exists {
		promListen = &value
		if debugLevel > 10 {
			log.Printf("key:%s, c.PromListen:%s", key, *promListen)
		}
	}

	key = "PROM_PATH"
	if value, exists := os.LookupEnv(key); exists {
		promPath = &value
		if debugLevel > 10 {
			log.Printf("key:%s, c.PromListen:%s", key, *promPath)
		}
	}
}

// environmentOverrideDebugLevel MUTATES d if env var is set
func environmentOverrideDebugLevel(d *uint, debugLevel uint) {
	key := "DEBUG_LEVEL"
	if value, exists := os.LookupEnv(key); exists {
		if i, err := strconv.Atoi(value); err == nil {
			*d = uint(i)
			if debugLevel > 10 {
				log.Printf("key:%s, d:%d", key, d)
			}
		}
	}
}

// environmentOverrideDebugLevel MUTATES d if env var is set
func environmentOverrideKafkaDebugLevel(d *uint, debugLevel uint) {
	key := "KAFKA_DEBUG_LEVEL"
	if value, exists := os.LookupEnv(key); exists {
		if i, err := strconv.Atoi(value); err == nil {
			*d = uint(i)
			if debugLevel > 10 {
				log.Printf("key:%s, d:%d", key, d)
			}
		}
	}
}

// environmentOverrideGoMaxProcs MUTATES goMaxProcs if env var is set
func environmentOverrideGoMaxProcs(goMaxProcs *uint, debugLevel uint) {
	key := "GOMAXPROCS"
	if value, exists := os.LookupEnv(key); exists {
		if i, err := strconv.Atoi(value); err == nil {
			*goMaxProcs = uint(i)
			if debugLevel > 10 {
				log.Printf("key:%s, goMaxProcs:%d", key, goMaxProcs)
			}
		}
	}
}

// environmentOverrideConfig MUTATES the config if environment variables exist
// this is to allow the environment variables to override the arguments
// (probably poor form to be mutatating)
func environmentOverrideConfig(c *gdp_config.GdpConfig, debugLevel uint) {

	var key string

	key = "POLL_FREQUENCY"
	if value, exists := os.LookupEnv(key); exists {
		if d, err := time.ParseDuration(value); err == nil {
			c.PollFrequency = durationpb.New(d)
			if debugLevel > 10 {
				log.Printf("key:%s, c.PollingFrequency:%s", key, c.PollFrequency.String())
			}
		}
	}

	key = "POLL_TIMEOUT"
	if value, exists := os.LookupEnv(key); exists {
		if d, err := time.ParseDuration(value); err == nil {
			c.PollTimeout = durationpb.New(d)
			if debugLevel > 10 {
				log.Printf("key:%s, c.PollingFrequency:%s", key, c.PollTimeout.String())
			}
		}
	}

	key = "MAX_LOOPS"
	if value, exists := os.LookupEnv(key); exists {
		if i, err := strconv.ParseInt(value, base10, sixtyFour); err == nil {
			c.MaxLoops = uint64(i)
			if debugLevel > 10 {
				log.Printf("key:%s, c.MaxLoops:%d", key, c.MaxLoops)
			}
		}
	}

	key = "DEST"
	if value, exists := os.LookupEnv(key); exists {
		c.Dest = value
		if debugLevel > 10 {
			log.Printf("key:%s, c.Dest:%s", key, c.Dest)
		}
	}

	key = "DEST_WRITE_FILES"
	if value, exists := os.LookupEnv(key); exists {
		if i, err := strconv.Atoi(value); err == nil {
			c.DestWriteFiles = uint32(i)
			if debugLevel > 10 {
				log.Printf("key:%s, c.DestWriteFiles:%d", key, c.DestWriteFiles)
			}
		}
	}

	key = "PROM_PROTO_FILE"
	if value, exists := os.LookupEnv(key); exists {
		c.PromProtoFile = value
		if debugLevel > 10 {
			log.Printf("key:%s, c.PromProtoFile:%s", key, c.PromProtoFile)
		}
	}

	key = "PROM_PROTO_LIST_FILE"
	if value, exists := os.LookupEnv(key); exists {
		c.PromListProtoFile = value
		if debugLevel > 10 {
			log.Printf("key:%s, c.PromListProtoFile:%s", key, c.PromListProtoFile)
		}
	}

	key = "KAFKA_SCHEMA_URL"
	if value, exists := os.LookupEnv(key); exists {
		c.KafkaSchemaUrl = value
		if debugLevel > 10 {
			log.Printf("key:%s, c.KafkaSchemaUrl:%s", key, c.KafkaSchemaUrl)
		}
	}

	key = "KAFKA_PRODUCE_TIMEOUT"
	if value, exists := os.LookupEnv(key); exists {
		if d, err := time.ParseDuration(value); err == nil {
			c.KafkaProduceTimeout = durationpb.New(d)
			if debugLevel > 10 {
				log.Printf("key:%s, c.KafkaProduceTimeout:%s", key, c.KafkaProduceTimeout.AsDuration())
			}
		}
	}

	key = "LABEL"
	if value, exists := os.LookupEnv(key); exists {
		c.Label = value
		if debugLevel > 10 {
			log.Printf("key:%s, c.Label:%s", key, c.Label)
		}
	}

	key = "TAG"
	if value, exists := os.LookupEnv(key); exists {
		c.Tag = value
		if debugLevel > 10 {
			log.Printf("key:%s, c.Tag:%s", key, c.Tag)
		}
	}

}

func printConfig(c *gdp_config.GdpConfig, comment string) {
	fmt.Println(comment)
	fmt.Println("c.PollFrequency:", c.PollFrequency.AsDuration())
	fmt.Println("c.PollTimeout:", c.PollTimeout.AsDuration())
	fmt.Println("c.MaxLoops:", c.MaxLoops)
	fmt.Println("c.Dest:", c.Dest)
	fmt.Println("c.DestWriteFiles:", c.DestWriteFiles)
	fmt.Println("c.PromProtoFile:", c.PromProtoFile)
	fmt.Println("c.PromListProtoFile:", c.PromListProtoFile)
	fmt.Println("c.KafkaSchemaUrl:", c.KafkaSchemaUrl)
	fmt.Println("c.KafkaProduceTimeout:", c.KafkaProduceTimeout.AsDuration())
	fmt.Println("c.DebugLevel:", c.DebugLevel)
	fmt.Println("c.Label:", c.Label)
	fmt.Println("c.Tag:", c.Tag)
}
