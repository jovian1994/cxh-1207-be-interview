package unify_response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIError struct {
	Code        int                 `json:"-"`
	ErrorCode   int                 `json:"code"`
	Message     string              `json:"message"`
	RequestPath string              `json:"requestPath"`
	Data        any                 `json:"data,omitempty"`
	Count       int64               `json:"count,omitempty"`
	ErrorFields []map[string]string `json:"errorFields,omitempty"`
}

type ListResponse struct {
	Code        int    `json:"-"`
	ErrorCode   int    `json:"code"`
	Message     string `json:"message"`
	RequestPath string `json:"requestPath"`
	Data        any    `json:"data,omitempty"`
	Count       int64  `json:"count"`
}

func (a *APIError) Error() string {
	return a.Message
}

func newApiError(code int, errorCode int, msg string) *APIError {
	return &APIError{
		Code:      code,
		ErrorCode: errorCode,
		Message:   msg,
	}
}

func ServerError(message string) *APIError {
	return newApiError(http.StatusInternalServerError, serverErrorCode, message)
}

func ParameterError(message string, errorFields ...map[string]string) *APIError {
	e := newApiError(http.StatusBadRequest, parameterErrorCode, message)
	if len(errorFields) != 0 {
		e.ErrorFields = errorFields
	} else {
		e.ErrorFields = make([]map[string]string, 0)
	}
	return e
}

func DBError(message string) *APIError {
	if message == "" {
		message = "数据库异常"
	}
	return newApiError(http.StatusInternalServerError, dbErrorCode, message)
}

func NotFound() *APIError {
	return newApiError(http.StatusNotFound, notFoundCode, http.StatusText(http.StatusNotFound))
}

type handleFunc func(ctx *gin.Context) error

func UnifyResponseWrapper(handler handleFunc) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		var err error
		err = handler(ctx)
		if err != nil {
			var apiError *APIError
			var h *APIError
			if errors.As(err, &h) {
				apiError = h
			}
			if apiError.ErrorCode == getListSuccessCode {
				data := ListResponse{
					Code:        apiError.Code,
					ErrorCode:   apiError.ErrorCode,
					Message:     apiError.Message,
					RequestPath: apiError.RequestPath,
					Data:        apiError.Data,
					Count:       apiError.Count,
				}
				ctx.JSON(apiError.Code, data)
				return
			}
			ctx.JSON(apiError.Code, apiError)
			return
		}
	}
}

func NewOk() *APIError {
	return newApiError(http.StatusOK, okCode, "ok")
}

func GetListSuccess(data any, count int64, msg string) *APIError {
	if msg == "" {
		msg = "success"
	}
	return &APIError{
		Code:      http.StatusOK,
		ErrorCode: getListSuccessCode,
		Message:   msg,
		Data:      data,
		Count:     count,
	}
}

func GetObjectSuccess(data any) *APIError {
	return &APIError{
		Code:      http.StatusOK,
		ErrorCode: getObjectSuccessCode,
		Message:   "success",
		Data:      data,
	}
}

func NewForbidden(msg string) *APIError {
	if msg == "" {
		msg = "forbidden"
	}
	return &APIError{
		Code:      http.StatusForbidden,
		ErrorCode: forbiddenCode,
		Message:   msg,
	}
}

// NewUnAuthorize 接口未授权
func NewUnAuthorize(msg string) *APIError {
	if msg == "" {
		msg = "unauthorized"
	}
	return &APIError{
		Code:      http.StatusUnauthorized,
		ErrorCode: UnAuthorizeCode,
		Message:   msg,
	}
}

// 用户已存在
func UserAlreadyExist(msg string) *APIError {
	if msg == "" {
		msg = "user already exist"
	}
	return &APIError{
		Code:      http.StatusBadRequest,
		ErrorCode: userAlreadyExistCode,
		Message:   msg,
	}
}
