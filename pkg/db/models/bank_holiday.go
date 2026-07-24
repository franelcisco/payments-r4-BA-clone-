package models

import "time"

type BankHoliday struct {
	ID        int       `gorm:"primaryKey;autoIncrement" json:"id"`
	Date      time.Time `gorm:"type:date;not null;unique" json:"date"`
	Year      int       `gorm:"not null" json:"year"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
