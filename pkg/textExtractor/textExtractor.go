package textExtractor

import (
	"fmt"
	"github.com/juvndvv/ocr/pkg"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type TextExtractor struct {
}

func NewTextExtractor() TextExtractor {
	return TextExtractor{}
}

func (te *TextExtractor) ExtractText(path string) (string, error) {
	return extractTextFromPDF(path)
}

func extractTextFromPDF(path string) (string, error) {
	text, err := extractWithPdfToText(path)
	if text != "" {
		return text, nil
	}

	ocrText, err := extractWithOCR(path)
	if err != nil {
		return "", ErrExtractionFailed
	}

	return ocrText, nil
}

func extractWithOCR(pdfPath string) (string, error) {
	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		return "", pkg.ErrFileNotFound
	}

	tempDir, err := os.MkdirTemp("", "ocr_")
	if err != nil {
		return "", pkg.ErrCreatingDir
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			fmt.Printf("Error al eliminar directorio temporal: %v\n", err)
		}
	}()

	cmd := exec.Command("pdftoppm", "-png", "-r", "300", pdfPath, filepath.Join(tempDir, "page"))

	_, err = cmd.CombinedOutput()
	if err != nil {
		return "", ErrConversionError
	}

	var text strings.Builder
	files, err := os.ReadDir(tempDir)
	if err != nil {
		return "", pkg.ErrReadingDir
	}

	processedPages := 0
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".png") {
			continue
		}

		imgPath := filepath.Join(tempDir, file.Name())

		cmd := exec.Command("tesseract", imgPath, "-", "-l", "spa+eng")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return "", ErrConversionError
		}

		text.Write(output)
		text.WriteString("\n")
		processedPages++
	}

	if text.Len() == 0 {
		return "", ErrConversionError
	}

	return text.String(), nil
}

func extractWithPdfToText(path string) (string, error) {
	// Ejecutar pdftotext
	cmd := exec.Command("pdftotext", path, "-")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", ErrConversionError
	}

	return string(output), nil
}
