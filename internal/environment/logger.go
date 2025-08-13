package environment

import (
	"io"
	"log"
	"os"
)

func SetupLogger() {
	// TODO: Setup envvar for log file enabled and path

	logDir := "logs"
	logFilePath := logDir + "/log"

	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		log.Fatalf("Failed to create log directory: %v", err)
	}

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	if err != nil {
		log.Fatalf("Could not open log file %v", err)
	}

	// Setup output to be directed to both logfile and console
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)
}
