package types

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
