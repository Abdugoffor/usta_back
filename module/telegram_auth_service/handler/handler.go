package telegram_auth_handler

import (
	"main_service/helper"
	telegram_auth_dto "main_service/module/telegram_auth_service/dto"
	telegram_auth_service "main_service/module/telegram_auth_service/service"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
)

type Handler struct {
	svc telegram_auth_service.Service
}

func New(router *httprouter.Router, group string, svc telegram_auth_service.Service) {
	h := &Handler{svc: svc}

	router.POST(group+"/auth/telegram/start", h.Start)
	router.GET(group+"/auth/telegram/status", h.Status)
}

// Start godoc
// @Summary      Telegram login: bir martalik token olish
// @Tags         Auth
// @Produce      json
// @Success      200 {object} telegram_auth_dto.StartResponse
// @Router       /auth/telegram/start [post]
func (h *Handler) Start(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	token, deep, err := h.svc.StartLogin(r.Context())
	if err != nil {
		helper.WriteInternalError(w, err)
		return
	}
	helper.WriteJSON(w, http.StatusOK, telegram_auth_dto.StartResponse{
		Token: token, DeepLink: deep,
	})
}

// Status godoc
// @Summary      Telegram login: status tekshirish (frontdan polling)
// @Tags         Auth
// @Produce      json
// @Param        token query string true "Start endpointidan olingan token"
// @Success      200 {object} telegram_auth_dto.StatusResponse
// @Failure      400 {object} map[string]string
// @Router       /auth/telegram/status [get]
func (h *Handler) Status(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	token := strings.TrimSpace(r.URL.Query().Get("token"))
	if token == "" {
		helper.WriteError(w, http.StatusBadRequest, "token required")
		return
	}
	status, jwt, user, err := h.svc.GetStatus(r.Context(), token)
	if err != nil {
		helper.WriteInternalError(w, err)
		return
	}
	helper.WriteJSON(w, http.StatusOK, telegram_auth_dto.StatusResponse{
		Status: status, JWT: jwt, User: user,
	})
}
