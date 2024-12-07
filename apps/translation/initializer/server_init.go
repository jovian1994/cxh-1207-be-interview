package initializer

import (
	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	"github.com/gin-gonic/gin"
	"github.com/jovian1994/cxh-1207-be-interview/apps/translation/api"
	"github.com/jovian1994/cxh-1207-be-interview/apps/translation/config"
	"github.com/jovian1994/cxh-1207-be-interview/apps/translation/dao"
	"github.com/jovian1994/cxh-1207-be-interview/apps/translation/service"
	"github.com/jovian1994/cxh-1207-be-interview/middlewares"
	"github.com/jovian1994/cxh-1207-be-interview/pkg/jwt"
	"github.com/jovian1994/cxh-1207-be-interview/pkg/llm"
	"github.com/jovian1994/cxh-1207-be-interview/pkg/mysql_tool"
	"github.com/jovian1994/cxh-1207-be-interview/pkg/unify_response"
	"strconv"
)

const (
	dbClientName = "translation-task"
)

func initMysql() {
	mysqlConfig := config.GetConfig().MysqlConfig
	if mysqlConfig == nil {
		panic("myslq配置为空")
	}
	err := mysql_tool.
		InitMysqlClient("translation-task",
			mysqlConfig.User,
			mysqlConfig.Password, mysqlConfig.Host, strconv.Itoa(mysqlConfig.Port))
	if err != nil {
		panic(err)
	}
}

func ServerInit(e *gin.Engine) {
	initMysql()
	e.GET("/health", func(c *gin.Context) {
		c.Status(200)
	})

	e.Use(middlewares.Cors())
	e.Use(middlewares.GinRecovery(true))
	e.Use(middlewares.LoggerRecord())
	e.NoRoute(middlewares.HandleNotFound)
	e.NoMethod(middlewares.HandleNotFound)

	tokenVerify := jwt.NewTokenVerify()
	llmClient := llm.NewLLMClient()
	userDao := dao.NewUserDao(dbClientName, tokenVerify)
	taskDao := dao.NewTaskDao(dbClientName)

	userService := service.NewUserService(userDao)
	taskService := service.NewTaskService(taskDao, llmClient)

	userApi := api.NewUserApi(userService)
	taskApi := api.NewTaskApi(taskService)

	rateLimit := getRateLimit()
	r := e.Group("/v1")
	{
		r.POST("/user/login", unify_response.UnifyResponseWrapper(userApi.Login))
		r.POST("/user/register", unify_response.UnifyResponseWrapper(userApi.Register))
		r.POST("/task/create", middlewares.RateLimitMiddleware(rateLimit), middlewares.LoginRequired(tokenVerify), unify_response.UnifyResponseWrapper(taskApi.CreateTask))
		r.POST("/task/execute", middlewares.RateLimitMiddleware(rateLimit), middlewares.LoginRequired(tokenVerify), unify_response.UnifyResponseWrapper(taskApi.ExecTask))
		r.GET("/task/detail", middlewares.RateLimitMiddleware(rateLimit), middlewares.LoginRequired(tokenVerify), unify_response.UnifyResponseWrapper(taskApi.GetTaskDetail))
		r.GET("/task/download", middlewares.RateLimitMiddleware(rateLimit), middlewares.LoginRequired(tokenVerify), unify_response.UnifyResponseWrapper(taskApi.DownloadTask))
	}
}

func getRateLimit() *limiter.Limiter {
	r := config.GetConfig().RateLimit
	if r == nil {
		return tollbooth.NewLimiter(1, nil)
	}
	return tollbooth.NewLimiter(float64(r.Limit), nil)
}
