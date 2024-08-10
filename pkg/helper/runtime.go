package helper

import (
	"runtime"
	"strings"
)

func GetCurrentFunctionName() string {
	pc, _, _, ok := runtime.Caller(1)
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
