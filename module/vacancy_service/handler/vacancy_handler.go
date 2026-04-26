package vacancy_handler

import (
	"encoding/json"
	"main_service/helper"
	"main_service/middleware"
	vacancy_dto "main_service/module/vacancy_service/dto"
	vacancy_service "main_service/module/vacancy_service/service"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"
)

type vacancyHandler struct {
	service vacancy_service.VacancyService
}

func NewVacancyHandler(router *httprouter.Router, group string, db *pgxpool.Pool) {
	h := &vacancyHandler{service: vacancy_service.NewVacancyService(db)}

	routes := group + "/vacancies"
	{
		router.POST(routes, middleware.CheckRole(h.Create))

		router.GET(routes, middleware.CheckRole(h.List))

		router.PUT(routes+"/:id", middleware.CheckRole(h.Update))

		router.DELETE(routes+"/:id", middleware.CheckRole(h.Delete))
	}

	router.GET(group+"/count/vacancies", h.Count)
}

// Create godoc
// @Summary      Vakansiya yaratish (auth user)
// @Description  category_ids — bir nechta kategoriya berish mumkin
// @Tags         Vacancies
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body vacancy_dto.CreateVacancyRequest true "Vakansiya"
// @Success      201 {object} vacancy_dto.VacancyResponse
// @Failure      401 {object} map[string]string
// @Failure      422 {object} map[string]any
// @Router       /vacancies [post]
func (h *vacancyHandler) Create(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	userID := middleware.GetUserID(r)

	if userID == 0 {
		helper.WriteError(w, http.StatusUnauthorized, "unauthorized")

		return
	}

	var req vacancy_dto.CreateVacancyRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.WriteError(w, http.StatusBadRequest, "invalid JSON")

		return
	}

	req.Name = strings.TrimSpace(req.Name)

	req.Title = strings.TrimSpace(req.Title)

	req.Adress = strings.TrimSpace(req.Adress)

	if errs := helper.ValidateStruct(req); errs != nil {
		helper.WriteValidation(w, errs)

		return
	}

	resp, err := h.service.Create(r.Context(), int64(userID), req)

	if err != nil {
		helper.WriteInternalError(w, err)

		return
	}

	helper.WriteJSON(w, http.StatusCreated, resp)
}

// List godoc
// @Summary      Vakansiyalar ro'yxati (cursor pagination, auth user)
// @Tags         Vacancies
// @Produce      json
// @Param        cursor       query string false "Keyset cursor"
// @Param        limit        query int    false "Default 10, max 100"
// @Param        sort_by      query string false "id|price"
// @Param        sort_order   query string false "asc|desc"
// @Param        name         query string false "Name ILIKE"
// @Param        title        query string false "Title ILIKE"
// @Param        search       query string false "Name+title+text bo'yicha qidirish"
// @Param        user_id      query int    false "Egasi ID"
// @Param        region_id    query int    false "Region ID"
// @Param        district_id  query int    false "District ID"
// @Param        mahalla_id   query int    false "Mahalla ID"
// @Param        is_active    query bool   false "Holat"
// @Param        min_price    query int    false "Minimal narx"
// @Param        max_price    query int    false "Maksimal narx"
// @Param        category_id  query int    false "Aniq bitta kategoriya"
// @Param        category_ids query string false "Bir nechta kategoriya, vergulli (1,2,3)"
// @Success      200 {object} map[string]any
// @Router       /vacancies [get]
func (h *vacancyHandler) List(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	userID := middleware.GetUserID(r)
	{
		if userID == 0 {
			helper.WriteError(w, http.StatusUnauthorized, "unauthorized")

			return
		}
	}

	q := r.URL.Query()

	cursor, limit := helper.ParseCursorPayload(r)

	f := vacancy_dto.VacancyFilter{
		Name:        helper.QueryString(r, "name", 100),
		Title:       helper.QueryString(r, "title", 100),
		Search:      helper.QueryString(r, "search", 100),
		SortBy:      strings.TrimSpace(q.Get("sort_by")),
		SortOrder:   strings.TrimSpace(q.Get("sort_order")),
		UserID:      helper.QueryInt64(r, "user_id"),
		RegionID:    helper.QueryInt64(r, "region_id"),
		DistrictID:  helper.QueryInt64(r, "district_id"),
		MahallaID:   helper.QueryInt64(r, "mahalla_id"),
		IsActive:    helper.QueryBool(r, "is_active"),
		MinPrice:    helper.QueryInt64(r, "min_price"),
		MaxPrice:    helper.QueryInt64(r, "max_price"),
		CategoryID:  helper.QueryInt64(r, "category_id"),
		CategoryIDs: helper.ParseIDList(q.Get("category_ids"), 20),
	}

	role := middleware.GetRole(r)
	{
		if role != "admin" {
			f.UserID = &userID
		}
	}

	items, hasMore, err := h.service.List(r.Context(), f, cursor, limit)

	if err != nil {
		helper.WriteInternalError(w, err)

		return
	}

	var lastID int64

	var lastValue string

	if len(items) > 0 {
		lastID = items[len(items)-1].ID

		if f.SortBy == "price" && items[len(items)-1].Price != nil {
			lastValue = strconv.FormatInt(*items[len(items)-1].Price, 10)
		}
	}

	helper.WriteJSON(w, http.StatusOK, map[string]any{
		"data": items,
		"meta": helper.NewCursorMetaWithValue(limit, hasMore, lastID, lastValue, 0),
	})
}

// Count godoc
// @Summary      Vakansiyalar soni (public)
// @Tags         Vacancies
// @Produce      json
// @Param        name         query string false "Name ILIKE"
// @Param        title        query string false "Title ILIKE"
// @Param        search       query string false "Name+title+text"
// @Param        user_id      query int    false "Egasi ID"
// @Param        region_id    query int    false "Region ID"
// @Param        district_id  query int    false "District ID"
// @Param        mahalla_id   query int    false "Mahalla ID"
// @Param        is_active    query bool   false "Holat"
// @Param        min_price    query int    false "Minimal narx"
// @Param        max_price    query int    false "Maksimal narx"
// @Param        category_id  query int    false "Aniq bitta kategoriya"
// @Param        category_ids query string false "Vergulli (1,2,3)"
// @Success      200 {object} map[string]int64
// @Router       /count/vacancies [get]
func (h *vacancyHandler) Count(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	q := r.URL.Query()

	f := vacancy_dto.VacancyFilter{
		Name:        helper.QueryString(r, "name", 100),
		Title:       helper.QueryString(r, "title", 100),
		Search:      helper.QueryString(r, "search", 100),
		UserID:      helper.QueryInt64(r, "user_id"),
		RegionID:    helper.QueryInt64(r, "region_id"),
		DistrictID:  helper.QueryInt64(r, "district_id"),
		MahallaID:   helper.QueryInt64(r, "mahalla_id"),
		IsActive:    helper.QueryBool(r, "is_active"),
		MinPrice:    helper.QueryInt64(r, "min_price"),
		MaxPrice:    helper.QueryInt64(r, "max_price"),
		CategoryID:  helper.QueryInt64(r, "category_id"),
		CategoryIDs: helper.ParseIDList(q.Get("category_ids"), 20),
	}

	total, err := h.service.Count(r.Context(), f)

	if err != nil {
		helper.WriteInternalError(w, err)

		return
	}

	helper.WriteJSON(w, http.StatusOK, map[string]int64{"total": total})
}

// Update godoc
// @Summary      Vakansiyani yangilash (egasi yoki admin)
// @Description  category_ids berilsa — to'liq almashtiriladi
// @Tags         Vacancies
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path int                              true "Vacancy ID"
// @Param        body body vacancy_dto.UpdateVacancyRequest true "Yangilanadigan maydonlar"
// @Success      200 {object} vacancy_dto.VacancyResponse
// @Failure      401 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Failure      422 {object} map[string]any
// @Router       /vacancies/{id} [put]
func (h *vacancyHandler) Update(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userID := middleware.GetUserID(r)

	if userID == 0 {
		helper.WriteError(w, http.StatusUnauthorized, "unauthorized")

		return
	}

	id, err := strconv.ParseInt(ps.ByName("id"), 10, 64)

	if err != nil || id <= 0 {
		helper.WriteError(w, http.StatusBadRequest, "invalid id")

		return
	}

	var req vacancy_dto.UpdateVacancyRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.WriteError(w, http.StatusBadRequest, "invalid JSON")

		return
	}

	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)

		req.Name = &trimmed
	}

	if req.Title != nil {
		trimmed := strings.TrimSpace(*req.Title)

		req.Title = &trimmed
	}

	if req.Adress != nil {
		trimmed := strings.TrimSpace(*req.Adress)

		req.Adress = &trimmed
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		helper.WriteValidation(w, errs)

		return
	}

	isAdmin := middleware.GetRole(r) == "admin"

	resp, err := h.service.Update(r.Context(), id, int64(userID), isAdmin, req)

	if err != nil {
		helper.WriteError(w, http.StatusNotFound, err.Error())

		return
	}

	helper.WriteJSON(w, http.StatusOK, resp)
}

// Delete godoc
// @Summary      Vakansiyani o'chirish (egasi yoki admin)
// @Tags         Vacancies
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Vacancy ID"
// @Success      200 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Router       /vacancies/{id} [delete]
func (h *vacancyHandler) Delete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userID := middleware.GetUserID(r)

	if userID == 0 {
		helper.WriteError(w, http.StatusUnauthorized, "unauthorized")

		return
	}

	id, err := strconv.ParseInt(ps.ByName("id"), 10, 64)

	if err != nil || id <= 0 {
		helper.WriteError(w, http.StatusBadRequest, "invalid id")

		return
	}

	isAdmin := middleware.GetRole(r) == "admin"

	if err := h.service.Delete(r.Context(), id, int64(userID), isAdmin); err != nil {
		helper.WriteError(w, http.StatusNotFound, err.Error())

		return
	}

	helper.WriteJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}
