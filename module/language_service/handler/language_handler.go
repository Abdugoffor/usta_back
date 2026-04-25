package language_handler

import (
	"encoding/json"
	"main_service/helper"
	"main_service/middleware"
	language_dto "main_service/module/language_service/dto"
	language_service "main_service/module/language_service/service"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"
)

type languageHandler struct {
	service language_service.LanguageService
}

func NewLanguageHandler(router *httprouter.Router, group string, db *pgxpool.Pool) {
	h := &languageHandler{service: language_service.NewLanguageService(db)}

	routes := group + "/languages"
	{
		router.GET(routes, middleware.CheckRole(h.List, "admin"))

		router.GET(routes+"/:id", middleware.CheckRole(h.Show, "admin"))

		router.POST(routes, middleware.CheckRole(h.Create, "admin"))

		router.PUT(routes+"/:id", middleware.CheckRole(h.Update, "admin"))

		router.DELETE(routes+"/:id", middleware.CheckRole(h.Delete, "admin"))
	}

	router.GET(group+"/count/languages", middleware.CheckRole(h.Count, "admin"))

	router.GET(group+"/active/languages", h.ListActive)
}

// ListActive godoc
// @Summary      Faol tillar ro'yxati (public)
// @Tags         Languages
// @Produce      json
// @Success      200 {object} map[string]any
// @Router       /active/languages [get]
func (h *languageHandler) ListActive(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	items, err := h.service.ListActive(r.Context())

	if err != nil {
		helper.WriteInternalError(w, err)

		return
	}

	helper.WriteJSON(w, http.StatusOK, map[string]any{"data": items})
}

// List godoc
// @Summary      Tillar ro'yxati (admin, cursor pagination)
// @Tags         Languages
// @Produce      json
// @Security     BearerAuth
// @Param        cursor      query string false "Keyset cursor"
// @Param        limit       query int    false "Default 10, max 100"
// @Param        sort_by     query string false "id|name|is_active|created_at|updated_at"
// @Param        sort_order  query string false "asc|desc"
// @Param        name        query string false "Til kodi bo'yicha qidirish (ILIKE)"
// @Param        description query string false "Tavsif bo'yicha qidirish"
// @Param        is_active   query bool   false "Holat"
// @Success      200 {object} map[string]any
// @Router       /languages [get]
func (h *languageHandler) List(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	q := helper.ParseAdminList(r)

	f := language_dto.LanguageFilter{
		Name:        helper.QueryString(r, "name", 100),
		Description: helper.QueryString(r, "description", 200),
		IsActive:    helper.QueryBool(r, "is_active"),
	}

	items, hasMore, err := h.service.List(r.Context(), f, q)

	if err != nil {
		helper.WriteInternalError(w, err)

		return
	}

	var lastID int64

	var lastValue string

	if len(items) > 0 {
		last := items[len(items)-1]

		lastID = last.ID

		lastValue = languageSortValue(q.SortBy, last)
	}

	helper.WriteJSON(w, http.StatusOK, map[string]any{
		"data": items,
		"meta": helper.NewCursorMetaWithValue(q.Limit, hasMore, lastID, lastValue, 0),
	})
}

// Count godoc
// @Summary      Tillar soni (admin)
// @Tags         Languages
// @Produce      json
// @Security     BearerAuth
// @Param        name        query string false "Til kodi (ILIKE)"
// @Param        description query string false "Tavsif (ILIKE)"
// @Param        is_active   query bool   false "Holat"
// @Success      200 {object} map[string]int64
// @Router       /count/languages [get]
func (h *languageHandler) Count(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	f := language_dto.LanguageFilter{
		Name:        helper.QueryString(r, "name", 100),
		Description: helper.QueryString(r, "description", 200),
		IsActive:    helper.QueryBool(r, "is_active"),
	}

	total, err := h.service.Count(r.Context(), f)

	if err != nil {
		helper.WriteInternalError(w, err)

		return
	}

	helper.WriteJSON(w, http.StatusOK, map[string]int64{"total": total})
}

func languageSortValue(sortBy string, l *language_dto.LanguageResponse) string {
	switch sortBy {
	case "name":
		return l.Name
	case "is_active":
		return strconv.FormatBool(l.IsActive)
	case "created_at":
		return l.CreatedAt.Format("2006-01-02 15:04:05.000000-07:00")
	case "updated_at":
		return l.UpdatedAt.Format("2006-01-02 15:04:05.000000-07:00")
	}

	return ""
}

// Show godoc
// @Summary      Tilni ID bo'yicha olish (admin)
// @Tags         Languages
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Language ID"
// @Success      200 {object} language_dto.LanguageResponse
// @Failure      404 {object} map[string]string
// @Router       /languages/{id} [get]
func (h *languageHandler) Show(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.ParseInt(ps.ByName("id"), 10, 64)

	if err != nil || id <= 0 {
		helper.WriteError(w, http.StatusBadRequest, "invalid id")

		return
	}

	resp, err := h.service.Show(r.Context(), id)

	if err != nil {
		helper.WriteError(w, http.StatusNotFound, "language not found")

		return
	}

	helper.WriteJSON(w, http.StatusOK, resp)
}

// Create godoc
// @Summary      Yangi til qo'shish (admin)
// @Tags         Languages
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body language_dto.CreateLanguageRequest true "Til ma'lumotlari"
// @Success      201 {object} language_dto.LanguageResponse
// @Failure      422 {object} map[string]any
// @Router       /languages [post]
func (h *languageHandler) Create(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req language_dto.CreateLanguageRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.WriteError(w, http.StatusBadRequest, "invalid JSON")

		return
	}

	req.Name = strings.TrimSpace(req.Name)

	if errs := helper.ValidateStruct(req); errs != nil {
		helper.WriteValidation(w, errs)

		return
	}

	resp, err := h.service.Create(r.Context(), req)

	if err != nil {
		helper.WriteInternalError(w, err)

		return
	}

	helper.WriteJSON(w, http.StatusCreated, resp)
}

// Update godoc
// @Summary      Tilni yangilash (admin)
// @Tags         Languages
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path int                                true  "Language ID"
// @Param        body body language_dto.UpdateLanguageRequest true  "Yangilanadigan maydonlar"
// @Success      200 {object} language_dto.LanguageResponse
// @Failure      404 {object} map[string]string
// @Failure      422 {object} map[string]any
// @Router       /languages/{id} [put]
func (h *languageHandler) Update(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.ParseInt(ps.ByName("id"), 10, 64)

	if err != nil || id <= 0 {
		helper.WriteError(w, http.StatusBadRequest, "invalid id")

		return
	}

	var req language_dto.UpdateLanguageRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.WriteError(w, http.StatusBadRequest, "invalid JSON")

		return
	}

	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)

		req.Name = &trimmed
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		helper.WriteValidation(w, errs)

		return
	}

	resp, err := h.service.Update(r.Context(), id, req)

	if err != nil {
		helper.WriteError(w, http.StatusNotFound, "language not found")

		return
	}

	helper.WriteJSON(w, http.StatusOK, resp)
}

// Delete godoc
// @Summary      Tilni o'chirish (soft delete, admin)
// @Tags         Languages
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Language ID"
// @Success      200 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Router       /languages/{id} [delete]
func (h *languageHandler) Delete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.ParseInt(ps.ByName("id"), 10, 64)

	if err != nil || id <= 0 {
		helper.WriteError(w, http.StatusBadRequest, "invalid id")

		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		helper.WriteError(w, http.StatusNotFound, err.Error())

		return
	}

	helper.WriteJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}
