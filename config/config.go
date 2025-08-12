package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// LoadConfig loads configuration from .env and .ini files
func LoadConfig(envPath string) error {

	// Check if envPath exists and is a file
	fileInfo, err := os.Stat(envPath)
	if err != nil || fileInfo.IsDir() {
		// If envPath doesn't exist or is a directory, append .env to it
		newPath := filepath.Join(envPath, ".env")
		if newFileInfo, newErr := os.Stat(newPath); newErr == nil && !newFileInfo.IsDir() {
			// Use the new path only if it exists and is a file
			envPath = newPath
		}
	}

	// Load environment variables if the file exists
	if envFileInfo, envErr := os.Stat(envPath); envErr == nil && !envFileInfo.IsDir() {
		if err := godotenv.Load(envPath); err != nil {
			return fmt.Errorf("error loading .env file: %w", err)
		}
	}
	return nil
}
func GetStringValue(key, defaultValue string) string {
	envValue := os.Getenv(key)
	if envValue != "" {
		return envValue
	}
	return defaultValue
}
