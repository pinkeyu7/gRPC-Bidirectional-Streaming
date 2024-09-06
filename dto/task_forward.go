package dto

type UnaryRequest struct {
	WorkerID string `json:"worker_id"`
	TaskID   string `json:"task_id"`
}

type UnaryResponse struct {
	WorkerID    string `json:"worker_id"`
	TaskID      string `json:"task_id"`
	TaskMessage string `json:"task_message"`
}

type ClientStreamRequest struct {
	WorkerID string `json:"worker_id"`
}

type ClientStreamResponse struct {
	WorkerID string `json:"worker_id"`
	Model    string `json:"model"`
	IP       string `json:"ip"`
}
