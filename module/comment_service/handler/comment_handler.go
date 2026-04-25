package comment_handler

import (
	"encoding/json"
	"main_service/helper"
	"main_service/middleware"
	comment_dto "main_service/module/comment_service/dto"
	comment_service "main_service/module/comment_service/service"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"
)

type commentHandler struct {
	service comment_service.CommentService
}

func NewCommentHandler(router *httprouter.Router, group string, db *pgxpool.Pool) {
	h := &commentHandler{service: comment_service.NewCommentService(db)}

	routes := group + "/comments"
	{
		router.POST(routes, middleware.CheckRole(h.Create))
		router.GET(routes, h.List)
		router.GET(routes+"/:id", h.GetByID)
		router.PUT(routes+"/:id", middleware.CheckRole(h.Update))
		router.DELETE(routes+"/:id", middleware.CheckRole(h.Delete))
	}

	router.GET(group+"/count/comments", h.Count)
}

func (h *commentHandler) Create(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	userID := middleware.GetUserID(r)
	if userID == 0 {
		helper.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req comment_dto.CreateCommentRequest
	{
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			helper.WriteError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
	}

	if req.VakansiyaID == nil && req.ResumeID == nil {
		helper.WriteError(w, http.StatusUnprocessableEntity, "vakansiya_id or resume_id required")
		return
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		helper.WriteValidation(w, errs)
		return
	}

	resp, err := h.service.Create(r.Context(), int64(userID), req)
	{
		if err != nil {
			helper.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	helper.WriteJSON(w, http.StatusCreated, resp)
}

func (h *commentHandler) List(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	q := r.URL.Query()
	pq := helper.ParsePage(r)
	{
		if pq.Limit < 1 || pq.Limit > 100 {
			pq.Limit = 20
		}
	}

	f := comment_dto.CommentFilter{Type: q.Get("type")}
	{
		if v := q.Get("vakansiya_id"); v != "" {
			if n, err := strconv.ParseInt(v, 10, 64); err == nil {
				f.VakansiyaID = &n
			}
		}

		if v := q.Get("resume_id"); v != "" {
			if n, err := strconv.ParseInt(v, 10, 64); err == nil {
				f.ResumeID = &n
			}
		}

		if v := q.Get("user_id"); v != "" {
			if n, err := strconv.ParseInt(v, 10, 64); err == nil {
				f.UserID = &n
			}
		}
	}

	items, err := h.service.List(r.Context(), f, pq.Page, pq.Limit, pq.SortCol, pq.SortOrder)
	{
		if err != nil {
			helper.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	helper.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"data": items,
		"meta": helper.NewPageMeta(0, pq.Page, pq.Limit),
	})
}

func (h *commentHandler) Count(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	q := r.URL.Query()

	f := comment_dto.CommentFilter{Type: q.Get("type")}

	if v := q.Get("vakansiya_id"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			f.VakansiyaID = &n
		}
	}

	if v := q.Get("resume_id"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			f.ResumeID = &n
		}
	}

	if v := q.Get("user_id"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			f.UserID = &n
		}
	}

	total, err := h.service.Count(r.Context(), f)

	if err != nil {
		helper.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	helper.WriteJSON(w, http.StatusOK, map[string]int64{"total": total})
}

func (h *commentHandler) GetByID(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	id, err := strconv.ParseInt(ps.ByName("id"), 10, 64)
	{
		if err != nil || id <= 0 {
			helper.WriteError(w, http.StatusBadRequest, "invalid id")
			return
		}
	}

	resp, err := h.service.GetByID(r.Context(), id)
	{
		if err != nil {
			helper.WriteError(w, http.StatusNotFound, "comment not found")
			return
		}
	}

	helper.WriteJSON(w, http.StatusOK, resp)
}

func (h *commentHandler) Update(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	
	userID := middleware.GetUserID(r)
	{
		if userID == 0 {
			helper.WriteError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
	}

	id, err := strconv.ParseInt(ps.ByName("id"), 10, 64)
	{
		if err != nil || id <= 0 {
			helper.WriteError(w, http.StatusBadRequest, "invalid id")
			return
		}
	}

	var req comment_dto.UpdateCommentRequest
	{
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			helper.WriteError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		helper.WriteValidation(w, errs)
		return
	}

	resp, err := h.service.Update(r.Context(), id, int64(userID), req)
	{
		if err != nil {
			helper.WriteError(w, http.StatusNotFound, err.Error())
			return
		}
	}

	helper.WriteJSON(w, http.StatusOK, resp)
}

func (h *commentHandler) Delete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userID := middleware.GetUserID(r)
	{
		if userID == 0 {
			helper.WriteError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
	}

	id, err := strconv.ParseInt(ps.ByName("id"), 10, 64)
	{
		if err != nil || id <= 0 {
			helper.WriteError(w, http.StatusBadRequest, "invalid id")
			return
		}
	}

	if err := h.service.Delete(r.Context(), id, int64(userID)); err != nil {
		helper.WriteError(w, http.StatusNotFound, err.Error())
		return
	}

	helper.WriteJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}
