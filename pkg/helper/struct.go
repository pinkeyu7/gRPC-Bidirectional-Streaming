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

func SetFieldValue[T any](obj *T, fieldName string, value any) error {
	val := reflect.ValueOf(obj).Elem()

	// Find the field by name
	field := val.FieldByName(fieldName)

	// Check if the field is valid and can be set
	if !field.IsValid() {
		return fmt.Errorf("no such field: %s in obj", fieldName)
	}
	if !field.CanSet() {
		return fmt.Errorf("cannot set field: %s", fieldName)
	}

	// Set the field value
	fieldValue := reflect.ValueOf(value)
	if field.Type() != fieldValue.Type() {
		return fmt.Errorf("provided value type didn't match object field type")
	}

	field.Set(fieldValue)
	return nil
}
