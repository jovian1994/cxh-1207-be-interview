package models

type TaskModel struct{}

func (TaskModel) TableName() string {
	return "tasks"
}
