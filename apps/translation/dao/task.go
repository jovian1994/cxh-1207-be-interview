package dao

import (
	"github.com/jovian1994/cxh-1207-be-interview/models"
	"github.com/jovian1994/cxh-1207-be-interview/pkg/mysql_tool"
	"github.com/jovian1994/cxh-1207-be-interview/pkg/unify_response"
)

type ITaskDao interface {
	CreateTask(
		createBy string,
		lang string,
		content string, targetLang string) error
	GetTaskByIdAndUsername(username string, taskId int64) (*models.TaskModel, error)
	UpdateTaskStatus(taskId int64, updates map[string]any) error
}

type taskDao struct {
	dbClientName string
	db           *mysql_tool.DB
}

func NewTaskDao(dbClientName string) ITaskDao {
	return &taskDao{}
}

func (t *taskDao) CreateTask(createBy string, lang string,
	content string, targetLang string) error {

	task := models.TaskModel{}
	task.CreateBy = createBy
	task.Lang = lang
	task.Content = content
	task.TargetLang = targetLang
	task.Status = 0
	if lang == "" {
		lang = models.AutoDetect
	}
	err := t.getDBClient().Create(&task).Error
	if err != nil {
		return unify_response.DBError(err.Error())
	}
	return nil
}

func (t *taskDao) GetTaskByIdAndUsername(username string, taskId int64) (*models.TaskModel, error) {

	var task models.TaskModel
	err := t.getDBClient().
		Where("id = ?", taskId).
		Where("create_by =?", username).
		Scan(&task).Error
	if err != nil {
		return nil, unify_response.DBError(err.Error())
	}
	if task.ID == 0 {
		return nil, unify_response.NotFound()
	}
	return &task, nil

}

func (t *taskDao) GetTaskDetail(taskId int64) (*models.TaskModel, error) {

	var task models.TaskModel
	err := t.getDBClient().
		Where("id = ?", taskId).
		Scan(&task).Error
	if err != nil {
		return nil, unify_response.DBError(err.Error())
	}
	if task.ID == 0 {
		return nil, unify_response.NotFound()
	}
	return &task, nil
}

func (t *taskDao) getDBClient() *mysql_tool.DB {
	if t.db != nil {
		return t.db
	}
	t.db = mysql_tool.GetMysqlClient(t.dbClientName)
	return t.db
}

func (t *taskDao) UpdateTaskStatus(taskId int64, updates map[string]any) error {
	err := t.getDBClient().
		Model(updates).
		Where("id = ?", taskId).
		Updates(updates).Error
	if err != nil {
		return unify_response.DBError(err.Error())
	}
	return nil
}
