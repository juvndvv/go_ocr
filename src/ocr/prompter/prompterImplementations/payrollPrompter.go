package prompterImplementations

import (
	"go_ocr/src/ocr/prompter/prompterTypes"
)

type PayrollPrompter struct {
}

func NewPayrollPrompter() *PayrollPrompter {
	return &PayrollPrompter{}
}

func (p *PayrollPrompter) GetContext() string {
	return "You are a payroll specialist at a company. You are responsible for processing payroll for all employees."
}

func (p *PayrollPrompter) BuildPrompt(text string) (prompterTypes.Prompt, error) {
	prompt := `Extract and transform payroll data into a JSON structure following these exact requirements:

		1. JSON Structure (Mantain fields order and nesting):
		   {
			 "employee": {
			   "name": "(nombre completo)",
			   "tax_id": "(validar formato)"
			 },
			 "date_range": {
			   "start_date": "yyyy-mm-dd",
			   "end_date": "yyyy-mm-dd"
			 },
			 "employer_costs": (valor numérico),
			 "gross_amount": (valor numérico),
			 "deductions": (valor numérico),
			 "net_amount": (valor numérico)
		   }
		
		2. Reglas estrictas:
		   - Tax ID Validation: Si el ID fiscal comienza con A,B,C,D, busca una cadena que se refiera a un DNI o NIE, si no la encuentras devuelve null
		   - Employer Costs: Debe ser (gross_amount + deductions + company's contributions) si no se encuentra
		   - Fechas: Formato ISO 8601 (ej: 2023-10-01)
		   - Valores numéricos: Usar .0 decimal incluso para enteros
		   - Campos faltantes: Usar null
		
		3. Validaciones finales:
		   - employer_costs DEBE ser el valor más alto
		   - net_amount DEBE ser (gross_amount - deductions)
		   - Eliminar cualquier campo no especificado
		   - Nunca agregar comentarios/explicaciones
		
		Procesar el texto proporcionado aplicando estas reglas estrictamente.` + "\n\n" + text

	return prompterTypes.Prompt{
		Context: p.GetContext(),
		Prompt:  prompt,
	}, nil
}
