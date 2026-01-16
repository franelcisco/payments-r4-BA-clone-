package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"

	"bone_appetit_r4_service/internal/models"
	"bone_appetit_r4_service/pkg/r4bank"
)

type R4Service interface {
	GetBCVTasaUSD(ctx context.Context) (*models.BCVTasaUSDResponse, error)
	GenerateOTP(ctx context.Context, req *models.OTPRequest) error
	ValidateImmediateDebit(ctx context.Context, req *models.ValidateOTPRequest) (*models.ValidateDebitInmediateResponse, error)
	ChangePaid(ctx context.Context, req *models.ChangePaidRequest) (*models.ChangePaidResponse, error)
	GetOperationByID(ctx context.Context, operationID string) (*r4bank.GetOperationResponse, error)
}

type r4Service struct {
	r4Client *r4bank.RestClient
	Logger   *zap.Logger
}

var _DebitInmediateSpecialResponse = map[string]string{
	"AC01": "Número de cuenta incorrecto",
	"AM04": "Saldo insuficiente",
	"MD15": "Monto incorrecto",
	"TKCM": "Codigo OTP inválido",
	"AC00": "En espera de respuesta del banco",
	"ACCP": "Transacción Exitosa",
}

const _debitInmetiateGenericError = "ocurrió un error al procesar la solicitud"

// NewR4Service creates a new R4Service
func NewR4Service(logger *zap.Logger, r4Client *r4bank.RestClient) R4Service {
	return &r4Service{
		r4Client: r4Client,
		Logger:   logger,
	}
}

// GetBCVTasaUSD retrieves the BCV exchange rate for USD
func (r *r4Service) GetBCVTasaUSD(ctx context.Context) (*models.BCVTasaUSDResponse, error) {
	var (
		dateValue = time.Now().Format("2006-01-02")
		currency  = "USD"
	)

	hmacInput := dateValue + currency
	payload := map[string]string{
		"Moneda":     currency,
		"Fechavalor": dateValue,
	}

	resp, err := r.r4Client.Do(ctx, hmacInput, payload, "MBbcv")
	if err != nil {
		r.Logger.Error(err.Error(), zap.Any("payload", payload))
		return nil, fmt.Errorf("error en request: %w", err)
	}

	var r4Resp r4bank.BCVResponse
	if err := json.Unmarshal(resp, &r4Resp); err != nil {
		r.Logger.Error(err.Error(), zap.Any("response", string(resp)))
		return nil, fmt.Errorf("error decodificando respuesta: %w", err)
	}

	if r4Resp.Code != "00" {
		r.Logger.Error("R4 API error", zap.String("code", r4Resp.Code), zap.Any("payload", payload))
		return nil, errors.New("R4 API returned an error")
	}

	return &models.BCVTasaUSDResponse{
		Date: r4Resp.Fechavalor,
		Rate: r4Resp.Tipocambio,
	}, nil
}

// ChangePaid returns paid in Bolivares
func (r *r4Service) ChangePaid(ctx context.Context, req *models.ChangePaidRequest) (*models.ChangePaidResponse, error) {
	hmacInput := req.Phone + fmt.Sprintf("%.2f", req.Amount) + req.Bank + req.DNI
	payload := map[string]string{
		"TelefonoDestino": req.Phone,
		"Cedula":          req.DNI,
		"Banco":           req.Bank,
		"Monto":           fmt.Sprintf("%.2f", req.Amount),
		"Concepto":        req.Concept,
	}

	resp, err := r.r4Client.Do(ctx, hmacInput, payload, "MBvuelto")
	if err != nil {
		r.Logger.Error(err.Error(), zap.Any("payload", payload))
		return nil, fmt.Errorf("error en request: %w", err)
	}

	var changeResp r4bank.ChangePaidResponse
	if err := json.Unmarshal(resp, &changeResp); err != nil {
		r.Logger.Error(err.Error(), zap.Any("response", string(resp)))
		return nil, fmt.Errorf("error decodificando respuesta: %w", err)
	}

	if changeResp.Code != "00" {
		r.Logger.Error("R4 Change Paid API error", zap.String("code", changeResp.Code), zap.Any("payload", payload))
		return nil, errors.New("R4 Change Paid API returned an error")
	}

	return &models.ChangePaidResponse{
		Reference: fmt.Sprintf("%d", changeResp.Reference),
	}, nil
}

// GenerateOTP generates a one-time password (OTP) for secure transactions
func (r *r4Service) GenerateOTP(ctx context.Context, req *models.OTPRequest) error {
	hmacInput := req.Bank + fmt.Sprintf("%.2f", req.Amount) + req.Phone + req.DNI
	payload := map[string]string{
		"Banco":    req.Bank,
		"Monto":    fmt.Sprintf("%.2f", req.Amount),
		"Telefono": req.Phone,
		"Cedula":   req.DNI,
	}

	resp, err := r.r4Client.Do(ctx, hmacInput, payload, "GenerarOtp")
	if err != nil {
		r.Logger.Error(err.Error(), zap.Any("payload", payload))
		return fmt.Errorf("error en request: %w", err)
	}

	var otpResp r4bank.OTPResponse
	if err := json.Unmarshal(resp, &otpResp); err != nil {
		r.Logger.Error(err.Error(), zap.Any("response", string(resp)))
		return fmt.Errorf("error decodificando respuesta: %w", err)
	}

	if otpResp.Code != "202" {
		r.Logger.Error("R4 OTP API error", zap.String("code", otpResp.Message), zap.Any("payload", payload))
		return errors.New("R4 OTP API returned an error")
	}

	return nil
}

type ValidateDebitInmediateResponse struct {
	ID   string `json:"id"`
	Code string `json:"code"`
}

// ValidateImmediateDebit validates an immediate debit transaction using the provided OTP
func (r *r4Service) ValidateImmediateDebit(ctx context.Context, req *models.ValidateOTPRequest) (*models.ValidateDebitInmediateResponse, error) {

	hmacInput := req.Bank + req.DNI + req.Phone + fmt.Sprintf("%.2f", req.Amount) + req.OTP
	payload := map[string]string{
		"Banco":    req.Bank,
		"Monto":    fmt.Sprintf("%.2f", req.Amount),
		"Telefono": req.Phone,
		"Cedula":   req.DNI,
		"Nombre":   req.Name,
		"OTP":      req.OTP,
		"Concepto": req.Concept,
	}

	resp, err := r.r4Client.Do(ctx, hmacInput, payload, "DebitoInmediato")
	if err != nil {
		r.Logger.Error(err.Error(), zap.Any("payload", payload))
		return nil, err
	}

	var validateResp r4bank.ValidateDebitInmediateResponse
	if err := json.Unmarshal(resp, &validateResp); err != nil {
		r.Logger.Error(err.Error(), zap.Any("response", string(resp)))
		return nil, err
	}

	var operationResp *r4bank.GetOperationResponse
	intent := 0
	for intent < 7 {
		if validateResp.Code != "ACCP" {
			time.Sleep(3 * time.Second)
		}

		operationResp, err = r.GetOperationByID(ctx, validateResp.ID)
		if err != nil {
			r.Logger.Error(err.Error(), zap.Any("payload", payload), zap.Any("validateResp", validateResp.ID))
			return nil, err
		}

		if operationResp == nil {
			r.Logger.Error("nil response from GetOperationByID", zap.Any("payload", payload))
			return nil, nil
		}

		if operationResp.Code != "AC00" {
			break
		}

		intent++
	}

	message := _debitInmetiateGenericError
	if msg, exist := _DebitInmediateSpecialResponse[operationResp.Code]; exist {
		message = msg
	}

	return &models.ValidateDebitInmediateResponse{
		ID:        validateResp.ID,
		Code:      operationResp.Code,
		Reference: operationResp.Reference,
		Message:   message,
		Status:    operationResp.Success,
	}, nil
}

// GetOperationByID
func (r *r4Service) GetOperationByID(ctx context.Context, operationID string) (*r4bank.GetOperationResponse, error) {
	hmacInput := operationID
	payload := map[string]string{
		"id": operationID,
	}
	resp, err := r.r4Client.Do(ctx, hmacInput, payload, "ConsultarOperaciones")
	if err != nil {
		r.Logger.Error(err.Error(), zap.Any("payload", payload))
		return nil, err
	}

	var opResp r4bank.GetOperationResponse
	if err := json.Unmarshal(resp, &opResp); err != nil {
		r.Logger.Error(err.Error(), zap.Any("response", string(resp)))
		return nil, err
	}

	r.Logger.Info("Operation response", zap.Any("operation", opResp))

	return &r4bank.GetOperationResponse{
		Code:      opResp.Code,
		Success:   opResp.Success,
		Reference: opResp.Reference,
	}, nil
}
