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
		return
	}

	log.Println("Could not find any environment files")
	os.Exit(0)
}

func loadEnvironmentFile(filePath string) {
	currentWorkingDirectory, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	path := filepath.Join(currentWorkingDirectory, filePath)

	if _, err := os.Stat(path); err == nil {
		err := godotenv.Overload(path)

		if err != nil {
			log.Println("Error occurred loading environment file", path)
			log.Println(err)

			os.Exit(0)
		}

		log.Println("Loaded", filePath)
	}
}
