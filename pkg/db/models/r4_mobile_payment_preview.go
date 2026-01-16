package models

type R4BoneMobilePaymentPreview struct {
	ID     int     `gorm:"primaryKey;autoIncrement" json:"id"`
	Amount float64 `json:"amount"`
}

func (R4BoneMobilePaymentPreview) TableName() string {
	return "r4_mobile_payments_previews"
}
