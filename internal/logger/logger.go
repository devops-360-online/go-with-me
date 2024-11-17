package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

var (
	logFile *os.File
)

// InitLogger initializes the global logger
func InitLogger() error {
	var err error
	logFile, err = os.OpenFile("/var/log/app/app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}
	log.SetOutput(logFile)
	return nil
}

// CloseLogger closes the log file
func CloseLogger() {
	if logFile != nil {
		logFile.Close()
	}
}

// LogMessage logs a structured message with additional fields
func LogMessage(level string, message string, traceID string, additionalFields map[string]interface{}) {
	logEntry := map[string]interface{}{
		"timestamp": time.Now().Format("2006-01-02T15:04:05.999999999Z07:00"),
		"level":     level,
		"message":   message,
		"trace_id":  traceID,
	}
	for k, v := range additionalFields {
		logEntry[k] = v
	}
	logData, _ := json.Marshal(logEntry)
	fmt.Fprintln(logFile, string(logData))
}
