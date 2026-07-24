package repositories

import (
	"context"
	"errors"
	"time"
 "fmt"

	"bone_appetit_r4_service/pkg/db/models"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type BankRepository interface {
	NextWorkDayBounded(ctx context.Context, date time.Time) (time.Time, error)
}

type bankRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewBankRepository(db *gorm.DB, logger *zap.Logger) BankRepository {
	return &bankRepository{
		db:     db,
		logger: logger,
	}
}

func (b *bankRepository) NextWorkDayBounded(ctx context.Context, start time.Time) (time.Time, error) {
	limitDate := start.AddDate(0, 0, 14)
	var holidays []models.BankHoliday

	// Corrección: Usar el contexto y manejar errores de BD
	err := b.db.WithContext(ctx).
		Where("date > ? AND date <= ?", start, limitDate).
		Find(&holidays).Error
	if err != nil {
		return time.Time{}, fmt.Errorf("error consultando feriados: %w", err)
	}

	holidayMap := make(map[string]bool)
	for _, h := range holidays {
		// time.DateOnly equivale a "2006-01-02" (disponible desde Go 1.20)
		holidayMap[h.Date.Format(time.DateOnly)] = true
	}

		nextDay := start

	for i := 1; i <= 14; i++ {
		dateStr := nextDay.Format(time.DateOnly)
		isWeekend := nextDay.Weekday() == time.Saturday || nextDay.Weekday() == time.Sunday
		isHoliday := holidayMap[dateStr]

		if !isWeekend && !isHoliday {
			b.logger.Info("Día laborable encontrado", zap.String("fecha_resultado", dateStr))
			return nextDay, nil
		}

		nextDay = nextDay.AddDate(0, 0, 1)
	}


	return time.Time{}, errors.New("límite de 14 días superado: revisar configuración de feriados en BD")
}
