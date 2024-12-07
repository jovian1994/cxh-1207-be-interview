package unify_response

import (
	"errors"
	"github.com/go-playground/validator/v10"
	"reflect"
)

const (
	errorFieldKey   = "field"
	errFieldMessage = "message"
	errMsgTag       = "errMsg"
	errFieldTag     = "errField"
)

func GetValidatorError(requestMapper any, err error) []map[string]string {
	var errorInfo = make([]map[string]string, 0)
	if requestMapper == nil || err == nil {
		return errorInfo
	}
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		for _, e := range validationErrors {
			fieldName := e.Field()
			typeof := reflect.TypeOf(requestMapper)
			for typeof.Kind() == reflect.Ptr {
				typeof = typeof.Elem()
			}
			field, ok := typeof.FieldByName(fieldName)
			if ok {
				errorField := field.Tag.Get(errFieldTag)
				errMsg := field.Tag.Get(errMsgTag)
				if errorField != "" && errMsg != "" {
					errorInfo = append(errorInfo, map[string]string{
						errorFieldKey:   errorField,
						errFieldMessage: errMsg,
					})
				}
			}
		}
	}
	return errorInfo
}
