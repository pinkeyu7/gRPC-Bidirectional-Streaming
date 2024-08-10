package helper

import (
	"fmt"
	"reflect"
)

func GetFieldValue[T any](value T, fieldName string) (string, error) {
	val := reflect.ValueOf(value)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return "", fmt.Errorf("value must be a struct")
	}

	field := val.FieldByName(fieldName)
	if !field.IsValid() {
		return "", fmt.Errorf("field %s not found", fieldName)
	}

	return fmt.Sprintf("%v", field.Interface()), nil
}
