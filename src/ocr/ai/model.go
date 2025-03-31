package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go_ocr/src/ocr/prompter/prompterTypes"
	"io"
	"net/http"
	"os"
	"time"
)

type Model interface {
	SendPrompt(prompt prompterTypes.Prompt) (string, error)
}

type DeepSeekModel struct {
	url        string
	apiKey     string
	httpClient *http.Client
}

func NewDeepSeekModel() (*DeepSeekModel, error) {
	url := os.Getenv("DEEPSEEK_API_URL")

	if url == "" {
		return nil, fmt.Errorf("DEEPSEEK_API_URL environment variable not set")
	}

	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("DEEPSEEK_API_KEY environment variable not set")
	}

	return &DeepSeekModel{
		url:        url,
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (m *DeepSeekModel) SendPrompt(prompt prompterTypes.Prompt) (string, error) {
	if m.apiKey == "" {
		return "", fmt.Errorf("API key not configured")
	}

	requestBody := deepseekApiRequest{
		Model: "deepseek-reasoner",
		Messages: []deepseekMessage{
			{Role: "system", Content: prompt.Context},
			{Role: "user", Content: prompt.Prompt},
		},
		Stream: false,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("error marshaling request body: %w", err)
	}

	req, err := http.NewRequest("POST", m.url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+m.apiKey)

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var apiResponse deepseekApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return "", fmt.Errorf("error decoding API response: %w", err)
	}

	if len(apiResponse.Choices) == 0 {
		return "", fmt.Errorf("no choices in API response")
	}

	return apiResponse.Choices[0].Message.Content, nil
}

type deepseekApiRequest struct {
	Model    string            `json:"model"`
	Messages []deepseekMessage `json:"messages"`
	Stream   bool              `json:"stream"`
}

type deepseekMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type deepseekApiResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"deepseekMessage"`
	} `json:"choices"`
}
