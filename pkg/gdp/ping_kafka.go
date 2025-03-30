package gdp

import (
	"context"
	"log"
	"time"
)

// pingKafkaWithRetries pings kafka with a retry loop, and sleeps
func (g *GDP) pingKafkaWithRetries(ctx context.Context, retries int, sleepDuration time.Duration) (err error) {
	for i := 0; i < retries; i++ {
		err = g.pingKafka(ctx)
		if err != nil {
			s := sleepDuration * time.Duration(i+1)
			if g.debugLevel > 10 {
				log.Printf("pingKafkaWithRetries i:%d sleep:%0.3fs  %dms", i, s.Seconds(), s.Milliseconds())
			}
			time.Sleep(s)
			continue
		}
		break
	}
	return err
}

// pingKafka performs a kafka ping ( although I don't really know what this does )
func (g *GDP) pingKafka(ctx context.Context) (err error) {
	pCst, pCancel := context.WithTimeout(ctx, kafkaPingTimeoutCst)
	defer pCancel()
	pTime := time.Now()
	err = g.kClient.Ping(pCst)
	if err != nil {
		if g.debugLevel > 10 {
			log.Printf("pingKafka unable to kafka ping:%v time:%0.6fs", err, time.Since(pTime).Seconds())
		}
		return err
	}
	if g.debugLevel > 10 {
		log.Printf("pingKafka kafka ping time:%0.6fs %dms", time.Since(pTime).Seconds(), time.Since(pTime).Milliseconds())
	}
	return err
}
