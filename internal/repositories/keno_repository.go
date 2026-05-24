package repositories

import (
	"errors"
	"go_server/internal/models"

	"gorm.io/gorm"
)

type KenoRepository struct {
	db *gorm.DB
}

func NewKenoRepository(db *gorm.DB) *KenoRepository {
	return &KenoRepository{db: db}
}

func (r *KenoRepository) CreateGame(game *models.Keno) error {
	return r.db.Create(game).Error
}

func (r *KenoRepository) GetLatestGameUnfinished() (*models.Keno, error) {
	var game models.Keno
	if err := r.db.Where("tinh_trang = ?", models.TinhTrangDangCho).Order("phien desc").First(&game).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &game, nil
}
func (r *KenoRepository) GetLatestGameFinished() (*models.Keno, error) {
	var game models.Keno
	if err := r.db.Where("tinh_trang = ?", models.TinhTrangHoanTat).Order("phien desc").First(&game).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &game, nil
}

func (r *KenoRepository) UpdateGame(phien int, updates map[string]interface{}) error {
	return r.db.Model(&models.Keno{}).Where("phien = ?", phien).Updates(updates).Error
}

func (r *KenoRepository) GetGameByPhien(phien int) (*models.Keno, error) {
	var game models.Keno
	if err := r.db.Where("phien = ?", phien).First(&game).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &game, nil
}
