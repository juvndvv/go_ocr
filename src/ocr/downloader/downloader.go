package downloader

import (
	"fmt"
	"go_ocr/src/ocr/logger"
	"io"
	"net/http"
	"os"
	"time"
)

type DownloaderConfig struct {
	Timeout        time.Duration
	MaxRetries     int
	RetryDelay     time.Duration
	UserAgent      string
	MaxRedirects   int
	ValidatePDFURL bool
}

type Downloader struct {
	logger *logger.Logger
	config DownloaderConfig
}

type DownloaderOption func(*DownloaderConfig)

func NewDownloader(logger *logger.Logger, opts ...DownloaderOption) *Downloader {
	// Configuración por defecto
	config := DownloaderConfig{
		Timeout:        30 * time.Second,
		MaxRetries:     3,
		RetryDelay:     1 * time.Second,
		UserAgent:      "OCR-Downloader/1.0",
		MaxRedirects:   5,
		ValidatePDFURL: true,
	}

	// Aplicar opciones
	for _, opt := range opts {
		opt(&config)
	}

	return &Downloader{
		logger: logger,
		config: config,
	}
}

func WithTimeout(timeout time.Duration) DownloaderOption {
	return func(c *DownloaderConfig) {
		c.Timeout = timeout
	}
}

func WithMaxRetries(retries int) DownloaderOption {
	return func(c *DownloaderConfig) {
		c.MaxRetries = retries
	}
}

func (d *Downloader) Download(url string) (string, error) {
	startTime := time.Now()
	d.logger.Info("Iniciando descarga de PDF desde URL: %s", url)
	d.logger.Debug("Configuración: %+v", d.config)

	client := &http.Client{
		Timeout: d.config.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= d.config.MaxRedirects {
				return fmt.Errorf("demasiados redirects (%d)", d.config.MaxRedirects)
			}
			return nil
		},
	}

	var resp *http.Response
	var err error

	for attempt := 0; attempt <= d.config.MaxRetries; attempt++ {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return "", fmt.Errorf("error creando request: %w", err)
		}
		req.Header.Set("User-Agent", d.config.UserAgent)

		resp, err = client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}

		if attempt < d.config.MaxRetries {
			d.logger.Debug("Intento %d fallido, reintentando en %v...", attempt+1, d.config.RetryDelay)
			time.Sleep(d.config.RetryDelay)
		}
	}

	if err != nil {
		d.logger.Error("Error al descargar PDF: %v", err)
		return "", fmt.Errorf("error al descargar: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("Respuesta HTTP no OK: %s", resp.Status)
		d.logger.Error(errMsg)
		return "", fmt.Errorf(errMsg)
	}

	tmpFile, err := os.CreateTemp("", "pdf_*.pdf")
	if err != nil {
		d.logger.Error("Error al crear archivo temporal: %v", err)
		return "", fmt.Errorf("error al crear archivo temporal: %w", err)
	}
	defer tmpFile.Close()

	if _, err = io.Copy(tmpFile, resp.Body); err != nil {
		d.logger.Error("Error al guardar PDF: %v", err)
		return "", fmt.Errorf("error al guardar PDF: %w", err)
	}

	d.logger.Info("PDF descargado exitosamente en %s. Tiempo total: %v",
		tmpFile.Name(), time.Since(startTime))

	return tmpFile.Name(), nil
}

func (d *Downloader) CleanupFile(path string) {
	if path == "" {
		d.logger.Debug("CleanupFile llamado con path vacío")
		return
	}

	startTime := time.Now()
	d.logger.Info("Intentando eliminar archivo: %s", path)

	err := os.Remove(path)
	if err != nil {
		d.logger.Error("Error al eliminar archivo %s: %v", path, err)
	} else {
		d.logger.Info("Archivo %s eliminado exitosamente. Tiempo de ejecución: %v", path, time.Since(startTime))
	}
}
