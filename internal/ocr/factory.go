package ocr

type Factory struct {
}

func NewFactory() Factory {
	return Factory{}
}

func (f *Factory) Create(OcrType string) (*OCR, error) {
	switch OcrType {
	case "payroll":
		return f.createPayrollOcr()
	}

	return nil, ErrUnsupportedOcrType
}

func (f *Factory) createPayrollOcr() (*OCR, error) {
	return NewOCR(), nil
}
