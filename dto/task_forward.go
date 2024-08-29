package dto

type FooRequest struct {
	WorkerId string `json:"worker_id"`
	TaskId   string `json:"task_id"`
}

type FooResponse struct {
	WorkerId    string `json:"worker_id"`
	TaskId      string `json:"task_id"`
	TaskMessage string `json:"task_message"`
}
