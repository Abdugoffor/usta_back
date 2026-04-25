package helper

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var _v = validator.New()

func init() {
	_v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" || name == "" {
			return fld.Name
		}
		return name
	})
}

func ValidateStruct(s interface{}) map[string]string {
	err := _v.Struct(s)
	if err == nil {
		return nil
	}

	errs := make(map[string]string)
	for _, e := range err.(validator.ValidationErrors) {
		errs[e.Field()] = validationMsg(e)
	}
	return errs
}

func validationMsg(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "Bu maydon to'ldirilishi shart"
	case "min":
		if e.Kind() == reflect.String {
			return fmt.Sprintf("Kamida %s ta belgi kiriting", e.Param())
		}
		return fmt.Sprintf("Minimal qiymat: %s", e.Param())
	case "max":
		if e.Kind() == reflect.String {
			return fmt.Sprintf("Ko'pi bilan %s ta belgi kiriting", e.Param())
		}
		return fmt.Sprintf("Maksimal qiymat: %s", e.Param())
	case "email":
		return "Noto'g'ri email format"
	case "url":
		return "Noto'g'ri URL format"
	case "numeric":
		return "Faqat raqam kiriting"
	case "alpha":
		return "Faqat harf kiriting"
	case "oneof":
		return fmt.Sprintf("Quyidagilardan birini tanlang: %s", e.Param())
	case "len":
		return fmt.Sprintf("Uzunligi aynan %s ta belgi bo'lishi kerak", e.Param())
	case "gt":
		return fmt.Sprintf("%s dan katta bo'lishi kerak", e.Param())
	case "gte":
		return fmt.Sprintf("Kamida %s bo'lishi kerak", e.Param())
	case "lt":
		return fmt.Sprintf("%s dan kichik bo'lishi kerak", e.Param())
	case "lte":
		return fmt.Sprintf("Ko'pi bilan %s bo'lishi kerak", e.Param())
	default:
		return fmt.Sprintf("Noto'g'ri qiymat (%s)", e.Tag())
	}
}
