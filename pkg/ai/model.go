package ai

type Model interface {
	// SendPrompt envía un prompt al modelo y devuelve la respuesta
	SendPrompt(systemContext string, prompt string) (*string, error)
}
