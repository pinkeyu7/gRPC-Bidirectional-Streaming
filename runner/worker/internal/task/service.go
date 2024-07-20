package task

type Service interface {
	GetIds() []string
	GetInfo(taskId string) (string, error)
}
