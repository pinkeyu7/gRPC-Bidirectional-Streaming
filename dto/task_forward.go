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

type UpnpSearchRequest struct {
	WorkerId string `json:"worker_id"`
}

type UpnpSearchResponse struct {
	WorkerId string `json:"worker_id"`
	Model    string `json:"model"`
	Ip       string `json:"ip"`
}
