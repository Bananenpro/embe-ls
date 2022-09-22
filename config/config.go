package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
)

var (
	LogFile     string
	LogLevel    string
	GLSPLogFile *string
)

type Config struct {
	LogFile     string  `json:"log_file"`
	LogLevel    string  `json:"log_level"`
	GLSPLogFile *string `json:"lsp_log_file"`
}

func init() {
	path := filepath.Join(xdg.ConfigHome, "embe-ls", "config.json")
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	var config Config
	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to decode config file: %s", err)
		return
	}

	LogFile = config.LogFile
	LogLevel = config.LogLevel
	GLSPLogFile = config.GLSPLogFile
	if GLSPLogFile != nil && *GLSPLogFile == "" {
		GLSPLogFile = nil
	}

	if GLSPLogFile != nil {
		os.Remove(*GLSPLogFile)
	}
}
