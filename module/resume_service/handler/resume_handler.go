package resume_handler

import (
	"encoding/json"
	"main_service/helper"
	"main_service/middleware"
	resume_dto "main_service/module/resume_service/dto"
	resume_service "main_service/module/resume_service/service"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"
)

type resumeHandler struct {
	service resume_service.ResumeService
}

func NewResumeHandler(router *httprouter.Router, group string, db *pgxpool.Pool) {
	h := &resumeHandler{service: resume_service.NewResumeService(db)}

	routes := group + "/resumes"
	{
		router.POST(routes, middleware.CheckRole(h.Create))

		router.GET(routes, h.List)

		router.GET(routes+"/:slug", h.GetBySlug)

		router.PUT(routes+"/:id", middleware.CheckRole(h.Update))

		router.DELETE(routes+"/:id", middleware.CheckRole(h.Delete))
	}

	router.GET(group+"/count/resumes", h.Count)
}

// Create godoc
// @Summary      Resume yaratish (auth user)
// @Description  category_ids — bir nechta kategoriya berish mumkin
// @Tags         Resumes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body resume_dto.CreateResumeRequest true "Resume"
// @Success      201 {object} resume_dto.ResumeResponse
// @Failure      401 {object} map[string]string
// @Failure      422 {object} map[string]any
// @Router       /resumes [post]
func (h *resumeHandler) Create(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	userID := middleware.GetUserID(r)

	if userID == 0 {
		helper.WriteError(w, http.StatusUnauthorized, "unauthorized")

		return
	}

	var req resume_dto.CreateResumeRequest

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
// @Summary      Resumelar ro'yxati (cursor pagination, public)
// @Tags         Resumes
// @Produce      json
// @Param        cursor         query string false "Keyset cursor"
// @Param        limit          query int    false "Default 10, max 100"
// @Param        sort_by        query string false "id|price|experience_year"
// @Param        sort_order     query string false "asc|desc"
// @Param        name           query string false "Name ILIKE"
// @Param        title          query string false "Title ILIKE"
// @Param        search         query string false "Name+title+skills+text bo'yicha"
// @Param        user_id        query int    false "Egasi ID"
// @Param        region_id      query int    false "Region ID"
// @Param        district_id    query int    false "District ID"
// @Param        mahalla_id     query int    false "Mahalla ID"
// @Param        is_active      query bool   false "Holat"
// @Param        min_price      query int    false "Minimal narx"
// @Param        max_price      query int    false "Maksimal narx"
// @Param        min_experience query int    false "Minimal tajriba (yil)"
// @Param        category_id    query int    false "Aniq bitta kategoriya"
// @Param        category_ids   query string false "Bir nechta kategoriya, vergulli (1,2,3)"
// @Success      200 {object} map[string]any
// @Router       /resumes [get]
func (h *resumeHandler) List(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	q := r.URL.Query()

	cursor, limit := helper.ParseCursorPayload(r)

	f := resume_dto.ResumeFilter{
		Name:          helper.QueryString(r, "name", 100),
		Title:         helper.QueryString(r, "title", 100),
		Search:        helper.QueryString(r, "search", 100),
		SortBy:        strings.TrimSpace(q.Get("sort_by")),
		SortOrder:     strings.TrimSpace(q.Get("sort_order")),
		UserID:        helper.QueryInt64(r, "user_id"),
		RegionID:      helper.QueryInt64(r, "region_id"),
		DistrictID:    helper.QueryInt64(r, "district_id"),
		MahallaID:     helper.QueryInt64(r, "mahalla_id"),
		IsActive:      helper.QueryBool(r, "is_active"),
		MinPrice:      helper.QueryInt64(r, "min_price"),
		MaxPrice:      helper.QueryInt64(r, "max_price"),
		MinExperience: helper.QueryInt(r, "min_experience"),
		CategoryID:    helper.QueryInt64(r, "category_id"),
		CategoryIDs:   helper.ParseIDList(q.Get("category_ids"), 20),
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

		switch f.SortBy {
		case "price":
			if items[len(items)-1].Price != nil {
				lastValue = strconv.FormatInt(*items[len(items)-1].Price, 10)
			}
		case "experience_year":
			if items[len(items)-1].ExperienceYear != nil {
				lastValue = strconv.Itoa(*items[len(items)-1].ExperienceYear)
			}
		}
	}

	helper.WriteJSON(w, http.StatusOK, map[string]any{
		"data": items,
		"meta": helper.NewCursorMetaWithValue(limit, hasMore, lastID, lastValue, 0),
	})
}

// Count godoc
// @Summary      Resumelar soni (public)
// @Tags         Resumes
// @Produce      json
// @Param        name           query string false "Name ILIKE"
// @Param        title          query string false "Title ILIKE"
// @Param        search         query string false "Name+title+skills+text"
// @Param        user_id        query int    false "Egasi ID"
// @Param        region_id      query int    false "Region ID"
// @Param        district_id    query int    false "District ID"
// @Param        mahalla_id     query int    false "Mahalla ID"
// @Param        is_active      query bool   false "Holat"
// @Param        min_price      query int    false "Minimal narx"
// @Param        max_price      query int    false "Maksimal narx"
// @Param        min_experience query int    false "Minimal tajriba"
// @Param        category_id    query int    false "Aniq bitta kategoriya"
// @Param        category_ids   query string false "Vergulli (1,2,3)"
// @Success      200 {object} map[string]int64
// @Router       /count/resumes [get]
func (h *resumeHandler) Count(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	q := r.URL.Query()

	f := resume_dto.ResumeFilter{
		Name:          helper.QueryString(r, "name", 100),
		Title:         helper.QueryString(r, "title", 100),
		Search:        helper.QueryString(r, "search", 100),
		UserID:        helper.QueryInt64(r, "user_id"),
		RegionID:      helper.QueryInt64(r, "region_id"),
		DistrictID:    helper.QueryInt64(r, "district_id"),
		MahallaID:     helper.QueryInt64(r, "mahalla_id"),
		IsActive:      helper.QueryBool(r, "is_active"),
		MinPrice:      helper.QueryInt64(r, "min_price"),
		MaxPrice:      helper.QueryInt64(r, "max_price"),
		MinExperience: helper.QueryInt(r, "min_experience"),
		CategoryID:    helper.QueryInt64(r, "category_id"),
		CategoryIDs:   helper.ParseIDList(q.Get("category_ids"), 20),
	}

	total, err := h.service.Count(r.Context(), f)

	if err != nil {
		helper.WriteInternalError(w, err)

		return
	}

	helper.WriteJSON(w, http.StatusOK, map[string]int64{"total": total})
}

// GetBySlug godoc
// @Summary      Resumeni slug bo'yicha olish (public)
// @Tags         Resumes
// @Produce      json
// @Param        slug path string true "Resume slug"
// @Success      200 {object} resume_dto.ResumeResponse
// @Failure      404 {object} map[string]string
// @Router       /resumes/{slug} [get]
func (h *resumeHandler) GetBySlug(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	slug := ps.ByName("slug")

	if slug == "" {
		helper.WriteError(w, http.StatusBadRequest, "invalid slug")

		return
	}

	resp, err := h.service.GetBySlug(r.Context(), slug)

	if err != nil {
		helper.WriteError(w, http.StatusNotFound, "resume not found")

		return
	}

	helper.WriteJSON(w, http.StatusOK, resp)
}

// Update godoc
// @Summary      Resumeni yangilash (egasi yoki admin)
// @Description  category_ids berilsa — to'liq almashtiriladi
// @Tags         Resumes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path int                            true "Resume ID"
// @Param        body body resume_dto.UpdateResumeRequest true "Yangilanadigan maydonlar"
// @Success      200 {object} resume_dto.ResumeResponse
// @Failure      401 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Failure      422 {object} map[string]any
// @Router       /resumes/{id} [put]
func (h *resumeHandler) Update(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

	var req resume_dto.UpdateResumeRequest

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
// @Summary      Resumeni o'chirish (egasi yoki admin)
// @Tags         Resumes
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Resume ID"
// @Success      200 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Router       /resumes/{id} [delete]
func (h *resumeHandler) Delete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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
