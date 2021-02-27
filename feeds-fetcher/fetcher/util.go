package fetcher

import (
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
