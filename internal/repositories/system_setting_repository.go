package repositories

import (
	"go_server/internal/models"

	"gorm.io/gorm"
)

type SystemSettingRepository struct {
	db *gorm.DB
}

func NewSystemSettingRepository(db *gorm.DB) *SystemSettingRepository {
	return &SystemSettingRepository{db: db}
}

func (r *SystemSettingRepository) FindByType(settingType models.SettingType) (*models.SystemSetting, error) {
	var setting models.SystemSetting
	err := r.db.Where("type = ?", settingType).First(&setting).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &setting, nil
}
