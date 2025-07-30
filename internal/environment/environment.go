package environment

import (
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

func LoadEnvironmentVariables() {
	files := []string{
		".env",
		".env.development",
	}

	currentWorkingDirectory, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		path := filepath.Join(currentWorkingDirectory, file)

		if _, err := os.Stat(path); err == nil {
			err := godotenv.Load(path)

			if err != nil {
				log.Println("Error occurred loading environment file", path)
				log.Println(err)
				os.Exit(0)
			}

			log.Println("Loaded", file)
			return
		}
	}

	log.Println("Could not find any environment files")
	os.Exit(0)
}
