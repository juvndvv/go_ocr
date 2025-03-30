package downloader

import (
	"fmt"
	"go_ocr/src/services/logger"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	// Logger compartido para el paquete
	log = logger.NewLogger(false) // Cambia a true para escribir en archivo
)

func init() {
	// Opcional: Habilitar debug en desarrollo
	if os.Getenv("ENV") == "development" {
		log.EnableDebug()
	}
}

func DownloadPDF(url string) (string, error) {
	startTime := time.Now()
	log.Info("Iniciando descarga de PDF desde URL: %s", url)
	log.Debug("Parámetros de DownloadPDF - url: %s", url)

	// Verificar que sea una URL de PDF
	if !strings.HasSuffix(strings.ToLower(url), ".pdf") {
		errMsg := fmt.Sprintf("URL no es un PDF: %s", url)
		log.Error(errMsg)
		return "", fmt.Errorf("la URL no apunta a un PDF")
	}

	resp, err := http.Get(url)
	if err != nil {
		log.Error("Error al descargar PDF: %v", err)
		return "", fmt.Errorf("error al descargar: %v", err)
	}
	defer resp.Body.Close()

	log.Debug("Respuesta HTTP - Status: %s, ContentLength: %d", resp.Status, resp.ContentLength)

	if resp.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("Respuesta HTTP no OK: %s", resp.Status)
		log.Error(errMsg)
		return "", fmt.Errorf("respuesta no OK: %s", resp.Status)
	}

	// Crear archivo temporal
	tmpFile, err := os.CreateTemp("", "pdf_*.pdf")
	if err != nil {
		log.Error("Error al crear archivo temporal: %v", err)
		return "", fmt.Errorf("error al crear archivo temporal: %v", err)
	}
	defer tmpFile.Close()

	// Copiar contenido
	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		log.Error("Error al guardar PDF: %v", err)
		return "", fmt.Errorf("error al guardar PDF: %v", err)
	}

	log.Info("PDF descargado exitosamente en %s. Tiempo de ejecución: %v", tmpFile.Name(), time.Since(startTime))
	return tmpFile.Name(), nil
}

func CleanupFile(path string) {
	if path == "" {
		log.Debug("CleanupFile llamado con path vacío")
		return
	}

	startTime := time.Now()
	log.Info("Intentando eliminar archivo: %s", path)

	err := os.Remove(path)
	if err != nil {
		log.Error("Error al eliminar archivo %s: %v", path, err)
	} else {
		log.Info("Archivo %s eliminado exitosamente. Tiempo de ejecución: %v", path, time.Since(startTime))
	}
}
