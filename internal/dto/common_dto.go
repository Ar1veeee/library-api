package dto

// ErrorResponse represents custom error response format
// MENGAPA struct terpisah untuk error response?
// - Konsistensi format error di seluruh API
// - Memudahkan client untuk parsing error
// - TraceID untuk debugging dan correlation logs
type ErrorResponse struct {
	Message      string `json:"message"`
	ZiyadErrCode string `json:"ziyad_error_code"`
	TraceID      string `json:"trace_id"`
}

// SuccessResponse represents generic success response
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
