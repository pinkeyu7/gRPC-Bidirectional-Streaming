package dto

type UnaryRequest struct {
	WorkerId string `json:"worker_id"`
	TaskId   string `json:"task_id"`
}

type UnaryResponse struct {
	WorkerId    string `json:"worker_id"`
	TaskId      string `json:"task_id"`
	TaskMessage string `json:"task_message"`
}

type ClientStreamRequest struct {
	WorkerId string `json:"worker_id"`
}

type ClientStreamResponse struct {
	WorkerId string `json:"worker_id"`
	Model    string `json:"model"`
	Ip       string `json:"ip"`
}
