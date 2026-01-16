package models

type R4AppaMobilePaymentPreview struct {
	ID     int     `gorm:"primaryKey;autoIncrement" json:"id"`
	Amount float64 `json:"amount"`
}

func (R4AppaMobilePaymentPreview) TableName() string {
	return "r4_appa_mobile_payments_previews"
}
