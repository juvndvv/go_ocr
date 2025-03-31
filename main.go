package main

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	uniPdfLicense "github.com/unidoc/unipdf/v3/common/license"
	"go_ocr/src/ocr/ai"
	"go_ocr/src/ocr/logger"
	"go_ocr/src/ocr/pdf_extractor"
	"go_ocr/src/ocr/pdf_extractor/downloader"
	"net/http"
	"os"
	"time"
)

var (
	log = logger.NewLogger(false)
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("No se encontró archivo .env")
	}

	licenseKey := os.Getenv("UNIPDF_LICENSE_KEY")
	err = uniPdfLicense.SetMeteredKey(licenseKey)
	if err != nil {
		log.Fatal("Error al configurar licencia de UniPDF: %v", err)
	}

	// Configurar logger
	log.Info("Starting OCR Server")
	log.Debug("Environment: %s", os.Getenv("ENV"))

	// Configurar handler
	http.HandleFunc("/convert", convertHandler)

	// Configurar servidor
	port := ":" + os.Getenv("APP_PORT")
	server := &http.Server{
		Addr:         port,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	log.Info("Servidor escuchando en http://localhost%s", port)
	log.Fatal(server.ListenAndServe().Error())
}

func convertHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := time.Now().UnixNano()
}
