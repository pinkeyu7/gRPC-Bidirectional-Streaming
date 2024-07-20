package task

type Service interface {
	GetInfo(taskId string) (string, error)
}
