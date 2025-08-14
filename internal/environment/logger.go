package environment

import (
	"io"
	"log"
	"os"
	"strings"
	"time"
)

func SetupLogger() {
	if strings.EqualFold(os.Getenv(LOGGING_ENABLED), "false") {
		return
	}

	logDir := "logs"
	if envLogDir := os.Getenv(LOGGING_DIRECTORY); envLogDir != "" {
		logDir = envLogDir
	}

	logFileName := time.Now().Format("log_20060102")

	if envLogFileIsSingleFile := strings.EqualFold(os.Getenv(LOGGING_SINGLEFILE), "true"); envLogFileIsSingleFile == true {
		logFileName = "log"
	}

	logFilePath := logDir + "/" + logFileName

	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		log.Fatalf("Failed to create log directory: %v", err)
	}

	var logFile *os.File
	var err error

	if envLogTruncateExistingFile := strings.EqualFold(os.Getenv(LOGGING_NEW_FILE_ON_STARTUP), "true"); envLogTruncateExistingFile == true {
		logFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	} else {
		logFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	}

	if err != nil {
		log.Fatalf("Could not open log file %v", err)
	}

	// Setup output to be directed to both logfile and console
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)
}
