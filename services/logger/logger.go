package logger

import (
	"io"
	"log"
	"os"
)

type Logger struct {
	infoLogger    *log.Logger
	errorLogger   *log.Logger
	debugLogger   *log.Logger
	warningLogger *log.Logger
	fatalLogger   *log.Logger
	file          *os.File
}

// NewLogger crea una nueva instancia de Logger
func NewLogger(logToFile bool) *Logger {
	l := &Logger{
		infoLogger:    log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		errorLogger:   log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
		debugLogger:   log.New(io.Discard, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile),
		warningLogger: log.New(os.Stdout, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile),
		fatalLogger:   log.New(os.Stderr, "FATAL: ", log.Ldate|log.Ltime|log.Lshortfile),
	}

	if logToFile {
		file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			l.file = file

			// Configurar los writers para cada logger
			setupMultiWriter := func(logger *log.Logger) {
				logger.SetOutput(io.MultiWriter(logger.Writer(), file))
			}

			setupMultiWriter(l.infoLogger)
			setupMultiWriter(l.errorLogger)
			setupMultiWriter(l.warningLogger)
			setupMultiWriter(l.fatalLogger)
		} else {
			l.Error("No se pudo abrir archivo de log: %v", err)
		}
	}

	return l
}

// EnableDebug activa los mensajes de debug
func (l *Logger) EnableDebug() {
	var output io.Writer = os.Stdout
	if l.file != nil {
		output = io.MultiWriter(os.Stdout, l.file)
	}
	l.debugLogger.SetOutput(output)
}

// Info registra un mensaje informativo
func (l *Logger) Info(format string, v ...interface{}) {
	l.infoLogger.Printf(format, v...)
}

// Error registra un mensaje de error
func (l *Logger) Error(format string, v ...interface{}) {
	l.errorLogger.Printf(format, v...)
}

// Debug registra un mensaje de debug
func (l *Logger) Debug(format string, v ...interface{}) {
	l.debugLogger.Printf(format, v...)
}

// Warning registra un mensaje de advertencia
func (l *Logger) Warning(format string, v ...interface{}) {
	l.warningLogger.Printf(format, v...)
}

// Fatal registra un mensaje fatal y termina la aplicación
func (l *Logger) Fatal(format string, v ...interface{}) {
	l.fatalLogger.Printf(format, v...)
	l.Close()
	os.Exit(1)
}

// Close cierra los recursos del logger
func (l *Logger) Close() {
	if l.file != nil {
		l.file.Close()
	}
}
