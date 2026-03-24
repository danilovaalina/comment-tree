package api

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type CustomValidator struct {
	validator *validator.Validate
}

func NewCustomValidator() *CustomValidator {
	v := validator.New()

	// Настраиваем валидатор использовать имена из тегов
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		// 1. Пробуем достать имя из json (для Body)
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name != "" && name != "-" {
			return name
		}

		// 2. Если пусто, пробуем достать из param (для URL параметров)
		name = strings.SplitN(fld.Tag.Get("param"), ",", 2)[0]
		if name != "" && name != "-" {
			return name
		}

		// 3. Если и там пусто, возвращаем само имя поля (или пустую строку)
		return ""
	})

	return &CustomValidator{validator: v}
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func (a *API) validationError(err error) interface{} {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		var errMses []string
		for _, fe := range ve {
			var msg string
			switch fe.Tag() {
			case "required":
				msg = "field is required"
			case "gt":
				msg = fmt.Sprintf("must be greater than %s", fe.Param())
			case "min":
				msg = fmt.Sprintf("minimum value or length is %s", fe.Param())
			default:
				msg = fmt.Sprintf("validation failed on '%s' tag", fe.Tag())
			}
			errMses = append(errMses, fmt.Sprintf("%s: %s", fe.Field(), msg))
		}
		return echo.Map{"error": strings.Join(errMses, ", ")}
	}

	return echo.Map{"error": err.Error()}
}
