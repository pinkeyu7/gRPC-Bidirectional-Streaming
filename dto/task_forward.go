package dto

type FooRequest struct {
	WorkerId string
	TaskId   string
}

type FooResponse struct {
	WorkerId    string
	TaskId      string
	TaskMessage string
}
