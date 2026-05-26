package services

import (
	"context"
	"log"
	"math/rand"
	"strconv"
	"time"

	"go_server/internal/models"
	"go_server/internal/repositories"
	"go_server/internal/websockets"

	"github.com/lib/pq"
)

type KenoService struct {
	Hub               *websockets.Hub
	UserRepo          *repositories.UserRepository
	KenoRepo          *repositories.KenoRepository
	DatCuocRepo       *repositories.DatCuocKenoRepository
	SystemSettingRepo *repositories.SystemSettingRepository
	CurrentPhien      int
	State             models.TinhTrangGame
	TimeLeft          int // in seconds
	// Channel to stop the game loop
	StopChan chan struct{}
}

const TIMER = 20

func NewKenoService(hub *websockets.Hub, userRepo *repositories.UserRepository, kenoRepo *repositories.KenoRepository, datCuocRepo *repositories.DatCuocKenoRepository, systemSettingRepo *repositories.SystemSettingRepository) *KenoService {
	return &KenoService{
		Hub:               hub,
		UserRepo:          userRepo,
		KenoRepo:          kenoRepo,
		DatCuocRepo:       datCuocRepo,
		SystemSettingRepo: systemSettingRepo,
		CurrentPhien:      0,
		State:             models.TinhTrangDangCho,
		TimeLeft:          TIMER,
		StopChan:          make(chan struct{}),
	}
}

// StartGameLoop starts the main game loop using Goroutines and Tickers
func (g *KenoService) StartGameLoop(ctx context.Context) {
	g.LoadGame()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Game loop stopped gracefully")
			return
		case <-g.StopChan:
			log.Println("Game loop stopped by StopChan")
			return
		case <-ticker.C:
			g.tick()
		}
	}
}

func (g *KenoService) LoadGame() {
	// get latest game
	latestGame, err := g.KenoRepo.GetLatestGameUnfinished()
	if err != nil {
		g.StopChan <- struct{}{}
		return
	}

	if latestGame == nil {
		log.Println("Khong co phien nao")
		lastestFinishedGame, err := g.KenoRepo.GetLatestGameFinished()
		if err != nil {
			g.StopChan <- struct{}{}
			return
		}
		if lastestFinishedGame == nil {
			g.createNewPhien()
			return
		}
		g.CurrentPhien = lastestFinishedGame.Phien
		g.State = lastestFinishedGame.TinhTrang
		return
	}
	log.Println("Phien hien tai:", latestGame.Phien, latestGame.ID)
	g.CurrentPhien = latestGame.Phien
	g.State = latestGame.TinhTrang

}

func (g *KenoService) tick() {

	g.TimeLeft--

	g.handleState()
}

func (g *KenoService) generateKetQua() pq.Int64Array {
	ketQua := make(pq.Int64Array, 5)
	for i := range 5 {
		randomNumber := int64(rand.Intn(9) + 1)
		ketQua[i] = randomNumber
	}
	return ketQua
}

func (g *KenoService) resetGameState() {
	g.State = models.TinhTrangDangCho
	g.TimeLeft = TIMER
}

func (g *KenoService) createNewPhien() *models.Keno {
	g.CurrentPhien++
	now := time.Now()
	// Create new game in db
	game := &models.Keno{
		Phien:     g.CurrentPhien,
		KetQua:    pq.Int64Array{}, // empty array for now
		TinhTrang: models.TinhTrangDangCho,
		Timestamp: &now,
	}
	if err := g.KenoRepo.CreateGame(game); err != nil {
		log.Println("Error creating game:", err)
		// Stop the game loop
		g.StopChan <- struct{}{}
		return nil
	}
	g.resetGameState()

	return game
}

func (g *KenoService) handleState() {
	switch g.State {
	case models.TinhTrangDangCho:
		log.Println("Time Left:", g.TimeLeft)

		if g.TimeLeft <= 0 {
			g.State = models.TinhTrangDangQuay
			g.KenoRepo.UpdateGame(g.CurrentPhien, map[string]interface{}{
				"tinh_trang": models.TinhTrangDangQuay,
			})
		}
	case models.TinhTrangDangQuay:
		ketQua := g.generateKetQua()
		// log ketQua
		log.Println("Ket Qua:", ketQua)
		// Update database
		if err := g.KenoRepo.UpdateGame(g.CurrentPhien, map[string]interface{}{
			"ket_qua":    ketQua,
			"tinh_trang": models.TinhTrangDangTraThuong,
		}); err != nil {
			log.Println("Error updating game:", err)
			// Stop the game loop
			g.StopChan <- struct{}{}
			return
		}
		g.State = models.TinhTrangDangTraThuong

	case models.TinhTrangDangTraThuong:
		// TODO: Handle awarding
		// Sleep 3s to assume awarded
		log.Println("Dang Tra Thuong")
		g.traThuong()
		g.State = models.TinhTrangHoanTat
		g.KenoRepo.UpdateGame(g.CurrentPhien, map[string]interface{}{
			"tinh_trang": models.TinhTrangHoanTat,
		})

	case models.TinhTrangHoanTat:
		g.createNewPhien()

	}
}

func (g *KenoService) getSystemSetting() map[string]float64 {
	setting, err := g.SystemSettingRepo.FindByType(models.TI_LE_KENO)
	if err != nil {
		log.Println("Error getting system setting:", err)
		return nil
	}
	if setting == nil {
		log.Println("System setting not found")
		return nil
	}
	log.Println("System Setting: ", setting.Setting1, setting.Setting2)
	val, err := strconv.ParseFloat(setting.Setting1, 64)
	if err != nil {
		log.Println("Error parsing Setting1 to float64:", err)
		return nil
	}
	settingMap := make(map[string]float64)

	settingMap["tiLeTraThuong"] = val
	return settingMap
}

func (g *KenoService) traThuong() {
	currentPhien := g.CurrentPhien
	game, _ := g.KenoRepo.GetGameByPhien(currentPhien)
	if game == nil {
		log.Println("Error getting game by phien")
		return
	}

	settingMap := g.getSystemSetting()

	lichSuDatCuocs, _ := g.DatCuocRepo.GetLichSuDatCuoc(game.ID)

	mapKetQua := g.mapKetQua(game.KetQua)

	log.Println("Map Ket Qua: ", mapKetQua)

	for _, lichSuDatCuoc := range lichSuDatCuocs {
		tongThang := int64(0)
		datCuocs, _ := g.DatCuocRepo.GetListDatCuocByLichSuId(lichSuDatCuoc.ID)
		for _, datCuoc := range datCuocs {
			loaiBi := datCuoc.LoaiBi
			loaiCuoc := datCuoc.LoaiCuoc
			kq := mapKetQua[int64(loaiBi)][0]
			if kq == loaiCuoc {
				log.Println("Thang")
				tienThang := int64(float64(datCuoc.TienCuoc) * (settingMap["tiLeTraThuong"]))
				tongThang += tienThang
				// Update chi tiet
				g.DatCuocRepo.UpdateChiTietDatCuoc(datCuoc.ID, map[string]interface{}{
					"tien_thang": tienThang,
				})
				log.Println("Tien Thang: ", tienThang)

			} else {
				log.Println("Thua")

			}

			log.Println("DatCuoc: ", datCuoc.ID, datCuoc.LoaiBi, datCuoc.LoaiCuoc, datCuoc.TienCuoc)
		}
		// Update tien thang cho lich su
		g.DatCuocRepo.UpdateLichSuDatCuoc(lichSuDatCuoc.ID, map[string]interface{}{
			"tinh_trang": models.TinhTrangHoanTat,
		})
	}

}

func (g *KenoService) mapKetQua(ketQua []int64) map[int64][]string {
	cacKetQua := map[int64][]string{
		1: make([]string, 1),
		2: make([]string, 1),
		3: make([]string, 1),
		4: make([]string, 1),
		5: make([]string, 1),
	}

	for index, kq := range ketQua {
		if kq%2 == 0 {
			cacKetQua[int64(index+1)] = []string{"C"}
		} else {
			cacKetQua[int64(index+1)] = []string{"L"}
		}
	}
	return cacKetQua

}
