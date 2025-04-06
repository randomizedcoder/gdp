package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"flag"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

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

	for i := 0; ; i++ {

		if debugLevel > 1000 {
			log.Printf("i:%d, PollFetches", i)
		}

		fetches := cl.PollFetches(ctx)
		if errs := fetches.Errors(); len(errs) > 0 {
			log.Printf("fetch errors: %v", errs)
			continue
		}

		fetches.EachRecord(func(record *kgo.Record) {

			if *k {
				if len(record.Value) < KafkaHeaderSizeCst {
					if debugLevel > 10 {
						log.Printf("len(record.Value): %d is too short", len(record.Value))
					}
					return
				}

				// if debugLevel > 10 {
				// 	log.Printf("record.Value header: % X", record.Value[:KafkaHeaderSizeCst])
				// }
				// if debugLevel > 100 {
				// 	log.Printf("record.Value:% X", record.Value)
				// }

				schemaID := binary.BigEndian.Uint32(record.Value[1:5])
				protobufSchemaIndex := record.Value[5]
				if debugLevel > 10 {
					log.Printf("len(record.Value):%d, schemaID:%d, protobufSchemaIndex: %d, header:% X", len(record.Value), schemaID, protobufSchemaIndex, record.Value[:KafkaHeaderSizeCst])
				}

				printPayload(record.Value[KafkaHeaderSizeCst:], *pb)
			}

			if *delim {
				printPayloadDelim(record.Value, *pb)
			}

		})
	}
}

func printPayload(b []byte, t string) {

	if t == "row" {
		printRowPayload(b)
		return
	}
	printEnvelopePayload(b)
}

func printPayloadDelim(b []byte, t string) {

	if t == "row" {
		printRowPayloadDelim(b)
		return
	}
	printEnvelopePayloadDelim(b)
}

func printRowPayload(b []byte) {
	var row gdpp.Envelope_PromRecordCounter
	err := proto.Unmarshal(b, &row)
	if err != nil {
		log.Printf("Failed to unmarshal row protobuf: %v", err)
		return
	}
	// Convert to JSON
	jsonBytes, err := protojson.Marshal(&row)
	if err != nil {
		log.Printf("Failed to marshal row to JSON: %v", err)
		return
	}

	log.Printf("printPayload row JSON: %s", jsonBytes)
}

func printRowPayloadDelim(b []byte) {
	reader := bytes.NewReader(b)
	var row gdpp.Envelope_PromRecordCounter
	err := protodelim.UnmarshalFrom(reader, &row)
	if err != nil {
		log.Printf("Failed to unmarshal row protobuf: %v", err)
		return
	}
	// Convert to JSON
	jsonBytes, err := protojson.Marshal(&row)
	if err != nil {
		log.Printf("Failed to marshal row to JSON: %v", err)
		return
	}

	log.Printf("printPayloadDelim row JSON: %s", jsonBytes)
}

func printEnvelopePayload(b []byte) {

	var envelope gdpp.Envelope
	err := proto.Unmarshal(b, &envelope)
	if err != nil {
		log.Printf("Failed to unmarshal envelope protobuf: %v", err)
		return
	}

	jsonBytes, err := protojson.Marshal(&envelope)
	if err != nil {
		log.Printf("Failed to marshal row to JSON: %v", err)
		return
	}
	log.Printf("printPayload envelope JSON: %s", jsonBytes)

	// for _, row := range envelope.Rows {
	// 	// Convert to JSON
	// 	jsonBytes, err := protojson.Marshal(row)
	// 	if err != nil {
	// 		log.Printf("Failed to marshal row to JSON: %v", err)
	// 		return
	// 	}

	// 	log.Printf("printPayload envelope JSON: %s", jsonBytes)
	// }

}

func printEnvelopePayloadDelim(b []byte) {
	reader := bytes.NewReader(b)
	var envelope gdpp.Envelope
	for {
		err := protodelim.UnmarshalFrom(reader, &envelope)
		if err != nil {
			if err == io.EOF {
				break // End of file reached
			}
			log.Printf("Failed to unmarshal envelope protobuf: %v", err)
			return
		}

		jsonBytes, err := protojson.Marshal(&envelope)
		if err != nil {
			log.Printf("Failed to marshal envelope to JSON: %v", err)
			return
		}
		log.Printf("printPayloadDelim envelope JSON: %s", jsonBytes)
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

	log.Printf("Signal caught, cancel() called, and sleeping to allow goroutines to close")

	select {
	case <-complete:
		log.Printf("<-complete exit(0)")
	default:
		log.Printf("Sleep complete, goodbye! exit(0)")
	}

	os.Exit(0)
}
