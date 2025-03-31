package ocr

import (
	"errors"
	"fmt"
	"time"

	"go_ocr/src/ocr/ai"
	"go_ocr/src/ocr/downloader"
	"go_ocr/src/ocr/extractor"
	"go_ocr/src/ocr/logger"
	"go_ocr/src/ocr/prompter"
)

var (
	ErrUnsupportedOCRType = errors.New("unsupported OCR type")
)

type OCRBuilder struct {
	logger     *logger.Logger
	downloader *downloader.Downloader
	extractor  *extractor.TextExtractor
	model      ai.Model
	prompter   prompter.Prompter
}

func NewOCRBuider() *OCRBuilder {
	return &OCRBuilder{}
}

// Build crea una instancia de OCR configurada para el tipo especificado
func (b *OCRBuilder) Build(ocrType string) (*OCR, error) {
	if ocrType != "payroll" {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedOCRType, ocrType)
	}

	// Inicializar componentes básicos
	if err := b.initializeBaseComponents(); err != nil {
		return nil, err
	}

	// Configurar componentes específicos del tipo
	if err := b.configureTypeSpecificComponents(ocrType); err != nil {
		return nil, err
	}

	return NewOCR(
		b.logger,
		b.downloader,
		b.extractor,
		b.model,
		b.prompter,
	), nil
}

func (b *OCRBuilder) initializeBaseComponents() error {
	// TODO - Implement configuration in logger
	//b.logger = logger.NewLogger()

	// Downloader con configuración default
	b.downloader = downloader.NewDownloader(
		logger.NewLogger(true),
		downloader.WithTimeout(30*time.Second),
		downloader.WithMaxRetries(3),
	)

	// Extractor de texto
	b.extractor = extractor.NewTextExtractor(b.logger)

	return nil
}

func (b *OCRBuilder) configureTypeSpecificComponents(ocrType string) error {
	// Modelo de IA (DeepSeek como default)
	model, err := ai.NewDeepSeekModel()
	if err != nil {
		return fmt.Errorf("error creating AI model: %w", err)
	}
	b.model = model

	// Prompter específico para el tipo
	prompterFactory := prompter.NewPrompterFactory()
	p, err := prompterFactory.GetPrompter(ocrType)
	if err != nil {
		return fmt.Errorf("error creating prompter: %w", err)
	}
	b.prompter = p

	return nil
}
