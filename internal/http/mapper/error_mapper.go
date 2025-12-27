package mapper

import (
	"errors"
	"net/http"

	"github.com/Ar1veeee/library-api/internal/model"
)

func httpStatusFromAPIError(err model.APIError) int {
	switch err.ZiyadErrCode {
	case model.ErrCodeNotFound:
		return http.StatusNotFound

	case model.ErrCodeInvalidInput:
		return http.StatusBadRequest

	case model.ErrCodeAlreadyBorrowed,
		model.ErrCodeAlreadyReturned,
		model.ErrCodeQuotaExceeded,
		model.ErrCodeStockEmpty:
		return http.StatusConflict

	case model.ErrCodeTxFailed:
		return http.StatusInternalServerError

	default:
		return http.StatusBadRequest
	}
}

// MENGAPA menggunakan errors.As?
// - Kita perlu tahu apakah error dari business logic (APIError)
// - APIError = user-facing error dengan custom code
func HandleHTTPError(w http.ResponseWriter, err error) {
	var apiErr model.APIError
	if errors.As(err, &apiErr) {
		respondError(w, apiErr, httpStatusFromAPIError(apiErr))
		return
	}

	respondError(
		w,
		model.NewAPIError("Internal server error", model.ErrCodeTxFailed),
		http.StatusInternalServerError,
	)
}
