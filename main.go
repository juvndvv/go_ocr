package main

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"go_ocr/services/ai"
	"go_ocr/services/logger"
	"go_ocr/services/pdf_extractor"
	"go_ocr/services/pdf_extractor/downloader"
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

	log.Info("[Request:%d] New request received - Método: %s - URL: %s",
		requestID, r.Method, r.URL.String())
	log.Debug("[Request:%d] Headers: %v", requestID, r.Header)

	// Validar método HTTP
	if r.Method != http.MethodPost {
		errMsg := "Method not allowed"
		log.Warning("[Request:%d] %s", requestID, errMsg)
		http.Error(w, errMsg, http.StatusMethodNotAllowed)
		return
	}

	// Obtener URL del PDF
	url := r.FormValue("url")
	if url == "" {
		errMsg := "Se requiere el parámetro 'url'"
		log.Warning("[Request:%d] %s", requestID, errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	log.Info("[Request:%d] Procesando PDF desde URL: %s", requestID, url)

	// Descargar el PDF
	filePath, err := downloader.DownloadPDF(url)
	if err != nil {
		errMsg := fmt.Sprintf("Error al descargar PDF: %v", err)
		log.Error("[Request:%d] %s", requestID, errMsg)
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}
	log.Info("[Request:%d] PDF descargado en: %s", requestID, filePath)
	defer func() {
		downloader.CleanupFile(filePath)
	}()

	// Extraer texto
	text, err := pdf_extractor.ExtractTextFromPDF(filePath)
	if err != nil {
		errMsg := fmt.Sprintf("Error al extraer texto: %v", err)
		log.Error("[Request:%d] %s", requestID, errMsg)
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	// Extraer datos estructurados
	payrollData, err := ai.ExtractPayrollData(text)
	if err != nil {
		errMsg := fmt.Sprintf("Error al extraer datos: %v", err)
		log.Error("[Request:%d] %s", requestID, errMsg)
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	log.Info("[Request:%d] Datos extraídos exitosamente", requestID)
	log.Debug("[Request:%d] Datos extraídos: %+v", requestID, payrollData)

	// Convertir a JSON
	responseJSON, err := json.Marshal(payrollData)
	if err != nil {
		errMsg := fmt.Sprintf("Error al convertir a JSON: %v", err)
		log.Error("[Request:%d] %s", requestID, errMsg)
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	// Configurar headers y enviar respuesta
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(responseJSON); err != nil {
		log.Error("[Request:%d] Error al escribir respuesta: %v", requestID, err)
	} else {
		log.Info("[Request:%d] Respuesta enviada exitosamente. Tiempo total: %v",
			requestID, time.Since(startTime))
	}
}
