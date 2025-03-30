package pdf_extractor

import (
	"fmt"
	"github.com/unidoc/unipdf/v3/extractor"
	"github.com/unidoc/unipdf/v3/model"
	"go_ocr/src/services/logger"
	"go_ocr/src/services/pdf_extractor/ocr"
	"os"
	"os/exec"
	"strings"
	"time"
)

var (
	log = logger.NewLogger(false) // Logger compartido
)

func ExtractTextFromPDF(path string) (string, error) {
	startTime := time.Now()
	log.Info("Iniciando extracción de texto de PDF: %s", path)
	log.Debug("Parámetros de ExtractTextFromPDF - path: %s", path)

	text, err := extractWithPdfToText(path)
	if err != nil {
		log.Warning("Extracción con pdftotext fallida: %v", err)
	} else {
		log.Info("Extracción con pdftotext exitosa. Longitud del texto: %d", len(text))
		log.Debug("Texto extraído (primeros 100 caracteres): %.100q", text)
		log.Info("Proceso completado. Tiempo total: %v", time.Since(startTime))
		return text, nil
	}

	// Intentar extracción con UniPDF
	text, err = extractWithUniPDF(path)
	if err != nil {
		log.Warning("Extracción con UniPDF fallida: %v", err)
	} else if len(text) > 50 {
		log.Info("Extracción con UniPDF exitosa. Longitud del texto: %d", len(text))
		log.Debug("Texto extraído (primeros 100 caracteres): %.100q", text)
		log.Info("Proceso completado. Tiempo total: %v", time.Since(startTime))
		return text, nil
	} else {
		log.Warning("Texto extraído con UniPDF demasiado corto (%d caracteres), intentando con OCR", len(text))
	}

	// Fallback a OCR
	log.Info("Intentando extracción con OCR...")
	ocrText, err := ocr.ExtractWithOCR(path)
	if err != nil {
		log.Error("Extracción con OCR fallida: %v", err)
		return "", fmt.Errorf("fallaron ambos métodos de extracción: %v", err)
	}

	log.Info("Extracción con OCR exitosa. Longitud del texto: %d", len(ocrText))
	log.Debug("Texto OCR extraído (primeros 100 caracteres): %.100q", ocrText)
	log.Info("Proceso completado. Tiempo total: %v", time.Since(startTime))

	return ocrText, nil
}

func extractWithPdfToText(path string) (string, error) {
	log.Debug("Extrayendo texto de PDF con pdftotext: %s", path)

	// Ejecutar pdftotext
	cmd := exec.Command("pdftotext", path, "-")
	log.Debug("Ejecutando comando: %v", cmd.Args)

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error("Error al extraer texto con pdftotext: %v\nSalida: %s", err, string(output))
		return "", fmt.Errorf("error al extraer texto: %v\nSalida: %s", err, string(output))
	}

	log.Debug("Extracción con pdftotext completada. Longitud del texto: %d", len(output))
	return string(output), nil
}

func extractWithUniPDF(path string) (string, error) {
	log.Debug("Abriendo PDF con UniPDF: %s", path)

	// Abrir el archivo PDF
	f, err := os.Open(path)
	if err != nil {
		log.Error("Error al abrir archivo: %v", err)
		return "", fmt.Errorf("error al abrir archivo: %v", err)
	}
	defer f.Close()

	pdfReader, err := model.NewPdfReader(f)
	if err != nil {
		log.Error("Error al crear PDF reader: %v", err)
		return "", fmt.Errorf("error al crear PDF reader: %v", err)
	}

	var text strings.Builder
	totalPages, err := pdfReader.GetNumPages()
	if err != nil {
		log.Error("Error al obtener número de páginas: %v", err)
		return "", fmt.Errorf("error al obtener número de páginas: %v", err)
	}
	log.Info("Procesando PDF con %d páginas", totalPages)

	// Procesar cada página
	for i := 1; i <= totalPages; i++ {
		log.Debug("Extrayendo texto de página %d/%d", i, totalPages)

		page, err := pdfReader.GetPage(i)
		if err != nil {
			log.Error("Error al obtener página %d: %v", i, err)
			return "", fmt.Errorf("error en página %d: %v", i, err)
		}

		ex, err := extractor.New(page)
		if err != nil {
			log.Error("Error al crear extractor para página %d: %v", i, err)
			return "", fmt.Errorf("error en página %d: %v", i, err)
		}

		content, err := ex.ExtractText()
		if err != nil {
			log.Error("Error al extraer texto de página %d: %v", i, err)
			return "", fmt.Errorf("error en página %d: %v", i, err)
		}

		text.WriteString(content + "\n")
		log.Debug("Página %d procesada. Longitud acumulada: %d", i, text.Len())
	}

	if text.Len() == 0 {
		log.Warning("No se encontró texto legible en el PDF")
		return "", fmt.Errorf("no se encontró texto legible")
	}

	log.Info("Extracción con UniPDF completada. Longitud total del texto: %d", text.Len())
	return text.String(), nil
}
