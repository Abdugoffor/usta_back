package country_handler

import (
	"context"
	"encoding/json"
	"main_service/helper"
	"main_service/middleware"
	country_dto "main_service/module/country_service/dto"
	country_service "main_service/module/country_service/service"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"
)

type countryHandler struct {
	service country_service.CountryService
}

func NewCountryHandler(router *httprouter.Router, group string, db *pgxpool.Pool) {
	h := &countryHandler{service: country_service.NewCountryService(db)}

	routes := group + "/countries"
	{
		router.POST(routes, middleware.CheckRole(h.Create, "admin"))

		router.GET(routes, middleware.CheckRole(h.List, "admin"))

		router.GET(routes+"/:id", middleware.CheckRole(h.GetByID, "admin"))

		router.PUT(routes+"/:id", middleware.CheckRole(h.Update, "admin"))

		router.DELETE(routes+"/:id", middleware.CheckRole(h.Delete, "admin"))
	}

	router.GET(group+"/count/countries", middleware.CheckRole(h.Count, "admin"))
}

func (h *countryHandler) checkActiveLangs(ctx context.Context, name map[string]string) (map[string]string, error) {
	langs, err := h.service.GetActiveLanguages(ctx)

	if err != nil {
		return nil, err
	}

	return helper.ValidateMultilang(name, langs, true), nil
}

// Create godoc
// @Summary      Davlat/region yaratish (admin, ko'p tilli)
// @Description  name maydoni JSONB: default + har bir faol til kalitlari majburiy
// @Tags         Countries
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body country_dto.CreateCountryRequest true "Country ma'lumotlari"
// @Success      201 {object} country_dto.CountryResponse
// @Failure      422 {object} map[string]any
// @Router       /countries [post]
func (h *countryHandler) Create(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req country_dto.CreateCountryRequest

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

// List godoc
// @Summary      Countries ro'yxati (admin, cursor pagination)
// @Tags         Countries
// @Produce      json
// @Security     BearerAuth
// @Param        cursor      query string false "Keyset cursor"
// @Param        limit       query int    false "Default 10, max 100"
// @Param        sort_by     query string false "id|name|parent_id|is_active|created_at|updated_at"
// @Param        sort_order  query string false "asc|desc"
// @Param        name        query string false "name JSONB ichida ILIKE"
// @Param        parent_id   query int    false "Aniq parent_id"
// @Param        has_parent  query bool   false "true => parent_id NOT NULL, false => root"
// @Param        is_active   query bool   false "Holat"
// @Success      200 {object} map[string]any
// @Router       /countries [get]
func (h *countryHandler) List(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	q := helper.ParseAdminList(r)

	f := country_dto.CountryFilter{
		Name:      helper.QueryString(r, "name", 100),
		ParentID:  helper.QueryInt64(r, "parent_id"),
		IsActive:  helper.QueryBool(r, "is_active"),
		HasParent: helper.QueryBool(r, "has_parent"),
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

		lastValue = countrySortValue(q.SortBy, last)
	}

	helper.WriteJSON(w, http.StatusOK, map[string]any{
		"data": items,
		"meta": helper.NewCursorMetaWithValue(q.Limit, hasMore, lastID, lastValue, 0),
	})
}

// Count godoc
// @Summary      Countries soni (admin)
// @Tags         Countries
// @Produce      json
// @Security     BearerAuth
// @Param        name       query string false "Name JSONB ichida ILIKE"
// @Param        parent_id  query int    false "Aniq parent_id"
// @Param        has_parent query bool   false "true => parent_id NOT NULL"
// @Param        is_active  query bool   false "Holat"
// @Success      200 {object} map[string]int64
// @Router       /count/countries [get]
func (h *countryHandler) Count(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	f := country_dto.CountryFilter{
		Name:      helper.QueryString(r, "name", 100),
		ParentID:  helper.QueryInt64(r, "parent_id"),
		IsActive:  helper.QueryBool(r, "is_active"),
		HasParent: helper.QueryBool(r, "has_parent"),
	}

	total, err := h.service.Count(r.Context(), f)

	if err != nil {
		helper.WriteInternalError(w, err)

		return
	}

	helper.WriteJSON(w, http.StatusOK, map[string]int64{"total": total})
}

func countrySortValue(sortBy string, c *country_dto.CountryResponse) string {
	switch sortBy {
	case "name":
		if v, ok := c.Name["default"]; ok {
			return v
		}

		return ""
	case "parent_id":
		if c.ParentID != nil {
			return strconv.FormatInt(*c.ParentID, 10)
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

// GetByID godoc
// @Summary      Countryni ID bo'yicha olish (admin)
// @Tags         Countries
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Country ID"
// @Success      200 {object} country_dto.CountryResponse
// @Failure      404 {object} map[string]string
// @Router       /countries/{id} [get]
func (h *countryHandler) GetByID(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.ParseInt(ps.ByName("id"), 10, 64)

	if err != nil || id <= 0 {
		helper.WriteError(w, http.StatusBadRequest, "invalid id")

		return
	}

	resp, err := h.service.GetByID(r.Context(), id)

	if err != nil {
		helper.WriteError(w, http.StatusNotFound, "country not found")

		return
	}

	helper.WriteJSON(w, http.StatusOK, resp)
}

// Update godoc
// @Summary      Countryni yangilash (admin)
// @Tags         Countries
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path int                              true  "Country ID"
// @Param        body body country_dto.UpdateCountryRequest true  "Yangilanadigan maydonlar"
// @Success      200 {object} country_dto.CountryResponse
// @Failure      404 {object} map[string]string
// @Failure      422 {object} map[string]any
// @Router       /countries/{id} [put]
func (h *countryHandler) Update(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.ParseInt(ps.ByName("id"), 10, 64)

	if err != nil || id <= 0 {
		helper.WriteError(w, http.StatusBadRequest, "invalid id")

		return
	}

	var req country_dto.UpdateCountryRequest

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
		helper.WriteError(w, http.StatusNotFound, "country not found")

		return
	}

	helper.WriteJSON(w, http.StatusOK, resp)
}

// Delete godoc
// @Summary      Countryni o'chirish (soft delete, admin)
// @Tags         Countries
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Country ID"
// @Success      200 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Router       /countries/{id} [delete]
func (h *countryHandler) Delete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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
