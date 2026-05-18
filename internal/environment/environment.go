package environment

import (
	"errors"
	"log/slog"
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
			slog.Error("Environment", "err", err)
			os.Exit(1)
		}

		slog.Info("Environment: Failed to find config in CWD, changing CWD to executable path")

		executablePath, executableErr := os.Executable()
		if executableErr != nil {
			slog.Error("Environment:", "err", executableErr)
			os.Exit(1)
		}

		if chdirErr := os.Chdir(filepath.Dir(executablePath)); chdirErr != nil {
			slog.Error("Environment:", "err", chdirErr)
			os.Exit(1)
		}

		if retryErr := loadConfigs(); retryErr != nil {
			slog.Error("Environment:", "err", retryErr)
			os.Exit(1)
		}
	}

	setDefaultEnvironmentVariables()
}

func loadConfigs() error {
	if os.Getenv(appEnv) == "development" {
		return loadOptionalEnvironmentFile(envFileDevelopment)
	}

	if err := loadOptionalEnvironmentFile(envFileProduction); err != nil {
		return err
	}

	if os.Getenv(FrontendDisabled) == "" {
		if _, err := os.Stat(GetFrontendPath()); os.IsNotExist(err) {
			return errNoBuildDirectory
		} else if err != nil {
			return err
		}
	}

	return nil
}

func loadOptionalEnvironmentFile(fileName string) error {
	if _, err := os.Stat(fileName); errors.Is(err, os.ErrNotExist) {
		slog.Info("Environment file not found, continuing with system environment", "fileName", fileName)
		return nil
	} else if err != nil {
		return err
	}

	slog.Info("Environment: Loading `" + fileName + "`")
	if err := godotenv.Load(fileName); err != nil {
		return err
	}

	return nil
}

func GetFrontendPath() string {
	frontendPath := os.Getenv(frontendPath)
	if frontendPath == "" {
		return defaultFrontendPath
	}

	return frontendPath
}

func setDefaultEnvironmentVariables() {
	if os.Getenv(StreamProfilePath) == "" {
		slog.Info("Environment: Setting STREAM_PROFILE_PATH: profiles")
		err := os.Setenv(StreamProfilePath, "profiles")
		if err != nil {
			slog.Error("Error setting default value for STREAM_PROFILE_PATH")
			os.Exit(1)
		}
	}
}
