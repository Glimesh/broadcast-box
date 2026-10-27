package environment

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	currentDate string
	logMutex    sync.Mutex
)

func SetupLogger() {
	if strings.EqualFold(os.Getenv(loggingEnabled), "false") {
		return
	}

	startLogRotation()
}

func setupLoggerForDate(date string) {
	logFile, err := getLogFileWriter()
	if err != nil {
		log.Printf("Failed to open log file: %v", err)
		return
	}

	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)
	currentDate = date
}

func startLogRotation() {
	go func() {
		for {
			now := time.Now().Format("20060102")
			logMutex.Lock()
			if now != currentDate {
				setupLoggerForDate(now)
			}
			logMutex.Unlock()
			time.Sleep(1 * time.Minute)
		}
	}()
}

func GetLogFileReader() (logFile *os.File, err error) {
	logDir, _, _ := getLogfilePath()
	logFilePath, err := getLatestLogFile(logDir)
	if err != nil {
		log.Println("Logger Error:", err)
	}

	file, err := os.Open(logFilePath)
	if err != nil {
		log.Println("Logger Error:", err)
	}

	return file, err
}

func getLogFileWriter() (logFile *os.File, err error) {
	logDir, _, logFilePath := getLogfilePath()

	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		log.Fatalf("Failed to create log directory: %v", err)
	}

	if envLogTruncateExistingFile := strings.EqualFold(os.Getenv(loggingNewFileOnStartup), "true"); envLogTruncateExistingFile {
		logFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	} else {
		logFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	}

	if err != nil {
		log.Println("Logger Error: FilePath", logFilePath)
		log.Fatalf("Logger Error: %v", err)
		return nil, err
	}

	return logFile, nil
}

func getLogfilePath() (directory string, fileName string, logFilePath string) {
	logDir := "logs"
	if envLogDir := os.Getenv(loggingDirectory); envLogDir != "" {
		logDir = envLogDir
	}

	logFileName := time.Now().Format("20060102")

	if envLogFileIsSingleFile := strings.EqualFold(os.Getenv(loggingSingleFile), "true"); envLogFileIsSingleFile {
		logFileName = "log"
	}

	return logDir, logFileName, logDir + "/" + logFileName
}

func getLatestLogFile(logDir string) (string, error) {
	var dates []time.Time
	var fileMap = make(map[time.Time]string)

	err := filepath.WalkDir(logDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || strings.Contains(d.Name(), ".") {
			return nil
		}

		t, err := time.Parse("20060102", d.Name())
		if err != nil {
			return nil
		}

		dates = append(dates, t)
		fileMap[t] = path
		return nil
	})

	if err != nil {
		return "", err
	}

	if len(dates) == 0 {
		return "", fmt.Errorf("no log files found")
	}

	sort.Slice(dates, func(i, j int) bool {
		return dates[i].After(dates[j])
	})

	latest := dates[0]
	return fileMap[latest], nil
}
