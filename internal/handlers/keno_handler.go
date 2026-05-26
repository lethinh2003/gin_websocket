package handlers

import (
	"errors"
	"log"
	"net/http"

	"go_server/internal/models"
	"go_server/internal/repositories"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type KenoHandler struct {
	kenoRepo    *repositories.KenoRepository
	datCuocRepo *repositories.DatCuocKenoRepository
	userRepo    *repositories.UserRepository
}

func NewKenoHandler(kenoRepo *repositories.KenoRepository, datCuocRepo *repositories.DatCuocKenoRepository, userRepo *repositories.UserRepository) *KenoHandler {
	return &KenoHandler{kenoRepo: kenoRepo, datCuocRepo: datCuocRepo, userRepo: userRepo}
}

type CreateCuocRequest struct {
	LoaiBi   int    `json:"loaiBi" binding:"required,min=1,max=5"`
	LoaiCuoc string `json:"loaiCuoc" binding:"required,oneof=C L"`
	TienCuoc int    `json:"tienCuoc" binding:"required,min=1000"`
	Phien    int    `json:"phien" binding:"required"`
}

func (h *KenoHandler) CreateCuoc(c *gin.Context) {
	// Create a new bet ticket
	var req CreateCuocRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			out := make(map[string]string)
			for _, fe := range ve {
				out[fe.Field()] = getErrorMsg(fe)
			}
			c.JSON(http.StatusBadRequest, gin.H{"errors": out})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	game, err := h.kenoRepo.GetGameByPhien(req.Phien)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if game == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Phiên không tồn tại"})
		return
	}
	if game.TinhTrang != models.TinhTrangDangCho {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Phiên đã kết thúc"})
		return
	}

	userId := c.MustGet("userId").(string)

	err = h.datCuocRepo.PlaceBet(userId, req.Phien, req.LoaiBi, req.LoaiCuoc, req.TienCuoc)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Cược thành công",
	})
}

func getErrorMsg(fe validator.FieldError) string {
	log.Println(fe.Tag())

	switch fe.Field() {
	case "LoaiBi":
		if fe.Tag() == "required" {
			return "Loại bi là bắt buộc"
		}
		return "Loại bi phải từ 1 đến 5"
	case "LoaiCuoc":
		if fe.Tag() == "required" {
			return "Loại cược là bắt buộc"
		}
		return "Loại cược phải là một trong các giá trị: C (Chẵn), L (Lẻ)"
	case "TienCuoc":
		if fe.Tag() == "required" {
			return "Tiền cược là bắt buộc"
		}
		return "Tiền cược phải từ 1.000 trở lên"
	case "Phien":
		return "Mã phiên là bắt buộc"
	}
	return "Dữ liệu không hợp lệ"
}
