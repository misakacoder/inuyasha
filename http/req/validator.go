package req

import (
	"fmt"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/misakacoder/inuyasha/http/resp"
	"github.com/misakacoder/inuyasha/pkg/db/types"
	"reflect"
	"strings"
	"time"
)

type Enum interface {
	Values() []Enum
}

func init() {
	if validate, ok := binding.Validator.Engine().(*validator.Validate); ok {
		validate.RegisterValidation("trim", trim)
		validate.RegisterValidation("enum", enum)
		validate.RegisterValidation("datepicker", datepicker(false))
		validate.RegisterValidation("sameYearDatepicker", datepicker(true))
	}
}

func trim(fieldLevel validator.FieldLevel) bool {
	if str, ok := fieldLevel.Field().Interface().(string); ok {
		fieldLevel.Field().SetString(strings.TrimSpace(str))
	}
	return true
}

func enum(fieldLevel validator.FieldLevel) bool {
	if value, ok := fieldLevel.Field().Interface().(Enum); ok {
		values := value.Values()
		for _, v := range values {
			if value == v {
				return true
			}
		}
		fieldName := getFieldName(fieldLevel.FieldName(), fieldLevel.Parent().Type())
		panic(resp.ParameterError.Msg(fmt.Sprintf("%s必须是%v中的一个值", fieldName, values)))
	}
	return true
}

func datepicker(mustSameYear bool) func(validator.FieldLevel) bool {
	return func(fieldLevel validator.FieldLevel) bool {
		field := fieldLevel.Field().Interface()
		fieldName := getFieldName(fieldLevel.FieldName(), fieldLevel.Parent().Type())
		switch dateTime := field.(type) {
		case []types.DateTime:
			validateDatepickerTime(fieldName, dateTime[0].Time(), dateTime[1].Time(), mustSameYear)
		case []*types.DateTime:
			validateDatepickerTime(fieldName, dateTime[0].Time(), dateTime[1].Time(), mustSameYear)
		default:
		}
		return true
	}
}

func validateDatepickerTime(fieldName string, begin, end time.Time, mustSameYear bool) {
	if mustSameYear && begin.Year() != end.Year() {
		panic(resp.ParameterError.Msg(fmt.Sprintf("%s的值必须是同一年的时间", fieldName)))
	}
	if begin.After(end) {
		panic(resp.ParameterError.Msg(fmt.Sprintf("%s的开始时间不能大于结束时间", fieldName)))
	}
}

func getFieldName(fieldName string, tp reflect.Type) string {
	field, ok := tp.FieldByName(fieldName)
	if ok {
		if form := field.Tag.Get("form"); form != "" {
			fieldName = form
		} else if json := field.Tag.Get("json"); json != "" {
			fieldName = json
		}
	}
	return fieldName
}
