package repositories

import (
	"errors"
	"go_server/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DatCuocKenoRepository struct {
	db *gorm.DB
}

func NewDatCuocKenoRepository(db *gorm.DB) *DatCuocKenoRepository {
	return &DatCuocKenoRepository{db: db}
}

func (r *DatCuocKenoRepository) Create(model *models.LichSuDatCuocKeno) error {
	return r.db.Create(model).Error
}

func (r *DatCuocKenoRepository) GetByKenoIDAndUserID(kenoID string, userID string) (*models.LichSuDatCuocKeno, error) {
	var model models.LichSuDatCuocKeno
	if err := r.db.Where("keno_id = ? AND user_id = ?", kenoID, userID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &model, nil
}

func (r *DatCuocKenoRepository) AddNewCuocByLichSuID(lichSuID string, datCuoc []models.ChiTietDatCuocKeno) (error, []models.ChiTietDatCuocKeno) {
	// 1. Manually assign the foreign key (LichSuID) to each item
	for i := range datCuoc {
		datCuoc[i].LichSuID = lichSuID
	}

	// 2. Perform a clean bulk insert
	if err := r.db.Create(&datCuoc).Error; err != nil {
		return err, nil
	}
	return nil, datCuoc
}

func (r *DatCuocKenoRepository) PlaceBet(userId string, phien int, loaiBi int, loaiCuoc string, tienCuoc int) error {
	// Execute everything within a database transaction
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Fetch and Lock the User's record to prevent race conditions
		var user models.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, "id = ?", userId).Error; err != nil {
			return errors.New("người dùng không tồn tại")
		}
		// 2. Fetch the game session
		var game models.Keno
		if err := tx.First(&game, "phien = ?", phien).Error; err != nil {
			return errors.New("phiên không tồn tại")
		}
		// 3. Validate game state
		if game.TinhTrang != models.TinhTrangDangCho {
			return errors.New("phiên đã kết thúc")
		}
		// 4. Validate user's balance
		if user.Money < tienCuoc {
			return errors.New("số dư không đủ")
		}
		// 5. Deduct user balance
		if err := tx.Model(&user).Update("money", user.Money-tienCuoc).Error; err != nil {
			return err
		}
		// 6. Check if user already placed a bet ticket for this session
		var lichSu models.LichSuDatCuocKeno
		err := tx.Where("keno_id = ? AND user_id = ?", game.ID, user.ID).First(&lichSu).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create a new parent bet ticket and its first detail entry
			newDatCuoc := &models.LichSuDatCuocKeno{
				KenoID:    game.ID,
				Phien:     game.Phien,
				UserID:    user.ID,
				TinhTrang: models.TinhTrangDangCho,
				DatCuoc: []models.ChiTietDatCuocKeno{
					{
						LoaiBi:   loaiBi,
						LoaiCuoc: loaiCuoc,
						TienCuoc: tienCuoc,
					},
				},
			}
			return tx.Create(newDatCuoc).Error
		} else if err == nil {
			// Add a new bet detail child entry to the existing ticket
			newDetail := models.ChiTietDatCuocKeno{
				LichSuID: lichSu.ID,
				LoaiBi:   loaiBi,
				LoaiCuoc: loaiCuoc,
				TienCuoc: tienCuoc,
			}
			return tx.Create(&newDetail).Error
		}
		return err
	})
}
