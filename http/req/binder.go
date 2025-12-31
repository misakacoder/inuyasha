package req

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/misakacoder/inuyasha/http/resp"
	"github.com/misakacoder/kagome/str"
	"reflect"
)

func BindHeader[T any](ctx *gin.Context, object T) T {
	return BindAny(ctx.ShouldBindHeader, object)
}

func BindUri[T any](ctx *gin.Context, object T) T {
	return BindAny(ctx.ShouldBindUri, object)
}

func BindQuery[T any](ctx *gin.Context, object T) T {
	return BindAny(ctx.ShouldBindQuery, object)
}

func BindForm[T any](ctx *gin.Context, object T) T {
	return BindAny(shouldBindForm(ctx), object)
}

func BindJSON[T any](ctx *gin.Context, object T) T {
	return BindAny(ctx.ShouldBindJSON, object)
}

func Bind[T any](ctx *gin.Context, object T) T {
	return BindAny(ctx.ShouldBind, object)
}

func BindAny[T any](bind func(v any) error, object T) T {
	if err := bind(object); err != nil {
		validateError(object, err)
	}
	return object
}

func shouldBindForm(ctx *gin.Context) func(v any) error {
	return func(v any) error {
		return ctx.ShouldBindWith(v, binding.Form)
	}
}

func validateError(object any, err error) {
	var validationErrors validator.ValidationErrors
	var sliceValidationError binding.SliceValidationError
	if errors.As(err, &sliceValidationError) {
		for _, e := range sliceValidationError {
			validationErrors = append(validationErrors, e.(validator.ValidationErrors)...)
		}
	} else {
		errors.As(err, &validationErrors)
	}
	if len(validationErrors) > 0 {
		joiner := str.NewJoiner(", ", "", "")
		objectType := reflect.TypeOf(object).Elem()
		if objectType.Kind() == reflect.Slice {
			objectType = objectType.Elem().Elem()
		}
		for _, fieldError := range validationErrors {
			tag := fieldError.Tag()
			fieldName := getFieldName(fieldError.Field(), objectType)
			if tag == "required" {
				joiner.Append(fmt.Sprintf("%s不能为空", fieldName))
			} else if tag == "len" {
				param := fieldError.Param()
				joiner.Append(fmt.Sprintf("%s长度必须为%s", fieldName, param))
			} else {
				joiner.Append(fieldError.Error())
			}
		}
		if joiner.Size() > 0 {
			panic(resp.ParameterError.Msg(joiner.String()))
		}
	}
	panic(err)
}
