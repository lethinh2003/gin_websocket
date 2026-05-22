package services

import (
	"log"
	"math/rand"
	"time"

	"go_server/internal/models"
	"go_server/internal/repositories"
	"go_server/internal/websockets"
)

type GameService struct {
	Hub          *websockets.Hub
	UserRepo     *repositories.UserRepository
	GameRepo     *repositories.GameRepository
	CurrentPhien int
	State        models.TinhTrangGame
	TimeLeft     int // in seconds
}

func NewGameService(hub *websockets.Hub, userRepo *repositories.UserRepository, gameRepo *repositories.GameRepository) *GameService {
	return &GameService{
		Hub:          hub,
		UserRepo:     userRepo,
		GameRepo:     gameRepo,
		CurrentPhien: 1,
		State:        models.TinhTrangDangCho,
		TimeLeft:     59,
	}
}

// StartGameLoop starts the main game loop using Goroutines and Tickers
func (g *GameService) StartGameLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			g.tick()
		}
	}
}

func (g *GameService) tick() {
	g.TimeLeft--

	if g.TimeLeft <= 0 {
		g.State = models.TinhTrangDangQuay
		g.TimeLeft = 59

		// generate random number from 1 to 9
		ketQua := []int{}
		for range 5 {
			randomNumber := rand.Intn(9) + 1
			ketQua = append(ketQua, randomNumber)
		}
		// log ketQua
		log.Println("Ket Qua:", ketQua)
	} else {
		g.State = models.TinhTrangDangCho
		log.Println("Time Left:", g.TimeLeft)

	}
}

func (g *GameService) generateKetQua() []int {
	ketQua := []int{}
	for range 5 {
		randomNumber := rand.Intn(9) + 1
		ketQua = append(ketQua, randomNumber)
	}
	return ketQua
}

func (g *GameService) handleState() {
	switch g.State {
	case models.TinhTrangDangCho:
	case models.TinhTrangDangQuay:
		ketQua := g.generateKetQua()
		// log ketQua
		log.Println("Ket Qua:", ketQua)
	case models.TinhTrangDangTraThuong:
	case models.TinhTrangHoanTat:
	}
}
