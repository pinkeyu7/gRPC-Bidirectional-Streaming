package grpc_streaming

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"reflect"
	"runtime"
	"strings"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.IntN(len(letterBytes))]
	}
	return string(b)
}

func getPackageNameFromStruct(s interface{}) string {
	t := reflect.TypeOf(s)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	pkgPaths := strings.Split(t.PkgPath(), "/")
	return pkgPaths[len(pkgPaths)-1]
}

func getParentFunctionName(skip int) string {
	pc, _, _, ok := runtime.Caller(skip)
	if !ok {
		return "unknown"
	}

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "unknown"
	}

	// Extract the function name and strip the package path
	name := fn.Name()
	parts := strings.Split(name, "/")

	names := strings.Split(parts[len(parts)-1], ".")

	return names[len(names)-1]
}

func getError[T any](value T) (*Error, error) {
	val := reflect.ValueOf(value)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("value must be a struct")
	}

	errorField := val.FieldByName("Error")
	if !errorField.IsValid() {
		return nil, fmt.Errorf("can not retrieve error field value")
	}

	errorString, err := json.Marshal(errorField.Interface())
	if err != nil {
		return nil, err
	}

	// Using json string to check error is nil
	if string(errorString) == "null" {
		return nil, nil
	}

	// Convert to error
	errorFromValue := &Error{}
	err = json.Unmarshal(errorString, errorFromValue)
	if err != nil {
		return nil, err
	}

	return errorFromValue, nil
}

func getFieldValue[T any](value T, fieldName string) (string, error) {
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

func setFieldValue[T any](obj *T, fieldName string, value any) error {
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

func convert[Source any, Target any](s *Source, t *Target) error {
	// JSON marshal
	byteString, err := json.Marshal(s)
	if err != nil {
		return err
	}

	// JSON unmarshal
	err = json.Unmarshal(byteString, t)
	if err != nil {
		return err
	}

	return nil
}
