package ocr

import (
	"go_ocr/src/ocr/ai"
	"go_ocr/src/ocr/downloader"
	"go_ocr/src/ocr/extractor"
	"go_ocr/src/ocr/logger"
	"go_ocr/src/ocr/prompter"
)

type OCR struct {
	logger     *logger.Logger
	downloader *downloader.Downloader
	extractor  *extractor.TextExtractor
	prompter   prompter.Prompter
	model      ai.Model
}

func NewOCR(
	logger *logger.Logger,
	downloader *downloader.Downloader,
	extractor *extractor.TextExtractor,
	model ai.Model,
	prompter prompter.Prompter,
) *OCR {
	return &OCR{
		logger:     logger,
		downloader: downloader,
		extractor:  extractor,
		model:      model,
		prompter:   prompter,
	}
}

func (o *OCR) read(url string) (string, error) {
	// Download
	pdfPath, err := o.downloader.Download(url)
	if err != nil {
		o.logger.Error("Error al descargar PDF: %v", err)
		return "", err
	}

	// Extract text
	text, err := o.extractor.ExtractText(pdfPath)
	if err != nil {
		o.logger.Error("Error al extraer texto de PDF: %v", err)
		return "", err
	}

	// Build prompt
	prompt, err := o.prompter.BuildPrompt(text)
	if err != nil {
		o.logger.Error("Error al construir prompt: %v", err)
		return "", err
	}

	// Send prompt to model
	response, err := o.model.SendPrompt(prompt)
	if err != nil {
		o.logger.Error("Error al enviar texto al modelo: %v", err)
		return "", err
	}

	// TODO - Extract structured data from response
	data := response

	// Remove downloaded PDF
	o.downloader.CleanupFile(pdfPath)

	return data, nil
}
