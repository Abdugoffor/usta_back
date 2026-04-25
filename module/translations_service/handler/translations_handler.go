package translation_handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"main_service/helper"
	"main_service/middleware"
	translation_dto "main_service/module/translations_service/dto"
	translation_service "main_service/module/translations_service/service"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"
)

type translationHandler struct {
	service translation_service.TranslationService
}

func NewTranslationHandler(router *httprouter.Router, group string, db *pgxpool.Pool) {
	h := &translationHandler{service: translation_service.NewTranslationService(db)}

	routes := group + "/translations"
	{
		router.GET(routes, middleware.CheckRole(h.List, "admin"))

		router.GET(routes+"/:id", middleware.CheckRole(h.GetByID, "admin"))

		router.POST(routes, middleware.CheckRole(h.Create, "admin"))

		router.PUT(routes+"/:id", middleware.CheckRole(h.Update, "admin"))

		router.DELETE(routes+"/:id", middleware.CheckRole(h.Delete, "admin"))
	}

	router.GET(group+"/count/translations", middleware.CheckRole(h.Count, "admin"))

	router.GET(group+"/t", h.GetTranslation)
}

func (h *translationHandler) checkActiveLangs(ctx context.Context, name map[string]string) (map[string]string, error) {
	langs, err := h.service.GetActiveLanguages(ctx)

	if err != nil {
		return nil, err
	}

	return helper.ValidateMultilang(name, langs, true), nil
}

// GetTranslation godoc
// @Summary      Frontend uchun bitta tarjimani olish (public)
// @Description  slug+lang bo'yicha qidiradi; topilmasa default; hech narsa bo'lmasa key qaytadi
// @Tags         Translations
// @Produce      json
// @Param        key  query string true  "Translation slug"
// @Param        lang query string false "Til kodi (default 'default')"
// @Success      200 {object} translation_dto.TranslationKeyResponse
// @Router       /t [get]
func (h *translationHandler) GetTranslation(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	key := strings.TrimSpace(r.URL.Query().Get("key"))

	if key == "" {
		helper.WriteError(w, http.StatusBadRequest, "key is required")

		return
	}

	lang := strings.TrimSpace(r.URL.Query().Get("lang"))

	if lang == "" {
		lang = "default"
	}

	value := h.service.GetTranslation(r.Context(), key, lang)

	helper.WriteJSON(w, http.StatusOK, translation_dto.TranslationKeyResponse{
		Key:   key,
		Value: value,
		Lang:  lang,
	})
}

// Create godoc
// @Summary      Tarjima yaratish (admin, ko'p tilli)
// @Tags         Translations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body translation_dto.CreateTranslationRequest true "Tarjima ma'lumotlari"
// @Success      201 {object} translation_dto.TranslationResponse
// @Failure      400 {object} map[string]string
// @Failure      422 {object} map[string]any
// @Router       /translations [post]
func (h *translationHandler) Create(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req translation_dto.CreateTranslationRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.WriteError(w, http.StatusBadRequest, "invalid JSON")

		return
	}

	req.Slug = strings.TrimSpace(req.Slug)

	helper.SanitizeMultilang(req.Name)

	if errs := helper.ValidateStruct(req); errs != nil {
		helper.WriteValidation(w, errs)

		return
	}

	errs, err := h.checkActiveLangs(r.Context(), req.Name)

	if err != nil {
		helper.WriteInternalError(w, err)

		return
	}

	if errs != nil {
		helper.WriteValidation(w, errs)

		return
	}

	resp, err := h.service.Create(r.Context(), req)

	if err != nil {
		helper.WriteError(w, http.StatusBadRequest, err.Error())

		return
	}

	helper.WriteJSON(w, http.StatusCreated, resp)
}

// List godoc
// @Summary      Tarjimalar ro'yxati (admin, cursor pagination)
// @Tags         Translations
// @Produce      json
// @Security     BearerAuth
// @Param        cursor      query string false "Keyset cursor"
// @Param        limit       query int    false "Default 10, max 100"
// @Param        sort_by     query string false "id|slug|name|is_active|created_at|updated_at"
// @Param        sort_order  query string false "asc|desc"
// @Param        slug        query string false "Slug bo'yicha ILIKE"
// @Param        name        query string false "Name JSONB ichida ILIKE"
// @Param        is_active   query bool   false "Holat"
// @Success      200 {object} map[string]any
// @Router       /translations [get]
func (h *translationHandler) List(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	q := helper.ParseAdminList(r)

	f := translation_dto.TranslationFilter{
		Slug:     helper.QueryString(r, "slug", 150),
		Name:     helper.QueryString(r, "name", 100),
		IsActive: helper.QueryBool(r, "is_active"),
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

		lastValue = translationSortValue(q.SortBy, last)
	}

	helper.WriteJSON(w, http.StatusOK, map[string]any{
		"data": items,
		"meta": helper.NewCursorMetaWithValue(q.Limit, hasMore, lastID, lastValue, 0),
	})
}

// Count godoc
// @Summary      Tarjimalar soni (admin)
// @Tags         Translations
// @Produce      json
// @Security     BearerAuth
// @Param        slug      query string false "Slug ILIKE"
// @Param        name      query string false "Name JSONB ICHIDA ILIKE"
// @Param        is_active query bool   false "Holat"
// @Success      200 {object} map[string]int64
// @Router       /count/translations [get]
func (h *translationHandler) Count(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	f := translation_dto.TranslationFilter{
		Slug:     helper.QueryString(r, "slug", 150),
		Name:     helper.QueryString(r, "name", 100),
		IsActive: helper.QueryBool(r, "is_active"),
	}

	total, err := h.service.Count(r.Context(), f)

	if err != nil {
		helper.WriteInternalError(w, err)

		return
	}

	helper.WriteJSON(w, http.StatusOK, map[string]int64{"total": total})
}

func translationSortValue(sortBy string, t *translation_dto.TranslationResponse) string {
	switch sortBy {
	case "slug":
		return t.Slug
	case "name":
		if v, ok := t.Name["default"]; ok {
			return v
		}

		return ""
	case "is_active":
		return strconv.FormatBool(t.IsActive)
	case "created_at":
		return t.CreatedAt.Format("2006-01-02 15:04:05.000000-07:00")
	case "updated_at":
		return t.UpdatedAt.Format("2006-01-02 15:04:05.000000-07:00")
	}

	return ""
}

// GetByID godoc
// @Summary      Tarjimani ID bo'yicha olish (admin)
// @Tags         Translations
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Translation ID"
// @Success      200 {object} translation_dto.TranslationResponse
// @Failure      404 {object} map[string]string
// @Router       /translations/{id} [get]
func (h *translationHandler) GetByID(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.ParseInt(ps.ByName("id"), 10, 64)

	if err != nil || id <= 0 {
		helper.WriteError(w, http.StatusBadRequest, "invalid id")

		return
	}

	resp, err := h.service.GetByID(r.Context(), id)

	if err != nil {
		helper.WriteError(w, http.StatusNotFound, "translation not found")

		return
	}

	helper.WriteJSON(w, http.StatusOK, resp)
}

// Update godoc
// @Summary      Tarjimani yangilash (admin)
// @Tags         Translations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path int                                       true "Translation ID"
// @Param        body body translation_dto.UpdateTranslationRequest  true "Yangilanadigan maydonlar"
// @Success      200 {object} translation_dto.TranslationResponse
// @Failure      400 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Failure      422 {object} map[string]any
// @Router       /translations/{id} [put]
func (h *translationHandler) Update(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.ParseInt(ps.ByName("id"), 10, 64)

	if err != nil || id <= 0 {
		helper.WriteError(w, http.StatusBadRequest, "invalid id")

		return
	}

	var req translation_dto.UpdateTranslationRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.WriteError(w, http.StatusBadRequest, "invalid JSON")

		return
	}

	if req.Slug != nil {
		trimmed := strings.TrimSpace(*req.Slug)

		req.Slug = &trimmed
	}

	if req.Name != nil {
		helper.SanitizeMultilang(*req.Name)
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		helper.WriteValidation(w, errs)

		return
	}

	if req.Name != nil {
		errs, err := h.checkActiveLangs(r.Context(), *req.Name)

		if err != nil {
			helper.WriteInternalError(w, err)

			return
		}

		if errs != nil {
			helper.WriteValidation(w, errs)

			return
		}
	}

	resp, err := h.service.Update(r.Context(), id, req)

	if err != nil {
		helper.WriteError(w, http.StatusBadRequest, err.Error())

		return
	}

	helper.WriteJSON(w, http.StatusOK, resp)
}

// Delete godoc
// @Summary      Tarjimani o'chirish (soft delete, admin)
// @Tags         Translations
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Translation ID"
// @Success      200 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Router       /translations/{id} [delete]
func (h *translationHandler) Delete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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
