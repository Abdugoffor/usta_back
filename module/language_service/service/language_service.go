package language_service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"main_service/helper"
	language_dto "main_service/module/language_service/dto"

	"github.com/jackc/pgx/v5/pgxpool"
)

var LanguageSortMap = helper.SortMap{
	"id":         {Col: "id", Type: "bigint"},
	"name":       {Col: "name", Type: "varchar"},
	"is_active":  {Col: "is_active", Type: "boolean"},
	"created_at": {Col: "created_at", Type: "timestamptz"},
	"updated_at": {Col: "updated_at", Type: "timestamptz"},
}

type LanguageService interface {
	Create(ctx context.Context, req language_dto.CreateLanguageRequest) (*language_dto.LanguageResponse, error)

	Show(ctx context.Context, id int64) (*language_dto.LanguageResponse, error)

	Update(ctx context.Context, id int64, req language_dto.UpdateLanguageRequest) (*language_dto.LanguageResponse, error)

	Delete(ctx context.Context, id int64) error

	List(ctx context.Context, f language_dto.LanguageFilter, q helper.AdminListQuery) ([]*language_dto.LanguageResponse, bool, error)

	Count(ctx context.Context, f language_dto.LanguageFilter) (int64, error)
}

type languageService struct {
	db *pgxpool.Pool
}

func NewLanguageService(db *pgxpool.Pool) LanguageService {
	return &languageService{db: db}
}

func (s *languageService) Create(ctx context.Context, req language_dto.CreateLanguageRequest) (*language_dto.LanguageResponse, error) {
	isActive := true

	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	var r language_dto.LanguageResponse

	err := s.db.QueryRow(ctx, `INSERT INTO languages (name, description, is_active) VALUES ($1, $2, $3) RETURNING id, name, COALESCE(description, ''), is_active, created_at, updated_at`,
		strings.ToLower(strings.TrimSpace(req.Name)), req.Description, isActive,
	).Scan(&r.ID, &r.Name, &r.Description, &r.IsActive, &r.CreatedAt, &r.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (s *languageService) Show(ctx context.Context, id int64) (*language_dto.LanguageResponse, error) {
	var r language_dto.LanguageResponse

	err := s.db.QueryRow(ctx, `SELECT id, name, COALESCE(description, ''), is_active, created_at, updated_at FROM languages WHERE id = $1 AND deleted_at IS NULL`,
		id,
	).Scan(&r.ID, &r.Name, &r.Description, &r.IsActive, &r.CreatedAt, &r.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (s *languageService) Update(ctx context.Context, id int64, req language_dto.UpdateLanguageRequest) (*language_dto.LanguageResponse, error) {
	var name *string

	if req.Name != nil {
		v := strings.ToLower(strings.TrimSpace(*req.Name))

		name = &v
	}

	var r language_dto.LanguageResponse

	err := s.db.QueryRow(ctx, `UPDATE languages SET name = COALESCE($1::varchar, name), description = COALESCE($2::text, description), is_active = COALESCE($3::boolean, is_active), updated_at = NOW() WHERE id = $4 AND deleted_at IS NULL RETURNING id, name, COALESCE(description, ''), is_active, created_at, updated_at`,
		name, req.Description, req.IsActive, id,
	).Scan(&r.ID, &r.Name, &r.Description, &r.IsActive, &r.CreatedAt, &r.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("language not found")
	}

	return &r, nil
}

func (s *languageService) Delete(ctx context.Context, id int64) error {
	tag, err := s.db.Exec(ctx, `UPDATE languages SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`, time.Now(), id)

	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("language not found")
	}

	return nil
}

const languageListWhere = `l.deleted_at IS NULL AND ($1::varchar IS NULL OR l.name ILIKE $1) AND ($2::text IS NULL OR l.description ILIKE $2) AND ($3::boolean IS NULL OR l.is_active = $3)`

func languageListArgs(f language_dto.LanguageFilter) []any {
	return []any{helper.LikePattern(f.Name), helper.LikePattern(f.Description), f.IsActive}
}

func (s *languageService) Count(ctx context.Context, f language_dto.LanguageFilter) (int64, error) {
	var total int64

	err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM languages l WHERE `+languageListWhere, languageListArgs(f)...).Scan(&total)

	return total, err
}

func (s *languageService) List(ctx context.Context, f language_dto.LanguageFilter, q helper.AdminListQuery) ([]*language_dto.LanguageResponse, bool, error) {
	spec, _ := LanguageSortMap.Resolve(q.SortBy)

	orderDir := q.Order()

	args := languageListArgs(f)

	cursorSQL, cursorArgs := helper.BuildCursorClause(len(args)+1, "id", spec, orderDir, q.Cursor)

	if cursorSQL == "" {
		cursorSQL = "TRUE"
	}

	args = append(args, cursorArgs...)

	args = append(args, q.Limit+1)

	rows, err := s.db.Query(ctx, fmt.Sprintf(`SELECT id, name, description, is_active, created_at, updated_at FROM (SELECT l.id, l.name, COALESCE(l.description, '') AS description, l.is_active, l.created_at, l.updated_at FROM languages l WHERE %s) l WHERE %s ORDER BY %s %s, id %s LIMIT $%d`,
		languageListWhere, cursorSQL, spec.Col, orderDir, orderDir, len(args),
	), args...)

	if err != nil {
		return nil, false, err
	}

	defer rows.Close()

	items := make([]*language_dto.LanguageResponse, 0, q.Limit)

	for rows.Next() {
		var r language_dto.LanguageResponse

		if err := rows.Scan(&r.ID, &r.Name, &r.Description, &r.IsActive, &r.CreatedAt, &r.UpdatedAt); err != nil {
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
