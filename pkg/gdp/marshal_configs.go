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
func (g *GDP) initMarshalConfigs(wg *sync.WaitGroup) {

	defer wg.Done()

	g.MarshalConfigs = &sync.Map{}

	i := 0

	marshalTypes := []string{
		//"ProtobufSingle",
		"Protobuf",
		"ProtobufList",
	}

	for _, marshalType := range marshalTypes {

		for k := 0; k <= 1; k++ {
			for d := 0; d <= 1; d++ {
				kafka := k == 1
				delim := d == 1

				if kafka && !delim {
					g.MarshalConfigs.Store(i, &gdp_config.MarshalConfig{
						MarshalType: marshalType,
						KafkaHeader: kafka,
						Protodelim:  delim,
					})
					i++
				}
			}
		}
	}

	// for _, mc := range g.MarshalConfigs {
	// 	if g.debugLevel > 10 {
	// 		log.Printf("initMarshalConfigs mc:%+v", mc)
	// 	}
	// 	if err := protovalidate.Validate(mc); err != nil {
	// 		log.Fatal("config validation failed:", err)
	// 	}
	// }

	if g.debugLevel > 10 {
		log.Println("initMarshalConfigs complete")
	}
}
