package gdp

import (
	"log"
	"sync"

	"github.com/randomizedcoder/gdp/pkg/gdp_config"
)

// initMarshalConfigs creates the sets of config GDP is testing
// this iterates over each combination of protobuf type,
// and then each combination of kafka header and delimiting
// this is obviously very silly, but it's just a demo
func (g *GDP) InitMarshalConfigs(wg *sync.WaitGroup) {

	defer wg.Done()

	g.MarshalConfigs = &sync.Map{}

	i := 0

	g.MarshalConfigs.Store(i, &gdp_config.MarshalConfig{
		MarshalType:          "ProtobufSingle",
		KafkaHeader:          false,
		Protodelim:           false,
		CreateDatabaseSchema: true,
		MaxPollSend:          0,
	})

	i++
	g.MarshalConfigs.Store(i, &gdp_config.MarshalConfig{
		MarshalType:          "Protobuf",
		KafkaHeader:          false,
		Protodelim:           true,
		CreateDatabaseSchema: true,
		MaxPollSend:          0,
	})

	i++
	g.MarshalConfigs.Store(i, &gdp_config.MarshalConfig{
		MarshalType:          "ProtobufList",
		KafkaHeader:          false,
		Protodelim:           true,
		CreateDatabaseSchema: true,
		MaxPollSend:          0,
	})

	// add kafka headers -
	i++
	g.MarshalConfigs.Store(i, &gdp_config.MarshalConfig{
		MarshalType:          "ProtobufSingle",
		KafkaHeader:          true,
		Protodelim:           false,
		CreateDatabaseSchema: false, // clickhouse doesn't support reading the kafka header
		MaxPollSend:          1,     // only send one, so demonstrate redpanda schema
	})

	i++
	g.MarshalConfigs.Store(i, &gdp_config.MarshalConfig{
		MarshalType:          "Protobuf",
		KafkaHeader:          true,
		Protodelim:           false,
		CreateDatabaseSchema: false,
		MaxPollSend:          1, // only send one, so demonstrate redpanda schema
	})

	i++
	g.MarshalConfigs.Store(i, &gdp_config.MarshalConfig{
		MarshalType:          "ProtobufList",
		KafkaHeader:          true,
		Protodelim:           false,
		CreateDatabaseSchema: false,
		MaxPollSend:          1, // only send one, so demonstrate redpanda schema
	})

	// marshalTypes := []string{
	// 	"ProtobufSingle",
	// 	"Protobuf",
	// 	"ProtobufList",
	// }

	// i := 0
	// for _, marshalType := range marshalTypes {

	// 	for k := 0; k <= 1; k++ {
	// 		for d := 0; d <= 1; d++ {
	// 			kafka := k == 1
	// 			delim := d == 1

	// 			if g.debugLevel > 10 {
	// 				log.Printf("initMarshalConfigs store i:%d", i)
	// 			}

	// 			store := false

	// 			//if kafka && !delim {
	// 			//if kafka || delim {
	// 			if marshalType == "ProtobufSingle" {
	// 				if !kafka || !delim {
	// 					store = true
	// 				}
	// 			} else {
	// 				if !kafka && delim {
	// 					store = true
	// 				}
	// 			}

	// 			if store {
	// 				g.MarshalConfigs.Store(i, &gdp_config.MarshalConfig{
	// 					MarshalType: marshalType,
	// 					KafkaHeader: kafka,
	// 					Protodelim:  delim,
	// 				})
	// 				i++
	// 			}
	// 		}
	// 	}

	g.SetTopic(g.MarshalConfigs)
	g.SetFilename(g.MarshalConfigs)

	g.MarshalConfigs.Range(func(key, value interface{}) bool {
		mc := value.(*gdp_config.MarshalConfig)
		if g.debugLevel > 10 {
			log.Printf("initMarshalConfigs mc:%+v", mc)
		}
		// Originally I was validating, but hte
		// if err := protovalidate.Validate(mc); err != nil {
		// 	log.Fatal("config validation failed:", err)
		// }
		return true
	})

	if g.debugLevel > 10 {
		log.Println("initMarshalConfigs complete")
	}
}

func (g *GDP) SetTopic(MarshalConfigs *sync.Map) {
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

func (g *GDP) SetFilename(MarshalConfigs *sync.Map) {
	MarshalConfigs.Range(func(key, value interface{}) bool {
		mc := value.(*gdp_config.MarshalConfig)
		var filename string
		switch mc.MarshalType {
		case "ProtobufSingle":
			filename = g.Config.PromProtoFile
		case "Protobuf":
			filename = g.Config.PromProtoFile
		case "ProtobufList":
			filename = g.Config.PromListProtoFile
		default:
			log.Fatal("setFilename unknown mc.MarshalType")
		}
		mc.Filename = filename
		if g.debugLevel > 10 {
			log.Printf("setFilename mc:%v", mc)
		}
		return true
	})
}
