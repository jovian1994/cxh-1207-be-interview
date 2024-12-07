package models

import (
	"gorm.io/gorm"
)

type TaskModel struct {
	gorm.Model
	Status     int    `gorm:"column:status"`
	CreateBy   string `gorm:"column:create_by"`
	ResultKey  string `gorm:"column:result_key"`
	IsOss      int    `gorm:"column:is_oss"`
	Content    string `gorm:"column:content"`
	Lang       string `gorm:"column:lang"`
	TargetLang string `gorm:"column:target_lang"`
}

func (TaskModel) TableName() string {
	return "tasks"
}
