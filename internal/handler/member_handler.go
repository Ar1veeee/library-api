package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Ar1veeee/library-api/internal/model"
	"github.com/Ar1veeee/library-api/internal/service"
	"github.com/gorilla/mux"
)

type MemberHandler struct {
	memberService *service.MemberService
}

func NewMemberHandler(memberService *service.MemberService) *MemberHandler {
	return &MemberHandler{memberService: memberService}
}

func (h *MemberHandler) GetMemberLoans(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	memberID, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, model.NewAPIError("Invalid member ID", model.ErrCodeInvalidInput), http.StatusBadRequest)
		return
	}

	loans, err := h.memberService.GetMemberLoans(r.Context(), memberID)
	if err != nil {
		// MENGAPA menggunakan errors.As?
		// - Kita perlu tahu apakah error dari business logic (APIError)
		// - APIError = user-facing error dengan custom code
		var apiErr model.APIError
		if errors.As(err, &apiErr) {
			respondError(w, apiErr, http.StatusNotFound)
		}
		return
	}

	respondSuccess(w, loans, http.StatusOK)
}
