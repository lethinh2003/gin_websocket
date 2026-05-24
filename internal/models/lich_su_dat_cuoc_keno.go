package models

import (
	"time"

	"gorm.io/gorm"
)

// LichSuDatCuocKeno represents the main bet ticket (parent table)
type LichSuDatCuocKeno struct {
	ID        string        `gorm:"primaryKey;type:varchar(36);default:gen_random_uuid()" json:"id"`
	KenoID    string        `gorm:"type:varchar(36);uniqueIndex:idx_keno_user_unique;not null" json:"kenoId"`
	Keno      Keno          `gorm:"foreignKey:KenoID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"keno,omitempty"`
	Phien     int           `gorm:"type:integer;not null" json:"phien"`
	UserID    string        `gorm:"type:varchar(36);uniqueIndex:idx_keno_user_unique;not null" json:"userId"`
	User      User          `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"user,omitempty"`
	TinhTrang TinhTrangGame `gorm:"type:varchar(20);default:'DANG_CHO'" json:"tinhTrang"`

	// Has-Many Relationship with child table
	DatCuoc []ChiTietDatCuocKeno `gorm:"foreignKey:LichSuID;constraint:OnDelete:CASCADE;" json:"datCuoc"`

	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// ChiTietDatCuocKeno represents individual bet details (child table)
type ChiTietDatCuocKeno struct {
	ID        string    `gorm:"primaryKey;type:varchar(36);default:gen_random_uuid()" json:"id"`
	LichSuID  string    `gorm:"type:varchar(36);index;not null" json:"lichSuId"`
	LoaiBi    int       `gorm:"type:smallint;not null;check:loai_bi BETWEEN 1 AND 5" json:"loaiBi"`     // constraint 1-5
	LoaiCuoc  string    `gorm:"type:varchar(2);not null;check:loai_cuoc IN ('C', 'L')" json:"loaiCuoc"` // C or L
	TienCuoc  int       `gorm:"type:bigint;default:0" json:"tienCuoc"`
	CreatedAt time.Time `json:"createdAt"`
}
