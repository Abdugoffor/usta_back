package categorya_handler

import (
	"context"
	"encoding/json"
	"main_service/helper"
	"main_service/middleware"
	categorya_dto "main_service/module/categorya_service/dto"
	categorya_service "main_service/module/categorya_service/service"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"
)

type categoryHandler struct {
	service categorya_service.CategoryService
}

func NewCategoryHandler(router *httprouter.Router, group string, db *pgxpool.Pool) {
	h := &categoryHandler{service: categorya_service.NewCategoryService(db)}

	routes := group + "/categories"
	{
		router.POST(routes, middleware.CheckRole(h.Create, "admin"))

		router.GET(routes, middleware.CheckRole(h.List, "admin"))

		router.GET(routes+"/:id", middleware.CheckRole(h.Show, "admin"))

		router.PUT(routes+"/:id", middleware.CheckRole(h.Update, "admin"))

		router.DELETE(routes+"/:id", middleware.CheckRole(h.Delete, "admin"))
	}

	router.GET(group+"/count/categories", middleware.CheckRole(h.Count, "admin"))
}

func (h *categoryHandler) checkActiveLangs(ctx context.Context, name map[string]string) (map[string]string, error) {
	langs, err := h.service.GetActiveLanguages(ctx)

	if err != nil {
		return nil, err
	}

	return helper.ValidateMultilang(name, langs, false), nil
}

// Create godoc
// @Summary      Kategoriya yaratish (admin, ko'p tilli)
// @Description  name JSONB: har bir faol til kalitlari kiritilishi kerak
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body categorya_dto.CreateCategoryRequest true "Kategoriya ma'lumotlari"
// @Success      201 {object} categorya_dto.CategoryResponse
// @Failure      422 {object} map[string]any
// @Router       /categories [post]
func (h *categoryHandler) Create(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req categorya_dto.CreateCategoryRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.WriteError(w, http.StatusBadRequest, "invalid JSON")

		return
	}

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
		helper.WriteInternalError(w, err)

		return
	}

	helper.WriteJSON(w, http.StatusCreated, resp)
}

// Update godoc
// @Summary      Kategoriyani yangilash (admin)
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path int                                  true  "Category ID"
// @Param        body body categorya_dto.UpdateCategoryRequest  true  "Yangilanadigan maydonlar"
// @Success      200 {object} categorya_dto.CategoryResponse
// @Failure      404 {object} map[string]string
// @Failure      422 {object} map[string]any
// @Router       /categories/{id} [put]
func (h *categoryHandler) Update(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.ParseInt(ps.ByName("id"), 10, 64)

	if err != nil || id <= 0 {
		helper.WriteError(w, http.StatusBadRequest, "invalid id")

		return
	}

	var req categorya_dto.UpdateCategoryRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.WriteError(w, http.StatusBadRequest, "invalid JSON")

		return
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
		helper.WriteError(w, http.StatusNotFound, "category not found")

		return
	}

	helper.WriteJSON(w, http.StatusOK, resp)
}

// Delete godoc
// @Summary      Kategoriyani o'chirish (soft delete, admin)
// @Tags         Categories
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Category ID"
// @Success      200 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Router       /categories/{id} [delete]
func (h *categoryHandler) Delete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

// Show godoc
// @Summary      Kategoriyani ID bo'yicha olish (admin)
// @Tags         Categories
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Category ID"
// @Success      200 {object} categorya_dto.CategoryResponse
// @Failure      404 {object} map[string]string
// @Router       /categories/{id} [get]
func (h *categoryHandler) Show(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.ParseInt(ps.ByName("id"), 10, 64)

	if err != nil || id <= 0 {
		helper.WriteError(w, http.StatusBadRequest, "invalid id")

		return
	}

	resp, err := h.service.Show(r.Context(), id)

	if err != nil {
		helper.WriteError(w, http.StatusNotFound, "category not found")

		return
	}

	helper.WriteJSON(w, http.StatusOK, resp)
}

// List godoc
// @Summary      Kategoriyalar ro'yxati (admin, cursor pagination)
// @Tags         Categories
// @Produce      json
// @Security     BearerAuth
// @Param        cursor      query string false "Keyset cursor"
// @Param        limit       query int    false "Default 10, max 100"
// @Param        sort_by     query string false "id|name|is_active|created_at|updated_at"
// @Param        sort_order  query string false "asc|desc"
// @Param        name        query string false "Name JSONB ichida ILIKE"
// @Param        is_active   query bool   false "Holat"
// @Success      200 {object} map[string]any
// @Router       /categories [get]
func (h *categoryHandler) List(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	q := helper.ParseAdminList(r)

	f := categorya_dto.CategoryFilter{
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

		lastValue = categorySortValue(q.SortBy, last)
	}

	helper.WriteJSON(w, http.StatusOK, map[string]any{
		"data": items,
		"meta": helper.NewCursorMetaWithValue(q.Limit, hasMore, lastID, lastValue, 0),
	})
}

// Count godoc
// @Summary      Kategoriyalar soni (admin)
// @Tags         Categories
// @Produce      json
// @Security     BearerAuth
// @Param        name      query string false "Name JSONB ichida ILIKE"
// @Param        is_active query bool   false "Holat"
// @Success      200 {object} map[string]int64
// @Router       /count/categories [get]
func (h *categoryHandler) Count(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	f := categorya_dto.CategoryFilter{
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

func categorySortValue(sortBy string, c *categorya_dto.CategoryResponse) string {
	switch sortBy {
	case "name":
		if v, ok := c.Name["default"]; ok {
			return v
		}

		return ""
	case "is_active":
		return strconv.FormatBool(c.IsActive)
	case "created_at":
		return c.CreatedAt.Format("2006-01-02 15:04:05.000000-07:00")
	case "updated_at":
		return c.UpdatedAt.Format("2006-01-02 15:04:05.000000-07:00")
	}

	return ""
}
