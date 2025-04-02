package ocr

import (
	"github.com/juvndvv/ocr/pkg/ai"
	"github.com/juvndvv/ocr/pkg/downloader"
	"github.com/juvndvv/ocr/pkg/textExtractor"
)

type OCR struct {
	downloaderInstance    downloader.Downloader
	textExtractorInstance textExtractor.TextExtractor
	model                 ai.Model
}

func NewOCR(model ai.Model) *OCR {
	return &OCR{
		downloaderInstance:    downloader.NewDownloader(),
		textExtractorInstance: textExtractor.NewTextExtractor(),
		model:                 model,
	}
}

func (o *OCR) ExtractText(url string) (*string, error) {
	// Download
	path, err := o.downloaderInstance.Download(url)
	if err != nil {
		return nil, err
	}

	// Extract
	text, err := o.textExtractorInstance.ExtractText(path)
	if err != nil {
		return nil, err
	}

	// Send to model
	result, err := o.model.SendPrompt("", text)
	if err != nil {
		return nil, err
	}

	return result, nil
}
