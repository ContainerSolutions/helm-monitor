package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

var (
	addr           = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")
	prometheusAddr = flag.String("prometheus-address", ":9101", "The address to listen on for Prometheus metrics requests.")

	inFlightGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "in_flight_requests",
			Help: "A gauge of requests currently being served by the wrapped handler.",
		},
	)

	counter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "A counter for requests to the wrapped handler.",
		},
		[]string{"code", "method"},
	)

	duration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "request_duration_seconds",
			Help:    "A histogram of latencies for requests.",
			Buckets: []float64{.25, .5, 1, 2.5, 5, 10},
		},
		[]string{"code", "method"},
	)

	responseSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "response_size_bytes",
			Help:    "A histogram of response sizes for requests.",
			Buckets: []float64{200, 500, 900, 1500},
		},
		[]string{"code", "method"},
	)
)

func init() {
	prometheus.MustRegister(inFlightGauge, counter, duration, responseSize)

}

func promRequestHandler(handler http.Handler) http.Handler {
	return promhttp.InstrumentHandlerInFlight(inFlightGauge,
		promhttp.InstrumentHandlerDuration(duration,
			promhttp.InstrumentHandlerCounter(counter,
				promhttp.InstrumentHandlerResponseSize(responseSize, handler),
			),
		),
	)
}

func accessLogger(r *http.Request, status, size int, dur time.Duration) {
	hlog.FromRequest(r).Info().
		Str("host", r.Host).
		Int("status", status).
		Int("size", size).
		Dur("duration_ms", dur).
		Msg("request")
}

func main() {
	flag.Parse()

	log := zerolog.New(os.Stdout).With().
		Timestamp().
		Logger()

	go func() {
		log.Info().Msgf("Serving Prometheus metrics on port %s", *prometheusAddr)

		http.Handle("/metrics", promhttp.Handler())

		if err := http.ListenAndServe(*prometheusAddr, nil); err != nil {
			log.Error().Err(err).Msg("Starting Prometheus listener failed")
		}
	}()

	c := alice.New(hlog.NewHandler(log), hlog.AccessHandler(accessLogger))

	r := mux.NewRouter()

	s := r.Methods("GET").Subrouter()

	s.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello")
		return
	})

	log.Info().Msgf("Serving application on port %s", *addr)
	if err := http.ListenAndServe(*addr, c.Then(promRequestHandler(r))); err != nil {
		log.Fatal().Err(err).Msg("Startup failed")
	}
}
