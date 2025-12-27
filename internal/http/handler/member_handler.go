package handler

import (
	"net/http"
	"strconv"

	"github.com/Ar1veeee/library-api/internal/dto"
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
		HandleHTTPError(w, err)
		return
	}

	loans, err := h.memberService.GetMemberLoans(r.Context(), memberID)
	if err != nil {
		HandleHTTPError(w, err)
		return
	}

	response := dto.SuccessResponse{
		Message: "Berhasil mengambil riwayat peminjaman member",
		Data:    loans,
	}

	respondSuccess(w, response, http.StatusOK)
}
