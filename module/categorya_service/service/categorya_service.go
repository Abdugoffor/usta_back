package categorya_service

import (
	"context"
	"encoding/json"
	"fmt"
	"main_service/helper"
	categorya_dto "main_service/module/categorya_service/dto"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var CategorySortMap = helper.SortMap{
	"id":         {Col: "id", Type: "bigint"},
	"name":       {Col: `(name->>'default')`, Type: "text"},
	"is_active":  {Col: "is_active", Type: "boolean"},
	"created_at": {Col: "created_at", Type: "timestamptz"},
	"updated_at": {Col: "updated_at", Type: "timestamptz"},
}

type CategoryService interface {
	Create(ctx context.Context, req categorya_dto.CreateCategoryRequest) (*categorya_dto.CategoryResponse, error)

	Show(ctx context.Context, id int64) (*categorya_dto.CategoryResponse, error)

	Update(ctx context.Context, id int64, req categorya_dto.UpdateCategoryRequest) (*categorya_dto.CategoryResponse, error)

	Delete(ctx context.Context, id int64) error

	List(ctx context.Context, f categorya_dto.CategoryFilter, q helper.AdminListQuery) ([]*categorya_dto.CategoryResponse, bool, error)

	Count(ctx context.Context, f categorya_dto.CategoryFilter) (int64, error)

	GetActiveLanguages(ctx context.Context) ([]string, error)
}

type categoryService struct {
	db *pgxpool.Pool
}

func NewCategoryService(db *pgxpool.Pool) CategoryService {
	return &categoryService{db: db}
}

func (s *categoryService) GetActiveLanguages(ctx context.Context) ([]string, error) {
	return helper.FetchActiveLanguages(ctx, s.db)
}

func unmarshalName(b []byte) map[string]string {
	if len(b) == 0 {
		return map[string]string{}
	}

	var m map[string]string

	if err := json.Unmarshal(b, &m); err != nil {
		return map[string]string{}
	}

	return m
}

func (s *categoryService) Create(ctx context.Context, req categorya_dto.CreateCategoryRequest) (*categorya_dto.CategoryResponse, error) {
	isActive := true

	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	nameJSON, err := json.Marshal(helper.NormalizeMultilang(req.Name))

	if err != nil {
		return nil, err
	}

	var (
		r         categorya_dto.CategoryResponse
		nameBytes []byte
	)

	err = s.db.QueryRow(ctx, `INSERT INTO categories (name, is_active) VALUES ($1::jsonb, $2) RETURNING id, name, COALESCE(is_active, FALSE), created_at, updated_at, deleted_at`,
		string(nameJSON), isActive,
	).Scan(&r.ID, &nameBytes, &r.IsActive, &r.CreatedAt, &r.UpdatedAt, &r.DeletedAt)

	if err != nil {
		return nil, err
	}

	r.Name = unmarshalName(nameBytes)

	return &r, nil
}

func (s *categoryService) Show(ctx context.Context, id int64) (*categorya_dto.CategoryResponse, error) {
	var (
		r         categorya_dto.CategoryResponse
		nameBytes []byte
	)

	err := s.db.QueryRow(ctx, `SELECT id, name, COALESCE(is_active, FALSE), created_at, updated_at, deleted_at FROM categories WHERE id = $1 AND deleted_at IS NULL`,
		id,
	).Scan(&r.ID, &nameBytes, &r.IsActive, &r.CreatedAt, &r.UpdatedAt, &r.DeletedAt)

	if err != nil {
		return nil, err
	}

	r.Name = unmarshalName(nameBytes)

	return &r, nil
}

func (s *categoryService) Update(ctx context.Context, id int64, req categorya_dto.UpdateCategoryRequest) (*categorya_dto.CategoryResponse, error) {
	var nameJSON *string

	if req.Name != nil {
		b, err := json.Marshal(helper.NormalizeMultilang(*req.Name))

		if err != nil {
			return nil, err
		}

		v := string(b)

		nameJSON = &v
	}

	var (
		r         categorya_dto.CategoryResponse
		nameBytes []byte
	)

	err := s.db.QueryRow(ctx, `UPDATE categories SET name = COALESCE($1::jsonb, name), is_active = COALESCE($2::boolean, is_active), updated_at = NOW() WHERE id = $3 AND deleted_at IS NULL RETURNING id, name, COALESCE(is_active, FALSE), created_at, updated_at, deleted_at`,
		nameJSON, req.IsActive, id,
	).Scan(&r.ID, &nameBytes, &r.IsActive, &r.CreatedAt, &r.UpdatedAt, &r.DeletedAt)

	if err != nil {
		return nil, fmt.Errorf("category not found")
	}

	r.Name = unmarshalName(nameBytes)

	return &r, nil
}

func (s *categoryService) Delete(ctx context.Context, id int64) error {
	tag, err := s.db.Exec(ctx, `UPDATE categories SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`, time.Now(), id)

	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("category not found")
	}

	return nil
}

const categoryListWhere = `c.deleted_at IS NULL AND ($1::text IS NULL OR c.name::text ILIKE $1) AND ($2::boolean IS NULL OR c.is_active = $2)`

func categoryListArgs(f categorya_dto.CategoryFilter) []any {
	return []any{helper.LikePattern(f.Name), f.IsActive}
}

func (s *categoryService) Count(ctx context.Context, f categorya_dto.CategoryFilter) (int64, error) {
	var total int64

	err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM categories c WHERE `+categoryListWhere, categoryListArgs(f)...).Scan(&total)

	return total, err
}

func (s *categoryService) List(ctx context.Context, f categorya_dto.CategoryFilter, q helper.AdminListQuery) ([]*categorya_dto.CategoryResponse, bool, error) {
	spec, _ := CategorySortMap.Resolve(q.SortBy)

	orderDir := q.Order()

	args := categoryListArgs(f)

	cursorSQL, cursorArgs := helper.BuildCursorClause(len(args)+1, "id", spec, orderDir, q.Cursor)

	if cursorSQL == "" {
		cursorSQL = "TRUE"
	}

	args = append(args, cursorArgs...)

	args = append(args, q.Limit+1)

	rows, err := s.db.Query(ctx, fmt.Sprintf(`SELECT id, name, is_active, created_at, updated_at, deleted_at FROM (SELECT c.id, c.name, COALESCE(c.is_active, FALSE) AS is_active, c.created_at, c.updated_at, c.deleted_at FROM categories c WHERE %s) c WHERE %s ORDER BY %s %s NULLS LAST, id %s LIMIT $%d`,
		categoryListWhere, cursorSQL, spec.Col, orderDir, orderDir, len(args),
	), args...)

	if err != nil {
		return nil, false, err
	}

	defer rows.Close()

	items := make([]*categorya_dto.CategoryResponse, 0, q.Limit)

	for rows.Next() {
		var (
			r         categorya_dto.CategoryResponse
			nameBytes []byte
		)

		if err := rows.Scan(&r.ID, &nameBytes, &r.IsActive, &r.CreatedAt, &r.UpdatedAt, &r.DeletedAt); err != nil {
			return nil, false, err
		}

		r.Name = unmarshalName(nameBytes)

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
