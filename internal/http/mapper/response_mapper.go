package mapper

import (
	"encoding/json"
	"net/http"

	"github.com/Ar1veeee/library-api/internal/dto"
	"github.com/Ar1veeee/library-api/internal/model"
)

func respondError(w http.ResponseWriter, err model.APIError, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	_ = json.NewEncoder(w).Encode(dto.ErrorResponse{
		Message:      err.Message,
		ZiyadErrCode: err.ZiyadErrCode,
		TraceID:      err.TraceID,
	})
}

func RespondSuccess(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	_ = json.NewEncoder(w).Encode(data)
}
