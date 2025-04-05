package gdp

import (
	"bytes"
	"io"
	"os"
	"testing"
	"time"

	"github.com/randomizedcoder/gdp/pkg/gdp_config"
	gdpp "github.com/randomizedcoder/gdp/pkg/prometheus"
	"google.golang.org/protobuf/encoding/protodelim"
	"google.golang.org/protobuf/proto"
)

type mtest struct {
	comment   string
	envelope  *gdpp.Envelope
	protofile string
	mc        *gdp_config.MarshalConfig
}

// TestMarshal verifies that the marshal function can create the same data as the clickhouse
// export of the protobufList
func TestMarshal(t *testing.T) {
	tests := []mtest{
		{
			comment: "ProtobufListProtodelim",
			envelope: &gdpp.Envelope{
				Rows: []*gdpp.Envelope_PromRecordCounter{
					{
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
					},
					{
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
					},
				},
			},
			protofile: "test_data/gdp.ProtobufListProtodelim.bin_2025_04_05",
			mc: &gdp_config.MarshalConfig{
				MarshalType: "ProtobufList",
				Topic:       "ProtobufListProtodelim",
				KafkaHeader: false,
				Protodelim:  true,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.comment, func(t *testing.T) {
			data, err := os.ReadFile(test.protofile)
			if err != nil {
				t.Fatalf("Failed to read file %s: %v", test.protofile, err)
			}

			reader := bytes.NewReader(data)

			var unmarshaledEnvelope gdpp.Envelope
			for {
				err = protodelim.UnmarshalFrom(reader, &unmarshaledEnvelope)
				if err != nil {
					if err == io.EOF {
						break
					}
					t.Fatalf("Failed to unmarshal protobuf from file %s: %v", test.protofile, err)
				}

				if !proto.Equal(test.envelope, &unmarshaledEnvelope) {
					t.Errorf("Unmarshaled envelope does not match expected envelope\nExpected: %v\nGot: %v", test.envelope, &unmarshaledEnvelope)
				}
			}
		})
	}
}
