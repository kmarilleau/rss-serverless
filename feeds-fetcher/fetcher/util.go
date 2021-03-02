package fetcher

import (
	"os"
	"reflect"
	"strings"
)

func stringIsEmpty(str string) bool {
	return strings.TrimSpace(str) == ""
}

func getFieldTag(s interface{}, fieldName string, tagType string) string {
	structField, _ := reflect.TypeOf(s).FieldByName(fieldName)

	return string(structField.Tag.Get(tagType))
}

func getEnv() string {
	switch os.Getenv("GOOGLE_CLOUD_PROJECT") {
	case "PROD":
		return "PROD"
	default:
		return "TEST"
	}
}
