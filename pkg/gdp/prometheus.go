package gdp

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func (g *GDP) InitPromethus(wg *sync.WaitGroup) {

	defer wg.Done()

	g.pC = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: "gdp",
			Name:      "counts",
			Help:      "gdp counts",
		},
		[]string{"function", "variable", "type"},
	)

	g.pH = promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Subsystem: "gdp",
			Name:      "histograms",
			Help:      "gdp historgrams",
			Objectives: map[float64]float64{
				0.1:  quantileError,
				0.5:  quantileError,
				0.99: quantileError,
			},
			MaxAge: summaryVecMaxAge,
		},
		[]string{"function", "variable", "type"},
	)

	g.pG = promauto.NewGauge(
		prometheus.GaugeOpts{
			Subsystem: "gdp",
			Name:      "gauge",
			Help:      "gdp gauge",
		},
	)

}
