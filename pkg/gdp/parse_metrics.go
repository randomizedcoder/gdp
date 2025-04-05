package gdp

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	gdpp "github.com/randomizedcoder/gdp/pkg/prometheus"
)

const (
	regexMatchesCst = 4 + 1
	splitPartsCst   = 2
)

var (
	re = regexp.MustCompile(`function="([^"]+)",type="([^"]+)",variable="([^"]+)"\} ([^\s]+)$`)
)

func (g *GDP) parseMetricLinesRegex(lines []string, t time.Time, pollingLoops uint64) (*gdpp.Envelope, error) {

	e := g.GDPEnvelopePool.Get().(*gdpp.Envelope)
	// g.pC.WithLabelValues("GDPEnvelopePool", "Get", "count").Inc()

	for i, line := range lines {

		pc, err := g.parseMetricLineRegex(uint64(i), line)
		if err != nil {
			return e, err
		}

		pc.TimestampNs = float64(t.UnixNano()) / 1e9 // storing seconds, with nanosecond preciscion
		pc.Hostname = g.hostname
		pc.Label = g.Config.Label
		pc.Tag = g.Config.Tag
		pc.PollCounter = pollingLoops

		e.Rows = append(e.Rows, pc)
	}
	return e, nil

}

// xtcp_counts{function="Netlinker",type="error",variable="ParseNLPacket"} 1775
// xtcp_counts{function="Poller",type="bytes",variable="Destination"} 16687

// parseMetricLineRegex parses a Prometheus metric line using regular expressions.
func (g *GDP) parseMetricLineRegex(i uint64, line string) (*gdpp.Envelope_PromRecordCounter, error) {

	// if g.debugLevel > 10 {
	// 	log.Printf("parseMetricLineRegex:%v", line)
	// }

	pc := g.GDPRecordPool.Get().(*gdpp.Envelope_PromRecordCounter)
	// g.pC.WithLabelValues("GDPRecordPool", "Get", "count").Inc()

	matches := re.FindStringSubmatch(line)

	// if g.debugLevel > 10 {
	// 	for k, v := range matches {
	// 		log.Printf("%d. %s\n", k, v)
	// 	}
	// }

	if len(matches) != regexMatchesCst {
		return pc, fmt.Errorf("invalid metric line format")
	}

	value, err := strconv.ParseFloat(matches[4], 64)
	if err != nil {
		return pc, fmt.Errorf("invalid float value: %w", err)
	}

	pc.RecordCounter = i
	pc.Function = matches[1]
	pc.Type = matches[2]
	pc.Variable = matches[3]
	pc.Value = value

	// if g.debugLevel > 10 {
	// 	log.Printf("pc.Function:%v", pc.Function)
	// 	log.Printf("pc.Type:%v", pc.Type)
	// 	log.Printf("pc.Variable:%v", pc.Variable)
	// 	log.Printf("pc.Value:%v", pc.Value)
	// }

	return pc, nil
}

// Don't use my crappy string parsing function
// My function is faster on a single line, probably because the branch predictor gets good
// but for lots of different lines, it's not as good
// And given the mental complexity overhead, regex is a better trade off

// [das@t:~/Downloads/gdp/pkg/gdp]$ go test -bench=.
// goos: linux
// goarch: amd64
// pkg: github.com/randomizedcoder/gdp/pkg/gdp
// cpu: Intel(R) Core(TM) i9-10885H CPU @ 2.40GHz
// BenchmarkParseMetricLineRegex-16          747507              1504 ns/op
// BenchmarkParseMetricLineString-16         958314              1296 ns/op
// BenchmarkParseMetricLinesRegex-16          20839             58242 ns/op
// BenchmarkParseMetricLinesString-16         20020             61310 ns/op    <--- slower!
// PASS
// ok      github.com/randomizedcoder/gdp/pkg/gdp        6.092s

// xtcp_counts{function="Netlinker",type="error",variable="ParseNLPacket"} 1775
// xtcp_counts{function="Poller",type="bytes",variable="Destination"} 16687

// parseMetricLineString parses a Prometheus metric line using string manipulation.
func (g *GDP) parseMetricLineString(i uint64, line string) (*gdpp.Envelope_PromRecordCounter, error) {

	// if g.debugLevel > 10 {
	// 	log.Printf("parseMetricLineString:%v", line)
	// }

	pc := g.GDPRecordPool.Get().(*gdpp.Envelope_PromRecordCounter)
	// g.pC.WithLabelValues("GDPRecordPool", "Get", "count").Inc()

	var captures [4]string
	captureIndex := 0
	capture := false
	var spaceIndex int

	for i, char := range line {
		// if g.debugLevel > 10 {
		// 	log.Printf("parseMetricLineString i:%d char:%v captures[%d]:%v", i, char, captureIndex, captures[captureIndex])
		// }

		if char == '"' {
			if capture {
				capture = false
				captureIndex++
			} else {
				capture = true
			}
			continue
		}

		if capture {
			captures[captureIndex] += string(char)
			continue
		}

		if char == ' ' {
			spaceIndex = i
			break
		}

	}

	captures[3] = line[spaceIndex+1:]

	value, err := strconv.ParseFloat(captures[3], 64)
	if err != nil {
		return pc, fmt.Errorf("invalid float value: %w", err)
	}

	pc.RecordCounter = i
	pc.Function = captures[0]
	pc.Type = captures[1]
	pc.Variable = captures[2]
	pc.Value = value

	// if g.debugLevel > 10 {
	// 	log.Printf("pc.Function:%v", pc.Function)
	// 	log.Printf("pc.Type:%v", pc.Type)
	// 	log.Printf("pc.Variable:%v", pc.Variable)
	// 	log.Printf("pc.Value:%v", pc.Value)
	// }

	return pc, nil
}
