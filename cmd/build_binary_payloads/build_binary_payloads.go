package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/randomizedcoder/gdp/pkg/gdp"
	"github.com/randomizedcoder/gdp/pkg/gdp_config"
	gdpp "github.com/randomizedcoder/gdp/pkg/prometheus"
	"google.golang.org/protobuf/encoding/protodelim"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	debugLevelCst        = 11
	signalChannelSizeCst = 10

	//promProtoFileCst = "/prometheus.proto"
	promProtoFileCst     = "/prometheus_protolist.proto"
	promListProtoFileCst = "/prometheus_protolist.proto"

	outFolderCst = "./"
	fCst         = "gdp.ProtobufListProtodelim.bin"
)

var (
	debugLevel uint
)

func main() {

	of := flag.String("of", outFolderCst, "output folder")
	f := flag.String("f", fCst, "file name")
	w := flag.Bool("w", false, "write")
	r := flag.Bool("r", false, "read rows")
	d := flag.Uint("d", debugLevelCst, "debug level")

	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.LUTC | log.Lshortfile | log.Lmsgprefix)

	debugLevel = *d

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	complete := make(chan struct{}, signalChannelSizeCst)
	go initSignalHandler(cancel, complete)

	g := &gdp.GDP{}

	g.Config = &gdp_config.GdpConfig{
		PromProtoFile:     promProtoFileCst,
		PromListProtoFile: promListProtoFileCst,
	}

	wg := new(sync.WaitGroup)
	wg.Add(1)
	g.InitSyncPools(wg)
	wg.Wait()

	if *w {
		write(of, f, g)
	} else {
		read(of, f, *r)
	}

}

func read(of *string, f *string, readRows bool) {
	filename := *of + *f
	log.Printf("Reading file: %s", filename)

	// Read the entire file into a byte slice
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("Failed to read file %s: %v", filename, err)
	}

	// Create a new reader from the byte slice
	reader := bytes.NewReader(data)

	// Unmarshal the length-delimited protobuf
	var envelope gdpp.Envelope
	for {
		err = protodelim.UnmarshalFrom(reader, &envelope)
		if err != nil {
			if err == io.EOF {
				break // End of file reached
			}
			log.Fatalf("Failed to unmarshal protobuf from file %s: %v", filename, err)
		}

		if readRows {
			for _, row := range envelope.Rows {
				// Marshal the protobuf to JSON
				jsonBytes, err := protojson.Marshal(row)
				if err != nil {
					log.Fatalf("Failed to marshal protobuf to JSON: %v", err)
				}

				// Print the JSON to the console
				fmt.Println(string(jsonBytes))
			}
		} else {
			// Marshal the protobuf to JSON
			jsonBytes, err := protojson.Marshal(&envelope)
			if err != nil {
				log.Fatalf("Failed to marshal protobuf to JSON: %v", err)
			}

			// Print the JSON to the console
			fmt.Println(string(jsonBytes))
		}
	}
}

func write(of *string, f *string, g *gdp.GDP) {
	// ef1acd401213 :) SELECT * FROM gdp.ProtobufListProtodelim LIMIT 10;

	// SELECT *
	// FROM gdp.ProtobufListProtodelim
	// LIMIT 10

	// Query id: 02522c67-aaf1-4d50-ac65-5bc1b1471855

	// 	 ┌───────────────────TimestampNs─┬─Hostname─┬─Pop──┬─Label──┬─Tag──┬─PollCounter─┬─RecordCounter─┬─Function──┬─Variable──┬─Type──┬──Value─┐
	// 1. │ 2025-04-05 10:00:00.000000000 │ host1    │ pop1 │ label1 │ tag1 │         100 │             1 │ function1 │ variable1 │ type1 │ 123.45 │
	// 2. │ 2025-04-05 11:00:00.000000000 │ host2    │ pop2 │ label2 │ tag2 │         200 │             2 │ function2 │ variable2 │ type2 │  678.9 │
	// 	 └───────────────────────────────┴──────────┴──────┴────────┴──────┴─────────────┴───────────────┴───────────┴───────────┴───────┴────────┘

	// 2 rows in set. Elapsed: 0.002 sec.

	// 	ef1acd401213 :) SELECT
	// 	toUnixTimestamp64Nano(TimestampNs) / 1e9 AS timestamp_ns,
	// 	Hostname AS hostname,
	// 	Pop AS pop,
	// 	Label AS label,
	// 	Tag AS tag,
	// 	PollCounter AS poll_counter,
	// 	RecordCounter AS record_counter,
	// 	Function AS function,
	// 	Variable AS variable,
	// 	Type AS type,
	// 	Value AS value
	// FROM gdp.ProtobufListProtodelim
	// INTO OUTFILE 'gdp.ProtobufListProtodelim.bin'
	// FORMAT ProtobufList
	// SETTINGS format_schema = '/var/lib/clickhouse/format_schemas/prometheus_protolist.proto:PromRecordCounter';

	// SELECT
	// 	toUnixTimestamp64Nano(TimestampNs) / 1000000000. AS timestamp_ns,
	// 	Hostname AS hostname,
	// 	Pop AS pop,
	// 	Label AS label,
	// 	Tag AS tag,
	// 	PollCounter AS poll_counter,
	// 	RecordCounter AS record_counter,
	// 	Function AS function,
	// 	Variable AS variable,
	// 	Type AS type,
	// 	Value AS value
	// FROM gdp.ProtobufListProtodelim
	// INTO OUTFILE 'gdp.ProtobufListProtodelim.bin'
	// FORMAT ProtobufList
	// SETTINGS format_schema = '/var/lib/clickhouse/format_schemas/prometheus_protolist.proto:PromRecordCounter'

	// Query id: 1bd7d073-6102-4a6d-a797-cfe890a8d2a8

	// 2 rows in set. Elapsed: 0.002 sec.

	proto := &gdpp.Envelope{}
	proto.Rows = append(proto.Rows, &gdpp.Envelope_PromRecordCounter{
		TimestampNs:   float64(time.Date(2025, time.April, 5, 10, 0, 0, 0, time.UTC).UnixNano()) / 1e9,
		Hostname:      "host1",
		Pop:           "pop1",
		Label:         "label1",
		Tag:           "tag1",
		PollCounter:   100,
		RecordCounter: 1,
		Function:      "function1",
		Variable:      "variable1",
		Type:          "type1",
		Value:         123.45,
	})

	// Row 2
	proto.Rows = append(proto.Rows, &gdpp.Envelope_PromRecordCounter{
		TimestampNs:   float64(time.Date(2025, time.April, 5, 11, 0, 0, 0, time.UTC).UnixNano()) / 1e9,
		Hostname:      "host2",
		Pop:           "pop2",
		Label:         "label2",
		Tag:           "tag2",
		PollCounter:   200,
		RecordCounter: 2,
		Function:      "function2",
		Variable:      "variable2",
		Type:          "type2",
		Value:         678.90,
	})

	mc := &gdp_config.MarshalConfig{
		MarshalType: "ProtobufList",
		Topic:       "ProtobufListProtodelim",
		KafkaHeader: false,
		Protodelim:  true,
	}

	b := g.ReturnBinaryPayload(proto, mc)

	err := os.WriteFile(*of+*f, *b, 0644)
	if err != nil {
		log.Printf("Failed to write: %s err: %v", *of+*f, err)
	} else {
		log.Printf("Wrote file: %s", *of+*f)
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
