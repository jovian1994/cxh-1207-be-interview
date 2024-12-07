package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jovian1994/cxh-1207-be-interview/apps/translation/request_mapping"
	"github.com/jovian1994/cxh-1207-be-interview/apps/translation/service"
	"github.com/jovian1994/cxh-1207-be-interview/pkg/unify_response"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type ITaskApi interface {
	CreateTask(c *gin.Context) error
	ExecTask(c *gin.Context) error
	GetTaskDetail(c *gin.Context) error
	DownloadTask(c *gin.Context) error
}

func NewTaskApi(taskService service.ITaskService) ITaskApi {
	return &taskApi{taskService: taskService}
}

type taskApi struct {
	taskService            service.ITaskService
	receiverTaskChangeChan chan any
	clientsMap             *sync.Map
}

func (t *taskApi) CreateTask(c *gin.Context) error {
	var req = &request_mapping.CreateTaskReq{}
	err := req.Validate(c)
	if err != nil {
		return err
	}
	err = t.taskService.CreateTask(c.GetString("username"), req)
	if err != nil {
		return err
	}
	return unify_response.NewOk()
}

func (t *taskApi) ExecTask(c *gin.Context) error {
	username := c.GetString("username")
	req := &request_mapping.ExecuteTaskReq{}
	err := req.Validate(c)
	if err != nil {
		return err
	}
	err = t.taskService.ExecuteTask(username, req.TaskId)
	if err != nil {
		return err
	}
	return unify_response.NewOk()
}
func (t *taskApi) GetTaskDetail(c *gin.Context) error {
	id := c.Query("id")
	username := c.GetString("username")
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return unify_response.ParameterError("任务ID不能为空")
	}
	if id == "" {
		return unify_response.ParameterError("任务ID不能为空")
	}
	detail, err := t.taskService.GetTaskDetail(username, idInt)
	if err != nil {
		return err
	}
	return unify_response.GetObjectSuccess(detail)
}

func (t *taskApi) DownloadTask(c *gin.Context) error {
	username := c.GetString("username")
	id := c.Query("id")
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return unify_response.ParameterError("任务ID不能为空")
	}
	done, resultPath, err := t.taskService.GetTaskIsDoneAndFilePath(idInt, username)
	if err != nil {
		return err
	}
	if done {
		// 设置响应头，提示下载文件
		c.Header("Content-Description", "File Transfer")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", resultPath))
		c.Header("Content-Type", "application/octet-stream")
		c.Data(http.StatusOK, "application/octet-stream", []byte(resultPath))
		return nil
	} else {
		return unify_response.ParameterError("任务未完成")
	}
}

func (t *taskApi) WatchTaskStatus(c *gin.Context) error {

	// 将 HTTP 升级为 WebSocket

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return nil
	}
	for {
		select {
		case <-t.receiverTaskChangeChan:
			return unify_response.NewOk()
		case <-time.After(10 * time.Second):
			return unify_response.NewOk()
		}
	}
}
