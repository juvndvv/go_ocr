package ocr

import (
	"errors"
)

var (
	ErrUnsupportedOcrType = errors.New("unsupported OCR type")
	ErrOcrLectureFailed   = errors.New("ocr lecture failed")
)
