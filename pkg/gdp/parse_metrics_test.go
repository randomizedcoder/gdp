package gdp

import (
	"bufio"
	"os"
	"testing"

	gdpp "github.com/randomizedcoder/gdp/pkg/prometheus"
)

func TestParseMetricLine(t *testing.T) {
	lines := []string{
		`xtcp_counts{function="Deserialize",type="count",variable="start"} 51846`,
		`xtcp_counts{function="GetNetlinkSocketFDs",type="counter",variable="start"} 8344`,
		`xtcp_counts{function="Netlinker",type="count",variable="RecvfromCalls"} 6.1467395368e+10`,
		`xtcp_counts{function="Netlinker",type="count",variable="Timeout"} 1.6404240858e+10`,
	}

	expected := []*gdpp.Envelope_PromRecordCounter{
		{Function: "Deserialize", Type: "count", Variable: "start", Value: 51846},
		{Function: "GetNetlinkSocketFDs", Type: "counter", Variable: "start", Value: 8344},
		{Function: "Netlinker", Type: "count", Variable: "RecvfromCalls", Value: 6.1467395368e+10},
		{Function: "Netlinker", Type: "count", Variable: "Timeout", Value: 1.6404240858e+10},
	}

	g := newGDPforTests()

	for i, line := range lines {
		resultRegex, errRegex := g.parseMetricLineRegex(uint64(i), line)

		t.Logf("Line: %s", line)

		if errRegex != nil {
			t.Fatalf("Regex: Error parsing line %d: %v", i, errRegex)
		}
		resultString, errString := g.parseMetricLineString(uint64(i), line)
		if errString != nil {
			t.Fatalf("String: Error parsing line %d: %v", i, errString)
		}

		if resultRegex.Function != expected[i].Function ||
			resultRegex.Type != expected[i].Type ||
			resultRegex.Variable != expected[i].Variable ||
			resultRegex.Value != expected[i].Value {
			t.Fatalf("Regex: Line %d, expected %+v, got %+v", i, expected[i], resultRegex)
		}

		if resultString.Function != expected[i].Function ||
			resultString.Type != expected[i].Type ||
			resultString.Variable != expected[i].Variable ||
			resultString.Value != expected[i].Value {
			t.Fatalf("String: Line %d, expected %+v, got %+v", i, expected[i], resultString)
		}

	}
}

func BenchmarkParseMetricLineRegex(b *testing.B) {
	g := newGDPforTests()
	line := `xtcp_counts{function="Netlinker",type="count",variable="Timeout"} 1.6404240858e+10`
	var results []*gdpp.Envelope_PromRecordCounter
	for i := 0; i < b.N; i++ {
		result, _ := g.parseMetricLineRegex(uint64(i), line)
		results = append(results, result)
	}
	_ = results
}

func BenchmarkParseMetricLineString(b *testing.B) {
	g := newGDPforTests()
	line := `xtcp_counts{function="Netlinker",type="count",variable="Timeout"} 1.6404240858e+10`
	var results []*gdpp.Envelope_PromRecordCounter
	for i := 0; i < b.N; i++ {
		result, _ := g.parseMetricLineString(uint64(i), line)
		results = append(results, result)
	}
	_ = results
}

func TestParseMetricLines(t *testing.T) {
	g := newGDPforTests()
	lines := readTestData(t)

	for i, line := range lines {

		t.Logf("Line: %s", line)

		_, errRegex := g.parseMetricLineRegex(uint64(i), line)
		if errRegex != nil {
			t.Fatalf("Regex: Error parsing line: %v", errRegex)
		}
		_, errString := g.parseMetricLineString(uint64(i), line)
		if errString != nil {
			t.Fatalf("String: Error parsing line: %v", errString)
		}
	}
}

func BenchmarkParseMetricLinesRegex(b *testing.B) {
	g := newGDPforTests()
	lines := readTestData(b)
	var results []*gdpp.Envelope_PromRecordCounter
	for i := 0; i < b.N; i++ {
		for _, line := range lines {
			result, _ := g.parseMetricLineRegex(uint64(i), line)
			results = append(results, result)
		}
	}
	_ = results
}

func BenchmarkParseMetricLinesString(b *testing.B) {
	g := newGDPforTests()
	lines := readTestData(b)
	var results []*gdpp.Envelope_PromRecordCounter
	for i := 0; i < b.N; i++ {
		for _, line := range lines {
			result, _ := g.parseMetricLineString(uint64(i), line)
			results = append(results, result)
		}
	}
	_ = results
}

const (
	testDataFileNameCst = "test_data/metrics_lines"
)

func readTestData(t testing.TB) []string {
	file, err := os.Open(testDataFileNameCst)
	if err != nil {
		t.Fatalf("Failed to open test data: %v", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("Error scanning test data: %v", err)
	}

	return lines
}
