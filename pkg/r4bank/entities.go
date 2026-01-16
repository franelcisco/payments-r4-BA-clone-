package r4bank

// BCVResponse represent the response from the BCV API
type BCVResponse struct {
	Code       string  `json:"code"`
	Fechavalor string  `json:"fechavalor"`
	Tipocambio float64 `json:"tipocambio"`
}

type OTPResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Success bool   `json:"success"`
}

type ValidateOTPResponse struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Reference string `json:"reference"`
}

type ChangePaidResponse struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Reference int    `json:"reference"`
}

type ValidateDebitInmediateResponse struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	ID        string `json:"id"`
	Reference string `json:"reference"`
}

type GetOperationResponse struct {
	Code      string `json:"code"`
	Reference string `json:"reference"`
	Success   bool   `json:"success"`
}
