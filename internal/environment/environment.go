package environment

import (
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

const (
	envFileDevelopment  = ".env.development"
	envFileProduction   = ".env.production"
	defaultFrontendPath = "./web/build"
)

var errNoBuildDirectory = errors.New("build directory does not exist, run `npm install` and `npm run build` in the web directory")

func LoadEnvironmentVariables() {
	if err := loadConfigs(); err != nil {
		if errors.Is(err, errNoBuildDirectory) {
			log.Fatal("Environment:", err)
		}

		log.Println("Environment: Failed to find config in CWD, changing CWD to executable path")

		executablePath, executableErr := os.Executable()
		if executableErr != nil {
			log.Fatal("Environment:", executableErr)
		}

		if chdirErr := os.Chdir(filepath.Dir(executablePath)); chdirErr != nil {
			log.Fatal("Environment:", chdirErr)
		}

		if retryErr := loadConfigs(); retryErr != nil {
			log.Fatal("Environment:", retryErr)
		}
	}

	setDefaultEnvironmentVariables()
}

func loadConfigs() error {
	if os.Getenv(APP_ENV) == "development" {
		log.Println("Environment: Loading `" + envFileDevelopment + "`")
		return godotenv.Load(envFileDevelopment)
	}

	log.Println("Environment: Loading `" + envFileProduction + "`")
	if err := godotenv.Load(envFileProduction); err != nil {
		return err
	}

	if os.Getenv(FRONTEND_DISABLED) == "" {
		if _, err := os.Stat(GetFrontendPath()); os.IsNotExist(err) {
			return errNoBuildDirectory
		} else if err != nil {
			return err
		}
	}

	return nil
}

func GetFrontendPath() string {
	frontendPath := os.Getenv(FRONTEND_PATH)
	if frontendPath == "" {
		return defaultFrontendPath
	}

	return frontendPath
}

func setDefaultEnvironmentVariables() {
	if os.Getenv(STREAM_PROFILE_PATH) == "" {
		log.Println("Environment: Setting STREAM_PROFILE_PATH: profiles")
		err := os.Setenv(STREAM_PROFILE_PATH, "profiles")
		if err != nil {
			log.Panic("Error setting default value for STREAM_PROFILE_PATH")
		}
	}
}
