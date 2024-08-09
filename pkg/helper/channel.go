package helper

func IsClosed[T any](ch <-chan T) bool {
	select {
	case <-ch:
		return true
	default:
	}

	return false
}
