package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type DeepseekReasonerModel struct {
	APIKey string
	URL    string
}

// NewDeepseekReasonerModel crea una nueva instancia del modelo DeepseekReasoner
func NewDeepseekReasonerModel(apiKey string) DeepseekReasonerModel {
	return DeepseekReasonerModel{
		APIKey: apiKey,
		URL:    "https://api.deepseek.com/chat/completions",
	}
}

func (m *DeepseekReasonerModel) SendPrompt(systemContext string, prompt string) (*string, error) {
	requestBody := map[string]interface{}{
		"model": "deepseek-reasoner",
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": systemContext,
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"stream": false,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request body: %v", err)
	}

	req, err := http.NewRequest("POST", m.URL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+m.APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, body)
	}

	var apiResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("error unmarshaling API response: %v", err)
	}

	if len(apiResponse.Choices) == 0 {
		return nil, fmt.Errorf("no choices in API response")
	}

	jsonContent := apiResponse.Choices[0].Message.Content

	return &jsonContent, nil
}
