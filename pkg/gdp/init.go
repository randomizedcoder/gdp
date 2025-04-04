package gdp

import (
	"context"
	"log"
	"os"
	"sync"
	"time"
)

const (
	destinationReadyChSize    = 2
	changePollFrequencyChSize = 2
	pollRequestChSize         = 2
)

func (g *GDP) Init(ctx context.Context) {

	startTime := time.Now()

	if g.debugLevel > 10 {
		log.Println("Init starting")
	}

	g.initChannels()

	wg := new(sync.WaitGroup)

	wg.Add(1)
	go g.InitMarshalConfigs(wg)

	// wg.Add(1)
	// go g.InitMarshallers(wg)

	wg.Add(1)
	go g.InitDests(ctx, wg)

	wg.Wait()

	g.InputValidation()

	wg.Add(1)
	go g.InitPromethus(wg)

	wg.Add(1)
	go g.InitSyncPools(wg)

	wg.Add(1)
	go g.InitSchemaRegistry(ctx, wg)

	//g.initSyncMaps()

	g.initHostname()

	wg.Wait()

	if g.debugLevel > 10 {
		log.Printf("Init complete after:%0.3f", time.Since(startTime).Seconds())
	}
}

func (g *GDP) initChannels() {

	g.DestinationReady = make(chan struct{}, destinationReadyChSize)

}

func (g *GDP) initHostname() {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalf("os.Hostname() error:%s", err)
	}
	g.hostname = hostname
}

// func (g *GDP) initSyncMaps() {
// 	g.mcToTopic = &sync.Map{}
// 	g.mcToSchemaId = &sync.Map{}
// }
