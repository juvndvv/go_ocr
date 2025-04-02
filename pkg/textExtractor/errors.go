package textExtractor

import "errors"

var (
	ErrExtractionFailed = errors.New("text extraction failed")
	ErrConversionError  = errors.New("convert to image failed")
)
