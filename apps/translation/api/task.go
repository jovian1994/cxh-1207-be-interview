package api

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jovian1994/cxh-1207-be-interview/apps/translation/request_mapping"
	"github.com/jovian1994/cxh-1207-be-interview/apps/translation/service"
	"github.com/jovian1994/cxh-1207-be-interview/pkg/strutil"
	"github.com/jovian1994/cxh-1207-be-interview/pkg/unify_response"
	"net/http"
	"strconv"
	"sync"
)

type ITaskApi interface {
	CreateTask(c *gin.Context) error
	ExecTask(c *gin.Context) error
	GetTaskDetail(c *gin.Context) error
	DownloadTask(c *gin.Context) error
	WatchTaskStatus(c *gin.Context) error
}

func NewTaskApi(taskService service.ITaskService, receiverChannel chan map[string]any) ITaskApi {
	t := &taskApi{
		taskService:            taskService,
		connections:            make(map[string]*webSocketClient),
		receiverTaskChangeChan: receiverChannel,
	}
	t.pushMessage()
}

type taskApi struct {
	taskService            service.ITaskService
	receiverTaskChangeChan chan map[string]any
	connections            map[string]*webSocketClient
}

type webSocketClient struct {
	Conn     *websocket.Conn
	Username string
	ClientId string
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

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许跨域
	},
}

// 定义一个用于存储所有连接的map
var connections = make(map[*websocket.Conn]bool)

// 定义一个用于删除连接的函数
var locker = &sync.Mutex{}

func (t *taskApi) removeConnection(id string) {
	locker.Lock()
	defer locker.Unlock()
	delete(t.connections, id)
}

func (t *taskApi) addConnection(id string, conn *webSocketClient) {
	locker.Lock()
	defer locker.Unlock()
	t.connections[id] = conn
}

// 定义一个用于广播消息的函数
func (t *taskApi) broadcastMessage(message map[string]any) {
	for _, wsConn := range t.connections {
		username, ok := message["username"].(string)
		if ok && username == wsConn.Username {
			dataBytes, err := json.Marshal(message)
			if err == nil {
				wsConn.Conn.WriteMessage(websocket.TextMessage, dataBytes)
			}
		}
	}
}

func (t *taskApi) WatchTaskStatus(c *gin.Context) error {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return nil
	}
	defer conn.Close()
	clientId := strutil.NewUUID()
	username := c.GetString("username")
	wsClient := &webSocketClient{
		Conn:     conn,
		Username: username,
		ClientId: clientId,
	}
	t.addConnection(clientId, wsClient)
	ch := make(chan struct{}, 1)
	go func() {
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				fmt.Println("read:", err)
				t.removeConnection(wsClient.ClientId)
				ch <- struct{}{}
				break
			}

		}
	}()
	return nil
}

func (t *taskApi) pushMessage() error {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println(err)
			}
		}()
		for {
			select {
			case message := <-t.receiverTaskChangeChan:
				t.broadcastMessage(message)
			}
		}
	}()
	return nil
}
