package api

import (
	"github.com/gin-gonic/gin"
	"github.com/jovian1994/cxh-1207-be-interview/apps/translation/request_mapping"
	"github.com/jovian1994/cxh-1207-be-interview/apps/translation/service"
	"github.com/jovian1994/cxh-1207-be-interview/pkg/unify_response"
)

type IUserApi interface {
	Login(ctx *gin.Context) error
	Register(ctx *gin.Context) error
}

func NewUserApi(userService service.IUserService) IUserApi {
	return &user{userService: userService}
}

type user struct {
	userService service.IUserService
}

func (u *user) Login(ctx *gin.Context) error {
	var req = &request_mapping.UserLoginReq{}
	if err := req.Validate(ctx); err != nil {
		return err
	}
	token, err := u.userService.HandlerLogin(req.Username, req.Password)
	if err != nil {
		return err
	}
	return unify_response.GetObjectSuccess(map[string]interface{}{
		"token": token,
	})
}
func (u *user) Register(ctx *gin.Context) error {
	var req = &request_mapping.UserRegisterReq{}
	if err := req.Validate(ctx); err != nil {
		return err
	}
	err := u.userService.HandlerRegister(req.Username, req.Password)
	if err != nil {
		return err
	}
	return unify_response.NewOk()
}
