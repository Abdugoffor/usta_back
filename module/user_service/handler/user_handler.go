package user_handler

import (
	"encoding/json"
	"main_service/helper"
	"main_service/middleware"
	user_dto "main_service/module/user_service/dto"
	user_service "main_service/module/user_service/service"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"
)

type userHandler struct {
	service user_service.UserService
}

func NewUserHandler(router *httprouter.Router, group string, db *pgxpool.Pool) {
	h := &userHandler{service: user_service.NewUserService(db)}

	auth := group + "/auth"
	{
		router.POST(auth+"/register", h.Register)

		router.POST(auth+"/login", h.Login)
	}

	users := group + "/users"
	{
		router.GET(users, middleware.CheckRole(h.List, "admin"))

		router.GET(users+"/:id", middleware.CheckRole(h.Show, "admin"))

		router.POST(users, middleware.CheckRole(h.Create, "admin"))

		router.PUT(users+"/:id", middleware.CheckRole(h.Update, "admin"))

		router.DELETE(users+"/:id", middleware.CheckRole(h.Delete, "admin"))
	}

	router.GET(group+"/count/users", middleware.CheckRole(h.Count, "admin"))
}

// Register godoc
// @Summary      Foydalanuvchi ro'yxatdan o'tish
// @Description  Yangi foydalanuvchi yaratish va JWT token qaytarish
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body body user_dto.RegisterRequest true "Ro'yxatdan o'tish ma'lumotlari"
// @Success      201 {object} user_dto.AuthResponse
// @Failure      400 {object} map[string]string
// @Failure      422 {object} map[string]any
// @Router       /auth/register [post]
func (h *userHandler) Register(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req user_dto.RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.WriteError(w, http.StatusBadRequest, "invalid JSON")

		return
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		helper.WriteValidation(w, errs)

		return
	}

	resp, err := h.service.Register(r.Context(), req)

	if err != nil {
		helper.WriteError(w, http.StatusBadRequest, err.Error())

		return
	}

	helper.WriteJSON(w, http.StatusCreated, resp)
}

// Login godoc
// @Summary      Tizimga kirish
// @Description  Phone+password orqali JWT token olish
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body body user_dto.LoginRequest true "Kirish ma'lumotlari"
// @Success      200 {object} user_dto.AuthResponse
// @Failure      401 {object} map[string]string
// @Failure      422 {object} map[string]any
// @Router       /auth/login [post]
func (h *userHandler) Login(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req user_dto.LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.WriteError(w, http.StatusBadRequest, "invalid JSON")

		return
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		helper.WriteValidation(w, errs)

		return
	}

	resp, err := h.service.Login(r.Context(), req)

	if err != nil {
		helper.WriteError(w, http.StatusUnauthorized, err.Error())

		return
	}

	helper.WriteJSON(w, http.StatusOK, resp)
}

// Create godoc
// @Summary      Foydalanuvchi yaratish (admin)
// @Tags         Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body user_dto.CreateUserRequest true "User ma'lumotlari"
// @Success      201 {object} user_dto.UserResponse
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      403 {object} map[string]string
// @Failure      422 {object} map[string]any
// @Router       /users [post]
func (h *userHandler) Create(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req user_dto.CreateUserRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.WriteError(w, http.StatusBadRequest, "invalid JSON")

		return
	}

	req.FullName = strings.TrimSpace(req.FullName)

	req.Phone = strings.TrimSpace(req.Phone)

	if errs := helper.ValidateStruct(req); errs != nil {
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
// @Summary      Foydalanuvchilar ro'yxati (admin, cursor pagination)
// @Tags         Users
// @Produce      json
// @Security     BearerAuth
// @Param        cursor     query string false "Keyset cursor (next_cursor dan olingan)"
// @Param        limit      query int    false "Sahifa hajmi (default 10, max 100)"
// @Param        sort_by    query string false "id|full_name|phone|role|is_active|created_at|updated_at"
// @Param        sort_order query string false "asc|desc (default desc)"
// @Param        full_name  query string false "Ism bo'yicha qidirish (ILIKE)"
// @Param        phone      query string false "Telefon bo'yicha qidirish (ILIKE)"
// @Param        role       query string false "Aniq rol (user|employer|admin)"
// @Param        is_active  query bool   false "Holat filtri"
// @Success      200 {object} map[string]any
// @Failure      401 {object} map[string]string
// @Failure      403 {object} map[string]string
// @Router       /users [get]
func (h *userHandler) List(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	q := helper.ParseAdminList(r)

	f := user_dto.UserFilter{
		FullName: helper.QueryString(r, "full_name", 100),
		Phone:    helper.QueryString(r, "phone", 50),
		Role:     helper.QueryString(r, "role", 50),
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

		lastValue = userSortValue(q.SortBy, last)
	}

	helper.WriteJSON(w, http.StatusOK, map[string]any{
		"data": items,
		"meta": helper.NewCursorMetaWithValue(q.Limit, hasMore, lastID, lastValue, 0),
	})
}

// Count godoc
// @Summary      Foydalanuvchilar soni (admin)
// @Tags         Users
// @Produce      json
// @Security     BearerAuth
// @Param        full_name  query string false "Ism bo'yicha qidirish (ILIKE)"
// @Param        phone      query string false "Telefon bo'yicha qidirish (ILIKE)"
// @Param        role       query string false "Aniq rol"
// @Param        is_active  query bool   false "Holat"
// @Success      200 {object} map[string]int64
// @Router       /count/users [get]
func (h *userHandler) Count(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	f := user_dto.UserFilter{
		FullName: helper.QueryString(r, "full_name", 100),
		Phone:    helper.QueryString(r, "phone", 50),
		Role:     helper.QueryString(r, "role", 50),
		IsActive: helper.QueryBool(r, "is_active"),
	}

	total, err := h.service.Count(r.Context(), f)

	if err != nil {
		helper.WriteInternalError(w, err)

		return
	}

	helper.WriteJSON(w, http.StatusOK, map[string]int64{"total": total})
}

func userSortValue(sortBy string, u *user_dto.UserResponse) string {
	switch sortBy {
	case "full_name":
		return u.FullName
	case "phone":
		return u.Phone
	case "role":
		return u.Role
	case "is_active":
		return strconv.FormatBool(u.IsActive)
	case "created_at":
		return u.CreatedAt.Format("2006-01-02 15:04:05.000000-07:00")
	case "updated_at":
		return u.UpdatedAt.Format("2006-01-02 15:04:05.000000-07:00")
	}

	return ""
}

// Show godoc
// @Summary      Foydalanuvchini ID bo'yicha olish (admin)
// @Tags         Users
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "User ID"
// @Success      200 {object} user_dto.UserResponse
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      403 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Router       /users/{id} [get]
func (h *userHandler) Show(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.ParseInt(ps.ByName("id"), 10, 64)

	if err != nil || id <= 0 {
		helper.WriteError(w, http.StatusBadRequest, "invalid id")

		return
	}

	resp, err := h.service.Show(r.Context(), id)

	if err != nil {
		helper.WriteError(w, http.StatusNotFound, "user not found")

		return
	}

	helper.WriteJSON(w, http.StatusOK, resp)
}

// Update godoc
// @Summary      Foydalanuvchini yangilash (admin)
// @Tags         Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path int                       true  "User ID"
// @Param        body body user_dto.UpdateUserRequest true "Yangilanadigan maydonlar"
// @Success      200 {object} user_dto.UserResponse
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      403 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Failure      422 {object} map[string]any
// @Router       /users/{id} [put]
func (h *userHandler) Update(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.ParseInt(ps.ByName("id"), 10, 64)

	if err != nil || id <= 0 {
		helper.WriteError(w, http.StatusBadRequest, "invalid id")

		return
	}

	var req user_dto.UpdateUserRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.WriteError(w, http.StatusBadRequest, "invalid JSON")

		return
	}

	if req.FullName != nil {
		trimmed := strings.TrimSpace(*req.FullName)

		req.FullName = &trimmed
	}

	if req.Phone != nil {
		trimmed := strings.TrimSpace(*req.Phone)

		req.Phone = &trimmed
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		helper.WriteValidation(w, errs)

		return
	}

	resp, err := h.service.Update(r.Context(), id, req)

	if err != nil {
		helper.WriteError(w, http.StatusNotFound, err.Error())

		return
	}

	helper.WriteJSON(w, http.StatusOK, resp)
}

// Delete godoc
// @Summary      Foydalanuvchini o'chirish (soft delete, admin)
// @Tags         Users
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "User ID"
// @Success      200 {object} map[string]string
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      403 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Router       /users/{id} [delete]
func (h *userHandler) Delete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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
