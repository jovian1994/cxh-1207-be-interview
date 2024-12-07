package request_mapping

import (
	"github.com/gin-gonic/gin"
	"github.com/jovian1994/cxh-1207-be-interview/pkg/logger"
	"github.com/jovian1994/cxh-1207-be-interview/pkg/unify_response"
	"go.uber.org/zap"
)

type UserRegisterReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (u *UserRegisterReq) Validate(c *gin.Context) error {
	err := c.ShouldBindJSON(u)
	if err != nil {
		logger.Error("注册参数解析失败", zap.Any("err", err))
		return unify_response.ParameterError("错误的请求参数")
	}
	if u.Username == "" || u.Password == "" {
		return unify_response.ParameterError("用户名或密码不能为空")
	}
	return nil
}

type UserLoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (u *UserLoginReq) Validate(c *gin.Context) error {
	err := c.ShouldBindJSON(u)
	if err != nil {
		logger.Error("注册参数解析失败", zap.Any("err", err))
		return unify_response.ParameterError("错误的请求参数")
	}
	if u.Username == "" || u.Password == "" {
		return unify_response.ParameterError("用户名或密码不能为空")
	}
	return nil
}
