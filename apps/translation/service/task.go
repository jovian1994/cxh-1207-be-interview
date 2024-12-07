package service

import (
	"crypto/rand"
	"fmt"
	"github.com/jovian1994/cxh-1207-be-interview/apps/translation/config"
	"github.com/jovian1994/cxh-1207-be-interview/apps/translation/dao"
	"github.com/jovian1994/cxh-1207-be-interview/apps/translation/request_mapping"
	"github.com/jovian1994/cxh-1207-be-interview/models"
	"github.com/jovian1994/cxh-1207-be-interview/pkg/llm"
	"github.com/jovian1994/cxh-1207-be-interview/pkg/logger"
	"io/ioutil"
	"os"
	"path"
)

type ITaskService interface {
	GetTaskDetail(username string, taskId int64) (*TaskData, error)
	ExecuteTask(username string, taskId int64) error
	CreateTask(username string, req *request_mapping.CreateTaskReq) error
	GetTaskIsDoneAndFilePath(id int64, username string) (bool, string, error)
}

type taskService struct {
	taskDao       dao.ITaskDao
	llm           llm.ILLMClient
	notifyChannel chan map[string]any
}

func NewTaskService(
	taskDao dao.ITaskDao, client llm.ILLMClient, notifyChannel chan map[string]any) ITaskService {
	return &taskService{
		taskDao:       taskDao,
		llm:           client,
		notifyChannel: notifyChannel,
	}
}

func (t *taskService) GetTaskDetail(
	username string, taskId int64) (*TaskData, error) {

	data, err := t.taskDao.GetTaskByIdAndUsername(username, taskId)
	if err != nil {
		return nil, err
	}
	item := &TaskData{
		Id:         int(data.ID),
		Status:     data.Status,
		CreateBy:   data.CreateBy,
		Content:    data.Content,
		Lang:       data.Lang,
		TargetLang: data.TargetLang,
	}

	if data.Status == 2 && data.IsOss != 1 {
		filename := data.ResultKey
		_, err := os.Stat(path.Join(filename))
		if err == nil {
			// 读取文件内容
			fileData, err := ioutil.ReadFile(filename)
			if err == nil {
				data.Content = string(fileData)
			}
		}
	}
	return item, nil
}
func (t *taskService) ExecuteTask(username string, taskId int64) error {
	taskData, err := t.taskDao.GetTaskByIdAndUsername(username, taskId)
	if err != nil {
		return err
	}
	err = t.taskDao.UpdateTaskStatus(taskId, map[string]any{"status": 2})
	if err != nil {
		return err
	}
	err = t.execute(taskData)
	if err != nil {
		return err
	}
	return nil
}

func (t *taskService) CreateTask(username string, req *request_mapping.CreateTaskReq) error {
	err := t.taskDao.CreateTask(username, req.Content, req.TargetLang, req.Lang)
	if err != nil {
		return err
	}
	return nil
}

func (t *taskService) execute(taskData *models.TaskModel) error {
	go func(taskData *models.TaskModel, client llm.ILLMClient) {
		panicErr := recover()
		if panicErr != nil {
			fmt.Println(panicErr)
		}
		translate, err := client.Translate(
			taskData.Content, taskData.Lang, taskData.TargetLang)
		if err != nil {
			logger.Error("send message to llm error")
		}
		filename, err := t.generateRandomFilename()
		if err != nil {
			logger.Error("生成文件名失败")
			return
		}
		// 构造完整路径
		filePath := fmt.Sprintf("%s.txt", path.Join(config.GetConfig().TaskResultDir, filename))
		// 将内容写入文件
		err = ioutil.WriteFile(filePath, []byte(translate), 0644)
		if err != nil {
			t.taskDao.UpdateTaskStatus(
				int64(taskData.ID), map[string]any{
					"status":   3,
					"task_key": filePath,
				})
			logger.Error(fmt.Sprintf("failed to write to file: %s", err.Error()))
			return
		}
		err = t.taskDao.UpdateTaskStatus(
			int64(taskData.ID), map[string]any{
				"status":   2,
				"task_key": filePath,
			})
		if err != nil {
			logger.Error(fmt.Sprintf("failed to update task status: %s", err.Error()))
			return
		}
		t.notifyChannel <- map[string]any{
			"task_id":   int64(taskData.ID),
			"file_path": filePath,
			"status":    2,
		}
	}(taskData, t.llm)
	return nil
}

// generateRandomFilename 生成一个随机的文件名
func (t *taskService) generateRandomFilename() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return fmt.Sprintf("%x", b), nil
}

func (t *taskService) GetTaskIsDoneAndFilePath(id int64, username string) (bool, string, error) {
	taskData, err := t.taskDao.GetTaskByIdAndUsername(username, id)
	if err != nil {
		return false, "", err
	}
	if taskData.Status != 2 {
		return false, "", nil
	}
	return true, taskData.ResultKey, nil
}
