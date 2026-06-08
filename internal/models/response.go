package models

type SuccessResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Success bool   `json:"success"`
	Data    any    `json:"data"`
}

type ErrorResponse struct {
	Status      int    `json:"status"`
	Message     string `json:"message"`
	ErrorDetail string `json:"errorDetail"`
	Data        any    `json:"data"`
}
