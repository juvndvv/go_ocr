package pkg

import "errors"

var (
	ErrFileNotFound = errors.New("file not found")
	ErrCreatingDir  = errors.New("creating temp directory failed")
	ErrReadingDir   = errors.New("reading temp directory failed")
)
