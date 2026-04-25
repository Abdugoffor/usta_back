package translation_service

import (
	"context"
	"encoding/json"
	"fmt"
	"main_service/helper"
	translation_dto "main_service/module/translations_service/dto"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var TranslationSortMap = helper.SortMap{
	"id":         {Col: "id", Type: "bigint"},
	"slug":       {Col: "slug", Type: "varchar"},
	"name":       {Col: `(name->>'default')`, Type: "text"},
	"is_active":  {Col: "is_active", Type: "boolean"},
	"created_at": {Col: "created_at", Type: "timestamptz"},
	"updated_at": {Col: "updated_at", Type: "timestamptz"},
}

type TranslationService interface {
	Create(ctx context.Context, req translation_dto.CreateTranslationRequest) (*translation_dto.TranslationResponse, error)

	GetByID(ctx context.Context, id int64) (*translation_dto.TranslationResponse, error)

	Update(ctx context.Context, id int64, req translation_dto.UpdateTranslationRequest) (*translation_dto.TranslationResponse, error)

	Delete(ctx context.Context, id int64) error

	List(ctx context.Context, f translation_dto.TranslationFilter, q helper.AdminListQuery) ([]*translation_dto.TranslationResponse, bool, error)

	Count(ctx context.Context, f translation_dto.TranslationFilter) (int64, error)

	GetActiveLanguages(ctx context.Context) ([]string, error)

	GetTranslation(ctx context.Context, slug, lang string) string
}

type translationService struct {
	db *pgxpool.Pool
}

func NewTranslationService(db *pgxpool.Pool) TranslationService {
	return &translationService{db: db}
}

func (s *translationService) GetActiveLanguages(ctx context.Context) ([]string, error) {
	return helper.FetchActiveLanguages(ctx, s.db)
}

func (s *translationService) GetTranslation(ctx context.Context, slug, lang string) string {
	var nameBytes []byte

	err := s.db.QueryRow(ctx, `SELECT name FROM translations WHERE slug = $1 AND deleted_at IS NULL AND is_active = TRUE`, slug).Scan(&nameBytes)

	if err != nil {
		return slug
	}

	var name map[string]string

	if err := json.Unmarshal(nameBytes, &name); err != nil {
		return slug
	}

	if v, ok := name[strings.ToLower(lang)]; ok && strings.TrimSpace(v) != "" {
		return v
	}

	if v, ok := name["default"]; ok && strings.TrimSpace(v) != "" {
		return v
	}

	return slug
}

func (s *translationService) Create(ctx context.Context, req translation_dto.CreateTranslationRequest) (*translation_dto.TranslationResponse, error) {
	isActive := true

	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	nameJSON, err := json.Marshal(helper.NormalizeMultilang(req.Name))

	if err != nil {
		return nil, err
	}

	var (
		r         translation_dto.TranslationResponse
		nameBytes []byte
	)

	err = s.db.QueryRow(ctx, `INSERT INTO translations (slug, name, is_active) VALUES ($1, $2::jsonb, $3) RETURNING id, slug, name, COALESCE(is_active, FALSE), created_at, updated_at`,
		strings.TrimSpace(req.Slug), string(nameJSON), isActive,
	).Scan(&r.ID, &r.Slug, &nameBytes, &r.IsActive, &r.CreatedAt, &r.UpdatedAt)

	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return nil, fmt.Errorf("slug allaqachon mavjud")
		}

		return nil, err
	}

	full := map[string]string{}

	if len(nameBytes) > 0 {
		_ = json.Unmarshal(nameBytes, &full)
	}

	activeLangs, err := s.GetActiveLanguages(ctx)

	if err != nil {
		return nil, err
	}

	r.Name = helper.FilterMultilangByActive(full, activeLangs)

	return &r, nil
}

func (s *translationService) GetByID(ctx context.Context, id int64) (*translation_dto.TranslationResponse, error) {
	var (
		r         translation_dto.TranslationResponse
		nameBytes []byte
	)

	err := s.db.QueryRow(ctx, `SELECT id, slug, name, COALESCE(is_active, FALSE), created_at, updated_at FROM translations WHERE id = $1 AND deleted_at IS NULL`,
		id,
	).Scan(&r.ID, &r.Slug, &nameBytes, &r.IsActive, &r.CreatedAt, &r.UpdatedAt)

	if err != nil {
		return nil, err
	}

	full := map[string]string{}

	if len(nameBytes) > 0 {
		if err := json.Unmarshal(nameBytes, &full); err != nil {
			return nil, err
		}
	}

	activeLangs, err := s.GetActiveLanguages(ctx)

	if err != nil {
		return nil, err
	}

	r.Name = helper.FilterMultilangByActive(full, activeLangs)

	return &r, nil
}

func (s *translationService) Update(ctx context.Context, id int64, req translation_dto.UpdateTranslationRequest) (*translation_dto.TranslationResponse, error) {
	var slug, nameJSON *string

	if req.Slug != nil {
		v := strings.TrimSpace(*req.Slug)

		slug = &v
	}

	if req.Name != nil {
		b, err := json.Marshal(helper.NormalizeMultilang(*req.Name))

		if err != nil {
			return nil, err
		}

		v := string(b)

		nameJSON = &v
	}

	var (
		r         translation_dto.TranslationResponse
		nameBytes []byte
	)

	err := s.db.QueryRow(ctx, `UPDATE translations SET slug = COALESCE($1::varchar, slug), name = COALESCE($2::jsonb, name), is_active = COALESCE($3::boolean, is_active), updated_at = NOW() WHERE id = $4 AND deleted_at IS NULL RETURNING id, slug, name, COALESCE(is_active, FALSE), created_at, updated_at`,
		slug, nameJSON, req.IsActive, id,
	).Scan(&r.ID, &r.Slug, &nameBytes, &r.IsActive, &r.CreatedAt, &r.UpdatedAt)

	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return nil, fmt.Errorf("slug allaqachon mavjud")
		}

		return nil, fmt.Errorf("translation not found")
	}

	full := map[string]string{}

	if len(nameBytes) > 0 {
		_ = json.Unmarshal(nameBytes, &full)
	}

	activeLangs, err := s.GetActiveLanguages(ctx)

	if err != nil {
		return nil, err
	}

	r.Name = helper.FilterMultilangByActive(full, activeLangs)

	return &r, nil
}

func (s *translationService) Delete(ctx context.Context, id int64) error {
	tag, err := s.db.Exec(ctx, `UPDATE translations SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`, time.Now(), id)

	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("translation not found")
	}

	return nil
}

const translationListWhere = `t.deleted_at IS NULL AND ($1::varchar IS NULL OR t.slug ILIKE $1) AND ($2::text IS NULL OR t.name::text ILIKE $2) AND ($3::boolean IS NULL OR t.is_active = $3)`

func translationListArgs(f translation_dto.TranslationFilter) []any {
	return []any{helper.LikePattern(f.Slug), helper.LikePattern(f.Name), f.IsActive}
}

func (s *translationService) Count(ctx context.Context, f translation_dto.TranslationFilter) (int64, error) {
	var total int64

	err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM translations t WHERE `+translationListWhere, translationListArgs(f)...).Scan(&total)

	return total, err
}

func (s *translationService) List(ctx context.Context, f translation_dto.TranslationFilter, q helper.AdminListQuery) ([]*translation_dto.TranslationResponse, bool, error) {
	spec, _ := TranslationSortMap.Resolve(q.SortBy)

	orderDir := q.Order()

	args := translationListArgs(f)

	cursorSQL, cursorArgs := helper.BuildCursorClause(len(args)+1, "id", spec, orderDir, q.Cursor)

	if cursorSQL == "" {
		cursorSQL = "TRUE"
	}

	args = append(args, cursorArgs...)

	args = append(args, q.Limit+1)

	rows, err := s.db.Query(ctx, fmt.Sprintf(`SELECT id, slug, name, is_active, created_at, updated_at FROM (SELECT t.id, t.slug, t.name, COALESCE(t.is_active, FALSE) AS is_active, t.created_at, t.updated_at FROM translations t WHERE %s) t WHERE %s ORDER BY %s %s NULLS LAST, id %s LIMIT $%d`,
		translationListWhere, cursorSQL, spec.Col, orderDir, orderDir, len(args),
	), args...)

	if err != nil {
		return nil, false, err
	}

	defer rows.Close()

	activeLangs, err := s.GetActiveLanguages(ctx)

	if err != nil {
		return nil, false, err
	}

	items := make([]*translation_dto.TranslationResponse, 0, q.Limit)

	for rows.Next() {
		var (
			r         translation_dto.TranslationResponse
			nameBytes []byte
		)

		if err := rows.Scan(&r.ID, &r.Slug, &nameBytes, &r.IsActive, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, false, err
		}

		full := map[string]string{}

		if len(nameBytes) > 0 {
			if err := json.Unmarshal(nameBytes, &full); err != nil {
				return nil, false, err
			}
		}

		r.Name = helper.FilterMultilangByActive(full, activeLangs)

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
