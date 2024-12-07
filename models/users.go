package models

import "gorm.io/gorm"

type UserModel struct {
	gorm.Model
	Username string `gorm:"column:username"`
	Password string `gorm:"column:password"`
	Role     string `gorm:"column:role"`
	Version  int    `gorm:"column:version"`
}

func (UserModel) TableName() string {
	return "users"
}
