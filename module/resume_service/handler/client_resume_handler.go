package resume_handler

import (
	"main_service/helper"
	resume_dto "main_service/module/resume_service/dto"
	resume_service "main_service/module/resume_service/service"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"
)

type clientResumeHandler struct {
	service resume_service.ResumeService
}

func NewClientResumeHandler(router *httprouter.Router, group string, db *pgxpool.Pool) {
	h := &clientResumeHandler{service: resume_service.NewResumeService(db)}

	routes := group + "/resumes-client"
	{
		router.GET(routes, h.List)

		router.GET(routes+"/:slug", h.GetBySlug)
	}
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
// @Router       /resumes-client [get]
func (h *clientResumeHandler) List(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	q := r.URL.Query()

	cursor, limit := helper.ParseCursorPayload(r)

	f := resume_dto.ResumeFilter{
		Name:          helper.QueryString(r, "name", 100),
		Title:         helper.QueryString(r, "title", 100),
		Search:        helper.QueryString(r, "search", 100),
		SortBy:        strings.TrimSpace(q.Get("sort_by")),
		SortOrder:     strings.TrimSpace(q.Get("sort_order")),
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

// GetBySlug godoc
// @Summary      Resumeni slug bo'yicha olish (public)
// @Tags         Resumes
// @Produce      json
// @Param        slug path string true "Resume slug"
// @Success      200 {object} resume_dto.ResumeResponse
// @Failure      404 {object} map[string]string
// @Router       /resumes-client/{slug} [get]
func (h *clientResumeHandler) GetBySlug(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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
