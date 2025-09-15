package client

type ErrorResponse struct {
	ErrorCode    string `json:"errorCode"`
	ErrorID      string `json:"errorId"`
	ErrorMessage string `json:"errorMessage"`
}
