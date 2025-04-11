package gdp

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/randomizedcoder/gdp/pkg/gdp_config"
)

func (g *GDP) Poller(ctx context.Context, wg *sync.WaitGroup) {

	defer wg.Done()

	if g.debugLevel > 10 {
		log.Printf("Poller started")
	}

	<-g.DestinationReady
	if g.debugLevel > 10 {
		log.Printf("Poller DestinationReady")
	}

	timer := time.NewTimer(0) // fire immediately
	defer timer.Stop()
	//wf := g.Config.DestWriteFiles

breakPoint:
	for pollingLoops := uint64(0); MaxLoopsOrForEver(pollingLoops, g.Config.MaxLoops); pollingLoops++ {

		g.pC.WithLabelValues("Poller", "pollingLoops", "count").Inc()

		// if g.debugLevel > 10 {
		// 	log.Printf("Poller pollingLoops:%d PollFrequency:%s", pollingLoops, g.Config.PollFrequency.AsDuration().String())
		// }

		select {

		case <-ctx.Done():
			break breakPoint

		case <-timer.C:
			g.pC.WithLabelValues("Poller", "ticker", "count").Inc()

			// if g.debugLevel > 10 {
			// 	log.Printf("Poller <-ticker.C pollingLoops:%d", pollingLoops)
			// }

			pollStartTime := time.Now()

			g.performPoll(ctx, pollingLoops)

			pollDuration := time.Since(pollStartTime)

			if g.debugLevel > 10 {
				log.Printf("Poller pollingLoops:%d PollFrequency:%s pollDuration:%0.4fs %dms",
					pollingLoops, g.Config.PollFrequency.AsDuration().String(), pollDuration.Seconds(), pollDuration.Milliseconds())
			}

			g.pH.WithLabelValues("Poller", "pollDuration", "count").Observe(pollDuration.Seconds())

			timer.Reset(g.Config.PollFrequency.AsDuration())
		}

	}

	g.pC.WithLabelValues("Poller", "complete", "count").Inc()
}

// MaxLoopsOrForEver returns true if maxloops == 0, or pollingLoops < maxloops
// This function just allows us to embed if logic into the main pollingLoops for statement
func MaxLoopsOrForEver(pollingLoops uint64, maxLoops uint64) bool {
	if maxLoops != 0 {
		if pollingLoops > maxLoops {
			return false
		}
	}
	return true
}

func (g *GDP) performPoll(ctx context.Context, pollingLoops uint64) error {

	startTime := time.Now()
	defer func() {
		g.pH.WithLabelValues("performPoll", "complete", "count").Observe(time.Since(startTime).Seconds())
	}()
	g.pC.WithLabelValues("performPoll", "start", "count").Inc()

	// if g.debugLevel > 10 {
	// 	log.Println("performPoll start")
	// }

	metrics, err := g.getCountMetrics(ctx)
	if err != nil {
		return err
	}

	if len(metrics) == 0 {
		g.pC.WithLabelValues("performPoll", "noMetrics", "count").Inc()
		if g.debugLevel > 10 {
			log.Println("performPoll no metrics")
		}
		return nil
	}

	// if g.debugLevel > 10 {
	// 	log.Printf("performPoll len(metrics):%d", len(metrics))
	// 	for _, m := range metrics {
	// 		log.Printf("performPoll m:%s", m)
	// 	}
	// }

	// if pollingLoops == 10 {
	// 	err := os.WriteFile("metrics_loop_10.txt", []byte(strings.Join(metrics, "\n")), 0644)
	// 	if err != nil {
	// 		g.pC.WithLabelValues("performPoll", "writeFile", "error").Inc()
	// 		log.Printf("performPoll writeFile error: %v", err)
	// 	}
	// }

	envelope, err := g.parseMetricLinesRegex(metrics, startTime, pollingLoops)
	if err != nil {
		return err
	}

	if len((*envelope).Rows) == 0 {
		g.pC.WithLabelValues("performPoll", "noEnvelope", "count").Inc()
		if g.debugLevel > 10 {
			log.Println("performPoll no envelope")
		}
		return nil
	}

	// if g.debugLevel > 10 {
	// 	log.Printf("performPoll len((*envelope).Rows):%d", len((*envelope).Rows))
	// 	for _, r := range (*envelope).Rows {
	// 		jsonBytes, err := protojson.Marshal(r)
	// 		if err != nil {
	// 			log.Printf("performPoll marshalEnvelope error: %v", err)
	// 		} else {
	// 			log.Printf("performPoll r:%s", jsonBytes)
	// 		}
	// 	}
	// }

	// if pollingLoops == 10 {
	// 	jsonBytes, err := protojson.Marshal(envelope)
	// 	if err != nil {
	// 		log.Printf("performPoll marshalEnvelope error: %v", err)
	// 	} else {
	// 		err := os.WriteFile("envelope_loop_10.json", jsonBytes, 0644)
	// 		if err != nil {
	// 			log.Printf("performPoll writeEnvelope error: %v", err)
	// 		}
	// 	}
	// }

	wg := new(sync.WaitGroup)
	g.MarshalConfigs.Range(func(key, value interface{}) bool {
		mc := value.(*gdp_config.MarshalConfig)

		if mc.MaxPollSend == 00 || pollingLoops < uint64(mc.MaxPollSend) {

			wg.Add(1)
			go g.sendEnvelopeWithMarshalConfig(ctx, wg, pollingLoops, mc, envelope)
			//g.sendEnvelopeWithMarshalConfig(ctx, wg, pollingLoops, mc, envelope)

			if g.debugLevel > 10 {
				log.Printf("performPoll, mc:%v", mc)
			}
			return true
		}
		return true
	})

	wg.Wait()

	// if g.debugLevel > 10 {
	// 	log.Println("performPoll wg.Wait() complete")
	// }

	// Reset prom records and envelope, and return to the sync.Pools
	for _, pc := range envelope.Rows {
		pc.Reset()
		g.GDPRecordPool.Put(pc)
	}
	// g.pC.WithLabelValues("GDPRecordPool", "Put", "count").Add(float64(len(envelope.Rows)))
	envelope.Reset()
	g.GDPEnvelopePool.Put(envelope)
	// g.pC.WithLabelValues("GDPEnvelopePool", "Put", "count").Inc()

	// if g.debugLevel > 10 {
	// 	log.Printf("performPoll complete:%0.4f ms:%d", time.Since(startTime).Seconds(), time.Since(startTime).Microseconds())
	// }

	// // for debugging, stopping here.  please note, if you use a breakpoint, then all the go routines are stopped
	// // so franz-go stops.  so as a hack, just sleeping
	// if pollingLoops == 3 {
	// 	log.Printf("performPoll sleep")
	// 	time.Sleep(12 * time.Hour)
	// }

	return nil
}

// getCountMetrics fetches the metrics from prometheus, and then filters
// them for only the count metrics
func (g *GDP) getCountMetrics(ctx context.Context) ([]string, error) {

	if g.debugLevel > 10 {
		log.Println("getCountMetrics start")
	}

	metrics, err := g.fetchPrometheusMetrics(
		ctx,
		g.Config.RequestConfig.MetricsUrl,
		g.Config.RequestConfig.RequestBackoff.AsDuration(),
		g.Config.RequestConfig.MaxRetries)

	if err != nil {
		if g.debugLevel > 10 {
			log.Printf("Failed to fetch Prometheus metrics: %v", err)
		}
		return nil, err
	}

	counts := g.filterCounts(metrics, g.Config.RequestConfig.AllowPrefix)

	return counts, nil
}

const (
	fetchPrometheusMetricsBackoffCst = 100 * time.Millisecond
)

// fetchPrometheusMetrics is a simple http requester with a timeout, retry logic, and a small
// pause between requests
func (g *GDP) fetchPrometheusMetrics(
	_ context.Context,
	url string,
	timeout time.Duration,
	maxRetries uint32) (string, error) {

	startTime := time.Now()
	defer func() {
		g.pH.WithLabelValues("fetchPrometheusMetrics", "complete", "count").Observe(time.Since(startTime).Seconds())
	}()
	g.pC.WithLabelValues("fetchPrometheusMetrics", "start", "count").Inc()

	client := &http.Client{
		Timeout: timeout,
	}

	var (
		resp *http.Response
		err  error
	)

	for i := uint32(0); i < maxRetries; i++ {

		resp, err = client.Get(url)

		if err == nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				return "", fmt.Errorf("failed to read response body: %w", err)
			}
			g.pC.WithLabelValues("fetchPrometheusMetrics", "bodyBytes", "count").Add(float64(len(bodyBytes)))

			if i == 0 {
				g.pC.WithLabelValues("fetchPrometheusMetrics", "firstShot", "count").Inc()
			}
			g.pC.WithLabelValues("fetchPrometheusMetrics", "success", "count").Inc()
			return string(bodyBytes), nil
		}

		if err != nil {
			if g.debugLevel > 10 {
				log.Printf("Attempt %d failed: %v", i, err)
			}
		} else if resp != nil {
			if g.debugLevel > 10 {
				log.Printf("Attempt %d failed: status code %d", i, resp.StatusCode)
			}
			g.pC.WithLabelValues("fetchPrometheusMetrics", strconv.Itoa(resp.StatusCode), "error").Inc()
		}

		if i < maxRetries-1 {
			time.Sleep(fetchPrometheusMetricsBackoffCst)
		}
		g.pC.WithLabelValues("fetchPrometheusMetrics", "retry", "count").Inc()

	}

	g.pC.WithLabelValues("fetchPrometheusMetrics", "overall", "error").Inc()
	return "", fmt.Errorf("failed to fetch metrics after %d retries: %v", maxRetries, err)
}

func (g *GDP) filterCounts(metrics string, allowPrefix string) (counts []string) {
	var (
		filtered float64
	)
	scanner := bufio.NewScanner(strings.NewReader(metrics))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, allowPrefix) {
			counts = append(counts, line)
			continue
		}
		filtered++
	}

	if err := scanner.Err(); err != nil {
		g.pC.WithLabelValues("filterCounts", "scanner", "error").Inc()
		log.Printf("Error during scanning: %v", err)
	}

	g.pC.WithLabelValues("filterCounts", "counts", "count").Add(float64(len(counts)))
	g.pC.WithLabelValues("filterCounts", "filtered", "count").Add(filtered)

	return counts
}
