package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Id        int `gorm:"uniqueIndex"`
	Name      string
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time
}
