package extractor

import (
	"fmt"
	"go_ocr/src/ocr/logger"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type TextExtractor struct {
	logger *logger.Logger
}

func NewTextExtractor(logger *logger.Logger) *TextExtractor {
	return &TextExtractor{
		logger: logger,
	}
}

func (p *TextExtractor) ExtractText(path string) (string, error) {
	// TODO implement fallbacks
	return p.extractWithPdfToText(path)
}

func (p *TextExtractor) extractWithPdfToText(path string) (string, error) {
	p.logger.Debug("Extrayendo texto de PDF con pdftotext: %s", path)

	// Check if pdftotext is installed
	_, err := exec.LookPath("pdftotext")
	if err != nil {
		p.logger.Error("pdftotext no está instalado")
		return "", fmt.Errorf("pdftotext no está instalado")
	}

	// Execute pdftotext command
	cmd := exec.Command("pdftotext", path, "-")
	p.logger.Debug("Ejecutando comando: %v", cmd.Args)

	output, err := cmd.CombinedOutput()
	if err != nil {
		p.logger.Error("Error al extraer texto con pdftotext: %v\nSalida: %s", err, string(output))
		return "", fmt.Errorf("error al extraer texto: %v\nSalida: %s", err, string(output))
	}

	p.logger.Debug("Extracción con pdftotext completada. Longitud del texto: %d", len(output))
	return string(output), nil
}

func (p *TextExtractor) extractWithTesseract(pdfPath string) (string, error) {
	startTime := time.Now()
	p.logger.Info("Iniciando extracción OCR para archivo: %s", pdfPath)
	p.logger.Debug("Parámetros de extractWithOCR - pdfPath: %s", pdfPath)

	// Validar que el archivo existe
	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		p.logger.Error("El archivo PDF no existe: %s", pdfPath)
		return "", fmt.Errorf("el archivo PDF no existe: %s", pdfPath)
	}

	// Crear directorio temporal
	tempDir, err := os.MkdirTemp("", "ocr_")
	if err != nil {
		p.logger.Error("Error al crear directorio temporal: %v", err)
		return "", fmt.Errorf("error al crear directorio temporal: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			p.logger.Error("Error al eliminar directorio temporal %s: %v", tempDir, err)
		} else {
			p.logger.Debug("Directorio temporal eliminado: %s", tempDir)
		}
	}()

	p.logger.Info("Directorio temporal creado: %s", tempDir)

	// 1. Convertir PDF a imágenes (una por página)
	p.logger.Info("Convirtiendo PDF a imágenes...")
	cmd := exec.Command("pdftoppm", "-png", "-r", "300", pdfPath, filepath.Join(tempDir, "page"))
	p.logger.Debug("Ejecutando comando: %v", cmd.Args)

	output, err := cmd.CombinedOutput()
	if err != nil {
		p.logger.Error("Error al convertir PDF a imágenes: %v\nSalida: %s", err, string(output))
		return "", fmt.Errorf("error al convertir PDF a imágenes: %v\nSalida: %s", err, string(output))
	}

	p.logger.Debug("Conversión PDF a imágenes exitosa. Salida: %s", string(output))

	// 2. Procesar cada imagen con Tesseract
	p.logger.Info("Procesando imágenes con Tesseract OCR...")
	var text strings.Builder
	files, err := os.ReadDir(tempDir)
	if err != nil {
		p.logger.Error("Error al leer directorio temporal: %v", err)
		return "", fmt.Errorf("error al leer directorio temporal: %v", err)
	}

	processedPages := 0
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".png") {
			p.logger.Debug("Archivo ignorado (no es imagen PNG): %s", file.Name())
			continue
		}

		imgPath := filepath.Join(tempDir, file.Name())
		p.logger.Debug("Procesando página con OCR: %s", imgPath)

		cmd := exec.Command("tesseract", imgPath, "-", "-l", "spa+eng")
		output, err := cmd.CombinedOutput()
		if err != nil {
			p.logger.Error("Error en OCR para %s: %v\nSalida: %s", imgPath, err, string(output))
			return "", fmt.Errorf("error en OCR para %s: %v\nSalida: %s", imgPath, err, string(output))
		}

		text.Write(output)
		text.WriteString("\n")
		processedPages++
		p.logger.Debug("Página procesada exitosamente: %s", imgPath)
	}

	if text.Len() == 0 {
		p.logger.Error("No se pudo extraer texto con OCR. Páginas procesadas: %d", processedPages)
		return "", fmt.Errorf("no se pudo extraer texto con OCR")
	}

	p.logger.Info("Extracción OCR completada. Páginas procesadas: %d. Tiempo total: %v",
		processedPages, time.Since(startTime))
	p.logger.Debug("Texto extraído (primeros 100 caracteres): %.100q", text.String())

	return text.String(), nil
}
