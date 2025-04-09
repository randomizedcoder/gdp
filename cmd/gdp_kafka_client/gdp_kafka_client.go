package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	gdpp "github.com/randomizedcoder/gdp/pkg/prometheus"
	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/protobuf/encoding/protodelim"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const (
	schemaRegistryURLCst = "http://localhost:18081"
	brokerCst            = "localhost:19092"
	topicCst             = "gdp"
	groupIDCst           = "gdp-consumer-group"
	clientIDCst          = "gdp-client"
	pbCst                = "row" // or envelope
	kafkaHeaderCst       = false
	delimCst             = false

	debugLevelCst        = 11
	signalChannelSizeCst = 10
	KafkaHeaderSizeCst   = 6
)

var (
	debugLevel uint
)

func main() {

	d := flag.Uint("d", debugLevelCst, "debug level")
	topic := flag.String("topic", topicCst, "topic")
	group := flag.String("group", groupIDCst, "consumer group")
	pb := flag.String("pb", pbCst, "protobuf type, row/envelope")
	k := flag.Bool("k", kafkaHeaderCst, "kafka header")
	delim := flag.Bool("delim", delimCst, "protobuf delim")

	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.LUTC | log.Lshortfile | log.Lmsgprefix)

	debugLevel = *d

	if debugLevel > 10 {
		log.Printf("Parsed flags:")
		log.Printf("  Debug Level: %d", debugLevel)
		log.Printf("  Topic: %s", *topic)
		log.Printf("  Consumer Group: %s", *group)
		log.Printf("  Protobuf Type: %s", *pb)
		log.Printf("  Kafka Header: %t", *k)
		log.Printf("  Length-Delimited: %t", *delim)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	complete := make(chan struct{}, signalChannelSizeCst)
	go initSignalHandler(cancel, complete)

	// regClient, err := sr.NewClient(sr.URLs(schemaRegistryURLCst))
	// if err != nil {
	// 	log.Fatalf("unable to create schema registry client: %v", err)
	// }

	//serde := sr.NewSerde(regClient)

	opts := []kgo.Opt{
		kgo.SeedBrokers(brokerCst),
		kgo.ConsumerGroup(*group),
		kgo.ConsumeTopics(*topic),
		kgo.ClientID(clientIDCst),
		//kgo.RecordDeserializer(serde.RecordDeserializer()),
	}

	cl, err := kgo.NewClient(opts...)
	if err != nil {
		log.Fatalf("unable to create client: %v", err)
	}
	defer cl.Close()

	if debugLevel > 10 {
		log.Println("kgo.NewClient complete")
	}

	startTime := time.Now()

	for i := 0; ; i++ {
		pollStartTime := time.Now()
		processMessages(ctx, startTime, pollStartTime, cl, *k, *delim, *pb)
	}
}

func processMessages(ctx context.Context, startTime time.Time, pollStartTime time.Time, cl *kgo.Client, kafkaHeader bool, delim bool, pbType string) {
	fetches := cl.PollFetches(ctx)
	if errs := fetches.Errors(); len(errs) > 0 {
		log.Printf("fetch errors: %v", errs)
		return
	}

	fetches.EachRecord(func(record *kgo.Record) {
		processRecord(startTime, pollStartTime, record, kafkaHeader, delim, pbType)
	})
}

func processRecord(startTime time.Time, pollStartTime time.Time, record *kgo.Record, kafkaHeader bool, delim bool, pbType string) {
	if kafkaHeader {
		if len(record.Value) < KafkaHeaderSizeCst {
			if debugLevel > 10 {
				log.Printf("len(record.Value): %d is too short", len(record.Value))
			}
			return
		}
		processRecordInner(startTime, pollStartTime, record.Value[KafkaHeaderSizeCst:], pbType, delim)
	} else if delim {
		processRecordInner(startTime, pollStartTime, record.Value, pbType, delim)
	}
}

func processRecordInner(startTime time.Time, pollStartTime time.Time, b []byte, pbType string, delim bool) {
	var (
		msg interface{}
		err error
	)

	if pbType == "row" {
		msg = &gdpp.Envelope_PromRecordCounter{}
	} else {
		msg = &gdpp.Envelope{}
	}

	if debugLevel > 1000 {
		log.Printf("processRecordInner pbType:%s, len(b):%d", pbType, len(b))
	}

	reader := bytes.NewReader(b)
	if debugLevel > 1000 {
		log.Printf("processRecordInner reader created")
	}

	if debugLevel > 10 {
		log.Printf("processRecordInner before UnmarshalFrom")
	}

	if delim {
		switch msg := msg.(type) {
		case *gdpp.Envelope_PromRecordCounter:
			err = protodelim.UnmarshalFrom(reader, msg)
		case *gdpp.Envelope:
			err = protodelim.UnmarshalFrom(reader, msg)
		default:
			err = fmt.Errorf("unknown message type: %T", msg)
		}
	} else {
		switch msg := msg.(type) {
		case *gdpp.Envelope_PromRecordCounter:
			err = proto.Unmarshal(b, msg)
		case *gdpp.Envelope:
			err = proto.Unmarshal(b, msg)
		default:
			err = fmt.Errorf("unknown message type: %T", msg)
		}
	}

	if err != nil {
		log.Printf("Failed to unmarshal protobuf: %v", err)
		return
	}

	var jsonBytes []byte
	switch msg := msg.(type) {
	case *gdpp.Envelope_PromRecordCounter:
		jsonBytes, err = protojson.Marshal(msg)
	case *gdpp.Envelope:
		jsonBytes, err = protojson.Marshal(msg)
	default:
		err = fmt.Errorf("unknown message type: %T", msg)
	}

	if err != nil {
		log.Printf("Failed to marshal to JSON: %v", err)
		return
	}

	log.Printf("Processed st:%dms pt:%dms message: %s", time.Since(startTime).Milliseconds(), time.Since(pollStartTime).Milliseconds(), jsonBytes)
}

// initSignalHandler sets up signal handling for the process, and
// will call cancel() when received
func initSignalHandler(cancel context.CancelFunc, complete <-chan struct{}) {

	c := make(chan os.Signal, signalChannelSizeCst)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c
	log.Printf("Signal caught, closing application")
	cancel()

	log.Printf("Signal caught, cancel() called, and sleeping to allow goroutines to close")

	select {
	case <-complete:
		log.Printf("<-complete exit(0)")
	default:
		log.Printf("Sleep complete, goodbye! exit(0)")
	}

	os.Exit(0)
}
