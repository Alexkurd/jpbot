package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Define a counter metric
var (
	requestsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "jpbot_requests_total",
		Help: "The total number of requests received",
	})
	triggersTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "jpbot_triggers_total",
		Help: "The total number of triggers received",
	})
	triggersPrivateTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "jpbot_triggers_private_total",
		Help: "The total number of personal triggers received",
	})
	triggersPublicTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "jpbot_triggers_public_total",
		Help: "The total number of personal triggers received",
	})
	triggerUsage = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "jpbot_triggers_usage_total",
			Help: "Total number of times each term has been used.",
		},
		[]string{"term", "private"},
	)
	apiUsage = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "jpbot_api_usage_total",
			Help: "Total number of apiBan requests.",
		},
		[]string{"service"},
	)
	apiRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "jpbot_api_requests_total",
			Help: "Total number of requests to the external API",
		},
		[]string{"service", "status"},
	)
	apiRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "jpbot_api_request_duration_seconds",
			Help:    "Duration of external API requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "status"},
	)
	tgRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "jpbot_tg_requests_total",
			Help: "Total number of requests to the external API",
		},
		[]string{"method", "status"},
	)
)

// HealthCheckHandler handles the /health endpoint
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
}

// MetricsHandler handles the /metrics endpoint
func MetricsHandler(w http.ResponseWriter, r *http.Request) {
	promhttp.Handler().ServeHTTP(w, r)
}

func recordTriggerUsage(trigger string, private string) {
	// Increment the counter for the given term.
	triggerUsage.WithLabelValues(trigger, private).Inc()
}

//// RootHandler handles the root endpoint
//func RootHandler(w http.ResponseWriter, r *http.Request) {
//	requestsTotal.Inc()
//	w.WriteHeader(http.StatusOK)
//	fmt.Fprintf(w, "Welcome to the Go HTTP server!")
//}

func initMetrics() {
	prometheus.MustRegister(triggerUsage)
	prometheus.MustRegister(apiRequestsTotal)
	prometheus.MustRegister(apiRequestDuration)
	prometheus.MustRegister(tgRequestsTotal)
	// Register the metrics handler
	http.HandleFunc("/metrics", MetricsHandler)

	// Register the health check handler
	http.HandleFunc("/health", HealthCheckHandler)

	// Start the HTTP server
	server := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	fmt.Println("Server is running on http://localhost:8080")
	if err := server.ListenAndServe(); err != nil {
		fmt.Printf("Error starting server: %s\n", err)
	}
}
