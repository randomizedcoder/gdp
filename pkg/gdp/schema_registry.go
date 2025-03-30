package gdp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/randomizedcoder/gdp/pkg/gdp_config"
)

func (g *GDP) InitSchemaRegistry(ctx context.Context, wg *sync.WaitGroup) {

	defer wg.Done()

	if g.debugLevel > 10 {
		log.Println("InitSchemaRegistry start")
	}

	g.setTopic(g.MarshalConfigs)
	g.setFilename(g.MarshalConfigs)

	w := new(sync.WaitGroup)

	i := 0
	g.MarshalConfigs.Range(func(key, value interface{}) bool {

		mc := value.(*gdp_config.MarshalConfig)
		if g.debugLevel > 10 {
			log.Printf("InitSchemaRegistry i:%d, mc:%+v", i, mc)
		}

		if mc.KafkaHeader {
			if g.debugLevel > 10 {
				log.Printf("InitSchemaRegistry registering i:%d, mc:%+v", i, mc)
			}
			w.Add(1)
			go g.registerProtobufSchemaRestful(w, mc, i)
		}

		i++
		return true
	})

	w.Wait()

	if g.debugLevel > 10 {
		log.Println("InitSchemaRegistry w.Wait() ---------------------")
	}

	if g.debugLevel > 10 {
		var i int
		g.MarshalConfigs.Range(func(key, value interface{}) bool {
			log.Printf("InitSchemaRegistry MarshalConfigs i:%d Key: %v, Value: %v\n", i, key, value)
			i++
			return true
		})
	}

	if g.debugLevel > 10 {
		log.Println("InitSchemaRegistry complete -------------------------")
	}

}

func (g *GDP) setTopic(MarshalConfigs *sync.Map) {
	MarshalConfigs.Range(func(key, value interface{}) bool {
		mc := value.(*gdp_config.MarshalConfig)
		topic := mc.MarshalType
		if mc.KafkaHeader {
			topic += "Kafka"
		}
		if mc.Protodelim {
			topic += "Protodelim"
		}
		mc.Topic = topic
		if g.debugLevel > 10 {
			log.Printf("setTopic mc:%v", mc)
		}
		return true
	})
}

func (g *GDP) setFilename(MarshalConfigs *sync.Map) {
	MarshalConfigs.Range(func(key, value interface{}) bool {
		mc := value.(*gdp_config.MarshalConfig)
		var filename string
		switch mc.MarshalType {
		case "ProtobufSingle":
			filename = g.config.PromProtoFile
		case "Protobuf":
			filename = g.config.PromListProtoFile
		case "ProtobufList":
			filename = g.config.PromListProtoFile
		}
		mc.Filename = filename
		if g.debugLevel > 10 {
			log.Printf("setFilename mc:%v", mc)
		}
		return true
	})
}

// registerProtobufSchemaRestful is a HTTP REST way to register schema,
// rather than using the franz-go helper
func (g *GDP) registerProtobufSchemaRestful(wg *sync.WaitGroup, mc *gdp_config.MarshalConfig, i int) (int, error) {

	defer wg.Done()

	if g.debugLevel > 10 {
		log.Printf("registerProtobufSchema g.mcToTopic.Store i:%d, topic:%s", i, mc.Topic)
	}

	url := fmt.Sprintf("%s/subjects/%s-value/versions", g.config.KafkaSchemaUrl, mc.Topic)
	if g.debugLevel > 10 {
		log.Printf("registerProtobufSchema url:%s\n", url)
	}

	// if g.debugLevel > 10 {
	// 	log.Printf("registerProtobufSchema filename:%s\n", mc.Filename)
	// }

	data, err := os.ReadFile(mc.Filename)
	if err != nil {
		log.Fatalf("registerProtobufSchema failed to read proto file: %v", err)
	}

	// SchemaRequest represents the payload to send to the schema registry
	type SchemaRequest struct {
		Schema     string `json:"schema"`
		SchemaType string `json:"schemaType"`
		//Name       string `json:"name"` // Name only exists for Confluent kafka, not redpanda
	}

	reqBody := SchemaRequest{
		Schema:     string(data),
		SchemaType: "PROTOBUF",
		//Name:       "GDP_flat_record.v1.Envelope"
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		log.Fatalf("registerProtobufSchema failed to marshal request: %v", err)
	}

	resp, err := http.Post(url, "application/vnd.schemaregistry.v1+json", bytes.NewReader(bodyBytes))
	if err != nil {
		log.Fatalf("registerProtobufSchema failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("registerProtobufSchema unexpected status code: %d", resp.StatusCode)
	}

	var result struct {
		ID int `json:"id"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	if g.debugLevel > 10 {
		log.Printf("registerProtobufSchema i:%d, result.ID:%d, mc:%v", i, result.ID, mc)
	}

	mc.SchemaID = uint32(result.ID)

	if g.debugLevel > 10 {
		log.Printf("registerProtobufSchema schema registered successfully under subject:%s-value, id:%d", mc.Topic, mc.SchemaID)
	}

	return result.ID, nil

}

// g.registerProtobufSchema(ctx)

// schemaID, err := g.getLatestSchemaID()
// if err != nil {
// 	log.Fatalf("InitDestKafka g.getLatestSchemaID() err:%v", err)
// }

// if schemaID != g.schemaID {
// 	log.Fatalf("InitDestKafka schemaID:%d != g.schemaID:%d", schemaID, g.schemaID)
// }

// g.kSerde.Register(
// 	g.schemaID,
// 	&gdpp.Envelope_PromRecordCounter{},
// 	sr.EncodeFn(func(v any) ([]byte, error) {
// 		return *g.protobufListMarshalNoHeader(v.(*gdpp.Envelope_PromRecordCounter)), nil
// 	}),
// 	sr.Index(0),
// 	// No need to decode currently
// 	// sr.DecodeFn(func(b []byte, v any) error {
// 	// 	return avro.Unmarshal(avroSchema, b, v)
// 	// }),
// )
// https://github.com/cloudhut/owl-shop/blob/7095131ece7a0fee9a58d00b4fbc9f820a0d13be/pkg/shop/order_service.go#L184
// code lifted from
// https://github.com/twmb/franz-go/blob/35ab5e5f5327ca190b49d4b14f326db4365abb9f/examples/schema_registry/schema_registry.go#L65C1-L74C3
// https://pkg.go.dev/github.com/twmb/franz-go/pkg/sr@v1.3.0#Index

// getLatestSchemaID is a go http method for getting the schema id,
// rather than using the franz-go helpers
func (g *GDP) getLatestSchemaID() (int, error) {

	url := fmt.Sprintf("%s/subjects/%s-value/versions/latest", g.config.KafkaSchemaUrl, defaultTopicCst) // fix me

	if g.debugLevel > 10 {
		log.Printf("getLatestSchemaID url:%s\n", url)
	}

	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		log.Fatal("getLatestSchemaID http.StatusNotFound")
	}

	var result struct {
		ID int `json:"id"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	if g.debugLevel > 10 {
		log.Printf("getLatestSchemaID result.ID:%d", result.ID)
	}

	return result.ID, nil
}

// // registerProtobufSchema uses the franz-go helpers to register the schema
// // as apposed to registerProtobufSchemaRestful, which is a pure go http way to do it
// func (g *GDP) registerProtobufSchema(ctx context.Context) {

// 	var err error
// 	g.kRegClient, err = sr.NewClient(sr.URLs(g.config.KafkaSchemaUrl))
// 	if err != nil {
// 		log.Fatalf("unable to create schema registry client: %v", err)
// 	}

// 	schemaBytes, errF := os.ReadFile(g.config.PromProtoFile)
// 	if errF != nil {
// 		log.Fatalf("registerProtobufSchema failed to read proto file: %v", errF)
// 	}

// 	schema := sr.Schema{
// 		Schema: string(schemaBytes),
// 		Type:   sr.TypeProtobuf,
// 	}

// 	s, errC := g.kRegClient.CreateSchema(ctx, defaultTopicCst+"-value", schema) // fix me
// 	if errC != nil {
// 		log.Fatalf("registerProtobufSchema CreateSchema er: %v", errC)
// 	}

// 	g.schemaID = s.ID

// 	if g.debugLevel > 10 {
// 		log.Printf("registerProtobufSchema schema registered, g.schemaID:%d\n", g.schemaID)
// 	}

// }
