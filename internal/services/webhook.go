package services

import (
	"bone_appetit_r4_service/internal/models"
	dbModels "bone_appetit_r4_service/pkg/db/models"
	"fmt"
	"strconv"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type WebhookService interface {
	RegisterR4MobilePayment(payment *models.R4NotificaRequest, storeName string) error
	RegisterR4MobilePaymentPreview(preview *models.R4ConsultaRequest, storeName string) error
}

type webhookService struct {
	db     *gorm.DB
	logger *zap.Logger
	loc    *time.Location
}

func NewWebhookService(db *gorm.DB, loc *time.Location, logger *zap.Logger) WebhookService {
	return &webhookService{db: db, loc: loc, logger: logger}
}

// RegisterR4MobilePaymentPreview registers a new R4 mobile payment preview in the database
func (s *webhookService) RegisterR4MobilePaymentPreview(preview *models.R4ConsultaRequest, storeName string) error {
	amount, err := strconv.ParseFloat(preview.Monto, 64)
	if err != nil {
		s.logger.Error("failed to parse preview amount", zap.Error(err))
		return err
	}

	switch storeName {
	case "bone":
		if err := s.createR4BoneMobilePaymentPreview(amount); err != nil {
			s.logger.Error("failed to register R4 mobile payment preview", zap.Error(err))
			return err
		}
	case "appa":
		if err := s.createR4AppaMobilePaymentPreview(amount); err != nil {
			s.logger.Error("failed to register R4 Appa mobile payment preview", zap.Error(err))
			return err
		}
	default:
		return fmt.Errorf("unknown store name: %s", storeName)
	}

	return nil
}

// RegisterR4MobilePaymentProcess registers a new R4 mobile payment in the database
func (s *webhookService) RegisterR4MobilePayment(payment *models.R4NotificaRequest, storeName string) error {

	if exist, err := s.existReference(payment.Referencia, storeName); err != nil {
		s.logger.Error("failed to check existing reference", zap.Error(err))
		return err
	} else if exist {
		s.logger.Info("R4 mobile payment already registered", zap.String("reference", payment.Referencia))
		return nil
	}

	bank := payment.BancoEmisor
	if len(bank) == 3 {
		bank = fmt.Sprintf("0%s", payment.BancoEmisor)
	}

	amount, err := strconv.ParseFloat(payment.Monto, 64)
	if err != nil {
		s.logger.Error("failed to parse payment amount", zap.Error(err))
		return err
	}

	switch storeName {
	case "bone":
		if err := s.createR4BoneMobilePayment(payment, bank, amount); err != nil {
			s.logger.Error("failed to register R4 mobile payment", zap.Error(err))
			return err
		}
	case "appa":
		if err := s.createR4AppaMobilePayment(payment, bank, amount); err != nil {
			s.logger.Error("failed to register R4 Appa mobile payment", zap.Error(err))
			return err
		}
	default:
		return fmt.Errorf("unknown store name: %s", storeName)
	}

	return nil
}

func (s *webhookService) createR4BoneMobilePaymentPreview(amount float64) error {
	return s.db.Create(&dbModels.R4BoneMobilePaymentPreview{
		Amount: amount,
	}).Error
}

func (s *webhookService) createR4AppaMobilePaymentPreview(amount float64) error {
	return s.db.Create(&dbModels.R4AppaMobilePaymentPreview{
		Amount: amount,
	}).Error
}

// createR4BoneMobilePayment registers a new R4 mobile payment in the database
func (s *webhookService) createR4BoneMobilePayment(
	payment *models.R4NotificaRequest,
	bank string,
	amount float64,
) error {
	return s.db.Create(&dbModels.R4MobilePayment{
		IDCommerce:    payment.IdComercio,
		CommercePhone: payment.TelefonoComercio,
		SenderPhone:   payment.TelefonoEmisor,
		IssuingBank:   bank,
		Amount:        amount,
		Reference:     payment.Referencia,
		Date:          time.Now().In(s.loc),
		OrderID:       nil,
	}).Error
}

// createR4BoneMobilePayment registers a new R4 mobile payment in the database
func (s *webhookService) createR4AppaMobilePayment(
	payment *models.R4NotificaRequest,
	bank string,
	amount float64,
) error {
	return s.db.Create(&dbModels.R4AppaMobilePayment{
		IDCommerce:    payment.IdComercio,
		CommercePhone: payment.TelefonoComercio,
		SenderPhone:   payment.TelefonoEmisor,
		IssuingBank:   bank,
		Amount:        amount,
		Reference:     payment.Referencia,
		Date:          time.Now().In(s.loc),
		OrderID:       nil,
	}).Error
}

// existReference checks if a reference already exists in the specified table
func (s *webhookService) existReference(reference, storeName string) (bool, error) {
	var (
		tableName string
		count     int64
	)

	switch storeName {
	case "bone":
		tableName = "r4_mobile_payments"
	case "appa":
		tableName = "r4_appa_mobile_payments"
	default:
		return false, fmt.Errorf("unknown store name: %s", storeName)
	}

	if err := s.db.Table(tableName).Where("reference = ?", reference).Count(&count).Error; err != nil {
		return false, err
	}

	if count > 0 {
		return true, nil
	}

	return false, nil
}
