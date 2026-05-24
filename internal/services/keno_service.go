package services

import (
	"context"
	"log"
	"math/rand"
	"time"

	"go_server/internal/models"
	"go_server/internal/repositories"
	"go_server/internal/websockets"

	"github.com/lib/pq"
)

type KenoService struct {
	Hub          *websockets.Hub
	UserRepo     *repositories.UserRepository
	KenoRepo     *repositories.KenoRepository
	CurrentPhien int
	State        models.TinhTrangGame
	TimeLeft     int // in seconds
	// Channel to stop the game loop
	StopChan chan struct{}
}

const TIMER = 1000

func NewKenoService(hub *websockets.Hub, userRepo *repositories.UserRepository, kenoRepo *repositories.KenoRepository) *KenoService {
	return &KenoService{
		Hub:          hub,
		UserRepo:     userRepo,
		KenoRepo:     kenoRepo,
		CurrentPhien: 0,
		State:        models.TinhTrangDangCho,
		TimeLeft:     TIMER,
		StopChan:     make(chan struct{}),
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
		time.Sleep(3 * time.Second)
		g.State = models.TinhTrangHoanTat
		g.KenoRepo.UpdateGame(g.CurrentPhien, map[string]interface{}{
			"tinh_trang": models.TinhTrangHoanTat,
		})

	case models.TinhTrangHoanTat:
		g.createNewPhien()

	}
}
