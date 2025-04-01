package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go_ocr/src/services/logger"
	"io"
	"net/http"
	"os"
	"regexp"
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
	DateRange struct {
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
	} `json:"date_range"`
	EmployerCosts float64 `json:"employer_costs"`
	GrossAmount   float64 `json:"gross_amount"`
	Deductions    float64 `json:"deductions"`
	NetAmount     float64 `json:"net_amount"`
}

// ExtractPayrollData envía el texto al modelo de IA y devuelve los datos estructurados
func ExtractPayrollData(text string) (*PayrollData, error) {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")

	// Construir el prompt completo
	prompt := `{  
  "employee": {  
    "name": "(nombre completo)",  
    "tax_id": "(si comienza con A/B/C/D, validar como NIE (Ej: A1234567X); si es numérico + letra, validar como DNI (Ej: 12345678Z); si no, null)"  
  },  
  "date_range": {  
    "start_date": "yyyy-mm-dd",  
    "end_date": "yyyy-mm-dd"  
  },  
  "employer_costs": "(si no se encuentra explícito, calcular como: gross_amount + (gross_amount * [0.236 + (0.055 si 'indefinido' en el texto, 0.067 si 'temporal') + 0.002 + 0.006]))",  
  "gross_amount": "(valor numérico con .0)",  
  "deductions": "(valor numérico con .0)",  
  "net_amount": "(gross_amount - deductions)"  
}  ` + "\n\n" + text

	log.Info("Prompt: %s", prompt)

	// Estructura para la solicitud a la API
	requestBody := map[string]interface{}{
		"model": "deepseek-reasoner",
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

	log.Info("API response: %+v", apiResponse.Choices[0].Message.Content)

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

	// Define la expresión regular para DNI (8 dígitos + letra) y NIE (X/Y/Z + 7 dígitos + letra)
	taxIDRegex := `(?i)^(\d{8}[a-z]|[xyz]\d{7}[a-z])$`
	matched, err := regexp.MatchString(taxIDRegex, payrollData.Employee.TaxID)
	if err != nil {
		panic(err) // Maneja errores en la ejecución de la regex
	}

	if !matched {
		// Busca en la variable text cualquier coincidencia de DNI/NIE
		re := regexp.MustCompile(`(?i)(\d{8}[a-z]|[xyz]\d{7}[a-z])`)
		found := re.FindString(text)
		if found != "" {
			// Actualiza el TaxID con el valor encontrado (opcional: convertir a mayúsculas)
			payrollData.Employee.TaxID = strings.ToUpper(found)
		}
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
