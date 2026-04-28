package user_service

import (
	"context"
	"fmt"
	"main_service/helper"
	user_dto "main_service/module/user_service/dto"
	user_model "main_service/module/user_service/model"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

var UserSortMap = helper.SortMap{
	"id":         {Col: "id", Type: "bigint"},
	"full_name":  {Col: "full_name", Type: "varchar"},
	"phone":      {Col: "phone", Type: "varchar"},
	"role":       {Col: "role", Type: "varchar"},
	"is_active":  {Col: "is_active", Type: "boolean"},
	"created_at": {Col: "created_at", Type: "timestamptz"},
	"updated_at": {Col: "updated_at", Type: "timestamptz"},
}

type UserService interface {
	Register(ctx context.Context, req user_dto.RegisterRequest) (*user_dto.AuthResponse, error)

	Login(ctx context.Context, req user_dto.LoginRequest) (*user_dto.AuthResponse, error)

	Create(ctx context.Context, req user_dto.CreateUserRequest) (*user_dto.UserResponse, error)

	List(ctx context.Context, f user_dto.UserFilter, q helper.AdminListQuery) ([]*user_dto.UserResponse, bool, error)

	Count(ctx context.Context, f user_dto.UserFilter) (int64, error)

	Show(ctx context.Context, id int64) (*user_dto.UserResponse, error)

	Update(ctx context.Context, id int64, req user_dto.UpdateUserRequest) (*user_dto.UserResponse, error)

	Delete(ctx context.Context, id int64) error
}

type userService struct {
	db *pgxpool.Pool
}

func NewUserService(db *pgxpool.Pool) UserService {
	return &userService{db: db}
}

func (s *userService) Register(ctx context.Context, req user_dto.RegisterRequest) (*user_dto.AuthResponse, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

	if err != nil {
		return nil, err
	}

	var (
		user     user_model.User
		password string
	)

	err = s.db.QueryRow(ctx, `INSERT INTO users (full_name, phone, password, role) VALUES ($1, $2, $3, $4) RETURNING id, full_name, photo, phone, password, role, is_active, created_at, updated_at, deleted_at`,
		req.FullName, req.Phone, string(hash), "user",
	).Scan(
		&user.ID, &user.FullName, &user.Photo, &user.Phone, &password,
		&user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("phone allaqachon ro'yxatdan o'tgan")
	}

	token, err := helper.GenerateToken(int(user.ID), user.Role)

	if err != nil {
		return nil, err
	}

	return &user_dto.AuthResponse{Token: token, User: user}, nil
}

func (s *userService) Login(ctx context.Context, req user_dto.LoginRequest) (*user_dto.AuthResponse, error) {
	var (
		user     user_model.User
		password string
	)

	err := s.db.QueryRow(ctx, `SELECT id, full_name, photo, phone, password, role, is_active, created_at, updated_at, deleted_at FROM users WHERE phone = $1 AND deleted_at IS NULL`,
		req.Phone,
	).Scan(
		&user.ID, &user.FullName, &user.Photo, &user.Phone, &password,
		&user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("phone yoki parol noto'g'ri")
	}

	if !user.IsActive {
		return nil, fmt.Errorf("account faol emas")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(password), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("phone yoki parol noto'g'ri")
	}

	token, err := helper.GenerateToken(int(user.ID), user.Role)

	if err != nil {
		return nil, err
	}

	return &user_dto.AuthResponse{Token: token, User: user}, nil
}

func (s *userService) Create(ctx context.Context, req user_dto.CreateUserRequest) (*user_dto.UserResponse, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

	if err != nil {
		return nil, err
	}

	role := "user"

	if req.Role != "" {
		role = req.Role
	}

	isActive := true

	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	var r user_dto.UserResponse

	err = s.db.QueryRow(ctx, `INSERT INTO users (full_name, phone, photo, password, role, is_active) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, full_name, photo, phone, role, is_active, created_at, updated_at`,
		strings.TrimSpace(req.FullName), strings.TrimSpace(req.Phone), req.Photo, string(hash), role, isActive,
	).Scan(&r.ID, &r.FullName, &r.Photo, &r.Phone, &r.Role, &r.IsActive, &r.CreatedAt, &r.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("phone allaqachon ro'yxatdan o'tgan")
	}

	return &r, nil
}

func (s *userService) Show(ctx context.Context, id int64) (*user_dto.UserResponse, error) {
	var r user_dto.UserResponse

	err := s.db.QueryRow(ctx, `SELECT id, full_name, photo, phone, role, is_active, created_at, updated_at FROM users WHERE id = $1 AND deleted_at IS NULL`,
		id,
	).Scan(&r.ID, &r.FullName, &r.Photo, &r.Phone, &r.Role, &r.IsActive, &r.CreatedAt, &r.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (s *userService) Update(ctx context.Context, id int64, req user_dto.UpdateUserRequest) (*user_dto.UserResponse, error) {
	var fullName, phone, photo, password, role *string

	if req.FullName != nil {
		v := strings.TrimSpace(*req.FullName)

		fullName = &v
	}

	if req.Phone != nil {
		v := strings.TrimSpace(*req.Phone)

		phone = &v
	}

	if req.Photo != nil {
		photo = req.Photo
	}

	if req.Password != nil {
		hash, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)

		if err != nil {
			return nil, err
		}

		v := string(hash)

		password = &v
	}

	if req.Role != nil {
		role = req.Role
	}

	var r user_dto.UserResponse

	err := s.db.QueryRow(ctx, `UPDATE users SET full_name = COALESCE($1::varchar, full_name), phone = COALESCE($2::varchar, phone), photo = COALESCE($3::varchar, photo), password = COALESCE($4::text, password), role = COALESCE($5::varchar, role), is_active = COALESCE($6::boolean, is_active), updated_at = NOW() WHERE id = $7 AND deleted_at IS NULL RETURNING id, full_name, photo, phone, role, is_active, created_at, updated_at`,
		fullName, phone, photo, password, role, req.IsActive, id,
	).Scan(&r.ID, &r.FullName, &r.Photo, &r.Phone, &r.Role, &r.IsActive, &r.CreatedAt, &r.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	return &r, nil
}

func (s *userService) Delete(ctx context.Context, id int64) error {
	tag, err := s.db.Exec(ctx, `UPDATE users SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`, time.Now(), id)

	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

const userListWhere = `u.deleted_at IS NULL AND ($1::varchar IS NULL OR u.full_name ILIKE $1) AND ($2::varchar IS NULL OR u.phone ILIKE $2) AND ($3::varchar IS NULL OR u.role = $3) AND ($4::boolean IS NULL OR u.is_active = $4)`

func userListArgs(f user_dto.UserFilter) []any {
	var role *string

	if f.Role != "" {
		v := f.Role

		role = &v
	}

	return []any{helper.LikePattern(f.FullName), helper.LikePattern(f.Phone), role, f.IsActive}
}

func (s *userService) Count(ctx context.Context, f user_dto.UserFilter) (int64, error) {
	var total int64

	err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM users u WHERE `+userListWhere, userListArgs(f)...).Scan(&total)

	return total, err
}

func (s *userService) List(ctx context.Context, f user_dto.UserFilter, q helper.AdminListQuery) ([]*user_dto.UserResponse, bool, error) {
	spec, _ := UserSortMap.Resolve(q.SortBy)

	orderDir := q.Order()

	args := userListArgs(f)

	cursorSQL, cursorArgs := helper.BuildCursorClause(len(args)+1, "id", spec, orderDir, q.Cursor)

	if cursorSQL == "" {
		cursorSQL = "TRUE"
	}

	args = append(args, cursorArgs...)

	args = append(args, q.Limit+1)

	rows, err := s.db.Query(ctx, fmt.Sprintf(`SELECT id, full_name, photo, phone, role, is_active, created_at, updated_at FROM (SELECT u.id, u.full_name, u.photo, u.phone, u.role, u.is_active, u.created_at, u.updated_at FROM users u WHERE %s) u WHERE %s ORDER BY %s %s, id %s LIMIT $%d`,
		userListWhere, cursorSQL, spec.Col, orderDir, orderDir, len(args),
	), args...)

	if err != nil {
		return nil, false, err
	}

	defer rows.Close()

	items := make([]*user_dto.UserResponse, 0, q.Limit)

	for rows.Next() {
		var r user_dto.UserResponse

		if err := rows.Scan(&r.ID, &r.FullName, &r.Photo, &r.Phone, &r.Role, &r.IsActive, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, false, err
		}

		items = append(items, &r)
	}

	if err := rows.Err(); err != nil {
		return nil, false, err
	}

	hasMore := len(items) > q.Limit

	if hasMore {
		items = items[:q.Limit]
	}

	return items, hasMore, nil
}
