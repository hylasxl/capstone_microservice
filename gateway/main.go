package main

import (
	"gateway/middlewares"
	"gateway/routes"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
)

// Middleware to track metrics
func metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/metrics" {
			promhttp.Handler().ServeHTTP(w, r)
			return
		}
		timer := prometheus.NewTimer(httpRequestDuration.WithLabelValues(r.Method, r.URL.Path))
		next.ServeHTTP(w, r)
		timer.ObserveDuration()

		// Record request count
		httpRequestsTotal.WithLabelValues(r.Method, r.URL.Path, http.StatusText(http.StatusOK)).Inc()
	})
}

func main() {
	// Configure router
	router := mux.NewRouter()
	prometheus.MustRegister(httpRequestsTotal, httpRequestDuration)
	router.Use(metricsMiddleware)
	
	router.Handle("/metrics", promhttp.Handler())

	// Set timezone
	if err := os.Setenv("TZ", "Asia/Bangkok"); err != nil {
		log.Fatalf("Failed to set timezone: %v", err)
	}
	time.Local = time.FixedZone("UTC+7", 7*3600)

	clients, err := routes.InitializeServiceClients()
	if err != nil {
		log.Fatalf("Failed to initialize service clients: %v", err)
	}
	defer clients.CloseConnections()

	// Initialize routes
	routes.InitializeRoutes(router, clients)

	corsRouter := middlewares.CorsMiddleware(router)
	// Determine port from environment or use default
	port := os.Getenv("GATEWAY_PORT")
	if port == "" {
		port = "8080"
	}

	// Start the server
	log.Printf("Starting Gateway on port %s", port)
	if err := http.ListenAndServe(":"+port, corsRouter); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
