package api

import "github.com/gin-gonic/gin"

type ITaskApi interface {
	CreateTask(c *gin.Context) error
	ExecTask(c *gin.Context) error
	GetTaskDetail(c *gin.Context) error
	DownloadTask(c *gin.Context) error
}

func NewTaskApi() ITaskApi {
	return &taskApi{}
}

type taskApi struct{}

func (t *taskApi) CreateTask(c *gin.Context) error {
	return nil
}

func (t *taskApi) ExecTask(c *gin.Context) error {
	return nil
}
func (t *taskApi) GetTaskDetail(c *gin.Context) error {
	return nil
}
func (t *taskApi) DownloadTask(c *gin.Context) error {
	return nil
}
