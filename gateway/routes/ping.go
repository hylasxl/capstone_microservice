package routes

import (
	"github.com/gorilla/mux"
	"net/http"
)

func InitializePingRoutes(router *mux.Router) {
	router.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("pong"))
	}).Methods("GET")
}
