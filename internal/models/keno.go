package models

import (
	"fmt"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Keno struct {
	ID        string         `gorm:"primaryKey;type:varchar(36);default:gen_random_uuid()" json:"id"`
	Phien     int            `gorm:"uniqueIndex" json:"phien"`
	KetQua    pq.Int64Array  `gorm:"type:integer[]" json:"ketQua"` // Mảng các số nguyên từ 1 đến 9
	TinhTrang TinhTrangGame  `json:"tinhTrang"`
	Timestamp *time.Time     `json:"timestamp"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (k *Keno) BeforeSave(tx *gorm.DB) (err error) {
	if len(k.KetQua) != 5 {
		return fmt.Errorf("kết quả phiên Keno phải chứa đúng 5 phần tử")
	}
	return nil
}
