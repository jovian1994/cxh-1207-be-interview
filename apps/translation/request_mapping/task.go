package request_mapping

import (
	"github.com/gin-gonic/gin"
	"github.com/jovian1994/cxh-1207-be-interview/pkg/unify_response"
)

type CreateTaskReq struct {
	Content    string `json:"content"`
	Lang       string `json:"lang"`
	TargetLang string `json:"target_lang"`
}

func (req *CreateTaskReq) Validate(c *gin.Context) error {
	err := c.ShouldBindJSON(req)
	if err != nil {
		return unify_response.ParameterError("参数错误")
	}
	if req.Content == "" {
		return unify_response.ParameterError("内容不可以为空")
	}
	if req.Lang == "" {
		req.Lang = "auto-detect"
	}
	if req.TargetLang == "" {
		return unify_response.ParameterError("目标语言不可以为空")
	}
	return nil
}

type ExecuteTaskReq struct {
	TaskId int64 `json:"task_id"`
}

func (req *ExecuteTaskReq) Validate(c *gin.Context) error {
	err := c.ShouldBindJSON(req)
	if err != nil {
		return unify_response.ParameterError("参数错误")
	}
	if req.TaskId == 0 {
		return unify_response.ParameterError("任务ID不可以为空")
	}
	return nil
}
