package ocr

import (
	"fmt"
	"go_ocr/services/logger"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var (
	log = logger.NewLogger(false)
)

func ExtractWithOCR(pdfPath string) (string, error) {
	startTime := time.Now()
	log.Info("Iniciando extracción OCR para archivo: %s", pdfPath)
	log.Debug("Parámetros de extractWithOCR - pdfPath: %s", pdfPath)

	// Validar que el archivo existe
	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		log.Error("El archivo PDF no existe: %s", pdfPath)
		return "", fmt.Errorf("el archivo PDF no existe: %s", pdfPath)
	}

	// Crear directorio temporal
	tempDir, err := os.MkdirTemp("", "ocr_")
	if err != nil {
		log.Error("Error al crear directorio temporal: %v", err)
		return "", fmt.Errorf("error al crear directorio temporal: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			log.Error("Error al eliminar directorio temporal %s: %v", tempDir, err)
		} else {
			log.Debug("Directorio temporal eliminado: %s", tempDir)
		}
	}()

	log.Info("Directorio temporal creado: %s", tempDir)

	// 1. Convertir PDF a imágenes (una por página)
	log.Info("Convirtiendo PDF a imágenes...")
	cmd := exec.Command("pdftoppm", "-png", "-r", "300", pdfPath, filepath.Join(tempDir, "page"))
	log.Debug("Ejecutando comando: %v", cmd.Args)

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error("Error al convertir PDF a imágenes: %v\nSalida: %s", err, string(output))
		return "", fmt.Errorf("error al convertir PDF a imágenes: %v\nSalida: %s", err, string(output))
	}

	log.Debug("Conversión PDF a imágenes exitosa. Salida: %s", string(output))

	// 2. Procesar cada imagen con Tesseract
	log.Info("Procesando imágenes con Tesseract OCR...")
	var text strings.Builder
	files, err := os.ReadDir(tempDir)
	if err != nil {
		log.Error("Error al leer directorio temporal: %v", err)
		return "", fmt.Errorf("error al leer directorio temporal: %v", err)
	}

	processedPages := 0
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".png") {
			log.Debug("Archivo ignorado (no es imagen PNG): %s", file.Name())
			continue
		}

		imgPath := filepath.Join(tempDir, file.Name())
		log.Debug("Procesando página con OCR: %s", imgPath)

		cmd := exec.Command("tesseract", imgPath, "-", "-l", "spa+eng")
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Error("Error en OCR para %s: %v\nSalida: %s", imgPath, err, string(output))
			return "", fmt.Errorf("error en OCR para %s: %v\nSalida: %s", imgPath, err, string(output))
		}

		text.Write(output)
		text.WriteString("\n")
		processedPages++
		log.Debug("Página procesada exitosamente: %s", imgPath)
	}

	if text.Len() == 0 {
		log.Error("No se pudo extraer texto con OCR. Páginas procesadas: %d", processedPages)
		return "", fmt.Errorf("no se pudo extraer texto con OCR")
	}

	log.Info("Extracción OCR completada. Páginas procesadas: %d. Tiempo total: %v",
		processedPages, time.Since(startTime))
	log.Debug("Texto extraído (primeros 100 caracteres): %.100q", text.String())

	return text.String(), nil
}
