package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go_ocr/services/logger"
	"io"
	"net/http"
	"os"
	"strings"
)

var (
	log = logger.NewLogger(false) // Logger compartido
)

// PayrollData representa la estructura del JSON que esperamos recibir
type PayrollData struct {
	Employee struct {
		Name  string `json:"name"`
		TaxID string `json:"tax_id"`
	} `json:"employee"`
	EmployerCosts float64 `json:"employer_costs"`
	GrossAmount   float64 `json:"gross_amount"`
	Deductions    float64 `json:"deductions"`
	NetAmount     float64 `json:"net_amount"`
}

// ExtractPayrollData envía el texto al modelo de IA y devuelve los datos estructurados
func ExtractPayrollData(text string) (*PayrollData, error) {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")

	// Construir el prompt completo
	prompt := `Extract the following payroll data from the provided text and return ONLY a JSON with these exact fields (use empty strings or null for missing data). The employee.tax_id field must not start with A, B, C or D. Do not include any other fields or comments:
	{
	  "employee": {
		"name": "",
		"tax_id": ""
	  },
      "date_range": {
		"start_date": "yyyy-mm-dd",
		"end_date": "yyyy-mm-dd"
	  },
	  "employer_costs": 0.0,
	  "gross_amount": 0.0,
	  "deductions": 0.0,
	  "net_amount": 0.0
	}` + "\n\n" + text

	log.Info("Prompt: %s", prompt)

	// Estructura para la solicitud a la API
	requestBody := map[string]interface{}{
		"model": "deepseek-chat",
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "Extract payroll data",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"stream": false,
	}

	// Convertir a JSON
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request body: %v", err)
	}

	// Crear la solicitud HTTP
	req, err := http.NewRequest("POST", "https://api.deepseek.com/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	// Añadir headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// Realizar la solicitud
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	// Leer la respuesta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	// Verificar el código de estado
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, body)
	}

	// Parsear la respuesta de la API
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

	log.Info("API response: %+v", apiResponse)

	if len(apiResponse.Choices) == 0 {
		return nil, fmt.Errorf("no choices in API response")
	}

	// Extraer el contenido JSON de la respuesta
	jsonContent := apiResponse.Choices[0].Message.Content
	jsonContent = cleanJSONResponse(jsonContent)

	// Parsear el JSON a nuestra estructura
	var payrollData PayrollData
	if err := json.Unmarshal([]byte(jsonContent), &payrollData); err != nil {
		return nil, fmt.Errorf("error unmarshaling payroll data: %v", err)
	}

	return &payrollData, nil
}

func cleanJSONResponse(response string) string {
	// Eliminar los triple backticks y la palabra "json" si están presentes
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	return strings.TrimSpace(response)
}
