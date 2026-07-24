package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"

	"bone_appetit_r4_service/internal/models"
	"bone_appetit_r4_service/internal/repositories"
	"bone_appetit_r4_service/pkg/r4bank"
)

type R4Service interface {
	GetBCVTasaUSD(ctx context.Context) (*models.BCVTasaUSDResponse, error)
	GenerateOTP(ctx context.Context, req *models.OTPRequest) error
	ValidateImmediateDebit(ctx context.Context, req *models.ValidateOTPRequest) (*models.ValidateDebitInmediateResponse, error)
	ChangePaid(ctx context.Context, req *models.ChangePaidRequest) (*models.ChangePaidResponse, error)
	GetOperationByID(ctx context.Context, operationID string) (*r4bank.GetOperationResponse, error)
	DirectDebitAccount(ctx context.Context, req *models.DebitDirectAccountRequest) (*models.DebitDirectAccountResponse, error)
	DirectDebitPhone(ctx context.Context, req *models.DebitDirectPhoneRequest) (*models.DebitDirectPhoneResponse, error)
}

type r4Service struct {
	r4Client *r4bank.RestClient
	bankRepo repositories.BankRepository
	Logger   *zap.Logger
}

var _DebitInmediateSpecialResponse = map[string]string{
	"AC01": "Número de cuenta incorrecto",
	"AM04": "Saldo insuficiente",
	"MD15": "Monto incorrecto",
	"TKCM": "Codigo OTP inválido",
	"AC00": "En espera de respuesta del banco",
	"ACCP": "Transacción Exitosa",
	"VE01": "Fuera del horario permitido",
	"MD01": "No posee afiliación",
	"BE01": "Datos del cliente no corresponden a la cuenta",
}

const _debitInmetiateGenericError = "ocurrió un error al procesar la solicitud"

// Endpoints usados en las llamadas a r4Client.Do (privados al paquete)
const (
	bcvTasaUSDEndpoint             = "MBbcv"
	changePaidEndpoint             = "MBvuelto"
	generateOTPEndpoint            = "GenerarOtp"
	validateImmediateDebitEndpoint = "DebitoInmediato"
	getOperationByIDEndpoint       = "ConsultarOperaciones"
	directDebitAccountEndpoint     = "TransferenciaOnline/DomiciliacionCNTA"
	directDebitPhoneEndpoint       = "TransferenciaOnline/DomiciliacionCELE"
	sendOperationTrueCode          = "202"
	inBreakpointCode               = "AC00"
)

// NewR4Service creates a new R4Service
func NewR4Service(logger *zap.Logger, r4Client *r4bank.RestClient, bankRepo repositories.BankRepository) R4Service {
	return &r4Service{
		r4Client: r4Client,
		bankRepo: bankRepo,
		Logger:   logger,
	}
}

// getDirectDebitResponse is a helper function to process the response from direct debit operations
func (r *r4Service) getDirectDebitResponse(ctx context.Context, id string) (*r4bank.GetOperationResponse, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	respCh := make(chan *r4bank.GetOperationResponse)

	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				opResp, err := r.GetOperationByID(ctxWithTimeout, id)
				if err != nil {
					r.Logger.Error(err.Error(), zap.Any("operationID", id))
					return
				}

				if opResp == nil {
					r.Logger.Error("nil response from GetOperationByID", zap.Any("operationID", id))
					return
				}

				if opResp.Code != inBreakpointCode {
					respCh <- opResp
					return
				}
			case <-ctxWithTimeout.Done():
				r.Logger.Error("context timeout while waiting for operation to complete", zap.Any("operationID", id))
				respCh <- nil
				return
			}
		}
	}()

	select {
	case opResp := <-respCh:
		if opResp == nil {
			return nil, errors.New("operation did not complete in time")
		}
		return opResp, nil

	case <-ctxWithTimeout.Done():
		r.Logger.Error("context timeout while waiting for operation to complete", zap.Any("operationID", id))
		return nil, errors.New("operation did not complete in time")
	}
}

// GetBCVTasaUSD retrieves the BCV exchange rate for USD
func (r *r4Service) GetBCVTasaUSD(ctx context.Context) (*models.BCVTasaUSDResponse, error) {
	currency := "USD"

	workDay, err := r.bankRepo.NextWorkDayBounded(ctx, time.Now())
	if err != nil {
		r.Logger.Error("failed to get next work day", zap.Error(err))
		return nil, fmt.Errorf("failed to get next work day: %w", err)
	}

	hmacInput := workDay.Format("2006-01-02") + currency
	payload := map[string]string{
		"Moneda":     currency,
		"Fechavalor": workDay.Format("2006-01-02"),
	}

	resp, err := r.r4Client.Do(ctx, hmacInput, payload, bcvTasaUSDEndpoint)
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

	resp, err := r.r4Client.Do(ctx, hmacInput, payload, changePaidEndpoint)
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

	resp, err := r.r4Client.Do(ctx, hmacInput, payload, generateOTPEndpoint)
	if err != nil {
		r.Logger.Error(err.Error(), zap.Any("payload", payload))
		return fmt.Errorf("error en request: %w", err)
	}

	var otpResp r4bank.OTPResponse
	if err := json.Unmarshal(resp, &otpResp); err != nil {
		r.Logger.Error(err.Error(), zap.Any("response", string(resp)))
		return fmt.Errorf("error decodificando respuesta: %w", err)
	}

	if otpResp.Code != sendOperationTrueCode {
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

	resp, err := r.r4Client.Do(ctx, hmacInput, payload, validateImmediateDebitEndpoint)
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
	resp, err := r.r4Client.Do(ctx, hmacInput, payload, getOperationByIDEndpoint)
	if err != nil {
		r.Logger.Error(err.Error(), zap.Any("payload", payload))
		return nil, err
	}

	var opResp r4bank.GetOperationResponse
	if err := json.Unmarshal(resp, &opResp); err != nil {
		r.Logger.Error(err.Error(), zap.Any("response", string(resp)))
		return nil, err
	}

	r.Logger.Debug("Operation response", zap.Any("operation", opResp))

	return &r4bank.GetOperationResponse{
		Code:      opResp.Code,
		Success:   opResp.Success,
		Reference: opResp.Reference,
	}, nil
}

// DirectDebitAccount debits a specified amount directly from a user's account
func (r *r4Service) DirectDebitAccount(ctx context.Context, req *models.DebitDirectAccountRequest) (*models.DebitDirectAccountResponse, error) {
	hmacInput := req.Account
	payload := map[string]string{
		"cuenta":   req.Account,
		"docId":    req.DNI,
		"monto":    fmt.Sprintf("%.2f", *req.Amount),
		"nombre":   req.Name,
		"concepto": req.Concept,
	}

	resp, err := r.r4Client.Do(ctx, hmacInput, payload, directDebitAccountEndpoint)
	if err != nil {
		r.Logger.Error(err.Error(), zap.Any("payload", payload))
		return nil, err
	}

	var debitResp r4bank.DebitDirectAccountResponse
	if err := json.Unmarshal(resp, &debitResp); err != nil {
		r.Logger.Error(err.Error(), zap.Any("response", string(resp)))
		return nil, err
	}

	if debitResp.Code != sendOperationTrueCode {
		r.Logger.Error("R4 Debit Direct Account API error", zap.String("code", debitResp.Code), zap.Any("payload", payload))
		return nil, errors.New("R4 Debit Direct Account API returned an error")
	}

	operation, err := r.getDirectDebitResponse(ctx, debitResp.UUID)
	if err != nil {
		r.Logger.Error(err.Error(), zap.Any("operationID", debitResp.UUID))
		return nil, fmt.Errorf("error obteniendo estado de la operación: %w", err)
	}

	return &models.DebitDirectAccountResponse{
		ID:        debitResp.UUID,
		Code:      operation.Code,
		Message:   debitResp.Message,
		Reference: operation.Reference,
		Success:   operation.Success,
	}, nil
}

func (r *r4Service) DirectDebitPhone(ctx context.Context, req *models.DebitDirectPhoneRequest) (*models.DebitDirectPhoneResponse, error) {
	hmacInput := req.Phone
	payload := map[string]string{
		"telefono": req.Phone,
		"docId":    req.DNI,
		"nombre":   req.Name,
		"banco":    req.Bank,
		"monto":    fmt.Sprintf("%.2f", *req.Amount),
		"concepto": req.Concept,
	}

	resp, err := r.r4Client.Do(ctx, hmacInput, payload, directDebitPhoneEndpoint)
	if err != nil {
		r.Logger.Error(err.Error(), zap.Any("payload", payload))
		return nil, err
	}

	var debitResp r4bank.DebitDirectPhoneResponse
	if err := json.Unmarshal(resp, &debitResp); err != nil {
		r.Logger.Error(err.Error(), zap.Any("response", string(resp)))
		return nil, err
	}

	var success bool
	if _, ok := _DebitInmediateSpecialResponse[debitResp.Code]; ok {
		success = true
	}

	return &models.DebitDirectPhoneResponse{
		ID:        debitResp.UUID,
		Code:      debitResp.Code,
		Message:   debitResp.Message,
		Reference: "",
		Success:   success,
	}, nil
}
