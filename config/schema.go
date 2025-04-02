package config

import (
	"encoding/json"
	"os"
	"sync"
)

var (
	config Schema
	once   sync.Once
	mu     sync.RWMutex
)

type Schema struct {
	Name       string `json:"name"`
	OcrFeature bool   `json:"ocr_feature"`
}

// GetConfig devuelve una copia segura de la configuración
func GetConfig() Schema {
	mu.RLock()
	defer mu.RUnlock()
	return config
}

// LoadConfig carga la configuración una sola vez
func LoadConfig(filePath string) error {
	var initErr error

	once.Do(func() {
		file, err := os.ReadFile(filePath)
		if err != nil {
			initErr = err
			return
		}

		mu.Lock()
		defer mu.Unlock()

		initErr = json.Unmarshal(file, &config)
	})

	return initErr
}
