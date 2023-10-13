package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID            uint      `gorm:"primaryKey;autoIncrement;not null"`
	Name          string    `gorm:"not null"`
	Email         string    `gorm:"not null;unique"`
	Age           uint8     `gorm:"not null"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	Password      string    `gorm:"not null"`
	UserTypeRefer uint
	UserType      UserType `gorm:"foreignKey:UserTypeRefer"`
}

type UserType struct {
	ID       uint   `gorm:"primaryKey;autoIncrement;not null"`
	CodeName string `gorm:"not null;unique"`
}
