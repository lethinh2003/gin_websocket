package models

import (
	"time"

	"gorm.io/gorm"
)

type UserRole string

const (
	RoleUser  UserRole = "USER"
	RoleAdmin UserRole = "ADMIN"
)

type User struct {
	ID             string         `gorm:"primaryKey;type:varchar(36);default:gen_random_uuid()" json:"id"`
	Name           string         `json:"name"`
	Email          string         `gorm:"uniqueIndex" json:"email"`
	Role           *UserRole      `gorm:"default:'USER'" json:"role"`
	Money          int            `gorm:"default:0" json:"money"`
	HashedPassword string         `json:"-"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}
