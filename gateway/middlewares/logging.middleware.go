package middlewares

import (
	"log"
	"net/http"
	"os"
	"time"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		log.Println("Logging middleware triggered") // Debugging

		logEntry := time.Now().Format("2006-01-02 15:04:05") + " | " + r.Method + " " + r.RequestURI + " from " + r.RemoteAddr
		logToFile("logging/request.log", logEntry)

		next.ServeHTTP(w, r)

		durationEntry := "Completed in " + time.Since(start).String()
		logToFile("logging/request.log", durationEntry)
	})
}

func logToFile(filename, entry string) {
	// Ensure the logging directory exists
	dir := "logging"
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		log.Println("🚨 Error creating logging directory:", err)
		return
	}

	// Ensure the log file exists
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Println("🚨 Error opening log file:", err)
		return
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	// Write log entry
	logger := log.New(file, "", 0)
	logger.Println(entry)

	err = file.Sync()
	if err != nil {
		return
	}

	// Debugging
	log.Println("✅ Successfully logged:", entry)
}
