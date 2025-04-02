package downloader

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

type Downloader interface {
	Download(url string) (string, error)
	CleanUpFile(path string) (bool, error)
}

type FileDownloader struct{}

func NewDownloader() *FileDownloader {
	return &FileDownloader{}
}

func (d *FileDownloader) Download(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("error al descargar: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("Error al cerrar respuesta: %v\n", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("respuesta no OK: %s", resp.Status)
	}

	tmpFile, err := os.CreateTemp("", "*")
	if err != nil {
		return "", fmt.Errorf("error al crear archivo temporal: %v", err)
	}
	defer func(tmpFile *os.File) {
		err := tmpFile.Close()
		if err != nil {
			fmt.Printf("Error al cerrar archivo temporal: %v\n", err)
		}
	}(tmpFile)

	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return "", fmt.Errorf("error al guardar PDF: %v", err)
	}

	return tmpFile.Name(), nil
}

func (d *FileDownloader) CleanUpFile(path string) (bool, error) {
	if path == "" {
		return false, fmt.Errorf("ruta vacía")
	}

	err := os.Remove(path)
	if err != nil {
		return false, fmt.Errorf("error al eliminar archivo: %v", err)
	}

	return true, nil
}
