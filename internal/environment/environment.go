package environment

import (
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

func LoadEnvironmentVariables() {
	files := []string{
		".env.development",
		".env.production",
	}

	// Load base environment file if available
	loadEnvironmentFile(".env")

	for _, file := range files {
		loadEnvironmentFile(file)
		setDefaultEnvironmentVariables()
		return
	}

	log.Println("Environment: Could not find any environment files")
	os.Exit(0)
}

func loadEnvironmentFile(filePath string) {
	currentWorkingDirectory, err := os.Getwd()
	if err != nil {
		log.Fatal("Environment:", err)
	}

	path := filepath.Join(currentWorkingDirectory, filePath)

	if _, err := os.Stat(path); err == nil {
		err := godotenv.Overload(path)

		if err != nil {
			log.Println("Environment: Error occurred loading environment file", path)
			log.Println(err)

			os.Exit(0)
		}

		log.Println("Environment: Loaded", filePath)
	}
}

func setDefaultEnvironmentVariables() {
	if os.Getenv(STREAM_PROFILE_PATH) == "" {
		log.Println("Environment: Setting STREAM_PROFILE_PATH: profiles")
		os.Setenv(STREAM_PROFILE_PATH, "profiles")
	}
}
