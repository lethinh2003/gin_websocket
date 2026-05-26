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

func (r *DatCuocKenoRepository) GetLichSuDatCuoc(kenoID string) ([]models.LichSuDatCuocKeno, error) {
	var models []models.LichSuDatCuocKeno
	if err := r.db.Where("keno_id = ?", kenoID).Find(&models).Error; err != nil {
		return nil, err
	}
	return models, nil
}

func (r *DatCuocKenoRepository) GetListDatCuocByLichSuId(lichSuID string) ([]models.ChiTietDatCuocKeno, error) {
	var models []models.ChiTietDatCuocKeno
	if err := r.db.Where("lich_su_id = ?", lichSuID).Find(&models).Error; err != nil {
		return nil, err
	}
	return models, nil
}

func (r *DatCuocKenoRepository) UpdateLichSuDatCuoc(id string, updates map[string]interface{}) {
	r.db.Model(&models.LichSuDatCuocKeno{}).Where("id = ?", id).Updates(updates)
}

func (r *DatCuocKenoRepository) UpdateChiTietDatCuoc(id string, updates map[string]interface{}) {
	r.db.Model(&models.ChiTietDatCuocKeno{}).Where("id = ?", id).Updates(updates)
}

/*
func (r *DatCuocKenoRepository) ProcessPayout(gameID string, mapKetQua map[int64][]string, settingMap map[string]float64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Get all pending parent bet tickets for this game session
		var tickets []models.LichSuDatCuocKeno
		if err := tx.Where("keno_id = ? AND tinh_trang = ?", gameID, models.TinhTrangDangCho).Find(&tickets).Error; err != nil {
			return err
		}

		for _, ticket := range tickets {
			// Get all detail bets for this ticket
			var details []models.ChiTietDatCuocKeno
			if err := tx.Where("lich_su_id = ?", ticket.ID).Find(&details).Error; err != nil {
				return err
			}

			totalWin := 0
			for _, bet := range details {
				results, exists := mapKetQua[int64(bet.LoaiBi)]
				won := false
				if exists {
					for _, res := range results {
						if res == bet.LoaiCuoc {
							won = true
							break
						}
					}
				}

				tienThang := 0
				if won {
					multiplier, ok := settingMap[bet.LoaiCuoc]
					if !ok {
						multiplier = 1.95
					}
					tienThang = int(float64(bet.TienCuoc) * multiplier)
				}

				totalWin += tienThang

				// Update detail bet record with the winning amount
				if err := tx.Model(&bet).Update("tien_thang", tienThang).Error; err != nil {
					return err
				}
			}

			// Update parent ticket status to HOAN_TAT
			if err := tx.Model(&ticket).Update("tinh_trang", models.TinhTrangHoanTat).Error; err != nil {
				return err
			}

			// If the user won any bets, update their wallet balance
			if totalWin > 0 {
				var user models.User
				if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, "id = ?", ticket.UserID).Error; err != nil {
					return err
				}
				if err := tx.Model(&user).Update("money", user.Money+totalWin).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

*/
