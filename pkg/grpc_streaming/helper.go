package grpc_streaming

import (
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

func getParentFunctionName() string {
	pc, _, _, ok := runtime.Caller(2)
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
