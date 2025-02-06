package routes

import (
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func InitMonitoringRoutes(router *mux.Router) {
	monitoringRoute := router.PathPrefix("/monitoring").Subrouter()
	monitoringRoute.Handle("/metrics", promhttp.Handler())
}
