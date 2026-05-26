package models

import (
	"time"

	"gorm.io/gorm"
)

type SettingType string

const (
	TI_LE_KENO SettingType = "ti_le_keno"
)

type SystemSetting struct {
	ID          string      `gorm:"primaryKey;type:varchar(36);default:gen_random_uuid()" json:"id"`
	Type        SettingType `gorm:"type:varchar(20);not null" json:"type"`
	Setting1    string      `gorm:"type:varchar(20);not null" json:"setting1"`
	Setting2    string      `gorm:"type:varchar(20);not null" json:"setting2"`
	SettingNote string
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// SeedSystemSettings seeds initial values for SystemSetting if they do not exist
func SeedSystemSettings(db *gorm.DB) error {
	defaultSettings := []SystemSetting{
		{
			Type:        TI_LE_KENO,
			Setting1:    "1.95",
			Setting2:    "",
			SettingNote: "Tỉ lệ phát thưởng game Keno",
		},
	}

	for _, setting := range defaultSettings {
		var count int64
		err := db.Model(&SystemSetting{}).
			Where("type = ?", setting.Type).
			Count(&count).Error
		if err != nil {
			return err
		}
		if count == 0 {
			if err := db.Create(&setting).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
