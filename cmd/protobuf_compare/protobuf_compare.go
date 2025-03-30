package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"

	gdpp "github.com/randomizedcoder/gdp/pkg/prometheus"
	"google.golang.org/protobuf/proto"
)

const (
	recordFilename     = "record.bin"
	promRecordFilename = "promrecord.bin"
)

// tempRecord is a temporary struct to hold the test data
type tempRecord struct {
	Function      string
	Type          string
	Variable      string
	Value         float64
	TimestampNs   float64
	Hostname      string
	Label         string
	Tag           string
	RecordCounter uint64
}

var (
	// Passed by "go build -ldflags" for the show version
	commit  string
	date    string
	version string
)

func main() {

	v := flag.Bool("v", false, "show version")

	flag.Parse()

	// Print version information passed in via ldflags in the Makefile
	if *v {
		log.Printf("commit:%s\tdate(UTC):%s\tversion:%s", commit, date, version)
		os.Exit(0)
	}

	// Create sample data in the temporary struct
	temp := tempRecord{
		Function:      "TestFunction",
		Type:          "TestType",
		Variable:      "TestVariable",
		Value:         123.45,
		TimestampNs:   1234567890,
		Hostname:      "testHost",
		Label:         "testLabel",
		Tag:           "testTag",
		RecordCounter: 1,
	}

	// Create Envelope_PromRecordCounter and copy data from temp
	record := &gdpp.Envelope_PromRecordCounter{
		Function:      temp.Function,
		Type:          temp.Type,
		Variable:      temp.Variable,
		Value:         temp.Value,
		TimestampNs:   temp.TimestampNs,
		Hostname:      temp.Hostname,
		Label:         temp.Label,
		Tag:           temp.Tag,
		RecordCounter: temp.RecordCounter,
	}

	// Create PromRecordCounter and copy data from temp
	promRecord := &gdpp.PromRecordCounter{
		Function:      temp.Function,
		Type:          temp.Type,
		Variable:      temp.Variable,
		Value:         temp.Value,
		TimestampNs:   temp.TimestampNs,
		Hostname:      temp.Hostname,
		Label:         temp.Label,
		Tag:           temp.Tag,
		RecordCounter: temp.RecordCounter,
	}

	// Marshal Envelope_PromRecordCounter
	recordBytes, err := proto.Marshal(record)
	if err != nil {
		log.Fatalf("Failed to marshal record: %v", err)
	}

	// Marshal PromRecordCounter
	promRecordBytes, err := proto.Marshal(promRecord)
	if err != nil {
		log.Fatalf("Failed to marshal promRecord: %v", err)
	}

	// Write to files
	err = os.WriteFile(recordFilename, recordBytes, 0644)
	if err != nil {
		log.Fatalf("Failed to write record to file: %v", err)
	}
	log.Printf("Wrote record to file: %s", recordFilename)

	err = os.WriteFile(promRecordFilename, promRecordBytes, 0644)
	if err != nil {
		log.Fatalf("Failed to write promRecord to file: %v", err)
	}
	log.Printf("Wrote promRecord to file: %s", promRecordFilename)

	// Byte-level diff
	diff, err := compareFiles(recordFilename, promRecordFilename)
	if err != nil {
		log.Fatalf("Failed to compare files: %v", err)
	}

	if diff {
		log.Println("Files are different.")
	} else {
		log.Println("Files are identical.")
	}
}

func compareFiles(file1, file2 string) (bool, error) {
	data1, err := os.ReadFile(file1)
	if err != nil {
		return false, fmt.Errorf("failed to read file %s: %w", file1, err)
	}

	data2, err := os.ReadFile(file2)
	if err != nil {
		return false, fmt.Errorf("failed to read file %s: %w", file2, err)
	}

	return !bytes.Equal(data1, data2), nil
}
