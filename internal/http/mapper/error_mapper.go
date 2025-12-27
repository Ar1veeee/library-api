package mapper

import (
	"errors"
	"net/http"

	errorStruct "github.com/Ar1veeee/library-api/internal/errors"
)

func httpStatusFromAPIError(err errorStruct.APIError) int {
	switch err.ZiyadErrCode {
	case errorStruct.ErrCodeNotFound:
		return http.StatusNotFound

	case errorStruct.ErrCodeInvalidInput:
		return http.StatusBadRequest

	case errorStruct.ErrCodeAlreadyBorrowed,
		errorStruct.ErrCodeAlreadyReturned,
		errorStruct.ErrCodeQuotaExceeded,
		errorStruct.ErrCodeStockEmpty:
		return http.StatusConflict

	case errorStruct.ErrCodeTxFailed:
		return http.StatusInternalServerError

	default:
		return http.StatusBadRequest
	}
}

// MENGAPA menggunakan errors.As?
// - Kita perlu tahu apakah error dari business logic (APIError)
// - APIError = user-facing error dengan custom code
func HandleHTTPError(w http.ResponseWriter, err error) {
	var apiErr errorStruct.APIError
	if errors.As(err, &apiErr) {
		respondError(w, apiErr, httpStatusFromAPIError(apiErr))
		return
	}

	respondError(
		w,
		errorStruct.NewAPIError("Internal server error", errorStruct.ErrCodeTxFailed),
		http.StatusInternalServerError,
	)
}
