// Page 317
// Listing 13-19: Imports and command line flags for the metrics example
package metrics

import (
	"flag"

	// You import Go kit's metrics package, which provides the interfaces your
	// code will use, it's prometheus adapter so you can use Prometheus as your
	// metrics platform, and Go's Prometheus client package itself.
	// All Prometheus-related imports reside in this package.
	// The rest of your code will use Go kit's interfaces.
	// This allows you to swap out the underlying metrics platform without the
	// need to change your code's instrumentation.
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	prom "github.com/prometheus/client_golang/prometheus"
)

var (
	// Prometheus prefixes its metrics with a namespace and a subsystem.
	// You could use the service name for the namespace and the node or host name
	// for the subsystem, for example.
	// In this example, you'll use web for the namespace and server1 for the
	// subsystem by default.
	// As a result, your metrics will use the web_server1 prefix.
	Namespace = flag.String("namespace", "web", "metrics namespace")
	Subsystem = flag.String("subsystem", "server1", "metrics subsystem")

	// Page 318
	// Listing 13-20: Creating counters as Go kit interfaces
	// Each counter implements Go kit's metrics.Counter interface.
	// The concrete type for each counter comes from Go kit's prometheus adapter
	// and relies on a CounterOpts struct from the Prometheus client package for
	// configuration.
	// Aside from the namespace and subsystem values we covered, the other
	// important values you set are the metric name and its help string, which
	// describes the metric.
	Requests metrics.Counter = prometheus.NewCounterFrom(
		prom.CounterOpts{
			Namespace: *Namespace,
			Subsystem: *Subsystem,
			Name:      "request_count",
			Help:      "Total requests",
		},
		[]string{},
	)

	WriteErrors metrics.Counter = prometheus.NewCounterFrom(
		prom.CounterOpts{
			Namespace: *Namespace,
			Subsystem: *Subsystem,
			Name:      "write_errors_count",
			Help:      "Total write errors",
		},
		[]string{},
	)

	// Page 318
	// Listing 13-21: Creating a gauge as a Go kit interface
	// Creating a gauge is much like creating a counter.
	// You create a new variable of Go kit's metrics.Gauge interface and use the
	// NewGaugeFrom function from Go kit's prometheus adapter to create the
	// underlying type.
	// The Prometheus client's GaugeOpts struct provides the settings for your
	// new gauge.
	OpenConnections metrics.Gauge = prometheus.NewGaugeFrom(prom.GaugeOpts{
		Namespace: *Namespace,
		Subsystem: *Subsystem,
		Name:      "open_connections",
		Help:      "Current open connections",
	},
		[]string{},
	)

	// Page 319
	// Listing 13-22: Creating a histogram metric.
	// Both the summary and histogram metric types implement Go kit's
	// metrics.Histogram interface from its prometheus adapter.
	// Here, you're using a histogram metric type, using the Prometheus client's
	// HistogramOpts struct for configuration.
	// Since Prometheus's default bucket sizes are too large for the expected
	// request duration range when communicating over localhost, you define
	// custom bucket sizes.
	RequestDuration metrics.Histogram = prometheus.NewHistogramFrom(
		prom.HistogramOpts{
			Namespace: *Namespace,
			Subsystem: *Subsystem,
			Buckets: []float64{
				0.0000001, 0.0000002, 0.0000003, 0.0000004, 0.0000005,
				0.000001, 0.0000025, 0.000005, 0.0000075,
				0.00001, 0.0001, 0.001, 0.01,
			},
			Name: "request_duration_histogram_seconds",
			Help: "Total duration of all requests",
		},
		[]string{},
	)

	// Page 320
	// Listing 13-23: Optionally creating a summary metric.
	RequestDurationSummary metrics.Histogram = prometheus.NewSummaryFrom(
		prom.SummaryOpts{
			Namespace: *Namespace,
			Subsystem: *Subsystem,
			Name: "request_duration_summary_seconds",
			Help: "Total duration of all requests",
		},
		[]string{},
	)
)
